package manifest

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// ManifestStrategy parses YAML into github.com/beclab/api/manifest.AppConfiguration
// and runs Olares validation rules.
type ManifestStrategy struct{}

// Parse unmarshals YAML; it does not validate.
func (s *ManifestStrategy) Parse(rendered []byte) (Manifest, error) {
	var cfg AppConfiguration
	if err := yaml.Unmarshal(rendered, &cfg); err != nil {
		return nil, err
	}
	return &parsedManifest{cfg: &cfg}, nil
}

// Validate runs structural and cross-field checks.
func (s *ManifestStrategy) Validate(m Manifest) error {
	mm, ok := m.(*parsedManifest)
	if !ok {
		return fmt.Errorf("manifest strategy received foreign manifest type: %T", m)
	}
	return ValidateAppConfiguration(mm.cfg)
}

type parsedManifest struct {
	cfg *AppConfiguration
}

func (m *parsedManifest) APIVersion() string {
	if m.cfg.APIVersion == "" {
		return APIVersionV1
	}
	return m.cfg.APIVersion
}

func (m *parsedManifest) ConfigVersion() string { return m.cfg.ConfigVersion }
func (m *parsedManifest) ConfigType() string    { return m.cfg.ConfigType }
func (m *parsedManifest) AppName() string       { return m.cfg.Metadata.Name }
func (m *parsedManifest) AppVersion() string    { return m.cfg.Metadata.Version }

func (m *parsedManifest) Entrances() []EntranceInfo {
	out := make([]EntranceInfo, 0, len(m.cfg.Entrances))
	for _, e := range m.cfg.Entrances {
		out = append(out, EntranceInfo{Name: e.Name, Host: e.Host, Port: e.Port})
	}
	return out
}

func (m *parsedManifest) OptionsImages() []string {
	if len(m.cfg.Options.Images) == 0 {
		return nil
	}
	out := make([]string, len(m.cfg.Options.Images))
	copy(out, m.cfg.Options.Images)
	return out
}

func (m *parsedManifest) PermissionAppData() bool { return m.cfg.Permission.AppData }

func (m *parsedManifest) Raw() any { return m.cfg }
