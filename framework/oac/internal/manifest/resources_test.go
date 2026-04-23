package manifest

import (
	"fmt"
	"strings"
	"testing"
)

// newResourcesConfig builds a config that exercises the resource block at a
// version above the gate (so checkResources actually runs). Because the
// modern path rejects every legacy flat spec.required*/spec.limited* field
// (Rule 7), this helper explicitly clears the legacy quantities populated
// by newValidConfig.
func newResourcesConfig(modes ...ResourceMode) *AppConfiguration {
	c := newValidConfig()
	c.ConfigVersion = "0.13.0" // >= 0.12.0 -> rules apply
	c.APIVersion = APIVersionV1
	c.Spec.SupportArch = []string{"amd64", "arm64"}
	c.Spec.Resources = modes
	c.Spec.RequiredCPU = ""
	c.Spec.LimitedCPU = ""
	c.Spec.RequiredMemory = ""
	c.Spec.LimitedMemory = ""
	c.Spec.RequiredDisk = ""
	c.Spec.LimitedDisk = ""
	c.Spec.RequiredGPU = ""
	c.Spec.LimitedGPU = ""
	return c
}

func TestResourcesCheckApplies(t *testing.T) {
	cases := []struct {
		v      string
		want   bool
		reason string
	}{
		{"", false, "missing"},
		{"junk", false, "malformed"},
		{"0.11.0", false, "below gate"},
		{"0.12.0", true, "boundary is inclusive"},
		{"0.12.1", true, "above gate"},
		{"1.0.0", true, "well above gate"},
	}
	for _, tc := range cases {
		if got := resourcesCheckApplies(tc.v); got != tc.want {
			t.Errorf("resourcesCheckApplies(%q) = %v, want %v (%s)", tc.v, got, tc.want, tc.reason)
		}
	}
}

func TestIsModernResourcesManifest(t *testing.T) {
	cases := []struct {
		v    string
		want bool
	}{
		{"", false},
		{"junk", false},
		{"0.11.0", false},
		{"0.12.0", true},
		{"0.12.1", true},
		{"1.0.0", true},
	}
	for _, tc := range cases {
		if got := IsModernResourcesManifest(tc.v); got != tc.want {
			t.Errorf("IsModernResourcesManifest(%q) = %v, want %v", tc.v, got, tc.want)
		}
	}
}

// ResourceRequirementToLimits flattens an inline ResourceRequirement into
// the eight-field limits envelope verbatim — empty fields stay empty.
func TestResourceRequirementToLimits_Inline(t *testing.T) {
	rr := ResourceRequirement{
		RequiredCPU: "100m", LimitedCPU: "200m",
		RequiredMemory: "256Mi", LimitedMemory: "512Mi",
		RequiredDisk: "1Gi", LimitedDisk: "2Gi",
	}
	got := ResourceRequirementToLimits(rr)
	want := ResourceRequirementLimits{
		RequiredCPU: "100m", LimitedCPU: "200m",
		RequiredMemory: "256Mi", LimitedMemory: "512Mi",
		RequiredDisk: "1Gi", LimitedDisk: "2Gi",
	}
	if got != want {
		t.Fatalf("ResourceRequirementToLimits mismatch: got %+v, want %+v", got, want)
	}
}

func TestResourceRequirementToLimits_Empty(t *testing.T) {
	got := ResourceRequirementToLimits(ResourceRequirement{})
	if got != (ResourceRequirementLimits{}) {
		t.Fatalf("empty ResourceRequirement must yield zero struct, got %+v", got)
	}
}

func TestResourceMode_Required(t *testing.T) {
	rm := ResourceMode{
		Mode: "",
		ResourceRequirement: ResourceRequirement{
			RequiredCPU: "100m",
		},
	}
	if err := ValidateResourceMode(rm); err == nil {
		t.Fatal("expected error: mode is required")
	}
}

