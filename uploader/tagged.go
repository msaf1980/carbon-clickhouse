package uploader

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/lomik/carbon-clickhouse/helper/RowBinary"
	"github.com/msaf1980/stringutils"
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

func (u *Tagged) writeMetric(wb *RowBinary.WriteBuffer, tagsBuf *RowBinary.WriteBuffer,
	path string, mname string, tags []string, day uint16, version uint32) {

	tagsBuf.WriteString(mname)

	// don't upload any other tag but __name__
	// if either main metric (m.Path) or each metric (*) is ignored
	ignoreAllButName := u.ignoredMetrics[mname] || u.ignoredMetrics["*"]
	tagsWritten := 1
	for _, tag := range tags {
		tagsBuf.WriteString(tag)
	}

	if !ignoreAllButName {
		tagsWritten += len(tags)
	}

	wb.WriteUint16(day)
	wb.WriteString(mname)
	wb.WriteString(path)
	wb.WriteUVarint(uint64(tagsWritten))
	wb.Write(tagsBuf.Bytes())
	wb.WriteUint32(version)
	if !ignoreAllButName {
		for i := 0; i < len(tags); i++ {
			wb.WriteUint16(day)
			wb.WriteString(tags[i])
			wb.WriteString(path)
			wb.WriteUVarint(uint64(tagsWritten))
			wb.Write(tagsBuf.Bytes())
			wb.WriteUint32(version)
		}
	}
}

func (u *Tagged) parseFile(filename string, out io.Writer, outNotify chan bool) (uint64, uint64, uint64, map[string]bool, error) {
	var reader *RowBinary.Reader
	var err error
	var n uint64

	defer func() { outNotify <- true }()

	reader, err = RowBinary.NewReader(filename, false)
	if err != nil {
		return n, 0, 0, nil, err
	}
	defer reader.Close()

	version := uint32(time.Now().Unix())

	newTagged := make(map[string]bool)

	wb := RowBinary.GetWriteBuffer()
	tagsBuf := RowBinary.GetWriteBuffer()
	defer wb.Release()
	defer tagsBuf.Release()

	outNotify <- true
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

		day := reader.Days()
		path := unsafeString(name)
		key := strconv.Itoa(int(day)) + ":" + path

		if u.existsCache.Exists(key) {
			continue LineLoop
		}

		if newTagged[key] {
			continue LineLoop
		}
		n++

		mname, tags, err := tagsParse(path)
		if err != nil {
			continue
		}

		newTagged[key] = true

		wb.Reset()
		tagsBuf.Reset()

		u.writeMetric(wb, tagsBuf, path, mname, tags, day, version)

		_, err = out.Write(wb.Bytes())
		if err != nil {
			return n, 0, 0, nil, err
		}
	}

	return n, 0, 0, newTagged, nil
}
