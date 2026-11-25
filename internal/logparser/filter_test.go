package logparser

import (
	"testing"
	"time"
)

func TestTimeIntervalInRange(t *testing.T) {
	from := time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, 2)
	toExclusive := to.AddDate(0, 0, 1)

	tests := []struct {
		name    string
		from    time.Time
		to      time.Time
		logTime time.Time
		want    bool
	}{
		{
			name:    "no boundaries",
			logTime: from,
			want:    true,
		},
		{
			name:    "only from before boundary",
			from:    from,
			logTime: from.Add(-time.Second),
			want:    false,
		},
		{
			name:    "only from exact boundary",
			from:    from,
			logTime: from,
			want:    true,
		},
		{
			name:    "only from after boundary",
			from:    from,
			logTime: toExclusive,
			want:    true,
		},
		{
			name:    "only to before boundary",
			to:      to,
			logTime: from.AddDate(0, 0, -10),
			want:    true,
		},
		{
			name:    "only to last moment of day",
			to:      to,
			logTime: toExclusive.Add(-time.Nanosecond),
			want:    true,
		},
		{
			name:    "only to next day excluded",
			to:      to,
			logTime: toExclusive,
			want:    false,
		},
		{
			name:    "both before from",
			from:    from,
			to:      to,
			logTime: from.Add(-time.Second),
			want:    false,
		},
		{
			name:    "both exact from",
			from:    from,
			to:      to,
			logTime: from,
			want:    true,
		},
		{
			name:    "both inside interval",
			from:    from,
			to:      to,
			logTime: from.AddDate(0, 0, 1),
			want:    true,
		},
		{
			name:    "both last moment of to",
			from:    from,
			to:      to,
			logTime: toExclusive.Add(-time.Nanosecond),
			want:    true,
		},
		{
			name:    "both next day excluded",
			from:    from,
			to:      to,
			logTime: toExclusive,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interval := NewTimeInterval(tt.from, tt.to)
			got := interval.InRange(tt.logTime)
			if got != tt.want {
				t.Errorf("inRange expected: %t, got: %t", tt.want, got)
			}
		})
	}
}
