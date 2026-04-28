package dashboard

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"
)

// OutputFormat is the on-the-wire choice for `--output / -o`. Default is
// `table` (human-readable, fixed columns) — agents pass `json` for the
// canonical envelope defined below.
type OutputFormat string

const (
	OutputTable OutputFormat = "table"
	OutputJSON  OutputFormat = "json"
)

// ValidOutputFormats returns the values cobra uses for ValidArgs / shell
// completion of the `--output` flag.
func ValidOutputFormats() []string {
	return []string{string(OutputTable), string(OutputJSON)}
}

// ParseOutputFormat normalises and validates a user-supplied `--output`
// value. Empty defaults to `table` (matches the SPA's "human view" default
// and what `olares-cli files ls` already does).
func ParseOutputFormat(s string) (OutputFormat, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "table":
		return OutputTable, nil
	case "json":
		return OutputJSON, nil
	default:
		return "", fmt.Errorf("unknown output format %q (valid: table, json)", s)
	}
}

// ----------------------------------------------------------------------------
// Envelope — the dual-shape JSON the CLI emits.
// ----------------------------------------------------------------------------
//
// Two shapes are supported:
//
//	Leaf      — a homogeneous list of entries:
//	            { kind, meta, items: [ { raw, display } ] }
//
//	Aggregated — heterogeneous named sections (used by `dashboard overview`):
//	            { kind, meta, sections: { <key>: { kind, meta, items|... } } }
//
// Every leaf command emits exactly one Envelope. An aggregated command emits
// one Envelope whose Sections map carries one nested Envelope per section.
// Both shapes share the same Meta layout so a generic JSON consumer can
// branch on the presence of `items` vs `sections`.

// Envelope is the top-level JSON document a command renders. Exactly one of
// Items or Sections is populated; the other is left zero / nil so its key
// is suppressed by `omitempty`.
type Envelope struct {
	// Kind is one of the constants in schema.go. Required.
	Kind string `json:"kind"`

	// Meta carries non-payload context: timestamps, the active profile,
	// pagination hints, polling cadence, and (for failed `--watch`
	// iterations) the typed error message.
	Meta Meta `json:"meta"`

	// Items is the leaf shape: a flat list of records the command
	// produced. Each Item carries a stable machine-friendly `raw` and an
	// SPA-aligned `display`.
	Items []Item `json:"items,omitempty"`

	// Sections is the aggregated shape: a map of section-key → nested
	// Envelope. Only `dashboard overview` (default action) uses this.
	// Section keys MUST be stable: `identity`, `quota`, `cluster`,
	// `ranking`. Iteration order is fixed by the parent command (encoded
	// JSON keeps insertion order via json.Marshaler-ish wrapping).
	Sections map[string]Envelope `json:"sections,omitempty"`
}

// Item is one row of a leaf Envelope.
//
// The split between Raw and Display is deliberate:
//
//   - Raw carries machine-friendly canonical values (numbers as numbers,
//     timestamps as Unix-seconds float64, no thousand separators, no
//     temperature unit conversion). Agents read this.
//
//   - Display mirrors the SPA's rendered strings (units appended, percentages
//     formatted, `--temp-unit` honored). Humans read the table view, which
//     pulls columns out of Display.
type Item struct {
	Raw     map[string]any `json:"raw,omitempty"`
	Display map[string]any `json:"display,omitempty"`
}

