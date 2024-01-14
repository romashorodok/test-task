package file

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/romanshorodok/test-task/pkg/protocol"
)

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

var _ protocol.File = (*RemoteFile)(nil)

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
