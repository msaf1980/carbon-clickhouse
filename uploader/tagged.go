package uploader

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/lomik/carbon-clickhouse/helper/RowBinary"
)

type Tagged struct {
	*cached
	ignoredMetrics map[string]bool
}

var _ Uploader = &Tagged{}
var _ UploaderWithReset = &Tagged{}

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

	newTagged := make(map[string]bool)

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

		if newTagged[key] {
			continue LineLoop
		}
		n++

		m, err := urlParse(unsafeString(name))
		if err != nil {
			continue
		}

		newTagged[key] = true

		wb.Reset()
		tagsBuf.Reset()
		tag1 = tag1[:0]

		t := fmt.Sprintf("__name__=%s", m.Path)
		tag1 = append(tag1, t)
		tagsBuf.WriteString(t)

		// don't upload any other tag but __name__
		// if either main metric (m.Path) or each metric (*) is ignored
		ignoreAllButName := u.ignoredMetrics[m.Path] || u.ignoredMetrics["*"]
		tagsWritten := 1
		for k, v := range m.Query() {
			t := fmt.Sprintf("%s=%s", k, v[0])
			tagsBuf.WriteString(t)
			tagsWritten++

			if !ignoreAllButName {
				tag1 = append(tag1, t)
			}
		}

		for i := 0; i < len(tag1); i++ {
			wb.WriteUint16(reader.Days())
			wb.WriteString(tag1[i])
			wb.WriteBytes(name)
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
