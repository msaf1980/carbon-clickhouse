package binary_store

import (
	"bufio"
	"encoding/binary"
	"os"

	"github.com/msaf1980/go-stringutils"
)

type binaryReader struct {
	file *os.File
	r    *bufio.Reader
}

func NewReader() *binaryReader {
	br := &binaryReader{}

	return br
}

func (br *binaryReader) IsOpen() bool {
	return br.file != nil
}

func (br *binaryReader) Open(fileName string) error {
	var err error
	if br.file != nil {
		return ErrNotClosed
	}
	br.file, err = os.Open(fileName)
	if err == nil {
		br.r = bufio.NewReader(br.file)
	}

	return err
}

func (br *binaryReader) Close() {
	if br.file != nil {
		br.file.Close()
		br.file = nil
	}
}

func (br *binaryReader) GetUint64() (uint64, error) {
	if br.file == nil {
		return 0, os.ErrClosed
	}
	var buf [SIZE_INT64]byte
	if n, err := br.r.Read(buf[0:SIZE_INT64]); err != nil {
		return 0, err
	} else if n != SIZE_INT64 {
		return 0, ErrUnexpectedEOR
	}

	return binary.LittleEndian.Uint64(buf[0:SIZE_INT64]), nil
}

func (br *binaryReader) Get() ([]byte, error) {
	length, err := br.GetUint64()
	if err != nil {
		return nil, err
	} else if length == 0 {
		return nil, nil
	}
	p := make([]byte, length)
	if n, err := br.r.Read(p); err != nil {
		return nil, err
	} else if uint64(n) != length {
		return nil, ErrUnexpectedEOR
	}

	return p, nil
}

func (br *binaryReader) GetString() (string, error) {
	if p, err := br.Get(); err != nil {
		return "", err
	} else {
		return stringutils.UnsafeString(p), nil
	}
}

func (br *binaryReader) GetKeyValue() (string, uint64, error) {
	key, err := br.GetString()
	if err != nil {
		return "", 0, err
	}
	if v, err := br.GetUint64(); err != nil {
		return "", 0, err
	} else {
		return key, v, nil
	}
}
