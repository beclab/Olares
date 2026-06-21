package appstate

import (
	"testing"
	"time"

	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

func TestIsStateTransitionValid(t *testing.T) {
	cases := []struct {
		from, to appsv1.ApplicationManagerState
		want     bool
	}{
		{appsv1.Pending, appsv1.Downloading, true},
		{appsv1.Downloading, appsv1.Installing, true},
		{appsv1.Installing, appsv1.Initializing, true},
		{appsv1.Installing, appsv1.InstallFailed, true},
		{appsv1.InstallFailed, appsv1.Uninstalling, true},
		{appsv1.Initializing, appsv1.Running, true},
		{appsv1.Running, appsv1.Uninstalling, true},
		{appsv1.Uninstalling, appsv1.Uninstalled, true},

		// Edges that were previously missing from the table but happen at
		// runtime — must now be reported valid.
		{"", appsv1.Pending, true},
		{appsv1.Installing, appsv1.Running, true},
		{appsv1.Upgrading, appsv1.Stopped, true},
		{appsv1.Running, appsv1.Uninstalled, true},
		{appsv1.UninstallFailed, appsv1.Uninstalled, true},
		{appsv1.InitializingCanceling, appsv1.InstallingCancelFailed, true},
		{appsv1.ResumingCanceling, appsv1.ResumingCancelFailed, true},
		{appsv1.Resuming, appsv1.Stopping, true},
		{appsv1.ResumeFailed, appsv1.Upgrading, true},
		{appsv1.ResumeFailed, appsv1.ApplyingEnv, true},
		{appsv1.InstallingCancelFailed, appsv1.Uninstalling, true},
		{appsv1.UpgradingCancelFailed, appsv1.UpgradingCanceling, true},
		{appsv1.UpgradingCancelFailed, appsv1.Uninstalling, true},
		{appsv1.ApplyingEnvCancelFailed, appsv1.ApplyingEnvCanceling, true},
		{appsv1.ApplyingEnvCancelFailed, appsv1.Uninstalling, true},
		{appsv1.PendingCanceled, appsv1.Pending, true},
		{appsv1.DownloadingCanceled, appsv1.Pending, true},
		{appsv1.InstallingCanceled, appsv1.Pending, true},
		{appsv1.Uninstalled, appsv1.Pending, true},
		// UpgradeFailed allows InstallOp as a recovery path (release lost).
		{appsv1.UpgradeFailed, appsv1.Pending, true},

		// invalid jumps (still rejected after the table was augmented).
		{appsv1.Pending, appsv1.Running, false},
		{appsv1.Running, appsv1.Installing, false},
		{appsv1.Uninstalled, appsv1.Running, false},
		{appsv1.Initializing, appsv1.Uninstalled, false},
	}
	for _, c := range cases {
		if got := IsStateTransitionValid(c.from, c.to); got != c.want {
			t.Errorf("IsStateTransitionValid(%s,%s)=%v want %v", c.from, c.to, got, c.want)
		}
	}
}

// IsStateTransitionAllowed is the runtime guard used by updateStatus. It must
// pass everything IsStateTransitionValid passes AND additionally allow
// same-state (from == to) writes so idempotent retries and re-assertions are
// not rejected.
func TestIsStateTransitionAllowedAllowsSelfWrite(t *testing.T) {
	selfWrites := []appsv1.ApplicationManagerState{
		appsv1.Pending,
		appsv1.Running,
		appsv1.InstallFailed,
		appsv1.Uninstalled,
		// Self-write must work even for states with NO declared outgoing
		// edge (e.g. ResumingCancelFailed today) — otherwise an idempotent
		// re-assertion of a stuck terminal state could not refresh its
		// message/reason.
		appsv1.ResumingCancelFailed,
	}
	for _, s := range selfWrites {
		if !IsStateTransitionAllowed(s, s) {
			t.Errorf("IsStateTransitionAllowed(%s,%s)=false, want true (self-write)", s, s)
		}
	}

	// Disallowed jumps remain disallowed.
	if IsStateTransitionAllowed(appsv1.Pending, appsv1.Running) {
		t.Error("IsStateTransitionAllowed(Pending,Running)=true, want false")
	}
	if IsStateTransitionAllowed(appsv1.Initializing, appsv1.Uninstalled) {
		t.Error("IsStateTransitionAllowed(Initializing,Uninstalled)=true, want false")
	}
}

// Every (state, op) pair declared in OperationAllowedInState that has an
// obvious target state (via the handler that drives it) must also appear in
// StateTransitions. This invariant catches the historical drift between the
// two tables (e.g. OperationAllowedInState allowing UninstallOp from a state
// while StateTransitions had no edge to Uninstalling).
func TestOperationAllowedAlignsWithStateTransitions(t *testing.T) {
	// op -> target state driven by the corresponding handler/state-machine
	// path. CancelOp is intentionally excluded: the target canceling state
	// depends on the source state (Installing -> InstallingCanceling,
	// Resuming -> ResumingCanceling, etc.), so it is covered case-by-case
	// elsewhere.
	opTarget := map[appsv1.OpType]appsv1.ApplicationManagerState{
		appsv1.InstallOp:   appsv1.Pending,
		appsv1.UninstallOp: appsv1.Uninstalling,
		appsv1.UpgradeOp:   appsv1.Upgrading,
		appsv1.ApplyEnvOp:  appsv1.ApplyingEnv,
		appsv1.ResumeOp:    appsv1.Resuming,
		appsv1.StopOp:      appsv1.Stopping,
	}
	for state, ops := range OperationAllowedInState {
		for op, allowed := range ops {
			if !allowed {
				continue
			}
			to, ok := opTarget[op]
			if !ok {
				continue
			}
			if !IsStateTransitionValid(state, to) {
				t.Errorf("OperationAllowedInState[%s][%s]=true but StateTransitions[%s] has no edge to %s",
					state, op, state, to)
			}
		}
	}
}

// Every transition declared in the table must be reported valid by the helper.
func TestStateTransitionsSelfConsistent(t *testing.T) {
	for from, tos := range StateTransitions {
		for _, to := range tos {
			if !IsStateTransitionValid(from, to) {
				t.Errorf("declared transition %s->%s not reported valid", from, to)
			}
		}
	}
}

func TestIsOperationAllowed(t *testing.T) {
	cases := []struct {
		state appsv1.ApplicationManagerState
		op    appsv1.OpType
		want  bool
	}{
		{"", appsv1.InstallOp, true},
		{"", appsv1.UninstallOp, false},
		{appsv1.Running, appsv1.UninstallOp, true},
		{appsv1.Running, appsv1.UpgradeOp, true},
		{appsv1.Running, appsv1.InstallOp, false},
		{appsv1.Uninstalling, appsv1.UninstallOp, false},
		{appsv1.InstallFailed, appsv1.InstallOp, true},
		{appsv1.InstallFailed, appsv1.UninstallOp, true},
		{appsv1.UpgradeFailed, appsv1.InstallOp, true},
		{appsv1.UpgradeFailed, appsv1.UpgradeOp, true},
		{appsv1.Pending, appsv1.CancelOp, true},
		{appsv1.Pending, appsv1.UninstallOp, false},
	}
	for _, c := range cases {
		if got := IsOperationAllowed(c.state, c.op); got != c.want {
			t.Errorf("IsOperationAllowed(%q,%q)=%v want %v", c.state, c.op, got, c.want)
		}
	}
}

func TestIsCancelableAndCanceling(t *testing.T) {
	if !IsCancelable(appsv1.Installing) {
		t.Error("Installing should be cancelable")
	}
	if IsCancelable(appsv1.Running) {
		t.Error("Running should not be cancelable")
	}
	if !IsCanceling(appsv1.InstallingCanceling) {
		t.Error("InstallingCanceling should be a canceling state")
	}
	if IsCanceling(appsv1.Installing) {
		t.Error("Installing should not be a canceling state")
	}
}

func TestStateToDuration(t *testing.T) {
	if got := StateToDuration(appsv1.Installing); got != 30*time.Minute {
		t.Errorf("Installing duration=%v want 30m", got)
	}
	if got := StateToDuration(appsv1.Downloading); got != 30*24*time.Hour {
		t.Errorf("Downloading duration=%v want 720h", got)
	}
	// unknown state falls back to the 10 minute default.
	if got := StateToDuration(appsv1.Running); got != 10*time.Minute {
		t.Errorf("default duration=%v want 10m", got)
	}
}

// Cancelable and Canceling states must all be operating states.
func TestCancelableAndCancelingAreOperating(t *testing.T) {
	for s := range CancelableStates {
		if !OperatingStates[s] {
			t.Errorf("cancelable state %s is not an operating state", s)
		}
	}
	for s := range CancelingStates {
		if !OperatingStates[s] {
			t.Errorf("canceling state %s is not an operating state", s)
		}
	}
}

func TestIsTerminalReinstallable(t *testing.T) {
	cases := []struct {
		state appsv1.ApplicationManagerState
		want  bool
	}{
		{appsv1.Uninstalled, true},
		{appsv1.InstallingCanceled, true},
		{appsv1.PendingCanceled, true},
		{appsv1.DownloadingCanceled, true},
		{appsv1.DownloadFailed, true},
		{appsv1.InstallFailed, true},

		{appsv1.Running, false},
		{appsv1.Installing, false},
		{appsv1.InstallingCanceling, false},
		{appsv1.UninstallFailed, false},
		{appsv1.UpgradeFailed, false},
		{appsv1.Stopped, false},
		{"", false},
	}
	for _, c := range cases {
		if got := IsTerminalReinstallable(c.state); got != c.want {
			t.Errorf("IsTerminalReinstallable(%q)=%v want %v", c.state, got, c.want)
		}
	}
}

func TestIsSafelyDeletable(t *testing.T) {
	cases := []struct {
		state appsv1.ApplicationManagerState
		want  bool
	}{
		{appsv1.Uninstalled, true},
		{appsv1.InstallingCanceled, true},
		{appsv1.InstallFailed, true},
		{appsv1.PendingCanceled, true},
		{appsv1.DownloadingCanceled, true},
		{appsv1.DownloadFailed, true},

		// Active / canceling / non-clean failure states.
		{appsv1.Running, false},
		{appsv1.Installing, false},
		{appsv1.InstallingCanceling, false},
		{appsv1.UninstallFailed, false},
		{appsv1.UpgradeFailed, false},
		{appsv1.Stopped, false},
		{appsv1.Stopping, false},
		{appsv1.Pending, false},
		{appsv1.PendingCancelFailed, false},
		{appsv1.InstallingCancelFailed, false},
		{"", false},
	}
	for _, c := range cases {
		if got := IsSafelyDeletable(c.state); got != c.want {
			t.Errorf("IsSafelyDeletable(%q)=%v want %v", c.state, got, c.want)
		}
	}
}

// Every IsSafelyDeletable state must also be IsTerminalReinstallable — the
// "deletable" set is a strict subset of the "reinstallable" set, since you
// can only safely delete a CR you'd otherwise allow a reinstall on.
func TestSafelyDeletableImpliesReinstallable(t *testing.T) {
	for _, s := range All {
		if IsSafelyDeletable(s) && !IsTerminalReinstallable(s) {
			t.Errorf("state %q is IsSafelyDeletable but not IsTerminalReinstallable", s)
		}
	}
}

// Every IsTerminalReinstallable state must allow InstallOp; the two predicates
// describe the same lifecycle decision from different angles and they must
// agree, otherwise a stale AM can be re-targeted but the state-machine will
// reject the request, or vice versa.
func TestTerminalReinstallableAllowsInstallOp(t *testing.T) {
	for _, s := range All {
		if IsTerminalReinstallable(s) && !IsOperationAllowed(s, appsv1.InstallOp) {
			t.Errorf("state %q is IsTerminalReinstallable but disallows InstallOp", s)
		}
	}
}
