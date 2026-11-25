// Package source reads logs from local files and HTTP sources.
package source

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/thexolod/incident-radar/internal/logparser"
	"github.com/thexolod/incident-radar/internal/logreader"
	"github.com/thexolod/incident-radar/internal/processor"
	"github.com/thexolod/incident-radar/internal/stats"
)

// ReadFromURL reads an HTTP log source and processes its records
// using the supplied state and time interval.
func ReadFromURL(url string, state *stats.State, interval *logparser.TimeInterval) error {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("get error: %w", err)
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			fmt.Fprintf(os.Stderr, "response body close error: %v", closeErr)
		}
	}()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("response status code: %v", resp.StatusCode)
	}
	err = logreader.Read(resp.Body, func(data []byte) {
		processor.ParserWrap(data, state, interval)
	})
	if err != nil {
		return fmt.Errorf("read file error: %w", err)
	}
	return nil
}
