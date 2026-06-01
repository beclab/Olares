package manifest

import (
	"strings"
	"testing"
)

// newValidConfig is a legacy (olaresManifest.version < 0.12.0) baseline
// fixture: every flat spec.required*/spec.limited* field that the legacy
// validator now requires is populated up front, so individual tests only
// need to mutate the field they care about.
func newValidConfig() *AppConfiguration {
	return &AppConfiguration{
		ConfigVersion: "0.11.0",
		APIVersion:    APIVersionV1,
		Metadata: AppMetaData{
			Name:        "firefox",
			Icon:        "https://example.com/icon.png",
			Description: "a browser",
			Title:       "Firefox",
			Version:     "1.2.3",
		},
		Entrances: []Entrance{{
			Name:       "main",
			Host:       "firefox",
			Port:       8080,
			Title:      "Main",
			Icon:       "https://example.com/entrance.png",
			AuthLevel:  "public",
			OpenMethod: "default",
		}},
		Spec: AppSpec{
			RequiredCPU:    "100m",
			LimitedCPU:     "200m",
			RequiredMemory: "128Mi",
			LimitedMemory:  "256Mi",
			RequiredDisk:   "1Gi",
		},
	}
}

func TestAppConfiguration_Valid(t *testing.T) {
	if err := ValidateAppConfiguration(newValidConfig()); err != nil {
		t.Fatalf("baseline fixture must pass: %v", err)
	}
}

func TestAppConfiguration_MissingConfigVersion(t *testing.T) {
	c := newValidConfig()
	c.ConfigVersion = ""
	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("expected error for missing olaresManifest.version")
	}
	if !strings.Contains(err.Error(), "olaresManifest.version") {
		t.Fatalf("error should mention olaresManifest.version, got: %v", err)
	}
}

func TestAppConfiguration_APIVersionEnum(t *testing.T) {
	c := newValidConfig()
	c.APIVersion = "v99"
	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("expected error for apiVersion=v99")
	}
	if !strings.Contains(err.Error(), "not supported version") {
		t.Fatalf("error should mention not supported version, got: %v", err)
	}

	c = newValidConfig()
	c.APIVersion = ""
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("empty apiVersion should be accepted: %v", err)
	}

	c = newValidConfig()
	c.APIVersion = APIVersionV3
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("apiVersion=v3 should be accepted: %v", err)
	}
}

func TestValidateKnownAPIVersion(t *testing.T) {
	if err := ValidateKnownAPIVersion(""); err != nil {
		t.Fatalf("empty: %v", err)
	}
	for _, v := range []string{"v1", "V1", "v2", "V3", " v3 "} {
		v := v
		t.Run(v, func(t *testing.T) {
			if err := ValidateKnownAPIVersion(v); err != nil {
				t.Fatalf("ValidateKnownAPIVersion(%q): %v", v, err)
			}
		})
	}
	errV0 := ValidateKnownAPIVersion("v0")
	if errV0 == nil {
		t.Fatal("expected error for v0")
	}
	if !strings.Contains(errV0.Error(), "not supported version") {
		t.Fatalf("got: %v", errV0)
	}
}

