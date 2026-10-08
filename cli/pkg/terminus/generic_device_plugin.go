package terminus

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/pkg/errors"
)

const genericDevicePluginManifest = "wizard/config/generic-device-plugin/generic-device-plugin.yaml"

// GenericDevicePluginTask returns the shared install/upgrade task for the
// platform generic-device-plugin. Keeping the apply and rollout check in one
// task makes both paths idempotent and prevents their manifests from drifting.
func GenericDevicePluginTask() task.Interface {
	return &task.LocalTask{
		Name:    "InstallGenericDevicePlugin",
		Action:  new(InstallGenericDevicePlugin),
		Retry:   3,
		Delay:   5 * time.Second,
		Timeout: 5 * time.Minute,
	}
}

// InstallGenericDevicePluginModule wires the shared task into a fresh Linux
// installation after Kubernetes is available.
type InstallGenericDevicePluginModule struct {
	common.KubeModule
}

func (m *InstallGenericDevicePluginModule) Init() {
	m.Name = "InstallGenericDevicePlugin"
	m.Tasks = []task.Interface{GenericDevicePluginTask()}
}

var _ module.Module = (*InstallGenericDevicePluginModule)(nil)

// InstallGenericDevicePlugin applies the complete RBAC and DaemonSet manifest
// and does not return until the rollout is ready.
type InstallGenericDevicePlugin struct {
	common.KubeAction
}

func (t *InstallGenericDevicePlugin) Execute(runtime connector.Runtime) error {
	manifestPath := filepath.Join(runtime.GetInstallerDir(), genericDevicePluginManifest)
	if !util.IsExist(manifestPath) {
		return fmt.Errorf("generic-device-plugin manifest not found at %s", manifestPath)
	}

	kubectl, err := util.GetCommand(common.CommandKubectl)
	if err != nil {
		return errors.Wrap(err, "kubectl not found")
	}
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s apply -f %q", kubectl, manifestPath), false, true); err != nil {
		return errors.Wrap(err, "failed to apply generic-device-plugin manifest")
	}
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s rollout status daemonset/generic-device-plugin -n kube-system --timeout=180s", kubectl), false, true); err != nil {
		return errors.Wrap(err, "generic-device-plugin rollout failed")
	}
	return nil
}
