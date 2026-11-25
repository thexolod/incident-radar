// Package stats accumulates request statistics and builds report data.
package stats

import "github.com/thexolod/incident-radar/internal/logparser"

// State accumulates statistics from log records.
type State struct {
	reqCnt          uint64
	responseCode    map[int]uint64
	resources       map[string]uint64
	responseSizeSum uint64
	maxResponseSize uint64
	responseSizes   []uint64
}

// NewState creates an empty state with initialized maps.
func NewState() *State {
	return &State{
		responseCode: make(map[int]uint64),
		resources:    make(map[string]uint64),
	}
}

// GetReqCnt returns the number of accumulated requests.
func (s State) GetReqCnt() uint64 {
	return s.reqCnt
}

// Add updates the statistics with one log record.
func (s *State) Add(log logparser.LogRecord) {
	s.reqCnt++
	s.responseCode[log.GetStatus()]++
	s.resources[log.GetRequest().GetResource()]++
	s.responseSizeSum += log.GetBodyBytesSent()
	s.maxResponseSize = max(s.maxResponseSize, log.GetBodyBytesSent())
	s.responseSizes = append(s.responseSizes, log.GetBodyBytesSent())
}
