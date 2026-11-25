package compute

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
)

func exclusiveCard(id string) Device {
	return Device{ID: id, Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeExclusive}
}

func consumerReq(name, perPod string, replicas int32, requiredGPU int64) Requirement {
	return Requirement{
		Mode:              utils.NvidiaCardType,
		RequiredGPU:       requiredGPU,
		LimitedGPU:        requiredGPU,
		SupportMultiCards: true,
		SupportMultiNodes: true,
		Consumer:          name,
		PerPod:            perPod,
		Replicas:          replicas,
	}
}

func cardsByNode(rows []Allocation) map[string]int {
	out := map[string]int{}
	for _, row := range rows {
		out[row.NodeName]++
	}
	return out
}

func TestSplitPicksOneCardPerReplicaAcrossNodes(t *testing.T) {
	app := &appcfg.ApplicationConfig{AppName: "llm", OwnerName: "alice"}
	req := consumerReq("server", appcfg.PerPodSplit, 2, 20*gi)
	nodes := []Node{nvidiaNode("a", exclusiveCard("a0")), nvidiaNode("b", exclusiveCard("b0"))}
	rows, validation, err := placeConsumers(app, []Requirement{req}, nil, nodes, PressureSnapshot{})
	if err != nil || validation != nil {
		t.Fatalf("unexpected failure %v %v", err, validation)
	}
	if got := cardsByNode(rows); got["a"] != 1 || got["b"] != 1 {
		t.Fatalf("expected one card on each node, got %v", got)
	}
	for _, row := range rows {
		if row.Consumer != "server" {
			t.Fatalf("row lost its consumer: %+v", row)
		}
	}
}

func TestSplitGrowsGroupsUntilMemoryFits(t *testing.T) {
	app := &appcfg.ApplicationConfig{AppName: "llm", OwnerName: "alice"}
	req := consumerReq("server", appcfg.PerPodSplit, 2, 40*gi)
	nodes := []Node{nvidiaNode("a", exclusiveCard("a0"), exclusiveCard("a1"), exclusiveCard("a2"), exclusiveCard("a3"))}
	rows, validation, err := placeConsumers(app, []Requirement{req}, nil, nodes, PressureSnapshot{})
	if err != nil || validation != nil {
		t.Fatalf("unexpected failure %v %v", err, validation)
	}
	if len(rows) != 4 {
		t.Fatalf("two replicas of 32Gi fall short of 40Gi, expected k=2 and four cards, got %d", len(rows))
	}
}

func TestSplitSelectionRules(t *testing.T) {
	req := consumerReq("server", appcfg.PerPodSplit, 2, 0)
	resolve := func(nodes ...string) []resolvedSelection {
		out := make([]resolvedSelection, 0, len(nodes))
		for _, node := range nodes {
			out = append(out, resolvedSelection{node: Node{NodeName: node}})
		}
		return out
	}
	if got := validateSplitSelection(req, resolve("a", "a", "a")); got == nil || got.Code != "split-cards-not-divisible" {
		t.Fatalf("three cards for two replicas must be refused, got %+v", got)
	}
	if got := validateSplitSelection(req, resolve("a", "a", "a", "b")); got == nil || got.Code != "split-group-not-on-one-node" {
		t.Fatalf("a group split across nodes must be refused, got %+v", got)
	}
	if got := validateSplitSelection(req, resolve("a", "a", "b", "b")); got != nil {
		t.Fatalf("two groups of two on two nodes is valid, got %+v", got)
	}
}

