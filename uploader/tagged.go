package uploader

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/lomik/carbon-clickhouse/helper/RowBinary"
	"github.com/lomik/zapwriter"
	"go.uber.org/zap"
)

type Tagged struct {
	*cached
	ignoredMetrics map[string]bool
}

var _ Uploader = &Tagged{}
var _ UploaderWithReset = &Tagged{}

const tagQueryS = "SELECT Path FROM %s WHERE "
const tagQueryE = " GROUP BY Path"

type tagRow struct {
	days  uint16
	name  []byte
	found bool
}

func NewTagged(base *Base) *Tagged {
	u := &Tagged{}
	u.cached = newCached(base)
	u.cached.parser = u.parseFile
	u.query = fmt.Sprintf("%s (Date, Tag1, Path, Tags, Version)", u.config.TableName)

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

func tagQueryString(sb *strings.Builder, days uint16, tableName string) string {
	var fsb strings.Builder
	fsb.WriteString(fmt.Sprintf(tagQueryS, tableName))

	t := time.Unix(86400*int64(days), 0)
	s := fmt.Sprintf("(Date = '%04d-%02d-%02d'", t.Year(), t.Month(), t.Day())
	fsb.WriteString(s)
	fsb.WriteString(" AND Path IN (")
	fsb.WriteString(sb.String())
	fsb.WriteString("))")

	fsb.WriteString(tagQueryE)
	return fsb.String()
}

func (u *Tagged) cacheQueryCheck(connect *sql.DB, sb *strings.Builder, tags map[string]tagRow, days uint16, filename string) error {
	logger := zapwriter.Logger("tags")
	query := tagQueryString(sb, days, u.config.TableName)
	logger.Debug("cache", zap.String("query", query), zap.String("filename", filename))
	rows, err := connect.Query(query)
	if err != nil {
		return err
	}
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return err
		}
		key := fmt.Sprintf("%d:%s", days, path)
		v, ok := tags[key]
		if ok {
			v.found = true
			tags[key] = v
			//fmt.Println(path)
		} else {
			return fmt.Errorf("key '%s' not found during tag lookup, may be wrong filter generated", key)
		}
	}

	return nil
}

func (u *Tagged) cacheBatchRecheck(tags map[string]tagRow, filename string) (map[string]bool, error) {
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

	n := 0
	var days uint16
	var filterSb strings.Builder
	for _, v := range tags {
		if n > 5000 || (n > 0 && days != v.days) {
			err = u.cacheQueryCheck(connect, &filterSb, tags, days, filename)
			if err != nil {
				return nil, err
			}

			filterSb.Reset()
			n = 0
		}
		if days != v.days {
			days = v.days
		}
		filterAddPath(&filterSb, unsafeString(v.name))
		n++
	}

	if n > 0 {
		err = u.cacheQueryCheck(connect, &filterSb, tags, days, filename)
		if err != nil {
			return nil, err
		}

		filterSb.Reset()
	}

	for key, _ := range tags {
		newTags[key] = true
	}
	return newTags, nil
}

func (u *Tagged) parseFile(filename string, out io.Writer) (uint64, uint64, uint64, map[string]bool, error) {
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

		key := fmt.Sprintf("%d:%s", reader.Days(), unsafeString(name))

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

	newTagged, err := u.cacheBatchRecheck(nocacheSeries, filename)
	if err != nil {
		return n, skipped, skippedTree, nil, err
	}

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
		}

		t := fmt.Sprintf("__name__=%s", m.Path)
		tag1 = append(tag1, t)
		tagsBuf.WriteString(t)

		// don't upload any other tag but __name__
		// if either main metric (m.Path) or each metric (*) is ignored
		ignoreAllButName := u.ignoredMetrics[m.Path] || u.ignoredMetrics["*"]
		tagsWritten := 1
		for k, vl := range m.Query() {
			t := fmt.Sprintf("%s=%s", k, vl[0])
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

	return n, skipped, skippedTree, newTagged, nil
}
