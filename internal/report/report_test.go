package report

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/thexolod/incident-radar/internal/stats"
)

func TestReportToJSON(t *testing.T) {
	tests := []struct {
		name    string
		report  stats.Report
		want    string
		wantErr bool
	}{
		{
			name: "all report fields",
			report: stats.Report{
				TopResources: []stats.Pair{
					{Path: "/orders", Count: 2},
					{Path: "/missing", Count: 1},
				},
				ReqCnt:       3,
				ResponseCode: map[int]uint64{200: 2, 404: 1},
				Files:        []string{"access.log", "https://example.com/access.txt"},
				ResponseSizeInBytes: stats.ResponseSizeStats{
					AvgSize:         200.33,
					MaxResponseSize: 301,
					P95:             301,
				},
			},
			want: `{"resources":[{"path":"/orders","count":2},{"path":"/missing","count":1}],"totalRequestsCount":3,"responseCodes":{"200":2,"404":1},"files":["access.log","https://example.com/access.txt"],"responseSizeInBytes":{"average":200.33,"max":301,"p95":301}}`,
		},
		{
			name: "empty report",
			report: stats.Report{
				TopResources: []stats.Pair{},
				ResponseCode: map[int]uint64{},
				Files:        []string{},
			},
			want: `{"resources":[],"totalRequestsCount":0,"responseCodes":{},"files":[],"responseSizeInBytes":{"average":0,"max":0,"p95":0}}`,
		},
		{
			name: "invalid average",
			report: stats.Report{
				ResponseSizeInBytes: stats.ResponseSizeStats{AvgSize: math.NaN()},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, gotErr := ToJSON(&tt.report)
			if (gotErr == nil) == tt.wantErr {
				t.Fatalf("wantErr expected: %t, got: %t (error: %v)", tt.wantErr, gotErr != nil, gotErr)
			}
			if tt.wantErr {
				return
			}
			var compact bytes.Buffer
			if err := json.Compact(&compact, data); err != nil {
				t.Fatalf("JSON compact error: %v", err)
			}
			got := compact.String()
			if got != tt.want {
				t.Errorf("JSON expected: %s, got: %s", tt.want, got)
			}
		})
	}
}

func TestReportToMarkdown(t *testing.T) {
	tests := []struct {
		name   string
		report stats.Report
		want   string
	}{
		{
			name: "all report fields and code ordering",
			report: stats.Report{
				TopResources: []stats.Pair{
					{Path: "/orders", Count: 2},
					{Path: "/missing", Count: 1},
				},
				ReqCnt:       3,
				ResponseCode: map[int]uint64{404: 1, 200: 2},
				Files:        []string{"access.log", "https://example.com/access.txt"},
				ResponseSizeInBytes: stats.ResponseSizeStats{
					AvgSize:         200.33,
					MaxResponseSize: 301,
					P95:             301,
				},
			},
			want: `# Анализ NGINX-логов

| Метрика | Значение |
| --- | ---: |
| Всего запросов | 3 |
| Средний размер ответа, байт | 200.33 |
| Максимальный размер ответа, байт | 301 |
| p95, байт | 301 |

## Источники

| Файл или URL |
| --- |
| access.log |
| https://example.com/access.txt |

## Ресурсы

| Ресурс | Количество запросов |
| --- | ---: |
| /orders | 2 |
| /missing | 1 |

## Коды ответа

| Код | Количество запросов |
| ---: | ---: |
| 200 | 2 |
| 404 | 1 |
`,
		},
		{
			name: "empty report",
			want: `# Анализ NGINX-логов

| Метрика | Значение |
| --- | ---: |
| Всего запросов | 0 |
| Средний размер ответа, байт | 0.00 |
| Максимальный размер ответа, байт | 0 |
| p95, байт | 0 |

## Источники

| Файл или URL |
| --- |

## Ресурсы

| Ресурс | Количество запросов |
| --- | ---: |

## Коды ответа

| Код | Количество запросов |
| ---: | ---: |
`,
		},
		{
			name: "escaped paths",
			report: stats.Report{
				Files:        []string{"access|one.log"},
				TopResources: []stats.Pair{{Path: "/a\\b|c\nnext", Count: 1}},
			},
			want: `# Анализ NGINX-логов

| Метрика | Значение |
| --- | ---: |
| Всего запросов | 0 |
| Средний размер ответа, байт | 0.00 |
| Максимальный размер ответа, байт | 0 |
| p95, байт | 0 |

## Источники

| Файл или URL |
| --- |
| access\|one.log |

## Ресурсы

| Ресурс | Количество запросов |
| --- | ---: |
| /a\\b\|c next | 1 |

## Коды ответа

| Код | Количество запросов |
| ---: | ---: |
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(ToMarkdown(&tt.report))
			if got != tt.want {
				t.Errorf("Markdown expected:\n%s\ngot:\n%s", tt.want, got)
			}
		})
	}
}

func TestSave(t *testing.T) {
	tests := []struct {
		name       string
		filename   string
		data       string
		want       string
		createFile bool
		wantErr    error
	}{
		{
			name:     "new file",
			filename: "report.json",
			data:     "new report",
			want:     "new report",
		},
		{
			name:     "empty data",
			filename: "empty.json",
		},
		{
			name:       "existing file stays unchanged",
			filename:   "report.json",
			data:       "new report",
			want:       "original report",
			createFile: true,
			wantErr:    os.ErrExist,
		},
		{
			name:     "missing directory",
			filename: filepath.Join("missing", "report.json"),
			data:     "new report",
			wantErr:  os.ErrNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tt.filename)

			if tt.createFile {
				if err := os.WriteFile(path, []byte(tt.want), 0o600); err != nil {
					t.Fatalf("create file error: %v", err)
				}
			}

			gotErr := Save(path, []byte(tt.data))
			if !errors.Is(gotErr, tt.wantErr) {
				t.Fatalf("error expected: %v, got: %v", tt.wantErr, gotErr)
			}
			if tt.wantErr != nil && !tt.createFile {
				return
			}

			data, err := os.ReadFile(path) // #nosec G304 -- Reads a test-controlled file under t.TempDir.
			if err != nil {
				t.Fatalf("read file error: %v", err)
			}
			if got := string(data); got != tt.want {
				t.Errorf("contents expected: %q, got: %q", tt.want, got)
			}
		})
	}
}
