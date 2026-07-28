package compute

import (
	"math"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"
	"k8s.io/apimachinery/pkg/api/resource"

	"pgregory.net/rapid"
)

func oneNodeSnapshot(threshold float64, u prometheus.NodeResourceUsage) (PressureSnapshot, Node) {
	return PressureSnapshot{
			Threshold:   threshold,
			UsageByNode: map[string]prometheus.NodeResourceUsage{"n1": u},
		}, Node{
			NodeName: "n1",
		}
}

// A balanced node already using half of every dimension. Adding the
// app's request is what may or may not push a dimension past 90%.
func halfUsedUsage() prometheus.NodeResourceUsage {
	return prometheus.NodeResourceUsage{
		CPUCapacity:     1000,
		CPUUtilization:  0.5, // used 500m
		MemoryCapacity:  1000,
		MemoryAvailable: 500, // used 500
		DiskCapacity:    1000,
		DiskAvailable:   500, // used 500
	}
}

func TestWouldPressure_PerDimension(t *testing.T) {
	cases := []struct {
		name  string
		added AddedResources
		want  bool
	}{
		{"headroom on every dimension", AddedResources{CPU: 100, Memory: 100, Disk: 100}, false},
		{"cpu pushes over 90%", AddedResources{CPU: 450}, true},
		{"memory pushes over 90%", AddedResources{Memory: 450}, true},
		{"disk pushes over 90%", AddedResources{Disk: 450}, true},
		{"lands exactly on threshold is not over", AddedResources{CPU: 400}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			snap, node := oneNodeSnapshot(0.9, halfUsedUsage())
			if got := snap.WouldPressure(node, tc.added); got != tc.want {
				t.Errorf("WouldPressure(%+v)=%v, want %v", tc.added, got, tc.want)
			}
		})
	}
}

// A node with no reported usage (capacities zero, e.g. metrics missing
// for that node) must never be considered under pressure -- otherwise a
// monitoring gap would block every install.
func TestWouldPressure_UnknownNodeNeverPressured(t *testing.T) {
	snap := PressureSnapshot{Threshold: 0.9, UsageByNode: map[string]prometheus.NodeResourceUsage{}}
	node := Node{NodeName: "missing"}
	if snap.WouldPressure(node, AddedResources{CPU: 1 << 30, Memory: 1 << 40, Disk: 1 << 50}) {
		t.Fatal("node without usage data must not be reported as pressured")
	}
}

// A zero Threshold must fall back to the default 0.90 rather than
// treating everything as over-pressure.
func TestWouldPressure_ZeroThresholdUsesDefault(t *testing.T) {
	usage := prometheus.NodeResourceUsage{CPUCapacity: 1000, CPUUtilization: 0}
	snap, node := oneNodeSnapshot(0, usage)
	if !snap.WouldPressure(node, AddedResources{CPU: 950}) {
		t.Error("950m on an idle 1000m node should exceed the default 0.9 threshold")
	}
	if snap.WouldPressure(node, AddedResources{CPU: 850}) {
		t.Error("850m on an idle 1000m node should stay under the default 0.9 threshold")
	}
}

func TestWouldPressure_OverflowFailsClosed(t *testing.T) {
	snap, node := oneNodeSnapshot(1, prometheus.NodeResourceUsage{
		CPUCapacity:     math.MaxInt64,
		CPUUtilization:  1,
		MemoryCapacity:  math.MaxInt64,
		MemoryAvailable: 0,
		DiskCapacity:    math.MaxInt64,
		DiskAvailable:   0,
	})
	got := snap.Evaluate(node, AddedResources{CPU: 1, Memory: 1, Disk: 1})
	for _, dimension := range got {
		if !dimension.Pressured {
			t.Fatalf("overflow must fail closed: %#v", got)
		}
	}
}

