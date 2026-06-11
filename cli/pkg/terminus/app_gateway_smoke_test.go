package terminus

import (
	"testing"

	"github.com/beclab/Olares/cli/pkg/core/task"
)

func TestAppGatewaySystemSingleHelmReleaseName(t *testing.T) {
	if appGatewaySystemReleaseName != "app-gateway-system" {
		t.Fatalf("release name: got %q want app-gateway-system", appGatewaySystemReleaseName)
	}
	if appGatewaySystemDirName != appGatewaySystemReleaseName {
		t.Fatalf("chart dir %q must match release name %q", appGatewaySystemDirName, appGatewaySystemReleaseName)
	}
}

func TestInstallAppGatewaySystemModule_RetryOnConvergence(t *testing.T) {
	initTerminusTestLogger(t)
	mod := &InstallAppGatewaySystemModule{}
	mod.Init()

	var installTask *task.LocalTask
	for _, item := range mod.Tasks {
		lt, ok := item.(*task.LocalTask)
		if !ok {
			continue
		}
		if lt.Name == "InstallAppGatewaySystem" {
			installTask = lt
			break
		}
	}
	if installTask == nil {
		t.Fatal("missing InstallAppGatewaySystem task")
	}
	if installTask.Retry < 1 {
		t.Fatalf("InstallAppGatewaySystem retry: got %d want >= 1 for interrupted install convergence", installTask.Retry)
	}
}
