package manifest

import (
	"errors"
	"fmt"

	"github.com/Masterminds/semver/v3"
)

const legacyManifestThreshold = "0.12.0"

var legacyVersionBoundary = semver.MustParse(legacyManifestThreshold)

var (
	ErrNilRenderer = errors.New("manifest: legacy manifest requires a Renderer")
)

// EntranceInfo is a cross-version, read-only view of a single entrance entry.
type EntranceInfo struct {
	Name string
	Host string
	Port int32
}

// Manifest is the cross-version, read-only view of a parsed OlaresManifest.
type Manifest interface {
	APIVersion() string
	ConfigVersion() string
	ConfigType() string
	AppName() string
	AppVersion() string
	Entrances() []EntranceInfo
	OptionsImages() []string
	PermissionAppData() bool
	Raw() any
}

// Strategy parses and validates a manifest of one specific version family.
type Strategy interface {
	Parse(rendered []byte) (Manifest, error)
	Validate(Manifest) error
}

// Renderer resolves helm-style template placeholders inside OlaresManifest.yaml.
type Renderer func(raw []byte, owner, admin string) ([]byte, error)

// Pipeline is the version-aware manifest processing flow.
//
// Validate accepts an optional "parsed" argument — the Manifest produced by
// a prior Parse call on the same raw bytes and (owner, admin) tuple. When
// present, single-owner pipelines can skip the expensive re-parse step; for
// pipelines whose validation runs under different (owner, admin) scenarios
// than Parse did, the argument is ignored. Passing nil is always safe and
// restores the "parse from raw" fallback.
type Pipeline interface {
	Parse(raw []byte, render Renderer, owner, admin string) (Manifest, error)
	Validate(raw []byte, parsed Manifest, render Renderer, owner, admin string) error
	Strategy() Strategy
}

// NewPipeline picks the version-aware Pipeline for a given manifest.
func NewPipeline(olaresVersion string, strat Strategy) Pipeline {
	if isLegacyVersion(olaresVersion) {
		return dualOwnerPipeline{strat: strat}
	}
	return singlePipeline{strat: strat}
}

func isLegacyVersion(v string) bool {
	if v == "" {
		return false
	}
	got, err := semver.NewVersion(v)
	if err != nil {
		return false
	}
	return got.LessThan(legacyVersionBoundary)
}

type singlePipeline struct{ strat Strategy }

func (p singlePipeline) Parse(raw []byte, _ Renderer, _, _ string) (Manifest, error) {
	return p.strat.Parse(raw)
}

// Validate reuses parsed when the caller already parsed raw; the raw bytes
// and their parse result cannot disagree in the single-owner flow (there is
// no per-scenario render indirection), so a second Parse would be a pure
// overhead. When parsed is nil we fall back to parsing raw ourselves.
func (p singlePipeline) Validate(raw []byte, parsed Manifest, _ Renderer, _, _ string) error {
	m := parsed
	if m == nil {
		var err error
		m, err = p.strat.Parse(raw)
		if err != nil {
			return err
		}
	}
	return p.strat.Validate(m)
}

func (p singlePipeline) Strategy() Strategy { return p.strat }

type dualOwnerPipeline struct{ strat Strategy }

const (
	dualOwnerFallbackAdmin = "admin"
	dualOwnerFallbackUser  = "user"
)

func (p dualOwnerPipeline) Parse(raw []byte, render Renderer, owner, admin string) (Manifest, error) {
	if render == nil {
		return nil, ErrNilRenderer
	}
	rendered, err := render(raw, owner, admin)
	if err != nil {
		return nil, fmt.Errorf("render manifest template: %w", err)
	}
	return p.strat.Parse(rendered)
}

// Validate ignores the pre-parsed Manifest: the two scenarios it runs use
// their own (owner, admin) values that intentionally differ from the ones
// Parse was invoked with, so any manifest carried in from Parse would not
// cover either scenario.
func (p dualOwnerPipeline) Validate(raw []byte, _ Manifest, render Renderer, owner, admin string) error {
	if render == nil {
		return ErrNilRenderer
	}
	adminValue := admin
	if adminValue == "" {
		adminValue = dualOwnerFallbackAdmin
	}
	userValue := owner
	if userValue == "" || userValue == adminValue {
		userValue = dualOwnerFallbackUser
	}
	scenarios := []struct {
		label string
		owner string
		admin string
	}{
		{"admin=owner", adminValue, adminValue},
		{"admin!=owner", userValue, adminValue},
	}
	var errs []error
	for _, sc := range scenarios {
		rendered, err := render(raw, sc.owner, sc.admin)
		if err != nil {
			errs = append(errs, fmt.Errorf("render %s scenario: %w", sc.label, err))
			continue
		}
		m, err := p.strat.Parse(rendered)
		if err != nil {
			errs = append(errs, fmt.Errorf("parse %s scenario: %w", sc.label, err))
			continue
		}
		if err := p.strat.Validate(m); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", sc.label, err))
		}
	}
	return errors.Join(errs...)
}

func (p dualOwnerPipeline) Strategy() Strategy { return p.strat }
