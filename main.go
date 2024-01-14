package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
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
		if err != nil && buf == nil {
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
				t.oldBuf = buf[offset+step:]
				return result, wordLen, nil
			}

			result = append(result, byte(r))
			wordLen += step
		}
	}
}

func NewTokenizer(source io.Reader) *tokenizer {
	return &tokenizer{
		source: source,
		// 1 byte = uint8
		// buf: make([]byte, 1024),
		// buf: make([]byte, 24),
		buf: make([]byte, 2),
	}
}

type StringArrayVar []string

func (s *StringArrayVar) Set(val string) error {
	*s = append(*s, val)
	return nil
}

func (s *StringArrayVar) String() string {
	return fmt.Sprint(*s)
}

var _ flag.Value = (*StringArrayVar)(nil)

type File interface {
	io.ReadCloser
	GetFilename() string
}

type LocalFile struct {
	path string
	file *os.File
}

// GetFilename implements File.
func (f *LocalFile) GetFilename() string {
	return f.path
}

func (f *LocalFile) Close() error {
	if f.file == nil {
		return errors.New("file not exist")
	}
	return f.file.Close()
}

func (f *LocalFile) Read(p []byte) (n int, err error) {
	return f.file.Read(p)
}

var _ File = (*LocalFile)(nil)

func NewLocalFile(path string) (*LocalFile, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, os.ModeType)
	if err != nil {
		return nil, err
	}

	return &LocalFile{
		path: path,
		file: file,
	}, nil
}

type RemoteFile struct {
	path      string
	respErr   error
	content   io.ReadCloser
	contentWg sync.WaitGroup
}

func (f *RemoteFile) GetFilename() string {
	return f.path
}

func (f *RemoteFile) Close() error {
	if f.content == nil {
		return errors.New("content not exist")
	}
	return f.content.Close()
}

func (f *RemoteFile) Read(p []byte) (n int, err error) {
	f.contentWg.Wait()
	if f.respErr != nil || f.content == nil {
		return 0, f.respErr
	}
	return f.content.Read(p)
}

var _ File = (*RemoteFile)(nil)

func NewRemoteFile(path string) (*RemoteFile, error) {
	f := &RemoteFile{path: path}

	go func() {
		defer f.contentWg.Done()
		f.contentWg.Add(1)

		resp, err := http.Get(path)
		if err != nil {
			f.respErr = err
			return
		}

		if resp.StatusCode != http.StatusOK {
			// NOTE: Server may return long response message too
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				f.respErr = err
			}
			f.respErr = fmt.Errorf("%s", body)
			resp.Body.Close()
			return
		}

		f.content = resp.Body
	}()

	return f, nil
}

var httpUrlRegex = regexp.MustCompile(`^https?://[a-zA-Z0-9.-]+(:\d{1,5})?([/?].*)?$`)

func IsRemoteFile(path string) bool {
	return httpUrlRegex.MatchString(path)
}

func BatchExec[T any](vals []T, batchSize int, fn func(T)) {
	var wg sync.WaitGroup
	workerPool := make(chan struct{}, batchSize)
	defer close(workerPool)
	for i := 0; i < batchSize; i++ {
		workerPool <- struct{}{}
	}

	idxCh := make(chan int, batchSize)
	defer close(idxCh)

	go func() {
		for {
			select {
			case idx := <-idxCh:
				_, _ = <-workerPool
				go func(idx int) {
					defer func() {
						workerPool <- struct{}{}
						wg.Done()
					}()
					fn(vals[idx])
				}(idx)
			}
		}
	}()

	for idx := range vals {
		wg.Add(1)
		idxCh <- idx
	}

	wg.Wait()
}

func GetFileType(path string) string {
	cmd := exec.Command("file", path)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("Unable read stdout")
	}
	defer stdout.Close()

	var result []byte
	go func() {
		result, _ = io.ReadAll(stdout)
	}()

	if err = cmd.Run(); err != nil {
		log.Println(err)
		return ""
	}

	return string(result)
}

type AnalyticsConsumer struct {
	state map[string]*atomic.Uint64

	doneCh chan struct{}
}

func counterInit() *atomic.Uint64 {
	var counter atomic.Uint64
	counter.Add(1)
	return &counter
}

const LARGER_THEN_9_KEY = ">9"

func (a *AnalyticsConsumer) Consume(event chan int) {
	for wordLen := range event {
		if wordLen <= 9 && wordLen > 0 {
			key := fmt.Sprint(wordLen)
			counter, exist := a.state[key]
			if !exist {
				a.state[key] = counterInit()
				continue
			}
			counter.Add(1)
		}

		if wordLen > 9 {
			counter, exist := a.state[LARGER_THEN_9_KEY]
			if !exist {
				a.state[LARGER_THEN_9_KEY] = counterInit()
				continue
			}
			counter.Add(1)
		}

		if wordLen <= 0 {
			log.Println("Consumed zero or negative value.")
			continue
		}
	}
	close(a.doneCh)
}

func (a *AnalyticsConsumer) ShowAnalytics(filename string) {
	result := fmt.Sprintf("%s:  ", filename)

	for i := 0; i <= 9; i++ {
		counter, exists := a.state[fmt.Sprint(i)]
		if !exists {
			continue
		}
		result += fmt.Sprintf("[%s] = %s, ", fmt.Sprint(i), strconv.FormatUint(counter.Load(), 10))
	}

	counter, exists := a.state[LARGER_THEN_9_KEY]
	if exists {
		result += fmt.Sprintf("[%s] = %s", LARGER_THEN_9_KEY, strconv.FormatUint(counter.Load(), 10))
	} else {
		result = strings.TrimRight(result, ", ")
	}

	log.Println(result)
}

func (a *AnalyticsConsumer) Done() chan struct{} {
	return a.doneCh
}

func NewAnalyticsConsumer() *AnalyticsConsumer {
	return &AnalyticsConsumer{
		state:  make(map[string]*atomic.Uint64),
		doneCh: make(chan struct{}),
	}
}

func main() {
	var files StringArrayVar
	flag.Var(&files, "file", "Select a file to process")
	flag.Parse()

	var readers []File

	for _, path := range files {
		if IsRemoteFile(path) {
			file, err := NewRemoteFile(path)
			if err != nil {
				log.Printf("Skip file `%s` remote file. Err: %s", path, err)
				continue
			}
			readers = append(readers, file)
			defer file.Close()
			continue
		}

		if fileType := GetFileType(path); strings.Contains(fileType, "executable") {
			log.Printf("Skip file `%s` because it's executable!", path)
			continue
		}

		file, err := NewLocalFile(path)
		if err != nil {
			log.Printf("Skip file `%s` local file. Err: %s", path, err)
			continue
		}

		readers = append(readers, file)
		defer file.Close()
	}

	BatchExec(readers, 1, func(file File) {
		defer file.Close()
		tokenizer := NewTokenizer(file)
		eventBus := make(chan int, 1)

		// The data analytics may be different types of services/storages/brokers
		// Also, the data may be in different shapes/formats
		analytics := NewAnalyticsConsumer()
		// Also, good practice here to use context for canceling the process
		go analytics.Consume(eventBus)

		for {
			_, wordLen, err := tokenizer.NextWordLen()
			if err != nil {
				break
			}

			if wordLen == 0 {
				continue
			}

			eventBus <- wordLen
		}

		close(eventBus)

		select {
		case <-analytics.Done():
			analytics.ShowAnalytics(file.GetFilename())
		}
	})
}
