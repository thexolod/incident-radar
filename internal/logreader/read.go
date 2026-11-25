// Package logreader provides streaming line-by-line reading of log sources.
package logreader

import (
	"bufio"
	"io"
)

// Read reads the source line by line and passes each line to the handler.
func Read(source io.Reader, makeLog func([]byte)) error {
	scanner := bufio.NewScanner(source)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		data := scanner.Bytes()
		makeLog(data)
	}
	return scanner.Err()
}
