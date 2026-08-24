package preinstall

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// A declaration is what this program tells Market the device should have. One
// file per trunk version of Olares, named for it, written once and never
// rewritten: a device that moves between builds of the same release is not a
// device that should preinstall anything again, and a device that upgrades to a
// new release gets a new file beside the old one rather than over it.
//
// Both media produce the same file. An ISO declares entries whose chart travels
// with it; an upgrade declares entries the official catalog serves. Everything
// downstream of the file is one code path.
const (
	// DeclarationSchemaVersion is what Market accepts. The two files this
	// replaced were both "1", so refusing that value is what stops an old
	// bundle.json from being read as a declaration.
	DeclarationSchemaVersion = "2"

	declarationFilePrefix = "preinstall-"
	declarationFileSuffix = ".json"

	MaxDeclarationBytes = 8 << 20
	MaxDeclarationApps  = 256

	// ChartSourceCatalog takes the chart from the official catalog, so the app
	// installs whatever version the catalog currently serves.
	ChartSourceCatalog = ChartSource("catalog")
	// ChartSourceLocal takes the chart from the payload beside the declaration,
	// which is what an air-gapped first install has.
	ChartSourceLocal = ChartSource("local")

	// CatalogLatestVersion is what Market records for an entry that pins no
	// version. It is stated here because the fixture the two repositories share
	// carries it, not because this program writes it into a declaration.
	CatalogLatestVersion = "catalog-latest"
)

// generatedNow is provenance rather than behaviour: Market branches on nothing
// in it, and a declaration written without it still imports.
var generatedNow = func() string { return time.Now().UTC().Format(time.RFC3339) }

// ChartSource says where the chart of one declared app comes from.
type ChartSource string

func (s ChartSource) Valid() bool {
	return s == ChartSourceCatalog || s == ChartSourceLocal
}

type DeclarationV2 struct {
	SchemaVersion string             `json:"schemaVersion"`
	OSVersion     string             `json:"osVersion"`
	GeneratedAt   string             `json:"generatedAt,omitempty"`
	Apps          []DeclarationAppV2 `json:"apps"`
}

// DeclarationAppV2 is one app the OS declares. The fields below ChartSource
// belong to a local entry alone: a catalog entry names an app and nothing else,
// because everything else about it is the catalog's to say.
type DeclarationAppV2 struct {
	AppID        string       `json:"appId"`
	AppName      string       `json:"appName"`
	InstallScope InstallScope `json:"installScope"`
	InstallOrder int          `json:"installOrder"`
	ChartSource  ChartSource  `json:"chartSource"`

	Version         string                `json:"version,omitempty"`
	Chart           string                `json:"chart,omitempty"`
	ChartSHA256     string                `json:"chartSha256,omitempty"`
	SelectedGPUType string                `json:"selectedGpuType,omitempty"`
	Envs            map[string]string     `json:"envs,omitempty"`
	Artifacts       []DeclarationArtifact `json:"artifacts,omitempty"`
	AppEntry        json.RawMessage       `json:"appEntry,omitempty"`
}

func (a DeclarationAppV2) local() bool { return a.ChartSource == ChartSourceLocal }

// DeclarationArtifact is the same shape as BundleArtifactV1 and is kept separate
// so the input format on the medium can change without changing what Market
// reads.
type DeclarationArtifact struct {
	Kind           string `json:"kind"`
	Source         string `json:"source"`
	Repo           string `json:"repo"`
	Revision       string `json:"revision"`
	Manifest       string `json:"manifest"`
	ManifestSHA256 string `json:"manifestSha256"`
	TotalSize      int64  `json:"totalSize"`
}

// TrunkVersion is the part of an Olares version that decides which declaration
// applies. Everything after the first hyphen is a build of the same release -- a
// daily, an alpha, a release candidate -- and those share one declaration.
func TrunkVersion(version string) string {
	trunk := strings.TrimSpace(version)
	if index := strings.IndexByte(trunk, '-'); index >= 0 {
		trunk = trunk[:index]
	}
	return trunk
}

// DeclarationFileName is the contract with Market: one file per trunk, named for
// it, and looked for under that name and no other.
func DeclarationFileName(trunk string) string {
	return declarationFilePrefix + TrunkVersion(trunk) + declarationFileSuffix
}

