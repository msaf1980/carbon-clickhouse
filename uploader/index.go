package uploader

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/lomik/carbon-clickhouse/helper/RowBinary"
	_ "github.com/mailru/go-clickhouse"
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

const indexQueryS = "SELECT Date, Path FROM graphite_index WHERE "
const indexQueryE = " GROUP BY Path, Date ORDER BY Path, Date ASC"

type indexRow struct {
	days      uint16
	name      []byte
	sname     string
	found     bool
	foundTree bool
}

func NewIndex(base *Base) *Index {
	u := &Index{}
	u.cached = newCached(base)
	u.cached.parser = u.parseFile
	return u
}

func filterAddPath(sb *strings.Builder, ir indexRow) {
	if sb.Len() == 0 {
		sb.WriteString("'" + ir.sname + "'")
	} else {
		sb.WriteString(", '" + ir.sname + "'")
	}
}

func queryString(sb *strings.Builder, days uint16, treeDays uint16) string {
	var fsb strings.Builder
	fsb.WriteString(indexQueryS)

	t := time.Unix(86400*int64(days), 0)
	treeT := time.Unix(86400*int64(treeDays), 0)
	s := fmt.Sprintf("(Date IN ('%04d-%02d-%02d', '%04d-%02d-%02d')", treeT.Year(), treeT.Month(), treeT.Day(), t.Year(), t.Month(), t.Day())
	fsb.WriteString(s)
	fsb.WriteString(" AND Path IN (")
	fsb.WriteString(sb.String())
	fsb.WriteString("))")

	fsb.WriteString(indexQueryE)
	return fsb.String()
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

func (u *Index) cacheQueryCheck(connect *sql.DB, sb *strings.Builder, indexes map[string]indexRow, days uint16, treeDays uint16) error {
	query := fmt.Sprintf(queryString(sb, days, treeDays))
	rows, err := connect.Query(query)
	if err != nil {
		return err
	}
	var treePath string
	for rows.Next() {
		var date time.Time
		var path string
		if err := rows.Scan(&date, &path); err != nil {
			return err
		}
		d := RowBinary.SlowTimeToDays(date)
		key := fmt.Sprintf("%d:%s", d, path)
		if d == treeDays {
			treePath = path
		} else {
			if len(treePath) > 0 && treePath != path {
				err = checkTreeIndex(indexes, treePath, days)
				if err != nil {
					return err
				}
			}
			v, ok := indexes[key]
			if ok {
				v.found = true
				if treePath == path && !v.foundTree {
					v.foundTree = true
				}
				indexes[key] = v
			} else {
				return fmt.Errorf("key '%s' not found during index lookup, may be wrong filter generated", key)
			}
		}
	}
	if len(treePath) > 0 {
		err = checkTreeIndex(indexes, treePath, days)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *Index) cacheBatchRecheck(indexes map[string]indexRow, treeDays uint16) (map[string]bool, error) {
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

	n := 0
	var days uint16
	var filterSb strings.Builder
	for _, v := range indexes {
		if n > 5000 || (n > 0 && days != v.days) {
			err = u.cacheQueryCheck(connect, &filterSb, indexes, days, treeDays)
			if err != nil {
				return nil, err
			}

			filterSb.Reset()
			n = 0
		}
		if days != v.days {
			days = v.days
		}
		filterAddPath(&filterSb, v)
		n++
	}

	if n > 0 {
		err = u.cacheQueryCheck(connect, &filterSb, indexes, days, treeDays)
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

		sname := unsafeString(name)
		key := fmt.Sprintf("%d:%s", reader.Days(), sname)

		if u.existsCache.Exists(key) {
			continue LineLoop
		}

		if _, ok := nocacheSeries[key]; ok {
			continue LineLoop
		}

		nocacheSeries[key] = indexRow{
			days:  reader.Days(),
			name:  name,
			sname: sname,
		}
	}

	newSeries, err := u.cacheBatchRecheck(nocacheSeries, treeDate)
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