func TestResourceMode_BadMode(t *testing.T) {
	rm := ResourceMode{Mode: "rocm-mi300"}
	err := ValidateResourceMode(rm)
	if err == nil {
		t.Fatal("expected error for unknown mode")
	}
	if !strings.Contains(err.Error(), "must be one of") {
		t.Fatalf("error should list valid modes, got: %v", err)
	}
}

func TestResourceMode_BadQuantity(t *testing.T) {
	rm := ResourceMode{
		Mode: ResourceModeCPU,
		ResourceRequirement: ResourceRequirement{
			RequiredCPU:    "lots",
			RequiredMemory: "200Mi",
		},
	}
	err := ValidateResourceMode(rm)
	if err == nil {
		t.Fatal("expected quantity parse error")
	}
	if !strings.Contains(err.Error(), "requiredCpu") {
		t.Fatalf("error should mention requiredCpu, got: %v", err)
	}
}

// Rule 1: chip family <-> required CPU arch.
func TestRule1_ModeArchRequirement(t *testing.T) {
	cases := []struct {
		name        string
		mode        string
		supportArch []string
		wantErr     bool
		wantNeed    string // arch substring expected in the error
	}{
		{"nvidia needs amd64", ResourceModeNvidia, []string{"arm64"}, true, "amd64"},
		{"nvidia ok with amd64", ResourceModeNvidia, []string{"amd64"}, false, ""},
		{"amd-gpu needs amd64", ResourceModeAMDGPU, []string{"arm64"}, true, "amd64"},
		{"nvidia-gb10 needs arm64", ResourceModeNvidiaGB10, []string{"amd64"}, true, "arm64"},
		{"mthreads-m1000 needs arm64", ResourceModeMThreadsM1000, []string{"amd64"}, true, "arm64"},
		{"cpu has no arch constraint", ResourceModeCPU, []string{"amd64"}, false, ""},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rr := fullCPUMemoryDisk()
			if _, gpu := gpuMemoryModes[tc.mode]; gpu {
				rr.RequiredGPU = "1Gi"
				rr.LimitedGPU = "1Gi"
			}
			c := newResourcesConfig(ResourceMode{
				Mode:                tc.mode,
				ResourceRequirement: rr,
			})
			c.Spec.SupportArch = tc.supportArch
			err := ValidateAppConfiguration(c)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), tc.wantNeed) {
					t.Fatalf("error should mention required arch %q, got: %v", tc.wantNeed, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// fullCPUMemoryDisk returns a ResourceRequirement with every cpu/memory/
// disk field populated so that ensureSectionComplete does not trip. Tests
// add gpu fields on top when the mode is gpu-capable.
func fullCPUMemoryDisk() ResourceRequirement {
	return ResourceRequirement{
		RequiredCPU:    "100m",
		LimitedCPU:     "200m",
		RequiredMemory: "100Mi",
		LimitedMemory:  "200Mi",
		RequiredDisk:   "1Gi",
		LimitedDisk:    "2Gi",
	}
}

// Rule 2: completeness on inline GPU declaration. nvidia (GPU-capable)
// must declare both requiredGpu and limitedGpu once any GPU field is
// populated; the assertion uses fullCPUMemoryDisk so no other rule fires.
func TestRule2_NvidiaModeRequiresGPUPair(t *testing.T) {
	rr := fullCPUMemoryDisk()
	rr.RequiredGPU = "1Gi"
	c := newResourcesConfig(ResourceMode{
		Mode:                ResourceModeNvidia,
		ResourceRequirement: rr,
	})
	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("expected error: nvidia must declare both requiredGpu and limitedGpu")
	}
	if !strings.Contains(err.Error(), "limitedGpu") {
		t.Fatalf("error should mention limitedGpu, got: %v", err)
	}
}

func TestRule2_NonGPUModeAcceptsCompleteEnvelope(t *testing.T) {
	// Non-GPU-memory modes (cpu / amd-apu / apple-m / nvidia-gb10 /
	// mthreads-m1000) cannot declare standalone gpu fields. A fully
	// populated cpu/memory/disk envelope therefore must validate cleanly.
	c := newResourcesConfig(ResourceMode{
		Mode:                ResourceModeCPU,
		ResourceRequirement: fullCPUMemoryDisk(),
	})
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("cpu mode with complete envelope must be accepted: %v", err)
	}
}