// Meta is the context block attached to every Envelope (top-level and per-
// section). Optional fields are suppressed when zero so the JSON stays terse
// for one-shot leaf commands and richer for multi-iteration `--watch` runs.
type Meta struct {
	// FetchedAt is the wall-clock time at which the CLI received the
	// terminal HTTP response. RFC3339 with timezone, honoring the
	// `--timezone` override.
	FetchedAt string `json:"fetched_at"`

	// Profile is the OlaresID of the active profile (whichever
	// `--profile` resolved to). Surfaced for log auditing — agents
	// should NOT use it for routing; routing is implicit in the URL.
	Profile string `json:"profile,omitempty"`

	// User, when present, is the per-command target user a `--user`
	// override resolved to. Empty for self-targeting commands.
	User string `json:"user,omitempty"`

	// RecommendedPollSeconds is the cadence the SPA polls this endpoint
	// at. The watch loop refuses `--watch` against commands with 0 here
	// (one-shot commands like `applications users` etc.).
	RecommendedPollSeconds int `json:"recommended_poll_seconds,omitempty"`

	// Iteration / TotalIterations are populated only by `--watch`. The
	// first iteration is 1, not 0, to mirror the way humans count.
	Iteration       int `json:"iteration,omitempty"`
	TotalIterations int `json:"total_iterations,omitempty"`

	// Error, when non-empty, signals this iteration / section failed but
	// the surrounding stream / aggregate continued. The CLI exits non-
	// zero only when the whole command failed; per-iteration / per-
	// section failures keep the stream alive (NDJSON discipline) so
	// agents can post-hoc detect outages.
	Error string `json:"error,omitempty"`

	// ErrorKind classifies Error into a small enum so agents can branch
	// without parsing free-form text. Values: "timeout", "http_4xx",
	// "http_5xx", "transport", "decode", "auth", "unknown".
	ErrorKind string `json:"error_kind,omitempty"`

	// Empty signals that the upstream returned no data — distinct from
	// "data was loaded and happens to be []". Used by GPU and other
	// optional-hardware endpoints; lets agents distinguish "feature not
	// installed" from "no items match".
	Empty bool `json:"empty,omitempty"`

	// EmptyReason is the human-friendly cause of Empty. Common values:
	//   "no_vgpu_integration"   — HAMI integration not installed (HTTP 404)
	//   "vgpu_unavailable"      — HAMI installed but unhealthy (HTTP 5xx);
	//                             Meta.Error carries the upstream message,
	//                             Meta.HTTPStatus carries the original status
	//   "no_gpu_detected"       — HAMI installed and healthy but the
	//                             list / detail returned an empty payload
	//   "no_pods" / "no_users"  — query had no matches
	//   "not_olares_one"        — fan / cooling features need Olares One hardware
	//   "no_fan_integration"    — capi /system/fan absent on this BFF
	//
	// GPU subtree NEVER hard-blocks on admin role or CUDA labels — those
	// surface as advisory `Note` instead, mirroring the SPA which only
	// hides the sidebar card. Reasons "requires_admin" / "no_cuda_node"
	// are reserved for soft hints; current code path emits them via Note.
	EmptyReason string `json:"empty_reason,omitempty"`

	// Note is a free-form, single-sentence explanation that complements
	// EmptyReason for human readers. JSON consumers should branch on
	// EmptyReason; Note exists so a `--watch` NDJSON stream stays self-
	// describing without an agent having to memorise the reason enum.
	Note string `json:"note,omitempty"`

	// DeviceName mirrors the `device_name` field of
	// `/user-service/api/system/status` — populated by gates that depend
	// on the Olares One vs. generic-box distinction so agents can branch
	// on hardware profile without re-querying.
	DeviceName string `json:"device_name,omitempty"`

	// HTTPStatus is the upstream HTTP status when it's worth keeping
	// (mostly the empty-by-404 cases). Suppressed for 200s.
	HTTPStatus int `json:"http_status,omitempty"`

	// Window describes the time range used to build this envelope, when
	// applicable. Populated by the GPU detail-full / task-detail-full
	// commands so agents can replay the same Prom-style range query
	// without re-deriving start/end/step.
	Window *TimeWindow `json:"window,omitempty"`

	// Warnings collects per-section / per-query soft failures that did
	// NOT abort the command. Typical use: in detail-full, one of N
	// gauges hit a 5xx — its raw entry carries `error` and the parent
	// envelope's Warnings gets an entry like
	// `gauges[2] (gpu_utilization): HAMI returned HTTP 502`. Agents
	// branching on partial data should check len(Warnings) first.
	Warnings []string `json:"warnings,omitempty"`
}

