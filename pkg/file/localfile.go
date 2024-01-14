package file

import (
	"errors"
	"os"

	"github.com/romanshorodok/test-task/pkg/protocol"
)

type LocalFile struct {
	path string
	file *os.File
}

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

var _ protocol.File = (*LocalFile)(nil)

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
