package uploader

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/lomik/carbon-clickhouse/helper/RowBinary"
	"github.com/lomik/zapwriter"
	"github.com/stretchr/testify/assert"
)

func TestIndex_parseName_Overflow(t *testing.T) {
	wb := RowBinary.GetWriteBuffer()
	defer wb.Release()

	logger := zapwriter.Logger("upload")
	base := &Base{
		queue:   make(chan string, 1024),
		inQueue: make(map[string]bool),
		logger:  logger,
		config:  &Config{TableName: "test"},
	}
	var sb bytes.Buffer
	for i := 0; i < 400; i++ {
		if i > 0 {
			sb.WriteString(".")
		}
		sb.WriteString(fmt.Sprintf("very_long%d", i))
	}
	u := NewIndex(base)

	name := sb.Bytes()

	reverseName := make([]byte, len(name))
	RowBinary.ReverseBytesTo(reverseName, name)

	treeDate := uint16(DefaultTreeDate)
	days := uint16(10)
	version := uint32(time.Now().Unix())
	newUniq := make(map[string]bool)

	err := u.parseName(name, reverseName, treeDate, days, version, u.config.DisableDailyIndex, newUniq, wb)
	assert.Equal(t, errBufOverflow, err)
}
