package main

import (
	"errors"
	"flag"
	"io"
	"log"
	"strings"
	"time"

	"github.com/romanshorodok/test-task/pkg/analytic"
	"github.com/romanshorodok/test-task/pkg/file"
	"github.com/romanshorodok/test-task/pkg/protocol"
	"github.com/romanshorodok/test-task/pkg/tokenizer"
	"github.com/romanshorodok/test-task/pkg/tokenizer/buffer"
	"github.com/romanshorodok/test-task/pkg/utils"
)

func processFiles(filesArg file.StringArrayVar) []protocol.File {
	var result []protocol.File

	for _, path := range filesArg {
		if file.IsRemoteFile(path) {
			f, err := file.NewRemoteFile(path)
			if err != nil {
				log.Printf("Skip file `%s` remote file. Err: %s", path, err)
				continue
			}
			result = append(result, f)
			continue
		}

		if fileType := file.GetFileType(path); strings.Contains(fileType, "executable") {
			log.Printf("Skip file `%s` because it's executable!", path)
			continue
		}

		f, err := file.NewLocalFile(path)
		if err != nil {
			log.Printf("Skip file `%s` local file. Err: %s", path, err)
			continue
		}
		result = append(result, f)
	}

	return result
}

func main() {
	var filesArg file.StringArrayVar
	flag.Var(&filesArg, "file", "Select a file to process")
	flag.Parse()

	files := processFiles(filesArg)

	start := time.Now()

	utils.BatchExec(files, 2, utils.WithFileDuration(
		func(f protocol.File) {
			defer f.Close()

			buf := buffer.NewBuffer(f, 1024)
			defer buf.Close()

			tokenzr := tokenizer.NewTokenizer(buf)
			eventBus := make(chan int, 1)

			// The data analytics may be different types of services/storages/brokers
			// Also, the data may be in different shapes/formats
			analytics := analytic.NewAnalyticsConsumer()
			go analytics.Consume(eventBus)

			var topLevelError error
			for {
				_, wordLen, err := tokenzr.NextWordLen()
				if err != nil {
					if !errors.Is(err, io.EOF) {
						topLevelError = err
					}
					break
				}
				if wordLen == 0 {
					continue
				}

				eventBus <- wordLen
			}

			close(eventBus)

			if topLevelError != nil {
				log.Println(topLevelError)
				return
			}

			select {
			case <-analytics.Done():
				analytics.ShowAnalytics(f.GetFilename())
			}
		},
	))

	end := time.Since(start)
	log.Printf("Process: took %s", end)
}