func TestAppMetaData_VersionSemver(t *testing.T) {
	cases := []struct {
		name    string
		version string
		wantErr bool
	}{
		{"valid semver", "1.2.3", false},
		{"with prerelease", "1.2.3-beta", false},
		{"bad format", "not-a-version", true},
		{"empty", "", true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := newValidConfig()
			c.Metadata.Version = tc.version
			err := ValidateAppConfiguration(c)
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestEntrance_AuthLevelEnum(t *testing.T) {
	cases := []struct {
		value   string
		wantErr bool
	}{
		{"", false},
		{"public", false},
		{"private", false},
		{"bogus", true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.value, func(t *testing.T) {
			c := newValidConfig()
			c.Entrances[0].AuthLevel = tc.value
			err := ValidateAppConfiguration(c)
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestEntrance_OpenMethodEnum(t *testing.T) {
	cases := []struct {
		value   string
		wantErr bool
	}{
		{"", false},
		{"default", false},
		{"iframe", false},
		{"window", false},
		{"popup", true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.value, func(t *testing.T) {
			c := newValidConfig()
			c.Entrances[0].OpenMethod = tc.value
			err := ValidateAppConfiguration(c)
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestEntrance_IconURL(t *testing.T) {
	cases := []struct {
		icon    string
		wantErr bool
	}{
		{"", false},
		{"http://example.com/x.png", false},
		{"https://example.com/x.png", false},
		{"ftp://example.com/x.png", true},
		{"not-a-url", true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.icon, func(t *testing.T) {
			c := newValidConfig()
			c.Entrances[0].Icon = tc.icon
			err := ValidateAppConfiguration(c)
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestEntrance_Required(t *testing.T) {
	c := newValidConfig()
	c.Entrances = nil
	if err := ValidateAppConfiguration(c); err == nil {
		t.Fatal("expected error: entrances is required")
	}
}

func TestEntrance_UniqueNames(t *testing.T) {
	c := newValidConfig()
	c.Entrances = []Entrance{
		{Name: "dup", Host: "alpha", Port: 1, Title: "A"},
		{Name: "dup", Host: "bravo", Port: 2, Title: "B"},
	}
	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("expected error for duplicate entrance names")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("error should mention duplicate, got: %v", err)
	}
}

func TestEntrance_PortNegative(t *testing.T) {
	c := newValidConfig()
	c.Entrances[0].Port = -1
	if err := ValidateAppConfiguration(c); err == nil {
		t.Fatal("expected error: port must be > 0")
	}
}

func TestAppSpec_QuantityFields(t *testing.T) {
	c := newValidConfig()
	c.Spec.RequiredMemory = "not-a-quantity"
	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("expected error for invalid requiredMemory")
	}
}

// validateAppSpec required-vs-optional matrix for olaresManifest.version
// < 0.12.0:
//
//   - requiredMemory, requiredDisk, requiredCpu, limitedMemory, limitedCpu
//     are required. Because isLegacyEnvelopeMissing flips to true the
//     moment any one of the five is empty, a partial fill produces the
//     consolidated guidance message rather than a per-field cascade --
//     the subtests assert on that consolidated message.
//   - requiredGpu, limitedGpu are optional (the manifest still validates
//     when both are absent).
//
// Each subtest blanks out exactly one legacy field on a fully-populated
// fixture and expects the consolidated guidance to fire; the tail
// subtests pin the optional GPU contract (omitted is fine, set must
// parse).
func TestValidateAppSpec_LegacyRequiredFieldsBelowGate(t *testing.T) {
	const legacyVersion = "0.11.0"
	const consolidatedGuidance = "spec.requiredCpu / spec.limitedCpu / spec.requiredMemory / spec.limitedMemory / spec.requiredDisk are required for olaresManifest.version < 0.12.0; populate the legacy resource envelope"

	required := []struct {
		name  string
		clear func(*AppSpec)
	}{
		{name: "requiredMemory", clear: func(s *AppSpec) { s.RequiredMemory = "" }},
		{name: "requiredDisk", clear: func(s *AppSpec) { s.RequiredDisk = "" }},
		{name: "requiredCpu", clear: func(s *AppSpec) { s.RequiredCPU = "" }},
		{name: "limitedMemory", clear: func(s *AppSpec) { s.LimitedMemory = "" }},
		{name: "limitedCpu", clear: func(s *AppSpec) { s.LimitedCPU = "" }},
	}
	for _, tc := range required {
		tc := tc
		t.Run("missing_"+tc.name, func(t *testing.T) {
			c := newValidConfig()
			c.ConfigVersion = legacyVersion
			tc.clear(&c.Spec)
			err := ValidateAppConfiguration(c)
			if err == nil {
				t.Fatalf("expected error: %s missing should fail at olaresManifest.version=%s", tc.name, legacyVersion)
			}
			if !strings.Contains(err.Error(), consolidatedGuidance) {
				t.Fatalf("error should mention the consolidated guidance message, got: %v", err)
			}
		})
	}

	t.Run("baseline_valid_without_gpu_pair", func(t *testing.T) {
		c := newValidConfig()
		c.ConfigVersion = legacyVersion
		c.Spec.RequiredGPU = ""
		c.Spec.LimitedGPU = ""
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("legacy fixture without gpu fields must validate: %v", err)
		}
	})

	t.Run("optional_gpu_when_set_must_be_quantity", func(t *testing.T) {
		c := newValidConfig()
		c.ConfigVersion = legacyVersion
		c.Spec.RequiredGPU = "1"
		c.Spec.LimitedGPU = "2"
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("valid gpu quantities must pass: %v", err)
		}
	})

	t.Run("optional_gpu_invalid_quantity_rejected", func(t *testing.T) {
		c := newValidConfig()
		c.ConfigVersion = legacyVersion
		c.Spec.RequiredGPU = "lots-of-it"
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: requiredGpu must be a Kubernetes quantity when set")
		}
		if !strings.Contains(err.Error(), "RequiredGPU") &&
			!strings.Contains(err.Error(), "requiredGpu") {
			t.Fatalf("error should mention requiredGpu, got: %v", err)
		}
	})

	t.Run("required_invalid_quantity_reported_alongside_required", func(t *testing.T) {
		// Sanity check: the required-rule and the quantity-rule both fire
		// when the field is set to an invalid quantity (the field is not
		// "missing", but it is invalid).
		c := newValidConfig()
		c.ConfigVersion = legacyVersion
		c.Spec.RequiredMemory = "totally-not-a-quantity"
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error for malformed requiredMemory")
		}
		if !strings.Contains(err.Error(), "must be a valid Kubernetes quantity") {
			t.Fatalf("error should mention quantity rule, got: %v", err)
		}
	})

	// When *all* five required legacy fields are absent, the per-field
	// "is required" cascade collapses into a single, version-tagged
	// guidance message so users see one obvious next step instead of five
	// repeated lines. Partial fills (covered above) still get pinpointed
	// per-field errors.
	t.Run("all_required_missing_emits_consolidated_guidance", func(t *testing.T) {
		c := newValidConfig()
		c.ConfigVersion = legacyVersion
		c.Spec.RequiredCPU = ""
		c.Spec.LimitedCPU = ""
		c.Spec.RequiredMemory = ""
		c.Spec.LimitedMemory = ""
		c.Spec.RequiredDisk = ""
		c.Spec.LimitedDisk = ""
		c.Spec.RequiredGPU = ""
		c.Spec.LimitedGPU = ""
		c.Spec.Accelerator = nil

		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected consolidated legacy guidance when nothing is supplied")
		}
		want := "spec.requiredCpu / spec.limitedCpu / spec.requiredMemory / spec.limitedMemory / spec.requiredDisk are required for olaresManifest.version < 0.12.0; populate the legacy resource envelope"
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error should contain consolidated legacy guidance, got: %v", err)
		}

		for _, field := range []string{
			"spec.requiredCpu", "spec.limitedCpu",
			"spec.requiredMemory", "spec.limitedMemory",
			"spec.requiredDisk",
		} {
			cascade := field + " is required for olaresManifest.version < 0.12.0."
			if strings.Contains(err.Error(), cascade) {
				t.Fatalf("per-field cascade %q must be suppressed when the consolidated guidance fires, got: %v", cascade, err)
			}
		}
	})
}

// Modern manifests (olaresManifest.version >= 0.12.0):
//
//   - EITHER spec.resources[] OR the legacy flat envelope is required.
//     If neither is declared, the modern-branch closure emits a single
//     consolidated guidance line listing both options.
//   - When spec.resources[] is declared, each entry must populate every
//     standard field (cpu/memory/disk pair, plus the gpu pair on
//     gpu-capable modes).
//   - The two shapes are mutually exclusive (Rule 7 -- already covered
//     in resources_test.go but pinned here for completeness so a
//     regression on the modern branch lights up locally).
//   - spec.requiredGpu and spec.limitedGpu remain optional at the spec
//     level (with spec.resources[] declared, gpu fields belong inside an
//     entry; with the flat envelope they live at the spec level).
func TestValidateAppSpec_ModernResourcesRequiredAtOrAboveGate(t *testing.T) {
	t.Run("missing_both_shapes_emits_dual_guidance", func(t *testing.T) {
		c := newResourcesConfig() // no modes AND newResourcesConfig clears every flat field
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: at least one resource envelope shape must be declared on a modern manifest")
		}
		want := "either spec.resources[] or the legacy envelope (spec.requiredCpu / spec.limitedCpu / spec.requiredMemory / spec.limitedMemory / spec.requiredDisk) is required for olaresManifest.version >= 0.12.0"
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error should mention modern dual-shape guidance, got: %v", err)
		}
	})

	// completeFields is the modern resources[] envelope every test case
	// below tweaks: it carries cpu/memory/disk pairs so non-gpu modes
	// validate cleanly, and tests opt-in to gpu pairs as needed.
	completeFields := func() ResourceRequirement {
		return ResourceRequirement{
			RequiredCPU:    "100m",
			LimitedCPU:     "200m",
			RequiredMemory: "128Mi",
			LimitedMemory:  "256Mi",
			RequiredDisk:   "1Gi",
			LimitedDisk:    "2Gi",
		}
	}

	t.Run("complete_cpu_entry_valid", func(t *testing.T) {
		c := newResourcesConfig(ResourceMode{
			Mode:                ResourceModeCPU,
			ResourceRequirement: completeFields(),
		})
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("complete cpu entry must validate: %v", err)
		}
	})

	missing := []struct {
		name      string
		clear     func(*ResourceRequirement)
		errSuffix string
	}{
		{
			name:      "requiredCpu",
			clear:     func(r *ResourceRequirement) { r.RequiredCPU = "" },
			errSuffix: "requiredCpu is required to declare a complete resource envelope",
		},
		{
			name:      "limitedCpu",
			clear:     func(r *ResourceRequirement) { r.LimitedCPU = "" },
			errSuffix: "limitedCpu is required to declare a complete resource envelope",
		},
		{
			name:      "requiredMemory",
			clear:     func(r *ResourceRequirement) { r.RequiredMemory = "" },
			errSuffix: "requiredMemory is required to declare a complete resource envelope",
		},
		{
			name:      "limitedMemory",
			clear:     func(r *ResourceRequirement) { r.LimitedMemory = "" },
			errSuffix: "limitedMemory is required to declare a complete resource envelope",
		},
		{
			name:      "requiredDisk",
			clear:     func(r *ResourceRequirement) { r.RequiredDisk = "" },
			errSuffix: "requiredDisk is required to declare a complete resource envelope",
		},
		{
			name:      "limitedDisk",
			clear:     func(r *ResourceRequirement) { r.LimitedDisk = "" },
			errSuffix: "limitedDisk is required to declare a complete resource envelope",
		},
	}
	for _, tc := range missing {
		tc := tc
		t.Run("missing_field_"+tc.name, func(t *testing.T) {
			rr := completeFields()
			tc.clear(&rr)
			c := newResourcesConfig(ResourceMode{
				Mode:                ResourceModeCPU,
				ResourceRequirement: rr,
			})
			err := ValidateAppConfiguration(c)
			if err == nil {
				t.Fatalf("expected error: %s missing on a populated entry must fail", tc.name)
			}
			if !strings.Contains(err.Error(), tc.errSuffix) {
				t.Fatalf("error should mention %q, got: %v", tc.errSuffix, err)
			}
		})
	}

	t.Run("empty_entry_reports_every_missing_field", func(t *testing.T) {
		c := newResourcesConfig(ResourceMode{Mode: ResourceModeCPU})
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: bare ResourceMode entry with no quantities must be rejected")
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
	})

	t.Run("nvidia_entry_requires_gpu_pair", func(t *testing.T) {
		c := newResourcesConfig(ResourceMode{
			Mode:                ResourceModeNvidia,
			ResourceRequirement: completeFields(), // no gpu fields
		})
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: nvidia entry must declare requiredGpu / limitedGpu")
		}
		for _, field := range []string{"requiredGpu", "limitedGpu"} {
			want := field + " is required to declare a complete resource envelope"
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("error should mention %q, got: %v", want, err)
			}
		}
	})

	t.Run("nvidia_entry_with_gpu_pair_valid", func(t *testing.T) {
		rr := completeFields()
		rr.RequiredGPU = "8Gi"
		rr.LimitedGPU = "8Gi"
		c := newResourcesConfig(ResourceMode{
			Mode:                ResourceModeNvidia,
			ResourceRequirement: rr,
		})
		c.Spec.SupportArch = []string{"amd64"}
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("nvidia entry with full envelope + gpu pair must validate: %v", err)
		}
	})

	t.Run("legacy_flat_fields_cannot_coexist_with_resources", func(t *testing.T) {
		c := newResourcesConfig(ResourceMode{
			Mode:                ResourceModeCPU,
			ResourceRequirement: completeFields(),
		})
		c.Spec.RequiredMemory = "256Mi" // mutually exclusive with spec.resources[]
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected Rule 7 error: legacy flat field cannot coexist with spec.resources[]")
		}
		if !strings.Contains(err.Error(), "spec.requiredMemory must be empty when spec.resources[] is set") {
			t.Fatalf("error should mention mutual-exclusion violation, got: %v", err)
		}
	})

	t.Run("spec_level_gpu_remains_optional", func(t *testing.T) {
		// spec.requiredGpu / spec.limitedGpu being unset must not trip
		// any required-field rule at the modern gate -- gpu lives inside
		// spec.resources[] entries instead.
		c := newResourcesConfig(ResourceMode{
			Mode:                ResourceModeCPU,
			ResourceRequirement: completeFields(),
		})
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("modern manifest without spec-level gpu fields must validate: %v", err)
		}
	})
}

