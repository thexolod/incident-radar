package stats

import (
	"cmp"
	"math"
	"slices"
)

// Report contains the calculated log statistics and source names.
type Report struct {
	TopResources []Pair         `json:"resources"`
	ReqCnt       uint64         `json:"totalRequestsCount"`
	ResponseCode map[int]uint64 `json:"responseCodes"`
	Files        []string       `json:"files"`

	ResponseSizeInBytes ResponseSizeStats `json:"responseSizeInBytes"`
}

// ResponseSizeStats contains the average, maximum, and p95 response sizes in bytes.
type ResponseSizeStats struct {
	AvgSize         float64 `json:"average"`
	MaxResponseSize uint64  `json:"max"`
	P95             uint64  `json:"p95"`
}

// Pair stores a resource path and its request count.
type Pair struct {
	Path  string `json:"path"`
	Count uint64 `json:"count"`
}

// MakeReport builds a report from accumulated statistics.
// It sorts the response sizes in s in place.
func MakeReport(s *State, files []string) *Report {
	avgSize := 0.
	if s.reqCnt != 0 {
		avgSize = float64(s.responseSizeSum) / float64(s.reqCnt)
	}

	var pairSlices []Pair
	for key, val := range s.resources {
		pairSlices = append(pairSlices, Pair{
			Path:  key,
			Count: val,
		})
	}
	slices.SortFunc(pairSlices, func(a, b Pair) int {
		if b.Count == a.Count {
			return cmp.Compare(a.Path, b.Path)
		}
		return cmp.Compare(b.Count, a.Count)
	})
	pairSlices = pairSlices[:min(10, len(pairSlices))]

	return &Report{
		TopResources: pairSlices,
		ReqCnt:       s.reqCnt,
		ResponseSizeInBytes: ResponseSizeStats{
			AvgSize: math.Round(avgSize*100) / 100,
			P95: func() uint64 {
				length := len(s.responseSizes)
				if length == 0 {
					return 0
				}
				slices.Sort(s.responseSizes)
				return s.responseSizes[int(math.Ceil(0.95*float64(length)))-1]
			}(),
			MaxResponseSize: s.maxResponseSize,
		},
		ResponseCode: s.responseCode,
		Files:        files,
	}
}
