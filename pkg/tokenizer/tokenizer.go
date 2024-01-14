package tokenizer

import (
	"errors"
	"io"
	"sync"
	"unicode/utf8"

	"github.com/romanshorodok/test-task/pkg/tokenizer/buffer"
)

func isSpace(r rune) bool {
	switch r {
	case ' ', '\n':
		return true
	default:
		return false
	}
}

type tokenizer struct {
	nextWordLenMu sync.Mutex
	buf           *buffer.Buffer
}

func (t *tokenizer) NextWordLen() ([]byte, int, error) {
	t.nextWordLenMu.Lock()
	defer t.nextWordLenMu.Unlock()

	var result []byte
	var wordLen int

	for {
		n, buf, err := t.buf.Read()
		if !errors.Is(err, io.EOF) && buf == nil {
			return nil, -1, err
		}
		if n == 0 {
			return nil, -1, io.EOF
		}

		var r rune
		var step int

		for offset := 0; offset < n; offset += step {

			r, step = utf8.DecodeRune(buf[offset:])
			if step == 0 {
				return nil, -1, err
			}

			if isSpace(r) {
				t.buf.UseBytes(buf[offset+step:])
				return result, wordLen, nil
			}

			result = append(result, byte(r))
			wordLen += step
		}
	}
}

func NewTokenizer(buf *buffer.Buffer) *tokenizer {
	return &tokenizer{
		buf: buf,
	}
}