// validateAppSpec contract at the 0.12.0 boundary when the manifest opts
// into the legacy flat envelope (spec.requiredCpu / spec.limitedCpu /
// spec.requiredMemory / spec.limitedMemory / spec.requiredDisk + optional
// limitedDisk / requiredGpu / limitedGpu) instead of spec.resources[].
//
// 0.12.0 is the boundary at which the modern code path activates
// (resourcesCheckApplies is inclusive), so this is the case where the
// modern-branch closure (validateResourceModeValueFor) must accept the
// flat shape and route every populated field through
// validateFlatResourceQuantities for K8s-quantity format checking.
// Required-ness of individual flat fields is intentionally not enforced
// here -- a partial envelope still passes format validation, mirroring
// the loose modern semantics the closure documents.
func TestValidateAppSpec_ModernAcceptsLegacyFlatEnvelope(t *testing.T) {
	const modernBoundary = "0.12.0"

	// populated builds a modern (0.12.0) manifest that declares the
	// legacy flat envelope and leaves spec.resources[] empty.
	// newResourcesConfig clears every flat field; we re-populate the
	// five mandatory quantities here.
	populated := func() *AppConfiguration {
		c := newResourcesConfig() // no modes, ConfigVersion=0.13.0, every flat field cleared
		c.ConfigVersion = modernBoundary
		c.APIVersion = APIVersionV1
		c.Spec.RequiredCPU = "100m"
		c.Spec.LimitedCPU = "200m"
		c.Spec.RequiredMemory = "128Mi"
		c.Spec.LimitedMemory = "256Mi"
		c.Spec.RequiredDisk = "1Gi"
		return c
	}

	t.Run("complete_flat_envelope_valid_at_boundary", func(t *testing.T) {
		c := populated()
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("modern manifest at %s using the legacy flat envelope must validate: %v", modernBoundary, err)
		}
	})

	t.Run("complete_flat_envelope_valid_above_boundary", func(t *testing.T) {
		// 0.13.0 sits well above the gate. The closure must take the
		// same branch as the 0.12.0 boundary and accept the flat shape.
		c := populated()
		c.ConfigVersion = "0.13.0"
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("modern manifest at 0.13.0 using the legacy flat envelope must validate: %v", err)
		}
	})

	t.Run("flat_envelope_with_optional_disk_and_gpu_valid", func(t *testing.T) {
		c := populated()
		c.Spec.LimitedDisk = "2Gi"
		c.Spec.RequiredGPU = "1"
		c.Spec.LimitedGPU = "2"
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("modern manifest with optional disk/gpu fields must validate: %v", err)
		}
	})

	// Each populated flat field must parse as a Kubernetes quantity --
	// the closure's validateFlatResourceQuantities walks every non-empty
	// flat field and reports the first that fails the k8sQuantity regex.
	t.Run("invalid_required_quantity_rejected", func(t *testing.T) {
		c := populated()
		c.Spec.RequiredMemory = "totally-not-a-quantity"
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: invalid quantity on the flat envelope must be rejected")
		}
		want := "spec.requiredMemory must be a valid Kubernetes quantity"
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error should mention %q, got: %v", want, err)
		}
	})

	t.Run("invalid_limited_disk_quantity_rejected", func(t *testing.T) {
		// limitedDisk is optional but still quantity-checked when set:
		// confirm the optional fields go through validateFlatResourceQuantities
		// rather than getting silently accepted.
		c := populated()
		c.Spec.LimitedDisk = "not-a-quantity"
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: invalid limitedDisk quantity must be rejected")
		}
		want := "spec.limitedDisk must be a valid Kubernetes quantity"
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error should mention %q, got: %v", want, err)
		}
	})

	// Modern-with-flat must not accidentally trip the modern-with-resources
	// guidance ("either spec.resources[] or the legacy envelope ...") --
	// once any flat field is set, hasAnyFlatResourceQuantity flips to true
	// and the dual-shape guard stays quiet.
	t.Run("no_dual_shape_guidance_when_flat_envelope_set", func(t *testing.T) {
		c := populated()
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("baseline must pass: %v", err)
		}
		// Even a single set flat field should be enough.
		c = newResourcesConfig()
		c.ConfigVersion = modernBoundary
		c.Spec.RequiredCPU = "100m"
		err := ValidateAppConfiguration(c)
		if err != nil && strings.Contains(err.Error(), "either spec.resources[] or the legacy envelope") {
			t.Fatalf("dual-shape guidance must not fire once any flat field is set, got: %v", err)
		}
	})

	// Sanity-check the symmetric mutex contract: a modern manifest that
	// declared the flat envelope but later also adds spec.resources[]
	// must still be caught by Rule 7.
	t.Run("flat_envelope_plus_resources_triggers_mutex", func(t *testing.T) {
		c := populated()
		c.Spec.Accelerator = []ResourceMode{{
			Mode: ResourceModeCPU,
			ResourceRequirement: ResourceRequirement{
				RequiredCPU:    "100m",
				LimitedCPU:     "200m",
				RequiredMemory: "128Mi",
				LimitedMemory:  "256Mi",
				RequiredDisk:   "1Gi",
				LimitedDisk:    "2Gi",
			},
		}}
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected Rule 7 mutex error when both shapes coexist")
		}
		if !strings.Contains(err.Error(), "spec.requiredCpu must be empty when spec.resources[] is set") {
			t.Fatalf("error should mention mutex violation, got: %v", err)
		}
	})
}

