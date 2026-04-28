package dashboard

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/dashboard/format"
)

// CommonFlags holds the persistent + shared flag values every leaf command
// in the dashboard tree consumes. Subcommands embed it and call Bind() in
// their constructor; PreRun then calls Validate() to enforce cross-flag
// invariants before the leaf's RunE fires.
//
// Why a single struct rather than per-command flags? Because every
// `dashboard *` invocation goes through the same authentication / output /
// timezone / temperature pipeline, and we want one place that documents
// the cross-flag rules (e.g. --start/--end mutually exclusive with --since;
// --watch requires a non-zero recommended-poll-seconds).
type CommonFlags struct {
	// Output is the resolved --output / -o value. Defaults to OutputTable.
	Output OutputFormat

	// outputRaw is the raw flag string before normalisation. Cobra
	// binds onto this; Validate() turns it into Output.
	outputRaw string

	// Watch toggles the polling ticker (see watch.go). Defaults off.
	Watch bool
	// WatchInterval is the cadence between iterations. 0 means "use the
	// command's RecommendedPollSeconds". Lower values emit a stderr
	// warning but proceed.
	WatchInterval time.Duration
	// WatchIterations caps the total number of iterations. 0 = unbounded.
	WatchIterations int
	// WatchTimeout caps the total wall-clock duration. 0 = unbounded.
	WatchTimeout time.Duration

	// Since is the relative sliding window used by metric commands in
	// `--watch` mode. e.g. "1h" → request the last 1h on every
	// iteration. Mutually exclusive with Start/End.
	Since time.Duration

	// Start / End are the absolute fixed window. Same window on every
	// iteration; useful for replaying a known incident. Mutually
	// exclusive with Since.
	Start time.Time
	End   time.Time

	// startRaw / endRaw / sinceRaw / timezoneRaw store the raw flag
	// strings before parsing so Validate() can produce friendly error
	// messages.
	startRaw    string
	endRaw      string
	sinceRaw    string
	timezoneRaw string

	// Timezone, when non-nil, overrides time.Local for FormatTime /
	// Meta.FetchedAt rendering. Defaults to time.Local.
	Timezone *format.Location

	// TempUnit is the user's preferred temperature display unit (C / F /
	// K). Defaults to TempC. JSON `raw` always emits Celsius regardless;
	// this only affects table view + display.rendered fields.
	TempUnit format.TempUnit
	// tempUnitRaw is the raw flag string before normalisation.
	tempUnitRaw string

	// User, when non-empty, targets a different user than the active
	// profile (admin-only). Surfaced via Meta.User.
	User string

	// Limit / Page mirror the BFF's pagination knobs for endpoints that
	// support them (apps list, etc.). 0 means "use default" (no
	// override sent).
	Limit int
	Page  int

	// Head truncates the rendered output to the first N rows after
	// sorting (client-side). 0 = no truncation.
	Head int
}

// BindPersistent registers the persistent flags every dashboard command
// needs on `cmd`. Call this on the dashboard root command.
func (cf *CommonFlags) BindPersistent(cmd *cobra.Command) {
	pf := cmd.PersistentFlags()
	pf.StringVarP(&cf.outputRaw, "output", "o", "table",
		"output format (table or json)")
	pf.BoolVar(&cf.Watch, "watch", false,
		"poll the upstream endpoint and emit one envelope per iteration (NDJSON in JSON mode)")
	pf.DurationVar(&cf.WatchInterval, "watch-interval", 0,
		"interval between watch iterations (default: command's recommended-poll-seconds)")
	pf.IntVar(&cf.WatchIterations, "watch-iterations", 0,
		"stop after N iterations (0 = unbounded)")
	pf.DurationVar(&cf.WatchTimeout, "watch-timeout", 0,
		"stop after this much wall-clock time (0 = unbounded)")
	pf.StringVar(&cf.sinceRaw, "since", "",
		"relative window for metric commands; sliding when --watch (e.g. 5m, 1h)")
	pf.StringVar(&cf.startRaw, "start", "",
		"absolute window start (RFC3339); fixed across iterations when --watch")
	pf.StringVar(&cf.endRaw, "end", "",
		"absolute window end (RFC3339); fixed across iterations when --watch")
	pf.StringVar(&cf.timezoneRaw, "timezone", "",
		"timezone for table rendering (IANA name, default: $TZ / system local)")
	pf.StringVar(&cf.tempUnitRaw, "temp-unit", "C",
		"temperature display unit: C, F, or K (JSON raw always Celsius)")
	pf.StringVar(&cf.User, "user", "",
		"target a different user than the active profile (platform-admin only)")
	pf.IntVar(&cf.Limit, "limit", 0,
		"page size for paginated endpoints (0 = upstream default)")
	pf.IntVar(&cf.Page, "page", 0,
		"page index for paginated endpoints (0 = first)")
	pf.IntVar(&cf.Head, "head", 0,
		"truncate output to the first N rows after sorting (0 = no truncation)")
}

