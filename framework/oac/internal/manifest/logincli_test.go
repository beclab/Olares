package manifest

import (
	"strings"
	"testing"
)

func TestIsLoginOlaresCLIAllowed(t *testing.T) {
	t.Setenv(loginOlaresCLIAllowListEnv, "")
	if !IsLoginOlaresCLIAllowed("lares") {
		t.Error("lares must be on the loginOlaresCLI allowlist")
	}
	if !IsLoginOlaresCLIAllowed("  lares  ") {
		t.Error("IsLoginOlaresCLIAllowed must trim surrounding whitespace")
	}
	for _, s := range []string{
		"", "firefox", "LARES", "lares2", " lares-cli",
	} {
		if IsLoginOlaresCLIAllowed(s) {
			t.Errorf("IsLoginOlaresCLIAllowed(%q) = true, want false", s)
		}
	}
}

func TestIsLoginOlaresCLIAllowedEnvOverridesCompiledList(t *testing.T) {
	t.Setenv(loginOlaresCLIAllowListEnv, "firefox, nitro")
	if !IsLoginOlaresCLIAllowed("firefox") || !IsLoginOlaresCLIAllowed("nitro") {
		t.Fatal("names in LOGIN_OLARES_CLI_ALLOW_LIST must be allowed")
	}
	if IsLoginOlaresCLIAllowed("lares") {
		t.Fatal("a non-empty env list replaces the compiled allowlist, it does not merge")
	}
}

func TestIsLoginOlaresCLIAllowedEmptyEnvFallsBack(t *testing.T) {
	t.Setenv(loginOlaresCLIAllowListEnv, "  ,  ")
	if !IsLoginOlaresCLIAllowed("lares") {
		t.Fatal("whitespace-only LOGIN_OLARES_CLI_ALLOW_LIST must fall back to the compiled list")
	}
}

// newLoginCLIBaseline returns a modern manifest that already satisfies every
// prerequisite permission.loginOlaresCLI drags in, so each test below can
// toggle exactly one thing. loginOlaresCLI is a 1.12.6-only trigger field, so
// without the version bump / locked dep the manifest would fail on the
// version gate before the allowlist is ever consulted.
func newLoginCLIBaseline() *AppConfiguration {
	c := newValidConfig()
	c.ConfigVersion = "0.13.0"
	wr := WorkloadReplicas{c.Metadata.Name: 1}
	c.WorkloadReplicas = &wr
	c.Options.Dependencies = []Dependency{newOlaresSystemDep(c)}
	return c
}

// Lint rule: an app that is not on the allowlist cannot declare
// permission.loginOlaresCLI, and the error names both the field and the
// offending app so the submitter knows why gitbot refused the manifest.
func TestPermission_LoginOlaresCLINotAllowlisted(t *testing.T) {
	t.Setenv(loginOlaresCLIAllowListEnv, "")
	c := newLoginCLIBaseline() // metadata.name "firefox"
	c.Permission.LoginOlaresCLI = true
	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("permission.loginOlaresCLI must be rejected for an app off the allowlist")
	}
	if !strings.Contains(err.Error(), "permission.loginOlaresCLI") {
		t.Fatalf("error should mention permission.loginOlaresCLI, got: %v", err)
	}
	if !strings.Contains(err.Error(), "firefox") {
		t.Fatalf("error should echo the offending app name, got: %v", err)
	}
}

// Lint rule: an allowlisted app declaring the permission on a modern manifest
// passes untouched.
func TestPermission_LoginOlaresCLIAllowlisted(t *testing.T) {
	t.Setenv(loginOlaresCLIAllowListEnv, "")
	c := newLoginCLIBaseline()
	c.Metadata.Name = "lares"
	wr := WorkloadReplicas{c.Metadata.Name: 1}
	c.WorkloadReplicas = &wr
	c.Permission.LoginOlaresCLI = true
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("permission.loginOlaresCLI should be accepted for an allowlisted app: %v", err)
	}
}

// Lint rule: being on the allowlist does not exempt an app from the version
// gate. loginOlaresCLI is a 1.12.6-only field, so a legacy manifest is
// rejected by validateModernFieldRequiresManifestVersion even for lares.
func TestPermission_LoginOlaresCLIVersionGate(t *testing.T) {
	t.Setenv(loginOlaresCLIAllowListEnv, "")
	c := newValidConfig() // ConfigVersion "0.11.0"
	c.Metadata.Name = "lares"
	c.Permission.LoginOlaresCLI = true
	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("permission.loginOlaresCLI must be rejected on olaresManifest.version < 0.12.0")
	}
	if !strings.Contains(err.Error(), "permission.loginOlaresCLI") {
		t.Fatalf("error should mention permission.loginOlaresCLI, got: %v", err)
	}
	if !strings.Contains(err.Error(), minResourcesManifestVersion.String()) {
		t.Fatalf("error should state the required olaresManifest.version, got: %v", err)
	}
}

// Lint rule: a per-user credential has no meaning on a cluster-wide shared
// app, even when the app is on the allowlist.
func TestPermission_LoginOlaresCLIRejectsShared(t *testing.T) {
	t.Setenv(loginOlaresCLIAllowListEnv, "")
	c := newLoginCLIBaseline()
	c.Metadata.Name = "lares"
	c.APIVersion = APIVersionV3
	c.Options.Shared = true
	c.Spec.OnlyAdmin = true
	wr := WorkloadReplicas{c.Metadata.Name: 1}
	c.WorkloadReplicas = &wr
	c.Permission.LoginOlaresCLI = true
	err := ValidateAppConfiguration(c)
	if err == nil {
		t.Fatal("permission.loginOlaresCLI must be rejected when options.shared is true")
	}
	if !strings.Contains(err.Error(), "permission.loginOlaresCLI") {
		t.Fatalf("error should mention permission.loginOlaresCLI, got: %v", err)
	}
	if !strings.Contains(err.Error(), "options.shared") {
		t.Fatalf("error should mention options.shared, got: %v", err)
	}
}
