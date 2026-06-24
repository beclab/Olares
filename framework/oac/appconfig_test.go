package oac_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/oac"
	apimanifest "github.com/beclab/api/manifest"
)

func TestLoadAppConfiguration_Firefox(t *testing.T) {
	cfg, err := oac.LoadAppConfiguration(testdataFirefox, oac.WithOwnerAdmin("alice"))
	if err != nil {
		t.Fatalf("LoadAppConfiguration: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil *AppConfiguration")
	}
	if cfg.Metadata.Name == "" {
		t.Fatalf("expected metadata.name to be populated, got %+v", cfg.Metadata)
	}
	if cfg.ConfigVersion == "" {
		t.Fatal("expected olaresManifest.version to be populated")
	}
}

func TestLoadAppConfigurationContent_RoundTrip(t *testing.T) {
	const content = `olaresManifest.version: 0.12.0
olaresManifest.type: app
apiVersion: v1
metadata:
  name: demo
  title: Demo
  version: 1.0.0
spec:
  versionName: v1
`
	cfg, err := oac.LoadAppConfigurationContent([]byte(content))
	if err != nil {
		t.Fatalf("LoadAppConfigurationContent: %v", err)
	}
	if cfg.Metadata.Name != "demo" {
		t.Fatalf("metadata.name = %q, want demo", cfg.Metadata.Name)
	}
	if cfg.APIVersion != "v1" {
		t.Fatalf("apiVersion = %q, want v1", cfg.APIVersion)
	}
}

func TestAsAppConfiguration_NilManifest(t *testing.T) {
	cfg, ok := oac.AsAppConfiguration(nil)
	if ok || cfg != nil {
		t.Fatalf("nil Manifest must yield (nil, false), got (%+v, %v)", cfg, ok)
	}
}

func TestAsAppConfiguration_FromLoadedManifest(t *testing.T) {
	c := oac.New(oac.WithOwnerAdmin("alice"))
	m, err := c.LoadManifestFile(testdataFirefox)
	if err != nil {
		t.Fatalf("LoadManifestFile: %v", err)
	}
	cfg, ok := oac.AsAppConfiguration(m)
	if !ok || cfg == nil {
		t.Fatal("AsAppConfiguration should unwrap a freshly-loaded Manifest")
	}
	if cfg.Metadata.Name == "" {
		t.Fatalf("expected populated cfg, got %+v", cfg.Metadata)
	}
}

// TestLoadAppConfiguration_ReturnsAPIManifestAlias pins the contract that
// oac.LoadAppConfiguration's declared return type (*AppConfiguration)
// is a pure type alias for *github.com/beclab/api/manifest.AppConfiguration.
// That means an *apimanifest.AppConfiguration variable can receive the
// result directly — no type assertion, no conversion — and reflect.TypeOf
// reports identical types on both sides.
func TestLoadAppConfiguration_ReturnsAPIManifestAlias(t *testing.T) {
	var cfg *apimanifest.AppConfiguration
	cfg, err := oac.LoadAppConfiguration(testdataFirefox, oac.WithOwnerAdmin("alice"))
	if err != nil {
		t.Fatalf("LoadAppConfiguration: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil *apimanifest.AppConfiguration")
	}
	if cfg.Metadata.Name == "" {
		t.Fatalf("expected metadata.name to be populated, got %+v", cfg.Metadata)
	}

	var oacCfg *oac.AppConfiguration
	if got, want := reflect.TypeOf(cfg), reflect.TypeOf(oacCfg); got != want {
		t.Fatalf("*apimanifest.AppConfiguration (%v) and *oac.AppConfiguration (%v) must be the exact same reflect.Type",
			got, want)
	}
}

