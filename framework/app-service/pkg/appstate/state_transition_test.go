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
		{appsv1.Initializing, appsv1.Running, true},
		{appsv1.Running, appsv1.Uninstalling, true},
		{appsv1.Uninstalling, appsv1.Uninstalled, true},
		// invalid jumps
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
