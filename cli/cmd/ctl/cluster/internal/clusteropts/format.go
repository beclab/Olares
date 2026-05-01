package clusteropts

import (
	"fmt"
	"time"
)

// DashIfEmpty renders empty strings as "-" so columnar output keeps
// constant cell shape when an upstream field is absent / null. Used
// by every list-style renderer in the cluster tree.
func DashIfEmpty(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// Age renders a "kubectl-style" coarse age for an RFC3339 timestamp:
// seconds → minutes → hours → days. Used by every list/get renderer
// that prints a CreationTimestamp-derived AGE column. Returns "-"
// for empty / unparseable input so renderers can call Age unguarded.
//
// `now` is a parameter (not time.Now()) so tests can pin the clock.
func Age(ts string, now time.Time) string {
	if ts == "" {
		return "-"
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return "-"
	}
	d := now.Sub(t)
	if d < 0 {
		d = 0
	}
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}
