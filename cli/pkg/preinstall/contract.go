package preinstall

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

var (
	sha256Pattern   = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)
	revisionPattern = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)
	repoPattern     = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*/[A-Za-z0-9][A-Za-z0-9._-]*$`)
)

func DecodeBundle(data []byte) (BundleV1, error) {
	var bundle BundleV1
	if len(data) > MaxBundleJSONBytes {
		return bundle, fmt.Errorf("%s exceeds %d bytes", BundleFileName, MaxBundleJSONBytes)
	}
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return bundle, fmt.Errorf("decode bundle: %w", err)
	}
	schema, err := schemaVersion(data)
	if err != nil {
		return bundle, fmt.Errorf("decode bundle schema: %w", err)
	}
	if schema != SupportedSchemaVersion {
		return bundle, fmt.Errorf("bundle schemaVersion %q is not supported", schema)
	}
	if err := strictDecode(data, &bundle); err != nil {
		return bundle, fmt.Errorf("decode bundle: %w", err)
	}
	return bundle, nil
}

func LoadDirectory(dir string) (*ContractV1, error) {
	root, err := openDirectoryNoSymlink(dir)
	if err != nil {
		return nil, fmt.Errorf("open contract directory: %w", err)
	}
	defer root.Close()
	bundleData, err := readRootFileLimited(root, BundleFileName, MaxBundleJSONBytes)
	if err != nil {
		return nil, err
	}
	bundle, err := DecodeBundle(bundleData)
	if err != nil {
		return nil, err
	}
	profileData, err := readRootFileLimited(root, ProfileFileName, MaxProfileJSONBytes)
	if err != nil {
		return nil, err
	}
	if err := rejectDuplicateJSONKeys(profileData); err != nil {
		return nil, fmt.Errorf("decode profile: %w", err)
	}
	schema, err := schemaVersion(profileData)
	if err != nil {
		return nil, fmt.Errorf("decode profile schema: %w", err)
	}
	if schema != SupportedSchemaVersion {
		return nil, fmt.Errorf("profile schemaVersion %q is not supported", schema)
	}
	var profile InstallProfileV1
	if err := strictDecode(profileData, &profile); err != nil {
		return nil, fmt.Errorf("decode profile: %w", err)
	}
	if err := Validate(bundle, profile); err != nil {
		return nil, err
	}
	for _, app := range bundle.Apps {
		for _, artifact := range app.Artifacts {
			if _, err := LoadArtifactManifest(root, artifact); err != nil {
				return nil, fmt.Errorf("bundle app %q artifact manifest: %w", app.AppID, err)
			}
		}
	}
	return &ContractV1{Bundle: bundle, Profile: profile}, nil
}

func Validate(bundle BundleV1, profile InstallProfileV1) error {
	apps, err := validateBundle(bundle)
	if err != nil {
		return err
	}
	return validateProfile(profile, apps)
}

func validateBundle(bundle BundleV1) (map[string]BundleAppV1, error) {
	if bundle.SchemaVersion != SupportedSchemaVersion {
		return nil, fmt.Errorf("bundle schemaVersion %q is not supported", bundle.SchemaVersion)
	}
	if bundle.SourceID != OfficialSourceID {
		return nil, fmt.Errorf("bundle sourceId must be %q", OfficialSourceID)
	}
	if strings.TrimSpace(bundle.CatalogHash) == "" {
		return nil, fmt.Errorf("bundle catalogHash is required")
	}
	if _, err := time.Parse(time.RFC3339, bundle.GeneratedAt); err != nil {
		return nil, fmt.Errorf("bundle generatedAt must be RFC3339: %w", err)
	}
	if len(bundle.Apps) == 0 {
		return nil, fmt.Errorf("bundle apps must not be empty")
	}
	if len(bundle.Apps) > MaxBundleApps {
		return nil, fmt.Errorf("bundle apps must contain at most %d entries", MaxBundleApps)
	}

	apps := make(map[string]BundleAppV1, len(bundle.Apps))
	appNames := make(map[string]struct{}, len(bundle.Apps))
	charts := make(map[string]struct{}, len(bundle.Apps))
	chartPaths := make([]string, 0, len(bundle.Apps))
	type artifactPath struct {
		kind string
		path string
	}
	artifactPaths := make([]artifactPath, 0, len(bundle.Apps)*2)
	for i, app := range bundle.Apps {
		prefix := fmt.Sprintf("bundle apps[%d]", i)
		if strings.TrimSpace(app.AppID) == "" {
			return nil, fmt.Errorf("%s appId is required", prefix)
		}
		if _, exists := apps[app.AppID]; exists {
			return nil, fmt.Errorf("bundle duplicate appId %q", app.AppID)
		}
		if strings.TrimSpace(app.AppName) == "" {
			return nil, fmt.Errorf("%s appName is required", prefix)
		}
		if _, exists := appNames[app.AppName]; exists {
			return nil, fmt.Errorf("bundle duplicate appName %q", app.AppName)
		}
		if strings.TrimSpace(app.Version) == "" {
			return nil, fmt.Errorf("%s version is required", prefix)
		}
		if !app.InstallScope.Valid() {
			return nil, fmt.Errorf("%s installScope must be %q or %q", prefix, InstallScopeShared, InstallScopePerUser)
		}
		if err := validateChartPath(app.Chart); err != nil {
			return nil, fmt.Errorf("%s chart: %w", prefix, err)
		}
		if _, exists := charts[app.Chart]; exists {
			return nil, fmt.Errorf("bundle duplicate chart %q", app.Chart)
		}
		for _, existing := range chartPaths {
			if pathsOverlap(app.Chart, existing) {
				return nil, fmt.Errorf("bundle chart %q overlaps chart %q", app.Chart, existing)
			}
		}
		if !sha256Pattern.MatchString(app.ChartSHA256) {
			return nil, fmt.Errorf("%s chartSha256 must be 64 hexadecimal characters", prefix)
		}
		if !jsonObject(app.AppEntry) {
			return nil, fmt.Errorf("%s appEntry must be a JSON object", prefix)
		}
		if duplicate := duplicateString(app.AllowedEnvs); duplicate != "" {
			return nil, fmt.Errorf("%s allowedEnvs contains duplicate %q", prefix, duplicate)
		}
		allowedEnvs := stringSet(app.AllowedEnvs)
		for key := range app.DefaultEnvs {
			if strings.TrimSpace(key) == "" {
				return nil, fmt.Errorf("%s defaultEnvs key is required", prefix)
			}
			if sensitiveEnvKey(key) {
				return nil, fmt.Errorf("%s defaultEnvs env %q is sensitive", prefix, key)
			}
			if _, ok := allowedEnvs[key]; !ok {
				return nil, fmt.Errorf("%s defaultEnvs env %q is not allowed", prefix, key)
			}
		}
		if duplicate := duplicateString(app.AllowedGPUTypes); duplicate != "" {
			return nil, fmt.Errorf("%s allowedGpuTypes contains duplicate %q", prefix, duplicate)
		}
		if len(app.Artifacts) > 1 {
			return nil, fmt.Errorf("%s artifacts must contain at most one entry", prefix)
		}
		for j, artifact := range app.Artifacts {
			artifactPrefix := fmt.Sprintf("%s artifacts[%d]", prefix, j)
			for _, chartApp := range bundle.Apps {
				if pathsOverlap(artifact.Source, chartApp.Chart) {
					return nil, fmt.Errorf("%s source conflicts with chart %q", artifactPrefix, chartApp.Chart)
				}
				if pathsOverlap(artifact.Manifest, chartApp.Chart) {
					return nil, fmt.Errorf("%s manifest conflicts with chart %q", artifactPrefix, chartApp.Chart)
				}
			}
			if err := validateBundleArtifact(artifact); err != nil {
				return nil, fmt.Errorf("%s: %w", artifactPrefix, err)
			}
			if pathsOverlap(artifact.Source, artifact.Manifest) {
				return nil, fmt.Errorf("%s source %q overlaps artifact manifest %q", artifactPrefix, artifact.Source, artifact.Manifest)
			}
			for _, existing := range artifactPaths {
				for _, current := range []artifactPath{
					{kind: "source", path: artifact.Source},
					{kind: "manifest", path: artifact.Manifest},
				} {
					if pathsOverlap(current.path, existing.path) {
						return nil, fmt.Errorf("%s %s %q overlaps artifact %s %q", artifactPrefix, current.kind, current.path, existing.kind, existing.path)
					}
				}
			}
			artifactPaths = append(artifactPaths,
				artifactPath{kind: "source", path: artifact.Source},
				artifactPath{kind: "manifest", path: artifact.Manifest},
			)
		}
		apps[app.AppID] = app
		appNames[app.AppName] = struct{}{}
		charts[app.Chart] = struct{}{}
		chartPaths = append(chartPaths, app.Chart)
	}
	return apps, nil
}

func LoadArtifactManifest(root *os.Root, artifact BundleArtifactV1) (ArtifactManifestV1, error) {
	var manifest ArtifactManifestV1
	if err := validateBundleArtifact(artifact); err != nil {
		return manifest, fmt.Errorf("artifact declaration: %w", err)
	}
	data, err := readRootFileLimited(root, artifact.Manifest, MaxArtifactManifestBytes)
	if err != nil {
		return manifest, err
	}
	sum := sha256.Sum256(data)
	if fmt.Sprintf("%x", sum) != strings.ToLower(artifact.ManifestSHA256) {
		return manifest, fmt.Errorf("manifestSha256 does not match %q", artifact.Manifest)
	}
	if err := strictDecode(data, &manifest); err != nil {
		return manifest, fmt.Errorf("decode artifact manifest: %w", err)
	}
	if err := validateArtifactManifest(manifest, artifact); err != nil {
		return manifest, err
	}
	return manifest, nil
}

func validateBundleArtifact(artifact BundleArtifactV1) error {
	if artifact.Kind != ArtifactKindHFCache {
		return fmt.Errorf("kind must be %q", ArtifactKindHFCache)
	}
	if err := validateRelativePath(artifact.Source); err != nil {
		return fmt.Errorf("source: %w", err)
	}
	if !repoPattern.MatchString(artifact.Repo) {
		return fmt.Errorf("repo must be an owner/repo identifier")
	}
	if !revisionPattern.MatchString(artifact.Revision) {
		return fmt.Errorf("revision must be a 40 hexadecimal commit")
	}
	if err := validateRelativePath(artifact.Manifest); err != nil {
		return fmt.Errorf("manifest: %w", err)
	}
	if path.Ext(artifact.Manifest) != ".json" {
		return fmt.Errorf("manifest must have .json extension")
	}
	if !sha256Pattern.MatchString(artifact.ManifestSHA256) {
		return fmt.Errorf("manifestSha256 must be 64 hexadecimal characters")
	}
	if artifact.TotalSize <= 0 || artifact.TotalSize > MaxArtifactTotalSize {
		return fmt.Errorf("totalSize must be between 1 and %d", MaxArtifactTotalSize)
	}
	return nil
}

func validateArtifactManifest(manifest ArtifactManifestV1, artifact BundleArtifactV1) error {
	if manifest.SchemaVersion != 1 {
		return fmt.Errorf("artifact manifest schemaVersion must be 1")
	}
	if manifest.Repo != artifact.Repo {
		return fmt.Errorf("artifact manifest repo must match artifact repo")
	}
	if manifest.Revision != artifact.Revision {
		return fmt.Errorf("artifact manifest revision must match artifact revision")
	}
	if len(manifest.Entries) == 0 {
		return fmt.Errorf("artifact manifest entries must not be empty")
	}
	if len(manifest.Entries) > MaxArtifactEntries {
		return fmt.Errorf("artifact manifest entries must contain at most %d entries", MaxArtifactEntries)
	}
	seen := make(map[string]struct{}, len(manifest.Entries))
	var total int64
	for i, entry := range manifest.Entries {
		prefix := fmt.Sprintf("artifact manifest entries[%d]", i)
		if err := validateRelativePath(entry.Path); err != nil {
			return fmt.Errorf("%s path: %w", prefix, err)
		}
		if _, exists := seen[entry.Path]; exists {
			return fmt.Errorf("artifact manifest duplicate path %q", entry.Path)
		}
		seen[entry.Path] = struct{}{}
		switch entry.Type {
		case "file":
			if entry.Size < 0 {
				return fmt.Errorf("%s file size must be non-negative", prefix)
			}
			if !sha256Pattern.MatchString(entry.SHA256) {
				return fmt.Errorf("%s file sha256 must be 64 hexadecimal characters", prefix)
			}
			if entry.Target != "" {
				return fmt.Errorf("%s file target must be empty", prefix)
			}
			if entry.Size > MaxArtifactTotalSize-total {
				return fmt.Errorf("artifact manifest file total exceeds %d", MaxArtifactTotalSize)
			}
			total += entry.Size
		case "directory":
			if entry.Size != 0 || entry.SHA256 != "" || entry.Target != "" {
				return fmt.Errorf("%s directory must have zero size and no sha256 or target", prefix)
			}
		case "symlink":
			if entry.Size != 0 || entry.SHA256 != "" {
				return fmt.Errorf("%s symlink must have zero size and no sha256", prefix)
			}
			if err := validateSymlinkTarget(entry.Path, entry.Target); err != nil {
				return fmt.Errorf("%s target: %w", prefix, err)
			}
		default:
			return fmt.Errorf("%s type must be directory, file, or symlink", prefix)
		}
	}
	if total != artifact.TotalSize {
		return fmt.Errorf("artifact manifest file total %d does not match totalSize %d", total, artifact.TotalSize)
	}
	return nil
}

func validateProfile(profile InstallProfileV1, bundleApps map[string]BundleAppV1) error {
	if profile.SchemaVersion != SupportedSchemaVersion {
		return fmt.Errorf("profile schemaVersion %q is not supported", profile.SchemaVersion)
	}
	seen := make(map[string]struct{}, len(profile.Apps))
	for i, app := range profile.Apps {
		prefix := fmt.Sprintf("profile apps[%d]", i)
		if strings.TrimSpace(app.AppID) == "" {
			return fmt.Errorf("%s appId is required", prefix)
		}
		if _, exists := seen[app.AppID]; exists {
			return fmt.Errorf("profile duplicate appId %q", app.AppID)
		}
		seen[app.AppID] = struct{}{}
		bundleApp, exists := bundleApps[app.AppID]
		if !exists {
			return fmt.Errorf("%s appId %q is not present in bundle", prefix, app.AppID)
		}
		allowedEnvs := stringSet(bundleApp.AllowedEnvs)
		for key := range app.Envs {
			if sensitiveEnvKey(key) {
				return fmt.Errorf("%s env %q is sensitive", prefix, key)
			}
			if _, ok := allowedEnvs[key]; !ok {
				return fmt.Errorf("%s env %q is not allowed", prefix, key)
			}
		}
		if app.SelectedGPUType != "" {
			if _, ok := stringSet(bundleApp.AllowedGPUTypes)[app.SelectedGPUType]; !ok {
				return fmt.Errorf("%s selectedGpuType %q is not allowed", prefix, app.SelectedGPUType)
			}
		}
	}
	return nil
}

func strictDecode(data []byte, target any) error {
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values")
		}
		return err
	}
	return nil
}

func rejectDuplicateJSONKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := consumeJSONValue(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values")
		}
		return err
	}
	return nil
}

func consumeJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("JSON object key is not a string")
			}
			if _, exists := seen[key]; exists {
				return fmt.Errorf("duplicate JSON key %q", key)
			}
			seen[key] = struct{}{}
			if err := consumeJSONValue(decoder); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil {
			return err
		}
		if closing != json.Delim('}') {
			return fmt.Errorf("invalid JSON object")
		}
	case '[':
		for decoder.More() {
			if err := consumeJSONValue(decoder); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil {
			return err
		}
		if closing != json.Delim(']') {
			return fmt.Errorf("invalid JSON array")
		}
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delim)
	}
	return nil
}

func schemaVersion(data []byte) (string, error) {
	var header struct {
		SchemaVersion string `json:"schemaVersion"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return "", err
	}
	return header.SchemaVersion, nil
}

