package binary_store

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/msaf1980/go-stringutils"
)

const (
	SIZE_INT64 = 8
)

var (
	ErrUnexpectedEOR = fmt.Errorf("expected end of record")
	ErrOverflow      = fmt.Errorf("overflow")
	ErrNotClosed     = fmt.Errorf("file not closed")
)

type binaryWriter struct {
	written int64
	n       int
	buf     [4194304]byte

	file *os.File
}

func NewWriter() *binaryWriter {
	bw := &binaryWriter{}

	return bw
}

func (bw *binaryWriter) IsOpen() bool {
	return bw.file != nil
}

func (bw *binaryWriter) Open(fileName string) error {
	var err error
	if bw.file != nil {
		return ErrNotClosed
	}
	bw.file, err = os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0644)
	if err == nil {
		stat, _ := bw.file.Stat()
		bw.written = stat.Size()
	}

	return nil
}

func (bw *binaryWriter) Close() {
	if bw.file != nil {
		bw.file.Close()
		bw.file = nil
		bw.written = 0
	}
}

func (bw *binaryWriter) Flush() error {
	if bw.file == nil {
		return os.ErrClosed
	}
	if bw.n > 0 {
		_, err := bw.file.Write(bw.buf[0:bw.n])
		if err == nil {
			bw.written += int64(bw.n)
			bw.n = 0
		} else {
			bw.Close()
		}

		return err
	}
	return nil
}

func (bw *binaryWriter) BufBytes() []byte {
	return bw.buf[0:bw.n]
}

func (bw *binaryWriter) Len() int {
	return bw.n
}

func (bw *binaryWriter) Size() int64 {
	return bw.written + int64(bw.n)
}

func (bw *binaryWriter) BufReset() {
	bw.n = 0
}

func (bw *binaryWriter) PutUint64(v uint64) error {
	if bw.n+SIZE_INT64 > len(bw.buf) {
		if err := bw.Flush(); err != nil {
			return err
		}
	}
	binary.LittleEndian.PutUint64(bw.buf[bw.n:], v)
	bw.n += SIZE_INT64

	return nil
}

func (bw *binaryWriter) Put(p []byte) error {
	if bw.n+SIZE_INT64+len(p) > len(bw.buf) {
		if err := bw.Flush(); err != nil {
			return err
		}
	}
	bw.PutUint64(uint64(len(p)))
	copy(bw.buf[bw.n:], p)
	bw.n += len(p)

	return nil
}

func (bw *binaryWriter) PutString(s string) error {
	if bw.n+SIZE_INT64+len(s) > len(bw.buf) {
		if err := bw.Flush(); err != nil {
			return err
		}
	}
	bw.PutUint64(uint64(len(s)))
	copy(bw.buf[bw.n:], stringutils.UnsafeStringBytes(&s))
	bw.n += len(s)

	return nil
}

func (bw *binaryWriter) PutKeyValue(key string, v uint64) error {
	if bw.n+2*SIZE_INT64+len(key) > len(bw.buf) {
		if err := bw.Flush(); err != nil {
			return err
		}
	}
	bw.PutUint64(uint64(len(key)))
	copy(bw.buf[bw.n:], stringutils.UnsafeStringBytes(&key))
	bw.n += len(key)

	bw.PutUint64(v)
	bw.n += SIZE_INT64

	return nil
}