// Pressure is monotonic in the added request: if a smaller request
// already pressures the node, any larger request must pressure it too.
func TestWouldPressure_MonotonicInAddedRequest(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		usage := prometheus.NodeResourceUsage{
			CPUCapacity:     rapid.Int64Range(0, 1<<20).Draw(rt, "cpuCap"),
			CPUUtilization:  rapid.Float64Range(0, 1).Draw(rt, "cpuUtil"),
			MemoryCapacity:  rapid.Int64Range(0, 1<<30).Draw(rt, "memCap"),
			MemoryAvailable: rapid.Int64Range(0, 1<<30).Draw(rt, "memAvail"),
			DiskCapacity:    rapid.Int64Range(0, 1<<30).Draw(rt, "diskCap"),
			DiskAvailable:   rapid.Int64Range(0, 1<<30).Draw(rt, "diskAvail"),
		}
		snap, node := oneNodeSnapshot(rapid.Float64Range(0.1, 1).Draw(rt, "threshold"), usage)

		low := AddedResources{
			CPU:    rapid.Int64Range(0, 1<<20).Draw(rt, "lowCPU"),
			Memory: rapid.Int64Range(0, 1<<30).Draw(rt, "lowMem"),
			Disk:   rapid.Int64Range(0, 1<<30).Draw(rt, "lowDisk"),
		}
		high := AddedResources{
			CPU:    low.CPU + rapid.Int64Range(0, 1<<20).Draw(rt, "dCPU"),
			Memory: low.Memory + rapid.Int64Range(0, 1<<30).Draw(rt, "dMem"),
			Disk:   low.Disk + rapid.Int64Range(0, 1<<30).Draw(rt, "dDisk"),
		}

		if snap.WouldPressure(node, low) && !snap.WouldPressure(node, high) {
			rt.Fatalf("monotonicity violated: low=%+v pressured but high=%+v did not", low, high)
		}
	})
}

func TestAddedResourcesFromAppConfig(t *testing.T) {
	if got := AddedResourcesFromAppConfig(nil); got != (AddedResources{}) {
		t.Errorf("nil config: got %+v, want zero", got)
	}

	legacy := &appcfg.ApplicationConfig{AppName: "legacy"}
	legacy.Requirement.CPU = resource.NewMilliQuantity(1500, resource.DecimalSI)
	legacy.Requirement.Memory = resource.NewQuantity(2<<30, resource.BinarySI)
	legacy.Requirement.Disk = resource.NewQuantity(4<<30, resource.BinarySI)

	mode := &appcfg.ApplicationConfig{
		AppName:         "mode",
		SelectedGpuType: "nvidia",
		Accelerator: []appcfg.ResourceMode{{
			Mode: "nvidia",
			ResourceRequirement: appcfg.ResourceRequirement{
				RequiredCPU:    "2500m",
				RequiredMemory: "3Gi",
				RequiredDisk:   "5Gi",
			},
		}},
	}

	cases := []struct {
		name string
		cfg  *appcfg.ApplicationConfig
		want AddedResources
	}{
		{"legacy quantities", legacy, AddedResources{CPU: 1500, Memory: 2 << 30, Disk: 4 << 30}},
		{"selected resource mode", mode, AddedResources{CPU: 2500, Memory: 3 << 30, Disk: 5 << 30}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := AddedResourcesFromAppConfig(tc.cfg); got != tc.want {
				t.Errorf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestPressureSnapshotEvaluateReportsDimensionDetails(t *testing.T) {
	snap, node := oneNodeSnapshot(0.9, prometheus.NodeResourceUsage{
		CPUCapacity:     2000,
		CPUUtilization:  0.25,
		MemoryCapacity:  1000,
		MemoryAvailable: 400,
		DiskCapacity:    1000,
		DiskAvailable:   1200,
	})

	got := snap.Evaluate(node, AddedResources{CPU: 1400, Memory: 300, Disk: 901})
	want := []DimensionPressure{
		{Resource: PressureResourceCPU, Required: 1400, Used: 500, Capacity: 2000, Available: 1300, Pressured: true},
		{Resource: PressureResourceMemory, Required: 300, Used: 600, Capacity: 1000, Available: 300, Pressured: false},
		{Resource: PressureResourceDisk, Required: 901, Used: 0, Capacity: 1000, Available: 900, Pressured: true},
	}
	if len(got) != len(want) {
		t.Fatalf("Evaluate returned %d dimensions, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("dimension %d: got %#v, want %#v", i, got[i], want[i])
		}
	}
}
