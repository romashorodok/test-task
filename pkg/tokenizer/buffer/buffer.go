package buffer

import (
	"errors"
	"io"
	"sync"
)

type Buffer struct {
	buf    []byte
	oldBuf []byte
	readMu sync.Mutex
	source io.ReadCloser
}

func (t *Buffer) Read() (int, []byte, error) {
	t.readMu.Lock()
	defer t.readMu.Unlock()

	if len(t.oldBuf) != 0 {
		oldBufLen := len(t.oldBuf)

		_ = copy(t.buf[:oldBufLen], t.oldBuf[:oldBufLen])
		t.oldBuf = nil

		n, err := t.source.Read(t.buf[oldBufLen:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				return len(t.buf[:oldBufLen+n]), t.buf[:oldBufLen+n], nil
			}

			return -1, nil, err
		}

		return oldBufLen + n, t.buf[:oldBufLen+n], err
	}

	n, err := t.source.Read(t.buf)
	return n, t.buf, err
}

func (t *Buffer) UseBytes(b []byte) {
	t.oldBuf = b
}

func (t *Buffer) Close() error {
	t.oldBuf = nil
	t.buf = nil
	return t.source.Close()
}

func NewBuffer(source io.ReadCloser, bufferSize int) *Buffer {
	return &Buffer{
		source: source,
		buf:    make([]byte, bufferSize),
	}
}
