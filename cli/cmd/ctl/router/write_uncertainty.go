package router

// What to tell a caller whose write never got an answer.
//
// A request that fails at the transport level has two possible histories and
// the client cannot tell them apart: Router never saw it, or Router applied it
// and the reply was lost. For a GET the distinction does not matter. For a
// POST it decides whether the next step is "do it again" or "go and look".
//
// Reported as a bare wrapped error — `POST …/quotas: context deadline
// exceeded` — it reads unambiguously as the first, which is how a caller ends
// up issuing a second API key it did not want, or reading a 409 as a new
// problem rather than as proof the first attempt landed.
//
// Router has no idempotency key. It is deferred to v2 (ARCHITECTURE ADR-11),
// so a retry is a second request and not a repeat of the first, and nothing
// the CLI sends can change that. What is left is to say so, and to say what a
// second request would do to this particular route — which the CLI knows,
// because it knows which route it just called.

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

// WriteUncertainError is a write that left this machine without an answer.
//
// Typed rather than a formatted string, because it is a state a caller may
// want to branch on: an agent that can re-read a list should do that instead
// of retrying, and matching on the wording of a network error to decide is the
// kind of control flow that breaks when Go changes a message.
type WriteUncertainError struct {
	Method string
	Path   string
	Err    error
}

func (e *WriteUncertainError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s: %v\n", e.Method, e.Path, e.Err)
	b.WriteString("the request left this machine and no reply came back, so whether Router applied " +
		"it is unknown. Router has no idempotency key, so sending it again is a second request " +
		"rather than the same one")
	if note := repeatNote(e.Method, e.Path); note != "" {
		b.WriteString(". ")
		b.WriteString(note)
	} else {
		b.WriteString(" — check before retrying")
	}
	return b.String()
}

func (e *WriteUncertainError) Unwrap() error { return e.Err }

// reachedRouter is whether the request plausibly arrived. A connection that was
// never established did not, and saying "this may have been applied" about an
// unreachable host is noise that teaches a reader to skip the warning on the
// one occasion it is true.
//
// Everything else is treated as uncertain, including a deadline that expired
// while the connection was being made. Being wrong towards "go and look" costs
// a read; being wrong the other way costs a duplicate.
func reachedRouter(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) && opErr.Op == "dial" {
		return false
	}
	var dnsErr *net.DNSError
	return !errors.As(err, &dnsErr)
}

// repeatNote says what a second identical request would do to this route, and
// where to look instead.
//
// This is a hand copy of what Router's own constraints happen to enforce, and
// nothing here can check it: the UNIQUE indexes and the ON CONFLICT clauses
// that make some of these safe live in another repository. So the default is
// the cautious one — a route not named here gets no promise, only the advice
// to look — and a route that gains a duplicate-name constraint keeps working
// with a message that is merely less helpful than it could be.
//
// The reverse mistake is the one to avoid: telling somebody a retry is refused
// when it is not.
func repeatNote(method, path string) string {
	path = strings.SplitN(path, "?", 2)[0]

	switch method {
	case "PATCH", "PUT":
		return "This route sets fields rather than creating a row, so sending it again " +
			"lands on the same state either way"
	case "DELETE":
		return "This route removes something, so sending it again either removes it or " +
			"reports it is already gone"
	}

	// The data plane. Not a configuration write, and the only one here that
	// costs money: a completion Router finished and could not deliver was
	// still billed, and a retry bills again.
	if isDataPlanePath(path) {
		return "This was a model call, and a call Router completed was billed whether or not " +
			"the answer arrived. `olares-cli router usage list --limit 5` says whether it ran"
	}

	switch {
	case path == epAPIKeys:
		return "A second attempt mints a second key rather than being refused — " +
			"`olares-cli router key list` says whether the first one exists"

	case path == epProviders:
		return "Router refuses a duplicate provider name, so a retry is safe: a 409 means the " +
			"first attempt landed. `olares-cli router provider list` says so directly"

	case path == epModelRoutes:
		return "Router refuses a duplicate route name, so a retry is safe: a 409 means the " +
			"first attempt landed. `olares-cli router route list` says so directly"

	case path == epQuotas:
		// One row per (scope, kind, period), not one per scope: a key
		// legitimately carries a budget and an rpm and a tpm ceiling at once.
		// What makes the retry safe is that the repeat names the same kind.
		return "Router allows one ceiling of each kind per scope, so sending the same one again " +
			"is safe: a 409 means the first attempt landed. `olares-cli router quota list` says so directly"

	case strings.HasSuffix(path, "/predefined-models"):
		return "Adding models skips the ones already attached, so a retry is safe"

	case strings.HasSuffix(path, "/customizable-models"):
		return "Router refuses a model name the provider already has, so a retry is safe: a 409 " +
			"means the first attempt landed. `olares-cli router model list` says so directly"

	case strings.HasSuffix(path, "/sync-models"):
		return "A sync adds what is missing and removes what the upstream no longer lists, so a " +
			"retry converges rather than duplicating — but the removals may already have happened"

	case strings.Contains(path, "/rollback/"):
		return "Each rollback that succeeds becomes a new credential version, so a second one " +
			"appends another. `olares-cli router provider history <provider>` says how far it got"

	case strings.HasSuffix(path, "/validate"):
		return "Validation changes nothing, so a retry is safe"

	case path == epLocalRetry, path == epLocalEngineRestart, path == epEngineRestart:
		return "This restarts the process serving the model, so a second one restarts it again " +
			"and the model stops answering for longer. `olares-cli router model progress <model>` " +
			"says where it got to"
	}

	return ""
}
