package manifest

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	ResourceModeCPU           = "cpu"
	ResourceModeAMDAPU        = "amd-apu"
	ResourceModeAMDGPU        = "amd-gpu"
	ResourceModeAppleM        = "apple-m"
	ResourceModeNvidia        = "nvidia"
	ResourceModeNvidiaGB10    = "nvidia-gb10"
	ResourceModeMThreadsM1000 = "mthreads-m1000"
)

var validResourceModes = []any{
	ResourceModeCPU,
	ResourceModeAMDAPU,
	ResourceModeAMDGPU,
	ResourceModeAppleM,
	ResourceModeNvidia,
	ResourceModeNvidiaGB10,
	ResourceModeMThreadsM1000,
}

var modeArchRequirement = map[string]string{
	ResourceModeAMDGPU:        "amd64",
	ResourceModeNvidia:        "amd64",
	ResourceModeNvidiaGB10:    "arm64",
	ResourceModeMThreadsM1000: "arm64",
}

var gpuMemoryModes = map[string]struct{}{
	ResourceModeNvidia: {},
	ResourceModeAMDGPU: {},
}

var minResourcesManifestVersion = semver.MustParse("0.12.0")

// ValidateResourceMode applies per-element rules to a single ResourceMode:
// the `mode` enum, quantity validity, completeness of the inline envelope
// once any field is populated, and the limit >= required pairing.
func ValidateResourceMode(r ResourceMode) error {
	modeInvalid := validation.ValidateStruct(&r,
		validation.Field(&r.Mode,
			validation.Required.Error("resources[].mode is required"),
			validation.In(validResourceModes...).
				Error(fmt.Sprintf("resources[].mode must be one of %s", joinModes(validResourceModes))),
		),
	)
	if modeInvalid != nil {
		return modeInvalid
	}

	_, gpuAllowed := gpuMemoryModes[r.Mode]
	var errs []error
	if err := validateQuantities("", r.ResourceRequirement); err != nil {
		errs = append(errs, err)
	}
	if err := ensureSectionComplete("", r.Mode, r.ResourceRequirement, gpuAllowed); err != nil {
		errs = append(errs, err)
	}
	if !gpuAllowed {
		if err := ensureNoGPUSection("", r.Mode, r.ResourceRequirement); err != nil {
			errs = append(errs, err)
		}
	}
	if err := checkLimitGERequired("", r.ResourceRequirement); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// specResourceCrossFieldRules runs cross-field checks against spec.
// Behaviours split by olaresManifest.version:
//
//   - mutual exclusion (Rule 7): runs at every version. Legacy flat
//     spec.required*/spec.limited* fields cannot coexist with
//     spec.resources[] -- they are alternative expressions of the same
//     envelope, not complementary fields, so emitting this rule on legacy
//     manifests warns users who try to straddle both shapes.
//   - per-entry validation (mode -> supportArch and empty-envelope
//     completeness): only runs when olaresManifest.version >= 0.12.0 and
//     apiVersion is not v2 (v2 forbids spec.resources[] entirely).
//     spec.resources[] is not part of the legacy schema, so its inner
//     contents are intentionally not validated below the gate; users on
//     legacy versions will instead be guided to populate the flat fields
//     by validateAppSpec.
func specResourceCrossFieldRules(configVersion, apiVersion string, spec *AppSpec) error {
	var errs []error
	if err := ensureLegacyAndResourcesAreMutuallyExclusive(spec); err != nil {
		errs = append(errs, err)
	}

	if !resourcesCheckApplies(configVersion) || normalizeAPIVersion(apiVersion) == APIVersionV2 {
		return errors.Join(errs...)
	}

	supportArch := make(map[string]struct{}, len(spec.SupportArch))
	for _, a := range spec.SupportArch {
		supportArch[a] = struct{}{}
	}

	for i, rm := range spec.Resources {
		path := fmt.Sprintf("spec.resources[%d]", i)

		if required, ok := modeArchRequirement[rm.Mode]; ok {
			if _, has := supportArch[required]; !has {
				errs = append(errs, fmt.Errorf(
					"%s: mode=%s requires spec.supportArch to contain %q",
					path, rm.Mode, required,
				))
			}
		}

		if !hasAnyQuantity(rm.ResourceRequirement) {
			_, gpuAllowed := gpuMemoryModes[rm.Mode]
			if err := requireResourceEntryFields(path+".", rm.Mode, gpuAllowed); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

// requireResourceEntryFields produces one error per missing field for a
// completely-empty resource entry. ensureSectionComplete handles partial
// sections; this is its strict counterpart for entries that declare no
// quantities at all.
func requireResourceEntryFields(prefix, mode string, gpuAllowed bool) error {
	pairs := []string{
		"requiredCpu", "limitedCpu",
		"requiredMemory", "limitedMemory",
		"requiredDisk", "limitedDisk",
	}
	if gpuAllowed {
		pairs = append(pairs, "requiredGpu", "limitedGpu")
	}
	var errs []error
	for _, field := range pairs {
		errs = append(errs, fmt.Errorf(
			"%s%s is required to declare a complete resource envelope (mode=%s)",
			prefix, field, mode,
		))
	}
	return errors.Join(errs...)
}

func hasAnyQuantity(rr ResourceRequirement) bool {
	return rr.RequiredCPU != "" || rr.LimitedCPU != "" ||
		rr.RequiredMemory != "" || rr.LimitedMemory != "" ||
		rr.RequiredDisk != "" || rr.LimitedDisk != "" ||
		rr.RequiredGPU != "" || rr.LimitedGPU != ""
}

var legacySpecResourceFields = [...]struct {
	name string
	get  func(*AppSpec) string
}{
	{"spec.requiredCpu", func(s *AppSpec) string { return s.RequiredCPU }},
	{"spec.limitedCpu", func(s *AppSpec) string { return s.LimitedCPU }},
	{"spec.requiredMemory", func(s *AppSpec) string { return s.RequiredMemory }},
	{"spec.limitedMemory", func(s *AppSpec) string { return s.LimitedMemory }},
	{"spec.requiredDisk", func(s *AppSpec) string { return s.RequiredDisk }},
	{"spec.limitedDisk", func(s *AppSpec) string { return s.LimitedDisk }},
	{"spec.requiredGpu", func(s *AppSpec) string { return s.RequiredGPU }},
	{"spec.limitedGpu", func(s *AppSpec) string { return s.LimitedGPU }},
}

// isLegacyEnvelopeMissing reports whether the five legacy flat fields
// that olaresManifest.version < 0.12.0 demands -- spec.requiredCpu /
// spec.limitedCpu / spec.requiredMemory / spec.limitedMemory /
// spec.requiredDisk -- are all empty. validateAppSpec uses this signal at
// the legacy branch to short-circuit the per-field "is required" cascade
// and emit a single consolidated guidance message instead. spec.limited*
// disk and spec.required*/limited*Gpu are not part of the must-fill set
// at legacy versions and are intentionally excluded from the check.
func isLegacyEnvelopeMissing(spec *AppSpec) bool {
	return spec.RequiredCPU == "" &&
		spec.LimitedCPU == "" &&
		spec.RequiredMemory == "" &&
		spec.LimitedMemory == "" &&
		spec.RequiredDisk == ""
}

// ensureLegacyAndResourcesAreMutuallyExclusive reports each legacy flat
// spec.required*/spec.limited* field that is set when spec.resources[] is
// also populated. The two are alternative shapes of the same resource
// envelope and cannot coexist on a single manifest at any
// olaresManifest.version. When spec.resources[] is empty the legacy flat
// fields are unconstrained here -- callers (validateAppSpec) decide
// whether they are required or simply optional for the version in play.
func ensureLegacyAndResourcesAreMutuallyExclusive(spec *AppSpec) error {
	if len(spec.Resources) == 0 {
		return nil
	}
	var errs []error
	for _, f := range legacySpecResourceFields {
		if f.get(spec) != "" {
			errs = append(errs, fmt.Errorf(
				"%s must be empty when spec.resources[] is set; pick one or the other",
				f.name,
			))
		}
	}
	return errors.Join(errs...)
}

// resourcesCheckApplies return true if configVersion >= 0.12.0
func resourcesCheckApplies(v string) bool {
	got, err := semver.NewVersion(v)
	if err != nil {
		return false
	}
	return got.GreaterThanEqual(minResourcesManifestVersion)
}

func IsModernResourcesManifest(version string) bool {
	return resourcesCheckApplies(version)
}

// ResourceRequirementLimits holds required/limited CPU, memory, disk, and GPU
// quantities as unparsed Kubernetes quantity strings.
type ResourceRequirementLimits struct {
	RequiredCPU    string
	LimitedCPU     string
	RequiredMemory string
	LimitedMemory  string
	RequiredDisk   string
	LimitedDisk    string
	RequiredGPU    string
	LimitedGPU     string
}

// ResourceRequirementToLimits flattens an inline ResourceRequirement into
// the eight-field ResourceRequirementLimits envelope used by limit-comparison
// helpers.
func ResourceRequirementToLimits(rr ResourceRequirement) ResourceRequirementLimits {
	return ResourceRequirementLimits{
		RequiredCPU:    rr.RequiredCPU,
		LimitedCPU:     rr.LimitedCPU,
		RequiredMemory: rr.RequiredMemory,
		LimitedMemory:  rr.LimitedMemory,
		RequiredDisk:   rr.RequiredDisk,
		LimitedDisk:    rr.LimitedDisk,
		RequiredGPU:    rr.RequiredGPU,
		LimitedGPU:     rr.LimitedGPU,
	}
}

func ensureSectionComplete(sectionPath, mode string, rr ResourceRequirement, gpuAllowed bool) error {
	if !hasAnyQuantity(rr) {
		return nil
	}
	pairs := []struct {
		field string
		value string
	}{
		{"requiredCpu", rr.RequiredCPU},
		{"limitedCpu", rr.LimitedCPU},
		{"requiredMemory", rr.RequiredMemory},
		{"limitedMemory", rr.LimitedMemory},
		{"requiredDisk", rr.RequiredDisk},
		{"limitedDisk", rr.LimitedDisk},
	}
	if gpuAllowed {
		pairs = append(pairs,
			struct {
				field string
				value string
			}{"requiredGpu", rr.RequiredGPU},
			struct {
				field string
				value string
			}{"limitedGpu", rr.LimitedGPU},
		)
	}
	var errs []error
	for _, p := range pairs {
		if p.value != "" {
			continue
		}
		errs = append(errs, fmt.Errorf(
			"%s%s is required to declare a complete resource envelope (mode=%s)",
			sectionPath, p.field, mode,
		))
	}
	return errors.Join(errs...)
}

func ensureNoGPUSection(sectionPath, mode string, rr ResourceRequirement) error {
	var errs []error
	pairs := []struct {
		field string
		value string
	}{
		{"requiredGpu", rr.RequiredGPU},
		{"limitedGpu", rr.LimitedGPU},
	}
	for _, p := range pairs {
		if p.value == "" {
			continue
		}
		errs = append(errs, fmt.Errorf(
			"%s%s must be empty for mode=%s (only %s support a standalone gpu memory requirement)",
			sectionPath, p.field, mode, joinGPUMemoryModes(),
		))
	}
	return errors.Join(errs...)
}

func validateQuantities(prefix string, rr ResourceRequirement) error {
	pairs := []struct {
		field string
		value string
	}{
		{"requiredCpu", rr.RequiredCPU},
		{"limitedCpu", rr.LimitedCPU},
		{"requiredMemory", rr.RequiredMemory},
		{"limitedMemory", rr.LimitedMemory},
		{"requiredDisk", rr.RequiredDisk},
		{"limitedDisk", rr.LimitedDisk},
		{"requiredGpu", rr.RequiredGPU},
		{"limitedGpu", rr.LimitedGPU},
	}
	var errs []error
	for _, p := range pairs {
		if p.value == "" {
			continue
		}
		if !k8sQuantity.MatchString(p.value) {
			errs = append(errs, fmt.Errorf("%s%s must be a valid Kubernetes quantity (got %q)", prefix, p.field, p.value))
		}
	}
	return errors.Join(errs...)
}

func checkLimitGERequired(prefix string, rr ResourceRequirement) error {
	dims := []struct {
		name     string
		required string
		limited  string
	}{
		{"cpu", rr.RequiredCPU, rr.LimitedCPU},
		{"memory", rr.RequiredMemory, rr.LimitedMemory},
		{"disk", rr.RequiredDisk, rr.LimitedDisk},
		{"gpu", rr.RequiredGPU, rr.LimitedGPU},
	}
	var errs []error
	for _, d := range dims {
		if d.required == "" || d.limited == "" {
			continue
		}
		req, err := resource.ParseQuantity(d.required)
		if err != nil {
			continue
		}
		lim, err := resource.ParseQuantity(d.limited)
		if err != nil {
			continue
		}
		if lim.Cmp(req) < 0 {
			errs = append(errs, fmt.Errorf(
				"%slimited%s (%s) must be >= required%s (%s)",
				prefix, capFirst(d.name), d.limited, capFirst(d.name), d.required,
			))
		}
	}
	return errors.Join(errs...)
}

func capFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func joinModes(modes []any) string {
	parts := make([]string, 0, len(modes))
	for _, m := range modes {
		parts = append(parts, fmt.Sprintf("%q", m))
	}
	return strings.Join(parts, ", ")
}

func joinGPUMemoryModes() string {
	names := make([]string, 0, len(gpuMemoryModes))
	for m := range gpuMemoryModes {
		names = append(names, fmt.Sprintf("mode=%s", m))
	}
	sort.Strings(names)
	return strings.Join(names, " / ")
}
