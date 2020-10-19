package uploader

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/lomik/carbon-clickhouse/helper/RowBinary"
	"github.com/lomik/zapwriter"
	"github.com/msaf1980/stringutils"
	"go.uber.org/zap"
)

type Tagged struct {
	*cached
	ignoredMetrics map[string]bool
}

var _ Uploader = &Tagged{}
var _ UploaderWithReset = &Tagged{}

type tagRow struct {
	days  uint16
	name  []byte
	found bool
}

const tagCacheBatchSize = 50000
const tagQuery = "SELECT Path FROM %s WHERE Date = ? AND Path IN (?) GROUP BY Path"

func NewTagged(base *Base) *Tagged {
	u := &Tagged{}
	u.cached = newCached(base)
	u.cached.parser = u.parseFile
	u.query = fmt.Sprintf("%s (Date, Tag1, Path, Tags, Version)", u.config.TableName)
	u.cacheQuery = fmt.Sprintf(tagQuery, u.config.TableName)

	u.ignoredMetrics = make(map[string]bool, len(u.config.IgnoredTaggedMetrics))
	for _, metric := range u.config.IgnoredTaggedMetrics {
		u.ignoredMetrics[metric] = true
	}

	return u
}

func urlParse(rawurl string) (*url.URL, error) {
	p := strings.IndexByte(rawurl, '?')
	if p < 0 {
		return url.Parse(rawurl)
	}
	m, err := url.Parse(rawurl[p:])
	if m != nil {
		m.Path, err = url.PathUnescape(rawurl[:p])
		if err != nil {
			return nil, err
		}
	}
	return m, err
}

func (u *Tagged) cacheQueryCheck(connect *sql.DB, filterSb *stringutils.Builder, tags map[string]tagRow,
	days uint16, filename string, checks, totalchecks, total int, prestartTime time.Time) error {

	logger := zapwriter.Logger("tags")
	startTime := time.Now()
	date := daysToDate(days).Format("2006-01-02")
	rows, err := connect.Query(u.cacheQuery, date, filterSb.String())
	endTime := time.Now()
	if err != nil {
		logger.Debug("cache", zap.String("date", date), zap.String("filename", filename), zap.Error(err),
			zap.Duration("query_time", endTime.Sub(startTime)))
		return err
	}
	var found int
	var keyerror int
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return err
		}
		found++
		key := strconv.Itoa(int(days)) + ":" + path
		v, ok := tags[key]
		if ok {
			v.found = true
			if v.found {
				u.cached.existsCache.Add(key, startTime.Unix())
			}
			tags[key] = v
			//fmt.Println(path)
		} else {
			keyerror++
			err = fmt.Errorf("key '%s' not found during tag lookup, may be wrong filter generated", key)
			logger.Error("cache", zap.String("date", date), zap.Error(err))
		}
	}
	logger.Info("cache", zap.String("filename", filename),
		zap.Duration("query_time", endTime.Sub(startTime)), zap.Duration("time", time.Since(prestartTime)),
		zap.Int("keyerror", keyerror), zap.Int("found", found),
		zap.Int("checked", checks), zap.Int("processed", totalchecks), zap.Int("total", total))
	return nil
}

func (u *Tagged) cacheBatchRecheck(tags map[string]tagRow, filename string, filterSb *stringutils.Builder) (map[string]bool, error) {
	newTags := make(map[string]bool)
	if len(tags) == 0 {
		return newTags, nil
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
	prestartTime := time.Now()
	filterSb.Reset()
	for _, v := range tags {
		if n >= 50000 || (n > 0 && days != v.days) {
			checks += n
			err = u.cacheQueryCheck(connect, filterSb, tags, days, filename, n, checks, len(tags), prestartTime)
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
		err = u.cacheQueryCheck(connect, filterSb, tags, days, filename, n, checks, len(tags), prestartTime)
		if err != nil {
			return nil, err
		}
		filterSb.Reset()
	}

	for key, v := range tags {
		if !v.found {
			newTags[key] = true
		}
	}
	return newTags, nil
}

func (u *Tagged) parseFile(filename string, out io.Writer, outNotify chan bool) (uint64, uint64, uint64, map[string]bool, error) {
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

	nocacheSeries := make(map[string]tagRow)

	wb := RowBinary.GetWriteBuffer()
	tagsBuf := RowBinary.GetWriteBuffer()
	defer wb.Release()
	defer tagsBuf.Release()

	tag1 := make([]string, 0)

LineLoop:
	for {
		name, err := reader.ReadRecord()
		if err != nil { // io.EOF or corrupted file
			break
		}

		// skip not tagged
		if bytes.IndexByte(name, '?') < 0 {
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

		nocacheSeries[key] = tagRow{
			days: reader.Days(),
			name: []byte(string(name)),
		}
	}

	var filterSb stringutils.Builder
	if len(nocacheSeries) > 0 {
		size := tagCacheBatchSize * 200
		if len(nocacheSeries) < tagCacheBatchSize {
			size = len(nocacheSeries) * 200
		}
		if size > filterSb.Cap() {
			filterSb.Grow(size)
		}
	}
	newTagged, err := u.cacheBatchRecheck(nocacheSeries, filename, &filterSb)
	if err != nil {
		return n, skipped, skippedTree, nil, err
	}

	first := true
	for _, v := range nocacheSeries {
		if v.found {
			skipped++
			continue
		}
		n++

		wb.Reset()
		tagsBuf.Reset()
		tag1 = tag1[:0]

		m, err := urlParse(unsafeString(v.name))
		if err != nil {
			continue
		} else if first {
			outNotify <- true
			first = false
		}

		t := "__name__=" + m.Path
		tag1 = append(tag1, t)
		tagsBuf.WriteString(t)

		// don't upload any other tag but __name__
		// if either main metric (m.Path) or each metric (*) is ignored
		ignoreAllButName := u.ignoredMetrics[m.Path] || u.ignoredMetrics["*"]
		tagsWritten := 1
		for k, vl := range m.Query() {
			t := k + "=" + vl[0]
			tagsBuf.WriteString(t)
			tagsWritten++

			if !ignoreAllButName {
				tag1 = append(tag1, t)
			}
		}

		for i := 0; i < len(tag1); i++ {
			wb.WriteUint16(v.days)
			wb.WriteString(tag1[i])
			wb.WriteBytes(v.name)
			wb.WriteUVarint(uint64(tagsWritten))
			wb.Write(tagsBuf.Bytes())
			wb.WriteUint32(version)
		}

		_, err = out.Write(wb.Bytes())
		if err != nil {
			return n, skipped, skippedTree, nil, err
		}
	}

	logger.Info("cache", zap.String("filename", filename),
		zap.Duration("time", time.Since(startTime)),
		zap.Uint64("cachemiss", n), zap.Uint64("cachehit", uint64(len(nocacheSeries))-n))

	return n, skipped, skippedTree, newTagged, nil
}