func TestRule2_InlineMissingFieldRejected(t *testing.T) {
	// The completeness rule mandates every cpu/memory/disk field on any
	// populated section. Missing a single side (required or limited)
	// must be flagged.
	rr := fullCPUMemoryDisk()
	rr.LimitedCPU = ""
	c := newResourcesConfig(ResourceMode{
		Mode:                ResourceModeCPU,
		ResourceRequirement: rr,
	})
	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "limitedCpu is required") {
		t.Fatalf("error should mention limitedCpu, got: %v", err)
	}
}

// Rule 3: modes outside the gpuMemoryModes whitelist (nvidia / amd-gpu)
// cannot declare gpu fields. Disk is allowed on every mode after the
// relaxation of the old "cpu+memory only" constraint.
func TestRule3_NonGPUFamilyModesForbidGPU(t *testing.T) {
	for _, mode := range []string{
		ResourceModeCPU, ResourceModeAMDAPU, ResourceModeAppleM, ResourceModeNvidiaGB10, ResourceModeMThreadsM1000,
	} {
		mode := mode
		t.Run(mode+"_disk_allowed", func(t *testing.T) {
			// Disk used to be rejected for these modes. It is now accepted
			// alongside cpu/memory.
			c := newResourcesConfig(ResourceMode{
				Mode: mode,
				ResourceRequirement: ResourceRequirement{
					RequiredCPU:  "100m",
					RequiredDisk: "1Gi",
					LimitedDisk:  "2Gi",
				},
			})
			if err := ValidateAppConfiguration(c); err != nil {
				if strings.Contains(err.Error(), "requiredDisk") ||
					strings.Contains(err.Error(), "limitedDisk") {
					t.Fatalf("disk must be allowed for mode=%s, got: %v", mode, err)
				}
			}
		})
		t.Run(mode+"_gpu_forbidden", func(t *testing.T) {
			c := newResourcesConfig(ResourceMode{
				Mode: mode,
				ResourceRequirement: ResourceRequirement{
					RequiredCPU: "100m",
					RequiredGPU: "1",
				},
			})
			err := ValidateAppConfiguration(c)
			if err == nil {
				t.Fatalf("expected error: gpu not allowed for mode=%s", mode)
			}
			if !strings.Contains(err.Error(), "requiredGpu") {
				t.Fatalf("error should mention requiredGpu, got: %v", err)
			}
		})
	}
}

// Rule 4: nvidia and amd-gpu may declare a standalone gpu memory quantity;
// every other mode must leave requiredGpu/limitedGpu empty.
func TestRule4_GPUMemoryAllowedModes(t *testing.T) {
	cases := []struct {
		mode    string
		wantErr bool
	}{
		{ResourceModeNvidia, false},
		{ResourceModeAMDGPU, false},
		{ResourceModeCPU, true},           // non-GPU family
		{ResourceModeMThreadsM1000, true}, // GPU-family but not yet whitelisted
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.mode, func(t *testing.T) {
			rr := fullCPUMemoryDisk()
			rr.RequiredGPU = "8Gi"
			rr.LimitedGPU = "8Gi"
			c := newResourcesConfig(ResourceMode{
				Mode:                tc.mode,
				ResourceRequirement: rr,
			})
			err := ValidateAppConfiguration(c)
			if tc.wantErr && err == nil {
				t.Fatalf("mode=%s: expected error for gpu declaration, got nil", tc.mode)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("mode=%s: unexpected error: %v", tc.mode, err)
			}
		})
	}
}

