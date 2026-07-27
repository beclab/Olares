package whoami

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/beclab/Olares/cli/pkg/access"
	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/olares"
)

// detect.go is the unified "where/who/what am I relative to this Olares?"
// pass. It generalizes Run's "cache unless refresh, then fetch + persist"
// policy across three independent facts — network position (Location), role
// (/api/backend/v1/user-info) and backend version (/api/olares-info) — each
// backed by its single atomic fetcher (access.ProbeLocation, FetchAndCache,
// FetchAndCacheVersion). `profile whoami` is its command surface; the eager
// post-login/import fetch funnels through DetectAndCache too, so there is one
// detect implementation and the role/version fetches always use the
// just-detected connection method instead of a hard-wired public URL.

// DetectInput configures a full (server) detect pass.
type DetectInput struct {
	Cfg         *cliconfig.MultiProfileConfig
	OlaresID    string
	LocalPrefix string
	Insecure    bool
	// AccessToken is injected on every fetch. The detect path uses a fresh
	// token client (NewHTTPClientWithToken) rather than a Factory client so it
	// can target whichever Location it just probed.
	AccessToken string
	// AuthURLOverride, when non-empty, marks the profile as pinned to an
	// explicit auth endpoint (dev/internal use). It mirrors the login/import
	// probeProfileLocation + factory.maybeBackfillLocation contract: a pinned
	// profile never gets probed — its position is LocationExternal — so a
	// `whoami --refresh` / `list --refresh` can't silently rewrite the
	// service URLs to a probed LAN/host/cluster position while auth stays on
	// the override. Ignored when KnownLocation is already set.
	AuthURLOverride string
	// KnownLocation lets callers that already probed (profile login / import)
	// skip the probe and reuse the result. Zero/invalid → probe now.
	KnownLocation olares.Location
	// Now is injected for testability; nil → time.Now.
	Now func() time.Time
}

// DetectDisplay is the JSON/table view rendered by `profile whoami`. It is a
// superset of the role-only Display the settings aliases use.
type DetectDisplay struct {
	OlaresID       string `json:"olaresId"`
	Name           string `json:"name,omitempty"`
	Location       string `json:"location,omitempty"`
	Connection     string `json:"connection,omitempty"`
	Role           string `json:"role,omitempty"`
	RoleLabel      string `json:"roleLabel,omitempty"`
	BackendVersion string `json:"backendVersion,omitempty"`
	Source         string `json:"source"` // "server" (re-detected) or "cache"
}

// ConnectionLabel renders a human-readable description of the connection
// method implied by loc — the URL family plus how DNS is resolved.
func ConnectionLabel(loc olares.Location) string {
	switch loc {
	case olares.LocationExternal:
		return "public HTTPS via system DNS"
	case olares.LocationLAN:
		return "LAN over http://*.olares.local"
	case olares.LocationHost:
		return "intranet HTTPS via in-cluster DNS (" + olares.ClusterDNS + ")"
	case olares.LocationCluster:
		return "in-cluster HTTPS"
	default:
		return "unknown"
	}
}

// ErrCacheWrite marks a detect pass whose server-side facts are sound but
// could not be written to config.json (lock timeout, I/O error, ...). Callers
// can test for it with errors.Is to tell "the backend didn't answer" apart
// from "we know the answer but couldn't remember it".
var ErrCacheWrite = errors.New("detected values could not be cached")

// newDetectClient indirects client construction so tests can drive a detect
// pass without a network; production never reassigns it.
var newDetectClient = func(desktopURL, olaresID, accessToken string, insecure bool, loc olares.Location) Doer {
	return NewHTTPClientWithToken(desktopURL, olaresID, accessToken, insecure, loc)
}

