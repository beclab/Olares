package appstate

import (
	"testing"

	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// The init() in transition_guard.go must have populated the apputils-side
// guard variable so every UpdateAppMgrStatus call goes through the same
// transition table as updateStatus.
func TestStateTransitionGuardWiredIntoApputils(t *testing.T) {
	if apputils.StateTransitionGuard == nil {
		t.Fatal("apputils.StateTransitionGuard is nil; pkg/appstate init must wire it up")
	}

	cases := []struct {
		name     string
		from, to appsv1.ApplicationManagerState
		want     bool
	}{
		{"declared edge accepted", appsv1.Installing, appsv1.Running, true},
		{"self-write accepted", appsv1.InstallFailed, appsv1.InstallFailed, true},
		{"undeclared jump rejected", appsv1.Pending, appsv1.Uninstalled, false},
		{"running to installing rejected", appsv1.Running, appsv1.Installing, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := apputils.StateTransitionGuard(c.from, c.to); got != c.want {
				t.Errorf("guard(%s,%s)=%v want %v", c.from, c.to, got, c.want)
			}
		})
	}
}
