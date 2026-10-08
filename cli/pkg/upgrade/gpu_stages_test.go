package upgrade

import (
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/gpu"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestSharedGPUStackDetectsHardwareOnlyOnWorker(t *testing.T) {
	for _, mode := range []string{gpu.IntelType, gpu.IntelGpuType, gpu.AmdGpuType, gpu.AMDType} {
		nodes := []corev1.Node{
			{ObjectMeta: metav1.ObjectMeta{Name: "cpu-master"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "gpu-worker", Labels: map[string]string{gpu.GpuModeLabel(mode): "true"}}},
		}
		if !nodesHaveGPUMode(nodes, []string{mode}) {
			t.Fatalf("worker GPU %s was missed", mode)
		}
		if nodesHaveGPUMode(nodes[:1], []string{mode}) {
			t.Fatalf("CPU node claims %s", mode)
		}
	}
}

func TestGPUNodeChecksAndLabelsAreNotAdminOnly(t *testing.T) {
	plan, err := BuildPlan(semver.MustParse("1.12.8-20261020"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"LabelIntelGPUNode", "LabelAMDGPUNode", "CheckIntelGpu", "CheckXpumd"} {
		i, ok := taskIndex(plan, name)
		if !ok || plan.Stages[i].Placement != PlacementAllNodes {
			t.Fatalf("%s does not run on workers", name)
		}
	}
	for _, name := range []string{"ApplyIntelNFD", "ApplyIntelGPUPlugin"} {
		i, ok := taskIndex(plan, name)
		if !ok || plan.Stages[i].Placement != PlacementAdmin {
			t.Fatalf("shared %s is not admin-only", name)
		}
	}
}