func TestConsumersPlaceAtomically(t *testing.T) {
	app := &appcfg.ApplicationConfig{AppName: "llm", OwnerName: "alice"}
	reqs := []Requirement{
		consumerReq("asr", appcfg.PerPodOne, 1, 8*gi),
		consumerReq("llm", appcfg.PerPodSplit, 2, 8*gi),
	}
	nodes := []Node{nvidiaNode("a", exclusiveCard("a0"), exclusiveCard("a1"))}
	rows, _, err := placeConsumers(app, reqs, nil, nodes, PressureSnapshot{})
	var placement *consumerPlacementError
	if !errors.As(err, &placement) || placement.consumer != "llm" || rows != nil || !errors.Is(err, ErrConsumerUnplaceable) {
		t.Fatalf("expected llm to fail with nothing placed, got rows=%v err=%v", rows, err)
	}
	for _, device := range nodes[0].Devices {
		if len(device.Bindings) != 0 {
			t.Fatalf("a failed placement must not touch the caller's node view")
		}
	}
}

func TestConsumersSeeEachOthersCardsAndPreferOneNode(t *testing.T) {
	app := &appcfg.ApplicationConfig{AppName: "llm", OwnerName: "alice"}
	reqs := []Requirement{
		consumerReq("asr", appcfg.PerPodOne, 1, 8*gi),
		consumerReq("llm", appcfg.PerPodOne, 1, 8*gi),
	}
	nodes := []Node{
		nvidiaNode("a", exclusiveCard("a0"), exclusiveCard("a1")),
		nvidiaNode("b", exclusiveCard("b0"), exclusiveCard("b1")),
	}
	rows, validation, err := placeConsumers(app, reqs, nil, nodes, PressureSnapshot{})
	if err != nil || validation != nil {
		t.Fatalf("unexpected failure %v %v", err, validation)
	}
	if len(rows) != 2 || rows[0].NodeName != rows[1].NodeName || rows[0].DeviceID == rows[1].DeviceID {
		t.Fatalf("expected two distinct cards on one node, got %+v", rows)
	}
}

func TestManualConsumerSelectionIsValidated(t *testing.T) {
	app := &appcfg.ApplicationConfig{AppName: "llm", OwnerName: "alice"}
	reqs := []Requirement{consumerReq("llm", appcfg.PerPodSplit, 2, 0)}
	nodes := []Node{nvidiaNode("a", exclusiveCard("a0"), exclusiveCard("a1"), exclusiveCard("a2"))}
	manual := map[string][]BindingSelection{"llm": {
		{NodeName: "a", DeviceID: "a0"}, {NodeName: "a", DeviceID: "a1"}, {NodeName: "a", DeviceID: "a2"},
	}}
	_, validation, err := placeConsumers(app, reqs, manual, nodes, PressureSnapshot{})
	if err != nil || validation == nil || validation.Code != "split-cards-not-divisible" {
		t.Fatalf("expected the manual pick to be refused, got %v %+v", err, validation)
	}
}

func TestBindingNameKeepsLegacyRows(t *testing.T) {
	legacy := Allocation{AppName: "llm", Owner: "alice", DeviceID: "GPU-1"}
	withConsumer := legacy
	withConsumer.Consumer = "server"
	if hamiBindingName(legacy) != "llm-alice-gpu-1-"+shortHash("llm-alice-GPU-1") {
		t.Fatalf("legacy binding name changed: %s", hamiBindingName(legacy))
	}
	if hamiBindingName(legacy) == hamiBindingName(withConsumer) {
		t.Fatal("two consumers on one card need distinct bindings")
	}
}

func TestLegacyRowsFitConsumer(t *testing.T) {
	rows := []Allocation{{Mode: utils.NvidiaCardType, NodeName: "a"}, {Mode: utils.NvidiaCardType, NodeName: "b"}}
	if !LegacyRowsFitConsumer(consumerReq("s", appcfg.PerPodOne, 1, 0), rows) {
		t.Fatal("one card per pod accepts any rows of the same mode")
	}
	if !LegacyRowsFitConsumer(consumerReq("s", appcfg.PerPodSplit, 2, 0), rows) {
		t.Fatal("two cards on two nodes split into two groups of one")
	}
	if LegacyRowsFitConsumer(consumerReq("s", appcfg.PerPodSplit, 3, 0), rows) {
		t.Fatal("two cards cannot feed three replicas")
	}
	gb10 := consumerReq("s", appcfg.PerPodOne, 1, 0)
	gb10.Mode = utils.GB10ChipType
	if LegacyRowsFitConsumer(gb10, rows) {
		t.Fatal("rows of another mode must be re-placed")
	}
}