// BuildDeclaration folds the medium's bundle and the installer's choices into
// the one file Market reads. Selections are validated against what the bundle
// allows here rather than published beside it: a choice the bundle never offered
// is a mistake in this program, and the device it would break is this one.
func BuildDeclaration(
	osVersion string, bundle BundleV1, selections ProfileSelections, catalog []DeclarationAppV2,
) (DeclarationV2, error) {
	declaration := DeclarationV2{
		SchemaVersion: DeclarationSchemaVersion,
		OSVersion:     strings.TrimSpace(osVersion),
		GeneratedAt:   generatedNow(),
	}
	for _, app := range bundle.Apps {
		entry, err := localDeclarationApp(app, selections)
		if err != nil {
			return DeclarationV2{}, err
		}
		declaration.Apps = append(declaration.Apps, entry)
	}
	declaration.Apps = append(declaration.Apps, mergeCatalogApps(declaration.Apps, catalog)...)
	if err := ValidateDeclaration(&declaration); err != nil {
		return DeclarationV2{}, err
	}
	return declaration, nil
}

// mergeCatalogApps adds the apps every device of this version is expected to
// have, minus any the medium already carries a chart for. An air-gapped install
// has no catalog to fetch from, so a bundled entry wins: it is the only one of
// the two that can be installed without a network.
func mergeCatalogApps(declared []DeclarationAppV2, catalog []DeclarationAppV2) []DeclarationAppV2 {
	taken := make(map[string]struct{}, len(declared)*2)
	for _, app := range declared {
		taken[app.AppID] = struct{}{}
		taken[app.AppName] = struct{}{}
	}
	merged := make([]DeclarationAppV2, 0, len(catalog))
	for _, app := range catalog {
		if _, exists := taken[app.AppID]; exists {
			continue
		}
		if _, exists := taken[app.AppName]; exists {
			continue
		}
		app.ChartSource = ChartSourceCatalog
		merged = append(merged, app)
	}
	sort.SliceStable(merged, func(i, j int) bool { return merged[i].AppID < merged[j].AppID })
	return merged
}

func localDeclarationApp(app BundleAppV1, selections ProfileSelections) (DeclarationAppV2, error) {
	selection := selections.Apps[app.AppID]
	allowedEnvs := stringSet(app.AllowedEnvs)
	envs := cloneStringMap(app.DefaultEnvs)
	for key, value := range selection.Envs {
		if _, allowed := allowedEnvs[key]; !allowed {
			return DeclarationAppV2{}, fmt.Errorf("app %q env %q is not allowed", app.AppID, key)
		}
		if envs == nil {
			envs = make(map[string]string, len(selection.Envs))
		}
		envs[key] = value
	}
	for key := range envs {
		if _, allowed := allowedEnvs[key]; !allowed {
			return DeclarationAppV2{}, fmt.Errorf("app %q env %q is not allowed", app.AppID, key)
		}
	}
	selectedGPUType := selection.SelectedGPUType
	if selectedGPUType == "" && containsString(app.AllowedGPUTypes, selections.DetectedGPUType) {
		selectedGPUType = selections.DetectedGPUType
	}
	if selectedGPUType != "" && !containsString(app.AllowedGPUTypes, selectedGPUType) {
		return DeclarationAppV2{}, fmt.Errorf(
			"app %q selectedGpuType %q is not allowed", app.AppID, selectedGPUType)
	}
	entry := DeclarationAppV2{
		AppID:           app.AppID,
		AppName:         app.AppName,
		InstallScope:    app.InstallScope,
		InstallOrder:    app.InstallOrder,
		ChartSource:     ChartSourceLocal,
		Version:         app.Version,
		Chart:           app.Chart,
		ChartSHA256:     app.ChartSHA256,
		SelectedGPUType: selectedGPUType,
		Envs:            envs,
		AppEntry:        app.AppEntry,
	}
	for _, artifact := range app.Artifacts {
		entry.Artifacts = append(entry.Artifacts, DeclarationArtifact(artifact))
	}
	return entry, nil
}