// Rule 5: limit >= required within the same dimension. Error cases only
// need to populate enough fields to trigger the comparison; completeness
// will also complain, but the assertion only checks for ANY error. The
// passing cases use the full envelope so they do not trip completeness.
func TestRule5_LimitGERequired(t *testing.T) {
	full := fullCPUMemoryDisk()
	flippedCPU := full
	flippedCPU.RequiredCPU = "200m"
	flippedCPU.LimitedCPU = "100m"

	flippedMemory := full
	flippedMemory.RequiredMemory = "200Mi"
	flippedMemory.LimitedMemory = "100Mi"

	cases := []struct {
		name string
		rr   ResourceRequirement
		want bool // true => expect error
	}{
		{"cpu limit < required", flippedCPU, true},
		{"memory limit < required", flippedMemory, true},
		{"limit >= required is fine", full, false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := newResourcesConfig(ResourceMode{
				Mode:                ResourceModeCPU,
				ResourceRequirement: tc.rr,
			})
			err := ValidateAppConfiguration(c)
			if tc.want && err == nil {
				t.Fatal("expected error")
			}
			if !tc.want && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// The version gate: legacy manifests must be left alone.
func TestResources_VersionGateBelowThreshold(t *testing.T) {
	c := newResourcesConfig(ResourceMode{Mode: "completely-bogus"})
	c.ConfigVersion = "0.11.0" // below the gate
	if err := ValidateAppConfiguration(c); err != nil {
		// Mode validation runs via ValidateAppConfiguration -> ValidateResourceMode.
		// which is a per-element rule, BUT the cross-field checkResources is gated.
		// We accept either outcome here as long as the cross-field arch rule
		// (which we'd otherwise expect to trigger) is NOT in the error.
		if strings.Contains(err.Error(), "supportArch") {
			t.Fatalf("cross-field arch rule should be skipped below the gate, got: %v", err)
		}
	}
}

// The completeness rule demands every cpu/memory/disk field on any
// section that is populated at all. Each test case flips one field to
// empty on a fully-populated inline section and expects that exact field
// to show up in the error.
func TestSectionComplete_InlineMissingFieldReported(t *testing.T) {
	fields := []string{
		"requiredCpu", "limitedCpu",
		"requiredMemory", "limitedMemory",
		"requiredDisk", "limitedDisk",
	}
	mutators := map[string]func(*ResourceRequirement){
		"requiredCpu":    func(r *ResourceRequirement) { r.RequiredCPU = "" },
		"limitedCpu":     func(r *ResourceRequirement) { r.LimitedCPU = "" },
		"requiredMemory": func(r *ResourceRequirement) { r.RequiredMemory = "" },
		"limitedMemory":  func(r *ResourceRequirement) { r.LimitedMemory = "" },
		"requiredDisk":   func(r *ResourceRequirement) { r.RequiredDisk = "" },
		"limitedDisk":    func(r *ResourceRequirement) { r.LimitedDisk = "" },
	}
	for _, field := range fields {
		field := field
		t.Run("inline_missing_"+field, func(t *testing.T) {
			rr := fullCPUMemoryDisk()
			mutators[field](&rr)
			c := newResourcesConfig(ResourceMode{
				Mode:                ResourceModeCPU,
				ResourceRequirement: rr,
			})
			err := ValidateAppConfiguration(c)
			if err == nil {
				t.Fatalf("expected error: inline section missing %s", field)
			}
			if !strings.Contains(err.Error(), field+" is required to declare a complete resource envelope") {
				t.Fatalf("error should mention %s, got: %v", field, err)
			}
		})
	}
}

// Modern manifests (>= 0.12.0) require every spec.resources[] entry to
// declare a complete envelope. A bare ResourceMode that names a chip
// family but supplies no quantities must therefore be rejected, with one
// error per missing field so callers see the full picture in one pass.
func TestSectionComplete_EmptySectionRejected(t *testing.T) {
	c := newResourcesConfig(ResourceMode{Mode: ResourceModeCPU})
	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("expected error: empty resource section must be rejected for olaresManifest.version >= 0.12.0")
	}
	for _, field := range []string{
		"requiredCpu", "limitedCpu",
		"requiredMemory", "limitedMemory",
		"requiredDisk", "limitedDisk",
	} {
		want := field + " is required to declare a complete resource envelope"
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error should mention %q, got: %v", want, err)
		}
	}
}

// GPU-capable modes (nvidia / amd-gpu) must declare both gpu sides once
// any field on the section is populated. Only requiredGpu with no
// limitedGpu (or vice versa) is rejected.
func TestSectionComplete_GPUModePairsGPU(t *testing.T) {
	cases := []struct {
		name      string
		rr        ResourceRequirement
		wantField string
	}{
		{
			name: "missing limitedGpu",
			rr: func() ResourceRequirement {
				r := fullCPUMemoryDisk()
				r.RequiredGPU = "8Gi"
				return r
			}(),
			wantField: "limitedGpu",
		},
		{
			name: "missing requiredGpu",
			rr: func() ResourceRequirement {
				r := fullCPUMemoryDisk()
				r.LimitedGPU = "8Gi"
				return r
			}(),
			wantField: "requiredGpu",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := newResourcesConfig(ResourceMode{
				Mode:                ResourceModeNvidia,
				ResourceRequirement: tc.rr,
			})
			err := ValidateAppConfiguration(c)
			if err == nil {
				t.Fatalf("expected error: nvidia must pair %s", tc.wantField)
			}
			if !strings.Contains(err.Error(), tc.wantField+" is required to declare a complete resource envelope") {
				t.Fatalf("error should mention %s, got: %v", tc.wantField, err)
			}
		})
	}
}