// validateAppSpec contract under olaresManifest.version >= 0.12.0 (apiVersion
// != v2) when both shapes coexist on the same manifest:
//
//   - Setting any one of the eight legacy flat fields alongside a populated
//     spec.resources[] entry must trip Rule 7 and report
//     "<field> must be empty when spec.resources[] is set" for exactly that
//     field. The subtests cycle through cpu / memory / disk / gpu pairs so a
//     regression that drops one column lights up locally rather than only
//     showing up in the aggregated TestRule7_* tests over in
//     resources_test.go.
//   - Setting all eight at once must produce all eight errors in a single
//     errors.Join'd return so callers see the full picture without having
//     to fix-and-rerun.
//   - The per-entry validation of spec.resources[] continues to fire
//     alongside the mutex errors -- a malformed resources entry must not
//     mask the mutex violation and vice versa.
func TestValidateAppSpec_ModernRejectsBothShapesCoexisting(t *testing.T) {
	completeFields := func() ResourceRequirement {
		return ResourceRequirement{
			RequiredCPU:    "100m",
			LimitedCPU:     "200m",
			RequiredMemory: "128Mi",
			LimitedMemory:  "256Mi",
			RequiredDisk:   "1Gi",
			LimitedDisk:    "2Gi",
		}
	}

	perFieldCases := []struct {
		name  string
		apply func(*AppSpec)
		field string
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
	for _, tc := range perFieldCases {
		tc := tc
		t.Run("single_field_"+tc.name, func(t *testing.T) {
			c := newResourcesConfig(ResourceMode{
				Mode:                ResourceModeCPU,
				ResourceRequirement: completeFields(),
			})
			tc.apply(&c.Spec)
			err := ValidateAppConfiguration(c)
			if err == nil {
				t.Fatalf("expected Rule 7 violation: %s cannot coexist with spec.resources[]", tc.field)
			}
			want := tc.field + " must be empty when spec.resources[] is set"
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("error should mention %q, got: %v", want, err)
			}
		})
	}

	t.Run("all_eight_flat_fields_aggregated", func(t *testing.T) {
		c := newResourcesConfig(ResourceMode{
			Mode:                ResourceModeCPU,
			ResourceRequirement: completeFields(),
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
			t.Fatal("expected aggregated Rule 7 violations for every coexisting flat field")
		}
		msg := err.Error()
		for _, field := range []string{
			"spec.requiredCpu", "spec.limitedCpu",
			"spec.requiredMemory", "spec.limitedMemory",
			"spec.requiredDisk", "spec.limitedDisk",
			"spec.requiredGpu", "spec.limitedGpu",
		} {
			want := field + " must be empty when spec.resources[] is set"
			if !strings.Contains(msg, want) {
				t.Fatalf("error should mention %s mutual-exclusion violation, got: %v", field, err)
			}
		}
	})

	// Per-entry validation must still surface even when Rule 7 has plenty
	// to say. A malformed resources entry (missing limitedCpu) coexists
	// with a populated spec.requiredCpu -- the user should see both the
	// envelope-completeness error AND the mutex error in one shot.
	t.Run("entry_validation_runs_alongside_mutex_errors", func(t *testing.T) {
		rr := completeFields()
		rr.LimitedCPU = "" // breaks the per-entry envelope
		c := newResourcesConfig(ResourceMode{
			Mode:                ResourceModeCPU,
			ResourceRequirement: rr,
		})
		c.Spec.RequiredCPU = "100m" // trips Rule 7
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected aggregated error covering both mutex and entry validation")
		}
		msg := err.Error()
		if !strings.Contains(msg, "spec.requiredCpu must be empty when spec.resources[] is set") {
			t.Fatalf("error should report Rule 7 mutex violation, got: %v", err)
		}
		if !strings.Contains(msg, "limitedCpu is required to declare a complete resource envelope") {
			t.Fatalf("error should also report per-entry completeness violation, got: %v", err)
		}
	})
}