func TestAutoSelectConsumerModes(t *testing.T) {
	consumers := []appcfg.GPUConsumer{
		{Name: "asr", Modes: []appcfg.ResourceMode{{Mode: utils.NvidiaCardType}, {Mode: utils.AMDGPUType}}},
		{Name: "llm", Modes: []appcfg.ResourceMode{{Mode: utils.NvidiaCardType}, {Mode: utils.CPUType}}},
	}
	cluster := map[string]struct{}{utils.NvidiaCardType: {}}
	chosen, ambiguous, err := autoSelectConsumerModes(consumers, nil, cluster)
	if err != nil || chosen["asr"] != utils.NvidiaCardType || len(ambiguous) != 1 || ambiguous[0] != "llm" {
		t.Fatalf("unexpected %v %v %v", chosen, ambiguous, err)
	}
	chosen, ambiguous, err = autoSelectConsumerModes(consumers, map[string]string{"llm": utils.CPUType}, cluster)
	if err != nil || chosen["llm"] != utils.CPUType || len(ambiguous) != 0 {
		t.Fatalf("unexpected %v %v %v", chosen, ambiguous, err)
	}
	if _, _, err = autoSelectConsumerModes(consumers, map[string]string{"web": utils.CPUType}, cluster); !errors.Is(err, ErrUnknownConsumer) {
		t.Fatalf("expected unknown consumer, got %v", err)
	}
}

func shortHash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])[:8]
}

func TestConsumersReportCumulativePressure(t *testing.T) {
	app := &appcfg.ApplicationConfig{AppName: "llm", OwnerName: "alice"}
	reqs := []Requirement{consumerReq("a", appcfg.PerPodOne, 1, 8*gi), consumerReq("b", appcfg.PerPodOne, 1, 8*gi)}
	for i := range reqs {
		reqs[i].Mode = utils.NvidiaCardType
		reqs[i].RequiredMemory = 70 * gi
	}
	nodes := []Node{nvidiaNode("a", exclusiveCard("a0"), exclusiveCard("a1"))}
	_, _, err := placeConsumers(app, reqs, nil, nodes, nodePressure("a"))
	var placement *consumerPlacementError
	if !errors.As(err, &placement) || placement.consumer != "b" || !strings.HasPrefix(placement.reason, "node-pressure") {
		t.Fatalf("the second consumer should fail on the pair's memory, got %v", err)
	}
}

func TestConsumerPressureReasonNamesEveryCandidateNode(t *testing.T) {
	app := &appcfg.ApplicationConfig{AppName: "llm", OwnerName: "alice"}
	reqs := []Requirement{consumerReq("a", appcfg.PerPodOne, 1, 8*gi), consumerReq("b", appcfg.PerPodOne, 1, 8*gi)}
	for i := range reqs {
		reqs[i].RequiredMemory = 70 * gi
	}
	small := exclusiveCard("b0")
	small.Memory = 4 * gi
	nodes := []Node{nvidiaNode("a", exclusiveCard("a0"), exclusiveCard("a1")), nvidiaNode("b", small)}
	_, _, err := placeConsumers(app, reqs, nil, nodes, nodePressure("a", "b"))
	var placement *consumerPlacementError
	if !errors.As(err, &placement) || placement.consumer != "b" {
		t.Fatalf("the second consumer should fail, got %v", err)
	}
	for _, want := range []string{"node-pressure", "a (memory)", "no card fits on b (" + NodeStatusNotEnough + ")"} {
		if !strings.Contains(placement.reason, want) {
			t.Fatalf("reason %q should contain %q", placement.reason, want)
		}
	}
}
