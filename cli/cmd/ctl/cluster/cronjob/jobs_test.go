package cronjob

import (
	"testing"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/job"
)

// TestIsChildOfCronJob pins the parent/child predicate that drives
// `cluster cronjob jobs`. The earlier label-only implementation
// silently mis-attributed Jobs across CronJobs that happened to
// share labels (or that were manually created with the same
// labels) — switching to ownerReferences fixes that, and this
// test keeps the contract honest:
//
//   - UID equality is the ONLY safe match (Name+Kind would attribute
//     a renamed/recreated CronJob to its predecessor's children).
//   - Controller=true is required: a non-controller owner ref is a
//     "soft" reference written by some other actor and must not be
//     treated as the parent.
//   - Kind, when set on the ref, must be "CronJob"; an empty Kind
//     is tolerated for the K8s servers that omit it on internal
//     references.
//
// Each case below maps to a specific real-world miss-match the
// predicate has to either accept or reject — see the per-case
// comments for the rationale.
func TestIsChildOfCronJob(t *testing.T) {
	const parentUID = "cron-uid-aaa"

	tests := []struct {
		name string
		refs []job.JobOwnerRef
		uid  string
		want bool
	}{
		{
			// No owner refs at all — manually-created Job. Must NOT
			// be attributed to any CronJob.
			name: "no owner refs",
			refs: nil,
			uid:  parentUID,
			want: false,
		},
		{
			// Happy path: the CronJob controller stamps a Job with
			// a Controller=true OwnerReference whose UID matches.
			name: "controller owner ref with matching UID",
			refs: []job.JobOwnerRef{
				{Kind: "CronJob", UID: parentUID, Controller: true},
			},
			uid:  parentUID,
			want: true,
		},
		{
			// UID mismatch — a Job spawned by a DIFFERENT CronJob in
			// the same namespace. The label pre-narrow might still
			// hand it back if labels collide; this filter catches it.
			name: "controller owner ref with non-matching UID",
			refs: []job.JobOwnerRef{
				{Kind: "CronJob", UID: "cron-uid-bbb", Controller: true},
			},
			uid:  parentUID,
			want: false,
		},
		{
			// Non-controller owner ref — e.g. a label-based "soft"
			// reference written by a third-party operator. Must be
			// rejected even if the UID happens to match.
			name: "non-controller owner ref",
			refs: []job.JobOwnerRef{
				{Kind: "CronJob", UID: parentUID, Controller: false},
			},
			uid:  parentUID,
			want: false,
		},
		{
			// Foreign Kind sharing a UID is extraordinarily unlikely
			// in practice (apiserver assigns UIDs across the cluster)
			// but pin the defensive check anyway.
			name: "owner ref with wrong Kind",
			refs: []job.JobOwnerRef{
				{Kind: "Job", UID: parentUID, Controller: true},
			},
			uid:  parentUID,
			want: false,
		},
		{
			// Some apiserver flavors omit Kind on internal refs;
			// tolerate empty Kind so we don't silently miss
			// legitimate children.
			name: "owner ref with empty Kind tolerated",
			refs: []job.JobOwnerRef{
				{Kind: "", UID: parentUID, Controller: true},
			},
			uid:  parentUID,
			want: true,
		},
		{
			// Multi-owner Jobs are valid K8s (one Controller=true
			// plus zero-to-many soft refs). The predicate must scan
			// past the soft refs to find the controller.
			name: "multi-owner: soft ref first, controller second",
			refs: []job.JobOwnerRef{
				{Kind: "Foo", UID: "foo-uid", Controller: false},
				{Kind: "CronJob", UID: parentUID, Controller: true},
			},
			uid:  parentUID,
			want: true,
		},
		{
			// Empty parentUID is a misuse — the caller should have
			// guarded on `c.Metadata.UID == ""`. The predicate
			// refuses to match anything in that case to keep the
			// failure visible rather than letting it cascade into a
			// silent "match everything".
			name: "empty parentUID never matches",
			refs: []job.JobOwnerRef{
				{Kind: "CronJob", UID: "", Controller: true},
			},
			uid:  "",
			want: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			j := job.Job{Metadata: job.JobMetadata{OwnerReferences: tc.refs}}
			got := isChildOfCronJob(j, tc.uid)
			if got != tc.want {
				t.Fatalf("isChildOfCronJob(uid=%q, refs=%+v): got %v, want %v",
					tc.uid, tc.refs, got, tc.want)
			}
		})
	}
}
