package app

import (
	"errors"
	"math"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const testGi int64 = 1 << 30

func TestEvaluatePhysicalCapacity(t *testing.T) {
	metrics := &prometheus.ClusterMetrics{
		CPU:    prometheus.Value{Total: 2},
		Memory: prometheus.Value{Total: 4 * float64(testGi)},
		Disk:   prometheus.Value{Total: 10 * float64(testGi)},
	}
	pressure, err := EvaluatePhysicalCapacity(ResourceState{CPU: 2001}, metrics, AllResourceDimensions)
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if len(pressure) != 1 || pressure[0].Resource != "cpu" || !pressure[0].Pressured {
		t.Fatalf("unexpected pressure: %#v", pressure)
	}
}

func TestEvaluateClusterPressureUsesThresholdAndDiskReserve(t *testing.T) {
	metrics := &prometheus.ClusterMetrics{
		CPU:    prometheus.Value{Total: 10, Usage: 8},
		Memory: prometheus.Value{Total: 10 * float64(testGi), Usage: 8 * float64(testGi)},
		Disk:   prometheus.Value{Total: 100 * float64(testGi), Usage: 86 * float64(testGi)},
	}
	pressure, err := EvaluateClusterPressure(ResourceState{CPU: 1001}, metrics, AllResourceDimensions)
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if len(pressure) != 2 || pressure[0].Resource != "disk" || pressure[1].Resource != "cpu" {
		t.Fatalf("unexpected pressure: %#v", pressure)
	}
}

func TestEvaluateOwnerPressureTreatsZeroTotalsAsUnlimited(t *testing.T) {
	pressure, err := EvaluateOwnerPressure(ResourceState{CPU: 1000, Memory: testGi}, &prometheus.ClusterMetrics{}, AllResourceDimensions)
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if len(pressure) != 0 {
		t.Fatalf("zero owner quota should be unlimited: %#v", pressure)
	}
}

func TestEvaluateClusterPressureBlocksExistingOveruse(t *testing.T) {
	metrics := &prometheus.ClusterMetrics{
		CPU:    prometheus.Value{Total: 10, Usage: 10},
		Memory: prometheus.Value{Total: 10 * float64(testGi)},
		Disk:   prometheus.Value{Total: 100 * float64(testGi)},
	}
	pressure, err := EvaluateClusterPressure(ResourceState{}, metrics, AllResourceDimensions)
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if len(pressure) != 1 || pressure[0].Resource != "cpu" {
		t.Fatalf("existing overuse must remain pressure: %#v", pressure)
	}
}

func TestEvaluateK8sRequestUsesStrictComparison(t *testing.T) {
	cpu := resource.MustParse("1000m")
	memory := resource.MustParse("1Gi")
	pressure, err := EvaluateK8sRequest(ResourceState{CPU: 1000, Memory: testGi}, AllResourceDimensions, cpu, memory)
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if len(pressure) != 2 || pressure[0].Resource != "cpu" || pressure[1].Resource != "memory" {
		t.Fatalf("equal availability must fail: %#v", pressure)
	}
}

func TestResourceEvaluatorsRejectInvalidMetrics(t *testing.T) {
	if _, err := EvaluateClusterPressure(ResourceState{}, nil, AllResourceDimensions); err == nil {
		t.Fatal("nil metrics must fail")
	}
	metrics := &prometheus.ClusterMetrics{
		CPU:    prometheus.Value{Total: 1},
		Memory: prometheus.Value{Total: 1},
		Disk:   prometheus.Value{Total: 1, Usage: -1},
	}
	if _, err := EvaluateClusterPressure(ResourceState{}, metrics, AllResourceDimensions); err == nil {
		t.Fatal("negative usage must fail")
	}
}

func TestMetricErrorsIdentifyResource(t *testing.T) {
	metrics := &prometheus.ClusterMetrics{
		CPU:    prometheus.Value{Total: 0},
		Memory: prometheus.Value{Total: 10 * float64(testGi)},
		Disk:   prometheus.Value{Total: 100 * float64(testGi)},
	}
	_, err := EvaluatePhysicalCapacity(ResourceState{CPU: 1}, metrics, ResourceDimensions{CPU: true})
	var metricErr *MetricsUnavailableError
	if !errors.As(err, &metricErr) || metricErr.Resource != constants.CPU {
		t.Fatalf("physical capacity error=%v, want cpu MetricsUnavailableError", err)
	}

	metrics.CPU.Total = 10
	metrics.CPU.Usage = math.NaN()
	_, err = EvaluateClusterPressure(ResourceState{}, metrics, ResourceDimensions{CPU: true})
	if !errors.As(err, &metricErr) || metricErr.Resource != constants.CPU {
		t.Fatalf("cluster pressure error=%v, want cpu MetricsUnavailableError", err)
	}
}

func TestMetricRequirementFailureUsesStructuredFriendlyResponse(t *testing.T) {
	resourceType, reason, err, ok := metricRequirementFailure(
		&MetricsUnavailableError{Resource: constants.Memory, Detail: "usage"},
		v1alpha1.InstallOp,
	)
	if !ok {
		t.Fatal("metric error should be recognized")
	}
	if resourceType != constants.Memory || reason != constants.MetricsUnavailable {
		t.Fatalf("resource=%q reason=%q", resourceType, reason)
	}
	if got, want := err.Error(), "Resource metrics are temporarily unavailable. Unable to install the application. Please try again later."; got != want {
		t.Fatalf("message=%q, want %q", got, want)
	}
}

func TestResourceDimensionsPreserveOptionalRequirementSemantics(t *testing.T) {
	metrics := &prometheus.ClusterMetrics{
		CPU:    prometheus.Value{Total: 10, Usage: 10},
		Memory: prometheus.Value{Total: 10 * float64(testGi), Usage: 10 * float64(testGi)},
		Disk:   prometheus.Value{Total: 100 * float64(testGi)},
	}
	pressure, err := EvaluateClusterPressure(ResourceState{}, metrics, ResourceDimensions{})
	if err != nil {
		t.Fatalf("cluster pressure: %v", err)
	}
	if len(pressure) != 0 {
		t.Fatalf("absent cpu and memory requirements must be skipped: %#v", pressure)
	}

	pressure, err = EvaluateOwnerPressure(ResourceState{}, metrics, ResourceDimensions{})
	if err != nil {
		t.Fatalf("owner pressure: %v", err)
	}
	if len(pressure) != 0 {
		t.Fatalf("absent owner requirements must be skipped: %#v", pressure)
	}
}

func TestClusterDiskReserveDoesNotRequireDiskPresence(t *testing.T) {
	metrics := &prometheus.ClusterMetrics{
		CPU:    prometheus.Value{Total: 10},
		Memory: prometheus.Value{Total: 10 * float64(testGi)},
		Disk:   prometheus.Value{Total: 100 * float64(testGi), Usage: 86 * float64(testGi)},
	}
	pressure, err := EvaluateClusterPressure(ResourceState{}, metrics, ResourceDimensions{})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if len(pressure) != 1 || pressure[0].Resource != "disk" {
		t.Fatalf("disk reserve must always apply: %#v", pressure)
	}
}

func TestClusterPressureSkipsAbsentDiskRequirement(t *testing.T) {
	metrics := &prometheus.ClusterMetrics{
		CPU:    prometheus.Value{Total: 10},
		Memory: prometheus.Value{Total: 10 * float64(testGi)},
		Disk:   prometheus.Value{Total: 100 * float64(testGi)},
	}
	pressure, err := EvaluateClusterPressure(ResourceState{Disk: 100 * testGi}, metrics, ResourceDimensions{})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if len(pressure) != 0 {
		t.Fatalf("absent disk requirement must be skipped while reserve still applies: %#v", pressure)
	}
}

func TestClusterPressureChecksPresentDiskRequirement(t *testing.T) {
	metrics := &prometheus.ClusterMetrics{
		Disk: prometheus.Value{Total: 100 * float64(testGi), Usage: 40 * float64(testGi)},
	}
	pressure, err := EvaluateClusterPressure(
		ResourceState{Disk: 51 * testGi},
		metrics,
		ResourceDimensions{Disk: true},
	)
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if len(pressure) != 1 || pressure[0].Resource != "disk" {
		t.Fatalf("present disk requirement must respect headroom: %#v", pressure)
	}
}

func TestEvaluatePhysicalCapacityIgnoresUsage(t *testing.T) {
	metrics := &prometheus.ClusterMetrics{
		CPU:    prometheus.Value{Total: 2, Usage: -1},
		Memory: prometheus.Value{Total: 4 * float64(testGi)},
		Disk:   prometheus.Value{Total: 10 * float64(testGi)},
	}
	if _, err := EvaluatePhysicalCapacity(ResourceState{CPU: 1}, metrics, ResourceDimensions{CPU: true}); err != nil {
		t.Fatalf("usage must not affect physical capacity: %v", err)
	}
	metrics.CPU.Total = 0
	if _, err := EvaluatePhysicalCapacity(ResourceState{CPU: 1}, metrics, ResourceDimensions{CPU: true}); err == nil {
		t.Fatal("selected total must fail closed")
	}
}

func TestEvaluateK8sRequestDistinguishesAbsentAndExplicitZero(t *testing.T) {
	zero := resource.MustParse("0")
	if pressure, err := EvaluateK8sRequest(ResourceState{}, ResourceDimensions{}, zero, zero); err != nil || len(pressure) != 0 {
		t.Fatalf("absent requirements must be skipped: pressure=%#v err=%v", pressure, err)
	}
	pressure, err := EvaluateK8sRequest(ResourceState{}, ResourceDimensions{CPU: true, Memory: true}, zero, zero)
	if err != nil {
		t.Fatalf("explicit zero: %v", err)
	}
	if len(pressure) != 2 {
		t.Fatalf("explicit zero still requires positive availability: %#v", pressure)
	}
}
