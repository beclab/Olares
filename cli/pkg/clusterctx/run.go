package clusterctx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// Output reuses pkg/whoami.Output to keep the --output flag's allowed
// values in lockstep across `profile whoami` and `cluster context`.
// (Both ultimately render a small identity-shaped struct in either
// table or json.)
type Output = whoami.Output

// ParseOutput delegates to whoami.ParseOutput for the same reason
// Output does — single source of truth for "table" / "json" parsing.
func ParseOutput(s string) (Output, error) { return whoami.ParseOutput(s) }

// Display is the JSON-serialisable view we render to stdout. Wire role
// (`globalrole`) is the source of truth for scripts; `globalRoleLabel`
// is the friendly form via FriendlyGlobalRole; `source` tells the
// caller whether the line is fresh or from cache, mirroring
// whoami.Display.Source so existing tooling can reuse the same field.
type Display struct {
	OlaresID         string   `json:"olaresId"`
	Username         string   `json:"username,omitempty"`
	Email            string   `json:"email,omitempty"`
	GlobalRole       string   `json:"globalrole,omitempty"`
	GlobalRoleLabel  string   `json:"globalroleLabel,omitempty"`
	ClusterRole      string   `json:"clusterRole,omitempty"`
	Workspaces       []string `json:"workspaces,omitempty"`
	SystemNamespaces []string `json:"systemNamespaces,omitempty"`
	GrantedClusters  []string `json:"grantedClusters,omitempty"`
	Source           string   `json:"source"` // "cache" or "server"
	RefreshedAt      int64    `json:"refreshedAt,omitempty"`
	RefreshedAgoSecs int64    `json:"refreshedAgoSecs,omitempty"`
	PreviousRole     string   `json:"previousGlobalrole,omitempty"`
	Changed          bool     `json:"changed,omitempty"`
}

// Render writes `d` to `w` in the requested format. Table mode keeps
// the same `gh auth status`-feel two-line summary as whoami's table:
// identity on top, a freshness line, then the optional drift notice;
// the workspace/namespace lists follow when present.
func Render(w io.Writer, d Display, format Output) error {
	switch format {
	case whoami.OutputJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(d)
	case whoami.OutputTable, "":
		return renderTable(w, d)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

func renderTable(w io.Writer, d Display) error {
	role := d.GlobalRoleLabel
	if role == "" {
		role = "(unknown)"
	}
	id := d.OlaresID
	if d.Username != "" && d.Username != strings.Split(d.OlaresID, "@")[0] {
		// Surface the ControlHub `username` separately when it diverges
		// from the local part of olaresId — same convention as
		// whoami.renderTable.
		id = fmt.Sprintf("%s (username: %s)", d.OlaresID, d.Username)
	}
	if _, err := fmt.Fprintf(w, "%s — %s\n", id, role); err != nil {
		return err
	}

	freshness := "no cluster context cached yet"
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
		if _, err := fmt.Fprintf(w, "globalrole changed: %s -> %s\n",
			FriendlyGlobalRole(d.PreviousRole), role); err != nil {
			return err
		}
	}

	// Workspaces + system namespaces + granted clusters, when present.
	// Server may legitimately return empty arrays for a fresh user with
	// no access yet, so we only print the line when there's something
	// to render — keeps the default output dense.
	if len(d.Workspaces) > 0 {
		if _, err := fmt.Fprintf(w, "workspaces:        %s\n", strings.Join(d.Workspaces, ", ")); err != nil {
			return err
		}
	}
	if len(d.SystemNamespaces) > 0 {
		if _, err := fmt.Fprintf(w, "system namespaces: %s\n", strings.Join(d.SystemNamespaces, ", ")); err != nil {
			return err
		}
	}
	if len(d.GrantedClusters) > 0 {
		if _, err := fmt.Fprintf(w, "granted clusters:  %s\n", strings.Join(d.GrantedClusters, ", ")); err != nil {
			return err
		}
	}
	if d.ClusterRole != "" {
		if _, err := fmt.Fprintf(w, "cluster role:      %s\n", d.ClusterRole); err != nil {
			return err
		}
	}
	if d.Email != "" {
		if _, err := fmt.Fprintf(w, "email:             %s\n", d.Email); err != nil {
			return err
		}
	}
	return nil
}

// humanizeAgo mirrors whoami.humanizeAgo (kept private over there).
// Tiny enough that pulling in a duration formatting library would be
// overkill, and the wording stays consistent if we ever change one of
// the two.
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

// Run is the shared driver for `cluster context`. Encapsulates the
// "use cache unless --refresh or cache empty, then refetch and persist"
// policy. Mirrors pkg/whoami.Run so the two surfaces (BFL whoami /
// ControlHub context) stay symmetric in behavior.
//
// Inputs:
//   - client     a Doer (typically clusterclient.Client or
//                clusterctx.HTTPClient) wired to the active profile's
//                ControlHub URL.
//   - cfg        the loaded MultiProfileConfig. Mutated in place when a
//                refetch happens; saved by FetchAndCache via
//                SetClusterContext.
//   - olaresID   olaresId of the profile whose cluster context we
//                report.
//   - refresh    true means "force a server roundtrip even when the
//                cache is populated".
//   - format     resolved Output mode (table / json).
//   - now        injected clock for testability; pass time.Now in prod.
//   - w          stdout for the printed result.
//
// Returned error is the first hard failure (HTTP, decode, config
// write). Pre-existing soft anomalies (cache hit on a profile with no
// ClusterContextRefreshedAt; first-time fetch) flow through as
// Display.Changed flags, not errors.
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
		return fmt.Errorf("clusterctx: nil writer")
	}
	if now == nil {
		now = time.Now
	}

	prof := cfg.FindByOlaresID(olaresID)
	if prof == nil {
		return fmt.Errorf("clusterctx: profile %q not found in config", olaresID)
	}

	hadCache := prof.ClusterContext != nil
	mustFetch := refresh || !hadCache

	d := Display{OlaresID: olaresID}

	if mustFetch {
		res, err := FetchAndCache(ctx, client, cfg, olaresID, now)
		if err != nil {
			return err
		}
		populateDisplay(&d, res.Info)
		d.Source = "server"
		d.RefreshedAt = res.RefreshedAt
		d.Changed = res.Changed
		d.PreviousRole = res.PreviousGlobalRole
		return Render(w, d, format)
	}

	// Cache hit: render without a roundtrip so `cluster context` is
	// cheap and works offline as long as a previous online run
	// populated the cache.
	populateDisplay(&d, fromCache(prof.ClusterContext))
	d.Source = "cache"
	d.RefreshedAt = prof.ClusterContextRefreshedAt
	if prof.ClusterContextRefreshedAt > 0 {
		d.RefreshedAgoSecs = now().Unix() - prof.ClusterContextRefreshedAt
	}
	return Render(w, d, format)
}

func populateDisplay(d *Display, info Info) {
	d.Username = info.Username
	d.Email = info.Email
	d.GlobalRole = info.GlobalRole
	d.GlobalRoleLabel = FriendlyGlobalRole(info.GlobalRole)
	d.ClusterRole = info.ClusterRole
	d.Workspaces = info.Workspaces
	d.SystemNamespaces = info.SystemNamespaces
	d.GrantedClusters = info.GrantedClusters
}
