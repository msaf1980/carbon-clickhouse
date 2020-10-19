package uploader

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/lomik/carbon-clickhouse/helper/RowBinary"
	"github.com/lomik/zapwriter"
	"github.com/mailru/dbr"
	_ "github.com/mailru/go-clickhouse"
	"github.com/msaf1980/stringutils"
	"go.uber.org/zap"
)

type Index struct {
	*cached
}

var _ Uploader = &Index{}
var _ UploaderWithReset = &Index{}

const ReverseLevelOffset = 10000
const TreeLevelOffset = 20000
const ReverseTreeLevelOffset = 30000

const DefaultTreeDate = 42 // 1970-02-12

type indexRow struct {
	days      uint16
	name      []byte
	found     bool
	foundTree bool
}

const indexCacheBatchSize = 50000
const indexQuery = "SELECT Path, groupUniqArray(Date) as Dates FROM %s WHERE Date IN (?) AND Path IN (?) GROUP BY Path"

func NewIndex(base *Base) *Index {
	u := &Index{}
	u.cached = newCached(base)
	u.cached.parser = u.parseFile
	u.query = fmt.Sprintf(indexQuery, u.config.TableName)
	return u
}

func filterQueryAddPath(sb *stringutils.Builder, name string, n int) {
	if n == 0 {
		//sb.WriteString("'" + name + "'")
		sb.WriteString("'")
	} else {
		//sb.WriteString(", '" + name + "'")
		sb.WriteString(", '")
	}
	sb.WriteString(name)
	sb.WriteString("'")
}

func daysToDate(days uint16) time.Time {
	return time.Unix(86400*int64(days), 0)
}

func (u *Index) cacheQueryCheck(connect *dbr.Connection, filterSb *stringutils.Builder, indexes map[string]indexRow,
	days uint16, treeDays uint16, filename string, checks, totalchecks, total int, prestartTime time.Time) error {

	logger := zapwriter.Logger("index")
	startTime := time.Now()
	dates := []string{daysToDate(days).Format("2006-01-02"), daysToDate(treeDays).Format("2006-01-02")}
	rows, err := connect.Query(u.query, days, filterSb.String())
	endTime := time.Now()
	if err != nil {
		logger.Error("cache", zap.String("paths", filterSb.String()), zap.Strings("date", dates),
			zap.String("filename", filename), zap.Error(err),
			zap.Duration("query_time", endTime.Sub(startTime)))
		return err
	}
	var found int
	var keyerror int
	for rows.Next() {
		var path string
		var sDates []time.Time
		if err = rows.Scan(&path, &sDates); err != nil {
			return err
		}
		found++
		key := strconv.Itoa(int(days)) + ":" + path
		v, ok := indexes[key]
		if !ok {
			keyerror++
			err = fmt.Errorf("key '%s' not found during index lookup, may be wrong filter generated", key)
			logger.Error("cache", zap.String("date", dates[0]), zap.Error(err))
			continue
		}
		for _, dd := range sDates {
			d := RowBinary.SlowTimeToDays(dd)
			if d == days {
				v.found = true
			} else {
				v.foundTree = true
			}
		}
		if v.found || v.foundTree {
			indexes[key] = v
		}
		if v.found && v.foundTree {
			u.cached.existsCache.Add(key, startTime.Unix())
		}
	}
	logger.Info("cache", zap.String("filename", filename),
		zap.Duration("query_time", endTime.Sub(startTime)), zap.Duration("time", time.Since(prestartTime)),
		zap.Int("keyerror", keyerror), zap.Int("found", found),
		zap.Int("checked", checks), zap.Int("processed", totalchecks), zap.Int("total", total))
	return nil
}

func (u *Index) cacheBatchRecheck(indexes map[string]indexRow, treeDays uint16, filename string, filterSb *stringutils.Builder) (map[string]bool, error) {
	newSeries := make(map[string]bool)
	if len(indexes) == 0 {
		return newSeries, nil
	}

	connect, err := dbr.Open("clickhouse", u.config.URL, nil)
	if err != nil {
		return nil, err
	}

	var n int
	var checks int
	var days uint16
	prestartTime := time.Now()
	filterSb.Reset()
	for _, v := range indexes {
		if n >= indexCacheBatchSize || (n > 0 && days != v.days) {
			checks += n
			err = u.cacheQueryCheck(connect, filterSb, indexes, days, treeDays, filename, n, checks, len(indexes), prestartTime)
			filterSb.Reset()
			if err != nil {
				return nil, err
			}
			if days != v.days {
				days = v.days
			}
			prestartTime = time.Now()
			n = 0
		}
		if days != v.days {
			days = v.days
			prestartTime = time.Now()
		}
		filterQueryAddPath(filterSb, unsafeString(v.name), n)
		n++
	}

	if n > 0 {
		checks += n
		err = u.cacheQueryCheck(connect, filterSb, indexes, days, treeDays, filename, n, checks, len(indexes), prestartTime)
		if err != nil {
			return nil, err
		}
	}

	for key, v := range indexes {
		if !v.found || !v.foundTree {
			newSeries[key] = true
		}
	}
	return newSeries, nil
}

