package preinstall

import "encoding/json"

const (
	SupportedSchemaVersion   = "1"
	OfficialSourceID         = "market.olares"
	BundleFileName           = "bundle.json"
	ProfileFileName          = "install-profile.json"
	StaticRelativeDir        = "preinstall/market"
	RuntimeRelativeDir       = "userdata/Cache/market-preinstall"
	MaxBundleJSONBytes       = 8 << 20
	MaxProfileJSONBytes      = 8 << 20
	MaxBundleApps            = 256
	MaxChartBytes            = 256 << 20
	MaxTotalChartBytes       = 1 << 30
	MaxArtifactManifestBytes = 64 << 20
	MaxArtifactEntries       = 1_000_000
	MaxArtifactTotalSize     = 1 << 40
	ArtifactKindHFCache      = "hf-cache"
)

type BundleV1 struct {
	SchemaVersion string        `json:"schemaVersion"`
	SourceID      string        `json:"sourceId"`
	CatalogHash   string        `json:"catalogHash"`
	GeneratedAt   string        `json:"generatedAt"`
	Apps          []BundleAppV1 `json:"apps"`
}

type BundleAppV1 struct {
	AppID           string             `json:"appId"`
	AppName         string             `json:"appName"`
	Version         string             `json:"version"`
	Chart           string             `json:"chart"`
	ChartSHA256     string             `json:"chartSha256"`
	InstallOrder    int                `json:"installOrder"`
	AllowedEnvs     []string           `json:"allowedEnvs,omitempty"`
	DefaultEnvs     map[string]string  `json:"defaultEnvs,omitempty"`
	AllowedGPUTypes []string           `json:"allowedGpuTypes,omitempty"`
	Artifacts       []BundleArtifactV1 `json:"artifacts,omitempty"`
	AppEntry        json.RawMessage    `json:"appEntry"`
}

type BundleArtifactV1 struct {
	Kind           string `json:"kind"`
	Source         string `json:"source"`
	Repo           string `json:"repo"`
	Revision       string `json:"revision"`
	Manifest       string `json:"manifest"`
	ManifestSHA256 string `json:"manifestSha256"`
	TotalSize      int64  `json:"totalSize"`
}

type ArtifactManifestV1 struct {
	SchemaVersion int                       `json:"schemaVersion"`
	Repo          string                    `json:"repo"`
	Revision      string                    `json:"revision"`
	Entries       []ArtifactManifestEntryV1 `json:"entries"`
}

type ArtifactManifestEntryV1 struct {
	Path   string `json:"path"`
	Type   string `json:"type"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256,omitempty"`
	Target string `json:"target,omitempty"`
}

type InstallProfileV1 struct {
	SchemaVersion   string                `json:"schemaVersion"`
	HardwareProfile string                `json:"hardwareProfile,omitempty"`
	Apps            []InstallProfileAppV1 `json:"apps"`
}

type InstallProfileAppV1 struct {
	AppID           string            `json:"appId"`
	SelectedGPUType string            `json:"selectedGpuType,omitempty"`
	Envs            map[string]string `json:"envs,omitempty"`
}

type ContractV1 struct {
	Bundle  BundleV1
	Profile InstallProfileV1
}

type ProfileSelections struct {
	HardwareProfile string
	DetectedGPUType string
	Apps            map[string]AppSelection
}

type AppSelection struct {
	SelectedGPUType string
	Envs            map[string]string
}
