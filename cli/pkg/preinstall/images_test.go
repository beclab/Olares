package preinstall

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	pkgchart "github.com/beclab/Olares/cli/pkg/chart"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/utils"
)

func TestCheckStaticBundleWithoutInstallationManifestSkipsImageGate(t *testing.T) {
	dir := buildImageBundle(t, imageApp{name: "gateway", image: "docker.io/beclab/router:v2.5.0"})

	if err := CheckStaticBundle(dir, CheckOptions{}); err != nil {
		t.Fatalf("CheckStaticBundle() error = %v", err)
	}
}

func TestCheckStaticBundleOlaresYAMLAcceptsListedImage(t *testing.T) {
	dir := buildImageBundle(t, imageApp{name: "gateway", image: "docker.io/beclab/router:v2.5.0"})
	opts := CheckOptions{OlaresImages: []string{"beclab/router:v2.5.0"}}

	if err := CheckStaticBundle(dir, opts); err != nil {
		t.Fatalf("CheckStaticBundle() error = %v", err)
	}
}

func TestCheckStaticBundleOlaresYAMLIgnoresExtraDeclaredImages(t *testing.T) {
	dir := buildImageBundle(t, imageApp{name: "gateway", image: "docker.io/beclab/router:v2.5.0"})
	opts := CheckOptions{OlaresImages: []string{"beclab/router:v2.5.0", "beclab/unrelated:v1"}}

	if err := CheckStaticBundle(dir, opts); err != nil {
		t.Fatalf("CheckStaticBundle() error = %v", err)
	}
}

func TestCheckStaticBundleOlaresYAMLReportsUnlistedImage(t *testing.T) {
	dir := buildImageBundle(t, imageApp{name: "model", image: "docker.io/beclab/llama.cpp:b10548"})
	opts := CheckOptions{OlaresImages: []string{"beclab/router:v2.5.0"}}

	err := CheckStaticBundle(dir, opts)
	if err == nil {
		t.Fatal("CheckStaticBundle() error = nil, want an unlisted image")
	}
	if !strings.Contains(err.Error(), "beclab/llama.cpp:b10548") {
		t.Fatalf("CheckStaticBundle() error = %v, want the image named", err)
	}
	if strings.Contains(err.Error(), "docker.io/beclab/llama.cpp") {
		t.Fatalf("CheckStaticBundle() error = %v, want the familiar spelling", err)
	}
	if !strings.Contains(err.Error(), string(ImageGapUnlistedOlaresYAML)) {
		t.Fatalf("CheckStaticBundle() error = %v, want kind %q", err, ImageGapUnlistedOlaresYAML)
	}
	if strings.Contains(err.Error(), string(ImageGapUnlisted)) {
		t.Fatalf("CheckStaticBundle() error = %v, want the olares.yaml kind, not the manifest kind", err)
	}
}

func TestCheckStaticBundleOlaresYAMLNormalizesRegistryPrefix(t *testing.T) {
	tests := []struct {
		name       string
		chartImage string
		yamlImage  string
	}{
		{"long chart, short yaml", "docker.io/beclab/router:v2.5.0", "beclab/router:v2.5.0"},
		{"short chart, long yaml", "beclab/router:v2.5.0", "docker.io/beclab/router:v2.5.0"},
		{"non-docker registry", "ghcr.io/beclab/router:v2.5.0", "ghcr.io/beclab/router:v2.5.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := buildImageBundle(t, imageApp{name: "gateway", image: tt.chartImage})
			opts := CheckOptions{OlaresImages: []string{tt.yamlImage}}
			if err := CheckStaticBundle(dir, opts); err != nil {
				t.Fatalf("CheckStaticBundle() error = %v", err)
			}
		})
	}
}

func TestCheckStaticBundleOlaresYAMLEmptyListReportsEveryChartImage(t *testing.T) {
	dir := buildImageBundle(t, imageApp{name: "gateway", image: "docker.io/beclab/router:v2.5.0"})
	opts := CheckOptions{OlaresImages: []string{}}

	err := CheckStaticBundle(dir, opts)
	if err == nil || !strings.Contains(err.Error(), "beclab/router:v2.5.0") {
		t.Fatalf("CheckStaticBundle() error = %v, want the chart image reported", err)
	}
}

