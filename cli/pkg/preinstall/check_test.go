package preinstall

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckStaticBundleGoldenFixture(t *testing.T) {
	dir := filepath.Join("testdata", "materialized-bundle", "source", "preinstall", "market")
	if err := CheckStaticBundle(dir, CheckOptions{}); err != nil {
		t.Fatalf("CheckStaticBundle() error = %v", err)
	}
	if err := CheckStaticBundle(dir, CheckOptions{Full: true}); err != nil {
		t.Fatalf("CheckStaticBundle(Full) error = %v", err)
	}
}

func TestCheckStaticBundleRejectsBadChartDigest(t *testing.T) {
	dir := copyStaticMarket(t)
	bundlePath := filepath.Join(dir, BundleFileName)
	bundle := decodeCheckBundleFile(t, bundlePath)
	bundle.Apps[0].ChartSHA256 = strings.Repeat("0", 64)
	writeCheckBundleFile(t, bundlePath, bundle)

	err := CheckStaticBundle(dir, CheckOptions{})
	if err == nil || !strings.Contains(err.Error(), "digest mismatch") {
		t.Fatalf("CheckStaticBundle() error = %v, want digest mismatch", err)
	}
}

func TestCheckStaticBundleRejectsMissingChart(t *testing.T) {
	dir := copyStaticMarket(t)
	bundle := decodeCheckBundleFile(t, filepath.Join(dir, BundleFileName))
	chartPath := filepath.Join(dir, filepath.FromSlash(bundle.Apps[0].Chart))
	if err := os.Remove(chartPath); err != nil {
		t.Fatal(err)
	}

	err := CheckStaticBundle(dir, CheckOptions{})
	if err == nil || !strings.Contains(err.Error(), "chart") {
		t.Fatalf("CheckStaticBundle() error = %v, want missing chart", err)
	}
}

func TestCheckStaticBundleRejectsBadManifestDigest(t *testing.T) {
	dir := copyStaticMarket(t)
	bundlePath := filepath.Join(dir, BundleFileName)
	bundle := decodeCheckBundleFile(t, bundlePath)
	bundle.Apps[0].Artifacts[0].ManifestSHA256 = strings.Repeat("1", 64)
	writeCheckBundleFile(t, bundlePath, bundle)

	err := CheckStaticBundle(dir, CheckOptions{})
	if err == nil || !strings.Contains(err.Error(), "manifestSha256") {
		t.Fatalf("CheckStaticBundle() error = %v, want manifestSha256 mismatch", err)
	}
}

func TestCheckStaticBundleRejectsDuplicateHFCacheTarget(t *testing.T) {
	dir := copyStaticMarket(t)
	bundlePath := filepath.Join(dir, BundleFileName)
	bundle := decodeCheckBundleFile(t, bundlePath)
	second := bundle.Apps[0]
	second.AppID = "second-app"
	second.AppName = "second-app"
	second.Chart = "chart/second-app-1.0.0.tgz"
	secondArtifact := second.Artifacts[0]
	secondArtifact.Source = "artifacts/fixture--tiny-model-copy"
	secondArtifact.Manifest = "manifests/fixture--tiny-model-copy.json"
	second.Artifacts = []BundleArtifactV1{secondArtifact}
	bundle.Apps = append(bundle.Apps, second)
	writeCheckBundleFile(t, bundlePath, bundle)

	chartData, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(bundle.Apps[0].Chart)))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(second.Chart)), chartData, 0o644); err != nil {
		t.Fatal(err)
	}

	err = CheckStaticBundle(dir, CheckOptions{})
	if err == nil || !strings.Contains(err.Error(), "duplicate Hugging Face cache target") {
		t.Fatalf("CheckStaticBundle() error = %v, want duplicate cache target", err)
	}
}

func TestCheckStaticBundleFullRejectsTamperedArtifact(t *testing.T) {
	dir := copyStaticMarket(t)
	payload := filepath.Join(dir, "artifacts", "fixture--tiny-model", "tiny.bin")
	if err := os.WriteFile(payload, []byte("XXXXX"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := CheckStaticBundle(dir, CheckOptions{}); err != nil {
		t.Fatalf("contract check should pass without --full: %v", err)
	}
	err := CheckStaticBundle(dir, CheckOptions{Full: true})
	if err == nil || !strings.Contains(err.Error(), "digest mismatch") {
		t.Fatalf("CheckStaticBundle(Full) error = %v, want digest mismatch", err)
	}
}

func TestCheckStaticBundleFullRejectsHardlinkedArtifact(t *testing.T) {
	dir := copyStaticMarket(t)
	payload := filepath.Join(dir, "artifacts", "fixture--tiny-model", "tiny.bin")
	external := filepath.Join(filepath.Dir(dir), "external.bin")
	data, err := os.ReadFile(payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(external, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(payload); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(external, payload); err != nil {
		t.Fatal(err)
	}

	err = CheckStaticBundle(dir, CheckOptions{Full: true})
	if err == nil || !strings.Contains(err.Error(), "must not be a hardlink") {
		t.Fatalf("CheckStaticBundle(Full) error = %v, want hardlink rejection", err)
	}
}

func TestCheckStaticBundleFullIgnoresUndeclaredArtifactPathLikeInstaller(t *testing.T) {
	dir := copyStaticMarket(t)
	extra := filepath.Join(dir, "artifacts", "fixture--tiny-model", "extra.bin")
	if err := os.WriteFile(extra, []byte("extra"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := CheckStaticBundle(dir, CheckOptions{Full: true}); err != nil {
		t.Fatalf("CheckStaticBundle(Full) error = %v", err)
	}
}

func copyStaticMarket(t *testing.T) string {
	t.Helper()
	src := filepath.Join("testdata", "materialized-bundle", "source", "preinstall", "market")
	dst := filepath.Join(canonicalTempDir(t), "market")
	if err := os.CopyFS(dst, os.DirFS(src)); err != nil {
		t.Fatal(err)
	}
	return dst
}

func decodeCheckBundleFile(t *testing.T, path string) BundleV1 {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := DecodeBundle(data)
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

func writeCheckBundleFile(t *testing.T, path string, bundle BundleV1) {
	t.Helper()
	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}
