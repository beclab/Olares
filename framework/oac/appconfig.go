package oac

import (
	olm "github.com/beclab/Olares/framework/oac/internal/manifest"
	apimanifest "github.com/beclab/api/manifest"
)

// AppConfiguration is the parsed OlaresManifest.yaml payload. It is a
// direct type alias onto github.com/beclab/api/manifest.AppConfiguration,
// so reading every field works out of the box:
//
//	cfg, _ := oac.LoadAppConfiguration("./myapp")
//	for _, e := range cfg.Entrances { fmt.Println(e.Name, e.Port) }
//
// When you also need to construct sub-structs (AppMetaData, Entrance,
// Options, ResourceMode, …) — for tests, generators, or programmatic
// edits — import github.com/beclab/api/manifest (and, for Entrance and
// friends, github.com/beclab/api/api/app.bytetrade.io/v1alpha1) directly:
//
//	import (
//	    "github.com/beclab/oac"
//	    "github.com/beclab/api/manifest"
//	    appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
//	)
//	cfg.Spec.SubCharts = append(cfg.Spec.SubCharts, manifest.Chart{...})
//	cfg.Entrances      = append(cfg.Entrances,      appv1.Entrance{...})
//
// Because everything is type-aliased (not a new named type), values flow
// freely between oac and the upstream api packages without
// conversion.
type AppConfiguration = apimanifest.AppConfiguration

// LoadAppConfiguration reads OlaresManifest.yaml from oacPath, runs it
// through the version-aware parsing pipeline, and returns the concrete
// *AppConfiguration. No validation is performed — pair with
// ValidateManifestFile or ValidateAppConfiguration if you also want structural checks. Legacy
// manifests (<0.12.0) are template-rendered with the Checker's
// owner/admin before parsing; modern (>=0.12.0) manifests are parsed
// verbatim.
func (c *OAC) LoadAppConfiguration(oacPath string) (*AppConfiguration, error) {
	m, err := c.LoadManifestFile(oacPath)
	if err != nil {
		return nil, err
	}
	cfg, ok := AsAppConfiguration(m)
	if !ok {
		return nil, errAppConfigUnavailable
	}
	return cfg, nil
}

// LoadAppConfigurationContent is the byte-slice counterpart of
// LoadAppConfiguration.
func (c *OAC) LoadAppConfigurationContent(content []byte) (*AppConfiguration, error) {
	m, err := c.LoadManifestContent(content)
	if err != nil {
		return nil, err
	}
	cfg, ok := AsAppConfiguration(m)
	if !ok {
		return nil, errAppConfigUnavailable
	}
	return cfg, nil
}

// AsAppConfiguration unwraps a Manifest into the concrete
// *AppConfiguration. The second return value is false when the manifest
// was produced by a future Strategy whose Raw() type is not
// AppConfiguration; today every Strategy in the tree returns one, so
// the boolean is a forward-compat hook rather than a runtime concern.
func AsAppConfiguration(m Manifest) (*AppConfiguration, bool) {
	if m == nil {
		return nil, false
	}
	cfg, ok := m.Raw().(*AppConfiguration)
	return cfg, ok
}

// ValidateAppConfiguration runs structural and cross-field validation on an
// already-parsed manifest. It applies the same rules as
// ValidateManifestFile / ValidateManifestContent but starts from an
// in-memory *AppConfiguration instead of raw YAML — no chart rendering,
// no folder check, no custom validators (those need an oacPath and the
// chart templates, which a bare *AppConfiguration cannot supply).
//
// Respects SkipManifestCheck(): when set, the method returns nil without
// running any rule, consistent with how Lint and ValidateManifestFile
// react to the same option.
//
// A nil cfg is treated as a validation failure rather than a panic.
// Failures are wrapped as *ValidationError keyed off cfg.APIVersion
// (defaults to "v1" when empty, matching the parsing pipeline).
func (c *OAC) ValidateAppConfiguration(cfg *AppConfiguration) error {
	if c.skipManifest {
		return nil
	}
	if cfg == nil {
		return WrapValidation(olm.APIVersionV1, errAppConfigNil)
	}
	apiVersion := cfg.APIVersion
	if apiVersion == "" {
		apiVersion = olm.APIVersionV1
	}
	return WrapValidation(apiVersion, olm.ValidateAppConfiguration(cfg))
}

// ValidateAppConfiguration is the Checker-less shortcut for one-off callers.
// Options behave identically to the Checker form (today only
// SkipManifestCheck is meaningful here; the rest don't apply to a bare
// *AppConfiguration).
func ValidateAppConfiguration(cfg *AppConfiguration, opts ...Option) error {
	return New(opts...).ValidateAppConfiguration(cfg)
}

// LoadAppConfiguration is the Checker-less shortcut for one-off callers.
func LoadAppConfiguration(oacPath string, opts ...Option) (*AppConfiguration, error) {
	return New(opts...).LoadAppConfiguration(oacPath)
}

// LoadAppConfigurationContent is the byte-slice counterpart.
func LoadAppConfigurationContent(content []byte, opts ...Option) (*AppConfiguration, error) {
	return New(opts...).LoadAppConfigurationContent(content)
}