func TestCheckStaticBundleRejectsBothImageListSources(t *testing.T) {
	dir := buildImageBundle(t, imageApp{name: "gateway", image: "docker.io/beclab/router:v2.5.0"})
	opts := CheckOptions{
		InstallationManifest: buildInstallationManifest(t, "beclab/router:v2.5.0"),
		OlaresImages:         []string{"beclab/router:v2.5.0"},
	}

	err := CheckStaticBundle(dir, opts)
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("CheckStaticBundle() error = %v, want mutually exclusive sources", err)
	}
}

func TestCheckStaticBundleImageGateAcceptsListedImage(t *testing.T) {
	dir := buildImageBundle(t, imageApp{name: "gateway", image: "docker.io/beclab/router:v2.5.0"})
	opts := CheckOptions{InstallationManifest: buildInstallationManifest(t, "beclab/router:v2.5.0")}

	if err := CheckStaticBundle(dir, opts); err != nil {
		t.Fatalf("CheckStaticBundle() error = %v", err)
	}
}

// A chart may name its image through .Values, so the gate has to render rather
// than scan: a textual scan would compare the template string, never the image
// that gets pulled.
func TestCheckStaticBundleImageGateResolvesValuesTemplatedImage(t *testing.T) {
	dir := buildImageBundle(t, imageApp{
		name:   "gateway",
		image:  "{{ .Values.image.repository }}:{{ .Values.image.tag }}",
		values: "image:\n  repository: docker.io/beclab/router\n  tag: v2.5.0\n",
	})

	if err := CheckStaticBundle(dir, CheckOptions{
		InstallationManifest: buildInstallationManifest(t, "beclab/router:v2.5.0"),
	}); err != nil {
		t.Fatalf("CheckStaticBundle() error = %v", err)
	}

	err := CheckStaticBundle(dir, CheckOptions{
		InstallationManifest: buildInstallationManifest(t, "beclab/router:v2.4.0"),
	})
	if err == nil || !strings.Contains(err.Error(), "beclab/router:v2.5.0") {
		t.Fatalf("CheckStaticBundle() error = %v, want the rendered tag reported", err)
	}
}

// installation.manifest writes docker.io images short and charts generally write
// them long. Both spellings name the same image, so neither may be reported as a
// gap.
func TestCheckStaticBundleImageGateNormalizesRegistryPrefix(t *testing.T) {
	tests := []struct {
		name          string
		chartImage    string
		manifestImage string
	}{
		{"long chart, short manifest", "docker.io/beclab/router:v2.5.0", "beclab/router:v2.5.0"},
		{"short chart, long manifest", "beclab/router:v2.5.0", "docker.io/beclab/router:v2.5.0"},
		{"non-docker registry", "ghcr.io/beclab/router:v2.5.0", "ghcr.io/beclab/router:v2.5.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := buildImageBundle(t, imageApp{name: "gateway", image: tt.chartImage})
			opts := CheckOptions{InstallationManifest: buildInstallationManifest(t, tt.manifestImage)}
			if err := CheckStaticBundle(dir, opts); err != nil {
				t.Fatalf("CheckStaticBundle() error = %v", err)
			}
		})
	}
}

// The reported spelling is the one that md5s into the uploaded payload name, so
// a gap has to be printed the way images.mf writes it, never fully qualified.
func TestCheckStaticBundleImageGateReportsUnlistedImageInManifestSpelling(t *testing.T) {
	dir := buildImageBundle(t, imageApp{name: "model", image: "docker.io/beclab/llama.cpp:b10548"})
	opts := CheckOptions{InstallationManifest: buildInstallationManifest(t, "beclab/router:v2.5.0")}

	err := CheckStaticBundle(dir, opts)
	if err == nil {
		t.Fatal("CheckStaticBundle() error = nil, want an unlisted image")
	}
	if !strings.Contains(err.Error(), "beclab/llama.cpp:b10548") {
		t.Fatalf("CheckStaticBundle() error = %v, want the image named", err)
	}
	if strings.Contains(err.Error(), "docker.io/beclab/llama.cpp") {
		t.Fatalf("CheckStaticBundle() error = %v, want the manifest spelling", err)
	}
	if !strings.Contains(err.Error(), string(ImageGapUnlisted)) {
		t.Fatalf("CheckStaticBundle() error = %v, want kind %q", err, ImageGapUnlisted)
	}
}

