package upgrade

import (
	"fmt"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

const (
	// legacyAMDGPUDevicePluginImage is the repository of the AMD GPU device
	// plugin. Earlier releases referenced it in the DaemonSet without a tag,
	// which resolves to ":latest". Docker Hub has since re-pushed "latest" and
	// garbage-collected the manifest that was originally cached, leaving the
	// local image untagged (shown as "-" and counted as a reclaimable, unused
	// image). The DaemonSet is now pinned to an immutable tag, but hosts
	// upgraded from those releases still hold the orphaned image.
	legacyAMDGPUDevicePluginImage = "docker.io/rocm/k8s-device-plugin"
	// legacyAMDGPUDevicePluginImageTag is the immutable tag we attach to that
	// orphaned image so it is no longer untagged.
	legacyAMDGPUDevicePluginImageTag = legacyAMDGPUDevicePluginImage + ":1.31.0.9"
)

// retagLegacyAMDGPUImage gives the previously untagged AMD GPU device plugin
// image a stable version tag (1.31.0.9) via ctr, so it is no longer an orphan
// excluded from being recognized as a managed image.
func retagLegacyAMDGPUImage() []task.Interface {
	return []task.Interface{
		&task.LocalTask{
			Name:   "RetagLegacyAMDGPUImage",
			Action: new(retagLegacyAMDGPUImageAction),
		},
	}
}

type retagLegacyAMDGPUImageAction struct {
	common.KubeAction
}

func (a *retagLegacyAMDGPUImageAction) Execute(runtime connector.Runtime) error {
	// the image only exists on Linux hosts that previously ran the AMD GPU
	// device plugin; ctr is not available elsewhere (e.g. macOS/minikube).
	if !runtime.GetSystemInfo().IsLinux() {
		return nil
	}

	// the legacy image is stored under the bare repository name (untagged).
	cmd := fmt.Sprintf("ctr -n k8s.io i tag --force %s %s", legacyAMDGPUDevicePluginImage, legacyAMDGPUDevicePluginImageTag)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, false); err != nil {
		// the image is absent on this host, nothing to do, do not fail the upgrade.
		logger.Debugf("legacy AMD GPU device plugin image %s not found, skip retagging: %v", legacyAMDGPUDevicePluginImage, err)
		return nil
	}
	logger.Infof("tagged legacy AMD GPU device plugin image %s as %s", legacyAMDGPUDevicePluginImage, legacyAMDGPUDevicePluginImageTag)
	return nil
}