func TestSubCharts_OnlyEnforcedForV2(t *testing.T) {
	c := newValidConfig()
	c.APIVersion = APIVersionV1
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("v1 should not require subCharts: %v", err)
	}

	c = newValidConfig()
	c.APIVersion = APIVersionV2
	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("v2 without subCharts should fail")
	}
	if !strings.Contains(err.Error(), "subCharts") {
		t.Fatalf("error should mention subCharts, got: %v", err)
	}

	c.Spec.SubCharts = []Chart{{Name: "main", Shared: false}}
	err = ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("v2 needs at least one shared=true subChart")
	}
	if !strings.Contains(err.Error(), "shared=true") {
		t.Fatalf("error should mention shared=true, got: %v", err)
	}

	c.Spec.SubCharts = []Chart{{Name: "main", Shared: true}}
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("v2 with shared subchart should pass: %v", err)
	}
}

func TestSubCharts_V2TriggersRegardlessOfOlaresVersion(t *testing.T) {
	legacyVersions := []string{"0.11.0", "0.11.9"}
	for _, v := range legacyVersions {
		v := v
		t.Run("olaresManifest.version="+v, func(t *testing.T) {
			c := newValidConfig()
			c.ConfigVersion = v
			c.APIVersion = APIVersionV2
			err := ValidateAppConfiguration(c)
			if err == nil {
				t.Fatalf("v2 manifest (olaresManifest.version=%s) without subCharts must fail", v)
			}
			if !strings.Contains(err.Error(), "subCharts is required for apiVersion=v2") {
				t.Fatalf("error should mention subCharts requirement, got: %v", err)
			}
		})
	}
}

