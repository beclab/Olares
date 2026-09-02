package compute

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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

func TestBuildNodeResourceDiscreteRegisterAnnotation(t *testing.T) {
	intel := k8sNode("intel-dgpu", "32Gi", map[string]string{
		utils.NodeGPUTypeLabelPrefix + utils.IntelGPUType: "true",
	})
	intel.Annotations = map[string]string{
		constants.NodeIntelRegisterKey: "dgpu,card0,xe,Intel® Arc™ A770 Graphics,Xe-HPG,Alchemist,17179869184",
	}
	in := buildNodeResource(intel)
	if len(in.Devices) != 1 {
		t.Fatalf("expected one intel discrete device, got %+v", in.Devices)
	}
	if in.Devices[0].ID != "intel-dgpu-intel-gpu-card0" || in.Devices[0].CardModel != "Intel® Arc™ A770 Graphics" || in.Devices[0].Memory != 17179869184 {
		t.Fatalf("unexpected intel discrete device %+v", in.Devices[0])
	}

	amd := k8sNode("amd-dgpu", "64Gi", map[string]string{
		utils.NodeGPUTypeLabelPrefix + utils.AMDGPUType: "true",
	})
	amd.Annotations = map[string]string{
		constants.NodeAmdRegisterKey: "dgpu,card1,amdgpu,AMD Radeon AI PRO R9700,GC_12_0_0,7551,34208743424",
	}
	an := buildNodeResource(amd)
	if len(an.Devices) != 1 {
		t.Fatalf("expected one amd discrete device, got %+v", an.Devices)
	}
	if an.Devices[0].ID != "amd-dgpu-amd-gpu-card1" || an.Devices[0].CardModel != "AMD Radeon AI PRO R9700" || an.Devices[0].Memory != 34208743424 {
		t.Fatalf("unexpected amd discrete device %+v", an.Devices[0])
	}
}

// hamiRegisterAnnotation encodes one card the way HAMi's device plugin does
// (util.EncodeNodeDevices): ID,Count,Devmem,Devcore,Type,Numa,Health,Index,
// Mode,Architecture followed by the device separator.
func hamiRegisterAnnotation(uuid string, memMiB int, healthy bool) string {
	return fmt.Sprintf("%s,10,%d,100,NVIDIA GeForce RTX 5090,0,%t,0,hami-core,0%s",
		uuid, memMiB, healthy, constants.OneContainerMultiDeviceSplitSymbol)
}

func hamiNode(name string, handshake string, healthyCard bool) *corev1.Node {
	node := k8sNode(name, "32Gi", map[string]string{
		utils.NodeGPUTypeLabelPrefix + utils.NvidiaCardType: "true",
	})
	node.Annotations = map[string]string{
		constants.NodeNvidiaRegistryKey: hamiRegisterAnnotation("GPU-"+name, 24576, healthyCard),
	}
	if handshake != "" {
		node.Annotations[constants.NodeHandshakeKey] = handshake
	}
	return node
}

func hamiStamp(age time.Duration) string {
	return time.Now().UTC().Add(-age).Format(time.DateTime)
}

// TestDecodeHAMINvidiaDevicesNodeLiveness covers the gap that let a powered-off
// node keep offering GPUs: the health bit inside the register annotation is
// written by the device plugin while the node is up and is never cleared, so it
// has to be read together with the node's own liveness signals.
func TestDecodeHAMINvidiaDevicesNodeLiveness(t *testing.T) {
	tests := []struct {
		name       string
		handshake  string
		healthyBit bool
		poweredOff bool
		want       string
	}{
		{name: "plugin reported recently", handshake: "Reported " + time.Now().String(), healthyBit: true, want: deviceHealthYes},
		{name: "no handshake annotation at all", handshake: "", healthyBit: true, want: deviceHealthYes},
		{name: "handshake request still within the window", handshake: "Requesting_" + hamiStamp(10*time.Second), healthyBit: true, want: deviceHealthYes},
		{name: "handshake request went unanswered", handshake: "Requesting_" + hamiStamp(2*time.Minute), healthyBit: true, want: deviceHealthNo},
		{name: "handshake unparseable", handshake: "Requesting_not-a-timestamp", healthyBit: true, want: deviceHealthNo},
		{name: "hami already cleaned the node up", handshake: "Deleted_" + hamiStamp(time.Minute), healthyBit: true, want: deviceHealthNo},
		// The regression: the node is gone but its register annotation, and so
		// the card's own health bit, still says the card is fine.
		{name: "node powered off with a stale healthy register", handshake: "Reported " + time.Now().String(), healthyBit: true, poweredOff: true, want: deviceHealthNo},
		{name: "card reported unhealthy by the plugin", handshake: "Reported " + time.Now().String(), healthyBit: false, want: deviceHealthNo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := hamiNode("gpu-a", tt.handshake, tt.healthyBit)
			if tt.poweredOff {
				// What the node controller records once kubelet stops
				// heartbeating.
				node.Status.Conditions[0].Status = corev1.ConditionUnknown
			}
			devices := decodeHAMINvidiaDevices(node, utils.NvidiaCardType)
			if len(devices) != 1 {
				t.Fatalf("expected exactly one decoded card, got %+v", devices)
			}
			if devices[0].Health != tt.want {
				t.Fatalf("card health = %q, want %q", devices[0].Health, tt.want)
			}
		})
	}
}

