package logparser

import (
	"strings"
	"testing"
	"time"
)

func TestParseIt(t *testing.T) {
	const validLine = `192.0.2.10 - alice [17/May/2015:08:05:32 +0000] "GET /orders HTTP/1.1" 200 512 "-" "curl/8.0"`

	tests := []struct {
		name    string
		line    string
		want    LogRecord
		wantErr bool
	}{
		{
			name: "valid log",
			line: validLine,
			want: LogRecord{
				remoteAddr: "192.0.2.10",
				remoteUser: "alice",
				timeLocal: time.Date(
					2015, time.May, 17, 8, 5, 32, 0, time.UTC,
				),
				request: LogRequest{
					method:   "GET",
					resource: "/orders",
					protocol: "HTTP/1.1",
				},
				status:        200,
				bodyBytesSent: 512,
				httpReferer:   "-",
				httpUserAgent: "curl/8.0",
			},
			wantErr: false,
		},
		{
			name:    "empty line",
			line:    "",
			wantErr: true,
		},
		{
			name: "invalid date",
			line: strings.Replace(
				validLine, "17/May/2015", "35/May/2015", 1,
			),
			wantErr: true,
		},
		{
			name: "missing opening time bracket",
			line: strings.Replace(
				validLine, "[17/May", "17/May", 1,
			),
			wantErr: true,
		},
		{
			name: "empty time field",
			line: strings.Replace(
				validLine, "[17/May/2015:08:05:32 +0000]", "]", 1,
			),
			wantErr: true,
		},
		{
			name: "invalid status",
			line: strings.Replace(
				validLine, " 200 512 ", " abc 512 ", 1,
			),
			wantErr: true,
		},
		{
			name: "negative response size",
			line: strings.Replace(
				validLine, " 200 512 ", " 200 -1 ", 1,
			),
			wantErr: true,
		},
		{
			name: "response size overflow",
			line: strings.Replace(
				validLine, " 200 512 ",
				" 200 18446744073709551616 ", 1,
			),
			wantErr: true,
		},
		{
			name: "missing request protocol",
			line: strings.Replace(
				validLine, "GET /orders HTTP/1.1", "GET /orders", 1,
			),
			wantErr: true,
		},
		{
			name: "empty request resource",
			line: strings.Replace(
				validLine, "GET /orders HTTP/1.1", "GET  HTTP/1.1", 1,
			),
			wantErr: true,
		},
		{
			name: "missing user agent closing quote",
			line: strings.TrimSuffix(
				validLine, `"`,
			),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := Parse(tt.line)
			if (gotErr == nil) == tt.wantErr {
				if tt.wantErr {
					t.Error("expected error, but didn't get it")
				} else {
					t.Errorf("didn't expect error, but got it: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				return
			}

			if got.remoteAddr != tt.want.remoteAddr {
				t.Errorf("remoteAddr expected: %q, got: %q",
					tt.want.remoteAddr, got.remoteAddr)
			}
			if got.remoteUser != tt.want.remoteUser {
				t.Errorf("remoteUser expected: %q, got: %q",
					tt.want.remoteUser, got.remoteUser)
			}
			if !got.timeLocal.Equal(tt.want.timeLocal) {
				t.Errorf("timeLocal expected: %v, got: %v",
					tt.want.timeLocal, got.timeLocal)
			}
			if got.request != tt.want.request {
				t.Errorf("request expected: %+v, got: %+v",
					tt.want.request, got.request)
			}
			if got.status != tt.want.status {
				t.Errorf("status expected: %d, got: %d",
					tt.want.status, got.status)
			}
			if got.bodyBytesSent != tt.want.bodyBytesSent {
				t.Errorf("bodyBytesSent expected: %d, got: %d",
					tt.want.bodyBytesSent, got.bodyBytesSent)
			}
			if got.httpReferer != tt.want.httpReferer {
				t.Errorf("httpReferer expected: %q, got: %q",
					tt.want.httpReferer, got.httpReferer)
			}
			if got.httpUserAgent != tt.want.httpUserAgent {
				t.Errorf("httpUserAgent expected: %q, got: %q",
					tt.want.httpUserAgent, got.httpUserAgent)
			}
		})
	}
}
