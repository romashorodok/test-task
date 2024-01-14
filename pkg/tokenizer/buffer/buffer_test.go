package buffer

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"unicode/utf8"
)

func TestBuffer_Read(t *testing.T) {
	reader := io.NopCloser(bytes.NewReader([]byte("qwerty qwerty.")))

	testCases := map[string]struct {
		bufferSize   int
		source       io.ReadCloser
		expectedRune []rune
		bufferState  [][]byte
		oldBufState  []byte
	}{
		"WithSmallBufferSize": {
			bufferSize: 2,
			source:     reader,
			expectedRune: []rune{
				rune('q'),
				rune('w'),
				rune('e'),
				rune('r'),
				rune('t'),
				rune('y'),
				rune(' '),
				rune('q'),
				rune('w'),
				rune('e'),
				rune('r'),
				rune('t'),
				rune('y'),
				rune('.'),
			},
			bufferState: [][]byte{
				{byte('q'), byte('w')},
				{byte('w'), byte('e')},
				{byte('e'), byte('r')},
				{byte('r'), byte('t')},
				{byte('t'), byte('y')},
				{byte('y'), byte(' ')},
				{byte(' '), byte('q')},
				{byte('q'), byte('w')},
				{byte('w'), byte('e')},
				{byte('e'), byte('r')},
				{byte('r'), byte('t')},
				{byte('t'), byte('y')},
				{byte('y'), byte('.')},
				{byte('.')},
			},
			oldBufState: []byte{
				byte('w'),
				byte('e'),
				byte('r'),
				byte('t'),
				byte('y'),
				byte(' '),
				byte('q'),
				byte('w'),
				byte('e'),
				byte('r'),
				byte('t'),
				byte('y'),
				byte('.'),
				0,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			buffer := NewBuffer(testCase.source, testCase.bufferSize)

			var idx int
			for {
				_, buf, err := buffer.Read()
				if err != nil {
					if !errors.Is(err, io.EOF) {
						t.FailNow()
					}
					break
				}

				state := testCase.bufferState[idx]
				for i, b := range buf {
					if b != state[i] {
						t.Fatalf("Expect buffer %d got %d", b, state[i])
					}
				}

				r, step := utf8.DecodeRune(buf)
				if step == 0 {
					t.Fatal("Decoded wrong rune. Buf state:", buf)
				}

				if r != testCase.expectedRune[idx] {
					t.Fatalf("Expect %s got %s", string(testCase.expectedRune[idx]), string(r))
				}

				buffer.UseBytes(buf[step:])

				for _, alredyInMemory := range buffer.oldBuf {
					if alredyInMemory != testCase.oldBufState[idx] {
						t.Fatalf("OldBufState must be same as alredyInMemory loaded byte. Expect %s got %s at %d", string(alredyInMemory), string(testCase.oldBufState[idx]), idx)
					}
				}

				idx++
			}
		})
	}
}
