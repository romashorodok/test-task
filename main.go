package main

import (
	"errors"
	"io"
	"log"
	"os"
	"sync"
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

		stubBuffer := make([]byte, 24)

		n, err := t.source.Read(stubBuffer[oldBufLen:])

		copy(stubBuffer[:oldBufLen], t.oldBuf[:oldBufLen])

		log.Println("after transform", stubBuffer, n, err)

		log.Println(string(stubBuffer))

		return len(stubBuffer), stubBuffer, err
	}

	n, err := t.source.Read(t.buf)
	return n, t.buf, err
}

func (t *tokenizer) NextWordLen() ([]byte, int, error) {
	t.nextWordLenMu.Lock()
	defer t.nextWordLenMu.Unlock()

	n, buf, err := t.read()
	if err != nil {
		log.Println(err)
		return nil, -1, err
	}

	var r rune
	var step int

	for offset := 0; offset < n; offset += step {
		r, step = utf8.DecodeRune(buf[offset:])
		if step == 0 {
			return nil, -1, err
		}

		if isSpace(r) {
			t.oldBuf = buf[offset+1:]
			return buf[:offset], offset, nil
		}
	}

	return nil, -1, ErrNotFoundSpace
}

func NewTokenizer(source io.Reader) *tokenizer {
	return &tokenizer{
		source: source,
		// 1 byte = uint8
		buf: make([]byte, 24),
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

	word, wordLen, err := tokenizer.NextWordLen()
	log.Println(wordLen, string(word))

	word, wordLen, err = tokenizer.NextWordLen()
	log.Println(wordLen, string(word))

	word, wordLen, err = tokenizer.NextWordLen()
	log.Println(wordLen, string(word))

	word, wordLen, err = tokenizer.NextWordLen()
	log.Println(wordLen, string(word))

	word, wordLen, err = tokenizer.NextWordLen()
	log.Println(wordLen, string(word))

	word, wordLen, err = tokenizer.NextWordLen()
	log.Println(wordLen, string(word))

	word, wordLen, err = tokenizer.NextWordLen()
	log.Println(wordLen, string(word))

	word, wordLen, err = tokenizer.NextWordLen()
	log.Println(wordLen, string(word))

	word, wordLen, err = tokenizer.NextWordLen()
	log.Println(wordLen, string(word))

	word, wordLen, err = tokenizer.NextWordLen()
	log.Println(wordLen, string(word))
}
