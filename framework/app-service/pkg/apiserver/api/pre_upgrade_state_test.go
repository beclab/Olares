package api

import (
	"testing"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

func TestSnapshotPreUpgradeState(t *testing.T) {
	cases := []struct {
		name     string
		state    appv1alpha1.ApplicationManagerState
		existing string
		want     string
	}{
		{
			name:  "running writes running",
			state: appv1alpha1.Running,
			want:  string(appv1alpha1.Running),
		},
		{
			name:  "stopped writes stopped",
			state: appv1alpha1.Stopped,
			want:  string(appv1alpha1.Stopped),
		},
		{
			name:     "running refreshes over stale stopped",
			state:    appv1alpha1.Running,
			existing: string(appv1alpha1.Stopped),
			want:     string(appv1alpha1.Running),
		},
		{
			name:     "upgradeFailed keeps prior stopped",
			state:    appv1alpha1.UpgradeFailed,
			existing: string(appv1alpha1.Stopped),
			want:     string(appv1alpha1.Stopped),
		},
		{
			name:     "upgradeFailed keeps prior running",
			state:    appv1alpha1.UpgradeFailed,
			existing: string(appv1alpha1.Running),
			want:     string(appv1alpha1.Running),
		},
		{
			name:  "upgradeFailed with no prior defaults to running",
			state: appv1alpha1.UpgradeFailed,
			want:  string(appv1alpha1.Running),
		},
		{
			name:     "applyEnvFailed keeps prior stopped",
			state:    appv1alpha1.ApplyEnvFailed,
			existing: string(appv1alpha1.Stopped),
			want:     string(appv1alpha1.Stopped),
		},
		{
			name:     "resumeFailed keeps prior stopped",
			state:    appv1alpha1.ResumeFailed,
			existing: string(appv1alpha1.Stopped),
			want:     string(appv1alpha1.Stopped),
		},
		{
			name:     "upgradeFailed with garbage prior defaults to running",
			state:    appv1alpha1.UpgradeFailed,
			existing: "upgradeFailed",
			want:     string(appv1alpha1.Running),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ann := map[string]string{}
			if tc.existing != "" {
				ann[AppPreUpgradeStateKey] = tc.existing
			}
			SnapshotPreUpgradeState(ann, tc.state)
			if got := ann[AppPreUpgradeStateKey]; got != tc.want {
				t.Fatalf("annotation=%q want %q", got, tc.want)
			}
		})
	}
}

func TestSnapshotPreUpgradeStateNilAnnotations(t *testing.T) {
	// Must not panic when callers have not yet allocated the map.
	SnapshotPreUpgradeState(nil, appv1alpha1.Stopped)
}
