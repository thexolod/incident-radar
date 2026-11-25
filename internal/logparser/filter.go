package logparser

import "time"

// TimeInterval represents an optional time range with an inclusive start and exclusive end.
type TimeInterval struct {
	from time.Time
	to   time.Time
}

// NewTimeInterval creates an interval that includes the entire end date.
// A zero boundary means that side of the interval is unrestricted.
func NewTimeInterval(from, to time.Time) *TimeInterval {
	if !to.IsZero() {
		to = to.AddDate(0, 0, 1)
	}

	return &TimeInterval{
		from: from,
		to:   to,
	}
}

// InRange reports whether the timestamp falls within the interval.
func (ti *TimeInterval) InRange(logTime time.Time) bool {
	return (ti.from.IsZero() || !logTime.Before(ti.from)) && (ti.to.IsZero() || logTime.Before(ti.to))
}
