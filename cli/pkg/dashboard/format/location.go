package format

import "time"

// Location wraps a *time.Location so the format package can stay decoupled
// from time.Location's concrete API while still letting the CLI swap in the
// `--timezone` override. This is intentionally minimal — every other module
// keeps using time.Location directly.
type Location struct {
	loc *time.Location
}

// NewLocation wraps an existing *time.Location.
func NewLocation(loc *time.Location) *Location {
	if loc == nil {
		loc = time.Local
	}
	return &Location{loc: loc}
}

// LocalLocation returns a Location wrapping time.Local.
func LocalLocation() *Location {
	return &Location{loc: time.Local}
}

// LoadLocation is a thin wrapper around time.LoadLocation that returns a
// *Location instead, so callers can keep the format-package vocabulary.
// Errors are propagated verbatim — we don't wrap because the caller usually
// surfaces the message in user-facing flag-validation diagnostics.
func LoadLocation(name string) (*Location, error) {
	if name == "" {
		return LocalLocation(), nil
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, err
	}
	return &Location{loc: loc}, nil
}

// Time returns the underlying *time.Location.
func (l *Location) Time() *time.Location {
	if l == nil || l.loc == nil {
		return time.Local
	}
	return l.loc
}

// unixMilliToTime converts a unix-millisecond stamp to a time.Time in the
// wrapped location. Used by FormatTime.
func (l *Location) unixMilliToTime(ms int64) time.Time {
	return time.UnixMilli(ms).In(l.Time())
}
