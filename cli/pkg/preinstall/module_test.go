package preinstall

import (
	"errors"
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/storage"
)

func TestMaterializeModuleUsesProvidedPathsAndSelections(t *testing.T) {
	selections := ProfileSelections{HardwareProfile: "fixture"}
	module := &MaterializeModule{
		InstallerDir:      "/installer",
		RootDir:           "/olares",
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
	if action.InstallerDir != "/installer" || action.RootDir != "/olares" ||
		action.ProfileSelections.HardwareProfile != "fixture" {
		t.Fatalf("action = %#v", action)
	}
}

func TestMaterializeRuntimeDirIsRelativeToOlaresRoot(t *testing.T) {
	// The market chart mounts {{ .Values.rootPath }}/RuntimeRelativeDir, and
	// rootPath is rendered from storage.OlaresRootDir. Publishing anywhere else
	// (notably under the installer base directory) makes the pod mount an empty
	// directory and silently disables offline preinstall.
	if !strings.HasPrefix(storage.OlaresRootDir, "/") {
		t.Fatalf("olares root = %q, want an absolute host path", storage.OlaresRootDir)
	}
	if strings.HasPrefix(RuntimeRelativeDir, "/") {
		t.Fatalf("runtime dir = %q, want a path relative to the olares root", RuntimeRelativeDir)
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
