package source

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/thexolod/incident-radar/internal/logparser"
	"github.com/thexolod/incident-radar/internal/stats"
)

func TestReadCurrentFile(t *testing.T) {
	tmpInterval := logparser.NewTimeInterval(time.Time{}, time.Time{})
	tests := []struct {
		name       string
		filename   string
		data       []byte
		wantErr    error
		createFile bool
		wantReqCnt uint64
	}{
		{
			name:       "valid txt file",
			filename:   "file.txt",
			data:       []byte(`192.0.2.2 - - [01/Mar/2025:10:01:00 +0000] "GET /after-error HTTP/1.1" 500 500 "-" "curl/8.0"`),
			wantErr:    nil,
			createFile: true,
			wantReqCnt: 1,
		},
		{
			name:       "valid log file",
			filename:   "file.log",
			data:       []byte(`192.0.2.2 - - [01/Mar/2025:10:01:00 +0000] "GET /after-error HTTP/1.1" 500 500 "-" "curl/8.0"`),
			wantErr:    nil,
			createFile: true,
			wantReqCnt: 1,
		},
		{
			name:     "big txt file",
			filename: "file.txt",
			data: []byte(`192.0.2.1 - - [01/Mar/2025:12:00:00 +0000] "GET /orders HTTP/1.1" 200 100 "-" "` +
				strings.Repeat("a", 64*1024) + `"`),
			wantErr:    nil,
			createFile: true,
			wantReqCnt: 1,
		},
		{
			name:       "too big txt file",
			filename:   "file.txt",
			data:       make([]byte, 2*1024*1024),
			wantErr:    bufio.ErrTooLong,
			createFile: true,
			wantReqCnt: 0,
		},
		{
			name:       "invalid extension file",
			filename:   "csvFile.csv",
			data:       []byte(`192.0.2.2 - - [01/Mar/2025:10:01:00 +0000] "GET /after-error HTTP/1.1" 500 500 "-" "curl/8.0"`),
			wantErr:    ErrBadObject,
			createFile: true,
			wantReqCnt: 0,
		},
		{
			name:       "missed file",
			filename:   "missedTxt.txt",
			data:       []byte(`192.0.2.2 - - [01/Mar/2025:10:01:00 +0000] "GET /after-error HTTP/1.1" 500 500 "-" "curl/8.0"`),
			wantErr:    os.ErrNotExist,
			createFile: false,
			wantReqCnt: 0,
		},
		{
			name:     "three logs with invalid in the middle",
			filename: "file.txt",
			data: []byte(`192.0.2.2 - - [01/Mar/2025:10:01:00 +0000] "GET /after-error HTTP/1.1" 500 500 "-" "curl/8.0"` +
				"\ni'm log\n" + `192.0.2.2 - - [01/Mar/2025:10:01:00 +0000] "GET /after-error HTTP/1.1" 500 500 "-" "curl/8.0"`,
			),
			wantErr:    nil,
			createFile: true,
			wantReqCnt: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := t.TempDir()
			tmpState := stats.NewState()
			curPath := filepath.Join(path, tt.filename)
			if tt.createFile {
				err := os.WriteFile(curPath, tt.data, 0o600)
				if err != nil {
					t.Fatalf("write file error: %v", err)
				}
			}
			gotErr := ReadCurrentFile(curPath, tmpState, tmpInterval)
			if !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("error expected: %v, got %v", tt.wantErr, gotErr)
			}
			if tt.wantReqCnt != tmpState.GetReqCnt() {
				t.Errorf("wantReqCnt expected: %d, got %d", tt.wantReqCnt, tmpState.GetReqCnt())
			}
		})
	}
}
