package cronjob

import "testing"

// TestBuildGetPath pins the K8s native CronJob endpoint the package
// uses. Until 2025-Q? the path was `apis/batch/v1beta1/...`, but
// `batch/v1beta1` was removed in K8s 1.25 and the Olares cluster
// only serves `batch/v1`. A user-reported wave of 404s on `cronjob
// get / jobs / suspend / yaml` traced back to the stale path — this
// test keeps the regression visible if anyone reverts the migration.
//
// The path is shared by `cronjob get`, `cronjob yaml`, `cronjob
// suspend`, and `cronjob resume` (suspend/resume PATCH the same URL
// `cronjob get` GETs), so pinning the helper is enough — the verbs
// can't drift independently.
func TestBuildGetPath(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		objName   string
		want      string
	}{
		{
			name:      "plain ascii",
			namespace: "default",
			objName:   "backup-cron",
			want:      "/apis/batch/v1/namespaces/default/cronjobs/backup-cron",
		},
		{
			name:      "kube-system namespace, dotted name",
			namespace: "kube-system",
			objName:   "snapshot.daily",
			want:      "/apis/batch/v1/namespaces/kube-system/cronjobs/snapshot.daily",
		},
		{
			// Slashes in either segment must be percent-encoded so
			// the path can't be smuggled past the apiserver into a
			// different resource. The verbs rely on url.PathEscape
			// for that — pin it here in case someone "simplifies"
			// the helper later.
			name:      "name with slash gets percent-escaped",
			namespace: "default",
			objName:   "tricky/name",
			want:      "/apis/batch/v1/namespaces/default/cronjobs/tricky%2Fname",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildGetPath(tc.namespace, tc.objName)
			if got != tc.want {
				t.Fatalf("buildGetPath(%q, %q):\n  got:  %q\n  want: %q",
					tc.namespace, tc.objName, got, tc.want)
			}
		})
	}
}
