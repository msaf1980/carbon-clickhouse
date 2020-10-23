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

type indexRow struct {
	day       uint16
	path      string
	found     bool
	foundTree bool
}

type Index struct {
	*cached
}

var _ Uploader = &Index{}
var _ UploaderWithReset = &Index{}

const ReverseLevelOffset uint32 = 10000
const TreeLevelOffset uint32 = 20000
const ReverseTreeLevelOffset uint32 = 30000

const DefaultTreeDate = 42 // 1970-02-12

const indexQueryS = "SELECT Path, groupUniqArray(Date) as Dates FROM %s WHERE Date IN ('"
const indexQueryF = "') AND Path IN ("
const indexQueryE = ") GROUP BY Path FORMAT RowBinary"

func NewIndex(base *Base) *Index {
	u := &Index{}
	u.cached = newCached(base)
	u.cached.parser = u.parseFile
	u.cacheQuery = fmt.Sprintf(indexQueryS, u.config.TableName)
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
		sb.Write(stringutils.UnsafeStringBytes(&v.path))
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
	treeDay uint16, filename string, totalchecks, total int, prestartTime time.Time) error {
	options := clickhouse.Options{u.config.Timeout.Duration, u.config.Timeout.Duration}

	logger := zapwriter.Logger("index")

	startTime := time.Now()
	day := paths[0].day
	sql := u.cacheQuerySql(paths, day, treeDay)
	reader, err := clickhouse.Reader(ctx, u.config.URL, sql, options)
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

		key := strconv.Itoa(int(day)) + ":" + path
		v, ok := indexes[key]
		if !ok {
			keyerror++
			s := "key '" + key + "' not found during index lookup, may be wrong filter generated"
			logger.Error("cache", zap.String("date", daysToDate(day).Format("2006-01-02")), zap.String("error", s))
			continue
		}
		found++
		for i := 0; i < datesCount; i++ {
			if dates[i] == day {
				v.found = true
			} else if dates[i] == treeDay {
				v.foundTree = true
			} else {
				s := fmt.Sprintf("day '%d' not found during index lookup, may be RowBinary parser error", dates[i])
				logger.Error("cache", zap.String("date", daysToDate(day).Format("2006-01-02")), zap.String("error", s))
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
	endTime := time.Now()
	logger.Info("cache", zap.String("filename", filename),
		zap.Duration("time", endTime.Sub(startTime)),
		zap.Duration("total_time", endTime.Sub(prestartTime)),
		zap.Int("keyerror", keyerror), zap.Int("found", found),
		zap.Int("checked", len(paths)), zap.Int("processed", totalchecks), zap.Int("total", total))
	return nil
}

func geRequestIDFromFilename(filename string, timestamp int64) string {
	if i := strings.LastIndex(filename, "."); i == -1 {
		return strconv.Itoa(int(timestamp)) + ":" + path.Base(filename)
	} else {
		return strconv.Itoa(int(timestamp)) + ":" + filename[i+1:]
	}
}

// indexNextDay get next day location on sorted slice
func indexNextDay(indexes []indexRow, start int, max int) int {
	end := len(indexes) - 1
	day := indexes[start].day
	if indexes[end].day == day {
		if len(indexes) > start+max {
			return start + max
		}
		return len(indexes)
	}

	pos := start
	for pos < end {
		median := pos + (end-pos)/2

		if indexes[median].day > day {
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

func (u *Index) cacheBatchRecheck(indexes map[string]indexRow, treeDays uint16, filename string, prestartTime time.Time) error {
	if len(indexes) == 0 {
		return nil
	}

	var n int
	var checks int
	ctx := scope.WithRequestID(context.Background(), geRequestIDFromFilename(filename, prestartTime.Unix()))
	paths := make([]indexRow, len(indexes))
	for _, v := range indexes {
		paths[n] = v
		n++
	}
	// sort for index sequentional scan (on date/path)
	sort.Slice(paths, func(i, j int) bool {
		if paths[i].day == paths[j].day {
			return paths[i].path < paths[j].path
		}
		return paths[i].day < paths[j].day
	})

	loadTime := time.Now()
	logger := zapwriter.Logger("index")
	logger.Info("cache", zap.String("filename", filename),
		zap.Duration("load_time", loadTime.Sub(prestartTime)),
	)

	n = 0
	for {
		i := indexNextDay(paths, n, u.config.BatchQuerySize)
		checks += i - n
		err := u.cacheQueryCheck(ctx, paths[n:i], indexes, treeDays, filename, checks, len(indexes), prestartTime)
		if err != nil {
			return err
		} else if i == len(paths) {
			break
		}
		n = i
		prestartTime = time.Now()
	}

	return nil
}

func (u *Index) writeMetric(wb *RowBinary.WriteBuffer, path string, day uint16, treeDay uint16, version uint32, newUniq map[string]bool) {
	var index int

	bpath := stringutils.UnsafeStringBytes(&path)
	level := uint32(pathLevel(bpath))

	// Direct path with date
	wb.WriteUint16(day)
	wb.WriteUint32(level)
	wb.WriteBytes(bpath)
	wb.WriteUint32(version)

	reversePath := RowBinary.ReverseBytes(bpath)

	// Reverse path with date
	wb.WriteUint16(day)
	wb.WriteUint32(level + ReverseLevelOffset)
	wb.WriteBytes(reversePath)
	wb.WriteUint32(version)

	// Tree
	wb.WriteUint16(treeDay)
	wb.WriteUint32(level + TreeLevelOffset)
	wb.WriteBytes(bpath)
	wb.WriteUint32(version)

	p := bpath
	l := level
	for l--; l > 0; l-- {
		index = bytes.LastIndexByte(p, '.')
		if newUniq[unsafeString(p[:index+1])] {
			break
		}

		newUniq[string(p[:index+1])] = true

		wb.WriteUint16(treeDay)
		wb.WriteUint32(l + TreeLevelOffset)
		wb.WriteBytes(p[:index+1])
		wb.WriteUint32(version)

		p = p[:index]
	}

	// Reverse path without date
	wb.WriteUint16(treeDay)
	wb.WriteUint32(level + ReverseTreeLevelOffset)
	wb.WriteBytes(reversePath)
	wb.WriteUint32(version)
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
		return n, 0, 0, nil, err
	}
	defer reader.Close()

	prestartTime := time.Now()
	version := uint32(prestartTime.Unix())
	nocacheSeries := make(map[string]indexRow)
	newSeries := make(map[string]bool)
	newUniq := make(map[string]bool)
	wb := RowBinary.GetWriteBuffer()

	treeDate := uint16(DefaultTreeDate)
	if !u.config.TreeDate.IsZero() {
		treeDate = RowBinary.TimestampToDays(uint32(u.config.TreeDate.Unix()))
	}

	outNotify <- true
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

		day := reader.Days()
		key := strconv.Itoa(int(day)) + ":" + unsafeString(name)

		if u.existsCache.Exists(key) {
			continue LineLoop
		}

		if _, ok := nocacheSeries[key]; ok {
			continue LineLoop
		}

		nocacheSeries[key] = indexRow{
			day:  reader.Days(),
			path: string(name),
		}
	}

	if u.config.BatchQuerySize > 0 {
		err := u.cacheBatchRecheck(nocacheSeries, treeDate, filename, prestartTime)
		if err != nil {
			return n, skipped, skippedTree, nil, err
		}
	}

	outNotify <- true

	for key, v := range nocacheSeries {
		if v.found && v.foundTree {
			skipped++
			skippedTree++
			continue
		}
		n++
		newSeries[key] = true

		wb.Reset()

		u.writeMetric(wb, v.path, v.day, treeDate, version, newUniq)

		_, err = out.Write(wb.Bytes())
		if err != nil {
			return n, 0, 0, nil, err
		}
	}

	logger.Info("cache", zap.String("filename", filename),
		zap.Duration("time", time.Since(startTime)),
		zap.Uint64("cachemiss", n), zap.Uint64("cachehit", uint64(len(nocacheSeries))-n))

	return n, skipped, skippedTree, newSeries, nil
}
