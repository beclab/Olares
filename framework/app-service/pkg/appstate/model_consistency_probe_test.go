package appstate

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// TestProbe_AllEnumeratesEveryReferencedState is a bug-hunting probe, not a
// behavior pin. It collects every state the model references anywhere
// (transition table, operation-gate table, the cancelable/operating/canceling
// sets, the timeout map) and asserts each one is enumerated in All.
//
// Why this matters: All is consumed by the apiserver as the default "all known
// states" filter -- handler_app.go:appsStatus builds a sets.String from
// appstate.All and then does `if !stateSet.Has(am.Status.State) { continue }`
// (handler_app.go also at lines ~258/502, and handler_service.go ~113). Any
// state that the system can actually put an app into but that is missing from
// All is therefore silently dropped from unfiltered listings.
//
// Known failure (candidate bug S1-1): Uninstalled is a real transition target
// (StateTransitions[Uninstalling] -> Uninstalled) and carries an
// OperationAllowedInState entry, yet it is absent from All -> apps sitting in
// Uninstalled disappear from the default status listing.
func TestProbe_AllEnumeratesEveryReferencedState(t *testing.T) {
	// Regression for fixed bug S1-1 (Uninstalled was missing from All).
	inAll := make(map[appsv1.ApplicationManagerState]bool, len(All))
	for _, s := range All {
		inAll[s] = true
	}

	refs := map[appsv1.ApplicationManagerState][]string{}
	add := func(s appsv1.ApplicationManagerState, src string) {
		if s == "" {
			return
		}
		refs[s] = append(refs[s], src)
	}
	for k, targets := range StateTransitions {
		add(k, "StateTransitions(from)")
		for _, v := range targets {
			add(v, "StateTransitions(to)")
		}
	}
	for k := range OperationAllowedInState {
		add(k, "OperationAllowedInState")
	}
	for k := range CancelableStates {
		add(k, "CancelableStates")
	}
	for k := range OperatingStates {
		add(k, "OperatingStates")
	}
	for k := range CancelingStates {
		add(k, "CancelingStates")
	}
	for k := range StateToDurationMap {
		add(k, "StateToDurationMap")
	}

	var missing []string
	for s, srcs := range refs {
		if !inAll[s] {
			sort.Strings(srcs)
			missing = append(missing, fmt.Sprintf("  %-24s referenced by %v", s, uniq(srcs)))
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("states referenced by the model but absent from All "+
			"(apiserver uses All as the default state filter, so apps in these states are hidden from unfiltered listings):\n%s",
			strings.Join(missing, "\n"))
	}
}

// TestProbe_GateEntriesReachableViaTransitionTable flags states that carry an
// OperationAllowedInState entry (the ONLY enforced gate, consulted by ~12
// handlers) yet are never produced by any declared transition. Such entries are
// dead or at-risk: either the gate grants operations in a state the app can
// never legitimately reach, or the transition table is missing the edge that
// produces it.
//
// Candidate bug S2-1: UpgradingCancelFailed and ApplyingEnvCancelFailed have
// gate entries but are not transition targets. UpgradingCanceling/
// ApplyingEnvCanceling both declare only {Stopping}. ApplyingEnvCancelFailed is
// at least set imperatively on the canceling error path
// (applying_env_canceling_app.go), but UpgradingCancelFailed has NO producer at
// all (its recovery handler is a commented-out DoNothing) -- a fully dead state
// whose gate entry can never fire.
func TestProbe_GateEntriesReachableViaTransitionTable(t *testing.T) {
	// Regression for fixed bug S2-1 (UpgradingCancelFailed / ApplyingEnvCancelFailed
	// had gate entries but were unreachable via the transition table).
	inbound := map[appsv1.ApplicationManagerState]bool{}
	for _, targets := range StateTransitions {
		for _, to := range targets {
			inbound[to] = true
		}
	}

	var dead []string
	for s := range OperationAllowedInState {
		if s == "" { // the synthetic "app does not exist yet" entry
			continue
		}
		if !inbound[s] {
			dead = append(dead, string(s))
		}
	}
	sort.Strings(dead)
	if len(dead) > 0 {
		t.Fatalf("OperationAllowedInState grants ops in states no transition can produce (dead/at-risk gate entries):\n  %s",
			strings.Join(dead, "\n  "))
	}
}

func uniq(in []string) []string {
	seen := map[string]bool{}
	out := in[:0]
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
