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

// writeTemplateFile writes name under dir/templates, creating the directory.
func writeTemplateFile(t *testing.T, dir, name, body string) {
	t.Helper()
	tpl := filepath.Join(dir, "templates")
	if err := os.MkdirAll(filepath.Dir(filepath.Join(tpl, name)), 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tpl, name), []byte(body), 0o644); err != nil {
		t.Fatalf("write template %s: %v", name, err)
	}
}

func TestCheckWorkloadReplicaTemplates_Valid(t *testing.T) {
	dir := t.TempDir()
	writeTemplateFile(t, dir, "deployment.yaml", `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}
spec:
  replicas: {{ .Values.workloads.web.replicaCount }}
`)
	if err := checkWorkloadReplicaTemplates(dir, map[string]int32{"web": 1}); err != nil {
		t.Fatalf("valid replicas reference must pass: %v", err)
	}
}

func TestCheckWorkloadReplicaTemplates_IndexForm(t *testing.T) {
	dir := t.TempDir()
	writeTemplateFile(t, dir, "sts.yaml", `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis
spec:
  replicas: {{ (index .Values.workloads "my-redis").replicaCount }}
`)
	if err := checkWorkloadReplicaTemplates(dir, map[string]int32{"my-redis": 1}); err != nil {
		t.Fatalf("index-form replicas reference must pass: %v", err)
	}
}

func TestCheckWorkloadReplicaTemplates_Hardcoded(t *testing.T) {
	dir := t.TempDir()
	writeTemplateFile(t, dir, "deployment.yaml", `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
spec:
  replicas: 3
`)
	err := checkWorkloadReplicaTemplates(dir, map[string]int32{"web": 1})
	if err == nil {
		t.Fatal("expected error for hardcoded spec.replicas")
	}
	if !strings.Contains(err.Error(), "must reference .Values.workloads") {
		t.Fatalf("error should explain the values requirement, got: %v", err)
	}
}

func TestCheckWorkloadReplicaTemplates_UnknownWorkload(t *testing.T) {
	dir := t.TempDir()
	writeTemplateFile(t, dir, "deployment.yaml", `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
spec:
  replicas: {{ .Values.workloads.ghost.replicaCount }}
`)
	err := checkWorkloadReplicaTemplates(dir, map[string]int32{"web": 1})
	if err == nil {
		t.Fatal("expected error when referenced workload is not in workloadReplicas")
	}
	if !strings.Contains(err.Error(), "not declared in workloadReplicas") {
		t.Fatalf("error should flag the unknown workload, got: %v", err)
	}
}

func TestCheckWorkloadReplicaTemplates_NoReplicasFieldOK(t *testing.T) {
	dir := t.TempDir()
	writeTemplateFile(t, dir, "deployment.yaml", `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
spec:
  selector:
    matchLabels:
      app: web
`)
	if err := checkWorkloadReplicaTemplates(dir, map[string]int32{"web": 1}); err != nil {
		t.Fatalf("a workload without spec.replicas must pass: %v", err)
	}
}

func TestCheckWorkloadReplicaTemplates_NonWorkloadIgnored(t *testing.T) {
	dir := t.TempDir()
	// A bare replicas: in a non-workload doc must not be flagged.
	writeTemplateFile(t, dir, "other.yaml", `apiVersion: v1
kind: ConfigMap
metadata:
  name: cfg
data:
  replicas: "3"
`)
	if err := checkWorkloadReplicaTemplates(dir, map[string]int32{"web": 1}); err != nil {
		t.Fatalf("non-workload replicas must be ignored: %v", err)
	}
}

func TestCheckWorkloadReplicaTemplates_NestedTemplate(t *testing.T) {
	dir := t.TempDir()
	writeTemplateFile(t, dir, "web/deployment.yaml", `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
spec:
  replicas: 2
`)
	err := checkWorkloadReplicaTemplates(dir, map[string]int32{"web": 1})
	if err == nil {
		t.Fatal("expected error for hardcoded replicas in a nested template")
	}
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
