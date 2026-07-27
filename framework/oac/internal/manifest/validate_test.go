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
		Options: Options{
			Dependencies: []Dependency{{
				Name:    olaresSystemDepName,
				Version: olaresDepRulePreV3.requirement,
				Type:    "system",
			}},
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

	// apiVersion=v3 is a 1.12.6-only trigger, so the legacy baseline
	// trips validateModernFieldRequiresManifestVersion. Pair v3 with a
	// modern manifest version + locked Olares dep so the assertion
	// really tracks the apiVersion enum rule rather than the
	// manifest-version gate.
	c = newValidConfig()
	c.ConfigVersion = "0.13.0"
	c.APIVersion = APIVersionV3
	wr := WorkloadReplicas{c.Metadata.Name: 1}
	c.WorkloadReplicas = &wr
	c.Options.Dependencies = []Dependency{newOlaresSystemDep(c)}
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

func TestServicePortsValidation(t *testing.T) {
	validPort := ServicePort{Name: "api", Host: "api", Port: 4712}
	cases := []struct {
		name    string
		ports   []ServicePort
		wantErr string
	}{
		{name: "omitted"},
		{name: "default protocol and automatic expose port", ports: []ServicePort{validPort}},
		{name: "tcp with explicit expose port", ports: []ServicePort{{
			Name: "api", Host: "api", Port: 4712, ExposePort: 4712, Protocol: "tcp",
		}}},
		{name: "udp", ports: []ServicePort{{
			Name: "api", Host: "api", Port: 4712, Protocol: "udp",
		}}},
		{name: "missing name", ports: []ServicePort{{
			Host: "api", Port: 4712,
		}}, wantErr: "port.name is required"},
		{name: "invalid name", ports: []ServicePort{{
			Name: "bad_name", Host: "api", Port: 4712,
		}}, wantErr: "port.name must match"},
		{name: "missing host", ports: []ServicePort{{
			Name: "api", Port: 4712,
		}}, wantErr: "port.host is required"},
		{name: "invalid host", ports: []ServicePort{{
			Name: "api", Host: "API", Port: 4712,
		}}, wantErr: "port.host must match"},
		{name: "missing port", ports: []ServicePort{{
			Name: "api", Host: "api",
		}}, wantErr: "port.port is required"},
		{name: "port above range", ports: []ServicePort{{
			Name: "api", Host: "api", Port: 65536,
		}}, wantErr: "port.port must be between 1 and 65535"},
		{name: "negative expose port", ports: []ServicePort{{
			Name: "api", Host: "api", Port: 4712, ExposePort: -1,
		}}, wantErr: "port.exposePort must be between 1 and 65535"},
		{name: "expose port above range", ports: []ServicePort{{
			Name: "api", Host: "api", Port: 4712, ExposePort: 65536,
		}}, wantErr: "port.exposePort must be between 1 and 65535"},
		{name: "uppercase protocol", ports: []ServicePort{{
			Name: "api", Host: "api", Port: 4712, Protocol: "TCP",
		}}, wantErr: `port.protocol must be one of "", "tcp", "udp"`},
		{name: "duplicate name", ports: []ServicePort{
			validPort,
			{Name: "api", Host: "peer", Port: 4713},
		}, wantErr: `duplicate port name "api"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := newValidConfig()
			c.Ports = tc.ports
			err := ValidateAppConfiguration(c)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error should contain %q, got: %v", tc.wantErr, err)
			}
		})
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
	// newResourcesConfig clears every flat field (and pre-populates
	// workloadReplicas, which the modern gate also requires); we
	// re-populate the five mandatory quantities here.
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
	c.Options.Dependencies = []Dependency{newOlaresSystemDep(c)}
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
	olaresDep := Dependency{
		Name:    olaresSystemDepName,
		Version: olaresDepRulePreV3.requirement,
		Type:    "system",
	}
	c := newValidConfig()
	c.Options.Dependencies = []Dependency{{Name: "foo", Version: "1.0.0", Type: "bogus"}, olaresDep}
	if err := ValidateAppConfiguration(c); err == nil {
		t.Fatal("expected error for bad dependency type")
	}

	c.Options.Dependencies = []Dependency{{Name: "foo", Version: "1.0.0", Type: "system"}, olaresDep}
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("system is a valid dep type: %v", err)
	}

	c.Options.Dependencies = []Dependency{{Name: "foo", Version: "1.0.0", Type: "application"}, olaresDep}
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

// TestPermission_ExternalDataVersionGate documents that
// permission.externalData and permission.appCommon are both gated by
// olaresManifest.version >= 0.12.0:
//
//   - permission.externalData was introduced at 0.12.0; the gate is
//     enforced directly by validatePermission and emits a message naming
//     the field.
//   - permission.appCommon is one of the 1.12.6-only trigger fields, so
//     the gate fires from validateModernFieldRequiresManifestVersion and
//     additionally requires options.dependencies[name=olares].version to
//     be locked to ">=1.12.6-0".
func TestPermission_ExternalDataVersionGate(t *testing.T) {
	// externalData on a legacy (< 0.12.0) manifest is rejected.
	c := newValidConfig() // ConfigVersion "0.11.0"
	c.Permission.ExternalData = true
	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("expected error for permission.externalData on olaresManifest.version < 0.12.0")
	}
	if !strings.Contains(err.Error(), "permission.externalData") {
		t.Fatalf("error should mention permission.externalData, got: %v", err)
	}

	// externalData on a modern (>= 0.12.0) manifest is accepted.
	// workloadReplicas and the Olares system dependency are required at
	// this version (non-v2), so populate minimal ones to keep this test
	// focused on permission.externalData.
	c = newValidConfig()
	c.ConfigVersion = "0.12.0"
	c.Permission.ExternalData = true
	wr := WorkloadReplicas{c.Metadata.Name: 1}
	c.WorkloadReplicas = &wr
	c.Options.Dependencies = []Dependency{newOlaresSystemDep(c)}
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("permission.externalData should be accepted at >= 0.12.0: %v", err)
	}

	// appCommon on a legacy (< 0.12.0) manifest is rejected by
	// validateModernFieldRequiresManifestVersion: appCommon is one of
	// the 1.12.6-only trigger fields and demands both
	// olaresManifest.version >= 0.12.0 and a locked Olares dep.
	c = newValidConfig()
	c.Permission.AppCommon = true
	err = ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("expected error for permission.appCommon on olaresManifest.version < 0.12.0")
	}
	if !strings.Contains(err.Error(), "permission.appCommon") {
		t.Fatalf("error should mention permission.appCommon, got: %v", err)
	}

	// appCommon on a modern manifest validates once the prerequisites
	// (workloadReplicas + locked Olares dep) are in place.
	c = newValidConfig()
	c.ConfigVersion = "0.13.0"
	c.Permission.AppCommon = true
	wr = WorkloadReplicas{c.Metadata.Name: 1}
	c.WorkloadReplicas = &wr
	c.Options.Dependencies = []Dependency{newOlaresSystemDep(c)}
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("permission.appCommon should validate on modern manifest with locked dep: %v", err)
	}
}

func newValidOverlayEntrance() OverlayEntrance {
	return OverlayEntrance{
		Title:       "Jellyfin",
		Port:        8096,
		Workload:    "jellyfin",
		Description: "Access in LAN",
		Protocol:    "tcp",
	}
}

// newOverlayGatewayBaseline returns a modern (>= 0.12.0) baseline that
// declares the prerequisites overlayGateway requires:
//
//   - olaresManifest.version=0.13.0 — overlayGateway is a 1.12.6-only
//     feature field, so the legacy fixture would trip
//     validateModernFieldRequiresManifestVersion before the overlay
//     entrance is even inspected.
//   - workloadReplicas with one entry — required on every non-v2 modern
//     manifest, and the test wouldn't be exercising overlay-specific
//     behaviour if it failed for missing replicas instead.
//   - the olares system dependency locked to the pre-v3 window — v1
//     manifests always sit in the >=1.12.3-0,<1.12.6 window regardless
//     of 1.12.6-only feature fields; newOlaresSystemDep picks that
//     automatically.
//
// Each test then layers exactly one overlay entrance on top so the
// assertion focuses on overlay rules alone.
func newOverlayGatewayBaseline() *AppConfiguration {
	c := newValidConfig()
	c.ConfigVersion = "0.13.0"
	wr := WorkloadReplicas{c.Metadata.Name: 1}
	c.WorkloadReplicas = &wr
	c.Options.Dependencies = []Dependency{newOlaresSystemDep(c)}
	return c
}

func TestOverlayGateway_ValidEntrance(t *testing.T) {
	c := newOverlayGatewayBaseline()
	c.OverlayGateway = OverlayGateway{
		Enable:    true,
		Entrances: []OverlayEntrance{newValidOverlayEntrance()},
	}
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("valid overlayGateway entrance must pass: %v", err)
	}
}

func TestOverlayGateway_ProtocolEnum(t *testing.T) {
	cases := []struct {
		value   string
		wantErr bool
	}{
		{"", false}, // empty means both tcp and udp
		{"tcp", false},
		{"udp", false},
		{"sctp", true},
		{"TCP", true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.value, func(t *testing.T) {
			c := newOverlayGatewayBaseline()
			e := newValidOverlayEntrance()
			e.Protocol = tc.value
			c.OverlayGateway = OverlayGateway{Enable: true, Entrances: []OverlayEntrance{e}}
			err := ValidateAppConfiguration(c)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error for protocol %q", tc.value)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error for protocol %q: %v", tc.value, err)
			}
		})
	}
}

func TestOverlayGateway_RequiredAndPortRules(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*OverlayEntrance)
		field  string
	}{
		{"missing title", func(e *OverlayEntrance) { e.Title = "" }, "title"},
		{"missing workload", func(e *OverlayEntrance) { e.Workload = "" }, "workload"},
		{"zero port", func(e *OverlayEntrance) { e.Port = 0 }, "port"},
		{"negative port", func(e *OverlayEntrance) { e.Port = -1 }, "port"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := newValidConfig()
			e := newValidOverlayEntrance()
			tc.mutate(&e)
			c.OverlayGateway = OverlayGateway{Enable: true, Entrances: []OverlayEntrance{e}}
			err := ValidateAppConfiguration(c)
			if err == nil {
				t.Fatalf("expected error for %s", tc.name)
			}
			if !strings.Contains(err.Error(), tc.field) {
				t.Fatalf("error should mention %q, got: %v", tc.field, err)
			}
		})
	}
}

// TestOptions_TemplateOnlyRequiresAllowMultipleInstall pins down the
// templateOnly => allowMultipleInstall=true cross-field rule on Options:
// templateOnly apps install as multiple clones, so allowMultipleInstall is
// mandatory. The default zero value (false) must surface the error, the
// explicit-false case must too, and pairing templateOnly with
// allowMultipleInstall=true must validate cleanly.
func TestOptions_TemplateOnlyRequiresAllowMultipleInstall(t *testing.T) {
	c := newValidConfig()
	c.Options.TemplateOnly = true
	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("expected error: options.templateOnly=true requires allowMultipleInstall=true")
	}
	if !strings.Contains(err.Error(), "options.allowMultipleInstall must be true when options.templateOnly is true") {
		t.Fatalf("error should flag the templateOnly cross-field rule, got: %v", err)
	}

	c = newValidConfig()
	c.Options.TemplateOnly = true
	c.Options.AllowMultipleInstall = false
	if err := ValidateAppConfiguration(c); err == nil {
		t.Fatal("expected error: explicit allowMultipleInstall=false still violates the rule")
	}

	// templateOnly is a 1.12.6-only trigger, so the happy-path case
	// requires a modern manifest with the prerequisites
	// (workloadReplicas + locked Olares dep). Without them the new
	// validateModernFieldRequiresManifestVersion gate would mask the
	// templateOnly+allowMultipleInstall pairing under test here.
	c = newValidConfig()
	c.ConfigVersion = "0.13.0"
	c.Options.TemplateOnly = true
	c.Options.AllowMultipleInstall = true
	wr := WorkloadReplicas{c.Metadata.Name: 1}
	c.WorkloadReplicas = &wr
	c.Options.Dependencies = []Dependency{newOlaresSystemDep(c)}
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("templateOnly=true + allowMultipleInstall=true must pass: %v", err)
	}

	// allowMultipleInstall=true on its own is fine; the rule only fires
	// when templateOnly is true. Keeps backward compatibility for charts
	// that already opt into multi-install without being template-only.
	c = newValidConfig()
	c.Options.AllowMultipleInstall = true
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("allowMultipleInstall=true alone must remain valid: %v", err)
	}
}

// TestAppSpec_TemplateOnlyAllowsAutoOnNonDiskLegacy verifies that a
// template-only manifest may use AutoResourceValue ("-1") on the legacy
// flat cpu/memory fields. "Legacy" in the name refers to the legacy
// resource ENVELOPE (flat spec.required*/limited* fields) rather than
// the manifest version itself: templateOnly is a 1.12.6-only trigger,
// so the manifest must still declare olaresManifest.version >= 0.12.0
// and lock the Olares dep. The legacy flat envelope remains a supported
// shape at modern versions when spec.resources[] is not used.
func TestAppSpec_TemplateOnlyAllowsAutoOnNonDiskLegacy(t *testing.T) {
	c := newValidConfig()
	c.ConfigVersion = "0.13.0"
	c.Options.TemplateOnly = true
	c.Options.AllowMultipleInstall = true
	c.Spec.RequiredCPU = AutoResourceValue
	c.Spec.LimitedCPU = AutoResourceValue
	c.Spec.RequiredMemory = AutoResourceValue
	c.Spec.LimitedMemory = AutoResourceValue
	wr := WorkloadReplicas{c.Metadata.Name: 1}
	c.WorkloadReplicas = &wr
	c.Options.Dependencies = []Dependency{newOlaresSystemDep(c)}
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("template-only legacy envelope with -1 on cpu/memory must pass: %v", err)
	}
}

func TestAppSpec_TemplateOnlyRejectsAutoOnDiskLegacy(t *testing.T) {
	c := newValidConfig()
	c.Options.TemplateOnly = true
	c.Options.AllowMultipleInstall = true
	c.Spec.RequiredDisk = AutoResourceValue
	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("expected error: template-only apps cannot use -1 on requiredDisk")
	}
	if !strings.Contains(err.Error(), "requiredDisk") {
		t.Fatalf("error should mention requiredDisk, got: %v", err)
	}
}

func TestAppSpec_NonTemplateRejectsAutoOnLegacyFlat(t *testing.T) {
	c := newValidConfig()
	c.Spec.RequiredCPU = AutoResourceValue
	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("expected error: non-template legacy flat fields cannot use -1")
	}
	if !strings.Contains(err.Error(), "requiredCpu") {
		t.Fatalf("error should mention requiredCpu, got: %v", err)
	}
}

// TestRootProvider_ForbiddenAtOrAbove012 documents the modern gate on the
// top-level provider section (AppConfiguration.Provider, distinct from
// permission.provider): starting with olaresManifest.version 0.12.0 the
// section must be empty, regardless of apiVersion. Below 0.12.0 the
// legacy behaviour (arbitrary provider entries accepted) is preserved.
func TestRootProvider_ForbiddenAtOrAbove012(t *testing.T) {
	const wantMsg = "provider must be empty for olaresManifest.version >="

	providerEntry := Provider{Name: "ollamaclient", Entrance: "main", Paths: []string{"/api"}, Verbs: []string{"GET"}}

	t.Run("v1_modern_with_root_provider_rejected", func(t *testing.T) {
		c := newValidConfig()
		c.ConfigVersion = "0.12.0"
		c.Provider = []Provider{providerEntry}
		wr := WorkloadReplicas{c.Metadata.Name: 1}
		c.WorkloadReplicas = &wr
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: root provider must be empty at >= 0.12.0")
		}
		if !strings.Contains(err.Error(), wantMsg) {
			t.Fatalf("error should mention the root provider gate, got: %v", err)
		}
	})

	t.Run("v3_modern_with_root_provider_rejected", func(t *testing.T) {
		c := newValidConfig()
		c.ConfigVersion = "0.13.0"
		c.APIVersion = APIVersionV3
		c.Provider = []Provider{providerEntry}
		wr := WorkloadReplicas{c.Metadata.Name: 1}
		c.WorkloadReplicas = &wr
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: root provider must be empty at >= 0.12.0 (v3)")
		}
		if !strings.Contains(err.Error(), wantMsg) {
			t.Fatalf("error should mention the root provider gate, got: %v", err)
		}
	})

	t.Run("v2_modern_with_root_provider_rejected", func(t *testing.T) {
		c := newValidConfig()
		c.ConfigVersion = "0.13.0"
		c.APIVersion = APIVersionV2
		c.Spec.SubCharts = []Chart{{Name: "main", Shared: true}}
		c.Provider = []Provider{providerEntry}
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: root provider must be empty at >= 0.12.0 (v2)")
		}
		if !strings.Contains(err.Error(), wantMsg) {
			t.Fatalf("error should mention the root provider gate, got: %v", err)
		}
	})

	t.Run("modern_without_root_provider_accepted", func(t *testing.T) {
		c := newValidConfig()
		c.ConfigVersion = "0.13.0"
		wr := WorkloadReplicas{c.Metadata.Name: 1}
		c.WorkloadReplicas = &wr
		c.Options.Dependencies = []Dependency{newOlaresSystemDep(c)}
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("modern manifest without root provider must validate: %v", err)
		}
	})

	t.Run("legacy_with_root_provider_accepted", func(t *testing.T) {
		c := newValidConfig() // ConfigVersion 0.11.0
		c.Provider = []Provider{providerEntry}
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("legacy manifest with root provider must validate: %v", err)
		}
	})

	t.Run("modern_with_both_provider_sections_aggregated", func(t *testing.T) {
		// Both gates (permission.provider AND root provider) must surface
		// in a single Validate run so a manifest carrying both retired
		// shapes sees every offender at once.
		c := newValidConfig()
		c.ConfigVersion = "0.13.0"
		wr := WorkloadReplicas{c.Metadata.Name: 1}
		c.WorkloadReplicas = &wr
		c.Options.Dependencies = []Dependency{newOlaresSystemDep(c)}
		c.Provider = []Provider{providerEntry}
		c.Permission.Provider = []ProviderPermission{{AppName: "ollama", ProviderName: "ollamaclient"}}
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected aggregated errors for root provider + permission.provider")
		}
		msg := err.Error()
		if !strings.Contains(msg, "permission.provider must be empty") {
			t.Fatalf("error should mention permission.provider gate, got: %v", err)
		}
		if !strings.Contains(msg, wantMsg) {
			t.Fatalf("error should mention root provider gate, got: %v", err)
		}
	})
}

// TestPermission_ProviderForbiddenAtOrAbove012 documents the modern gate
// on permission.provider: starting with olaresManifest.version 0.12.0 the
// field must be empty, regardless of apiVersion. Below 0.12.0 the legacy
// behaviour (any number of provider entries accepted) is preserved.
func TestPermission_ProviderForbiddenAtOrAbove012(t *testing.T) {
	const wantMsg = "permission.provider must be empty for olaresManifest.version >="

	providerEntry := ProviderPermission{AppName: "ollama", ProviderName: "ollamaclient"}

	t.Run("v1_modern_with_provider_rejected", func(t *testing.T) {
		c := newValidConfig()
		c.ConfigVersion = "0.12.0"
		c.Permission.Provider = []ProviderPermission{providerEntry}
		wr := WorkloadReplicas{c.Metadata.Name: 1}
		c.WorkloadReplicas = &wr
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: permission.provider must be empty at >= 0.12.0")
		}
		if !strings.Contains(err.Error(), wantMsg) {
			t.Fatalf("error should mention the provider gate, got: %v", err)
		}
	})

	t.Run("v3_modern_with_provider_rejected", func(t *testing.T) {
		c := newValidConfig()
		c.ConfigVersion = "0.13.0"
		c.APIVersion = APIVersionV3
		c.Permission.Provider = []ProviderPermission{providerEntry}
		wr := WorkloadReplicas{c.Metadata.Name: 1}
		c.WorkloadReplicas = &wr
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: permission.provider must be empty at >= 0.12.0 (v3)")
		}
		if !strings.Contains(err.Error(), wantMsg) {
			t.Fatalf("error should mention the provider gate, got: %v", err)
		}
	})

	t.Run("v2_modern_with_provider_rejected", func(t *testing.T) {
		// The rule does not depend on apiVersion — the field is retired
		// platform-wide on the modern channel. v2 must hit the same gate.
		c := newValidConfig()
		c.ConfigVersion = "0.13.0"
		c.APIVersion = APIVersionV2
		c.Spec.SubCharts = []Chart{{Name: "main", Shared: true}}
		c.Permission.Provider = []ProviderPermission{providerEntry}
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: permission.provider must be empty at >= 0.12.0 (v2)")
		}
		if !strings.Contains(err.Error(), wantMsg) {
			t.Fatalf("error should mention the provider gate, got: %v", err)
		}
	})

	t.Run("modern_without_provider_accepted", func(t *testing.T) {
		c := newValidConfig()
		c.ConfigVersion = "0.13.0"
		wr := WorkloadReplicas{c.Metadata.Name: 1}
		c.WorkloadReplicas = &wr
		c.Options.Dependencies = []Dependency{newOlaresSystemDep(c)}
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("modern manifest without permission.provider must validate: %v", err)
		}
	})

	t.Run("legacy_with_provider_accepted", func(t *testing.T) {
		c := newValidConfig() // ConfigVersion 0.11.0
		c.Permission.Provider = []ProviderPermission{providerEntry}
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("legacy manifest with permission.provider must validate: %v", err)
		}
	})
}

// TestWorkloadReplicas_RequiredAtOrAbove012 documents the new gate: a
// modern (olaresManifest.version >= 0.12.0) manifest must declare a
// non-empty workloadReplicas map, except for apiVersion=v2, which carries
// workloads inside subCharts and therefore has no parent-level field to
// require. Legacy versions (< 0.12.0) remain unconstrained.
func TestWorkloadReplicas_RequiredAtOrAbove012(t *testing.T) {
	const wantMsg = "workloadReplicas is required for olaresManifest.version >="

	t.Run("v1_at_boundary_missing_rejected", func(t *testing.T) {
		c := newValidConfig()
		c.ConfigVersion = "0.12.0"
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: workloadReplicas missing at olaresManifest.version=0.12.0 (v1)")
		}
		if !strings.Contains(err.Error(), wantMsg) {
			t.Fatalf("error should mention the workloadReplicas gate, got: %v", err)
		}
	})

	t.Run("v1_above_boundary_empty_map_rejected", func(t *testing.T) {
		c := newValidConfig()
		c.ConfigVersion = "0.13.0"
		empty := WorkloadReplicas{}
		c.WorkloadReplicas = &empty
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: empty workloadReplicas map must not satisfy the gate")
		}
		if !strings.Contains(err.Error(), wantMsg) {
			t.Fatalf("error should mention the workloadReplicas gate, got: %v", err)
		}
	})

	t.Run("v1_above_boundary_with_entry_accepted", func(t *testing.T) {
		c := newValidConfig()
		c.ConfigVersion = "0.13.0"
		wr := WorkloadReplicas{c.Metadata.Name: 1}
		c.WorkloadReplicas = &wr
		c.Options.Dependencies = []Dependency{newOlaresSystemDep(c)}
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("modern v1 with workloadReplicas must validate: %v", err)
		}
	})

	t.Run("v3_at_boundary_missing_rejected", func(t *testing.T) {
		c := newValidConfig()
		c.ConfigVersion = "0.12.0"
		c.APIVersion = APIVersionV3
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: workloadReplicas missing at olaresManifest.version=0.12.0 (v3)")
		}
		if !strings.Contains(err.Error(), wantMsg) {
			t.Fatalf("error should mention the workloadReplicas gate, got: %v", err)
		}
	})

	t.Run("v3_above_boundary_with_entry_accepted", func(t *testing.T) {
		c := newValidConfig()
		c.ConfigVersion = "0.13.0"
		c.APIVersion = APIVersionV3
		wr := WorkloadReplicas{c.Metadata.Name: 1}
		c.WorkloadReplicas = &wr
		c.Options.Dependencies = []Dependency{newOlaresSystemDep(c)}
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("modern v3 with workloadReplicas must validate: %v", err)
		}
	})

	t.Run("empty_apiversion_defaults_to_v1_and_requires", func(t *testing.T) {
		c := newValidConfig()
		c.ConfigVersion = "0.13.0"
		c.APIVersion = ""
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: empty apiVersion is treated as v1 and must require workloadReplicas at >= 0.12.0")
		}
		if !strings.Contains(err.Error(), wantMsg) {
			t.Fatalf("error should mention the workloadReplicas gate, got: %v", err)
		}
	})

	t.Run("v2_modern_does_not_require", func(t *testing.T) {
		c := newValidConfig()
		c.ConfigVersion = "0.13.0"
		c.APIVersion = APIVersionV2
		c.Spec.SubCharts = []Chart{{Name: "main", Shared: true}}
		c.Options.Dependencies = []Dependency{newOlaresSystemDep(c)}
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("modern v2 must not require workloadReplicas (workloads live in subCharts): %v", err)
		}
	})

	legacyVersions := []string{"0.11.9", "0.11.0", "0.10.0"}
	for _, v := range legacyVersions {
		v := v
		t.Run("legacy_"+v+"_unconstrained", func(t *testing.T) {
			c := newValidConfig()
			c.ConfigVersion = v
			if err := ValidateAppConfiguration(c); err != nil {
				t.Fatalf("legacy version %s must not require workloadReplicas: %v", v, err)
			}
		})
	}
}

// TestOlaresDependency_ConstraintGate documents the rules on the Olares
// system dependency (options.dependencies entry with name="olares" and
// type="system"):
//
//   - The entry must exist on every manifest.
//   - apiVersion=v3 forces the constraint to restrict Olares to
//     >=1.12.6-0.
//   - apiVersion in {empty, v1, v2}: the constraint must restrict Olares
//     to >=1.12.3-0,<1.12.6 regardless of 1.12.6-only feature fields
//     (no leakage into 1.12.6+, no permission below the 1.12.3-0 floor).
//
// Each subtest builds a config that already satisfies every other modern
// gate so the assertion truly tracks validateOlaresDependency alone.
func TestOlaresDependency_ConstraintGate(t *testing.T) {
	// modernV1 builds a non-v2 modern manifest. workloadReplicas is
	// required for non-v2 modern manifests; apiVersion=v1 still selects
	// the pre-v3 (>=1.12.3-0,<1.12.6) window.
	modernV1 := func() *AppConfiguration {
		c := newValidConfig()
		c.ConfigVersion = "0.13.0"
		wr := WorkloadReplicas{c.Metadata.Name: 1}
		c.WorkloadReplicas = &wr
		return c
	}
	modernV3 := func() *AppConfiguration {
		c := modernV1()
		c.APIVersion = APIVersionV3
		return c
	}
	// modernV2NoTriggers is the only reachable shape that still exercises
	// the pre-v3 (>=1.12.3-0,<1.12.6) window: v2 manifests do not require
	// workloadReplicas, and the helper deliberately leaves every other
	// 1.12.6-only feature field unset.
	modernV2NoTriggers := func() *AppConfiguration {
		c := newValidConfig()
		c.ConfigVersion = "0.13.0"
		c.APIVersion = APIVersionV2
		c.Spec.SubCharts = []Chart{{Name: "main", Shared: true}}
		return c
	}

	setOlaresDep := func(c *AppConfiguration, version string) {
		c.Options.Dependencies = []Dependency{{
			Name:    olaresSystemDepName,
			Version: version,
			Type:    "system",
		}}
	}

	t.Run("modern_without_olares_dep_rejected", func(t *testing.T) {
		c := modernV1()
		c.Options.Dependencies = nil
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: modern manifest must declare the Olares system dependency")
		}
		if !strings.Contains(err.Error(), `options.dependencies must declare an entry with name="olares" and type="system"`) {
			t.Fatalf("error should mention the missing Olares dep, got: %v", err)
		}
	})

	t.Run("legacy_without_olares_dep_rejected", func(t *testing.T) {
		c := newValidConfig()
		c.Options.Dependencies = nil
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: legacy manifest must declare the Olares system dependency")
		}
		if !strings.Contains(err.Error(), `options.dependencies must declare an entry with name="olares" and type="system"`) {
			t.Fatalf("error should mention the missing Olares dep, got: %v", err)
		}
	})

	t.Run("v2_no_triggers_accepts_pre_v3_range", func(t *testing.T) {
		c := modernV2NoTriggers()
		setOlaresDep(c, ">=1.12.3-0,<1.12.6")
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("v2 manifest without 1.12.6-only fields must accept the pre-v3 window: %v", err)
		}
	})

	t.Run("v2_no_triggers_rejects_constraint_that_leaks_into_v3_range", func(t *testing.T) {
		c := modernV2NoTriggers()
		// >=1.12.0 allows 1.12.6 and 2.0.0 — both outside the pre-v3 window.
		setOlaresDep(c, ">=1.12.0")
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: pre-v3 dep constraint must not allow 1.12.6 / 2.0.0")
		}
		if !strings.Contains(err.Error(), `must restrict the Olares system version to ">=1.12.3-0,<1.12.6" for apiVersion=v2`) {
			t.Fatalf("error should mention the pre-v3 constraint requirement on v2, got: %v", err)
		}
	})

	t.Run("v2_no_triggers_rejects_constraint_below_floor", func(t *testing.T) {
		c := modernV2NoTriggers()
		// <1.12.6 allows everything below 1.12.6, including 1.12.2 (below the 1.12.3-0 floor).
		setOlaresDep(c, "<1.12.6")
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: pre-v3 dep constraint must enforce the >=1.12.3-0 floor")
		}
		if !strings.Contains(err.Error(), "1.12.2") {
			t.Fatalf("error should call out the below-floor sample 1.12.2, got: %v", err)
		}
	})

	t.Run("v2_rejects_v3_only_range_without_triggers", func(t *testing.T) {
		// Without a trigger, v2 must stay in the pre-v3 window — the v3
		// range allows 1.12.6, which is forbidden there.
		c := modernV2NoTriggers()
		setOlaresDep(c, ">=1.12.6-0")
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: v2 dep constraint must not target the v3-only Olares range")
		}
		if !strings.Contains(err.Error(), `apiVersion=v2`) {
			t.Fatalf("error should mention apiVersion=v2, got: %v", err)
		}
	})

	t.Run("v1_with_workload_replicas_accepts_pre_v3_range", func(t *testing.T) {
		c := modernV1()
		setOlaresDep(c, ">=1.12.3-0,<1.12.6")
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("v1 with workloadReplicas must accept the pre-v3 dep window: %v", err)
		}
	})

	t.Run("v1_with_workload_replicas_rejects_post_v3_range", func(t *testing.T) {
		c := modernV1()
		setOlaresDep(c, ">=1.12.6-0")
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: v1 dep constraint must stay in the pre-v3 window")
		}
		if !strings.Contains(err.Error(), `must restrict the Olares system version to ">=1.12.3-0,<1.12.6" for apiVersion=v1`) {
			t.Fatalf("error should mention the pre-v3 requirement on v1, got: %v", err)
		}
	})

	t.Run("empty_apiversion_treated_as_v1_pre_v3_range", func(t *testing.T) {
		c := modernV1()
		c.APIVersion = ""
		setOlaresDep(c, ">=1.12.3-0,<1.12.6")
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("empty apiVersion with workloadReplicas must accept the pre-v3 window: %v", err)
		}
	})

	t.Run("v3_accepts_in_range_constraint", func(t *testing.T) {
		c := modernV3()
		setOlaresDep(c, ">=1.12.6-0")
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("v3 manifest with >=1.12.6-0 must validate: %v", err)
		}
	})

	t.Run("v3_rejects_pre_1_12_6_constraint", func(t *testing.T) {
		c := modernV3()
		// Allows 1.12.5 (below the 1.12.6-0 floor for v3).
		setOlaresDep(c, ">=1.12.5,<2.0.0")
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: v3 dep constraint must restrict Olares to >=1.12.6-0")
		}
		if !strings.Contains(err.Error(), `must restrict the Olares system version to ">=1.12.6-0" for apiVersion=v3`) {
			t.Fatalf("error should mention the v3 constraint requirement, got: %v", err)
		}
	})

	t.Run("v3_message_omits_feature_promotion_clause", func(t *testing.T) {
		// On v3, the requirement comes from apiVersion itself; the
		// "because the manifest declares ..." clause is only meaningful
		// when rule 4 promoted a v1/v2 manifest, so it should be absent
		// here to avoid misleading users about the trigger.
		c := modernV3()
		setOlaresDep(c, ">=1.12.5,<2.0.0")
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: v3 dep constraint must restrict Olares to >=1.12.6-0")
		}
		if strings.Contains(err.Error(), "because the manifest declares") {
			t.Fatalf("v3 error must not attribute the requirement to feature triggers, got: %v", err)
		}
	})

	t.Run("malformed_constraint_rejected", func(t *testing.T) {
		c := modernV1()
		setOlaresDep(c, "not-a-semver-constraint")
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: malformed semver constraint must be rejected")
		}
		if !strings.Contains(err.Error(), "is not a valid semver constraint") {
			t.Fatalf("error should mention invalid semver constraint, got: %v", err)
		}
	})

	t.Run("wrong_type_does_not_count_as_olares_system_dep", func(t *testing.T) {
		c := modernV1()
		// Same name but type=application — does not satisfy the gate.
		c.Options.Dependencies = []Dependency{{
			Name:    olaresSystemDepName,
			Version: ">=1.12.6-0",
			Type:    "application",
		}}
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: only type=system Olares dependency satisfies the gate")
		}
		if !strings.Contains(err.Error(), `options.dependencies must declare an entry with name="olares" and type="system"`) {
			t.Fatalf("error should mention the missing system Olares dep, got: %v", err)
		}
	})
}

// TestOlaresDependency_FeatureTriggersDoNotPromoteV1V2Range verifies that
// 1.12.6-only feature fields on v1/v2 manifests do NOT promote the Olares
// dep constraint window — apiVersion alone selects pre-v3 vs post-v3.
func TestOlaresDependency_FeatureTriggersDoNotPromoteV1V2Range(t *testing.T) {
	const preV3Constraint = ">=1.12.3-0,<1.12.6"

	base := func() *AppConfiguration {
		c := newValidConfig()
		c.ConfigVersion = "0.13.0"
		c.APIVersion = APIVersionV2
		c.Spec.SubCharts = []Chart{{Name: "main", Shared: true}}
		c.Options.Dependencies = []Dependency{{
			Name:    olaresSystemDepName,
			Version: preV3Constraint,
			Type:    "system",
		}}
		return c
	}

	cases := []struct {
		name      string
		flip      func(c *AppConfiguration)
		extraSetup func(c *AppConfiguration)
	}{
		{
			name: "workloadReplicas",
			flip: func(c *AppConfiguration) {
				wr := WorkloadReplicas{c.Metadata.Name: 1}
				c.WorkloadReplicas = &wr
			},
		},
		{
			name: "overlayGateway",
			flip: func(c *AppConfiguration) {
				c.OverlayGateway = OverlayGateway{
					Enable:    true,
					Entrances: []OverlayEntrance{newValidOverlayEntrance()},
				}
			},
		},
		{
			name: "options.LLMGatewaySupported",
			flip: func(c *AppConfiguration) {
				c.Options.LLMGatewaySupported = true
			},
		},
		{
			name: "permission.appCommon",
			flip: func(c *AppConfiguration) {
				c.Permission.AppCommon = true
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name+"_accepts_pre_v3_with_feature", func(t *testing.T) {
			c := base()
			if tc.extraSetup != nil {
				tc.extraSetup(c)
			}
			tc.flip(c)
			if err := ValidateAppConfiguration(c); err != nil {
				t.Fatalf("declaring %s on v2 with pre-v3 dep must validate: %v", tc.name, err)
			}
		})

		t.Run(tc.name+"_rejects_post_v3_on_v2", func(t *testing.T) {
			c := base()
			if tc.extraSetup != nil {
				tc.extraSetup(c)
			}
			tc.flip(c)
			c.Options.Dependencies = []Dependency{{
				Name:    olaresSystemDepName,
				Version: ">=1.12.6-0",
				Type:    "system",
			}}
			err := ValidateAppConfiguration(c)
			if err == nil {
				t.Fatalf("expected error: v2 dep constraint must stay in the pre-v3 window even with %s", tc.name)
			}
			msg := err.Error()
			if !strings.Contains(msg, `must restrict the Olares system version to ">=1.12.3-0,<1.12.6" for apiVersion=v2`) {
				t.Fatalf("error should require the pre-v3 range, got: %v", err)
			}
			if strings.Contains(msg, "because the manifest declares") {
				t.Fatalf("v2 error must not attribute the requirement to feature triggers, got: %v", err)
			}
		})
	}

	t.Run("spec.accelerator_accepts_pre_v3_on_v1", func(t *testing.T) {
		c := newValidConfig()
		c.ConfigVersion = "0.13.0"
		c.APIVersion = APIVersionV1
		c.Spec.RequiredCPU = ""
		c.Spec.LimitedCPU = ""
		c.Spec.RequiredMemory = ""
		c.Spec.LimitedMemory = ""
		c.Spec.RequiredDisk = ""
		c.Spec.LimitedDisk = ""
		c.Spec.RequiredGPU = ""
		c.Spec.LimitedGPU = ""
		c.Spec.Accelerator = []ResourceMode{{
			Mode: ResourceModeCPU,
			ResourceRequirement: ResourceRequirement{
				RequiredCPU:    "100m",
				LimitedCPU:     "200m",
				RequiredMemory: "128Mi",
				LimitedMemory:  "256Mi",
				RequiredDisk:   "10Mi",
				LimitedDisk:    "20Mi",
			},
		}}
		wr := WorkloadReplicas{c.Metadata.Name: 1}
		c.WorkloadReplicas = &wr
		c.Options.Dependencies = []Dependency{{
			Name:    olaresSystemDepName,
			Version: preV3Constraint,
			Type:    "system",
		}}
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("v1 with spec.accelerator and pre-v3 dep must validate: %v", err)
		}
	})

	t.Run("options.shared_is_v3_only_post_v3_range", func(t *testing.T) {
		c := newValidConfig()
		c.ConfigVersion = "0.13.0"
		c.APIVersion = APIVersionV3
		c.Options.Shared = true
		wr := WorkloadReplicas{c.Metadata.Name: 1}
		c.WorkloadReplicas = &wr
		c.Options.Dependencies = []Dependency{{
			Name:    olaresSystemDepName,
			Version: preV3Constraint,
			Type:    "system",
		}}
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: v3 dep constraint must restrict Olares to >=1.12.6-0")
		}
		if !strings.Contains(err.Error(), `must restrict the Olares system version to ">=1.12.6-0" for apiVersion=v3`) {
			t.Fatalf("error should require the post-v3 range for v3, got: %v", err)
		}
	})

	t.Run("options.templateOnly_accepts_pre_v3_on_v2", func(t *testing.T) {
		c := base()
		c.Options.AllowMultipleInstall = true
		c.Options.TemplateOnly = true
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("v2 with options.templateOnly and pre-v3 dep must validate: %v", err)
		}
	})
}

// TestModernFieldRequiresManifestVersion is the inverse of
// TestOlaresDependency_FeatureTriggersDoNotPromoteV1V2Range: on a LEGACY manifest
// (olaresManifest.version < 0.12.0), declaring any 1.12.6-only feature
// field — or apiVersion=v3 — must be rejected with guidance to bump the
// manifest version AND lock the Olares dep to >=1.12.6-0.
//
// Each subtest starts from the legacy baseline (newValidConfig pins
// ConfigVersion=0.11.0), flips exactly one trigger, and asserts:
//
//  1. validation fails;
//  2. the error names the field that triggered the gate;
//  3. the error spells out the >=1.12.6-0 dep requirement so the user
//     doesn't have to discover it on a second Lint pass after bumping
//     olaresManifest.version.
//
// The "no triggers stay legacy" control case anchors the gate so future
// edits don't accidentally fire it on plain legacy manifests.
func TestModernFieldRequiresManifestVersion(t *testing.T) {
	cases := []struct {
		name       string
		flip       func(c *AppConfiguration)
		wantLabel  string
		extraSetup func(c *AppConfiguration)
	}{
		{
			name: "apiVersion=v3",
			flip: func(c *AppConfiguration) {
				c.APIVersion = APIVersionV3
			},
			wantLabel: "apiVersion=v3",
		},
		{
			name: "spec.accelerator",
			flip: func(c *AppConfiguration) {
				c.Spec.Accelerator = []ResourceMode{{
					Mode: ResourceModeCPU,
					ResourceRequirement: ResourceRequirement{
						RequiredCPU:    "100m",
						LimitedCPU:     "200m",
						RequiredMemory: "128Mi",
						LimitedMemory:  "256Mi",
						RequiredDisk:   "10Mi",
						LimitedDisk:    "20Mi",
					},
				}}
			},
			wantLabel: "spec.accelerator",
		},
		{
			name: "workloadReplicas",
			flip: func(c *AppConfiguration) {
				wr := WorkloadReplicas{c.Metadata.Name: 1}
				c.WorkloadReplicas = &wr
			},
			wantLabel: "workloadReplicas",
		},
		{
			name: "overlayGateway",
			flip: func(c *AppConfiguration) {
				c.OverlayGateway = OverlayGateway{
					Enable:    true,
					Entrances: []OverlayEntrance{newValidOverlayEntrance()},
				}
			},
			wantLabel: "overlayGateway",
		},
		{
			name: "options.LLMGatewaySupported",
			flip: func(c *AppConfiguration) {
				c.Options.LLMGatewaySupported = true
			},
			wantLabel: "options.LLMGatewaySupported",
		},
		{
			name: "options.templateOnly",
			flip: func(c *AppConfiguration) {
				// templateOnly cross-field rule also requires
				// allowMultipleInstall=true; set it so this assertion
				// stays focused on the manifest-version gate.
				c.Options.AllowMultipleInstall = true
				c.Options.TemplateOnly = true
			},
			wantLabel: "options.templateOnly",
		},
		{
			name: "options.shared",
			flip: func(c *AppConfiguration) {
				// shared has additional cross-field rules; the
				// manifest-version gate must fire regardless of those.
				c.Options.Shared = true
			},
			wantLabel: "options.shared",
		},
		{
			name: "permission.appCommon",
			flip: func(c *AppConfiguration) {
				c.Permission.AppCommon = true
			},
			wantLabel: "permission.appCommon",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name+"_rejected_on_legacy_manifest", func(t *testing.T) {
			c := newValidConfig() // ConfigVersion=0.11.0
			if tc.extraSetup != nil {
				tc.extraSetup(c)
			}
			tc.flip(c)
			err := ValidateAppConfiguration(c)
			if err == nil {
				t.Fatalf("expected error: declaring %s on olaresManifest.version=0.11.0 must be rejected", tc.wantLabel)
			}
			msg := err.Error()
			if !strings.Contains(msg, tc.wantLabel) {
				t.Fatalf("error should name the trigger %s, got: %v", tc.wantLabel, err)
			}
			if !strings.Contains(msg, "olaresManifest.version must be >= 0.12.0") {
				t.Fatalf("error should require bumping olaresManifest.version, got: %v", err)
			}
			if !strings.Contains(msg, `">=1.12.6-0"`) {
				t.Fatalf("error should mention the Olares dep lock, got: %v", err)
			}
		})
	}

	t.Run("legacy_without_triggers_accepted", func(t *testing.T) {
		// Anchor case: a plain legacy manifest must keep validating.
		// Catches regressions where the new gate accidentally fires on
		// every legacy manifest.
		c := newValidConfig()
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("legacy manifest without triggers must validate: %v", err)
		}
	})

	t.Run("modern_with_triggers_does_not_double_fire", func(t *testing.T) {
		// At olaresManifest.version>=0.12.0 the legacy gate must stay
		// silent — the dep window is owned by validateOlaresDependency.
		// permission.appCommon would trigger the legacy gate at 0.11.0
		// but here we expect the error (if any) to come from the dep
		// check, not from the legacy gate.
		c := newValidConfig()
		c.ConfigVersion = "0.13.0"
		c.Permission.AppCommon = true
		wr := WorkloadReplicas{c.Metadata.Name: 1}
		c.WorkloadReplicas = &wr
		c.Options.Dependencies = []Dependency{newOlaresSystemDep(c)}
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("modern manifest with permission.appCommon and locked dep must validate: %v", err)
		}
	})

	t.Run("multiple_triggers_listed_together", func(t *testing.T) {
		// When several triggers fire on one legacy manifest the gate
		// should list every offender in a single message so the author
		// fixes the version once instead of chasing them piecemeal.
		c := newValidConfig()
		c.APIVersion = APIVersionV3
		c.Permission.AppCommon = true
		wr := WorkloadReplicas{c.Metadata.Name: 1}
		c.WorkloadReplicas = &wr
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: legacy manifest with multiple triggers")
		}
		msg := err.Error()
		for _, want := range []string{"apiVersion=v3", "workloadReplicas", "permission.appCommon"} {
			if !strings.Contains(msg, want) {
				t.Fatalf("error should list trigger %q, got: %v", want, err)
			}
		}
	})
}

// TestOptions_SharedRequiresAPIVersionV3 pins down the shared => v3
// cross-field rule on Options: shared installs only make sense on the v3
// schema (a single install services multiple users). v1/v2/empty must
// reject; v3 must accept. The control case (shared=false on v1) confirms
// the rule does not fire for the default shape.
func TestOptions_SharedRequiresAPIVersionV3(t *testing.T) {
	cases := []struct {
		name       string
		apiVersion string
		shared     bool
		wantErr    bool
	}{
		{"v1 + shared=true rejected", APIVersionV1, true, true},
		{"v2 + shared=true rejected", APIVersionV2, true, true},
		{"empty + shared=true rejected", "", true, true},
		{"v3 + shared=true accepted", APIVersionV3, true, false},
		{"v1 + shared=false accepted", APIVersionV1, false, false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := newValidConfig()
			c.APIVersion = tc.apiVersion
			c.Options.Shared = tc.shared
			// v2 needs a shared subchart to satisfy checkSubCharts, otherwise the
			// test would fail for an unrelated reason. Keep the focus on the
			// options.shared cross-field rule.
			if tc.apiVersion == APIVersionV2 {
				c.Spec.SubCharts = []Chart{{Name: "main", Shared: true}}
			}
			// validateSharedAppRequirements also fires on shared=true and
			// would mask the apiVersion gate for the v3 happy path. Set
			// the minimum extra fields it demands so the assertion stays
			// focused on the shared+v3 rule.
			//
			// apiVersion=v3 (and options.shared=true) are 1.12.6-only
			// triggers, so the legacy fixture would also trip
			// validateModernFieldRequiresManifestVersion before the v3
			// happy path is reached. Bump the manifest to a modern
			// version and pin the Olares dep so the assertion really
			// tracks the shared/v3 rule and nothing else.
			if tc.apiVersion == APIVersionV3 && tc.shared {
				c.Spec.OnlyAdmin = true
				c.ConfigVersion = "0.13.0"
				wr := WorkloadReplicas{c.Metadata.Name: 1}
				c.WorkloadReplicas = &wr
				c.Options.Dependencies = []Dependency{newOlaresSystemDep(c)}
			}
			err := ValidateAppConfiguration(c)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %s", tc.name)
				}
				if !strings.Contains(err.Error(), "options.shared=true is only supported for apiVersion=v3") {
					t.Fatalf("error should flag the shared cross-field rule, got: %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error for %s: %v", tc.name, err)
				}
			}
		})
	}
}

// TestOptions_SharedAppRequirements documents the cross-field gates that
// fire on top of the apiVersion=v3 requirement whenever options.shared
// is true. Each subtest starts from a v3 manifest that satisfies every
// other validator gate (the helper is intentionally kept minimal: only
// what the shared/v3 path requires) and flips exactly one offending
// field. The negative cases assert both that the error fires AND that
// the message names the specific field, so a manifest with multiple
// violations gives the author a precise checklist.
func TestOptions_SharedAppRequirements(t *testing.T) {
	// validSharedV3 builds the canonical "everything set up correctly"
	// shared+v3 manifest used as the baseline for every subtest. Each
	// subtest copies it and flips exactly one field to verify a single
	// gate in isolation.
	validSharedV3 := func() *AppConfiguration {
		c := newValidConfig()
		c.ConfigVersion = "0.13.0"
		c.APIVersion = APIVersionV3
		c.Options.Shared = true
		c.Spec.OnlyAdmin = true
		wr := WorkloadReplicas{c.Metadata.Name: 1}
		c.WorkloadReplicas = &wr
		c.Options.Dependencies = []Dependency{newOlaresSystemDep(c)}
		return c
	}

	t.Run("baseline_accepted", func(t *testing.T) {
		// Anchor test: confirms the helper itself is a valid manifest so
		// every "flip one field" subtest below can attribute its failure
		// to the flipped field rather than to a baseline issue.
		if err := ValidateAppConfiguration(validSharedV3()); err != nil {
			t.Fatalf("shared+v3 baseline must validate: %v", err)
		}
	})

	t.Run("missing_only_admin_rejected", func(t *testing.T) {
		c := validSharedV3()
		c.Spec.OnlyAdmin = false
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: spec.onlyAdmin must be true when options.shared=true")
		}
		if !strings.Contains(err.Error(), "spec.onlyAdmin must be true when options.shared=true") {
			t.Fatalf("error should flag the onlyAdmin gate, got: %v", err)
		}
	})

	t.Run("subcharts_set_rejected", func(t *testing.T) {
		c := validSharedV3()
		c.Spec.SubCharts = []Chart{
			{Name: "ollamaserver", Shared: true},
			{Name: "ollamav2"},
		}
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: spec.subCharts must be empty when options.shared=true")
		}
		if !strings.Contains(err.Error(), "spec.subCharts must be empty when options.shared=true") {
			t.Fatalf("error should flag the subCharts gate, got: %v", err)
		}
	})

	t.Run("appscope_cluster_scoped_rejected", func(t *testing.T) {
		c := validSharedV3()
		c.Options.AppScope.ClusterScoped = true
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: options.appScope.clusterScoped must be false when options.shared=true")
		}
		if !strings.Contains(err.Error(), "options.appScope.clusterScoped must be false when options.shared=true") {
			t.Fatalf("error should flag the clusterScoped gate, got: %v", err)
		}
	})

	t.Run("appscope_app_ref_rejected", func(t *testing.T) {
		c := validSharedV3()
		c.Options.AppScope.AppRef = []string{"ollamav2"}
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected error: options.appScope.appRef must be empty when options.shared=true")
		}
		if !strings.Contains(err.Error(), "options.appScope.appRef must be empty when options.shared=true") {
			t.Fatalf("error should flag the appRef gate, got: %v", err)
		}
	})

	t.Run("multiple_violations_are_aggregated", func(t *testing.T) {
		// A manifest authored against the wrong schema can easily hit
		// every gate at once. The aggregator must surface them all in a
		// single Lint run so the author sees the full checklist instead
		// of fixing one field, re-running, and discovering the next.
		c := validSharedV3()
		c.Spec.OnlyAdmin = false
		c.Spec.SubCharts = []Chart{{Name: "ollamaserver", Shared: true}}
		c.Options.AppScope = AppScope{
			ClusterScoped: true,
			AppRef:        []string{"ollamav2"},
		}
		err := ValidateAppConfiguration(c)
		if err == nil {
			t.Fatal("expected aggregated errors for multiple shared-app violations")
		}
		msg := err.Error()
		wantFragments := []string{
			"spec.onlyAdmin must be true when options.shared=true",
			"spec.subCharts must be empty when options.shared=true",
			"options.appScope.clusterScoped must be false when options.shared=true",
			"options.appScope.appRef must be empty when options.shared=true",
		}
		for _, frag := range wantFragments {
			if !strings.Contains(msg, frag) {
				t.Fatalf("aggregated error should contain %q, got: %v", frag, err)
			}
		}
	})

	t.Run("rules_dormant_when_shared_false", func(t *testing.T) {
		// Every constraint here is gated on options.shared=true. Build
		// a config that would otherwise violate every gate, then leave
		// shared=false and confirm validation passes (the offending
		// fields are unrelated to the shared rule when shared is off).
		// This documents that the gate keys off shared, not off the
		// presence of subCharts/appScope/onlyAdmin themselves.
		c := newValidConfig()
		c.APIVersion = APIVersionV2
		c.Spec.OnlyAdmin = false
		c.Spec.SubCharts = []Chart{{Name: "main", Shared: true}}
		c.Options.AppScope = AppScope{
			ClusterScoped: true,
			AppRef:        []string{"other"},
		}
		// options.shared stays at its zero value (false).
		if err := ValidateAppConfiguration(c); err != nil {
			t.Fatalf("shared=false config must not trip the shared-app gates: %v", err)
		}
	})
}
