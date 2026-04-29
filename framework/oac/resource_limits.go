package oac

import (
	"fmt"
	"strings"

	"github.com/beclab/Olares/framework/oac/internal/manifest"
)

// ManifestResourceLimits is the full resource envelope (CPU, memory, disk, GPU
// required/limited pairs) for one spec.resources[] mode row.
type ManifestResourceLimits = manifest.ResourceRequirementLimits

// ResourceLimitsForResourceMode returns required/limited CPU, memory, disk,
// and GPU for the spec.resources[] element whose mode matches
// (case-insensitive). The inline ResourceRequirement on the matched row is
// returned verbatim — empty fields stay empty.
func ResourceLimitsForResourceMode(cfg *AppConfiguration, mode string) (ManifestResourceLimits, error) {
	if cfg == nil {
		return ManifestResourceLimits{}, fmt.Errorf("oac: AppConfiguration is nil")
	}
	for i := range cfg.Spec.Resources {
		rm := &cfg.Spec.Resources[i]
		if strings.EqualFold(rm.Mode, mode) {
			return manifest.ResourceRequirementToLimits(rm.ResourceRequirement), nil
		}
	}
	return ManifestResourceLimits{}, fmt.Errorf("oac: no spec.resources entry with mode %q", mode)
}
