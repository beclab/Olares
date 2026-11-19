package timeoutprofile

import (
	"errors"
	"fmt"
	"time"
)

var (
	// ErrUnknownClass is returned when an entrance declares a class the
	// catalog does not define.
	ErrUnknownClass = errors.New("unknown traffic class")
	// ErrOutOfBounds is returned when a declared value falls outside Bounds.
	// Values are rejected rather than clamped so the author sees the mistake.
	ErrOutOfBounds = errors.New("timeout out of bounds")
)

// Override is what an app entrance may declare. Zero fields are unset.
type Override struct {
	Idle  time.Duration
	Total time.Duration
}

// Source tells whether an effective value came from the catalog or the app.
type Source string

const (
	FromDefault  Source = "default"
	FromOverride Source = "override"
)

// Effective is the resolved timeout set of one entrance.
type Effective struct {
	Class       Class
	Idle        time.Duration
	IdleSource  Source
	Total       time.Duration
	TotalSource Source
}

// Resolve merges an entrance override into the defaults of class. An empty
// class means Interactive. A zero Total in the result means the platform
// enforces no total budget for the entrance.
func (c Catalog) Resolve(class Class, o Override) (Effective, error) {
	if class == "" {
		class = Interactive
	}
	p, ok := c.Classes[class]
	if !ok {
		return Effective{}, fmt.Errorf("%w: %q", ErrUnknownClass, class)
	}

	eff := Effective{
		Class:       class,
		Idle:        p.EdgeIdle,
		IdleSource:  FromDefault,
		TotalSource: FromDefault,
	}

	if o.Idle != 0 {
		if o.Idle < c.Bounds.IdleMin || o.Idle > c.Bounds.IdleMax {
			return Effective{}, fmt.Errorf("%w: idle %s outside [%s, %s]",
				ErrOutOfBounds, o.Idle, c.Bounds.IdleMin, c.Bounds.IdleMax)
		}
		eff.Idle = o.Idle
		eff.IdleSource = FromOverride
	}

	if o.Total != 0 {
		if o.Total < c.Bounds.TotalMin || o.Total > c.Bounds.TotalMax {
			return Effective{}, fmt.Errorf("%w: total %s outside [%s, %s]",
				ErrOutOfBounds, o.Total, c.Bounds.TotalMin, c.Bounds.TotalMax)
		}
		eff.Total = o.Total
		eff.TotalSource = FromOverride
	}

	return eff, nil
}