// Rule 7: legacy flat spec.required*/spec.limited* fields cannot coexist
// with spec.resources[]. The mutual-exclusion rule fires regardless of
// olaresManifest.version, so the test fixture pins both a populated
// resources entry (to engage the rule) and one legacy flat field per
// subtest (to drive the assertion).
func TestRule7_LegacySpecResourceFieldsRejected(t *testing.T) {
	cases := []struct {
		name  string
		apply func(*AppSpec)
		want  string
	}{
		{"requiredCpu", func(s *AppSpec) { s.RequiredCPU = "100m" }, "spec.requiredCpu"},
		{"limitedCpu", func(s *AppSpec) { s.LimitedCPU = "200m" }, "spec.limitedCpu"},
		{"requiredMemory", func(s *AppSpec) { s.RequiredMemory = "256Mi" }, "spec.requiredMemory"},
		{"limitedMemory", func(s *AppSpec) { s.LimitedMemory = "512Mi" }, "spec.limitedMemory"},
		{"requiredDisk", func(s *AppSpec) { s.RequiredDisk = "1Gi" }, "spec.requiredDisk"},
		{"limitedDisk", func(s *AppSpec) { s.LimitedDisk = "2Gi" }, "spec.limitedDisk"},
		{"requiredGpu", func(s *AppSpec) { s.RequiredGPU = "1" }, "spec.requiredGpu"},
		{"limitedGpu", func(s *AppSpec) { s.LimitedGPU = "2" }, "spec.limitedGpu"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := newResourcesConfig(ResourceMode{
				Mode:                ResourceModeCPU,
				ResourceRequirement: fullCPUMemoryDisk(),
			})
			tc.apply(&c.Spec)
			err := ValidateAppConfiguration(c)
			if err == nil {
				t.Fatalf("expected error: %s cannot coexist with spec.resources[]", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want+" must be empty when spec.resources[] is set") {
				t.Fatalf("error should mention %s, got: %v", tc.want, err)
			}
		})
	}
}

// When spec.resources[] is empty the legacy flat fields are unconstrained
// by the mutual-exclusion rule, regardless of olaresManifest.version. The
// test sets two flat fields on a legacy fixture (no resources entries) and
// asserts the mutual-exclusion error message never appears.
func TestRule7_LegacyFieldsPermittedWithoutResources(t *testing.T) {
	c := newResourcesConfig()
	c.ConfigVersion = "0.11.0"
	c.Spec.RequiredCPU = "100m"
	c.Spec.LimitedMemory = "512Mi"
	if err := ValidateAppConfiguration(c); err != nil {
		if strings.Contains(err.Error(), "must be empty when spec.resources[] is set") {
			t.Fatalf("legacy flat fields must be permitted when spec.resources[] is empty, got: %v", err)
		}
	}
}

