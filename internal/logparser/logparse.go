// Package logparser parses NGINX combined log lines and provides date-range filtering.
package logparser

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	errBadLog       = errors.New("malformed log line")
	protocolPattern = regexp.MustCompile(`^HTTP/[0-9]+\.[0-9]+$`)
)

// LogRequest contains the HTTP method, resource, and protocol.
type LogRequest struct {
	method   string
	resource string
	protocol string
}

// GetResource returns the requested resource.
func (lr LogRequest) GetResource() string {
	return lr.resource
}

// LogRecord represents a parsed NGINX log entry.
type LogRecord struct {
	remoteAddr    string
	remoteUser    string
	timeLocal     time.Time
	request       LogRequest
	status        int
	bodyBytesSent uint64
	httpReferer   string
	httpUserAgent string
}

// GetTimeLocal returns the request timestamp.
func (lr LogRecord) GetTimeLocal() time.Time {
	return lr.timeLocal
}

// GetStatus returns the HTTP response status code.
func (lr LogRecord) GetStatus() int {
	return lr.status
}

// GetRequest returns the parsed request.
func (lr LogRecord) GetRequest() LogRequest {
	return lr.request
}

// GetBodyBytesSent returns the response body size in bytes.
func (lr LogRecord) GetBodyBytesSent() uint64 {
	return lr.bodyBytesSent
}

func validateLog(curReq string) (LogRequest, bool) {
	reqComponents := strings.Split(curReq, " ")
	if len(reqComponents) != 3 || reqComponents[0] == "" ||
		reqComponents[1] == "" || reqComponents[2] == "" {
		return LogRequest{}, false
	}

	if !protocolPattern.MatchString(reqComponents[2]) {
		return LogRequest{}, false
	}

	return LogRequest{
		method:   reqComponents[0],
		resource: reqComponents[1],
		protocol: reqComponents[2],
	}, true
}

// Parse parses one NGINX combined log line.
func Parse(req string) (LogRecord, error) {
	addr, req, found := strings.Cut(req, " - ")
	if !found {
		return LogRecord{}, errBadLog
	}
	if net.ParseIP(addr) == nil {
		return LogRecord{}, errBadLog
	}

	user, req, found := strings.Cut(req, " ")
	if !found || user == "" {
		return LogRecord{}, errBadLog
	}

	locTimeString, req, found := strings.Cut(req, "] ")
	if !found {
		return LogRecord{}, errBadLog
	}
	if len(locTimeString) == 0 || locTimeString[0] != '[' {
		return LogRecord{}, errBadLog
	}
	locTime, err := time.Parse("02/Jan/2006:15:04:05 -0700", locTimeString[1:])
	if err != nil {
		return LogRecord{}, fmt.Errorf("time parse: %w", err)
	}

	curReq, req, found := strings.Cut(req, "\" ")
	if !found {
		return LogRecord{}, errBadLog
	}
	if len(curReq) == 0 || curReq[0] != '"' {
		return LogRecord{}, errBadLog
	}
	curReq = curReq[1:]
	if strings.Contains(curReq, "\"") {
		return LogRecord{}, errBadLog
	}
	curLogRequest, isValid := validateLog(curReq)
	if !isValid {
		return LogRecord{}, errBadLog
	}

	status, req, found := strings.Cut(req, " ")
	if !found {
		return LogRecord{}, errBadLog
	}
	intStatus, err := strconv.Atoi(status)
	if err != nil || intStatus < 100 || intStatus > 599 {
		return LogRecord{}, errBadLog
	}

	bytes, req, found := strings.Cut(req, " ")
	if !found {
		return LogRecord{}, errBadLog
	}
	uint64Bytes, err := strconv.ParseUint(bytes, 10, 64)
	if err != nil {
		return LogRecord{}, errBadLog
	}

	ref, req, found := strings.Cut(req, "\" ")
	if !found {
		return LogRecord{}, errBadLog
	}
	if len(ref) == 0 || ref[0] != '"' {
		return LogRecord{}, errBadLog
	}
	ref = ref[1:]
	if strings.Contains(ref, "\"") {
		return LogRecord{}, errBadLog
	}

	if strings.Count(req, "\"") != 2 ||
		req[0] != '"' || req[len(req)-1] != '"' {
		return LogRecord{}, errBadLog
	}
	agent := req[1 : len(req)-1]

	return LogRecord{
		remoteAddr:    addr,
		remoteUser:    user,
		timeLocal:     locTime,
		request:       curLogRequest,
		status:        intStatus,
		bodyBytesSent: uint64Bytes,
		httpReferer:   ref,
		httpUserAgent: agent,
	}, nil
}
