package oac_test

import (
	"github.com/beclab/Olares/framework/oac"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeV3Chart(t *testing.T, templateBody string) string {
	t.Helper()
	dir := t.TempDir()
	manifest := `olaresManifest.version: "0.12.0"
olaresManifest.type: app
apiVersion: v3
metadata:
  name: v3env
  icon: https://example.com/i.png
  description: d
  title: T
  version: 1.0.0
entrances:
- name: main
  host: v3env
  port: 8080
  title: Main
spec:
  requiredCpu: 100m
  limitedCpu: 200m
  requiredMemory: 128Mi
  limitedMemory: 256Mi
  requiredDisk: 1Gi
  supportArch: [amd64]
workloadReplicas:
  v3env: 1
envs:
  - envName: SMTP_HOST
    applyOnChange: true
    valueFrom:
      envName: OLARES_USER_SMTP_SERVER
options:
  policies: []
  resetCookie: { enabled: false }
`
	os.WriteFile(filepath.Join(dir, "Chart.yaml"), []byte("apiVersion: v2\nname: v3env\nversion: 1.0.0\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "values.yaml"), []byte("x: y\nworkloads:\n  v3env:\n    replicaCount: 1\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "OlaresManifest.yaml"), []byte(manifest), 0o644)
	tmpl := filepath.Join(dir, "templates")
	os.MkdirAll(tmpl, 0o755)
	os.WriteFile(filepath.Join(tmpl, "deployment.yaml"), []byte(templateBody), 0o644)
	return dir
}

const v3DeployHeader = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: v3env
spec:
  replicas: 1
  selector:
    matchLabels:
      app: v3env
  template:
    metadata:
      labels:
        app: v3env
    spec:
      containers:
        - name: main
          image: nginx:1
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 200m
              memory: 256Mi
          env:
`

func TestLint_V3_OlaresEnvInTemplateOK(t *testing.T) {
	dir := writeV3Chart(t, v3DeployHeader+`            - name: SMTP_HOST
              value: "{{ .Values.olaresEnv.SMTP_HOST }}"
`)
	if err := oac.Lint(dir, oac.SkipResourceCheck(), oac.SkipSameVersionCheck()); err != nil {
		t.Fatalf("Lint: %v", err)
	}
}

func TestLint_V3_OLARESUserInTemplateFails(t *testing.T) {
	dir := writeV3Chart(t, v3DeployHeader+`            - name: SMTP_HOST
              value: "{{ .Values.OLARES_USER_SMTP_SERVER }}"
`)
	err := oac.Lint(dir, oac.SkipResourceCheck(), oac.SkipSameVersionCheck())
	if err == nil {
		t.Fatal("expected lint error for OLARES_USER in chart template")
	}
	if !strings.Contains(err.Error(), "OLARES_USER") {
		t.Fatalf("error should mention OLARES_USER, got: %v", err)
	}
}

func TestLint_V3_OLARESUserInValuesFails(t *testing.T) {
	dir := writeV3Chart(t, v3DeployHeader+`            - name: SMTP_HOST
              value: "{{ .Values.olaresEnv.OLARES_USER_SMTP_HOST }}"
`)
	if err := os.WriteFile(filepath.Join(dir, "values.yaml"),
		[]byte("bad: OLARES_USER_SMTP_SERVER\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := oac.Lint(dir, oac.SkipResourceCheck(), oac.SkipSameVersionCheck())
	if err == nil {
		t.Fatal("expected lint error for OLARES_USER in values.yaml")
	}
}

func TestValidateManifestContent_V3_InvalidEnvName(t *testing.T) {
	dir := writeV3Chart(t, v3DeployHeader+`            - name: SMTP_HOST
              value: "{{ .Values.olaresEnv.SMTP_HOST }}"
`)
	content, err := os.ReadFile(filepath.Join(dir, "OlaresManifest.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	content = []byte(strings.Replace(string(content),
		"envName: SMTP_HOST", "envName: OLARES_USER_SMTP_SERVER", 1))

	err = oac.ValidateManifestContent(content)
	if err == nil {
		t.Fatal("expected error for envs[].envName with OLARES_USER prefix")
	}
}
