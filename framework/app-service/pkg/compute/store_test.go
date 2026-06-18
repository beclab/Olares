package compute

import (
	"sort"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func k8sNode(name string, memory string, labels map[string]string) *corev1.Node {
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name, Labels: labels},
		Status: corev1.NodeStatus{
			Capacity: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse(memory)},
			Conditions: []corev1.NodeCondition{
				{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
			},
		},
	}
}

func sortedStrings(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}

func devicesForMode(n Node, mode string) []Device {
	out := make([]Device, 0)
	for _, d := range n.Devices {
		if d.Mode == mode {
			out = append(out, d)
		}
	}
	return out
}

func TestBuildNodeResourcePureCPU(t *testing.T) {
	n := buildNodeResource(k8sNode("cpu-a", "16Gi", nil))
	if len(n.GPUTypes) != 0 {
		t.Fatalf("pure-cpu node should advertise no gpu types, got %v", n.GPUTypes)
	}
	if len(n.Devices) != 1 || n.Devices[0].Mode != utils.CPUType || n.Devices[0].SupportType != SupportTypeMemorySlice {
		t.Fatalf("expected a single cpu memory-slice device, got %+v", n.Devices)
	}
	if !n.SupportsMode(utils.CPUType) {
		t.Fatal("every node must support cpu")
	}
}

func TestBuildNodeResourceMultiMode(t *testing.T) {
	// Olares One style node: advertises both nvidia and intel, so it stays a
	// single Node carrying both modes with per-mode devices.
	node := k8sNode("olares-one", "32Gi", map[string]string{
		utils.NodeGPUTypeLabelPrefix + utils.NvidiaCardType: "true",
		utils.NodeGPUTypeLabelPrefix + utils.IntelType:      "true",
	})
	n := buildNodeResource(node)

	if got, want := sortedStrings(n.GPUTypes), []string{utils.IntelType, utils.NvidiaCardType}; !equalStringSlices(got, want) {
		t.Fatalf("expected GPUTypes %v, got %v", want, got)
	}
	if !n.SupportsMode(utils.NvidiaCardType) || !n.SupportsMode(utils.IntelType) || !n.SupportsMode(utils.CPUType) {
		t.Fatalf("node should support nvidia, intel and cpu; GPUTypes=%v", n.GPUTypes)
	}
	if n.SupportsMode(utils.AMDType) {
		t.Fatal("node must not claim support for amd")
	}

	// Intel is a unified-memory accelerator: one MemorySlice device, tagged
	// with its mode and the node-mode device id.
	intelDevs := devicesForMode(n, utils.IntelType)
	if len(intelDevs) != 1 || intelDevs[0].SupportType != SupportTypeMemorySlice {
		t.Fatalf("expected one memory-slice intel device, got %+v", intelDevs)
	}
	if intelDevs[0].ID != "olares-one-intel-0" {
		t.Fatalf("unexpected intel device id %q", intelDevs[0].ID)
	}

	// viewForMode projects to a single mode so the scheduler only sees that
	// mode's devices on this multi-mode node.
	view := n.viewForMode(utils.IntelType)
	if len(view.Devices) != 1 || view.Devices[0].Mode != utils.IntelType {
		t.Fatalf("viewForMode(intel) should keep only intel devices, got %+v", view.Devices)
	}
}

// TestBindingSelectionResolvesDeviceOnMultiModeNode guards the multi-mode node
// binding path: a node that exposes both nvidia and intel must let callers
// select either mode's device by (nodeName, deviceID). The single physical-node
// model keeps every device under one NodeName so findNode/findDevice locate the
// right one, and the validation keys the mode off the selected device.
func TestBindingSelectionResolvesDeviceOnMultiModeNode(t *testing.T) {
	node := Node{
		NodeName:       "olares-one",
		GPUTypes:       []string{utils.NvidiaCardType, utils.IntelType},
		memoryCapacity: 32 * gi,
		Devices: []Device{
			{ID: "gpu0", NodeName: "olares-one", Mode: utils.NvidiaCardType, Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeExclusive, AvailableSupportTypes: AvailableSupportTypes(utils.NvidiaCardType)},
			{ID: "olares-one-intel-0", NodeName: "olares-one", Mode: utils.IntelType, Memory: 24 * gi, Health: deviceHealthYes, SupportType: SupportTypeMemorySlice, AvailableSupportTypes: AvailableSupportTypes(utils.IntelType)},
		},
	}
	nodes := []Node{node}

	resolved, err := resolveSelection([]BindingSelection{{NodeName: "olares-one", DeviceID: "olares-one-intel-0"}}, nodes)
	if err != nil {
		t.Fatalf("resolveSelection on multi-mode node: %v", err)
	}
	if len(resolved) != 1 || resolved[0].device.ID != "olares-one-intel-0" || resolved[0].device.Mode != utils.IntelType {
		t.Fatalf("expected the intel device resolved, got %+v", resolved)
	}

	reqIntel := Requirement{Mode: utils.IntelType, RequiredMemory: 4 * gi, LimitedMemory: 4 * gi}
	if res := ValidateBindingSelection(reqIntel, []BindingSelection{{NodeName: "olares-one", DeviceID: "olares-one-intel-0"}}, nodes, PressureSnapshot{}); !res.OK {
		t.Fatalf("intel device should be valid for an intel request, got %+v", res)
	}

	reqNvidia := Requirement{Mode: utils.NvidiaCardType, RequiredGPU: 4 * gi, LimitedGPU: 4 * gi}
	if res := ValidateBindingSelection(reqNvidia, []BindingSelection{{NodeName: "olares-one", DeviceID: "olares-one-intel-0"}}, nodes, PressureSnapshot{}); res.OK {
		t.Fatalf("intel device must not satisfy an nvidia request, got %+v", res)
	}
}

// TestBuildNodeResourceDiscreteGPULabel documents that a node advertising a
// discrete-GPU mode (amd-gpu / intel-gpu) is treated like any other advertised
// mode now that there is no scheduling-time filter: it becomes a node-level
// device for that mode. In practice olares-cli never writes these labels, so
// this path is not exercised on real clusters.
func TestBuildNodeResourceDiscreteGPULabel(t *testing.T) {
	node := k8sNode("dgpu-a", "16Gi", map[string]string{
		utils.NodeGPUTypeLabelPrefix + utils.AMDGPUType: "true",
	})
	n := buildNodeResource(node)
	if len(n.GPUTypes) != 1 || n.GPUTypes[0] != utils.AMDGPUType {
		t.Fatalf("expected node to advertise amd-gpu, got %v", n.GPUTypes)
	}
	if len(n.Devices) != 1 || n.Devices[0].Mode != utils.AMDGPUType {
		t.Fatalf("expected a single amd-gpu device, got %+v", n.Devices)
	}
}