// DetectAndCache runs a full detect: it resolves the Location (probing unless
// KnownLocation is set), then fetches role + version through a client bound to
// that Location and persists all three in one locked write. It returns the
// assembled display plus any hard error encountered: a role/version fetch
// failure (the other facts are still cached), and/or ErrCacheWrite when the
// cache write itself failed. A probe failure (every method down) is returned
// as-is with a nil display.
func DetectAndCache(ctx context.Context, in DetectInput) (*DetectDisplay, error) {
	now := in.Now
	if now == nil {
		now = time.Now
	}
	id, err := olares.ParseID(in.OlaresID)
	if err != nil {
		return nil, err
	}

	loc := in.KnownLocation
	if !loc.Valid() {
		if in.AuthURLOverride != "" {
			// Pinned auth endpoint: do not probe (mirrors login/import's
			// probeProfileLocation and factory.maybeBackfillLocation), so a
			// refresh can't move the profile off external while auth stays on
			// the override.
			loc = olares.LocationExternal
		} else {
			probed, perr := access.ProbeLocation(ctx, id, in.LocalPrefix, in.Insecure)
			if perr != nil {
				return nil, perr
			}
			loc = probed
		}
	}

	ep := id.Endpoints(loc, in.LocalPrefix)
	client := newDetectClient(ep.Desktop, in.OlaresID, in.AccessToken, in.Insecure, loc)

	d := &DetectDisplay{
		OlaresID:   in.OlaresID,
		Location:   string(loc),
		Connection: ConnectionLabel(loc),
		Source:     "server",
	}

	// Fetch role + version WITHOUT each writing config (nil cfg = value-only);
	// the location + whatever was fetched are persisted together in a single
	// locked write below, so a detect pass touches config.json exactly once.
	var (
		role, version     string
		roleAt, versionAt int64
		firstErr          error
	)
	if roleRes, rerr := FetchAndCache(ctx, client, nil, in.OlaresID, now); rerr != nil {
		firstErr = rerr
	} else {
		d.Name = roleRes.Info.Name
		d.Role = roleRes.Info.OwnerRole
		d.RoleLabel = FriendlyLabel(roleRes.Info.OwnerRole)
		role = roleRes.Info.OwnerRole
		roleAt = roleRes.RefreshedAt
	}
	if verRes, verr := FetchAndCacheVersion(ctx, client, nil, in.OlaresID, now); verr != nil {
		if firstErr == nil {
			firstErr = verr
		}
	} else {
		d.BackendVersion = verRes.OsVersion
		if verRes.Version != nil {
			version = verRes.Version.Original()
		}
		versionAt = verRes.RefreshedAt
	}

	if in.Cfg != nil {
		if serr := in.Cfg.SetDetectResults(in.OlaresID, string(loc), now().Unix(), role, roleAt, version, versionAt); serr != nil {
			// The facts we're about to return are real, but nothing reached
			// config.json — every later command will still see the old
			// values. Report it rather than letting the caller print
			// "re-detected just now" over an unchanged cache.
			firstErr = errors.Join(firstErr, fmt.Errorf("%w: %w", ErrCacheWrite, serr))
		}
	}
	return d, firstErr
}

// DetectFromCache assembles a DetectDisplay from the persisted profile fields
// only — no network. Used by `profile whoami` without --refresh.
func DetectFromCache(cfg *cliconfig.MultiProfileConfig, olaresID string) *DetectDisplay {
	d := &DetectDisplay{OlaresID: olaresID, Source: "cache"}
	if cfg == nil {
		return d
	}
	p := cfg.FindByOlaresID(olaresID)
	if p == nil {
		return d
	}
	d.Name = p.Name
	d.Location = p.Location
	d.Connection = ConnectionLabel(olares.Location(p.Location))
	d.Role = p.OwnerRole
	d.RoleLabel = FriendlyLabel(p.OwnerRole)
	d.BackendVersion = p.BackendVersion
	return d
}

// RenderDetect writes d to w in the requested format.
func RenderDetect(w io.Writer, d *DetectDisplay, format Output) error {
	switch format {
	case OutputJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(d)
	case OutputTable, "":
		return renderDetectTable(w, d)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

func renderDetectTable(w io.Writer, d *DetectDisplay) error {
	id := d.OlaresID
	if d.Name != "" && d.Name != olares.ID(d.OlaresID).Local() {
		id = fmt.Sprintf("%s (name: %s)", d.OlaresID, d.Name)
	}
	role := d.RoleLabel
	if role == "" {
		role = "(unknown)"
	}
	where := d.Location
	if where == "" {
		where = "(undetected)"
	} else if d.Connection != "" {
		where = fmt.Sprintf("%s (%s)", where, d.Connection)
	}
	version := d.BackendVersion
	if version == "" {
		version = "-"
	}

	if _, err := fmt.Fprintln(w, id); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "where:   %s\n", where); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "role:    %s\n", role); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "version: %s\n", version); err != nil {
		return err
	}
	switch d.Source {
	case "server":
		_, err := fmt.Fprintln(w, "(re-detected just now)")
		return err
	default:
		_, err := fmt.Fprintln(w, "(from cache; run `olares-cli profile whoami --refresh` to re-detect)")
		return err
	}
}
