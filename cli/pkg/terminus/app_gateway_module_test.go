package terminus

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

func initTerminusTestLogger(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	logger.InitLog(dir, filepath.Join(dir, "console.log"), true)
}

func TestInstallAppGatewaySystemModule_TaskOrder(t *testing.T) {
	initTerminusTestLogger(t)
	mod := &InstallAppGatewaySystemModule{}
	mod.Init()

	want := []string{
		"ValidateAppGatewaySystemInstaller",
		"PrepareLinkerdPKI",
		"ApplyNetworkCRDs",
		"InstallAppGatewaySystem",
		"CheckAppGatewayControlPlaneReady",
	}
	if len(mod.Tasks) != len(want) {
		t.Fatalf("task count: got %d want %d", len(mod.Tasks), len(want))
	}
	for i, name := range want {
		lt, ok := mod.Tasks[i].(*task.LocalTask)
		if !ok {
			t.Fatalf("task %d: got %T, want *task.LocalTask", i, mod.Tasks[i])
		}
		if lt.Name != name {
			t.Fatalf("task %d: got %q want %q", i, lt.Name, name)
		}
	}

	retryByName := map[string]int{
		"PrepareLinkerdPKI":           2,
		"ApplyNetworkCRDs":            2,
		"InstallAppGatewaySystem":     2,
		"CheckAppGatewayControlPlaneReady": 20,
	}
	delayByName := map[string]time.Duration{
		"PrepareLinkerdPKI":              30 * time.Second,
		"ApplyNetworkCRDs":               30 * time.Second,
		"InstallAppGatewaySystem":        30 * time.Second,
		"CheckAppGatewayControlPlaneReady": 10 * time.Second,
	}
	for _, lt := range mod.Tasks {
		task, ok := lt.(*task.LocalTask)
		if !ok {
			continue
		}
		if wantRetry, ok := retryByName[task.Name]; ok && task.Retry != wantRetry {
			t.Fatalf("task %q retry: got %d want %d", task.Name, task.Retry, wantRetry)
		}
		if wantDelay, ok := delayByName[task.Name]; ok && task.Delay != wantDelay {
			t.Fatalf("task %q delay: got %s want %s", task.Name, task.Delay, wantDelay)
		}
	}
}