func TestAPIVersionV2_ModernOlaresUsesLegacyEnvelope(t *testing.T) {
	c := newValidConfig()
	c.ConfigVersion = "0.13.0"
	c.APIVersion = APIVersionV2
	c.Spec.SubCharts = []Chart{{Name: "main", Shared: true}}
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("v2 with olaresManifest.version >= 0.12.0 and legacy flat fields must validate: %v", err)
	}
}

func TestSupportedGpu_ForbiddenWhenOlaresAtOrAbove012(t *testing.T) {
	c := newResourcesConfig(ResourceMode{
		Mode: ResourceModeCPU,
		ResourceRequirement: ResourceRequirement{
			RequiredCPU:    "100m",
			LimitedCPU:     "200m",
			RequiredMemory: "128Mi",
			LimitedMemory:  "256Mi",
			RequiredDisk:   "1Gi",
			LimitedDisk:    "2Gi",
		},
	})
	c.Spec.SupportedGpu = []any{"nvidia"}
	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("expected error when spec.supportedGpu is set at olaresManifest.version >= 0.12.0")
	}
	if !strings.Contains(err.Error(), "spec.supportedGpu must be empty") {
		t.Fatalf("error should mention supportedGpu rule, got: %v", err)
	}
}

func TestSupportedGpu_AllowedBelow012(t *testing.T) {
	c := newValidConfig()
	c.Spec.SupportedGpu = []any{"nvidia"}
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("spec.supportedGpu with olaresManifest.version < 0.12.0 should validate: %v", err)
	}
}

