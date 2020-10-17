package uploader

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/lomik/carbon-clickhouse/helper/RowBinary"
	"github.com/lomik/zapwriter"
	_ "github.com/mailru/go-clickhouse"
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

func NewIndex(base *Base) *Index {
	u := &Index{}
	u.cached = newCached(base)
	u.cached.parser = u.parseFile
	return u
}

func filterQueryAddPath(sb *strings.Builder, name string, n int) {
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

const indexQueryS = "SELECT Date, Path FROM %s WHERE "
const indexQueryE = " GROUP BY Path, Date ORDER BY Path, Date ASC"

func indexQueryPrep(sb *strings.Builder, days uint16, treeDays uint16, tableName string) {
	sb.Reset()
	sb.WriteString(fmt.Sprintf(indexQueryS, tableName))

	t := time.Unix(86400*int64(days), 0)
	treeT := time.Unix(86400*int64(treeDays), 0)
	s := fmt.Sprintf("(Date IN ('%04d-%02d-%02d', '%04d-%02d-%02d')", treeT.Year(), treeT.Month(), treeT.Day(), t.Year(), t.Month(), t.Day())
	sb.WriteString(s)
	sb.WriteString(" AND Path IN (")
}

func indexQueryEnd(sb *strings.Builder) {
	sb.WriteString("))")
	sb.WriteString(indexQueryE)
}

func checkTreeIndex(indexes map[string]indexRow, path string, days uint16) error {
	keySearch := fmt.Sprintf("%d:%s", days, path)
	v, ok := indexes[keySearch]
	if ok {
		if !v.foundTree {
			v.foundTree = true
		}
		indexes[keySearch] = v
		return nil
	}
	return fmt.Errorf("root key '%s' not found during index lookup, may be wrong filter generated", keySearch)
}

func (u *Index) cacheQueryCheck(connect *sql.DB, sb *strings.Builder, indexes map[string]indexRow,
	days uint16, treeDays uint16, filename string, checks, totalchecks, total int, prestartTime time.Time) error {

	logger := zapwriter.Logger("index")
	indexQueryEnd(sb)
	query := sb.String()
	startTime := time.Now()
	rows, err := connect.Query(query)
	endTime := time.Now()
	if err != nil {
		logger.Error("cache", zap.String("query", query), zap.String("filename", filename), zap.Error(err),
			zap.Duration("query_time", endTime.Sub(startTime)))
		return err
	}
	var found int
	var keyerror int
	for rows.Next() {
		var date time.Time
		var path string
		if err := rows.Scan(&date, &path); err != nil {
			return err
		}
		found++
		d := RowBinary.SlowTimeToDays(date)
		if d == treeDays {
			err := checkTreeIndex(indexes, path, days)
			if err == nil {
			} else {
				logger.Error("cache", zap.String("filename", filename), zap.Uint16("days", days), zap.String("error", err.Error()))
			}
		} else {
			key := fmt.Sprintf("%d:%s", d, path)
			v, ok := indexes[key]
			if ok {
				v.found = true
				indexes[key] = v
			} else {
				keyerror++
				err = fmt.Errorf("key '%s' not found during index lookup, may be wrong filter generated", key)
				logger.Error("cache", zap.String("query", query), zap.Uint16("days", days), zap.Error(err))
			}
		}
	}
	logger.Info("cache", zap.String("filename", filename),
		zap.Duration("query_time", endTime.Sub(startTime)), zap.Duration("time", time.Since(prestartTime)),
		zap.Int("keyerror", keyerror), zap.Int("found", found),
		zap.Int("checked", checks), zap.Int("processed", totalchecks), zap.Int("total", total))
	return nil
}

func (u *Index) cacheBatchRecheck(indexes map[string]indexRow, treeDays uint16, filename string) (map[string]bool, error) {
	newSeries := make(map[string]bool)
	if len(indexes) == 0 {
		return newSeries, nil
	}

	connect, err := sql.Open("clickhouse", u.config.URL)
	if err != nil {
		return nil, err
	}
	if err := connect.Ping(); err != nil {
		return nil, err
	}

	var n int
	var checks int
	var days uint16
	var filterSb strings.Builder
	prestartTime := time.Now()
	for _, v := range indexes {
		if n >= 50000 || (n > 0 && days != v.days) {
			checks += n
			err = u.cacheQueryCheck(connect, &filterSb, indexes, days, treeDays, filename, n, checks, len(indexes), prestartTime)
			if err != nil {
				return nil, err
			}
			if days != v.days {
				days = v.days
			}
			prestartTime = time.Now()
			indexQueryPrep(&filterSb, days, treeDays, u.config.TableName)
			n = 0
		}
		if days != v.days {
			days = v.days
			prestartTime = time.Now()
			indexQueryPrep(&filterSb, days, treeDays, u.config.TableName)
		}
		filterQueryAddPath(&filterSb, unsafeString(v.name), n)
		n++
	}

	if n > 0 {
		checks += n
		err = u.cacheQueryCheck(connect, &filterSb, indexes, days, treeDays, filename, n, checks, len(indexes), prestartTime)
		if err != nil {
			return nil, err
		}

		filterSb.Reset()
	}

	for key, _ := range indexes {
		newSeries[key] = true
	}
	return newSeries, nil
}

func (u *Index) parseFile(filename string, out io.Writer) (uint64, uint64, uint64, map[string]bool, error) {
	var reader *RowBinary.Reader
	var err error
	var n uint64
	var skipped uint64
	var skippedTree uint64

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

		key := fmt.Sprintf("%d:%s", reader.Days(), unsafeString(name))

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

	newSeries, err := u.cacheBatchRecheck(nocacheSeries, treeDate, filename)
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

	return n, skipped, skippedTree, newSeries, nil
}
