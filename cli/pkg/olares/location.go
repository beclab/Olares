package olares

import (
	"fmt"
	"net/url"
	"strings"
)

// Location describes where olares-cli sits on the network relative to the
// Olares instance it talks to. It is stored per-profile (cliconfig.
// ProfileConfig.Location) and selects the "connection method" — URL scheme +
// host + DNS resolver — at runtime.
//
// The four positions map many-to-one onto a connection method:
//
//	external → https://<svc>.<terminus>           (system DNS → public IP; slowest, universal)
//	lan      → http://<svc>.<local>.olares.local  (system DNS → LAN IP; fast)
//	host     → https://<svc>.<terminus>           (CLI-side DNS via ClusterDNS → intranet IP)
//	cluster  → https://<svc>.<terminus>           (pod /etc/resolv.conf is already cluster DNS)
//
// external / host / cluster share the SAME URL string and differ only in the
// http.Transport's resolver, which is why Location must travel alongside the
// resolved URLs (see credential.ResolvedProfile and pkg/access.Transport).
type Location string

const (
	LocationExternal Location = "external"
	LocationLAN      Location = "lan"
	LocationHost     Location = "host"
	LocationCluster  Location = "cluster"
)

// ClusterDNS is the in-cluster CoreDNS service IP. From the Olares host (the
// `host` Location) the CLI points its own resolver here so that the public
// `<svc>.<terminus>` hostnames resolve to intranet addresses without the user
// having to change their system DNS. Mirrors daemon/pkg/utils/cluster_api.go.
const ClusterDNS = "10.233.0.3"

// LANDomain is the suffix used for local-network access. Mirrors the web app's
// `*.olares.local` derivation (apps/packages/app/src/stores/desktop/token.ts
// getAppLocalUrl / getAuthURL), which uses ONLY the local part of the olaresId
// plus this suffix — the user's real domain is not part of the LAN hostname.
const LANDomain = "olares.local"

// Valid reports whether l is one of the four known positions.
func (l Location) Valid() bool {
	switch l {
	case LocationExternal, LocationLAN, LocationHost, LocationCluster:
		return true
	}
	return false
}

// UsesLocalDomain reports whether l resolves through the `.local` LAN domain
// over plain HTTP. Only the `lan` position does.
func (l Location) UsesLocalDomain() bool { return l == LocationLAN }

// UsesClusterResolver reports whether HTTP clients for l must resolve the
// public `<svc>.<terminus>` hostnames via the in-cluster DNS (ClusterDNS)
// rather than the system resolver. Only the `host` position does (cluster
// pods already inherit cluster DNS via /etc/resolv.conf; external/lan use the
// system resolver).
func (l Location) UsesClusterResolver() bool { return l == LocationHost }

// scheme is the URL scheme implied by l: plain HTTP for the LAN domain,
// HTTPS everywhere else (the edge terminates TLS on the public hostnames).
func (l Location) scheme() string {
	if l == LocationLAN {
		return "http"
	}
	return "https"
}

// hostSuffix returns the terminus-name portion of a service hostname for loc:
// "<local>.olares.local" for the LAN position, "<local>.<domain>" otherwise.
func (id ID) hostSuffix(loc Location) string {
	if loc == LocationLAN {
		return id.Local() + "." + LANDomain
	}
	return id.TerminusName()
}

// Endpoints bundles every per-service base URL the CLI talks to for a single
// (Location, localPrefix) pair. Returned by ID.Endpoints so callers derive
// all URLs from one place instead of calling the per-service helpers and
// risking a Location mismatch between them.
type Endpoints struct {
	Auth       string
	Vault      string
	Desktop    string
	Settings   string
	Files      string
	Market     string
	Dashboard  string
	ControlHub string
}

// Endpoints derives all per-service base URLs for loc. localPrefix is the
// dev-only label inserted between the service subdomain and the terminus name
// (e.g. "dev." → "auth.dev.alice.olares.com"); pass "" in production.
//
// An invalid/empty loc is treated as LocationExternal so callers that haven't
// probed yet still get working public URLs.
func (id ID) Endpoints(loc Location, localPrefix string) Endpoints {
	if !loc.Valid() {
		loc = LocationExternal
	}
	scheme := loc.scheme()
	base := id.hostSuffix(loc)
	svc := func(sub string) string {
		return fmt.Sprintf("%s://%s.%s%s", scheme, sub, localPrefix, base)
	}
	return Endpoints{
		Auth:       svc("auth"),
		Vault:      svc("vault") + "/server",
		Desktop:    svc("desktop"),
		Settings:   svc("settings"),
		Files:      svc("files"),
		Market:     svc("market"),
		Dashboard:  svc("dashboard"),
		ControlHub: svc("control-hub"),
	}
}

// RebaseURL rewrites u to target the same service under a different Location,
// preserving the path / query / fragment. It takes the first DNS label of
// u.Host as the service subdomain (e.g. "files" from
// "files.alice.olares.com") and rebuilds scheme + host for loc + localPrefix.
//
// Used by the reprobe path (cmdutil.Factory) to retry an in-flight request
// against a newly-detected Location without re-deriving which service it was
// hitting. external↔host↔cluster leave the URL byte-identical (only the
// transport changes); lan↔anything flips both scheme (http↔https) and the
// domain suffix (olares.local↔<domain>). Returns nil when u is nil.
func (id ID) RebaseURL(u *url.URL, loc Location, localPrefix string) *url.URL {
	if u == nil {
		return nil
	}
	if !loc.Valid() {
		loc = LocationExternal
	}
	out := *u
	sub := u.Hostname()
	if i := strings.Index(sub, "."); i >= 0 {
		sub = sub[:i]
	}
	out.Scheme = loc.scheme()
	out.Host = fmt.Sprintf("%s.%s%s", sub, localPrefix, id.hostSuffix(loc))
	return &out
}
