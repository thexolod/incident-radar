package stats

import (
	"fmt"
	"reflect"
	"testing"
)

func TestMakeReport(t *testing.T) {
	manyResources := make(map[string]uint64)
	var wantTop []Pair
	var manySizes []uint64

	for i := range 12 {
		path := fmt.Sprintf("/%02d", i)
		manyResources[path] = 1
		manySizes = append(manySizes, 1)
		if i < 10 {
			wantTop = append(wantTop, Pair{Path: path, Count: 1})
		}
	}

	tests := []struct {
		name  string
		state State
		files []string
		want  Report
	}{
		{
			name:  "empty state",
			state: State{},
			want:  Report{},
		},
		{
			name: "statistics and resource ordering",
			state: State{
				reqCnt: 5,
				responseCode: map[int]uint64{
					200: 3,
					404: 2,
				},
				resources: map[string]uint64{
					"/b": 2,
					"/a": 2,
					"/c": 1,
				},
				responseSizeSum: 1500,
				maxResponseSize: 500,
				responseSizes:   []uint64{500, 100, 400, 200, 300},
			},
			files: []string{"access.log", "https://example.com/access.txt"},
			want: Report{
				TopResources: []Pair{
					{Path: "/a", Count: 2},
					{Path: "/b", Count: 2},
					{Path: "/c", Count: 1},
				},
				ReqCnt:       5,
				ResponseCode: map[int]uint64{200: 3, 404: 2},
				Files: []string{
					"access.log",
					"https://example.com/access.txt",
				},
				ResponseSizeInBytes: ResponseSizeStats{
					AvgSize:         300,
					MaxResponseSize: 500,
					P95:             500,
				},
			},
		},
		{
			name: "average rounding",
			state: State{
				reqCnt:          3,
				responseCode:    map[int]uint64{200: 3},
				resources:       map[string]uint64{"/a": 3},
				responseSizeSum: 11,
				maxResponseSize: 8,
				responseSizes:   []uint64{1, 8, 2},
			},
			want: Report{
				TopResources: []Pair{{Path: "/a", Count: 3}},
				ResponseCode: map[int]uint64{200: 3},
				ReqCnt:       3,
				ResponseSizeInBytes: ResponseSizeStats{
					AvgSize:         3.67,
					MaxResponseSize: 8,
					P95:             8,
				},
			},
		},
		{
			name: "top ten limit with equal counts",
			state: State{
				reqCnt:          12,
				responseCode:    map[int]uint64{200: 12},
				resources:       manyResources,
				responseSizeSum: 12,
				maxResponseSize: 1,
				responseSizes:   manySizes,
			},
			want: Report{
				TopResources: wantTop,
				ReqCnt:       12,
				ResponseCode: map[int]uint64{200: 12},
				ResponseSizeInBytes: ResponseSizeStats{
					AvgSize:         1,
					MaxResponseSize: 1,
					P95:             1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MakeReport(&tt.state, tt.files)
			if !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("report expected: %+v, got: %+v", tt.want, *got)
			}
		})
	}
}
