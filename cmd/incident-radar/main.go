// Package main implements the Incident Radar CLI application.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/thexolod/incident-radar/internal/logparser"
	"github.com/thexolod/incident-radar/internal/report"
	"github.com/thexolod/incident-radar/internal/source"
	"github.com/thexolod/incident-radar/internal/stats"
)

func main() {
	var path string
	flag.StringVar(&path, "path", "", "path to logs")
	flag.StringVar(&path, "p", "", "path to logs")

	var format string
	flag.StringVar(&format, "format", "json", "log format")
	flag.StringVar(&format, "f", "json", "log format")

	var output string
	flag.StringVar(&output, "output", "", "output path")
	flag.StringVar(&output, "o", "", "output path")

	from := flag.String("from", "", "start date")
	to := flag.String("to", "", "end date")

	flag.Parse()

	if output == "" {
		fmt.Fprint(os.Stderr, "output argument is missing.")
		os.Exit(2)
	}
	if _, err := os.Stat(output); !errors.Is(err, os.ErrNotExist) {
		if err == nil {
			fmt.Fprint(os.Stderr, "output file already exists")
		} else {
			fmt.Fprintf(os.Stderr, "output file stat error: %v", err)
		}
		os.Exit(2)
	}

	if path == "" {
		fmt.Fprint(os.Stderr, "path argument is missing.")
		os.Exit(2)
	}

	if format != "json" && format != "markdown" {
		fmt.Fprintf(os.Stderr, "invalid format argument: %s", format)
		os.Exit(2)
	}
	ext := filepath.Ext(output)
	if ext != ".md" && ext != ".json" {
		fmt.Fprintf(os.Stderr, "invalid extension: %s", ext)
		os.Exit(2)
	}
	if format == "markdown" && ext != ".md" || format == "json" && ext != ".json" {
		fmt.Fprintf(os.Stderr, "extension: %s mismatch with the format: %s", ext, format)
		os.Exit(2)
	}

	var fromDate, toDate time.Time
	if *from != "" {
		var err error
		fromDate, err = time.Parse("2006-01-02", *from)
		if err != nil {
			fmt.Fprintf(os.Stderr, "from date: %v", err)
			os.Exit(2)
		}
	}
	if *to != "" {
		var err error
		toDate, err = time.Parse("2006-01-02", *to)
		if err != nil {
			fmt.Fprintf(os.Stderr, "to date: %v", err)
			os.Exit(2)
		}
		if *from != "" && !fromDate.Before(toDate) {
			fmt.Fprintf(os.Stderr, "'to' date must go after 'from' date")
			os.Exit(2)
		}
	}

	var files []string
	interval := logparser.NewTimeInterval(fromDate, toDate)
	globalState := stats.NewState()

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		err := source.ReadFromURL(path, globalState, interval)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(2)
		}
		files = append(files, path)
	} else {
		fPaths, err := filepath.Glob(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "path error %v", err)
			os.Exit(2)
		}
		for _, curPath := range fPaths {
			err := source.ReadCurrentFile(curPath, globalState, interval)
			if errors.Is(err, source.ErrBadObject) {
				continue
			}
			if err != nil {
				fmt.Fprint(os.Stderr, err)
				os.Exit(2)
			}
			files = append(files, curPath)
		}
	}

	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "no suitable .log or .txt files found for path: %s", path)
		os.Exit(2)
	}

	reportToSave := stats.MakeReport(globalState, files)
	var data []byte
	var err error
	switch format {
	case "json":
		data, err = report.ToJSON(reportToSave)
	case "markdown":
		data = report.ToMarkdown(reportToSave)
	}
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	err = report.Save(output, data)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		if errors.Is(err, os.ErrExist) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