// TimeWindow describes a relative + absolute time range. All fields are
// strings so JSON consumers don't have to deal with timezone parsing —
// `since` is the user-supplied "1h"/"8h" form (or "" when --start/--end
// drove the window), `start`/`end` are RFC-3339 wall-clock, `step` is
// the SPA-style coarse-grain step ("30m" / "10m" / etc.).
type TimeWindow struct {
	Since string `json:"since,omitempty"`
	Start string `json:"start"`
	End   string `json:"end"`
	Step  string `json:"step,omitempty"`
}

// NewMeta returns a Meta pre-populated with FetchedAt and the optional
// profile / user fields. Other fields stay zero so json `omitempty` keeps
// the envelope terse.
func NewMeta(now time.Time, profile, user string) Meta {
	return Meta{
		FetchedAt: now.Format(time.RFC3339),
		Profile:   profile,
		User:      user,
	}
}

// ----------------------------------------------------------------------------
// Renderers
// ----------------------------------------------------------------------------

// WriteJSON marshals env as a single-line JSON document terminated by `\n`.
// Used for both one-shot output and individual NDJSON lines in `--watch`
// mode (the iteration / Error fields handle the per-line state).
func WriteJSON(w io.Writer, env Envelope) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(env)
}

// TableColumn names a table column and how to extract its value. Each leaf
// command supplies its own []TableColumn so the renderer stays agnostic.
type TableColumn struct {
	Header string
	// Get pulls the cell value out of an Item — typically by reading
	// item.Display[<key>]. Should never return nil; render "-" instead.
	Get func(Item) string
}

// WriteTable emits a tabwriter-based table for items. Empty inputs emit a
// single "-" row so the user has a visible signal that the call succeeded
// but produced nothing.
//
// Header / footer matching the SPA aesthetic (two-space gutter, no border
// chars) is intentional — bash piping (`| awk '{print $2}'`) stays simple.
func WriteTable(w io.Writer, columns []TableColumn, items []Item) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	headers := make([]string, len(columns))
	for i, c := range columns {
		headers[i] = c.Header
	}
	if _, err := fmt.Fprintln(tw, strings.Join(headers, "\t")); err != nil {
		return err
	}

	if len(items) == 0 {
		dash := make([]string, len(columns))
		for i := range dash {
			dash[i] = "-"
		}
		if _, err := fmt.Fprintln(tw, strings.Join(dash, "\t")); err != nil {
			return err
		}
		return tw.Flush()
	}

	for _, it := range items {
		row := make([]string, len(columns))
		for i, c := range columns {
			row[i] = c.Get(it)
		}
		if _, err := fmt.Fprintln(tw, strings.Join(row, "\t")); err != nil {
			return err
		}
	}
	return tw.Flush()
}

// DisplayString is a small helper for table column getters: pull `key` out
// of item.Display and stringify it, falling back to "-" when missing or
// empty. Centralises the rendering of `nil` / "" / 0-length so callers stay
// declarative.
func DisplayString(it Item, key string) string {
	if it.Display == nil {
		return "-"
	}
	v, ok := it.Display[key]
	if !ok || v == nil {
		return "-"
	}
	switch x := v.(type) {
	case string:
		if x == "" {
			return "-"
		}
		return x
	default:
		return fmt.Sprintf("%v", v)
	}
}

// HeadItems truncates items to at most `n`. n<=0 means "no truncation".
// Mirrors --head's "first-N rows" semantics on top of any sort order the
// command established.
func HeadItems(items []Item, n int) []Item {
	if n <= 0 || n >= len(items) {
		return items
	}
	return items[:n]
}
