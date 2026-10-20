package chart

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFromComposeNoInterpolateResolvesPortsAndKeepsEnvironment(t *testing.T) {
	root := t.TempDir()
	composePath := filepath.Join(root, "compose.yml")
	compose := []byte(`x-common-env: &common-env
  SHARED_TOKEN: ${SHARED_TOKEN}
services:
  web:
    image: nginx:1.27
    ports:
      - "${WEB_PORT:-8088}:8088"
    environment:
      <<: *common-env
      SECRET_KEY: ${SECRET_KEY}
      LISTEN: 0.0.0.0:${INNER_PORT:-8088}
  worker:
    image: alpine:3.20
    command: ["sleep", "3600"]
    environment:
      - QUEUE=${QUEUE_NAME}
`)
	if err := os.WriteFile(composePath, compose, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SECRET_KEY", "from-the-shell")

	chartDir := filepath.Join(root, "web")
	result, err := FromCompose(Options{
		ComposeFiles:  []string{composePath},
		OutputDir:     chartDir,
		Name:          "web",
		NoInterpolate: true,
	})
	if err != nil {
		t.Fatalf("FromCompose() error: %v", err)
	}
	if result.EntrancePort != 8088 {
		t.Fatalf("entrance port = %d, want 8088 resolved from ${WEB_PORT:-8088}", result.EntrancePort)
	}

	web := readTemplate(t, chartDir, "deployment-web.yaml")
	for _, want := range []string{"value: ${SECRET_KEY}", "value: ${SHARED_TOKEN}", "value: 0.0.0.0:${INNER_PORT:-8088}", "containerPort: 8088"} {
		if !strings.Contains(web, want) {
			t.Fatalf("deployment-web.yaml lacks %q:\n%s", want, web)
		}
	}
	if strings.Contains(web, "from-the-shell") {
		t.Fatalf("--no-interpolate baked a shell value into the chart:\n%s", web)
	}
	worker := readTemplate(t, chartDir, "deployment-worker.yaml")
	if !strings.Contains(worker, "value: ${QUEUE_NAME}") {
		t.Fatalf("list-form environment lost its reference:\n%s", worker)
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".olares-compose-") {
			t.Fatalf("escaped compose copy %q was left behind", e.Name())
		}
	}
}
