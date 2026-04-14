package mtgpu

import (
	"context"

	"github.com/beclab/Olares/cli/pkg/clientset"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/gpu"
	"k8s.io/utils/ptr"

	"github.com/pkg/errors"
)

// UpdateNodeMThreadsGPUInfo updates Kubernetes node labels with MThreads GPU information.
type UpdateNodeMThreadsGPUInfo struct {
	common.KubeAction
}

func (u *UpdateNodeMThreadsGPUInfo) Execute(runtime connector.Runtime) error {
	client, err := clientset.NewKubeClient()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubeclient create error")
	}

	// Check if MThreads AI Book M1000 exists
	m1000Exists := connector.IsMThreadsAIBookM1000(runtime)
	if !m1000Exists {
		logger.Info("MThreads AI Book M1000 is not detected")
		return nil
	}

	if runtime.GetSystemInfo().IsAmdApu() {
		return gpu.UpdateNodeGpuLabel(context.Background(), client.Kubernetes(), nil, nil, nil, ptr.To(gpu.MThreadsM1000Type))
	}

	return nil
}
