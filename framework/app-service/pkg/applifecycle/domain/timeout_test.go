package domain

import (
	"testing"
	"time"

	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

func TestDefaultTable_TimeoutOnlyForCallWorkloadStates(t *testing.T) {
	tab := NewDefaultTable()
	want := map[State]time.Duration{
		appsv1.Downloading:                 24 * time.Hour,
		appsv1.UpgradeDownloading:          24 * time.Hour,
		appsv1.Installing:                  30 * time.Minute,
		appsv1.Initializing:                time.Hour,
		appsv1.Upgrading:                   time.Hour,
		appsv1.ApplyingEnv:                 30 * time.Minute,
		appsv1.Resuming:                    time.Hour,
		appsv1.Stopping:                    30 * time.Minute,
		appsv1.Uninstalling:                30 * time.Minute,
		appsv1.PendingCanceling:            10 * time.Minute,
		appsv1.DownloadingCanceling:        10 * time.Minute,
		appsv1.UpgradePendingCanceling:     10 * time.Minute,
		appsv1.UpgradeDownloadingCanceling: 10 * time.Minute,
		appsv1.InstallingCanceling:         10 * time.Minute,
		appsv1.UpgradingCanceling:          10 * time.Minute,
	}
	for state, ttl := range want {
		got, ok := tab.Timeout(state)
		if !ok || got != ttl {
			t.Errorf("Timeout(%s) = (%v, %v), want (%v, true)", state, got, ok, ttl)
		}
	}
	for _, state := range []State{
		appsv1.Pending,
		appsv1.UpgradePending,
		appsv1.InitializingCanceling,
		appsv1.ResumingCanceling,
		appsv1.ApplyingEnvCanceling,
		appsv1.Running,
		appsv1.Stopped,
		appsv1.Uninstalled,
		appsv1.InstallFailed,
	} {
		if got, ok := tab.Timeout(state); ok {
			t.Errorf("Timeout(%s) = (%v, true), want disabled", state, got)
		}
	}
}