func TestSupportedGpu_ForbiddenForV2ModernOlares(t *testing.T) {
	c := newValidConfig()
	c.ConfigVersion = "0.13.0"
	c.APIVersion = APIVersionV2
	c.Spec.SubCharts = []Chart{{Name: "main", Shared: true}}
	c.Spec.SupportedGpu = []any{"nvidia"}
	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("expected error: supportedGpu forbidden for olaresManifest.version >= 0.12.0 even on apiVersion=v2")
	}
	if !strings.Contains(err.Error(), "spec.supportedGpu must be empty") {
		t.Fatalf("error should mention supportedGpu rule, got: %v", err)
	}
}

func TestAPIVersionV2_ModernOlaresRejectsSpecResources(t *testing.T) {
	c := newResourcesConfig(ResourceMode{
		Mode: ResourceModeCPU,
		ResourceRequirement: ResourceRequirement{
			RequiredCPU:    "100m",
			LimitedCPU:     "200m",
			RequiredMemory: "128Mi",
			LimitedMemory:  "256Mi",
			RequiredDisk:   "1Gi",
			LimitedDisk:    "2Gi",
		},
	})
	c.APIVersion = APIVersionV2
	c.Spec.SubCharts = []Chart{{Name: "main", Shared: true}}
	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("expected error: apiVersion=v2 must not allow spec.resources even when olaresManifest.version >= 0.12.0")
	}
	want := "spec.resources is not supported for apiVersion=v2"
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error should mention %q, got: %v", want, err)
	}
}

