package job

import "testing"

// TestJobPodSelector pins how `cluster job pods` (and `cluster job
// rerun`) decide which labelSelector to send to the apiserver.
//
// The contract is intentionally simple:
//
//   - When spec.selector.matchLabels is present, build a sorted
//     comma-joined `k=v[,k=v]` clause out of it. spec.selector is the
//     K8s-native source of truth — it's the same selector the Job
//     controller uses to find its child Pods — so honoring it lets
//     the CLI keep working across the K8s 1.27 controller-uid →
//     batch.kubernetes.io/controller-uid relabel (KEP-3850) and
//     across manualSelector=true Jobs (where the operator picks
//     their own keys).
//   - When spec.selector is empty/missing (very old clusters or
//     hand-crafted bare Jobs), fall back to the historical
//     `controller-uid=<metadata.uid>` clause so the verb never
//     silently degrades to a cross-Job listing.
//   - When BOTH are unavailable (no spec.selector AND no UID — a
//     pathological response), return "" so runPods can refuse
//     instead of issuing an unscoped list.
//
// Sorted key order matters: the hint message in runPods echoes the
// selector back to the user, and a stable string makes that hint
// reproducible across runs (Go map iteration is randomized).
func TestJobPodSelector(t *testing.T) {
	tests := []struct {
		name        string
		job         Job
		wantSel     string
		wantSource  string
	}{
		{
			// Modern (K8s 1.27+) auto-generated Job: spec.selector
			// carries batch.kubernetes.io/controller-uid as the
			// authoritative label key. We should send exactly that
			// — NOT the legacy controller-uid — because the new key
			// is the one the controller actually used to claim
			// pods.
			name: "modern K8s 1.27+ selector: batch.kubernetes.io/controller-uid",
			job: Job{
				Metadata: JobMetadata{UID: "abc-123"},
				Spec: JobSpec{
					Selector: &JobSelector{
						MatchLabels: map[string]string{
							"batch.kubernetes.io/controller-uid": "abc-123",
						},
					},
				},
			},
			wantSel:    "batch.kubernetes.io/controller-uid=abc-123",
			wantSource: "spec.selector.matchLabels",
		},
		{
			// Legacy K8s (<1.27) auto-generated Job: spec.selector
			// uses the bare `controller-uid` label. The CLI must
			// still pick the value off spec.selector rather than
			// reconstructing it from metadata.uid so we don't
			// double-up on what the controller already authored.
			name: "legacy K8s selector: controller-uid",
			job: Job{
				Metadata: JobMetadata{UID: "abc-123"},
				Spec: JobSpec{
					Selector: &JobSelector{
						MatchLabels: map[string]string{
							"controller-uid": "abc-123",
						},
					},
				},
			},
			wantSel:    "controller-uid=abc-123",
			wantSource: "spec.selector.matchLabels",
		},
		{
			// manualSelector=true Job: the operator picked an
			// arbitrary key. Most production-grade Jobs don't do
			// this, but the very few that do (custom batch
			// frameworks, hand-tuned analytics jobs) would be
			// completely broken by a hardcoded controller-uid=
			// clause. Reading spec.selector keeps them working.
			name: "manualSelector with custom keys is honored",
			job: Job{
				Metadata: JobMetadata{UID: "abc-123"},
				Spec: JobSpec{
					Selector: &JobSelector{
						MatchLabels: map[string]string{
							"app":  "etl",
							"tier": "batch",
						},
					},
				},
			},
			// Keys sorted alphabetically so the hint string stays
			// reproducible — Go map iteration order is randomized,
			// so a naive range loop would emit a different selector
			// every other run.
			wantSel:    "app=etl,tier=batch",
			wantSource: "spec.selector.matchLabels",
		},
		{
			// Pathological: spec.selector is present but empty
			// (matchLabels={}). Treat it the same as missing —
			// otherwise we'd issue an unscoped pods list which
			// would surface every Pod in the namespace as a
			// "child" of this Job. Falling back to metadata.uid
			// keeps the scope tight.
			name: "empty matchLabels falls back to UID",
			job: Job{
				Metadata: JobMetadata{UID: "abc-123"},
				Spec: JobSpec{
					Selector: &JobSelector{MatchLabels: map[string]string{}},
				},
			},
			wantSel:    "controller-uid=abc-123",
			wantSource: "metadata.uid fallback",
		},
		{
			// Hand-crafted bare Job that somehow has no
			// spec.selector at all (kubectl-applied YAML missing
			// the auto-managed field, or a Job restored from a
			// partial backup). Fall back to UID so we still
			// produce a scoped selector.
			name: "no spec.selector falls back to UID",
			job: Job{
				Metadata: JobMetadata{UID: "abc-123"},
				Spec:     JobSpec{},
			},
			wantSel:    "controller-uid=abc-123",
			wantSource: "metadata.uid fallback",
		},
		{
			// Worst case: neither selector nor UID. runPods must
			// refuse rather than send an unscoped query. Returning
			// "" here is the signal the caller checks for.
			name:       "no selector and no UID returns empty",
			job:        Job{},
			wantSel:    "",
			wantSource: "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotSel, gotSource := jobPodSelector(tc.job)
			if gotSel != tc.wantSel {
				t.Fatalf("selector: got %q, want %q", gotSel, tc.wantSel)
			}
			if gotSource != tc.wantSource {
				t.Fatalf("source: got %q, want %q", gotSource, tc.wantSource)
			}
		})
	}
}
