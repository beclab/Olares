package whoami

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

// Soft preflight + 403 hint helpers shared between command surfaces.
//
// "Soft" means: every check here is best-effort. The server is the only
// authoritative source for "can this user do this" — the helpers here
// just save round-trips when the local cache already says no, and
// translate server-side 403 / unauthorized responses into a CLI-friendly
// hint that points at `whoami --refresh`.
//
// Why both halves? The two failure modes are distinct:
//
//   - Cache says no, server hasn't been asked yet:
//     PreflightRole short-circuits with a hint to refresh — no wasted
//     API call, the user has a clear next step.
//
//   - Cache says yes (or is empty) but the server rejects on permission:
//     WrapPermissionErr catches it after the API call and explains
//     what likely happened (role drift since last cache write, e.g.
//     owner demoted them). Same suggested action: refresh + retry.
//
// Both paths share the same suggested-next-step phrasing so the user
// learns exactly one trick (`olares-cli profile whoami --refresh`).

// PreflightRole short-circuits a verb before any API call when the
// locally-cached role is provably below `required`. Empty / missing
// cached role → no opinion, returns nil (server decides). Unknown
// `required` value → no opinion either; we don't gate on labels we
// don't understand because future-CLI/server skew shouldn't break
// existing verbs.
//
// `verbDescr` is a short human-friendly description of what the verb
// is trying to do (e.g. "list users", "set FRP server"); it's
// interpolated into the error message so the user knows which call
// got blocked. Keep it lowercase, present-tense, no trailing
// punctuation.
//
// Important: this never reaches out to the network, so callers can
// invoke it cheaply at the top of every gated RunE. Verbs that are
// fine for any authenticated user (e.g. `settings me whoami`) should
// simply not call this — passing required=RoleNormal to "be safe"
// adds nothing because empty/unknown roles always pass through.
func PreflightRole(cfg *cliconfig.MultiProfileConfig, olaresID string, required string, verbDescr string) error {
	if cfg == nil || olaresID == "" || required == "" {
		return nil
	}
	requiredRank := Rank(required)
	if requiredRank == 0 {
		return nil
	}
	prof := cfg.FindByOlaresID(olaresID)
	if prof == nil || prof.OwnerRole == "" {
		return nil
	}
	have := Rank(prof.OwnerRole)
	if have == 0 || have >= requiredRank {
		return nil
	}
	return fmt.Errorf(
		"this command needs role %q or higher to %s, but profile %q is cached as %q\n"+
			"  if your role on the server changed recently, run:\n"+
			"      olares-cli profile whoami --refresh\n"+
			"  and retry. Otherwise ask the instance owner to grant you the right role.",
		FriendlyLabel(required), verbDescr, olaresID, FriendlyLabel(prof.OwnerRole))
}

// WrapPermissionErr post-processes an HTTP error returned by any of the
// CLI's edge-API clients. When the underlying error is a 403 (or a 401
// the call site decided is permission-shaped), we add the same "refresh
// and retry" hint PreflightRole emits — closing the loop on the
// "cache stale → server rejected" branch.
//
// `verbDescr` follows the same convention as PreflightRole.
//
// We intentionally don't try to parse the upstream JSON body for the
// role reason: BFL / user-service / terminusd all phrase 403s slightly
// differently, and chasing every variant would be a maintenance
// treadmill. A generic "the server rejected this; here's how to refresh
// your cached role" is good enough — the user can always read the
// wrapped error if they want the verbatim server message.
func WrapPermissionErr(err error, olaresID, verbDescr string) error {
	if err == nil {
		return nil
	}
	var status int
	switch {
	case isHTTPStatus(err, http.StatusForbidden):
		status = http.StatusForbidden
	case isHTTPStatus(err, http.StatusUnauthorized):
		status = http.StatusUnauthorized
	default:
		return err
	}
	return fmt.Errorf("%w\n"+
		"  HTTP %d while attempting to %s.\n"+
		"  if your role on the server changed recently, run:\n"+
		"      olares-cli profile whoami --refresh\n"+
		"  and retry. Otherwise this likely needs a higher role on profile %q.",
		err, status, verbDescr, olaresID)
}

// isHTTPStatus is a deliberately simple substring sniffer. The shared
// settings client formats HTTP errors as
// "<METHOD> <url>: HTTP <status>: <body>", so we can reliably detect the
// status without plumbing a dedicated typed error through every call
// site.
//
// If we later add a typed *settings.HTTPError, switch this to errors.As.
func isHTTPStatus(err error, status int) bool {
	if err == nil {
		return false
	}
	needle := fmt.Sprintf("HTTP %d", status)
	return errorContains(err, needle)
}

func errorContains(err error, needle string) bool {
	for err != nil {
		if strings.Contains(err.Error(), needle) {
			return true
		}
		err = errors.Unwrap(err)
	}
	return false
}
