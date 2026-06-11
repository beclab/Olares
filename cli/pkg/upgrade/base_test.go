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

func TestUpgradeSystemComponents_AppGatewayTaskOrder(t *testing.T) {
	tasks := (upgraderBase{}).UpgradeSystemComponents()
	names := taskNames(tasks)

	pkiIdx := indexOf(names, "PrepareLinkerdPKI")
	crdIdx := indexOf(names, "ApplyNetworkCRDs")
	upgradeIdx := indexOf(names, "UpgradeAppGatewaySystem")
	if pkiIdx < 0 || crdIdx < 0 || upgradeIdx < 0 {
		t.Fatalf("missing app-gateway upgrade tasks in %v", names)
	}
	if !(pkiIdx < crdIdx && crdIdx < upgradeIdx) {
		t.Fatalf("app-gateway upgrade order: got %v want PrepareLinkerdPKI before ApplyNetworkCRDs before UpgradeAppGatewaySystem", names)
	}
}

func TestUpgradeSystemComponents_PrepareLinkerdPKIPreservesSecretOnUpgrade(t *testing.T) {
	tasks := (upgraderBase{}).UpgradeSystemComponents()
	localByName := localTasksByName(tasks)
	pkiTask, ok := localByName["PrepareLinkerdPKI"]
	if !ok {
		t.Fatal("missing PrepareLinkerdPKI in upgrade chain")
	}
	if _, ok := pkiTask.Action.(*terminus.PrepareLinkerdPKI); !ok {
		t.Fatalf("unexpected action type %T, want *terminus.PrepareLinkerdPKI", pkiTask.Action)
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

func taskNames(tasks []task.Interface) []string {
	names := make([]string, 0, len(tasks))
	for _, item := range tasks {
		if lt, ok := item.(*task.LocalTask); ok {
			names = append(names, lt.Name)
		}
	}
	return names
}

func indexOf(names []string, target string) int {
	for i, name := range names {
		if name == target {
			return i
		}
	}
	return -1
}
