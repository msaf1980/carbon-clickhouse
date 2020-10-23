package uploader

import (
	"bytes"
	"io"
	"strconv"
	"time"

	"github.com/lomik/carbon-clickhouse/helper/RowBinary"
)

type Index struct {
	*cached
}

var _ Uploader = &Index{}
var _ UploaderWithReset = &Index{}

const ReverseLevelOffset uint32 = 10000
const TreeLevelOffset uint32 = 20000
const ReverseTreeLevelOffset uint32 = 30000

const DefaultTreeDate = 42 // 1970-02-12

func NewIndex(base *Base) *Index {
	u := &Index{}
	u.cached = newCached(base)
	u.cached.parser = u.parseFile
	return u
}

func (u *Index) writeMetric(wb *RowBinary.WriteBuffer, path []byte, day uint16, treeDay uint16, version uint32, newUniq map[string]bool) {
	var index int

	level := uint32(pathLevel(path))

	// Direct path with date
	wb.WriteUint16(day)
	wb.WriteUint32(level)
	wb.WriteBytes(path)
	wb.WriteUint32(version)

	reversePath := RowBinary.ReverseBytes(path)

	// Reverse path with date
	wb.WriteUint16(day)
	wb.WriteUint32(level + ReverseLevelOffset)
	wb.WriteBytes(reversePath)
	wb.WriteUint32(version)

	// Tree
	wb.WriteUint16(treeDay)
	wb.WriteUint32(level + TreeLevelOffset)
	wb.WriteBytes(path)
	wb.WriteUint32(version)

	p := path
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

func (u *Index) parseFile(filename string, out io.Writer) (uint64, uint64, uint64, map[string]bool, error) {
	var reader *RowBinary.Reader
	var err error
	var n uint64

	reader, err = RowBinary.NewReader(filename, false)
	if err != nil {
		return n, 0, 0, nil, err
	}
	defer reader.Close()

	version := uint32(time.Now().Unix())
	newSeries := make(map[string]bool)
	newUniq := make(map[string]bool)
	wb := RowBinary.GetWriteBuffer()

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

		day := reader.Days()
		key := strconv.Itoa(int(day)) + ":" + unsafeString(name)

		if u.existsCache.Exists(key) {
			continue LineLoop
		}

		if newSeries[key] {
			continue LineLoop
		}
		n++

		wb.Reset()

		newSeries[key] = true

		u.writeMetric(wb, name, day, treeDate, version, newUniq)

		_, err = out.Write(wb.Bytes())
		if err != nil {
			return n, 0, 0, nil, err
		}
	}

	wb.Release()

	return n, 0, 0, newSeries, nil
}
