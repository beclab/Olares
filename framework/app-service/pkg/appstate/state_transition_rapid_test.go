package appstate

import (
	"sort"
	"testing"
	"time"

	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"pgregory.net/rapid"
)

func allStateSet() map[appsv1.ApplicationManagerState]bool {
	m := make(map[appsv1.ApplicationManagerState]bool, len(All))
	for _, s := range All {
		m[s] = true
	}
	return m
}

// recognizedStateSet is the union of every state the package declares anywhere:
// the All enumeration, the transition-table keys and the operation-table keys.
// It is broader than All on purpose -- e.g. Uninstalled is a valid transition
// target and appears in OperationAllowedInState but is missing from All -- so we
// can flag genuine typos without coupling to that gap.
func recognizedStateSet() map[appsv1.ApplicationManagerState]bool {
	m := allStateSet()
	for k := range StateTransitions {
		m[k] = true
	}
	for k := range OperationAllowedInState {
		if k != "" {
			m[k] = true
		}
	}
	return m
}

func sortedTransitionKeys() []appsv1.ApplicationManagerState {
	keys := make([]appsv1.ApplicationManagerState, 0, len(StateTransitions))
	for k := range StateTransitions {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

// A random walk over the declared transition table must always take valid
// transitions and must never reach a state that is not in All.
func TestStateMachineRandomWalkStaysValid(t *testing.T) {
	known := recognizedStateSet()
	keys := sortedTransitionKeys()

	rapid.Check(t, func(rt *rapid.T) {
		cur := keys[rapid.IntRange(0, len(keys)-1).Draw(rt, "start")]
		steps := rapid.IntRange(1, 50).Draw(rt, "steps")
		for i := 0; i < steps; i++ {
			nexts, ok := StateTransitions[cur]
			if !ok || len(nexts) == 0 {
				return
			}
			next := nexts[rapid.IntRange(0, len(nexts)-1).Draw(rt, "next")]
			if !IsStateTransitionValid(cur, next) {
				rt.Fatalf("declared transition %s->%s reported invalid", cur, next)
			}
			if !known[next] {
				rt.Fatalf("transition %s->%s leads to an unrecognized state", cur, next)
			}
			cur = next
		}
	})
}

// The state helpers must never panic and must return safe defaults for states
// or operations they do not know about.
func TestStateHelpersRobustOnArbitraryInput(t *testing.T) {
	known := allStateSet()

	rapid.Check(t, func(rt *rapid.T) {
		s := appsv1.ApplicationManagerState(rapid.String().Draw(rt, "state"))
		op := appsv1.OpType(rapid.String().Draw(rt, "op"))

		_ = IsStateTransitionValid(s, s)
		allowed := IsOperationAllowed(s, op)
		dur := StateToDuration(s)

		if _, isKnownState := OperationAllowedInState[s]; !isKnownState && allowed {
			rt.Fatalf("operation %q must not be allowed in unknown state %q", op, s)
		}
		if _, hasDur := StateToDurationMap[s]; !hasDur && dur != 10*time.Minute {
			rt.Fatalf("unknown state %q duration=%v, want default 10m", s, dur)
		}
		if !known[s] && (IsCancelable(s) || IsCanceling(s)) {
			rt.Fatalf("unknown state %q must not be cancelable/canceling", s)
		}
	})
}

// Every cancelable/canceling state must be an operating state, for any state
// drawn from the declared universe.
func TestCancelStatesAreOperatingProperty(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		s := All[rapid.IntRange(0, len(All)-1).Draw(rt, "state")]
		if IsCancelable(s) && !OperatingStates[s] {
			rt.Fatalf("cancelable state %q is not operating", s)
		}
		if IsCanceling(s) && !OperatingStates[s] {
			rt.Fatalf("canceling state %q is not operating", s)
		}
	})
}