// TestPoweredOffNvidiaNodeIsNotPicked drives the whole selection path rather
// than the decoder alone: install and resume both choose out of
// listAvailableForLaunch, so a node that is gone must neither be offered to the
// user nor picked automatically.
func TestPoweredOffNvidiaNodeIsNotPicked(t *testing.T) {
	// Both nodes carry an identical, healthy-looking register annotation —
	// that is the whole point: nothing about the powered-off node's GPU
	// advertisement changes when it goes down.
	deadNode := hamiNode("gpu-dead", "Reported "+time.Now().String(), true)
	deadNode.Status.Conditions[0].Status = corev1.ConditionUnknown

	live := buildNodeResource(hamiNode("gpu-live", "Reported "+time.Now().String(), true))
	dead := buildNodeResource(deadNode)

	req := Requirement{Mode: utils.NvidiaCardType, RequiredGPU: 8 * gi, LimitedGPU: 8 * gi}
	availability := listAvailableForLaunch(req, []Node{live, dead}, PressureSnapshot{})

	for _, node := range availability.Nodes {
		if node.NodeName == "gpu-dead" {
			t.Fatalf("a powered-off node must not appear in the resume picker: %+v", node)
		}
		for _, device := range node.Devices {
			if node.NodeName == "gpu-live" && !device.Operable {
				t.Fatalf("the live node's card should still be offered: %+v", device)
			}
		}
	}

	// Repeated because the picker breaks ties at random: a single run could
	// pass on the live node by luck even if the dead one were still a
	// candidate.
	for i := 0; i < 50; i++ {
		selections, ok := pickLaunchSelection(req, availability, PressureSnapshot{}, allocationOptions{checkPressure: true})
		if !ok || len(selections) != 1 {
			t.Fatalf("expected the live node to be picked, got %+v (ok=%t)", selections, ok)
		}
		if selections[0].NodeName != "gpu-live" {
			t.Fatalf("picked a powered-off node: %+v", selections[0])
		}
	}

	// With only the dead node left there must be no placement at all, rather
	// than a fallback onto its stale cards.
	deadOnly := listAvailableForLaunch(req, []Node{dead}, PressureSnapshot{})
	if len(deadOnly.Nodes) != 0 {
		t.Fatalf("a powered-off-only cluster must not list the dead node, got %#v", deadOnly.Nodes)
	}
	if _, ok := pickLaunchSelection(req, deadOnly, PressureSnapshot{}, allocationOptions{checkPressure: true}); ok {
		t.Fatal("a cluster whose only GPU node is powered off must not yield a placement")
	}
}

// TestNotReadyNodeIsHiddenFromResumePicker pins the rule on node readiness
// alone: kubelet has stopped heartbeating, so the node is dropped from both
// flows no matter what its cards claim. The dead node's card is deliberately
// left healthy — a shape the cluster cannot actually produce, since
// decodeHAMINvidiaDevices folds node liveness into every card — so that the
// test fails if the decision ever starts depending on the devices again.
func TestNotReadyNodeIsHiddenFromResumePicker(t *testing.T) {
	req := Requirement{Mode: utils.NvidiaCardType, RequiredGPU: 8 * gi, LimitedGPU: 8 * gi}
	dead := nvidiaNode("zero0", Device{ID: "gpu-4060", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeTimeSlice})
	dead.Health = deviceHealthNo
	live := nvidiaNode("olares-26", Device{ID: "gpu-5090", Memory: 24 * gi, Health: deviceHealthYes, SupportType: SupportTypeTimeSlice})
	live.Health = deviceHealthYes

	result := listAvailableForLaunch(req, []Node{dead, live}, PressureSnapshot{})
	if len(result.Nodes) != 1 || result.Nodes[0].NodeName != "olares-26" {
		t.Fatalf("NotReady node must be dropped from resume options, got %#v", result.Nodes)
	}

	// Install auto-picks out of the very same view, so it must not reach the
	// dead node either. This is the assertion the previous fix was missing:
	// it only checked the pick and the Operable bit, both of which pass while
	// the node is still listed. Repeated because ties are broken at random.
	for i := 0; i < 50; i++ {
		selections, ok := pickLaunchSelection(req, result, PressureSnapshot{}, allocationOptions{checkPressure: true})
		if !ok || len(selections) != 1 || selections[0].NodeName != "olares-26" {
			t.Fatalf("install must pick the live node, got %+v (ok=%t)", selections, ok)
		}
	}

	// With the NotReady node alone there is nothing left to offer, and no
	// devices are reported for it at all.
	only := listAvailableForLaunch(req, []Node{dead}, PressureSnapshot{})
	if len(only.Nodes) != 0 || only.Schedulable {
		t.Fatalf("a NotReady-only cluster must offer nothing, got %#v", only)
	}
}

func TestFetchSchedulableNodeComputeAllocationsExcludesNotReadyAndCordonedGPU(t *testing.T) {
	labels := map[string]string{utils.NodeGPUTypeLabelPrefix + utils.NvidiaCardType: "true"}
	ready := k8sNode("ready", "16Gi", labels)
	notReady := k8sNode("not-ready", "16Gi", labels)
	notReady.Status.Conditions[0].Status = corev1.ConditionFalse
	cordoned := k8sNode("cordoned", "16Gi", labels)
	cordoned.Spec.Unschedulable = true
	c := fake.NewClientBuilder().WithObjects(ready, notReady, cordoned).Build()

	all, err := FetchNodeComputeAllocations(context.Background(), c)
	if err != nil {
		t.Fatalf("fetch all nodes: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("install allocation view changed, got nodes %#v", all)
	}

	schedulable, err := FetchSchedulableNodeComputeAllocations(context.Background(), c)
	if err != nil {
		t.Fatalf("fetch schedulable nodes: %v", err)
	}
	if len(schedulable) != 1 || schedulable[0].NodeName != "ready" ||
		!schedulable[0].SupportsMode(utils.NvidiaCardType) {
		t.Fatalf("schedulable nodes=%#v", schedulable)
	}
}
