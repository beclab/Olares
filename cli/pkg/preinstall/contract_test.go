package preinstall

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDecodeBundleV1Strictly(t *testing.T) {
	bundle, err := DecodeBundle([]byte(validBundleJSON))
	if err != nil {
		t.Fatalf("DecodeBundle() error = %v", err)
	}
	if bundle.SourceID != OfficialSourceID || bundle.Apps[0].AllowedGPUTypes[0] != "nvidia" {
		t.Fatalf("DecodeBundle() = %#v", bundle)
	}

	_, err = DecodeBundle([]byte(strings.Replace(validBundleJSON, `"sourceId"`, `"unknown":true,"sourceId"`, 1)))
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("DecodeBundle() unknown field error = %v", err)
	}
}

func TestDecodeBundleAcceptsDefaultEnvs(t *testing.T) {
	raw := strings.Replace(validBundleJSON, `"allowedEnvs":["WORKER_COUNT"]`, `"allowedEnvs":["HF_HUB_OFFLINE","MODEL_PATH","MODEL_REVISION"],"defaultEnvs":{"HF_HUB_OFFLINE":"1","MODEL_PATH":"/models/default","MODEL_REVISION":""}`, 1)

	bundle, err := DecodeBundle([]byte(raw))

	if err != nil {
		t.Fatalf("DecodeBundle() error = %v", err)
	}
	if got := bundle.Apps[0].DefaultEnvs; len(got) != 3 || got["HF_HUB_OFFLINE"] != "1" || got["MODEL_REVISION"] != "" {
		t.Fatalf("DecodeBundle() defaultEnvs = %#v", got)
	}
}

func TestValidateInstallScope(t *testing.T) {
	tests := []struct {
		name         string
		installScope string
		wantErr      bool
	}{
		{name: "shared", installScope: "shared"},
		{name: "per-user", installScope: "per-user"},
		{name: "missing", wantErr: true},
		{name: "unknown", installScope: "cluster", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := validBundleJSON
			if tt.installScope == "" {
				raw = strings.Replace(raw, `    "installScope":"shared",`+"\n", "", 1)
			} else {
				raw = strings.Replace(raw, `"installScope":"shared"`, `"installScope":"`+tt.installScope+`"`, 1)
			}
			bundle := decodeBundle(t, raw)

			err := Validate(bundle, InstallProfileV1{SchemaVersion: SupportedSchemaVersion})

			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), "installScope") {
					t.Fatalf("Validate() error = %v, want installScope rejection", err)
				}
			} else if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}

