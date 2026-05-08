package oac

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/beclab/Olares/framework/oac/internal/manifest"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
)

// defaultStrategy is the schema-aware parser/validator shared by every
// Pipeline we hand out. manifest.ManifestStrategy is stateless, so a single
// package-level instance is safe to share across Checkers and goroutines.
var defaultStrategy manifest.Strategy = &manifest.ManifestStrategy{}

// ManifestFileName is the well-known file name that holds the OlaresManifest.
const ManifestFileName = "OlaresManifest.yaml"

const manifestRenderKey = "chart/OlaresManifest.yaml"

const (
	defaultRenderOwner = "default"
	defaultRenderAdmin = "default"
)

// ManifestVersions is the lightweight pair of top-level version fields read
// from raw OlaresManifest.yaml before full parsing. APIVersion is the value
// after the apiVersion key; OlaresManifestVersion is the value after
// olaresManifest.version.
type ManifestVersions struct {
	APIVersion            string
	OlaresManifestVersion string
}

// PeekManifestVersions extracts apiVersion and olaresManifest.version from
// content using the same line-oriented probe as the manifest parsing pipeline
// (tolerates unrendered Helm template fragments in the scalar). Missing keys
// yield empty strings; only I/O-style scanner failures produce a non-nil error.
func PeekManifestVersions(content []byte) (ManifestVersions, error) {
	h, err := manifest.Peek(content)
	if err != nil {
		return ManifestVersions{}, err
	}
	return ManifestVersions{
		APIVersion: func() string {
			if h.APIVersion == "" {
				return "v1"
			}
			return h.APIVersion
		}(),
		OlaresManifestVersion: h.OlaresVersion,
	}, nil
}

// Manifest is the cross-version, read-only view of a parsed OlaresManifest.
// Raw() yields *github.com/beclab/api/manifest.AppConfiguration (same type as
// oac.AppConfiguration — the two are type-aliased together).
type Manifest = manifest.Manifest

// EntranceInfo is a lightweight view of a manifest entrance shared between
// versions.
type EntranceInfo = manifest.EntranceInfo

// LoadManifestFile reads OlaresManifest.yaml from oacPath and returns the
// parsed manifest. Legacy (<0.12.0) payloads are template-rendered with the
// checker's owner/admin before parsing; modern (>=0.12.0) payloads are parsed
// verbatim. No validation is performed here — use
// ValidateManifestFile for that.
func (c *OAC) LoadManifestFile(oacPath string) (Manifest, error) {
	data, err := readManifestFile(oacPath)
	if err != nil {
		return nil, err
	}
	return c.parseManifest(data)
}

// LoadManifestContent is the byte-slice counterpart of LoadManifestFile.
func (c *OAC) LoadManifestContent(content []byte) (Manifest, error) {
	return c.parseManifest(content)
}

func readManifestFile(oacPath string) ([]byte, error) {
	if !strings.HasSuffix(oacPath, string(filepath.Separator)) {
		oacPath += string(filepath.Separator)
	}
	f, err := os.Open(filepath.Join(oacPath, ManifestFileName))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func (c *OAC) parseManifest(raw []byte) (Manifest, error) {
	pipe, err := c.resolvePipeline(raw)
	if err != nil {
		return nil, err
	}
	m, err := pipe.Parse(raw, c.manifestRenderer(), c.owner, c.admin)
	if err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	return m, nil
}

func (c *OAC) validateManifestBytes(raw []byte, m Manifest) error {
	pipe, err := c.resolvePipeline(raw)
	if err != nil {
		return err
	}
	if err := pipe.Validate(raw, m, c.manifestRenderer(), c.owner, c.admin); err != nil {
		return WrapValidation(m.APIVersion(), err)
	}
	return nil
}

func (c *OAC) resolvePipeline(raw []byte) (manifest.Pipeline, error) {
	header, err := manifest.Peek(raw)
	if err != nil {
		return nil, fmt.Errorf("probe manifest version: %w", err)
	}
	return manifest.NewPipeline(header.OlaresVersion, defaultStrategy), nil
}

func (c *OAC) manifestRenderer() manifest.Renderer {
	return func(raw []byte, owner, admin string) ([]byte, error) {
		return renderManifestTemplate(raw, owner, admin)
	}
}

func renderManifestTemplate(content []byte, owner, admin string) ([]byte, error) {
	if owner == "" {
		owner = defaultRenderOwner
	}
	if admin == "" {
		admin = defaultRenderAdmin
	}
	values := map[string]interface{}{
		"admin": admin,
		"bfl":   map[string]string{"username": owner},
	}
	ch := &chart.Chart{
		Metadata:  &chart.Metadata{Name: "chart", Version: "0.0.1"},
		Templates: []*chart.File{{Name: ManifestFileName, Data: content}},
	}
	renderValues, err := chartutil.ToRenderValues(ch, values, chartutil.ReleaseOptions{}, nil)
	if err != nil {
		return nil, err
	}
	e := engine.Engine{}
	rendered, err := e.Render(ch, renderValues)
	if err != nil {
		return nil, err
	}
	out, ok := rendered[manifestRenderKey]
	if !ok {
		return nil, fmt.Errorf("helm engine did not emit %q", manifestRenderKey)
	}
	return []byte(out), nil
}
