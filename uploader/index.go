package uploader

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lomik/carbon-clickhouse/helper/RowBinary"
	"github.com/lomik/graphite-clickhouse/helper/clickhouse"
	"github.com/lomik/graphite-clickhouse/pkg/scope"
	"github.com/lomik/zapwriter"
	"github.com/msaf1980/stringutils"
	"go.uber.org/zap"
)

type Index struct {
	*cached
	opts clickhouse.Options
}

var _ Uploader = &Index{}
var _ UploaderWithReset = &Index{}

const ReverseLevelOffset = 10000
const TreeLevelOffset = 20000
const ReverseTreeLevelOffset = 30000

const DefaultTreeDate = 42 // 1970-02-12

type indexRow struct {
	days      uint16
	path      []byte
	found     bool
	foundTree bool
}

const indexCacheBatchSize = 50000

const indexQueryS = "SELECT Path, groupUniqArray(Date) as Dates FROM %s WHERE Date IN ('"
const indexQueryF = "') AND Path IN ("
const indexQueryE = ") GROUP BY Path FORMAT RowBinary"

func NewIndex(base *Base) *Index {
	u := &Index{}
	u.cached = newCached(base)
	u.cached.parser = u.parseFile
	u.cacheQuery = fmt.Sprintf(indexQueryS, u.config.TableName)
	u.opts.ConnectTimeout = u.config.Timeout.Duration
	u.opts.Timeout = u.config.Timeout.Duration
	return u
}

func daysToDate(days uint16) time.Time {
	return time.Unix(86400*int64(days), 0)
}

func (u *Index) cacheQuerySql(paths []indexRow, days uint16, treeDays uint16) string {
	var sb stringutils.Builder
	sb.Grow(len(paths) * 200)
	sb.WriteString(u.cacheQuery)
	sb.WriteString(daysToDate(days).Format("2006-01-02"))
	sb.WriteString("', '")
	sb.WriteString(daysToDate(treeDays).Format("2006-01-02"))
	sb.WriteString(indexQueryF)
	for n, v := range paths {
		if n == 0 {
			sb.WriteRune('\'')
		} else {
			sb.WriteString(", '")
		}
		sb.Write(v.path)
		sb.WriteRune('\'')
	}
	sb.WriteString(indexQueryE)

	return sb.String()
}

// TODO: may be move to github.com/lomik/graphite-clickhouse/helper/clickhouse like ReadUvarint
func readString(b []byte) (string, int, error) {
	n, readBytes, err := clickhouse.ReadUvarint(b)
	if err != nil {
		return "", 0, clickhouse.ErrClickHouseResponse
	}
	pathLen := int(n)
	if len(b) < pathLen+readBytes {
		return "", 0, clickhouse.ErrClickHouseResponse
	}
	return stringutils.UnsafeStringFromPtr(&b[readBytes], pathLen), pathLen + readBytes, nil
}

// dataPathDatesSplit is a split function for bufio.Scanner for read row binary response for queries with fileds:
// Path, array(Date)
func dataPathDatesSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(data) == 0 && atEOF {
		// stop
		return 0, nil, nil
	}
	var tokenLen int
	var fieldLen uint64
	var seek int

	// path
	fieldLen, seek, err = clickhouse.ReadUvarint(data)
	if err != nil {
		if atEOF {
			return 0, nil, err
		} else {
			// Request more data.
			return 0, nil, nil
		}
	}
	tokenLen += seek + int(fieldLen)
	if len(data) < tokenLen+1 {
		if atEOF {
			return 0, nil, clickhouse.ErrClickHouseResponse
		} else {
			// Request more data.
			return 0, nil, nil
		}
	}

	// dates
	fieldLen, seek, err = clickhouse.ReadUvarint(data[tokenLen:])
	if err != nil {
		if atEOF {
			return 0, nil, err
		} else {
			// Request more data.
			return 0, nil, nil
		}
	}
	tokenLen += seek + int(fieldLen)*2 // sizeof(UInt16) * array_size
	if len(data) < tokenLen {
		if atEOF {
			return 0, nil, clickhouse.ErrClickHouseResponse
		} else {
			// Request more data.
			return 0, nil, nil
		}
	}

	return tokenLen, data[:tokenLen], nil
}

