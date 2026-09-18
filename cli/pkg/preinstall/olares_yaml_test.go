package preinstall

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestReadOlaresYAMLContainers(t *testing.T) {
	path := writeOlaresYAML(t, `apiVersion: v1
target: prebuilt
output:
  containers:
    - name: beclab/router:v2.5.0
    - name: "  "
    -
      name: docker.io/beclab/llama.cpp:b10548
`)

	got, err := ReadOlaresYAMLContainers(path)
	if err != nil {
		t.Fatalf("ReadOlaresYAMLContainers() error = %v", err)
	}
	want := []string{"beclab/router:v2.5.0", "docker.io/beclab/llama.cpp:b10548"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ReadOlaresYAMLContainers() = %v, want %v", got, want)
	}
}

func TestReadOlaresYAMLContainersEmptyIsNonNil(t *testing.T) {
	path := writeOlaresYAML(t, "apiVersion: v1\noutput: {}\n")

	got, err := ReadOlaresYAMLContainers(path)
	if err != nil {
		t.Fatalf("ReadOlaresYAMLContainers() error = %v", err)
	}
	if got == nil {
		t.Fatal("ReadOlaresYAMLContainers() = nil, want a non-nil empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("ReadOlaresYAMLContainers() = %v, want empty", got)
	}
}

func TestReadOlaresYAMLContainersRejectsInvalidYAML(t *testing.T) {
	path := writeOlaresYAML(t, "output: [\n")

	_, err := ReadOlaresYAMLContainers(path)
	if err == nil {
		t.Fatal("ReadOlaresYAMLContainers() error = nil, want decode failure")
	}
}

func writeOlaresYAML(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "Olares.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
