package validation

import (
	"context"
	"math"
	"reflect"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/compute"
	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestPreflightSnapshotCollectorCollectsRawFacts(t *testing.T) {
	var nodesCalls, pressureCalls, clusterCalls, k8sCalls int
	userCalls := map[string]int{}
	cpu := resource.MustParse("3500m")
	memory := resource.MustParse("12Gi")
	cluster := &prometheus.ClusterMetrics{
		CPU:    prometheus.Value{Total: 8, Usage: 2},
		Memory: prometheus.Value{Total: 32 << 30, Usage: 8 << 30},
		Disk:   prometheus.Value{Total: 1 << 40, Usage: 100 << 30},
	}
	collector := &PreflightSnapshotCollector{
		nodes: func(context.Context, client.Client) ([]compute.Node, error) {
			nodesCalls++
			return []compute.Node{{NodeName: "node-a"}}, nil
		},
		pressure: func(context.Context) (compute.PressureSnapshot, error) {
			pressureCalls++
			return compute.PressureSnapshot{
				Threshold: 0.9,
				UsageByNode: map[string]prometheus.NodeResourceUsage{
					"node-a": {},
				},
			}, nil
		},
		clusterMetrics: func(string) (*prometheus.ClusterMetrics, []string, error) {
			clusterCalls++
			return cluster, nil, nil
		},
		userMetrics: func(_ context.Context, owner string) (*prometheus.ClusterMetrics, error) {
			userCalls[owner]++
			return &prometheus.ClusterMetrics{CPU: prometheus.Value{Total: 4}}, nil
		},
		k8sAvailable: func() (resource.Quantity, resource.Quantity, error) {
			k8sCalls++
			return cpu, memory, nil
		},
	}

	snapshot, err := collector.Collect(context.Background(), "token", []string{"alice", "shared", "bob", "alice"})
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	if nodesCalls != 1 || pressureCalls != 1 || clusterCalls != 1 || k8sCalls != 1 {
		t.Fatalf("source calls nodes=%d pressure=%d cluster=%d k8s=%d", nodesCalls, pressureCalls, clusterCalls, k8sCalls)
	}
	if !reflect.DeepEqual(userCalls, map[string]int{"alice": 1, "bob": 1}) {
		t.Fatalf("user calls=%v", userCalls)
	}
	if snapshot.Cluster != cluster || len(snapshot.Nodes) != 1 || snapshot.Pressure.Threshold != 0.9 {
		t.Fatalf("raw snapshot facts not preserved: %#v", snapshot)
	}
	if len(snapshot.Owners) != 2 || snapshot.Owners["alice"].CPU.Total != 4 {
		t.Fatalf("owner facts=%#v", snapshot.Owners)
	}
	if snapshot.K8sAvailable.CPU != 3500 || snapshot.K8sAvailable.Memory != 12<<30 {
		t.Fatalf("k8s facts=%#v", snapshot.K8sAvailable)
	}
}

func TestPreflightSnapshotCollectorRejectsInvalidMetrics(t *testing.T) {
	tests := []prometheus.ClusterMetrics{
		{Memory: prometheus.Value{Total: 1}, Disk: prometheus.Value{Total: 1}},
		{CPU: prometheus.Value{Total: math.NaN()}, Memory: prometheus.Value{Total: 1}, Disk: prometheus.Value{Total: 1}},
		{CPU: prometheus.Value{Total: 1}, Memory: prometheus.Value{Total: 1}, Disk: prometheus.Value{Total: 1, Usage: -1}},
	}
	for _, metrics := range tests {
		collector := validCollector(&metrics)
		if _, err := collector.Collect(context.Background(), "", nil); err == nil {
			t.Fatalf("invalid metrics must fail: %#v", metrics)
		}
	}
}

func TestPreflightSnapshotCollectorRejectsInvalidOwnerMetrics(t *testing.T) {
	metrics := &prometheus.ClusterMetrics{
		CPU:    prometheus.Value{Total: 1},
		Memory: prometheus.Value{Total: 1},
		Disk:   prometheus.Value{Total: 1},
	}
	collector := validCollector(metrics)
	collector.userMetrics = func(context.Context, string) (*prometheus.ClusterMetrics, error) {
		return nil, nil
	}
	if _, err := collector.Collect(context.Background(), "", []string{"alice"}); err == nil {
		t.Fatal("missing owner metrics must fail closed")
	}
}

func TestPreflightSnapshotCollectorRejectsMissingNodeMetrics(t *testing.T) {
	metrics := &prometheus.ClusterMetrics{
		CPU:    prometheus.Value{Total: 1},
		Memory: prometheus.Value{Total: 1},
		Disk:   prometheus.Value{Total: 1},
	}
	collector := validCollector(metrics)
	collector.pressure = func(context.Context) (compute.PressureSnapshot, error) {
		return compute.PressureSnapshot{Threshold: 0.9}, nil
	}
	if _, err := collector.Collect(context.Background(), "", nil); err == nil {
		t.Fatal("missing node metrics must fail snapshot collection")
	}
}

func validCollector(metrics *prometheus.ClusterMetrics) *PreflightSnapshotCollector {
	return &PreflightSnapshotCollector{
		nodes: func(context.Context, client.Client) ([]compute.Node, error) {
			return []compute.Node{{NodeName: "node-a"}}, nil
		},
		pressure: func(context.Context) (compute.PressureSnapshot, error) {
			return compute.PressureSnapshot{
				Threshold: 0.9,
				UsageByNode: map[string]prometheus.NodeResourceUsage{
					"node-a": {},
				},
			}, nil
		},
		clusterMetrics: func(string) (*prometheus.ClusterMetrics, []string, error) {
			return metrics, nil, nil
		},
		k8sAvailable: func() (resource.Quantity, resource.Quantity, error) {
			return resource.MustParse("1"), resource.MustParse("1"), nil
		},
	}
}
