package compute

import (
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"

	"k8s.io/apimachinery/pkg/api/resource"
)

// UpgradeResourceDelta returns the non-negative per-dimension increase
// from prev to next app declared requirements. Negative deltas (resource
// reductions) are clamped to zero: upgrade pressure/quota checks only
// ask whether the cluster can absorb additional demand; releasing
// resources does not need a headroom check.
//
// Units match AddedResourcesFromAppConfig: CPU in milli-cores, Memory
// and Disk in bytes.
func UpgradeResourceDelta(prev, next *appcfg.ApplicationConfig) AddedResources {
	old := AddedResourcesFromAppConfig(prev)
	neu := AddedResourcesFromAppConfig(next)
	return AddedResources{
		CPU:    maxInt64(0, neu.CPU-old.CPU),
		Memory: maxInt64(0, neu.Memory-old.Memory),
		Disk:   maxInt64(0, neu.Disk-old.Disk),
	}
}

// AppConfigWithRequirement returns a shallow copy of base whose
// Requirement CPU/Memory/Disk are set from added (nil for zero dims so
// resourceStateFromConfig skips them). Accelerator is cleared so any
// SelectedRequirement lookup cannot override the injected delta.
// OwnerName and other identity fields are preserved for user-quota
// lookups.
func AppConfigWithRequirement(base *appcfg.ApplicationConfig, added AddedResources) *appcfg.ApplicationConfig {
	if base == nil {
		return nil
	}
	cfg := *base
	cfg.Accelerator = nil
	cfg.Requirement = appcfg.AppRequirement{}
	if added.CPU > 0 {
		cfg.Requirement.CPU = resource.NewMilliQuantity(added.CPU, resource.DecimalSI)
	}
	if added.Memory > 0 {
		cfg.Requirement.Memory = resource.NewQuantity(added.Memory, resource.BinarySI)
	}
	if added.Disk > 0 {
		cfg.Requirement.Disk = resource.NewQuantity(added.Disk, resource.BinarySI)
	}
	return &cfg
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
