package processor

import (
	"testing"
	"time"

	"github.com/thexolod/incident-radar/internal/logparser"
	"github.com/thexolod/incident-radar/internal/stats"
)

func TestParserWrap(t *testing.T) {
	const validLine = `192.0.2.10 - alice [01/Mar/2025:12:00:00 +0000] "GET /orders HTTP/1.1" 200 512 "-" "curl/8.0"`

	date := time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		lines   []string
		from    time.Time
		to      time.Time
		wantCnt uint64
	}{
		{
			name:    "valid record",
			lines:   []string{validLine},
			wantCnt: 1,
		},
		{
			name:    "malformed record",
			lines:   []string{"broken log"},
			wantCnt: 0,
		},
		{
			name:    "continues after malformed record",
			lines:   []string{"broken log", validLine},
			wantCnt: 1,
		},
		{
			name:    "before from",
			lines:   []string{validLine},
			from:    date.AddDate(0, 0, 1),
			wantCnt: 0,
		},
		{
			name:    "after to",
			lines:   []string{validLine},
			to:      date.AddDate(0, 0, -1),
			wantCnt: 0,
		},
		{
			name:    "to day included",
			lines:   []string{validLine},
			to:      date,
			wantCnt: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := stats.NewState()
			interval := logparser.NewTimeInterval(tt.from, tt.to)

			for _, line := range tt.lines {
				ParserWrap([]byte(line), state, interval)
			}

			if state.GetReqCnt() != tt.wantCnt {
				t.Errorf("reqCnt expected: %d, got: %d",
					tt.wantCnt, state.GetReqCnt())
			}
		})
	}
}
