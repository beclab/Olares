package compute

import (
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"

	"k8s.io/apimachinery/pkg/api/resource"
)

func legacyConfig(cpuMilli, memory, disk int64) *appcfg.ApplicationConfig {
	cfg := &appcfg.ApplicationConfig{AppName: "app", OwnerName: "alice"}
	if cpuMilli > 0 {
		cfg.Requirement.CPU = resource.NewMilliQuantity(cpuMilli, resource.DecimalSI)
	}
	if memory > 0 {
		cfg.Requirement.Memory = resource.NewQuantity(memory, resource.BinarySI)
	}
	if disk > 0 {
		cfg.Requirement.Disk = resource.NewQuantity(disk, resource.BinarySI)
	}
	return cfg
}

func TestUpgradeResourceDelta(t *testing.T) {
	cases := []struct {
		name string
		prev *appcfg.ApplicationConfig
		next *appcfg.ApplicationConfig
		want AddedResources
	}{
		{
			name: "increase all dims",
			prev: legacyConfig(1000, 1<<30, 2<<30),
			next: legacyConfig(2500, 3<<30, 5<<30),
			want: AddedResources{CPU: 1500, Memory: 2 << 30, Disk: 3 << 30},
		},
		{
			name: "unchanged",
			prev: legacyConfig(1000, 1<<30, 2<<30),
			next: legacyConfig(1000, 1<<30, 2<<30),
			want: AddedResources{},
		},
		{
			name: "decrease clamps to zero",
			prev: legacyConfig(2500, 3<<30, 5<<30),
			next: legacyConfig(1000, 1<<30, 2<<30),
			want: AddedResources{},
		},
		{
			name: "missing prev dims treat as zero",
			prev: &appcfg.ApplicationConfig{AppName: "app"},
			next: legacyConfig(1500, 2<<30, 4<<30),
			want: AddedResources{CPU: 1500, Memory: 2 << 30, Disk: 4 << 30},
		},
		{
			name: "nil prev",
			prev: nil,
			next: legacyConfig(500, 1<<20, 1<<20),
			want: AddedResources{CPU: 500, Memory: 1 << 20, Disk: 1 << 20},
		},
		{
			name: "nil next",
			prev: legacyConfig(500, 1<<20, 1<<20),
			next: nil,
			want: AddedResources{},
		},
		{
			name: "partial increase",
			prev: legacyConfig(1000, 2<<30, 4<<30),
			next: legacyConfig(1000, 3<<30, 2<<30),
			want: AddedResources{Memory: 1 << 30},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := UpgradeResourceDelta(tc.prev, tc.next); got != tc.want {
				t.Errorf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestAppConfigWithRequirement(t *testing.T) {
	if got := AppConfigWithRequirement(nil, AddedResources{CPU: 1}); got != nil {
		t.Fatalf("nil base: got %+v, want nil", got)
	}

	base := &appcfg.ApplicationConfig{
		AppName:         "app",
		OwnerName:       "alice",
		SelectedGpuType: "nvidia",
		Accelerator: []appcfg.ResourceMode{{
			Mode: "nvidia",
			ResourceRequirement: appcfg.ResourceRequirement{
				RequiredCPU:    "4",
				RequiredMemory: "8Gi",
			},
		}},
	}
	base.Requirement.CPU = resource.NewMilliQuantity(4000, resource.DecimalSI)

	got := AppConfigWithRequirement(base, AddedResources{CPU: 500, Memory: 1 << 30})
	if got == nil {
		t.Fatal("got nil")
	}
	if got.OwnerName != "alice" || got.AppName != "app" {
		t.Errorf("identity fields not preserved: %+v", got)
	}
	if got.Accelerator != nil {
		t.Errorf("Accelerator should be cleared, got %+v", got.Accelerator)
	}
	if got.Requirement.CPU == nil || got.Requirement.CPU.MilliValue() != 500 {
		t.Errorf("CPU=%v, want 500m", got.Requirement.CPU)
	}
	if got.Requirement.Memory == nil || got.Requirement.Memory.Value() != 1<<30 {
		t.Errorf("Memory=%v, want 1Gi", got.Requirement.Memory)
	}
	if got.Requirement.Disk != nil {
		t.Errorf("Disk should be nil for zero delta, got %v", got.Requirement.Disk)
	}
	if base.Accelerator == nil || base.Requirement.CPU.MilliValue() != 4000 {
		t.Error("base was mutated")
	}
}