// ValidateDeclaration mirrors what Market refuses to read. Checking it here is
// what turns a disagreement between the two programs into a failed install on a
// machine that can still say why, rather than a Market that will not start.
func ValidateDeclaration(declaration *DeclarationV2) error {
	if declaration.SchemaVersion != DeclarationSchemaVersion {
		return fmt.Errorf("declaration schemaVersion %q is not supported", declaration.SchemaVersion)
	}
	if strings.TrimSpace(declaration.OSVersion) == "" {
		return fmt.Errorf("declaration osVersion is required")
	}
	if generated := strings.TrimSpace(declaration.GeneratedAt); generated != "" {
		if _, err := time.Parse(time.RFC3339, generated); err != nil {
			return fmt.Errorf("declaration generatedAt must be RFC3339: %w", err)
		}
	}
	if len(declaration.Apps) > MaxDeclarationApps {
		return fmt.Errorf("declaration apps must contain at most %d entries", MaxDeclarationApps)
	}
	appIDs := make(map[string]struct{}, len(declaration.Apps))
	appNames := make(map[string]struct{}, len(declaration.Apps))
	payloadPaths := make([]string, 0, len(declaration.Apps)*3)
	for index := range declaration.Apps {
		app := &declaration.Apps[index]
		prefix := fmt.Sprintf("declaration apps[%d]", index)
		if strings.TrimSpace(app.AppID) == "" || strings.TrimSpace(app.AppName) == "" {
			return fmt.Errorf("%s appId and appName are required", prefix)
		}
		if _, exists := appIDs[app.AppID]; exists {
			return fmt.Errorf("declaration duplicate appId %q", app.AppID)
		}
		if _, exists := appNames[app.AppName]; exists {
			return fmt.Errorf("declaration duplicate appName %q", app.AppName)
		}
		appIDs[app.AppID] = struct{}{}
		appNames[app.AppName] = struct{}{}
		if !app.InstallScope.Valid() {
			return fmt.Errorf("%s installScope must be %q or %q",
				prefix, InstallScopeShared, InstallScopePerUser)
		}
		if !app.ChartSource.Valid() {
			return fmt.Errorf("%s chartSource must be %q or %q",
				prefix, ChartSourceCatalog, ChartSourceLocal)
		}
		if app.InstallOrder < 0 {
			return fmt.Errorf("%s installOrder must not be negative", prefix)
		}
		paths, err := validateDeclarationApp(prefix, app)
		if err != nil {
			return err
		}
		for _, candidate := range paths {
			for _, existing := range payloadPaths {
				if pathsOverlap(candidate, existing) {
					return fmt.Errorf("declaration payload %q overlaps %q", candidate, existing)
				}
			}
		}
		payloadPaths = append(payloadPaths, paths...)
	}
	return nil
}

// validateDeclarationApp checks one entry and returns the payload paths it
// claims, so the caller can refuse two entries that name the same file.
func validateDeclarationApp(prefix string, app *DeclarationAppV2) ([]string, error) {
	if !app.local() {
		// A catalog entry that carried a version, a chart or an app entry would
		// be describing something only the catalog can say. Refused rather than
		// dropped, because a reader that drops it installs something other than
		// what whoever wrote it meant.
		switch {
		case strings.TrimSpace(app.Version) != "":
			return nil, fmt.Errorf("%s catalog entry must not carry version", prefix)
		case strings.TrimSpace(app.Chart) != "" || strings.TrimSpace(app.ChartSHA256) != "":
			return nil, fmt.Errorf("%s catalog entry must not carry a chart", prefix)
		case len(app.Artifacts) > 0:
			return nil, fmt.Errorf("%s catalog entry must not carry artifacts", prefix)
		case len(app.AppEntry) > 0:
			return nil, fmt.Errorf("%s catalog entry must not carry appEntry", prefix)
		case len(app.Envs) > 0 || strings.TrimSpace(app.SelectedGPUType) != "":
			return nil, fmt.Errorf("%s catalog entry must not carry envs or selectedGpuType", prefix)
		}
		return nil, nil
	}
	if strings.TrimSpace(app.Version) == "" {
		return nil, fmt.Errorf("%s local entry requires version", prefix)
	}
	if err := validateChartPath(app.Chart); err != nil {
		return nil, fmt.Errorf("%s chart: %w", prefix, err)
	}
	if !sha256Pattern.MatchString(app.ChartSHA256) {
		return nil, fmt.Errorf("%s chartSha256 must be 64 hexadecimal characters", prefix)
	}
	if !jsonObject(app.AppEntry) {
		return nil, fmt.Errorf("%s appEntry must be a JSON object", prefix)
	}
	for key := range app.Envs {
		if strings.TrimSpace(key) == "" {
			return nil, fmt.Errorf("%s envs key is required", prefix)
		}
		if sensitiveEnvKey(key) {
			return nil, fmt.Errorf("%s envs %q is sensitive", prefix, key)
		}
	}
	if len(app.Artifacts) > 1 {
		return nil, fmt.Errorf("%s artifacts must contain at most one entry", prefix)
	}
	paths := []string{app.Chart}
	for position, artifact := range app.Artifacts {
		artifactPrefix := fmt.Sprintf("%s artifacts[%d]", prefix, position)
		if err := validateBundleArtifact(BundleArtifactV1(artifact)); err != nil {
			return nil, fmt.Errorf("%s: %w", artifactPrefix, err)
		}
		if pathsOverlap(artifact.Source, artifact.Manifest) {
			return nil, fmt.Errorf("%s source %q overlaps manifest %q",
				artifactPrefix, artifact.Source, artifact.Manifest)
		}
		paths = append(paths, artifact.Source, artifact.Manifest)
	}
	return paths, nil
}
