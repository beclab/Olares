package handlers_test

import (
	"context"
	"testing"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/handlers"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

func TestTimeoutHandlers_Table(t *testing.T) {
	cases := []struct {
		name    string
		from    appv1.ApplicationManagerState
		to      appv1.ApplicationManagerState
		reason  string
		message string
		event   string
		record  bool
	}{
		{"downloading", appv1.Downloading, appv1.DownloadingCanceling, constants.DownloadCancelBySystem, constants.DownloadCanceledByTimeout, "Normal", false},
		{"upgradedownloading", appv1.UpgradeDownloading, appv1.UpgradeDownloadingCanceling, constants.UpgradeCancelBySystem, constants.UpgradeCanceledByTimeout, "Normal", false},
		{"initializing", appv1.Initializing, appv1.InitializingCanceling, constants.InitializeCancelBySystem, constants.InitializeCanceledByTimeout, "Normal", false},
		{"upgrading", appv1.Upgrading, appv1.UpgradingCanceling, constants.UpgradeCancelBySystem, constants.UpgradeCanceledByTimeout, "Normal", false},
		{"applyingenv", appv1.ApplyingEnv, appv1.ApplyingEnvCanceling, constants.ApplyEnvCancelBySystem, constants.ApplyEnvCanceledByTimeout, "Normal", false},
		{"resuming", appv1.Resuming, appv1.ResumingCanceling, constants.ResumeCancelBySystem, constants.ResumeCanceledByTimeout, "Normal", false},
		{"stopping", appv1.Stopping, appv1.StopFailed, "StopTimedOut", "stop timed out", "Warning", true},
		{"uninstalling", appv1.Uninstalling, appv1.UninstallFailed, "UninstallTimedOut", "uninstall timed out", "Warning", true},
		{"pendingcanceling", appv1.PendingCanceling, appv1.PendingCancelFailed, "PendingCancelTimedOut", "pending cancel timed out", "Warning", true},
		{"downloadingcanceling", appv1.DownloadingCanceling, appv1.DownloadingCancelFailed, "DownloadingCancelTimedOut", "downloading cancel timed out", "Warning", true},
		{"installingcanceling", appv1.InstallingCanceling, appv1.InstallingCancelFailed, "InstallingCancelTimedOut", "installing cancel timed out", "Warning", true},
		{"upgradingcanceling", appv1.UpgradingCanceling, appv1.UpgradingCancelFailed, "UpgradingCancelTimedOut", "upgrading cancel timed out", "Warning", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			am := &appv1.ApplicationManager{
				ObjectMeta: metav1.ObjectMeta{Name: "app1"},
				Spec:       appv1.ApplicationManagerSpec{AppName: "app1", AppOwner: "alice", Source: "market"},
				Status:     appv1.ApplicationManagerStatus{State: tc.from, OpType: appv1.InstallOp, OpID: "op", Payload: map[string]string{"phase": "x"}},
			}
			res := handlers.OnTimeout(context.Background(), domain.World{AM: am, Now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Action: domain.TimeoutFired{}})
			if res.Err != nil {
				t.Fatalf("Err = %v", res.Err)
			}
			if res.Next != tc.to {
				t.Fatalf("Next = %s, want %s", res.Next, tc.to)
			}
			patch, ok := findEffect[domain.PatchAM](res.Effects)
			if !ok || patch.Status.State != tc.to || patch.Status.Reason != tc.reason || patch.Status.Message != tc.message {
				t.Fatalf("patch = %#v", patch.Status)
			}
			if len(patch.Status.Payload) != 0 {
				t.Fatalf("payload = %#v, want cleared", patch.Status.Payload)
			}
			ev, ok := findEffect[domain.EmitEvent](res.Effects)
			if !ok || ev.Type != tc.event {
				t.Fatalf("event = %#v, want %s", ev, tc.event)
			}
			_, hasRecord := findEffect[domain.AppendOpRecord](res.Effects)
			if hasRecord != tc.record {
				t.Fatalf("op record present = %v, want %v", hasRecord, tc.record)
			}
		})
	}
}

func TestInstallingTimeout_EntersCancelingWithInstallFailedIntent(t *testing.T) {
	am := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "app1"},
		Spec:       appv1.ApplicationManagerSpec{AppName: "app1"},
		Status:     appv1.ApplicationManagerStatus{State: appv1.Installing, Payload: map[string]string{handlers.PayloadInstallingPhase: handlers.PhaseInstallingHelmInstall}},
	}
	res := handlers.OnTimeout(context.Background(), domain.World{AM: am, Action: domain.TimeoutFired{}})
	if res.Next != appv1.InstallingCanceling {
		t.Fatalf("Next = %s, want InstallingCanceling", res.Next)
	}
	patch, ok := findEffect[domain.PatchAM](res.Effects)
	if !ok {
		t.Fatal("no patch")
	}
	if got := patch.Status.Payload[handlers.PayloadInstallingCancelingTerminal]; got != string(appv1.InstallFailed) {
		t.Fatalf("terminal = %q", got)
	}
	if patch.Status.Payload[handlers.PayloadInstallingFailReason] != constants.InstallCancelBySystem ||
		patch.Status.Payload[handlers.PayloadInstallingFailMessage] != constants.InstallCanceledByTimeout {
		t.Fatalf("stashed diagnostic = %#v", patch.Status.Payload)
	}
}

// TestInstallFailed_HasNoTimeout pins that install-failure cleanup is
// never cut short: a timeout would leave InstallFailed while the
// namespace may still be terminating.
func TestInstallFailed_HasNoTimeout(t *testing.T) {
	tab := domain.NewDefaultTable()
	handlers.Register(tab)
	if d, ok := tab.Timeout(appv1.InstallFailed); ok {
		t.Fatalf("Timeout(InstallFailed) = %s, want disabled", d)
	}
	if got := tab.Candidates(appv1.InstallFailed, domain.TimeoutFired{}.Kind()); len(got) != 0 {
		t.Fatalf("InstallFailed has %d TimeoutFired handlers, want none", len(got))
	}
}