func jsonObject(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	var object map[string]json.RawMessage
	return json.Unmarshal(raw, &object) == nil && object != nil
}

func duplicateString(values []string) string {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			return value
		}
		seen[value] = struct{}{}
	}
	return ""
}

func stringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
}

func sensitiveEnvKey(key string) bool {
	upper := strings.ToUpper(key)
	for _, marker := range []string{
		"SECRET", "PASSWORD", "TOKEN", "CREDENTIAL", "PRIVATE_KEY", "API_KEY",
		"ACCESS_KEY", "SECRET_KEY", "PASSPHRASE", "COOKIE", "AUTH",
	} {
		if strings.Contains(upper, marker) {
			return true
		}
	}
	return false
}

func validateChartPath(chart string) error {
	if chart == "" {
		return fmt.Errorf("path is required")
	}
	if strings.Contains(chart, `\`) {
		return fmt.Errorf("path must use forward slashes")
	}
	if path.IsAbs(chart) || path.Clean(chart) != chart {
		return fmt.Errorf("path must be clean and relative")
	}
	if !strings.HasPrefix(chart, "chart/") && !strings.HasPrefix(chart, "charts/") {
		return fmt.Errorf("path must be under chart/ or charts/")
	}
	if path.Ext(chart) != ".tgz" {
		return fmt.Errorf("path must have .tgz extension")
	}
	return nil
}

func validateRelativePath(value string) error {
	if value == "" {
		return fmt.Errorf("path is required")
	}
	if strings.Contains(value, `\`) {
		return fmt.Errorf("path must use forward slashes")
	}
	if value == "." || path.IsAbs(value) || path.Clean(value) != value || strings.HasPrefix(value, "../") {
		return fmt.Errorf("path must be clean and relative")
	}
	return nil
}

func validateSymlinkTarget(entryPath, target string) error {
	if target == "" {
		return fmt.Errorf("path is required")
	}
	if strings.Contains(target, `\`) || path.IsAbs(target) || path.Clean(target) != target {
		return fmt.Errorf("path must be clean and relative")
	}
	resolved := path.Clean(path.Join(path.Dir(entryPath), target))
	if resolved == ".." || strings.HasPrefix(resolved, "../") {
		return fmt.Errorf("path escapes artifact root")
	}
	return nil
}

func pathsOverlap(left, right string) bool {
	return left == right || strings.HasPrefix(left, right+"/") || strings.HasPrefix(right, left+"/")
}
