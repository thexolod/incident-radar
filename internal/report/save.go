// Package report formats log statistics as JSON or Markdown and saves reports to files.
package report

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/thexolod/incident-radar/internal/stats"
)

// ToJSON formats report data as JSON.
func ToJSON(report *stats.Report) ([]byte, error) {
	return json.MarshalIndent(report, "", "\t")
}

// ToMarkdown formats report data as Markdown.
func ToMarkdown(data *stats.Report) []byte {
	var b strings.Builder
	cell := strings.NewReplacer(
		"\\", "\\\\",
		"|", "\\|",
		"\r", " ",
		"\n", " ",
	)

	size := data.ResponseSizeInBytes
	fmt.Fprintf(&b, `# Анализ NGINX-логов

| Метрика | Значение |
| --- | ---: |
| Всего запросов | %d |
| Средний размер ответа, байт | %.2f |
| Максимальный размер ответа, байт | %d |
| p95, байт | %d |

## Источники

| Файл или URL |
| --- |
`, data.ReqCnt, size.AvgSize, size.MaxResponseSize, size.P95)

	for _, file := range data.Files {
		fmt.Fprintf(&b, "| %s |\n", cell.Replace(file))
	}
	fmt.Fprint(&b, `
## Ресурсы

| Ресурс | Количество запросов |
| --- | ---: |
`)

	for _, resource := range data.TopResources {
		fmt.Fprintf(&b, "| %s | %d |\n",
			cell.Replace(resource.Path), resource.Count)
	}
	fmt.Fprint(&b, `
## Коды ответа

| Код | Количество запросов |
| ---: | ---: |
`)

	codes := make([]int, 0, len(data.ResponseCode))
	for code := range data.ResponseCode {
		codes = append(codes, code)
	}

	slices.Sort(codes)
	for _, code := range codes {
		fmt.Fprintf(&b, "| %d | %d |\n", code, data.ResponseCode[code])
	}
	return []byte(b.String())
}

// Save writes report data to a new file without overwriting an existing file.
func Save(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_EXCL|os.O_CREATE|os.O_WRONLY, 0o600) // #nosec G304 -- CLI intentionally accepts user-provided output paths; O_EXCL prevents overwriting existing files.
	if err != nil {
		return fmt.Errorf("save: openfile error: %w", err)
	}
	_, err = file.Write(data)
	if err != nil {
		err = errors.Join(err, file.Close())
		return fmt.Errorf("save: write error: %w", err)
	}
	if closeErr := file.Close(); closeErr != nil {
		return fmt.Errorf("save: closefile error: %w", closeErr)
	}
	return nil
}
