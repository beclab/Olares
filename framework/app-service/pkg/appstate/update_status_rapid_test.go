package appstate

import (
	"context"
	"strconv"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"pgregory.net/rapid"
)

// Driving updateStatus through a random sequence of states and records against
// a real fake client must always preserve the storage-level invariants:
// OpGeneration increases by exactly one per call, records stay newest-first and
// never exceed the cap of 20.
//
// updateStatus now rejects transitions that are not declared in
// StateTransitions, so the walk picks the next state from
// StateTransitions[current] (or a same-state self-write) instead of any random
// target. Self-writes are exercised deliberately because they are the
// idempotent-retry path that must NOT be rejected.
func TestUpdateStatusInvariantsUnderRandomSequence(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		am := testutil.NewAppManager("nginx-rapid", testutil.WithState(appsv1.Installing))
		c := testutil.NewFakeClient(am)
		b := &baseStatefulApp{manager: am, client: c}

		n := rapid.IntRange(1, 40).Draw(rt, "ops")
		recordsAdded := 0
		cur := appsv1.Installing
		for i := 0; i < n; i++ {
			// Pick a valid next state: either a declared edge or a
			// same-state self-write (both must be accepted).
			nexts := StateTransitions[cur]
			var st appsv1.ApplicationManagerState
			if len(nexts) == 0 || rapid.Bool().Draw(rt, "selfWrite") {
				st = cur
			} else {
				st = nexts[rapid.IntRange(0, len(nexts)-1).Draw(rt, "next")]
			}

			var rec *appsv1.OpRecord
			id := strconv.Itoa(i)
			if rapid.Bool().Draw(rt, "withRecord") {
				now := metav1.Now()
				rec = &appsv1.OpRecord{OpID: id, Message: id, StateTime: &now}
				recordsAdded++
			}

			if err := b.updateStatus(context.TODO(), am, st, rec, id, ""); err != nil {
				rt.Fatalf("updateStatus call %d (%s->%s): %v", i, cur, st, err)
			}
			cur = st

			got := getAM(t, b, "nginx-rapid")
			if got.Status.OpGeneration != int64(i+1) {
				rt.Fatalf("after call %d OpGeneration=%d, want %d", i, got.Status.OpGeneration, i+1)
			}
			if len(got.Status.OpRecords) > 20 {
				rt.Fatalf("after call %d OpRecords=%d exceeds cap 20", i, len(got.Status.OpRecords))
			}
			wantLen := recordsAdded
			if wantLen > 20 {
				wantLen = 20
			}
			if len(got.Status.OpRecords) != wantLen {
				rt.Fatalf("after call %d OpRecords=%d, want %d", i, len(got.Status.OpRecords), wantLen)
			}
			if rec != nil && got.Status.OpRecords[0].OpID != id {
				rt.Fatalf("after call %d newest record=%q, want %q", i, got.Status.OpRecords[0].OpID, id)
			}
		}
	})
}

// Companion test: invalid transitions must be rejected and must NOT mutate
// any storage-level state (no OpGeneration bump, no record prepend).
func TestUpdateStatusRejectsInvalidTransitionAndPreservesInvariants(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		am := testutil.NewAppManager("nginx-rapid-bad", testutil.WithState(appsv1.Pending))
		// Seed an existing record so we can also assert the rejected
		// call did not prepend a new one.
		am.Status.OpRecords = []appsv1.OpRecord{{OpID: "seed"}}
		c := testutil.NewFakeClient(am)
		b := &baseStatefulApp{manager: am, client: c}

		// Build the set of valid targets from Pending (incl. self) so we
		// can draw an arbitrary INVALID target.
		valid := map[appsv1.ApplicationManagerState]bool{appsv1.Pending: true}
		for _, s := range StateTransitions[appsv1.Pending] {
			valid[s] = true
		}
		// All non-empty states from the known universe.
		candidates := make([]appsv1.ApplicationManagerState, 0, len(All))
		for _, s := range All {
			if !valid[s] {
				candidates = append(candidates, s)
			}
		}
		if len(candidates) == 0 {
			return
		}
		target := candidates[rapid.IntRange(0, len(candidates)-1).Draw(rt, "invalidTarget")]

		err := b.updateStatus(context.TODO(), am, target, &appsv1.OpRecord{OpID: "rejected"}, "msg", "")
		if err == nil {
			rt.Fatalf("updateStatus(Pending -> %s) returned nil, want error", target)
		}

		got := getAM(t, b, "nginx-rapid-bad")
		if got.Status.State != appsv1.Pending {
			rt.Fatalf("state=%s want Pending (rejected write must not mutate)", got.Status.State)
		}
		if got.Status.OpGeneration != 0 {
			rt.Fatalf("OpGeneration=%d want 0 (rejected write must not bump)", got.Status.OpGeneration)
		}
		if len(got.Status.OpRecords) != 1 || got.Status.OpRecords[0].OpID != "seed" {
			rt.Fatalf("OpRecords=%+v want only the seed record (rejected write must not prepend)", got.Status.OpRecords)
		}
	})
}