// All eight prohibited fields populated simultaneously alongside a
// resources entry must be reported in one shot -- errors.Join lets
// callers see the full picture instead of fixing them one at a time.
func TestRule7_LegacyFieldsAggregated(t *testing.T) {
	c := newResourcesConfig(ResourceMode{
		Mode:                ResourceModeCPU,
		ResourceRequirement: fullCPUMemoryDisk(),
	})
	c.Spec.RequiredCPU = "100m"
	c.Spec.LimitedCPU = "200m"
	c.Spec.RequiredMemory = "256Mi"
	c.Spec.LimitedMemory = "512Mi"
	c.Spec.RequiredDisk = "1Gi"
	c.Spec.LimitedDisk = "2Gi"
	c.Spec.RequiredGPU = "1"
	c.Spec.LimitedGPU = "2"
	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("expected aggregated error listing every coexisting field")
	}
	msg := err.Error()
	for _, field := range []string{
		"spec.requiredCpu", "spec.limitedCpu",
		"spec.requiredMemory", "spec.limitedMemory",
		"spec.requiredDisk", "spec.limitedDisk",
		"spec.requiredGpu", "spec.limitedGpu",
	} {
		if !strings.Contains(msg, field+" must be empty when spec.resources[] is set") {
			t.Fatalf("error should mention %s mutual-exclusion violation, got: %v", field, err)
		}
	}
}

// When spec.resources[] declares an entry whose ResourceRequirement is
// empty (only Mode is set) and all eight legacy flat fields are populated
// at the spec level, validation must aggregate two distinct error classes:
//
//   - Rule 7 mutual exclusion: every legacy flat field must be reported as
//     "must be empty when spec.resources[] is set" (because spec.resources
//     is non-empty and the legacy field is set).
//   - Empty-entry completeness: every standard cpu/memory/disk field on
//     spec.resources[0] must be reported as missing for the declared mode
//     (because the inline ResourceRequirement has no quantities at all,
//     triggering requireResourceEntryFields).
//
// Mode=cpu so the gpu pair must NOT be reported as missing on
// spec.resources[0] -- mode=cpu does not allow a standalone gpu memory
// requirement and Rule 4-empty respects that.
func TestRule7_LegacyFieldsAggregated_EmptyResourceRequirement(t *testing.T) {
	c := newResourcesConfig(ResourceMode{
		Mode: ResourceModeCPU,
	})
	c.Spec.RequiredCPU = "100m"
	c.Spec.LimitedCPU = "200m"
	c.Spec.RequiredMemory = "256Mi"
	c.Spec.LimitedMemory = "512Mi"
	c.Spec.RequiredDisk = "1Gi"
	c.Spec.LimitedDisk = "2Gi"
	c.Spec.RequiredGPU = "1"
	c.Spec.LimitedGPU = "2"

	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("expected aggregated error covering legacy fields + empty resources[0] envelope")
	}
	msg := err.Error()

	for _, field := range []string{
		"spec.requiredCpu", "spec.limitedCpu",
		"spec.requiredMemory", "spec.limitedMemory",
		"spec.requiredDisk", "spec.limitedDisk",
		"spec.requiredGpu", "spec.limitedGpu",
	} {
		if !strings.Contains(msg, field+" must be empty when spec.resources[] is set") {
			t.Fatalf("error should mention %s mutual-exclusion violation, got: %v", field, err)
		}
	}

	for _, field := range []string{
		"requiredCpu", "limitedCpu",
		"requiredMemory", "limitedMemory",
		"requiredDisk", "limitedDisk",
	} {
		want := "spec.resources[0]." + field + " is required to declare a complete resource envelope (mode=cpu)"
		if !strings.Contains(msg, want) {
			t.Fatalf("error should mention %s envelope-completeness violation, got: %v", field, err)
		}
	}

	for _, gpuField := range []string{"requiredGpu", "limitedGpu"} {
		bad := "spec.resources[0]." + gpuField + " is required to declare a complete resource envelope"
		if strings.Contains(msg, bad) {
			t.Fatalf("mode=cpu must not require %s on spec.resources[0], got: %v", gpuField, err)
		}
	}
}