// Validate parses the raw flag strings into the typed fields and enforces
// cross-flag rules. Call from PreRunE, after cobra has populated the raw
// strings.
//
// Invariants enforced here (the test suite asserts each one):
//
//   - --output must be one of {table, json}
//   - --temp-unit must be one of {C, F, K}
//   - --since is mutually exclusive with --start/--end
//   - --start and --end must come together; --start must be < --end
//   - --watch-iterations / --watch-timeout require --watch
//   - negative durations are rejected
//
// Validate is idempotent — calling it twice returns the same result.
func (cf *CommonFlags) Validate() error {
	out, err := ParseOutputFormat(cf.outputRaw)
	if err != nil {
		return err
	}
	cf.Output = out

	tu, err := parseTempUnit(cf.tempUnitRaw)
	if err != nil {
		return err
	}
	cf.TempUnit = tu

	if cf.timezoneRaw != "" {
		loc, err := format.LoadLocation(cf.timezoneRaw)
		if err != nil {
			return fmt.Errorf("--timezone: %w", err)
		}
		cf.Timezone = loc
	} else if cf.Timezone == nil {
		cf.Timezone = format.LocalLocation()
	}

	if cf.sinceRaw != "" {
		d, err := time.ParseDuration(cf.sinceRaw)
		if err != nil {
			return fmt.Errorf("--since: %w", err)
		}
		if d < 0 {
			return errors.New("--since: must be non-negative")
		}
		cf.Since = d
	}
	if cf.startRaw != "" {
		t, err := time.Parse(time.RFC3339, cf.startRaw)
		if err != nil {
			return fmt.Errorf("--start: %w", err)
		}
		cf.Start = t
	}
	if cf.endRaw != "" {
		t, err := time.Parse(time.RFC3339, cf.endRaw)
		if err != nil {
			return fmt.Errorf("--end: %w", err)
		}
		cf.End = t
	}

	hasSince := cf.Since != 0
	hasStart := !cf.Start.IsZero()
	hasEnd := !cf.End.IsZero()

	if hasSince && (hasStart || hasEnd) {
		return errors.New("--since is mutually exclusive with --start/--end")
	}
	if hasStart != hasEnd {
		return errors.New("--start and --end must be specified together")
	}
	if hasStart && hasEnd && !cf.Start.Before(cf.End) {
		return errors.New("--start must be before --end")
	}

	if !cf.Watch {
		if cf.WatchIterations > 0 {
			return errors.New("--watch-iterations requires --watch")
		}
		if cf.WatchTimeout > 0 {
			return errors.New("--watch-timeout requires --watch")
		}
		if cf.WatchInterval > 0 {
			return errors.New("--watch-interval requires --watch")
		}
	}
	if cf.WatchInterval < 0 || cf.WatchTimeout < 0 || cf.WatchIterations < 0 {
		return errors.New("watch durations / iteration count must be non-negative")
	}
	if cf.Limit < 0 || cf.Page < 0 || cf.Head < 0 {
		return errors.New("--limit / --page / --head must be non-negative")
	}

	return nil
}

// HasAbsoluteWindow reports whether --start/--end were both provided.
func (cf *CommonFlags) HasAbsoluteWindow() bool {
	return !cf.Start.IsZero() && !cf.End.IsZero()
}

// ResolveWindow returns the effective [start, end] for the iteration at
// `now`. For absolute windows the same pair is returned every call; for
// sliding windows we anchor end=now and back off by Since.
//
// When neither flag is set we fall back to defaultDuration ending at `now`
// — defaultDuration is whatever the per-page config.ts uses (e.g. 5m for
// CPU). Caller passes it.
func (cf *CommonFlags) ResolveWindow(now time.Time, defaultDuration time.Duration) (start, end time.Time) {
	if cf.HasAbsoluteWindow() {
		return cf.Start, cf.End
	}
	dur := cf.Since
	if dur == 0 {
		dur = defaultDuration
	}
	return now.Add(-dur), now
}

// parseTempUnit normalises and validates the --temp-unit string.
func parseTempUnit(s string) (format.TempUnit, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "", "C", "CELSIUS":
		return format.TempC, nil
	case "F", "FAHRENHEIT":
		return format.TempF, nil
	case "K", "KELVIN":
		return format.TempK, nil
	default:
		return "", fmt.Errorf("--temp-unit: unknown unit %q (valid: C, F, K)", s)
	}
}

// ParseStep is a small helper for `--step` flags that accept either a Go
// duration ("30s") or a Prometheus-flavoured integer of seconds ("30").
// Returned as time.Duration. Used by metric commands that expose --step.
func ParseStep(raw string) (time.Duration, error) {
	if raw == "" {
		return 0, nil
	}
	if d, err := time.ParseDuration(raw); err == nil {
		if d <= 0 {
			return 0, errors.New("--step: must be positive")
		}
		return d, nil
	}
	if n, err := strconv.Atoi(raw); err == nil {
		if n <= 0 {
			return 0, errors.New("--step: must be positive")
		}
		return time.Duration(n) * time.Second, nil
	}
	return 0, fmt.Errorf("--step: %q is not a duration or integer seconds", raw)
}
