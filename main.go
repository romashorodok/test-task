package main

import (
	"errors"
	"io"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"unicode/utf8"
)

// math.Big
// test := atomic.Uint64{}
// _ = test.Add(1)

type tokenizer struct {
	source        io.Reader
	buf           []byte
	oldBuf        []byte
	nextWordLenMu sync.Mutex
}

func isSpace(r rune) bool {
	switch r {
	case ' ', '\n':
		return true
	default:
		return false
	}
}

var ErrNotFoundSpace = errors.New("Not found space")

func (t *tokenizer) read() (int, []byte, error) {
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

func (t *tokenizer) NextWordLen() ([]byte, int, error) {
	t.nextWordLenMu.Lock()
	defer t.nextWordLenMu.Unlock()

	var result []byte
	var wordLen int

	for {
		n, buf, err := t.read()
		if err != nil {
			t.buf = nil
			t.oldBuf = nil
			log.Println(err)
			return nil, -1, err
		}

		if n == 0 {
			break
		}

		var r rune
		var step int

		for offset := 0; offset < n; offset += step {
			r, step = utf8.DecodeRune(buf[offset:])
			if step == 0 {
				return nil, -1, err
			}

			if isSpace(r) {
				t.oldBuf = buf[offset+step:]
				return result, wordLen, nil
			}

			result = append(result, byte(r))
			wordLen += step
		}
	}

	return result, wordLen, nil
}

func NewTokenizer(source io.Reader) *tokenizer {
	return &tokenizer{
		source: source,
		// 1 byte = uint8
		// buf: make([]byte, 1024),
		buf: make([]byte, 24),
		// buf: make([]byte, 2),
	}
}

func main() {
	file, err := os.OpenFile("input_text", os.O_RDONLY, os.ModeType)
	if err != nil {
		notExistFile := errors.Is(err, os.ErrNotExist)
		log.Println(notExistFile)
	}
	defer file.Close()

	tokenizer := NewTokenizer(file)
	var counter atomic.Uint64

	for {
		word, wordLen, err := tokenizer.NextWordLen()
		if err != nil {
			break
		}
		if wordLen == 0 {
			continue
		}

		log.Println(wordLen, string(word))
		counter.Add(1)
	}
	log.Println(counter.Load())
}