// A manifest line is only half of what prepare needs: LoadImages still has to
// find a tarball named after the md5 of that line, or it aborts.
func TestCheckStaticBundleImageGateRequiresImagePayload(t *testing.T) {
	dir := buildImageBundle(t, imageApp{name: "gateway", image: "docker.io/beclab/router:v2.5.0"})
	opts := CheckOptions{
		InstallationManifest: buildInstallationManifest(t, "beclab/router:v2.5.0"),
		ImagesDir:            t.TempDir(),
	}

	err := CheckStaticBundle(dir, opts)
	if err == nil || !strings.Contains(err.Error(), string(ImageGapNoPayload)) {
		t.Fatalf("CheckStaticBundle() error = %v, want kind %q", err, ImageGapNoPayload)
	}

	writeImagePayload(t, opts.ImagesDir, "beclab/router:v2.5.0")
	if err := CheckStaticBundle(dir, opts); err != nil {
		t.Fatalf("CheckStaticBundle() error = %v", err)
	}
}

// A stale image list usually misses one image per app, so a single run has to
// show the whole shape of the problem rather than the first app that trips.
func TestCheckStaticBundleImageGateAggregatesAcrossApps(t *testing.T) {
	dir := buildImageBundle(t,
		imageApp{name: "model", image: "docker.io/beclab/llama.cpp:b10548"},
		imageApp{name: "gateway", image: "docker.io/beclab/router:v2.5.0"},
	)
	opts := CheckOptions{InstallationManifest: buildInstallationManifest(t, "beclab/unrelated:v1")}

	err := CheckStaticBundle(dir, opts)
	if err == nil {
		t.Fatal("CheckStaticBundle() error = nil, want both apps reported")
	}
	for _, want := range []string{"beclab/llama.cpp:b10548", "beclab/router:v2.5.0"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("CheckStaticBundle() error = %v, want %q reported", err, want)
		}
	}
}

