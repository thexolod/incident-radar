package main_test

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLI(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "incident-radar")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.CommandContext(t.Context(), "go", "build", "-o", binary, ".") // #nosec G204 -- Fixed go command with test-controlled build arguments.
	if logs, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build error: %v\n%s", err, logs)
	}

	const inputLogs = `192.0.2.1 - - [01/Mar/2025:10:00:00 +0000] "GET /orders HTTP/1.1" 200 100 "-" "curl"
192.0.2.2 - - [01/Mar/2025:10:01:00 +0000] "GET /missing HTTP/1.1" 404 300 "-" "curl"
`
	const oldReport = "original report"

	tests := []struct {
		name         string
		format       string
		outputName   string
		createOutput bool
		wantCode     int
	}{
		{
			name:       "JSON report",
			format:     "json",
			outputName: "report.json",
			wantCode:   0,
		},
		{
			name:       "Markdown report",
			format:     "markdown",
			outputName: "report.md",
			wantCode:   0,
		},
		{
			name:         "existing output",
			format:       "json",
			outputName:   "report.json",
			createOutput: true,
			wantCode:     2,
		},
		{
			name:       "missing output directory",
			format:     "json",
			outputName: filepath.Join("missing", "report.json"),
			wantCode:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			input := filepath.Join(dir, "access.log")
			output := filepath.Join(dir, tt.outputName)

			if err := os.WriteFile(input, []byte(inputLogs), 0o600); err != nil {
				t.Fatalf("create input error: %v", err)
			}
			if tt.createOutput {
				if err := os.WriteFile(output, []byte(oldReport), 0o600); err != nil {
					t.Fatalf("create output error: %v", err)
				}
			}

			cmd := exec.CommandContext(t.Context(), binary,
				"--path", input,
				"--format", tt.format,
				"--output", output,
			) // #nosec G204 -- Runs the test-built CLI executable from t.TempDir.
			logs, err := cmd.CombinedOutput()

			gotCode := 0
			if err != nil {
				exitErr, ok := errors.AsType[*exec.ExitError](err)
				if !ok {
					t.Fatalf("CLI start error: %v\n%s", err, logs)
				}
				gotCode = exitErr.ExitCode()
			}
			if gotCode != tt.wantCode {
				t.Fatalf("exit code expected: %d, got: %d\n%s",
					tt.wantCode, gotCode, logs)
			}

			if tt.wantCode != 0 && !tt.createOutput {
				return
			}

			data, err := os.ReadFile(output) // #nosec G304 -- Reads a test-controlled output path under t.TempDir.
			if err != nil {
				t.Fatalf("read output error: %v", err)
			}

			if tt.createOutput {
				if string(data) != oldReport {
					t.Errorf("contents expected: %q, got: %q",
						oldReport, string(data))
				}
				return
			}

			switch tt.format {
			case "json":
				if !json.Valid(data) {
					t.Errorf("valid JSON expected, got:\n%s", data)
				}
			case "markdown":
				want := "| Всего запросов | 2 |"
				if !strings.Contains(string(data), want) {
					t.Errorf("Markdown row expected: %q, got:\n%s", want, data)
				}
			}
		})
	}
}
