package utils

import (
	"log"
	"time"

	"github.com/romanshorodok/test-task/pkg/protocol"
)

func WithFileDuration(fn func(protocol.File)) func(protocol.File) {
	start := time.Now()
	return func(file protocol.File) {
		fn(file)
		end := time.Since(start)
		log.Printf("%s:  took %s", file.GetFilename(), end)
	}
}
