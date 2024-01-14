package file

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
)

type StringArrayVar []string

func (s *StringArrayVar) Set(val string) error {
	*s = append(*s, val)
	return nil
}

func (s *StringArrayVar) String() string {
	return fmt.Sprint(*s)
}

var _ flag.Value = (*StringArrayVar)(nil)

var httpUrlRegex = regexp.MustCompile(`^https?://[a-zA-Z0-9.-]+(:\d{1,5})?([/?].*)?$`)

func IsRemoteFile(path string) bool {
	return httpUrlRegex.MatchString(path)
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