// The mutual-exclusion rule fires regardless of olaresManifest.version: a
// legacy (< 0.12.0) manifest that declares both legacy flat quantities
// and a populated spec.resources[] entry must be rejected too.
func TestRule7_MutualExclusionFiresBelowGate(t *testing.T) {
	c := newResourcesConfig(ResourceMode{
		Mode:                ResourceModeCPU,
		ResourceRequirement: fullCPUMemoryDisk(),
	})
	c.ConfigVersion = "0.11.0"
	c.Spec.RequiredCPU = "100m"
	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("expected error: legacy manifest cannot mix flat fields with spec.resources[]")
	}
	if !strings.Contains(err.Error(), "spec.requiredCpu must be empty when spec.resources[] is set") {
		t.Fatalf("error should mention mutual-exclusion violation, got: %v", err)
	}
}

// Legacy version (< 0.12.0) with all eight legacy flat fields empty AND a
// malformed spec.resources[0] entry (mode set, ResourceRequirement empty).
// At legacy versions spec.resources[] is not part of the recognised schema
// so its inner contents are intentionally not validated -- the user is
// instead steered toward populating the flat fields. The test pins three
// behaviours:
//
//  1. The consolidated legacy guidance fires once (no per-field cascade).
//  2. NO per-entry empty-envelope errors leak out of spec.resources[0]:
//     the version gate inside specResourceCrossFieldRules suppresses the
//     mode -> supportArch and completeness checks below 0.12.0.
//  3. Rule 7 mutual-exclusion still does NOT fire here -- it only
//     triggers when a legacy flat field is actually set alongside
//     spec.resources[], and every legacy field is empty in this fixture.
func TestValidateAppSpec_LegacyEnvelopeMissingWithMalformedResourcesEntry(t *testing.T) {
	c := newResourcesConfig(ResourceMode{
		Mode: ResourceModeCPU,
	})
	c.ConfigVersion = "0.11.0"

	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("expected error: legacy envelope missing")
	}
	msg := err.Error()
	fmt.Println("msg: ", msg)
	legacyGuidance := "spec.requiredCpu / spec.limitedCpu / spec.requiredMemory / spec.limitedMemory / spec.requiredDisk are required for olaresManifest.version < 0.12.0"
	if !strings.Contains(msg, legacyGuidance) {
		t.Fatalf("error should contain consolidated legacy guidance, got: %v", err)
	}

	for _, field := range []string{
		"requiredCpu", "limitedCpu",
		"requiredMemory", "limitedMemory",
		"requiredDisk", "limitedDisk",
	} {
		bad := "spec.resources[0]." + field + " is required to declare a complete resource envelope"
		if strings.Contains(msg, bad) {
			t.Fatalf("spec.resources[] must not be validated below 0.12.0, but %q appeared: %v", bad, err)
		}
	}

	for _, field := range []string{
		"spec.requiredCpu", "spec.limitedCpu",
		"spec.requiredMemory", "spec.limitedMemory",
		"spec.requiredDisk", "spec.limitedDisk",
		"spec.requiredGpu", "spec.limitedGpu",
	} {
		bad := field + " must be empty when spec.resources[] is set"
		if strings.Contains(msg, bad) {
			t.Fatalf("Rule 7 must not fire when %s is empty, got: %v", field, err)
		}
	}

	for _, field := range []string{
		"spec.requiredCpu", "spec.limitedCpu",
		"spec.requiredMemory", "spec.limitedMemory",
		"spec.requiredDisk",
	} {
		bad := field + " is required for olaresManifest.version < 0.12.0."
		if strings.Contains(msg, bad) {
			t.Fatalf("per-field cascade %q must be suppressed by the consolidated guidance, got: %v", bad, err)
		}
	}
}
