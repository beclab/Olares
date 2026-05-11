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
	if !strings.Contains(err.Error(), "不支持该版本") {
		t.Fatalf("error should mention 不支持该版本, got: %v", err)
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
	if !strings.Contains(errV0.Error(), "不支持该版本") {
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
//     are required (one error per missing field).
//   - requiredGpu, limitedGpu are optional (the manifest still validates
//     when both are absent).
//
// Each subtest blanks out exactly one legacy field on a fully-populated
// fixture and expects the corresponding error message; the tail subtests
// pin the optional GPU contract (omitted is fine, set must parse).
func TestValidateAppSpec_LegacyRequiredFieldsBelowGate(t *testing.T) {
	const legacyVersion = "0.11.0"

	required := []struct {
		name      string
		clear     func(*AppSpec)
		errSuffix string
	}{
		{
			name:      "requiredMemory",
			clear:     func(s *AppSpec) { s.RequiredMemory = "" },
			errSuffix: "spec.requiredMemory is required for olaresManifest.version < 0.12.0",
		},
		{
			name:      "requiredDisk",
			clear:     func(s *AppSpec) { s.RequiredDisk = "" },
			errSuffix: "spec.requiredDisk is required for olaresManifest.version < 0.12.0",
		},
		{
			name:      "requiredCpu",
			clear:     func(s *AppSpec) { s.RequiredCPU = "" },
			errSuffix: "spec.requiredCpu is required for olaresManifest.version < 0.12.0",
		},
		{
			name:      "limitedMemory",
			clear:     func(s *AppSpec) { s.LimitedMemory = "" },
			errSuffix: "spec.limitedMemory is required for olaresManifest.version < 0.12.0",
		},
		{
			name:      "limitedCpu",
			clear:     func(s *AppSpec) { s.LimitedCPU = "" },
			errSuffix: "spec.limitedCpu is required for olaresManifest.version < 0.12.0",
		},
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
			if !strings.Contains(err.Error(), tc.errSuffix) {
				t.Fatalf("error should mention %q, got: %v", tc.errSuffix, err)
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
		c.Spec.Resources = nil

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
//   - spec.resources is required (missing/empty rejected).
//   - Each declared spec.resources[] entry must populate every standard
//     field (cpu/memory/disk pair, plus the gpu pair on gpu-capable modes).
//   - The legacy flat spec.required*/spec.limited* fields must be empty
//     (Rule 7 -- already covered in resources_test.go but pinned here for
//     completeness so a regression on the modern branch lights up locally).
//   - spec.requiredGpu and spec.limitedGpu remain optional at the spec
//     level (gpu fields belong inside spec.resources[] entries instead).
func TestValidateAppSpec_ModernResourcesRequiredAtOrAboveGate(t *testing.T) {
	t.Run("missing_resources_rejected", func(t *testing.T) {
		c := newResourcesConfig() // no modes -> Resources is empty
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: spec.resources is required for olaresManifest.version >= 0.12.0")
		}
		want := "spec.resources is required for olaresManifest.version >= 0.12.0; declare at least one entry"
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error should mention modern consolidated guidance, got: %v", err)
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
