package upgrade

import (
	"context"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/gpu"
	"github.com/beclab/Olares/cli/pkg/gpu/amdgpu"
	"github.com/beclab/Olares/cli/pkg/gpu/intelgpu"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Shared GPU charts are installed once by the admin, based on every node's
// hardware labels. A CPU-only master must not skip a worker's GPU stack.
type clusterHasGPU struct {
	common.KubePrepare
	modes []string
}

func (p *clusterHasGPU) PreCheck(connector.Runtime) (bool, error) {
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return false, err
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	return nodesHaveGPUMode(nodes.Items, p.modes), nil
}

func nodesHaveGPUMode(nodes []corev1.Node, modes []string) bool {
	for _, node := range nodes {
		for _, mode := range modes {
			if node.Labels[gpu.GpuModeLabel(mode)] == "true" {
				return true
			}
		}
	}
	return false
}

func intelGPUClusterTasks() []task.Interface {
	hasIntel := func() *clusterHasGPU { return &clusterHasGPU{modes: []string{gpu.IntelType, gpu.IntelGpuType}} }
	return []task.Interface{
		&task.LocalTask{Name: "ApplyIntelNFD", Prepare: hasIntel(), Action: new(intelgpu.ApplyIntelNFD), Retry: 5, Delay: 5 * time.Second},
		&task.LocalTask{Name: "ApplyIntelNodeFeatureRules", Prepare: hasIntel(), Action: new(intelgpu.ApplyIntelNodeFeatureRules), Retry: 10, Delay: 6 * time.Second},
		&task.LocalTask{Name: "ApplyIntelGPUPlugin", Prepare: hasIntel(), Action: new(intelgpu.ApplyIntelGPUPlugin), Retry: 3, Delay: 5 * time.Second},
	}
}

func intelGPUNodeChecks() []task.Interface {
	return []task.Interface{
		&task.LocalTask{Name: "CheckIntelGpu", Prepare: new(intelgpu.HasAnyIntelGPU), Action: new(intelgpu.CheckIntelGpu), Retry: 50, Delay: 10 * time.Second},
		// Qualification depends on this machine's GPU IDs and kernel. This
		// serial node stage preserves that gate without a master-only probe.
		&task.LocalTask{Name: "InstallXpumd", Prepare: new(intelgpu.HasQualifyingIntelDGPU), Action: new(intelgpu.InstallXpumd), Retry: 3, Delay: 5 * time.Second},
		&task.LocalTask{Name: "CheckXpumd", Prepare: new(intelgpu.HasQualifyingIntelDGPU), Action: new(intelgpu.CheckXpumd), Retry: 30, Delay: 10 * time.Second},
	}
}

func amdMetricsNodeChecks() []task.Interface {
	return []task.Interface{
		&task.LocalTask{Name: "CheckDeviceMetricsExporter", Prepare: new(amdgpu.HasAmdDiscreteGPU), Action: new(amdgpu.CheckDeviceMetricsExporter), Retry: 30, Delay: 10 * time.Second},
	}
}