// TestCheckerValidateAppConfiguration_Firefox is the happy-path contract:
// a cfg that just came back from LoadAppConfiguration must pass
// (*Checker).ValidateAppConfiguration cleanly — round-tripping a real
// fixture covers the common "load, tweak, re-validate" workflow.
func TestCheckerValidateAppConfiguration_Firefox(t *testing.T) {
	c := oac.New(oac.WithOwnerAdmin("alice"))
	cfg, err := c.LoadAppConfiguration(testdataFirefox)
	if err != nil {
		t.Fatalf("LoadAppConfiguration: %v", err)
	}
	if err := c.ValidateAppConfiguration(cfg); err != nil {
		t.Fatalf("ValidateAppConfiguration on a freshly-loaded firefox cfg must pass, got: %v", err)
	}
}

// TestCheckerValidateAppConfiguration_ReportsValidationError breaks a
// required field and asserts the Checker method surfaces it as a
// *ValidationError (same error model as ValidateManifestContent).
func TestCheckerValidateAppConfiguration_ReportsValidationError(t *testing.T) {
	c := oac.New(oac.WithOwnerAdmin("alice"))
	cfg, err := c.LoadAppConfiguration(testdataFirefox)
	if err != nil {
		t.Fatalf("LoadAppConfiguration: %v", err)
	}

	cfg.Metadata.Name = "" // required by the schema
	err = c.ValidateAppConfiguration(cfg)
	if err == nil {
		t.Fatal("expected validation error when metadata.name is blanked out")
	}
	var vErr *oac.ValidationError
	if !errors.As(err, &vErr) {
		t.Fatalf("expected *oac.ValidationError, got %T: %v", err, err)
	}
	if vErr.Version == "" {
		t.Fatalf("ValidationError.Version should be populated, got %+v", vErr)
	}
}

// TestCheckerValidateAppConfiguration_SkipManifest mirrors how
// SkipManifestCheck() disables the manifest stage in Lint and
// ValidateManifestFile: the method must return nil even against an
// obviously invalid cfg.
func TestCheckerValidateAppConfiguration_SkipManifest(t *testing.T) {
	cfg := &oac.AppConfiguration{} // everything zero — normally a hard fail

	c := oac.New(oac.SkipManifestCheck())
	if err := c.ValidateAppConfiguration(cfg); err != nil {
		t.Fatalf("SkipManifestCheck must short-circuit validation, got: %v", err)
	}

	// Sanity check: without the skip, the same cfg must fail.
	if err := oac.New().ValidateAppConfiguration(cfg); err == nil {
		t.Fatal("empty cfg should fail validation when SkipManifestCheck is not set")
	}
}

// TestCheckerValidateAppConfiguration_NilCfg guarantees we don't panic on
// a nil pointer and that the caller still gets a *ValidationError they
// can inspect with errors.As — same contract as every other Validate*
// entry point.
func TestCheckerValidateAppConfiguration_NilCfg(t *testing.T) {
	err := oac.New().ValidateAppConfiguration(nil)
	if err == nil {
		t.Fatal("expected error for nil cfg, got nil")
	}
	var vErr *oac.ValidationError
	if !errors.As(err, &vErr) {
		t.Fatalf("expected *oac.ValidationError, got %T: %v", err, err)
	}
}

// TestValidateAppConfiguration_PackageLevelShortcut asserts that the
// Checker-less convenience wrapper accepts the same Option values (at
// least SkipManifestCheck — the others have no effect on a bare cfg)
// and produces identical behaviour.
func TestValidateAppConfiguration_PackageLevelShortcut(t *testing.T) {
	cfg := &oac.AppConfiguration{}
	if err := oac.ValidateAppConfiguration(cfg); err == nil {
		t.Fatal("package-level ValidateAppConfiguration must fail on an empty cfg")
	}
	if err := oac.ValidateAppConfiguration(cfg, oac.SkipManifestCheck()); err != nil {
		t.Fatalf("SkipManifestCheck must disable the package-level shortcut too, got: %v", err)
	}
}

