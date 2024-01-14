package protocol

import "io"

type File interface {
	io.ReadCloser
	GetFilename() string
}
