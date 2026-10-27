package gpu

import (
	"context"
	"os"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// newFakeNodeClient builds a client holding a single node named after the local
// hostname, since that is the node updateCurrentNodeLabels resolves and patches.
func newFakeNodeClient(t *testing.T, labels map[string]string) ctrlclient.Client {
	t.Helper()

	hostname, err := os.Hostname()
	if err != nil {
		t.Fatalf("get hostname: %v", err)
	}

	s := runtime.NewScheme()
	if err := scheme.AddToScheme(s); err != nil {
		t.Fatalf("add corev1 to scheme: %v", err)
	}

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   hostname,
			Labels: labels,
		},
	}

	return fake.NewClientBuilder().WithScheme(s).WithObjects(node).Build()
}

func nodeLabels(t *testing.T, client ctrlclient.Client) map[string]string {
	t.Helper()

	hostname, err := os.Hostname()
	if err != nil {
		t.Fatalf("get hostname: %v", err)
	}

	var node corev1.Node
	if err := client.Get(context.Background(), ctrlclient.ObjectKey{Name: hostname}, &node); err != nil {
		t.Fatalf("get node: %v", err)
	}
	return node.GetLabels()
}

// mixedAcceleratorLabels mimics a node with both an NVIDIA discrete card and an
// Intel iGPU, plus a stale legacy type label and an unrelated label.
func mixedAcceleratorLabels() map[string]string {
	return map[string]string{
		GpuModeLabel(NvidiaCardType): "true",
		GpuModeLabel(IntelType):      "true",
		GpuModeLabel(AMDType):        "true",
		GpuDriverLabel:               "595.84",
		GpuCudaLabel:                 "13.2",
		GpuCudaSupportedLabel:        "true",
		GpuType:                      NvidiaCardType,
		"kubernetes.io/arch":         "amd64",
	}
}

func TestRemoveAllNodeGpuLabels(t *testing.T) {
	client := newFakeNodeClient(t, mixedAcceleratorLabels())

	if err := RemoveAllNodeGpuLabels(context.Background(), client); err != nil {
		t.Fatalf("RemoveAllNodeGpuLabels: %v", err)
	}

	labels := nodeLabels(t, client)
	for _, key := range []string{
		GpuModeLabel(NvidiaCardType),
		GpuModeLabel(IntelType),
		GpuModeLabel(AMDType),
		GpuDriverLabel,
		GpuCudaLabel,
		GpuCudaSupportedLabel,
		GpuType,
	} {
		if _, ok := labels[key]; ok {
			t.Errorf("label %s should have been removed", key)
		}
	}
	if labels["kubernetes.io/arch"] != "amd64" {
		t.Error("unrelated labels should be preserved")
	}
}

func TestRemoveNvidiaNodeGpuLabelsKeepsOtherAccelerators(t *testing.T) {
	client := newFakeNodeClient(t, mixedAcceleratorLabels())

	if err := RemoveNvidiaNodeGpuLabels(context.Background(), client); err != nil {
		t.Fatalf("RemoveNvidiaNodeGpuLabels: %v", err)
	}

	labels := nodeLabels(t, client)
	for _, key := range []string{
		GpuModeLabel(NvidiaCardType),
		GpuModeLabel(GB10ChipType),
		GpuDriverLabel,
		GpuCudaLabel,
		GpuCudaSupportedLabel,
		GpuType,
	} {
		if _, ok := labels[key]; ok {
			t.Errorf("nvidia-owned label %s should have been removed", key)
		}
	}
	for _, key := range []string{
		GpuModeLabel(IntelType),
		GpuModeLabel(AMDType),
	} {
		if labels[key] != "true" {
			t.Errorf("label %s of a non-nvidia accelerator should have been preserved", key)
		}
	}
	if labels["kubernetes.io/arch"] != "amd64" {
		t.Error("unrelated labels should be preserved")
	}
}

// NvidiaGpuModeTypes must stay a subset of AllGpuModeTypes, otherwise the
// uninstall path would leave an nvidia mode label behind.
func TestNvidiaGpuModeTypesIsSubsetOfAllGpuModeTypes(t *testing.T) {
	all := make(map[string]struct{}, len(AllGpuModeTypes))
	for _, mode := range AllGpuModeTypes {
		all[mode] = struct{}{}
	}
	for _, mode := range NvidiaGpuModeTypes {
		if _, ok := all[mode]; !ok {
			t.Errorf("nvidia mode %s is missing from AllGpuModeTypes", mode)
		}
	}
}