// TestLoadAppConfigurationContent_NormalizesAppID pins the loader contract
// that metadata.appid is overwritten with md5(metadata.name)[:8] regardless
// of what the manifest declared. The value matches
// framework/app-service/pkg/appcfg.(AppName).GetAppID for non-system apps,
// so callers can rely on the load layer producing the same id the runtime
// computes downstream.
func TestLoadAppConfigurationContent_NormalizesAppID(t *testing.T) {
	const content = `olaresManifest.version: 0.12.0
olaresManifest.type: app
apiVersion: v1
metadata:
  name: demo
  title: Demo
  version: 1.0.0
  appid: whatever-user-wrote
spec:
  versionName: v1
`
	cfg, err := oac.LoadAppConfigurationContent([]byte(content))
	if err != nil {
		t.Fatalf("LoadAppConfigurationContent: %v", err)
	}
	// md5("demo") = "fe01ce2a7fbac8fafaed7c982a04e229" -> first 8 hex chars.
	const want = "fe01ce2a"
	if cfg.Metadata.AppID != want {
		t.Fatalf("metadata.appid = %q, want %q (md5(metadata.name)[:8])", cfg.Metadata.AppID, want)
	}
}

// TestLoadAppConfigurationContent_EmptyNameLeavesAppIDEmpty pins the
// edge-case branch of normalizeAppID: when metadata.name is missing the
// loader returns an empty appid rather than md5("")[:8] (a meaningless
// constant). The follow-up validation error will then key off the missing
// name instead of confusing the user with a synthetic hash.
func TestLoadAppConfigurationContent_EmptyNameLeavesAppIDEmpty(t *testing.T) {
	const content = `olaresManifest.version: 0.12.0
olaresManifest.type: app
apiVersion: v1
metadata:
  title: Demo
  version: 1.0.0
spec:
  versionName: v1
`
	cfg, err := oac.LoadAppConfigurationContent([]byte(content))
	if err != nil {
		t.Fatalf("LoadAppConfigurationContent: %v", err)
	}
	if cfg.Metadata.AppID != "" {
		t.Fatalf("metadata.appid = %q, want empty when metadata.name is missing", cfg.Metadata.AppID)
	}
}

// TestLoadAppConfiguration_NormalizesAppID exercises the file-based loader
// against the firefox fixture: the manifest declares appid: firefox (see
// testdata/firefox/OlaresManifest.yaml), which the load layer must
// overwrite with md5("firefox")[:8].
func TestLoadAppConfiguration_NormalizesAppID(t *testing.T) {
	cfg, err := oac.LoadAppConfiguration(testdataFirefox, oac.WithOwnerAdmin("alice"))
	if err != nil {
		t.Fatalf("LoadAppConfiguration: %v", err)
	}
	// md5("firefox") = "d6a5c9544eca9b5ce2266d1c34a93222" -> first 8 hex chars.
	const want = "d6a5c954"
	if cfg.Metadata.AppID != want {
		t.Fatalf("metadata.appid = %q, want %q (md5(metadata.name)[:8])", cfg.Metadata.AppID, want)
	}
}

// TestLoadAppConfiguration_PropagatesParseError ensures that file-IO and
// parser errors surface via the convenience loader (we don't swallow them
// or wrap them in errAppConfigUnavailable by mistake).
//
// errAppConfigUnavailable is internal, so we pin the contract by its
// message substring rather than by identity: the previous version
// compared against a freshly-built errors.New(...) via errors.Is, which
// never matched because each errors.New allocates a distinct value.
func TestLoadAppConfiguration_PropagatesParseError(t *testing.T) {
	_, err := oac.LoadAppConfiguration("/no/such/path/here")
	if err == nil {
		t.Fatal("expected error for missing path")
	}
	if strings.Contains(err.Error(), "manifest is not backed by *AppConfiguration") {
		t.Fatalf("must surface IO error, not the type-assertion sentinel: %v", err)
	}
}
