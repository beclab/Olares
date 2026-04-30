package whoami

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

// Output is the table-vs-json selector. Kept as a free type rather than a
// bool so future shape switches (yaml? ndjson?) don't churn callers.
type Output string

const (
	OutputTable Output = "table"
	OutputJSON  Output = "json"
)

// ParseOutput normalizes the user-supplied --output flag. Empty defaults to
// OutputTable. Anything unrecognised returns an error so we fail fast on
// typos rather than silently render the table form.
func ParseOutput(s string) (Output, error) {
	v := strings.ToLower(strings.TrimSpace(s))
	switch v {
	case "", string(OutputTable):
		return OutputTable, nil
	case string(OutputJSON):
		return OutputJSON, nil
	default:
		return "", fmt.Errorf("unsupported --output %q (allowed: table, json)", s)
	}
}

// Display is the JSON-serialisable view we render to stdout. Wire role
// (`role`) is the source of truth for scripts; `roleLabel` is the
// SPA-aligned friendly label ("User" for "normal"); `source` tells the
// caller whether the line is fresh or from cache, which matters when
// scripting around drift detection.
type Display struct {
	OlaresID         string `json:"olaresId"`
	Name             string `json:"name,omitempty"`
	Role             string `json:"role,omitempty"`
	RoleLabel        string `json:"roleLabel,omitempty"`
	Source           string `json:"source"` // "cache" or "server"
	RefreshedAt      int64  `json:"refreshedAt,omitempty"`
	RefreshedAgoSecs int64  `json:"refreshedAgoSecs,omitempty"`
	PreviousRole     string `json:"previousRole,omitempty"` // populated only when Changed
	Changed          bool   `json:"changed,omitempty"`
}

// Render writes `d` to `w` in the requested format.
//
// Table mode writes a compact, two-line summary intentionally close to
// `gh auth status` in feel: identity on top, freshness underneath. JSON
// mode emits the Display struct verbatim — wire role + friendly label both
// included so callers can pick whichever they prefer.
func Render(w io.Writer, d Display, format Output) error {
	switch format {
	case OutputJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(d)
	case OutputTable, "":
		return renderTable(w, d)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

func renderTable(w io.Writer, d Display) error {
	role := d.RoleLabel
	if role == "" {
		role = "(unknown)"
	}
	id := d.OlaresID
	if d.Name != "" && d.Name != strings.Split(d.OlaresID, "@")[0] {
		// Surface the BFL `name` field separately when it diverges from
		// the local part of olaresId. Most installs keep them identical
		// (alice@olares.com → name="alice") but it's not enforced, so
		// don't pretend they're the same.
		id = fmt.Sprintf("%s (name: %s)", d.OlaresID, d.Name)
	}

	if _, err := fmt.Fprintf(w, "%s — %s\n", id, role); err != nil {
		return err
	}
	freshness := "no role cached yet"
	switch d.Source {
	case "server":
		freshness = "fetched just now from " + Endpoint
	case "cache":
		if d.RefreshedAt > 0 {
			freshness = fmt.Sprintf("from cache, refreshed %s",
				humanizeAgo(d.RefreshedAgoSecs))
		} else {
			freshness = "from cache (no refresh timestamp)"
		}
	}
	if _, err := fmt.Fprintln(w, freshness); err != nil {
		return err
	}
	if d.Changed && d.PreviousRole != "" {
		if _, err := fmt.Fprintf(w, "role changed: %s -> %s\n",
			FriendlyLabel(d.PreviousRole), role); err != nil {
			return err
		}
	}
	return nil
}

// humanizeAgo turns a delta in seconds into a coarse "5 minutes ago" /
// "3 hours ago" string. Tiny enough that pulling in a duration library
// for it would be overkill.
func humanizeAgo(secs int64) string {
	if secs < 0 {
		secs = 0
	}
	switch {
	case secs < 60:
		return fmt.Sprintf("%ds ago", secs)
	case secs < 3600:
		return fmt.Sprintf("%dm ago", secs/60)
	case secs < 86400:
		return fmt.Sprintf("%dh ago", secs/3600)
	default:
		return fmt.Sprintf("%dd ago", secs/86400)
	}
}

// Run is the shared driver for `profile whoami` / `settings users me` /
// `settings me whoami`. It encapsulates the "use cache unless --refresh or
// cache empty, then refetch and persist" policy so the three cobra
// wrappers stay cosmetic.
//
// Inputs:
//   - client     the SettingsClient-shaped Doer (DoJSON only); resolved by
//                the caller from cmdutil.Factory.
//   - cfg        the loaded MultiProfileConfig. Mutated in place when a
//                refetch happens; saved by FetchAndCache via SetOwnerRole.
//   - olaresID   the olaresId of the profile whose role we report.
//   - refresh    true means "force a server roundtrip even when the cache
//                is populated".
//   - format     the resolved Output mode (table / json).
//   - now        injected clock for testability; pass time.Now in prod.
//   - w          stdout for the printed result.
//
// Returned error is the first hard failure (HTTP, decode, config write).
// Pre-existing soft anomalies (cache hit on a profile with no
// WhoamiRefreshedAt; first-time fetch) flow through as Display.Changed
// flags, not errors.
func Run(
	ctx context.Context,
	client Doer,
	cfg *cliconfig.MultiProfileConfig,
	olaresID string,
	refresh bool,
	format Output,
	now func() time.Time,
	w io.Writer,
) error {
	if w == nil {
		return fmt.Errorf("whoami: nil writer")
	}
	if now == nil {
		now = time.Now
	}

	prof := cfg.FindByOlaresID(olaresID)
	if prof == nil {
		return fmt.Errorf("whoami: profile %q not found in config", olaresID)
	}

	hadCache := prof.OwnerRole != ""
	mustFetch := refresh || !hadCache

	d := Display{
		OlaresID: olaresID,
	}

	if mustFetch {
		res, err := FetchAndCache(ctx, client, cfg, olaresID, now)
		if err != nil {
			return err
		}
		d.Name = res.Info.Name
		d.Role = res.Info.OwnerRole
		d.RoleLabel = FriendlyLabel(res.Info.OwnerRole)
		d.Source = "server"
		d.RefreshedAt = res.RefreshedAt
		d.Changed = res.Changed
		d.PreviousRole = res.PreviousRole
		return Render(w, d, format)
	}

	// Cache hit path: render without a roundtrip so `whoami` is cheap and
	// works offline as long as a previous online run populated the cache.
	d.Role = prof.OwnerRole
	d.RoleLabel = FriendlyLabel(prof.OwnerRole)
	d.Source = "cache"
	d.RefreshedAt = prof.WhoamiRefreshedAt
	if prof.WhoamiRefreshedAt > 0 {
		d.RefreshedAgoSecs = now().Unix() - prof.WhoamiRefreshedAt
	}
	return Render(w, d, format)
}
