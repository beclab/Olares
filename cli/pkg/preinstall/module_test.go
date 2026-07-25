package preinstall

import (
	"errors"
	"testing"

	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/storage"
)

func TestMaterializeModuleUsesProvidedPathsAndSelections(t *testing.T) {
	selections := ProfileSelections{HardwareProfile: "fixture"}
	module := &MaterializeModule{
		InstallerDir:      "/installer",
		BaseDir:           "/base",
		ProfileSelections: selections,
	}

	module.Init()

	if module.GetName() != "MaterializeMarketPreinstall" {
		t.Fatalf("module name = %q", module.GetName())
	}
	tasks := module.GetTasks()
	if len(tasks) != 1 || tasks[0].GetName() != "MaterializeMarketPreinstall" {
		t.Fatalf("tasks = %#v", tasks)
	}
	localTask, ok := tasks[0].(*task.LocalTask)
	if !ok {
		t.Fatalf("task type = %T", tasks[0])
	}
	action, ok := localTask.Action.(*MaterializeAction)
	if !ok {
		t.Fatalf("action type = %T", localTask.Action)
	}
	if action.InstallerDir != "/installer" || action.BaseDir != "/base" ||
		action.ProfileSelections.HardwareProfile != "fixture" {
		t.Fatalf("action = %#v", action)
	}
}

func TestHFCacheMaterializeModuleUsesRuntimeInstallerAndStorageTarget(t *testing.T) {
	module := &HFCacheMaterializeModule{}

	module.Init()

	if module.GetName() != "MaterializeHuggingFaceCache" {
		t.Fatalf("module name = %q", module.GetName())
	}
	tasks := module.GetTasks()
	if len(tasks) != 1 || tasks[0].GetName() != "MaterializeHuggingFaceCache" {
		t.Fatalf("tasks = %#v", tasks)
	}
	localTask, ok := tasks[0].(*task.LocalTask)
	if !ok {
		t.Fatalf("task type = %T", tasks[0])
	}
	action, ok := localTask.Action.(*HFCacheMaterializeAction)
	if !ok {
		t.Fatalf("action type = %T", localTask.Action)
	}
	wantErr := errors.New("fixture failure")
	var gotInstaller, gotTarget string
	action.materialize = func(installer, target string, owner *hfOwnership) error {
		gotInstaller, gotTarget = installer, target
		if owner == nil {
			t.Fatal("module must enable recursive ownership")
		}
		return wantErr
	}
	runtime := &installerOnlyRuntime{installerDir: "/installer"}

	err := action.Execute(runtime)

	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v", err)
	}
	if gotInstaller != "/installer" || gotTarget != storage.HuggingFaceCacheDir {
		t.Fatalf("materialize paths = %q, %q", gotInstaller, gotTarget)
	}
}

type installerOnlyRuntime struct {
	connector.Runtime
	installerDir string
}

func (r *installerOnlyRuntime) GetInstallerDir() string {
	return r.installerDir
}
