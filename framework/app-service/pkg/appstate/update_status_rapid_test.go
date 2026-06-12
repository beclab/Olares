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
func TestUpdateStatusInvariantsUnderRandomSequence(t *testing.T) {
	candidateStates := []appsv1.ApplicationManagerState{
		appsv1.Installing,
		appsv1.Initializing,
		appsv1.Running,
		appsv1.Stopping,
		appsv1.Stopped,
		appsv1.Upgrading,
		appsv1.ApplyingEnv,
		appsv1.Uninstalling,
	}

	rapid.Check(t, func(rt *rapid.T) {
		am := testutil.NewAppManager("nginx-rapid", testutil.WithState(appsv1.Installing))
		c := testutil.NewFakeClient(am)
		b := &baseStatefulApp{manager: am, client: c}

		n := rapid.IntRange(1, 40).Draw(rt, "ops")
		recordsAdded := 0
		for i := 0; i < n; i++ {
			st := candidateStates[rapid.IntRange(0, len(candidateStates)-1).Draw(rt, "state")]

			var rec *appsv1.OpRecord
			id := strconv.Itoa(i)
			if rapid.Bool().Draw(rt, "withRecord") {
				now := metav1.Now()
				rec = &appsv1.OpRecord{OpID: id, Message: id, StateTime: &now}
				recordsAdded++
			}

			if err := b.updateStatus(context.TODO(), am, st, rec, id, ""); err != nil {
				rt.Fatalf("updateStatus call %d: %v", i, err)
			}

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