func (u *Index) cacheQueryCheck(ctx context.Context, paths []indexRow, indexes map[string]indexRow,
	treeDays uint16, filename string, totalchecks, total int, prestartTime time.Time) error {

	logger := zapwriter.Logger("index")

	startTime := time.Now()
	days := paths[0].days
	sql := u.cacheQuerySql(paths, days, treeDays)
	reader, err := clickhouse.Reader(ctx, u.config.URL, sql, u.opts)
	endTime := time.Now()
	if err != nil {
		return err
	}
	var found int
	var keyerror int
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 1048576), 0)
	scanner.Split(dataPathDatesSplit)
	for scanner.Scan() {
		var pos int
		var n uint64
		var seek int
		var path string
		var dates [2]uint16
		// var sDates []time.Time
		rowStart := scanner.Bytes()
		if len(rowStart) == 0 {
			continue
		}
		path, seek, err = readString(rowStart)
		if err != nil {
			return err
		}
		pos += seek

		n, seek, err = clickhouse.ReadUvarint(rowStart[pos:])
		if err != nil {
			return err
		}
		datesCount := int(n)
		if len(rowStart) < datesCount+2*datesCount {
			return clickhouse.ErrClickHouseResponse
		}
		pos += seek
		for i := 0; i < datesCount; i++ {
			dates[i] = binary.LittleEndian.Uint16(rowStart[pos : pos+2])
			pos += 2
		}

		key := strconv.Itoa(int(days)) + ":" + path
		v, ok := indexes[key]
		if !ok {
			keyerror++
			s := "key '" + key + "' not found during index lookup, may be wrong filter generated"
			logger.Error("cache", zap.String("date", daysToDate(days).Format("2006-01-02")), zap.String("error", s))
			continue
		}
		found++
		for i := 0; i < datesCount; i++ {
			if dates[i] == days {
				v.found = true
			} else if dates[i] == treeDays {
				v.foundTree = true
			} else {
				s := fmt.Sprintf("day '%d' not found during index lookup, may be RowBinary parser error", dates[i])
				logger.Error("cache", zap.String("date", daysToDate(days).Format("2006-01-02")), zap.String("error", s))
			}
		}
		if v.found || v.foundTree {
			indexes[key] = v
		}
		if v.found && v.foundTree {
			u.cached.existsCache.Add(key, startTime.Unix())
		}
	}
	if err = scanner.Err(); err != nil {
		return err
	}
	logger.Info("cache", zap.String("filename", filename),
		zap.Duration("query_time", endTime.Sub(startTime)), zap.Duration("time", time.Since(prestartTime)),
		zap.Int("keyerror", keyerror), zap.Int("found", found),
		zap.Int("checked", len(paths)), zap.Int("processed", totalchecks), zap.Int("total", total))
	return nil
}

func geRequestIDFromFilename(filename string) string {
	if i := strings.LastIndex(filename, "."); i == -1 {
		return path.Base(filename)
	} else {
		return filename[i+1:]
	}
}

// indexNextDay get next day location on sorted slice
func indexNextDay(indexes []indexRow, start int, max int) int {
	end := len(indexes) - 1
	days := indexes[start].days
	if indexes[end].days == days {
		if len(indexes) > start+max {
			return start + max
		}
		return len(indexes)
	}

	pos := start
	for pos < end {
		median := pos + (end-pos)/2

		if indexes[median].days > days {
			end = median
		} else {
			pos = median + 1
		}
	}
	if pos > start+max {
		return start + max
	}
	return pos
}

func (u *Index) cacheBatchRecheck(indexes map[string]indexRow, treeDays uint16, filename string) (map[string]bool, error) {
	newSeries := make(map[string]bool)
	if len(indexes) == 0 {
		return newSeries, nil
	}

	if !u.config.NoQueryCache {
		var n int
		var checks int
		ctx := scope.WithRequestID(context.Background(), geRequestIDFromFilename(filename))
		prestartTime := time.Now()
		paths := make([]indexRow, len(indexes))
		for _, v := range indexes {
			paths[n] = v
			n++
		}
		// sort for index sequentional scan (on date/path)
		sort.Slice(paths, func(i, j int) bool {
			if paths[i].days == paths[j].days {
				return unsafeString(paths[i].path) < unsafeString(paths[j].path)
			}
			return paths[i].days < paths[j].days
		})

		n = 0
		for {
			i := indexNextDay(paths, n, indexCacheBatchSize)
			checks += i - n
			err := u.cacheQueryCheck(ctx, paths[n:i], indexes, treeDays, filename, checks, len(indexes), prestartTime)
			if err != nil {
				return nil, err
			} else if i == len(paths) {
				break
			}
			n = i
			prestartTime = time.Now()
		}
	}
	for key, v := range indexes {
		if !v.found || !v.foundTree {
			newSeries[key] = true
		}
	}
	return newSeries, nil
}

func (u *Index) parseFile(filename string, out io.Writer, outNotify chan bool) (uint64, uint64, uint64, map[string]bool, error) {
	var reader *RowBinary.Reader
	var err error
	var n uint64
	var skipped uint64
	var skippedTree uint64

	logger := zapwriter.Logger("index")
	startTime := time.Now()

	defer func() { outNotify <- true }()
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
			path: []byte(string(name)),
		}
	}

	newSeries, err := u.cacheBatchRecheck(nocacheSeries, treeDate, filename)
	if err != nil {
		return n, skipped, skippedTree, nil, err
	}

	first := true
	for _, v := range nocacheSeries {
		if v.found && v.foundTree {
			skipped++
			skippedTree++
			continue
		} else if first {
			outNotify <- true
			first = false
		}
		n++

		wb.Reset()

		reverseName := RowBinary.ReverseBytes(v.path)
		if !v.found {
			level = pathLevel(v.path)
			// Direct path with date
			wb.WriteUint16(v.days)
			wb.WriteUint32(uint32(level))
			wb.WriteBytes(v.path)
			wb.WriteUint32(version)

			// Reverse path with date
			wb.WriteUint16(v.days)
			wb.WriteUint32(uint32(level + ReverseLevelOffset))
			wb.WriteBytes(reverseName)
			wb.WriteUint32(version)
		}

		if v.foundTree {
			skippedTree++
		} else {
			// Tree
			wb.WriteUint16(treeDate)
			wb.WriteUint32(uint32(level + TreeLevelOffset))
			wb.WriteBytes(v.path)
			wb.WriteUint32(version)

			p = v.path
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
