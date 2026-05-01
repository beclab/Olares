// Package clusterctx centralizes the "who am I against this Olares
// cluster?" round-trip used by `olares-cli cluster context` and by any
// future verb that wants to render the active profile's identity /
// role / accessible workspaces from the ControlHub side.
//
// It is the moral counterpart of pkg/whoami:
//
//   - pkg/whoami caches BFL identity (`/api/backend/v1/user-info`) onto
//     ProfileConfig.OwnerRole — the "who am I in BFL terms?" answer.
//   - pkg/clusterctx caches ControlHub identity
//     (`/capi/app/detail`) onto ProfileConfig.ClusterContext — the
//     "who am I in K8s/KubeSphere terms?" answer (globalrole +
//     accessible workspaces + system namespaces + granted clusters).
//
// The split is intentional: BFL OwnerRole gates the settings tree; the
// ControlHub identity scopes the cluster tree's display. Verb code in
// the cluster tree MUST NOT consult this cache to decide whether a call
// is allowed — server-side ControlHub does the actual gating, and
// trying to second-guess it client-side both adds attack surface
// (a tampered local cache could falsely "permit" something the server
// will then reject) and gets stale silently after role drift. The cache
// exists for display + error-message context only.
//
// The package is deliberately cobra-free: command files in
// cli/cmd/ctl/cluster import this package and wrap it in a thin RunE;
// the heavy lifting (HTTP, decode, drift detection, atomic config
// write) lives here once.
package clusterctx

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

// Endpoint is the ControlHub aggregator path that returns the active
// user's identity + accessible scope. See the SPA reference in
// apps/packages/app/src/apps/controlPanelCommon/network/network.ts:222
// (`AppDetailResponse`) and the wire-shape mapping below.
const Endpoint = "/capi/app/detail"

// Doer is the minimal HTTP surface FetchAndCache needs. ClusterClient
// (cli/pkg/clusterclient.Client) and clusterctx.HTTPClient both satisfy
// it; defining it locally keeps the import graph tidy.
//
// Signature is intentionally identical to pkg/whoami.Doer so a single
// http.Client can satisfy both packages without a wrapping shim.
type Doer interface {
	DoJSON(ctx context.Context, method, path string, body, out interface{}) error
}

// Info is the in-memory model of /capi/app/detail. Field names keep the
// wire JSON tags so callers can `--output json` straight from a Display
// without an extra mapping layer.
//
// Unmapped wire fields (config, ksConfig, user.lang, user.lastLoginTime,
// user.globalRules) are intentionally NOT modeled here — they're either
// megabytes of nested ks-installer state (config / ksConfig) or
// per-resource ACL maps (globalRules) the CLI never evaluates locally.
// Add them as later phases find specific uses.
type Info struct {
	Username         string   `json:"username,omitempty"`
	GlobalRole       string   `json:"globalrole,omitempty"`
	Email            string   `json:"email,omitempty"`
	Workspaces       []string `json:"workspaces,omitempty"`
	SystemNamespaces []string `json:"systemNamespaces,omitempty"`
	GrantedClusters  []string `json:"grantedClusters,omitempty"`
	ClusterRole      string   `json:"clusterRole,omitempty"`
}

// detailResponse mirrors the on-the-wire shape of /capi/app/detail. We
// keep it private and decode into Info via toInfo so callers see only
// the curated subset.
//
// Why not decode straight into Info: the SPA wraps user-facing fields
// inside a nested `user` object while top-level workspaces /
// systemNamespaces / clusterRole live at the document root. Flattening
// here keeps Info ergonomic.
type detailResponse struct {
	ClusterRole      string   `json:"clusterRole"`
	Workspaces       []string `json:"workspaces"`
	SystemNamespaces []string `json:"systemNamespaces"`
	User             struct {
		Email           string   `json:"email"`
		Globalrole      string   `json:"globalrole"`
		GrantedClusters []string `json:"grantedClusters"`
		Username        string   `json:"username"`
	} `json:"user"`
}

func (d detailResponse) toInfo() Info {
	return Info{
		Username:         d.User.Username,
		GlobalRole:       d.User.Globalrole,
		Email:            d.User.Email,
		Workspaces:       d.Workspaces,
		SystemNamespaces: d.SystemNamespaces,
		GrantedClusters:  d.User.GrantedClusters,
		ClusterRole:      d.ClusterRole,
	}
}

// toCacheEntry shapes the cliconfig persistence record for SetClusterContext.
// Returned as a pointer so callers can pass it through SetClusterContext's
// nullable contract (passing nil there means "explicit clear").
func (i Info) toCacheEntry() *cliconfig.ClusterContextCache {
	return &cliconfig.ClusterContextCache{
		Username:         i.Username,
		GlobalRole:       i.GlobalRole,
		Email:            i.Email,
		Workspaces:       i.Workspaces,
		SystemNamespaces: i.SystemNamespaces,
		GrantedClusters:  i.GrantedClusters,
		ClusterRole:      i.ClusterRole,
	}
}