func (u *Index) parseFile(filename string, out io.Writer) (uint64, uint64, uint64, map[string]bool, error) {
	var reader *RowBinary.Reader
	var err error
	var n uint64
	var skipped uint64
	var skippedTree uint64

	logger := zapwriter.Logger("index")
	startTime := time.Now()

	reader, err = RowBinary.NewReader(filename, false)
	if err != nil {
		return n, skipped, skippedTree, nil, err
	}
	defer reader.Close()

	version := uint32(time.Now().Unix())
	nocacheSeries := make(map[string]indexRow)
	newUniq := make(map[string]bool)
	wb := RowBinary.GetWriteBuffer()

	var level, index, l int
	var p []byte

	treeDate := uint16(DefaultTreeDate)
	if !u.config.TreeDate.IsZero() {
		treeDate = RowBinary.TimestampToDays(uint32(u.config.TreeDate.Unix()))
	}

LineLoop:
	for {
		name, err := reader.ReadRecord()
		if err != nil { // io.EOF or corrupted file
			break
		}

		// skip tagged
		if bytes.IndexByte(name, '?') >= 0 {
			continue
		}

		//key := fmt.Sprintf("%d:%s", reader.Days(), unsafeString(name))
		key := strconv.Itoa(int(reader.Days())) + ":" + unsafeString(name)

		if u.existsCache.Exists(key) {
			continue LineLoop
		}

		if _, ok := nocacheSeries[key]; ok {
			continue LineLoop
		}

		nocacheSeries[key] = indexRow{
			days: reader.Days(),
			name: []byte(string(name)),
		}
	}

	var filterSb stringutils.Builder
	if len(nocacheSeries) > 0 {
		size := indexCacheBatchSize * 200
		if len(nocacheSeries) < indexCacheBatchSize {
			size = len(nocacheSeries) * 200
		}
		if size > filterSb.Cap() {
			filterSb.Grow(size)
		}
	}
	newSeries, err := u.cacheBatchRecheck(nocacheSeries, treeDate, filename, &filterSb)
	if err != nil {
		return n, skipped, skippedTree, nil, err
	}

	for _, v := range nocacheSeries {
		if v.found {
			skipped++
			skippedTree++
			continue
		}
		n++

		wb.Reset()

		level = pathLevel(v.name)
		// Direct path with date
		wb.WriteUint16(v.days)
		wb.WriteUint32(uint32(level))
		wb.WriteBytes(v.name)
		wb.WriteUint32(version)

		reverseName := RowBinary.ReverseBytes(v.name)

		// Reverse path with date
		wb.WriteUint16(v.days)
		wb.WriteUint32(uint32(level + ReverseLevelOffset))
		wb.WriteBytes(reverseName)
		wb.WriteUint32(version)

		if v.foundTree {
			skippedTree++
		} else {
			// Tree
			wb.WriteUint16(treeDate)
			wb.WriteUint32(uint32(level + TreeLevelOffset))
			wb.WriteBytes(v.name)
			wb.WriteUint32(version)

			p = v.name
			l = level
			for l--; l > 0; l-- {
				index = bytes.LastIndexByte(p, '.')
				if newUniq[unsafeString(p[:index+1])] {
					break
				}

				newUniq[string(p[:index+1])] = true

				wb.WriteUint16(treeDate)
				wb.WriteUint32(uint32(l + TreeLevelOffset))
				wb.WriteBytes(p[:index+1])
				wb.WriteUint32(version)

				p = p[:index]
			}

			// Reverse path without date
			wb.WriteUint16(treeDate)
			wb.WriteUint32(uint32(level + ReverseTreeLevelOffset))
			wb.WriteBytes(reverseName)
			wb.WriteUint32(version)
		}

		_, err = out.Write(wb.Bytes())
		if err != nil {
			return n, skipped, skippedTree, nil, err
		}
	}

	wb.Release()

	logger.Info("cache", zap.String("filename", filename),
		zap.Duration("time", time.Since(startTime)),
		zap.Uint64("cachemiss", n), zap.Uint64("cachehit", uint64(len(nocacheSeries))-n))

	return n, skipped, skippedTree, newSeries, nil
}