func TestValidateRejectsInvalidDefaultEnvs(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr string
	}{
		{name: "outside allowlist", key: "OTHER", wantErr: "not allowed"},
		{name: "blank", key: " ", wantErr: "required"},
		{name: "sensitive", key: "API_TOKEN", wantErr: "sensitive"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := decodeBundle(t, validBundleJSON)
			bundle.Apps[0].AllowedEnvs = append(bundle.Apps[0].AllowedEnvs, tt.key)
			if tt.name == "outside allowlist" {
				bundle.Apps[0].AllowedEnvs = bundle.Apps[0].AllowedEnvs[:1]
			}
			bundle.Apps[0].DefaultEnvs = map[string]string{tt.key: ""}

			err := Validate(bundle, InstallProfileV1{SchemaVersion: SupportedSchemaVersion})

			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestDecodeBundleRejectsDuplicateJSONKeysRecursively(t *testing.T) {
	tests := []string{
		strings.Replace(validBundleJSON, `"sourceId":"market.olares"`, `"sourceId":"market.olares","sourceId":"upload"`, 1),
		strings.Replace(validBundleJSON, `"appEntry":{}`, `"appEntry":{"nested":{"name":"a","name":"b"}}`, 1),
	}
	for _, raw := range tests {
		_, err := DecodeBundle([]byte(raw))
		if err == nil || !strings.Contains(err.Error(), "duplicate JSON key") {
			t.Fatalf("DecodeBundle() duplicate key error = %v", err)
		}
	}
}

func TestLoadDirectoryRejectsDuplicateProfileKeys(t *testing.T) {
	dir := canonicalTempDir(t)
	if err := os.WriteFile(filepath.Join(dir, BundleFileName), []byte(validBundleJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	profile := `{"schemaVersion":"1","apps":[],"apps":[]}`
	if err := os.WriteFile(filepath.Join(dir, ProfileFileName), []byte(profile), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadDirectory(dir)

	if err == nil || !strings.Contains(err.Error(), "duplicate JSON key") {
		t.Fatalf("LoadDirectory() error = %v", err)
	}
}

func TestLoadDirectoryRejectsOversizedProfile(t *testing.T) {
	dir := writeContractDirectory(t)
	if err := os.Truncate(filepath.Join(dir, ProfileFileName), MaxProfileJSONBytes+1); err != nil {
		t.Fatal(err)
	}

	_, err := LoadDirectory(dir)

	if err == nil || !strings.Contains(err.Error(), "install-profile.json exceeds") {
		t.Fatalf("LoadDirectory() error = %v", err)
	}
}

func TestLoadDirectoryRejectsSymlinks(t *testing.T) {
	t.Run("bundle leaf", func(t *testing.T) {
		dir := writeContractDirectory(t)
		bundle := filepath.Join(dir, BundleFileName)
		outside := filepath.Join(t.TempDir(), BundleFileName)
		if err := os.Rename(bundle, outside); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, bundle); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadDirectory(dir); err == nil || !strings.Contains(err.Error(), "symlink") {
			t.Fatalf("LoadDirectory() error = %v", err)
		}
	})

	t.Run("profile leaf", func(t *testing.T) {
		dir := writeContractDirectory(t)
		profile := filepath.Join(dir, ProfileFileName)
		outside := filepath.Join(t.TempDir(), ProfileFileName)
		if err := os.Rename(profile, outside); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, profile); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadDirectory(dir); err == nil || !strings.Contains(err.Error(), "symlink") {
			t.Fatalf("LoadDirectory() error = %v", err)
		}
	})

	t.Run("ancestor is allowed", func(t *testing.T) {
		dir := writeContractDirectory(t)
		root := canonicalTempDir(t)
		link := filepath.Join(root, "linked-parent")
		if err := os.Symlink(filepath.Dir(dir), link); err != nil {
			t.Fatal(err)
		}
		linkedDir := filepath.Join(link, filepath.Base(dir))
		if _, err := LoadDirectory(linkedDir); err != nil {
			t.Fatalf("LoadDirectory() through symlinked ancestor error = %v", err)
		}
	})
}

func TestValidateRejectsTooManyApps(t *testing.T) {
	bundle := decodeBundle(t, validBundleJSON)
	bundle.Apps = make([]BundleAppV1, MaxBundleApps+1)

	err := Validate(bundle, InstallProfileV1{SchemaVersion: SupportedSchemaVersion})

	if err == nil || !strings.Contains(err.Error(), "at most") {
		t.Fatalf("Validate() error = %v", err)
	}
}

// path.Clean("..") is ".." and does not start with "../", so the parent
// directory itself used to pass the only traversal gate the bundle contract
// has. Artifact entry paths are written under the model root by the installer.
func TestValidateRelativePathRejectsParentTraversal(t *testing.T) {
	for _, value := range []string{"", ".", "..", "../", "../escape", "/absolute", "a/../b", `a\b`} {
		if err := validateRelativePath(value); err == nil {
			t.Errorf("validateRelativePath(%q) = nil, want an error", value)
		}
	}
	if err := validateRelativePath("artifacts/models/owner--repo"); err != nil {
		t.Fatalf("validateRelativePath() rejected a clean relative path: %v", err)
	}
}

func TestValidateRejectsOverlappingCharts(t *testing.T) {
	bundle := decodeBundle(t, validBundleJSON)
	second := bundle.Apps[0]
	second.AppID = "app-b"
	second.AppName = "app-b"
	second.Chart = bundle.Apps[0].Chart + "/nested.tgz"
	bundle.Apps = append(bundle.Apps, second)

	err := Validate(bundle, InstallProfileV1{SchemaVersion: SupportedSchemaVersion})

	if err == nil || !strings.Contains(err.Error(), "overlaps chart") {
		t.Fatalf("Validate() overlapping chart error = %v", err)
	}
}

func TestValidateRejectsInvalidBundleAndProfile(t *testing.T) {
	bundle := decodeBundle(t, validBundleJSON)
	profile := InstallProfileV1{
		SchemaVersion:   SupportedSchemaVersion,
		HardwareProfile: "nvidia-cuda",
		Apps: []InstallProfileAppV1{{
			AppID:           "app-a",
			SelectedGPUType: "nvidia",
			Envs:            map[string]string{"WORKER_COUNT": "1"},
		}},
	}
	if err := Validate(bundle, profile); err != nil {
		t.Fatalf("Validate() valid contract error = %v", err)
	}

	tests := []struct {
		name    string
		mutate  func(*BundleV1, *InstallProfileV1)
		wantErr string
	}{
		{"schema", func(b *BundleV1, _ *InstallProfileV1) { b.SchemaVersion = "2" }, "schemaVersion"},
		{"source", func(b *BundleV1, _ *InstallProfileV1) { b.SourceID = "upload" }, "sourceId"},
		{"duplicate app", func(b *BundleV1, _ *InstallProfileV1) { b.Apps = append(b.Apps, b.Apps[0]) }, "duplicate appId"},
		{"absolute chart", func(b *BundleV1, _ *InstallProfileV1) { b.Apps[0].Chart = "/charts/app.tgz" }, "chart"},
		{"short digest", func(b *BundleV1, _ *InstallProfileV1) { b.Apps[0].ChartSHA256 = "abc" }, "chartSha256"},
		{"unknown app", func(_ *BundleV1, p *InstallProfileV1) { p.Apps[0].AppID = "other" }, "not present"},
		{"unknown env", func(_ *BundleV1, p *InstallProfileV1) { p.Apps[0].Envs["OTHER"] = "x" }, "not allowed"},
		{"sensitive env", func(b *BundleV1, p *InstallProfileV1) {
			b.Apps[0].AllowedEnvs = append(b.Apps[0].AllowedEnvs, "API_TOKEN")
			p.Apps[0].Envs["API_TOKEN"] = "x"
		}, "sensitive"},
		{"unknown gpu", func(_ *BundleV1, p *InstallProfileV1) { p.Apps[0].SelectedGPUType = "amd" }, "selectedGpuType"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := decodeBundle(t, validBundleJSON)
			p := profile
			p.Apps = append([]InstallProfileAppV1(nil), profile.Apps...)
			p.Apps[0].Envs = map[string]string{"WORKER_COUNT": "1"}
			tt.mutate(&b, &p)
			err := Validate(b, p)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestHFCacheArtifactContract(t *testing.T) {
	manifest := validArtifactManifestJSON()
	artifact := validHFCacheArtifact(manifest)
	bundle := decodeBundle(t, validBundleJSON)
	bundle.Apps[0].Artifacts = []BundleArtifactV1{artifact}
	profile := InstallProfileV1{SchemaVersion: SupportedSchemaVersion}
	if err := Validate(bundle, profile); err != nil {
		t.Fatalf("Validate() valid artifact error = %v", err)
	}

	tests := []struct {
		name    string
		mutate  func(*BundleV1)
		wantErr string
	}{
		{"multiple artifacts", func(b *BundleV1) {
			b.Apps[0].Artifacts = append(b.Apps[0].Artifacts, artifact)
		}, "at most one"},
		{"unknown kind", func(b *BundleV1) { b.Apps[0].Artifacts[0].Kind = "destination" }, "kind"},
		{"invalid source", func(b *BundleV1) { b.Apps[0].Artifacts[0].Source = "../models" }, "source"},
		{"invalid repo", func(b *BundleV1) { b.Apps[0].Artifacts[0].Repo = "model" }, "repo"},
		{"mutable revision", func(b *BundleV1) { b.Apps[0].Artifacts[0].Revision = "main" }, "revision"},
		{"invalid manifest", func(b *BundleV1) { b.Apps[0].Artifacts[0].Manifest = "/manifest.json" }, "manifest"},
		{"invalid manifest digest", func(b *BundleV1) { b.Apps[0].Artifacts[0].ManifestSHA256 = "abc" }, "manifestSha256"},
		{"zero size", func(b *BundleV1) { b.Apps[0].Artifacts[0].TotalSize = 0 }, "totalSize"},
		{"oversized", func(b *BundleV1) { b.Apps[0].Artifacts[0].TotalSize = MaxArtifactTotalSize + 1 }, "totalSize"},
		{"chart source conflict", func(b *BundleV1) { b.Apps[0].Artifacts[0].Source = "charts" }, "conflicts with chart"},
		{"chart manifest conflict", func(b *BundleV1) { b.Apps[0].Artifacts[0].Manifest = b.Apps[0].Chart }, "conflicts with chart"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := decodeBundle(t, mustJSON(t, bundle))
			tt.mutate(&candidate)
			err := Validate(candidate, profile)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}

	bundle.Apps = append(bundle.Apps, bundle.Apps[0])
	bundle.Apps[1].AppID = "app-b"
	bundle.Apps[1].AppName = "app-b"
	bundle.Apps[1].Chart = "charts/app-b-1.0.0.tgz"
	bundle.Apps[1].Artifacts[0].Source = "artifacts/models/subdir"
	if err := Validate(bundle, profile); err == nil || !strings.Contains(err.Error(), "overlaps") {
		t.Fatalf("Validate() overlapping artifact source error = %v", err)
	}

	for _, field := range []string{"source", "manifest"} {
		t.Run("cross app chart "+field, func(t *testing.T) {
			candidate := decodeBundle(t, mustJSON(t, decodeBundle(t, validBundleJSON)))
			second := candidate.Apps[0]
			second.AppID = "app-b"
			second.AppName = "app-b"
			second.Chart = "charts/app-b-1.0.0.tgz"
			second.Artifacts = []BundleArtifactV1{artifact}
			if field == "source" {
				second.Artifacts[0].Source = candidate.Apps[0].Chart
			} else {
				second.Artifacts[0].Manifest = candidate.Apps[0].Chart
			}
			candidate.Apps = append(candidate.Apps, second)
			err := Validate(candidate, profile)
			if err == nil || !strings.Contains(err.Error(), "conflicts with chart") {
				t.Fatalf("Validate() cross-app chart conflict error = %v", err)
			}
		})
	}
}

func TestValidateRejectsAllArtifactPathOverlaps(t *testing.T) {
	manifest := validArtifactManifestJSON()
	artifact := validHFCacheArtifact(manifest)
	tests := []struct {
		name   string
		mutate func(*BundleV1)
	}{
		{"same artifact source manifest", func(bundle *BundleV1) {
			bundle.Apps[0].Artifacts[0].Manifest = bundle.Apps[0].Artifacts[0].Source + "/manifest.json"
		}},
		{"cross artifact source manifest", func(bundle *BundleV1) {
			bundle.Apps[1].Artifacts[0].Manifest = bundle.Apps[0].Artifacts[0].Source + "/manifest.json"
		}},
		{"cross artifact manifest source", func(bundle *BundleV1) {
			bundle.Apps[1].Artifacts[0].Source = bundle.Apps[0].Artifacts[0].Manifest + "/payload"
		}},
		{"cross artifact manifests", func(bundle *BundleV1) {
			bundle.Apps[1].Artifacts[0].Manifest = bundle.Apps[0].Artifacts[0].Manifest + "/nested.json"
		}},
		{"duplicate artifact manifests", func(bundle *BundleV1) {
			bundle.Apps[1].Artifacts[0].Manifest = bundle.Apps[0].Artifacts[0].Manifest
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := twoArtifactBundle(t, artifact)
			tt.mutate(&bundle)
			err := Validate(bundle, InstallProfileV1{SchemaVersion: SupportedSchemaVersion})
			if err == nil || !strings.Contains(err.Error(), "overlaps") {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}

func TestLoadArtifactManifestValidatesWire(t *testing.T) {
	valid := validArtifactManifestJSON()
	artifact := validHFCacheArtifact(valid)
	tests := []struct {
		name    string
		raw     string
		mutate  func(*BundleArtifactV1)
		wantErr string
	}{
		{"valid", valid, nil, ""},
		{"unknown field", strings.Replace(valid, `"entries":`, `"destination":"/tmp","entries":`, 1), nil, "unknown field"},
		{"duplicate key", strings.Replace(valid, `"repo":`, `"repo":"other/repo","repo":`, 1), nil, "duplicate JSON key"},
		{"schema", strings.Replace(valid, `"schemaVersion":1`, `"schemaVersion":2`, 1), nil, "schemaVersion"},
		{"repo mismatch", strings.Replace(valid, `"repo":"owner/model"`, `"repo":"other/model"`, 1), nil, "repo"},
		{"revision mismatch", strings.Replace(valid, artifact.Revision, strings.Repeat("b", 40), 1), nil, "revision"},
		{"empty entries", strings.Replace(valid, manifestEntriesJSON(), `[]`, 1), nil, "entries"},
		{"duplicate path", strings.Replace(valid, manifestEntriesJSON(), `[{"path":"weights","type":"directory","size":0},{"path":"weights","type":"directory","size":0}]`, 1), nil, "duplicate"},
		{"invalid path", strings.Replace(valid, `"path":"weights/model.bin"`, `"path":"../model.bin"`, 1), nil, "path"},
		{"file target", strings.Replace(valid, `"sha256":"`+strings.Repeat("c", 64)+`"`, `"sha256":"`+strings.Repeat("c", 64)+`","target":"other"`, 1), nil, "target"},
		{"directory size", strings.Replace(valid, `"path":"weights","type":"directory","size":0`, `"path":"weights","type":"directory","size":1`, 1), nil, "directory"},
		{"escaping symlink", strings.Replace(valid, `"target":"model.bin"`, `"target":"../../outside"`, 1), nil, "target"},
		{"total mismatch", valid, func(a *BundleArtifactV1) { a.TotalSize++ }, "totalSize"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := canonicalTempDir(t)
			if err := os.MkdirAll(filepath.Join(dir, "manifests"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, artifact.Manifest), []byte(tt.raw), 0o644); err != nil {
				t.Fatal(err)
			}
			candidate := artifact
			sum := sha256.Sum256([]byte(tt.raw))
			candidate.ManifestSHA256 = fmt.Sprintf("%x", sum)
			if tt.mutate != nil {
				tt.mutate(&candidate)
			}
			root, err := openDirectoryNoSymlink(dir)
			if err != nil {
				t.Fatal(err)
			}
			defer root.Close()
			_, err = LoadArtifactManifest(root, candidate)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("LoadArtifactManifest() error = %v", err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("LoadArtifactManifest() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoadArtifactManifestRejectsSizeAndEntryLimits(t *testing.T) {
	t.Run("manifest bytes", func(t *testing.T) {
		dir := canonicalTempDir(t)
		artifact := validHFCacheArtifact(validArtifactManifestJSON())
		if err := os.MkdirAll(filepath.Join(dir, "manifests"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, artifact.Manifest), nil, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Truncate(filepath.Join(dir, artifact.Manifest), MaxArtifactManifestBytes+1); err != nil {
			t.Fatal(err)
		}
		root, err := openDirectoryNoSymlink(dir)
		if err != nil {
			t.Fatal(err)
		}
		defer root.Close()
		if _, err := LoadArtifactManifest(root, artifact); err == nil || !strings.Contains(err.Error(), "exceeds") {
			t.Fatalf("LoadArtifactManifest() oversized error = %v", err)
		}
	})

	t.Run("entry count", func(t *testing.T) {
		artifact := validHFCacheArtifact(validArtifactManifestJSON())
		manifest := ArtifactManifestV1{
			SchemaVersion: 1,
			Repo:          artifact.Repo,
			Revision:      artifact.Revision,
			Entries:       make([]ArtifactManifestEntryV1, MaxArtifactEntries+1),
		}
		if err := validateArtifactManifest(manifest, artifact); err == nil || !strings.Contains(err.Error(), "at most") {
			t.Fatalf("validateArtifactManifest() entry limit error = %v", err)
		}
	})
}

// The publisher owns both marker names: one says a publish is unfinished, the
// other says the tree is the artifact it claims to be. A manifest entry that
// could write either name could tell both lies.
func TestValidateArtifactManifestRejectsTheMarkerNames(t *testing.T) {
	artifact := validHFCacheArtifact(validArtifactManifestJSON())
	for _, name := range []string{
		hfCacheMarkerFileName,
		hfStageMarkerFileName,
		"snapshots/" + hfCacheMarkerFileName,
		"snapshots/" + hfStageMarkerFileName,
	} {
		t.Run(name, func(t *testing.T) {
			manifest := ArtifactManifestV1{
				SchemaVersion: 1, Repo: artifact.Repo, Revision: artifact.Revision,
				Entries: []ArtifactManifestEntryV1{{
					Path: name, Type: "file", Size: 0,
					SHA256: strings.Repeat("0", 64),
				}},
			}

			err := validateArtifactManifest(manifest, artifact)

			if err == nil || !strings.Contains(err.Error(), "reserved name") {
				t.Fatalf("validateArtifactManifest() error = %v, want a reserved name rejection", err)
			}
		})
	}
}

func TestLoadArtifactManifestRejectsInvalidDeclarationBeforeIO(t *testing.T) {
	dir := canonicalTempDir(t)
	root, err := openDirectoryNoSymlink(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	artifact := validHFCacheArtifact(validArtifactManifestJSON())
	artifact.Manifest = "../outside.json"
	if _, err := LoadArtifactManifest(root, artifact); err == nil || !strings.Contains(err.Error(), "manifest:") {
		t.Fatalf("LoadArtifactManifest() invalid declaration error = %v", err)
	}
}

func TestDecodeBundleRejectsArtifactDestination(t *testing.T) {
	raw := strings.Replace(validBundleJSON, `"appEntry":{}`, `"artifacts":[{"kind":"hf-cache","source":"artifacts/model","repo":"owner/model","revision":"`+strings.Repeat("a", 40)+`","manifest":"manifests/model.json","manifestSha256":"`+strings.Repeat("b", 64)+`","totalSize":1,"destination":"/tmp"}],"appEntry":{}`, 1)
	if _, err := DecodeBundle([]byte(raw)); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("DecodeBundle() destination error = %v", err)
	}
}

const validBundleJSON = `{
  "schemaVersion":"1",
  "sourceId":"market.olares",
  "catalogHash":"catalog",
  "generatedAt":"2026-07-25T00:00:00Z",
  "apps":[{
    "appId":"app-a",
    "appName":"app-a",
    "version":"1.0.0",
    "installScope":"shared",
    "chart":"charts/app-a-1.0.0.tgz",
    "chartSha256":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
    "installOrder":10,
    "allowedEnvs":["WORKER_COUNT"],
    "allowedGpuTypes":["nvidia"],
    "appEntry":{}
  }]
}`

func decodeBundle(t *testing.T, raw string) BundleV1 {
	t.Helper()
	var bundle BundleV1
	if err := json.Unmarshal([]byte(raw), &bundle); err != nil {
		t.Fatal(err)
	}
	return bundle
}

func writeContractDirectory(t *testing.T) string {
	t.Helper()
	dir := canonicalTempDir(t)
	if err := os.WriteFile(filepath.Join(dir, BundleFileName), []byte(validBundleJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ProfileFileName), []byte(`{"schemaVersion":"1","apps":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func validHFCacheArtifact(manifest string) BundleArtifactV1 {
	sum := sha256.Sum256([]byte(manifest))
	return BundleArtifactV1{
		Kind:           ArtifactKindHFCache,
		Source:         "artifacts/models/owner--model",
		Repo:           "owner/model",
		Revision:       strings.Repeat("a", 40),
		Manifest:       "manifests/owner--model.json",
		ManifestSHA256: fmt.Sprintf("%x", sum),
		TotalSize:      3,
	}
}

func validArtifactManifestJSON() string {
	return `{"schemaVersion":1,"repo":"owner/model","revision":"` + strings.Repeat("a", 40) + `","entries":` + manifestEntriesJSON() + `}`
}

func manifestEntriesJSON() string {
	return `[{"path":"weights","type":"directory","size":0},{"path":"weights/model.bin","type":"file","size":3,"sha256":"` + strings.Repeat("c", 64) + `"},{"path":"weights/latest","type":"symlink","size":0,"target":"model.bin"}]`
}

func twoArtifactBundle(t *testing.T, artifact BundleArtifactV1) BundleV1 {
	t.Helper()
	bundle := decodeBundle(t, validBundleJSON)
	bundle.Apps[0].Artifacts = []BundleArtifactV1{artifact}
	second := bundle.Apps[0]
	second.AppID = "app-b"
	second.AppName = "app-b"
	second.Chart = "charts/app-b-1.0.0.tgz"
	second.Artifacts = []BundleArtifactV1{artifact}
	second.Artifacts[0].Source = "artifacts/models/other--model"
	second.Artifacts[0].Manifest = "manifests/other--model.json"
	bundle.Apps = append(bundle.Apps, second)
	return bundle
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
