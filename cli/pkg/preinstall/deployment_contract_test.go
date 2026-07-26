package preinstall

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMarketDeploymentMountsPreinstallReadOnlyBesideWritableData(t *testing.T) {
	path := filepath.Join("..", "..", "..", "framework", "market", ".olares", "config", "cluster", "deploy", "market_deploy.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	deployment := deploymentDocument(t, string(data), "market-deployment")
	container := between(t, deployment, "      - name: appstore-backend\n", "      volumes:\n")
	volumes := after(t, deployment, "      volumes:\n")

	// The two preinstall variables are rendered only for an installer that
	// published a bundle, so an ordinary install does not have Market open an
	// empty mount on every boot. They are asserted as one block for that
	// reason: an enabled flag without a directory turns the importer on with
	// nothing to read.
	preinstallEnv := "{{ if .Values.preinstall }}\n" +
		"          - name: PREINSTALL_ENABLED\n            value: 'true'\n" +
		"          - name: PREINSTALL_BUNDLE_DIR\n            value: /opt/app/preinstall\n" +
		"{{ end }}"
	for _, required := range []string{
		preinstallEnv,
		"- name: opt-data\n            mountPath: /opt/app/data",
		"- name: market-preinstall\n            mountPath: /opt/app/preinstall\n            readOnly: true",
	} {
		if !strings.Contains(container, required) {
			t.Errorf("appstore-backend container missing contract:\n%s", required)
		}
	}
	if strings.Count(container, "PREINSTALL_ENABLED") != 1 {
		t.Errorf("PREINSTALL_ENABLED must be rendered from one gated block")
	}
	// The mount is derived from RuntimeRelativeDir so that moving the publish
	// target in the CLI cannot silently diverge from what the pod mounts.
	preinstallVolume := fmt.Sprintf(
		"- name: market-preinstall\n        hostPath:\n          path: '{{ .Values.rootPath }}/%s'\n          type: DirectoryOrCreate",
		RuntimeRelativeDir,
	)
	for _, required := range []string{
		"- name: opt-data\n        hostPath:\n          path: '{{ .Values.rootPath }}/userdata/Cache/chartrepo'\n          type: DirectoryOrCreate",
		preinstallVolume,
	} {
		if !strings.Contains(volumes, required) {
			t.Errorf("market deployment volumes missing contract:\n%s", required)
		}
	}
	if strings.Count(container, "- name: market-preinstall\n") != 1 ||
		strings.Count(volumes, "- name: market-preinstall\n") != 1 {
		t.Errorf("market-preinstall must have one backend mount and one deployment volume")
	}
}

func TestMarketPreinstallDeploymentPublishesManifestsButNotPayloads(t *testing.T) {
	installerDir, baseDir := writeStaticBundle(t)
	artifact, _ := addMaterializeArtifactFixture(t, installerDir)
	if err := Materialize(installerDir, baseDir, ProfileSelections{}); err != nil {
		t.Fatalf("Materialize() error = %v", err)
	}
	target := filepath.Join(baseDir, RuntimeRelativeDir)
	t.Cleanup(func() { _ = makeWritable(target) })

	allowed := map[string]bool{
		BundleFileName:  true,
		ProfileFileName: true,
		"charts":        true,
		"manifests":     true,
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !allowed[entry.Name()] {
			t.Errorf("market preinstall contains undeclared top-level entry %q", entry.Name())
		}
	}
	if _, err := os.Stat(filepath.Join(target, filepath.FromSlash(artifact.Manifest))); err != nil {
		t.Fatalf("artifact manifest is not deployable: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(target, filepath.FromSlash(artifact.Source))); !os.IsNotExist(err) {
		t.Fatalf("artifact payload crossed deployment boundary: %v", err)
	}
}

func deploymentDocument(t *testing.T, template, name string) string {
	t.Helper()
	for _, document := range strings.Split(template, "\n---") {
		if strings.Contains(document, "kind: Deployment\n") &&
			strings.Contains(document, "  name: "+name+"\n") {
			return document
		}
	}
	t.Fatalf("deployment %q not found", name)
	return ""
}

func between(t *testing.T, value, start, end string) string {
	t.Helper()
	startIndex := strings.Index(value, start)
	if startIndex < 0 {
		t.Fatalf("start marker %q not found", start)
	}
	endIndex := strings.Index(value[startIndex:], end)
	if endIndex < 0 {
		t.Fatalf("end marker %q not found", end)
	}
	return value[startIndex : startIndex+endIndex]
}

func after(t *testing.T, value, marker string) string {
	t.Helper()
	index := strings.Index(value, marker)
	if index < 0 {
		t.Fatalf("marker %q not found", marker)
	}
	return value[index+len(marker):]
}
