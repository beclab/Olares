package oac

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeValues(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "values.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write values.yaml: %v", err)
	}
	return dir
}

func TestCheckWorkloadReplicaValues_AllPresent(t *testing.T) {
	dir := writeValues(t, `
workloads:
  affine:
    replicaCount: 1
  db:
    replicaCount: 2
`)
	if err := checkWorkloadReplicaValues(dir, map[string]int32{"affine": 1, "db": 1}); err != nil {
		t.Fatalf("all replicaCount present must pass: %v", err)
	}
}

func TestCheckWorkloadReplicaValues_MissingWorkloadKey(t *testing.T) {
	dir := writeValues(t, `
workloads:
  affine:
    replicaCount: 1
`)
	err := checkWorkloadReplicaValues(dir, map[string]int32{"affine": 1, "db": 1})
	if err == nil {
		t.Fatal("expected error when a workload is absent from values.yaml")
	}
	if !strings.Contains(err.Error(), "workloads.db.replicaCount") {
		t.Fatalf("error should mention workloads.db.replicaCount, got: %v", err)
	}
}

func TestCheckWorkloadReplicaValues_MissingReplicaCount(t *testing.T) {
	dir := writeValues(t, `
workloads:
  affine:
    image: foo
`)
	err := checkWorkloadReplicaValues(dir, map[string]int32{"affine": 1})
	if err == nil {
		t.Fatal("expected error when replicaCount is absent for a workload")
	}
	if !strings.Contains(err.Error(), "workloads.affine.replicaCount") {
		t.Fatalf("error should mention workloads.affine.replicaCount, got: %v", err)
	}
}
