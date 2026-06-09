package upgrade

import (
	"testing"

	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/terminus"
)

func TestUpgradeSystemComponents_ReplacesLegacyAppGatewayTasks(t *testing.T) {
	tasks := (upgraderBase{}).UpgradeSystemComponents()
	localByName := localTasksByName(tasks)

	for _, required := range []string{
		"PrepareLinkerdPKI",
		"ApplyNetworkCRDs",
		"UpgradeAppGatewaySystem",
	} {
		if _, ok := localByName[required]; !ok {
			t.Fatalf("missing required task %q in upgrade chain", required)
		}
	}

	for _, removed := range []string{
		"UpgradeAppGatewayVendor",
		"UpgradeAppGatewayChart",
		"WaitAppGatewayDataPlaneMeshed",
	} {
		if _, ok := localByName[removed]; ok {
			t.Fatalf("legacy task %q should not exist in upgrade chain", removed)
		}
	}
}

func TestUpgradeSystemComponents_UpgradeTaskUsesInstallAppGatewaySystemAction(t *testing.T) {
	tasks := (upgraderBase{}).UpgradeSystemComponents()
	localByName := localTasksByName(tasks)
	upgradeTask, ok := localByName["UpgradeAppGatewaySystem"]
	if !ok {
		t.Fatal("missing task UpgradeAppGatewaySystem")
	}

	if _, ok := upgradeTask.Action.(*terminus.InstallAppGatewaySystem); !ok {
		t.Fatalf("unexpected action type %T, want *terminus.InstallAppGatewaySystem", upgradeTask.Action)
	}
}

func localTasksByName(tasks []task.Interface) map[string]*task.LocalTask {
	res := make(map[string]*task.LocalTask, len(tasks))
	for _, item := range tasks {
		if lt, ok := item.(*task.LocalTask); ok {
			res[lt.Name] = lt
		}
	}
	return res
}
