package logreader

import (
	"bufio"
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestRead(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    []string
		wantErr error
	}{
		{
			name: "valid test",
			data: `93.180.71.3 - - [28/Feb/2025:23:59:59 +0000] "GET /health HTTP/1.1" 200 100 "-" "incident-radar-sample/1.0"`,
			want: []string{
				`93.180.71.3 - - [28/Feb/2025:23:59:59 +0000] "GET /health HTTP/1.1" 200 100 "-" "incident-radar-sample/1.0"`,
			},
		},
		{
			name:    "too long token",
			data:    strings.Repeat("a", 1024*1024),
			wantErr: bufio.ErrTooLong,
		},
		{
			name: "empty input",
			data: "",
			want: []string{},
		},
		{
			name: "multiple lines",
			data: "first\nsecond\nthird",
			want: []string{"first", "second", "third"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := make([]string, 0)

			source := strings.NewReader(tt.data)
			gotErr := Read(source, func(b []byte) {
				got = append(got, string(b))
			})

			if !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("error expected: %v, got: %v", tt.wantErr, gotErr)
			}
			if tt.wantErr == nil && !slices.Equal(got, tt.want) {
				t.Errorf("data expected: %v, got: %v", tt.want, got)
			}
		})
	}
}