func TestCheckStaticBundleImageGateRejectsFloatingTag(t *testing.T) {
	tests := []struct {
		name  string
		image string
		want  string
	}{
		{"explicit latest", "docker.io/beclab/router:latest", "latest tag"},
		{"implicit latest", "docker.io/beclab/router", "has no tag"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := buildImageBundle(t, imageApp{name: "gateway", image: tt.image})
			opts := CheckOptions{InstallationManifest: buildInstallationManifest(t, "beclab/router:v2.5.0")}
			err := CheckStaticBundle(dir, opts)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("CheckStaticBundle() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestCheckStaticBundleImageGateFailsOnUnrenderableChart(t *testing.T) {
	dir := buildImageBundle(t, imageApp{name: "gateway", image: "{{ .Values.missing.repository }}"})
	opts := CheckOptions{InstallationManifest: buildInstallationManifest(t, "beclab/router:v2.5.0")}

	err := CheckStaticBundle(dir, opts)
	if err == nil || !strings.Contains(err.Error(), "render chart") {
		t.Fatalf("CheckStaticBundle() error = %v, want a render failure", err)
	}
}

type imageApp struct {
	name   string
	image  string
	values string
}

// buildImageBundle writes a static market bundle whose charts each run one
// container, so a test can state the image set it wants the gate to see.
func buildImageBundle(t *testing.T, apps ...imageApp) string {
	t.Helper()
	dir := filepath.Join(canonicalTempDir(t), "market")
	chartDir := filepath.Join(dir, "chart")
	if err := os.MkdirAll(chartDir, 0o755); err != nil {
		t.Fatal(err)
	}

	bundle := BundleV1{
		SchemaVersion: SupportedSchemaVersion,
		SourceID:      OfficialSourceID,
		CatalogHash:   "image-gate-fixture",
		GeneratedAt:   "2026-07-25T00:00:00Z",
	}
	for i, app := range apps {
		archive := packageImageChart(t, chartDir, app)
		bundle.Apps = append(bundle.Apps, BundleAppV1{
			AppID:        fmt.Sprintf("app-%d", i),
			AppName:      app.name,
			Version:      "1.0.0",
			InstallScope: InstallScopeShared,
			Chart:        "chart/" + filepath.Base(archive),
			ChartSHA256:  fileSHA256(t, archive),
			InstallOrder: (i + 1) * 10,
			AppEntry:     json.RawMessage(fmt.Sprintf(`{"id":"app-%d","name":%q,"version":"1.0.0","cfgType":"app"}`, i, app.name)),
		})
	}

	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, BundleFileName), append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func packageImageChart(t *testing.T, outputDir string, app imageApp) string {
	t.Helper()
	source := filepath.Join(t.TempDir(), app.name)
	if err := os.MkdirAll(filepath.Join(source, "templates"), 0o755); err != nil {
		t.Fatal(err)
	}
	values := app.values
	if values == "" {
		values = "{}\n"
	}
	files := map[string]string{
		"Chart.yaml":                fmt.Sprintf("apiVersion: v2\nname: %s\nversion: 1.0.0\n", app.name),
		"values.yaml":               values,
		"OlaresManifest.yaml":       imageChartManifest(app.name),
		"templates/deployment.yaml": imageChartDeployment(app.name, app.image),
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(source, filepath.FromSlash(name)), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	archive, err := pkgchart.Package(source, outputDir)
	if err != nil {
		t.Fatal(err)
	}
	return archive
}

func imageChartManifest(name string) string {
	return fmt.Sprintf(`olaresManifest.version: "0.13.0"
olaresManifest.type: app
apiVersion: v3
metadata:
  name: %[1]s
  version: 1.0.0
  title: %[1]s
  description: Image gate fixture
  icon: https://example.com/icon.png
entrances:
  - name: %[1]s
    port: 80
    host: %[1]s
    title: %[1]s
spec:
  onlyAdmin: true
  supportArch:
    - amd64
  requiredCpu: 50m
  limitedCpu: 500m
  requiredMemory: 20Mi
  limitedMemory: 100Mi
  requiredDisk: 50Mi
  limitedDisk: 100Mi
workloadReplicas:
  %[1]s: 1
options:
  dependencies:
    - name: olares
      type: system
      version: ">=1.12.6-0"
  shared: true
`, name)
}

func imageChartDeployment(name, image string) string {
	return fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: %[1]s
spec:
  replicas: 1
  selector:
    matchLabels:
      app: %[1]s
  template:
    metadata:
      labels:
        app: %[1]s
    spec:
      containers:
        - name: %[1]s
          image: %[2]s
          resources:
            requests:
              cpu: 50m
              memory: 20Mi
            limits:
              cpu: 500m
              memory: 100Mi
`, name, image)
}

// buildInstallationManifest writes the image rows the way build-manifest.sh does
// and parses them back, so the gate is tested against the real line format
// rather than a hand-built map.
func buildInstallationManifest(t *testing.T, images ...string) manifest.InstallationManifest {
	t.Helper()
	lines := make([]string, 0, len(images))
	for _, image := range images {
		lines = append(lines, fmt.Sprintf(
			"%s.tar.gz,images,images.mf,https://cdn/amd64,amd64sum,https://cdn/arm64,arm64sum,%s",
			utils.MD5(image), image,
		))
	}
	path := filepath.Join(t.TempDir(), "installation.manifest")
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	installation, err := manifest.ReadAll(path)
	if err != nil {
		t.Fatal(err)
	}
	return installation
}

func writeImagePayload(t *testing.T, imagesDir, image string) {
	t.Helper()
	name := utils.MD5(image) + ".tar.gz"
	if err := os.WriteFile(filepath.Join(imagesDir, name), []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func fileSHA256(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
