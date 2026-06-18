package intelgpu

import (
	"context"

	"github.com/beclab/Olares/cli/pkg/clientset"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/gpu"

	"github.com/pkg/errors"
)

// UpdateNodeIntelGPUInfo labels the node as supporting the "intel" mode (Intel
// integrated GPU)
type UpdateNodeIntelGPUInfo struct {
	common.KubeAction
}

func (u *UpdateNodeIntelGPUInfo) Execute(runtime connector.Runtime) error {
	client, err := clientset.NewKubeClient()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubeclient create error")
	}

	if !connector.HasIntelGPU(runtime) {
		logger.Info("Intel GPU is not detected")
		return nil
	}

	return gpu.SetNodeGpuModeLabel(context.Background(), client.Kubernetes(), gpu.IntelType, nil, nil, nil)
}