func TestSpec_SupportArch_AcceptsKnownArches(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		c := newValidConfig()
		c.Spec.SupportArch = []string{arch}
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("supportArch=%q must be accepted: %v", arch, err)
		}
	}

	c := newValidConfig()
	c.Spec.SupportArch = []string{"amd64", "arm64"}
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("supportArch=[amd64 arm64] must be accepted: %v", err)
	}
}

func TestSpec_SupportArch_EmptyIsAccepted(t *testing.T) {
	c := newValidConfig()
	c.Spec.SupportArch = nil
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("empty supportArch must remain valid (no enum gate): %v", err)
	}

	c.Spec.SupportArch = []string{}
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("zero-length supportArch slice must remain valid: %v", err)
	}
}

func TestSpec_SupportArch_RejectsUnknownArch(t *testing.T) {
	for _, bad := range []string{"x86", "x86_64", "AMD64", "ARM64", "i386", "riscv64", ""} {
		c := newValidConfig()
		c.Spec.SupportArch = []string{bad}
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatalf("supportArch=%q must be rejected", bad)
		}
		if !strings.Contains(err.Error(), "amd64") || !strings.Contains(err.Error(), "arm64") {
			t.Fatalf("error should mention the enum constraint, got: %v", err)
		}
	}
}

func TestSpec_SupportArch_RejectsDuplicates(t *testing.T) {
	c := newValidConfig()
	c.Spec.SupportArch = []string{"amd64", "amd64"}
	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("expected duplicate supportArch entries to be rejected")
	}
	if !strings.Contains(err.Error(), "duplicate value") {
		t.Fatalf("error should mention duplicate, got: %v", err)
	}
}

func TestDependency_TypeEnum(t *testing.T) {
	c := newValidConfig()
	c.Options.Dependencies = []Dependency{{Name: "foo", Version: "1.0.0", Type: "bogus"}}
	if err := ValidateAppConfiguration(c); err == nil {
		t.Fatal("expected error for bad dependency type")
	}

	c.Options.Dependencies = []Dependency{{Name: "foo", Version: "1.0.0", Type: "system"}}
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("system is a valid dep type: %v", err)
	}

	c.Options.Dependencies = []Dependency{{Name: "foo", Version: "1.0.0", Type: "application"}}
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("application is a valid dep type: %v", err)
	}
}

func TestPolicy_DurationPattern(t *testing.T) {
	c := newValidConfig()
	c.Options.Policies = []Policy{{
		URIRegex: "^/",
		Level:    "one_factor",
		Duration: "not-a-duration",
	}}
	if err := ValidateAppConfiguration(c); err == nil {
		t.Fatal("expected error for bad validDuration")
	}

	c.Options.Policies[0].Duration = "3600s"
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("3600s should be a valid duration: %v", err)
	}
}
