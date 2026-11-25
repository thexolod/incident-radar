package source

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/thexolod/incident-radar/internal/logparser"
	"github.com/thexolod/incident-radar/internal/stats"
)

func TestReadFromURL(t *testing.T) {
	tmpInterval := logparser.NewTimeInterval(time.Time{}, time.Time{})
	tests := []struct {
		name    string
		code    int
		wantErr bool
		close   bool
	}{
		{
			name:    "valid test",
			code:    http.StatusOK,
			wantErr: false,
			close:   false,
		},
		{
			name:    "404 status code",
			code:    http.StatusNotFound,
			wantErr: true,
			close:   false,
		},
		{
			name:    "server closed",
			code:    http.StatusOK,
			wantErr: true,
			close:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpState := stats.NewState()
			URL := serverForReadFromURL(t, tt.code, tt.close)
			gotErr := ReadFromURL(URL, tmpState, tmpInterval)
			if (gotErr == nil) == tt.wantErr {
				if tt.wantErr {
					t.Error("expected error, but didn't get it")
				} else {
					t.Errorf("didn't expected error, but get it: %v", gotErr)
				}
			}
		})
	}
}

func serverForReadFromURL(t *testing.T, code int, isClosed bool) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(code)
		if code/100 == 2 {
			_, err := w.Write([]byte(`192.0.2.10 - - [03/Mar/2025:18:30:00 +0000] "GET /api/orders HTTP/1.1" 200 600 "-" "Mozilla/5.0"`))
			if err != nil {
				t.Errorf("write to response body error: %v", err)
			}
		}
	}))
	if isClosed {
		server.Close()
		return server.URL
	}
	t.Cleanup(func() {
		server.Close()
	})
	return server.URL
}