// fromCache rebuilds an Info from a persisted ClusterContextCache so the
// cache-hit render path goes through the same Info → Display pipeline as
// the server-fetch path.
func fromCache(c *cliconfig.ClusterContextCache) Info {
	if c == nil {
		return Info{}
	}
	return Info{
		Username:         c.Username,
		GlobalRole:       c.GlobalRole,
		Email:            c.Email,
		Workspaces:       c.Workspaces,
		SystemNamespaces: c.SystemNamespaces,
		GrantedClusters:  c.GrantedClusters,
		ClusterRole:      c.ClusterRole,
	}
}

// Result is what callers get back from FetchAndCache: the freshly-decoded
// Info plus drift-detection metadata so callers can render
// "globalrole changed: X -> Y" hints without re-querying the cache.
//
// Changed=true fires for actual transitions (e.g. platform-regular →
// platform-admin) AND for first-time writes (no prior cache → role
// known) so first-login UX gets the same "your role is X" line as
// genuine changes.
type Result struct {
	Info               Info
	Changed            bool
	PreviousGlobalRole string
	WroteToCache       bool  // false when caller passed cfg=nil
	RefreshedAt        int64 // Unix-second timestamp written to cache
}

// FetchAndCache hits Endpoint with `client`, decodes the ControlHub
// detail response, and (when cfg is non-nil) atomically updates the
// matching profile's ClusterContext + ClusterContextRefreshedAt fields.
//
// olaresID is required so the cache write targets the right profile —
// callers that already have a ResolvedProfile usually pass rp.OlaresID.
//
// `now` is injected for testability; pass time.Now in production.
//
// On HTTP / decode failure FetchAndCache returns the error untouched —
// the explicit `cluster context` caller surfaces it as a regular error.
// (No equivalent of whoami's eager-fetch path exists yet for the
// ControlHub side; if/when one shows up, the same "wrap as warning"
// idiom can apply.)
func FetchAndCache(
	ctx context.Context,
	client Doer,
	cfg *cliconfig.MultiProfileConfig,
	olaresID string,
	now func() time.Time,
) (*Result, error) {
	if client == nil {
		return nil, errors.New("clusterctx: nil http client")
	}
	if olaresID == "" {
		return nil, errors.New("clusterctx: empty olaresID")
	}
	if now == nil {
		now = time.Now
	}

	var raw detailResponse
	if err := client.DoJSON(ctx, "GET", Endpoint, nil, &raw); err != nil {
		return nil, err
	}

	info := raw.toInfo()
	res := &Result{
		Info:        info,
		RefreshedAt: now().Unix(),
	}
	if cfg == nil {
		// In-memory only — used when the resolved profile came from the
		// EnvProvider (no on-disk profile to update) or test scaffolds.
		return res, nil
	}

	target := cfg.FindByOlaresID(olaresID)
	if target == nil {
		return nil, fmt.Errorf("clusterctx: profile %q not found in config", olaresID)
	}
	if target.ClusterContext != nil {
		res.PreviousGlobalRole = target.ClusterContext.GlobalRole
	}

	changed, err := cfg.SetClusterContext(olaresID, info.toCacheEntry(), res.RefreshedAt)
	if err != nil {
		return nil, err
	}
	res.Changed = changed
	res.WroteToCache = true
	return res, nil
}

// FriendlyGlobalRole renders a KubeSphere global-role wire constant in a
// human-shaped form for table output. The known values come from
// KubeSphere's `installer/roletemplates/...` bundle:
//
//   - "platform-admin"             → "Cluster Admin"
//   - "platform-self-provisioner"  → "Self Provisioner"
//   - "platform-regular"           → "Regular User"
//   - "anonymous"                  → "Anonymous"
//
// Unknown values fall through to the wire string Title-cased, so future
// roles stay readable instead of dropping silently. JSON output keeps
// the wire spelling — only the table renderer goes through this.
func FriendlyGlobalRole(wire string) string {
	switch strings.ToLower(strings.TrimSpace(wire)) {
	case "platform-admin":
		return "Cluster Admin"
	case "platform-self-provisioner":
		return "Self Provisioner"
	case "platform-regular":
		return "Regular User"
	case "anonymous":
		return "Anonymous"
	case "":
		return "(unknown)"
	default:
		// Replace dashes with spaces and Title-case each word.
		parts := strings.Split(wire, "-")
		for i, p := range parts {
			if p == "" {
				continue
			}
			parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
		}
		return strings.Join(parts, " ")
	}
}
