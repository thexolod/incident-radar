package source

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/thexolod/incident-radar/internal/logparser"
	"github.com/thexolod/incident-radar/internal/logreader"
	"github.com/thexolod/incident-radar/internal/processor"
	"github.com/thexolod/incident-radar/internal/stats"
)

// ErrBadObject indicates an unsupported file extension or a non-regular file.
var ErrBadObject = errors.New("object does not meet the requirements")

// ReadCurrentFile reads a local log file and processes its records
// using the supplied state and time interval.
func ReadCurrentFile(curPath string, state *stats.State, interval *logparser.TimeInterval) error {
	ext := filepath.Ext(curPath)
	if ext != ".log" && ext != ".txt" {
		return ErrBadObject
	}
	stat, err := os.Stat(curPath)
	if err != nil {
		return fmt.Errorf("file stat error: %w", err)
	}
	if !stat.Mode().IsRegular() {
		return ErrBadObject
	}
	file, err := os.Open(curPath) // #nosec G304 -- CLI intentionally accepts user-provided log paths
	if err != nil {
		return fmt.Errorf("open file error: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Close file: %s error: %v\n", file.Name(), closeErr)
		}
	}()
	err = logreader.Read(file, func(data []byte) {
		processor.ParserWrap(data, state, interval)
	})
	if err != nil {
		return fmt.Errorf("read file error: %w", err)
	}
	return nil
}
