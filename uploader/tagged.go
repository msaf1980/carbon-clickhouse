package uploader

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
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

type Tagged struct {
	*cached
	opts           clickhouse.Options
	ignoredMetrics map[string]bool
}

var _ Uploader = &Tagged{}
var _ UploaderWithReset = &Tagged{}

type tagRow struct {
	days  uint16
	path  string
	name  string
	tags  []string
	found bool
}

const tagCacheBatchSize = 50000

const tagQueryS = "SELECT Path FROM %s WHERE Date = '"
const tagQueryF1 = "' AND Tag1 IN ("
const tagQueryF2 = ") AND Path IN ("
const tagQueryE = ") GROUP BY Path"

func NewTagged(base *Base) *Tagged {
	u := &Tagged{}
	u.cached = newCached(base)
	u.cached.parser = u.parseFile
	u.query = fmt.Sprintf("%s (Date, Tag1, Path, Tags, Version)", u.config.TableName)
	u.cacheQuery = fmt.Sprintf(tagQueryS, u.config.TableName)
	u.opts.ConnectTimeout = u.config.Timeout.Duration
	u.opts.Timeout = u.config.Timeout.Duration

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

// don't unescape special symbols
// escape also not needed (all is done in receiver/plain.go, Base.PlainParseLine)
func tagsParse(path string) (string, []string, error) {
	name, args, n := stringutils.Split2(path, "?")
	if n == 1 || args == "" {
		return name, nil, fmt.Errorf("incomplete tags in '%s'", path)
	}
	name = "__name__=" + name
	tags := strings.Split(args, "&")
	return name, tags, nil
}

func (u *Tagged) cacheQuerySql(names []string, paths []string, days uint16) string {
	var sb stringutils.Builder
	sb.Grow(len(paths) * 200)
	sb.WriteString(u.cacheQuery)
	sb.WriteString(daysToDate(days).Format("2006-01-02"))
	sb.WriteString(tagQueryF1)
	for n, name := range names {
		if n == 0 {
			sb.WriteRune('\'')
		} else {
			sb.WriteString(", '")
		}
		sb.WriteString(name)
		sb.WriteRune('\'')
	}
	sb.WriteString(tagQueryF2)
	for n, path := range paths {
		if n == 0 {
			sb.WriteRune('\'')
		} else {
			sb.WriteString(", '")
		}
		sb.WriteString(path)
		sb.WriteRune('\'')
	}
	sb.WriteString(tagQueryE)

	return sb.String()
}

func (u *Tagged) cacheQueryCheck(ctx context.Context, namesMap map[string]bool, paths []string, tags map[string]tagRow,
	days uint16, filename string, totalchecks, total int, prestartTime time.Time) error {

	logger := zapwriter.Logger("tags")
	startTime := time.Now()
	names := make([]string, len(namesMap))
	n := 0
	for name := range namesMap {
		names[n] = name
		n++
	}
	sql := u.cacheQuerySql(names, paths, days)
	reader, err := clickhouse.Reader(ctx, u.config.URL, sql, u.opts)
	if err != nil {
		return err
	}
	endTime := time.Now()
	var found int
	var keyerror int
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 1048576), 0)
	for scanner.Scan() {
		path := unsafeString(scanner.Bytes())
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
			s := "key '" + key + "' not found during tag lookup, may be wrong filter generated"
			logger.Error("cache", zap.String("date", daysToDate(days).Format("2006-01-02")), zap.String("error", s))
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

func (u *Tagged) cacheBatchRecheck(tags map[string]tagRow, filename string) (map[string]bool, error) {
	newTags := make(map[string]bool)
	if len(tags) == 0 {
		return newTags, nil
	}

	ctx := scope.WithRequestID(context.Background(), geRequestIDFromFilename(filename))

	var n int
	var checks int
	var days uint16
	prestartTime := time.Now()
	paths := make([]string, tagCacheBatchSize)
	names := make(map[string]bool)
	var namesSb stringutils.Builder
	namesSb.Grow(tagCacheBatchSize)
	tagsSorted := make([]tagRow, len(tags))
	for _, v := range tags {
		tagsSorted[n] = v
		n++
	}
	sort.Slice(tagsSorted, func(i, j int) bool {
		if tagsSorted[i].days == tagsSorted[j].days {
			return tagsSorted[i].path < tagsSorted[j].path
		}
		return tagsSorted[i].days < tagsSorted[j].days
	})
	n = 0
	for _, v := range tagsSorted {
		if n >= tagCacheBatchSize || (n > 0 && days != v.days) {
			checks += n
			err := u.cacheQueryCheck(ctx, names, paths[0:n], tags, days, filename, checks, len(tags), prestartTime)
			names = make(map[string]bool)
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
		paths[n] = v.path
		if _, ok := names[v.name]; !ok {
			names[string(v.name)] = true
		}
		n++
	}

	if n > 0 {
		checks += n
		err := u.cacheQueryCheck(ctx, names, paths[0:n], tags, days, filename, checks, len(tags), prestartTime)
		if err != nil {
			return nil, err
		}
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
		sname := unsafeString(name)
		key := strconv.Itoa(int(reader.Days())) + ":" + sname

		if u.existsCache.Exists(key) {
			continue LineLoop
		}

		if _, ok := nocacheSeries[key]; ok {
			continue LineLoop
		}

		sname = string(name)

		tagName, tags, err := tagsParse(sname)
		if err != nil {
			continue
		}
		nocacheSeries[key] = tagRow{
			days: reader.Days(),
			path: sname,
			name: tagName,
			tags: tags,
		}
	}

	newTagged, err := u.cacheBatchRecheck(nocacheSeries, filename)
	if err != nil {
		return n, skipped, skippedTree, nil, err
	}

	first := true
	tag1 := make([]string, 0)
	for _, v := range nocacheSeries {
		if v.found {
			skipped++
			continue
		} else if first {
			outNotify <- true
			first = false
		}
		n++

		wb.Reset()
		tagsBuf.Reset()
		tag1 = tag1[:0]

		tag1 = append(tag1, v.name)
		tagsBuf.WriteString(v.name)

		// don't upload any other tag but __name__
		// if either main metric (m.Path) or each metric (*) is ignored
		ignoreAllButName := u.ignoredMetrics[v.name] || u.ignoredMetrics["*"]
		tagsWritten := 1
		for _, t := range v.tags {
			tagsBuf.WriteString(t)
			tagsWritten++

			if !ignoreAllButName {
				tag1 = append(tag1, t)
			}
		}

		for i := 0; i < len(tag1); i++ {
			wb.WriteUint16(v.days)
			wb.WriteString(tag1[i])
			wb.WriteBytes(stringutils.UnsafeStringBytes(&v.path))
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
