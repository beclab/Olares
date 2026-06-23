package preflight

import (
	"context"
	"fmt"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// Version gates are the shared way every `olares-cli settings` verb expresses
// an Olares-backend version requirement, so individual areas don't re-roll the
// same OlaresBackendAtLeast + error-message boilerplate.
//
// Two opposite-direction gates are provided:
//
//   - RequireMinVersion — a verb that needs the backend to be at least some
//     version (a feature introduced in release X). Fail-closed: an
//     undetectable version is rejected, because the feature provably does not
//     exist on anything older and the call would 404.
//
//   - RejectIfRemoved — a verb whose backing API was removed at some version
//     (a legacy surface dropped in release X). Fail-open: an undetectable
//     version is let through so the server stays authoritative and older
//     backends where the endpoint still works keep behaving as before.
//
// Both consult the version cached by `profile login` / `profile whoami` (or
// the --olares-version override), so in the common case they add no network
// round-trip.

// MinVersionGate describes a verb gated to a minimum Olares backend version.
type MinVersionGate struct {
	// Verb is the human command name, e.g. "settings compute". Rendered in
	// backticks in the error.
	Verb string
	// MinVersion is the first Olares line that supports the verb, e.g. "1.12.6".
	MinVersion string
	// Reason is an optional short parenthetical, e.g. "compute-resources APIs".
	Reason string
	// Fallback is an optional trailing hint appended after the version error,
	// e.g. "use the legacy `olares-cli settings gpu list` on 1.12.5".
	Fallback string
}

// RequireMinVersion fails fast when the backend is below the gate's minimum,
// or when the version is undetectable (fail-closed). The --olares-version flag
// is suggested as the escape hatch for the undetectable case.
func RequireMinVersion(ctx context.Context, f *cmdutil.Factory, gate MinVersionGate) error {
	if f == nil {
		return nil
	}
	reason := ""
	if gate.Reason != "" {
		reason = " (" + gate.Reason + ")"
	}
	ok, err := f.OlaresBackendAtLeast(ctx, gate.MinVersion)
	if err != nil {
		return fmt.Errorf(
			"`%s` requires Olares >= %s%s, but the backend version could not be determined: %v; "+
				"pass --%s <version> to set it manually (e.g. --%s %s)",
			gate.Verb, gate.MinVersion, reason, err,
			cmdutil.FlagOlaresVersion, cmdutil.FlagOlaresVersion, gate.MinVersion)
	}
	if !ok {
		got := "unknown"
		if v, verr := f.OlaresBackendVersion(ctx); verr == nil && v != nil {
			got = v.Original()
		}
		msg := fmt.Sprintf("`%s` requires Olares >= %s%s, but this backend is %s",
			gate.Verb, gate.MinVersion, reason, got)
		if gate.Fallback != "" {
			msg += "; " + gate.Fallback
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}

// RemovedGate describes a verb whose backing API was removed at a given Olares
// backend version.
type RemovedGate struct {
	// Verb is the human command name, e.g. "settings gpu list".
	Verb string
	// Detail is an optional parenthetical naming the removed surface, e.g.
	// "legacy HAMI /api/gpu/list".
	Detail string
	// RemovedIn is the first Olares line that no longer ships the API, e.g. "1.12.6".
	RemovedIn string
	// Replacement is an optional replacement command to suggest, e.g.
	// "olares-cli settings compute resources list".
	Replacement string
}

// RejectIfRemoved fails fast when the backend is at or after the version that
// removed the verb, pointing the user at the replacement. Fail-open on an
// undetectable version: the call is let through so the server stays
// authoritative and older backends keep working.
func RejectIfRemoved(ctx context.Context, f *cmdutil.Factory, gate RemovedGate) error {
	if f == nil {
		return nil
	}
	ok, err := f.OlaresBackendAtLeast(ctx, gate.RemovedIn)
	if err != nil || !ok {
		return nil
	}
	detail := ""
	if gate.Detail != "" {
		detail = " (" + gate.Detail + ")"
	}
	msg := fmt.Sprintf("`%s`%s was removed in Olares %s", gate.Verb, detail, gate.RemovedIn)
	if gate.Replacement != "" {
		msg += fmt.Sprintf("; use `%s` instead", gate.Replacement)
	}
	return fmt.Errorf("%s", msg)
}
