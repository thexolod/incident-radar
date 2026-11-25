package stats

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/thexolod/incident-radar/internal/logparser"
)

func TestStateAdd(t *testing.T) {
	tests := []struct {
		name    string
		records []logparser.LogRecord
		want    State
	}{
		{
			name: "no records",
			want: State{
				responseCode: map[int]uint64{},
				resources:    map[string]uint64{},
			},
		},
		{
			name: "multiple records",
			records: []logparser.LogRecord{
				makeRecord(t, "/orders", 200, 100),
				makeRecord(t, "/health", 404, 300),
				makeRecord(t, "/orders", 200, 200),
			},
			want: State{
				reqCnt:          3,
				responseCode:    map[int]uint64{200: 2, 404: 1},
				resources:       map[string]uint64{"/orders": 2, "/health": 1},
				responseSizeSum: 600,
				maxResponseSize: 300,
				responseSizes:   []uint64{100, 300, 200},
			},
		},
		{
			name: "zero response size",
			records: []logparser.LogRecord{
				makeRecord(t, "/health", 304, 0),
			},
			want: State{
				reqCnt:          1,
				responseCode:    map[int]uint64{304: 1},
				resources:       map[string]uint64{"/health": 1},
				responseSizeSum: 0,
				maxResponseSize: 0,
				responseSizes:   []uint64{0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewState()
			for _, record := range tt.records {
				state.Add(record)
			}

			if !reflect.DeepEqual(*state, tt.want) {
				t.Errorf("state expected: %+v, got: %+v", tt.want, *state)
			}
		})
	}
}

func makeRecord(
	t *testing.T,
	resource string,
	status int,
	size uint64,
) logparser.LogRecord {
	t.Helper()

	line := fmt.Sprintf(
		`192.0.2.1 - - [01/Mar/2025:12:00:00 +0000] "GET %s HTTP/1.1" %d %d "-" "test"`,
		resource, status, size,
	)

	record, err := logparser.Parse(line)
	if err != nil {
		t.Fatalf("parse test log: %v", err)
	}

	return record
}
