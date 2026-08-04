package compute

import (
	"math"
	"reflect"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
)

func TestSimulatePreflightSingleDemand(t *testing.T) {
	report, err := SimulatePreflight([]PreflightDemand{cpuDemand("app", "alice", 100)}, installablePreflightSnapshot())
	if err != nil {
		t.Fatalf("simulate: %v", err)
	}
	if !report.Installable || report.FailedDemandID != "" || report.Reason != "" || len(report.Pressure) != 0 {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestSimulatePreflightAccumulatesInInputOrder(t *testing.T) {
	snapshot := installablePreflightSnapshot()
	snapshot.Pressure.UsageByNode["cpu-a"] = prometheus.NodeResourceUsage{CPUCapacity: 1000}
	b := cpuDemand("b", "alice", 600)
	a := cpuDemand("a", "alice", 400)

	report, err := SimulatePreflight([]PreflightDemand{b, a}, snapshot)
	if err != nil {
		t.Fatalf("simulate: %v", err)
	}
	if report.Installable || report.FailedDemandID != "a" || report.Reason != PreflightReasonNodePressure {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestSimulatePreflightStopsAtFirstFailure(t *testing.T) {
	snapshot := installablePreflightSnapshot()
	snapshot.Owners["alice"] = &prometheus.ClusterMetrics{
		CPU:    prometheus.Value{Total: 1},
		Memory: prometheus.Value{Total: 8 * float64(gi)},
	}

	report, err := SimulatePreflight([]PreflightDemand{
		cpuDemand("first", "alice", 1000),
		cpuDemand("never", "bob", 100000),
	}, snapshot)
	if err != nil {
		t.Fatalf("simulate: %v", err)
	}
	if report.Installable || report.FailedDemandID != "first" || report.Reason != PreflightReasonOwnerQuota {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestSimulatePreflightValidatesAllOwnerFactsBeforeSimulation(t *testing.T) {
	snapshot := installablePreflightSnapshot()
	delete(snapshot.Owners, "bob")
	snapshot.Cluster.CPU.Usage = snapshot.Cluster.CPU.Total

	_, err := SimulatePreflight([]PreflightDemand{
		cpuDemand("blocked", "alice", 1),
		cpuDemand("missing-owner", "bob", 1),
	}, snapshot)
	if err == nil {
		t.Fatal("missing facts for any demand must fail before simulation")
	}
}

func TestSimulatePreflightExclusiveGPUCompetition(t *testing.T) {
	snapshot := installablePreflightSnapshot()
	snapshot.Nodes = []Node{nvidiaNode("gpu-a", Device{
		ID: "gpu0", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeExclusive,
	})}
	snapshot.Pressure = nodePressure("gpu-a")

	report, err := SimulatePreflight([]PreflightDemand{
		gpuDemand("a", "alice", 8*gi),
		gpuDemand("b", "bob", 8*gi),
	}, snapshot)
	if err != nil {
		t.Fatalf("simulate: %v", err)
	}
	if report.Installable || report.FailedDemandID != "b" || report.Reason != PreflightReasonUnschedulable {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestSimulatePreflightTimeSliceHostMemory(t *testing.T) {
	snapshot := installablePreflightSnapshot()
	snapshot.Nodes = []Node{nvidiaNode("gpu-a",
		Device{ID: "gpu0", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeTimeSlice},
		Device{ID: "gpu1", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeTimeSlice},
	)}
	snapshot.Pressure = PressureSnapshot{
		Threshold: 0.9,
		UsageByNode: map[string]prometheus.NodeResourceUsage{
			"gpu-a": {MemoryCapacity: 32 * gi, MemoryAvailable: 32 * gi},
		},
	}
	demand := gpuDemand("multi", "alice", 24*gi)
	demand.Requirement.RequiredMemory = gi
	demand.Requirement.LimitedMemory = gi
	demand.Requirement.SupportMultiCards = true

	report, err := SimulatePreflight([]PreflightDemand{demand}, snapshot)
	if err != nil {
		t.Fatalf("simulate: %v", err)
	}
	if report.Installable || report.FailedDemandID != "multi" || report.Reason != PreflightReasonNodePressure {
		t.Fatalf("unexpected report: %#v", report)
	}
	if len(report.Pressure) != 1 || len(report.Pressure[0].Dimensions) != 1 ||
		report.Pressure[0].Dimensions[0].Required != 33*gi {
		t.Fatalf("expected pod plus both cards in host memory pressure: %#v", report.Pressure)
	}
}

func TestSimulatePreflightResourceChecks(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*PreflightSnapshot, *PreflightDemand)
		reason string
	}{
		{
			name: "physical capacity",
			mutate: func(s *PreflightSnapshot, d *PreflightDemand) {
				d.Requirement.RequiredMemory = 65 * gi
				d.Requirement.LimitedMemory = 65 * gi
			},
			reason: PreflightReasonClusterCapacity,
		},
		{
			name: "cluster pressure",
			mutate: func(s *PreflightSnapshot, d *PreflightDemand) {
				s.Cluster.Memory.Usage = 58 * float64(gi)
				d.Requirement.RequiredMemory = gi
				d.Requirement.LimitedMemory = gi
			},
			reason: PreflightReasonClusterPressure,
		},
		{
			name: "owner quota",
			mutate: func(s *PreflightSnapshot, d *PreflightDemand) {
				s.Owners["alice"] = &prometheus.ClusterMetrics{
					CPU:    prometheus.Value{Total: 1},
					Memory: prometheus.Value{Total: 8 * float64(gi)},
				}
				d.Requirement.RequiredCPU = 1000
			},
			reason: PreflightReasonOwnerQuota,
		},
		{
			name: "k8s strict comparison",
			mutate: func(s *PreflightSnapshot, d *PreflightDemand) {
				s.K8sAvailable.CPU = 100
				d.Requirement.RequiredCPU = 100
			},
			reason: PreflightReasonK8sRequest,
		},
		{
			name: "node pressure",
			mutate: func(s *PreflightSnapshot, d *PreflightDemand) {
				s.Pressure.UsageByNode["cpu-a"] = prometheus.NodeResourceUsage{CPUCapacity: 100}
				d.Requirement.RequiredCPU = 100
			},
			reason: PreflightReasonNodePressure,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			snapshot := installablePreflightSnapshot()
			demand := cpuDemand("app", "alice", 1)
			tc.mutate(&snapshot, &demand)
			report, err := SimulatePreflight([]PreflightDemand{demand}, snapshot)
			if err != nil {
				t.Fatalf("simulate: %v", err)
			}
			if report.Installable || report.FailedDemandID != "app" || report.Reason != tc.reason ||
				len(report.Pressure) == 0 {
				t.Fatalf("unexpected report: %#v", report)
			}
		})
	}
}

func TestSimulatePreflightChecksPressureOnUndeclaredDimensions(t *testing.T) {
	snapshot := installablePreflightSnapshot()
	snapshot.Cluster.Memory.Usage = 60 * float64(gi)
	demand := cpuDemand("strict", "alice", 1)

	installPressure, err := apputils.EvaluateClusterPressure(
		apputils.ResourceState{CPU: demand.Requirement.RequiredCPU},
		snapshot.Cluster,
		apputils.ResourceDimensions{CPU: true},
	)
	if err != nil {
		t.Fatalf("evaluate install pressure: %v", err)
	}
	if len(installPressure) != 0 {
		t.Fatalf("real install dimensions should skip undeclared memory: %#v", installPressure)
	}

	report, err := SimulatePreflight([]PreflightDemand{demand}, snapshot)
	if err != nil {
		t.Fatalf("simulate: %v", err)
	}
	if report.Installable || report.FailedDemandID != demand.ID || report.Reason != PreflightReasonClusterPressure {
		t.Fatalf("unexpected report: %#v", report)
	}
	if len(report.Pressure) != 1 || len(report.Pressure[0].Dimensions) != 1 ||
		report.Pressure[0].Dimensions[0].Resource != PressureResourceMemory {
		t.Fatalf("expected memory pressure: %#v", report.Pressure)
	}
}

func TestSimulatePreflightRejectsInvalidSnapshotAndOverflow(t *testing.T) {
	tests := []struct {
		name     string
		snapshot PreflightSnapshot
		demand   PreflightDemand
	}{
		{
			name:     "missing cluster metrics",
			snapshot: PreflightSnapshot{},
			demand:   cpuDemand("missing", "alice", 1),
		},
		{
			name: "missing node metrics",
			snapshot: func() PreflightSnapshot {
				s := installablePreflightSnapshot()
				s.Pressure.UsageByNode = nil
				return s
			}(),
			demand: cpuDemand("missing-node", "alice", 1),
		},
		{
			name: "overflow",
			snapshot: func() PreflightSnapshot {
				s := installablePreflightSnapshot()
				s.Pressure.UsageByNode["cpu-a"] = prometheus.NodeResourceUsage{
					CPUCapacity: math.MaxInt64, CPUUtilization: 1,
				}
				return s
			}(),
			demand: cpuDemand("overflow", "alice", 1),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := SimulatePreflight([]PreflightDemand{tc.demand}, tc.snapshot); err == nil {
				t.Fatal("invalid internal state must return an error")
			}
		})
	}
}

func TestSimulatePreflightDoesNotMutateSnapshot(t *testing.T) {
	snapshot := installablePreflightSnapshot()
	snapshot.Nodes = []Node{nvidiaNode("gpu-a", Device{
		ID: "gpu0", Memory: 16 * gi, Health: deviceHealthYes, SupportType: SupportTypeExclusive,
	})}
	snapshot.Pressure = nodePressure("gpu-a")
	before := clonePreflightSnapshot(snapshot)

	if _, err := SimulatePreflight([]PreflightDemand{gpuDemand("app", "alice", 8*gi)}, snapshot); err != nil {
		t.Fatalf("simulate: %v", err)
	}
	if !reflect.DeepEqual(snapshot, before) {
		t.Fatalf("input snapshot was mutated:\nbefore=%#v\nafter=%#v", before, snapshot)
	}
}

func installablePreflightSnapshot() PreflightSnapshot {
	return PreflightSnapshot{
		Nodes:    []Node{computeNode("cpu-a", utils.CPUType, 64*gi, SupportTypeMemorySlice)},
		Pressure: nodePressure("cpu-a"),
		Cluster: &prometheus.ClusterMetrics{
			CPU:    prometheus.Value{Total: 64},
			Memory: prometheus.Value{Total: 64 * float64(gi)},
			Disk:   prometheus.Value{Total: 1024 * float64(gi)},
		},
		Owners: map[string]*prometheus.ClusterMetrics{
			"alice": {},
			"bob":   {},
		},
		K8sAvailable: apputils.ResourceState{CPU: 64000, Memory: 64 * gi},
	}
}

func cpuDemand(id, owner string, cpu int64) PreflightDemand {
	return PreflightDemand{
		ID: id, Application: id, Owner: owner,
		Requirement: Requirement{Mode: utils.CPUType, RequiredCPU: cpu},
	}
}

func gpuDemand(id, owner string, gpu int64) PreflightDemand {
	return PreflightDemand{
		ID: id, Application: id, Owner: owner,
		Requirement: Requirement{Mode: utils.NvidiaCardType, RequiredGPU: gpu, LimitedGPU: gpu},
	}
}

func nodePressure(nodeNames ...string) PressureSnapshot {
	usage := make(map[string]prometheus.NodeResourceUsage, len(nodeNames))
	for _, nodeName := range nodeNames {
		usage[nodeName] = prometheus.NodeResourceUsage{
			CPUCapacity: 100000, MemoryCapacity: 128 * gi, MemoryAvailable: 128 * gi,
			DiskCapacity: 128 * gi, DiskAvailable: 128 * gi,
		}
	}
	return PressureSnapshot{Threshold: 0.9, UsageByNode: usage}
}
