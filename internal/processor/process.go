// Package processor parses log lines, filters records, and updates statistics.
package processor

import (
	"fmt"
	"os"

	"github.com/thexolod/incident-radar/internal/logparser"
	"github.com/thexolod/incident-radar/internal/stats"
)

// ParserWrap parses a line and adds it to state if it matches the interval.
// Malformed lines are skipped with a warning written to stderr.
func ParserWrap(data []byte, state *stats.State, interval *logparser.TimeInterval) {
	logRecord, err := logparser.Parse(string(data))
	if err != nil {
		fmt.Fprintln(os.Stderr, "WARN: skipped malformed log line")
		return
	}
	if interval.InRange(logRecord.GetTimeLocal()) {
		state.Add(logRecord)
	}
}
