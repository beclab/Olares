package manifest

import (
	"strings"
	"testing"
)

// v3EnvManifest returns a modern (>= 0.12.0) v3 manifest pre-populated
// with the trigger-side prerequisites — workloadReplicas and the Olares
// system dependency locked to the post-v3 window — so the assertion in
// each test focuses on v3 env semantics rather than the manifest-version
// gate (apiVersion=v3 is itself a 1.12.6-only trigger, so the legacy
// fixture would otherwise trip validateModernFieldRequiresManifestVersion
// before envs are even inspected).
func v3EnvManifest(envs []AppEnvVar) *AppConfiguration {
	c := newValidConfig()
	c.ConfigVersion = "0.13.0"
	c.APIVersion = APIVersionV3
	wr := WorkloadReplicas{c.Metadata.Name: 1}
	c.WorkloadReplicas = &wr
	c.Options.Dependencies = []Dependency{newOlaresSystemDep(c)}
	c.Envs = envs
	return c
}

func TestValidateV3Envs_AppLocalNameAllowed(t *testing.T) {
	c := v3EnvManifest([]AppEnvVar{{
		EnvVarSpec: EnvVarSpec{
			EnvName: "SMTP_HOST",
		},
		ApplyOnChange: true,
		ValueFrom: &ValueFrom{
			EnvName: "OLARES_USER_SMTP_SERVER",
		},
	}})
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("valueFrom may reference OLARES_USER_*: %v", err)
	}
}

func TestValidateV3Envs_EnvNameCannotUseOLARESUserPrefix(t *testing.T) {
	c := v3EnvManifest([]AppEnvVar{{
		EnvVarSpec: EnvVarSpec{
			EnvName: "OLARES_USER_SMTP_SERVER",
		},
	}})
	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("expected error when envs[].envName starts with OLARES_USER")
	}
	if !strings.Contains(err.Error(), "envs[0].envName") {
		t.Fatalf("error should point at envs[0].envName, got: %v", err)
	}
}

func TestValidateV3Envs_SkippedForV1(t *testing.T) {
	c := newValidConfig()
	c.Envs = []AppEnvVar{{
		EnvVarSpec: EnvVarSpec{
			EnvName: "OLARES_USER_EMAIL",
		},
	}}
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("OLARES_USER envName rule applies only to apiVersion=v3: %v", err)
	}
}
