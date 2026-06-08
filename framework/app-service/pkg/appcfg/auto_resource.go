package appcfg

import "strings"

// AutoResourceValue is the sentinel a manifest author writes into a
// spec.resources / accelerator resource field (e.g. requiredGPUMemory) to
// declare "this value is not known until the app is instantiated; compute it
// from the rendered chart at install time".
//
// It is used by template-style apps where the concrete
// resource demand depends on a user-supplied choice (model, gpu memory, ...)
// injected as an appenv and only materializes after the chart is rendered.

const AutoResourceValue = "-1"

// IsAutoResource reports whether a resource field value is the auto-compute
// sentinel.
func IsAutoResource(s string) bool {
	return strings.TrimSpace(s) == AutoResourceValue
}

// ResourceRequirementHasAuto reports whether any field of the given
// ResourceRequirement is the auto-compute sentinel.
func ResourceRequirementHasAuto(rr ResourceRequirement) bool {
	return IsAutoResource(rr.RequiredCPU) || IsAutoResource(rr.LimitedCPU) ||
		IsAutoResource(rr.RequiredMemory) || IsAutoResource(rr.LimitedMemory) ||
		IsAutoResource(rr.RequiredDisk) || IsAutoResource(rr.LimitedDisk) ||
		IsAutoResource(rr.RequiredGPU) || IsAutoResource(rr.LimitedGPU)
}

// HasAutoResource reports whether the app declares any auto-compute resource
// field in its accelerator/resources matrix. Legacy apps (no Accelerator
// matrix) never use the sentinel, so only the explicit matrix is inspected.
func (c *ApplicationConfig) HasAutoResource() bool {
	for _, mode := range c.Accelerator {
		if ResourceRequirementHasAuto(mode.ResourceRequirement) {
			return true
		}
	}
	return false
}
