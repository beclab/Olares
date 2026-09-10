package preinstall

import "encoding/json"

const (
	// SupportedSchemaVersion is the schema of the static bundle on the medium.
	// It is not the schema of what this program publishes: the bundle is an
	// input written by the image build, the declaration is the output Market
	// reads, and the two version independently.
	SupportedSchemaVersion = "1"
	OfficialSourceID       = "market.olares"
	BundleFileName         = "bundle.json"
	StaticRelativeDir      = "preinstall/market"
	// RuntimeRelativeDir is relative to the Olares root directory
	// (storage.OlaresRootDir), which is what the market chart renders as
	// .Values.rootPath for its read-only hostPath mount. It is not relative to
	// the installer base directory.
	RuntimeRelativeDir             = "userdata/Cache/market-preinstall"
	MaxBundleJSONBytes             = 8 << 20
	MaxBundleApps                  = 256
	MaxChartBytes            int64 = 256 << 20
	MaxTotalChartBytes             = 1 << 30
	MaxArtifactManifestBytes int64 = 64 << 20
	MaxArtifactEntries             = 1_000_000
	MaxArtifactTotalSize     int64 = 1 << 40
	ArtifactKindHFCache            = "hf-cache"
	InstallScopeShared             = InstallScope("shared")
	InstallScopePerUser            = InstallScope("per-user")
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
	InstallScope    InstallScope       `json:"installScope"`
	Chart           string             `json:"chart"`
	ChartSHA256     string             `json:"chartSha256"`
	InstallOrder    int                `json:"installOrder"`
	AllowedEnvs     []string           `json:"allowedEnvs,omitempty"`
	DefaultEnvs     map[string]string  `json:"defaultEnvs,omitempty"`
	AllowedGPUTypes []string           `json:"allowedGpuTypes,omitempty"`
	Artifacts       []BundleArtifactV1 `json:"artifacts,omitempty"`
	AppEntry        json.RawMessage    `json:"appEntry"`
}

type InstallScope string

func (s InstallScope) Valid() bool {
	return s == InstallScopeShared || s == InstallScopePerUser
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

// ProfileSelections is what the installer decided while running: the hardware it
// found and anything the operator chose per app. It is folded into the published
// declaration rather than written beside it, because Market has no use for the
// choices that were available and not taken.
type ProfileSelections struct {
	HardwareProfile string
	DetectedGPUType string
	Apps            map[string]AppSelection
}

type AppSelection struct {
	SelectedGPUType string
	Envs            map[string]string
}
