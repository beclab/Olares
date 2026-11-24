package security

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
)

func TestBFLRefreshIdentityIsOptInAndGrantsNoAPICredentials(t *testing.T) {
	const ns = "user-space-example"
	const file = "templates/bfl-cert-refresh-serviceaccount.yaml"
	data, err := os.ReadFile(filepath.Join("../../..", "bfl/.olares/config/launcher", file))
	if err != nil {
		t.Fatal(err)
	}
	c := &chart.Chart{Metadata: &chart.Metadata{Name: "launcher"}, Templates: []*chart.File{{Name: file, Data: data}}}
	render := func(enabled bool) string {
		t.Helper()
		values, err := chartutil.ToRenderValues(c, map[string]any{"bfl": map[string]any{"prepareCertRefreshMesh": enabled}},
			chartutil.ReleaseOptions{Name: "launcher", Namespace: ns}, nil)
		if err != nil {
			t.Fatal(err)
		}
		files, err := engine.Render(c, values)
		if err != nil {
			t.Fatal(err)
		}
		return strings.TrimSpace(files["launcher/"+file])
	}
	if doc := render(false); doc != "" {
		t.Fatalf("unexpected ServiceAccount before migration: %s", doc)
	}
	var got map[string]any
	if err := yaml.Unmarshal([]byte(render(true)), &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"apiVersion": "v1", "kind": "ServiceAccount", "automountServiceAccountToken": false,
		"metadata": map[string]any{"name": "bfl-cert-refresh", "namespace": ns},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected certificate refresh identity: %v", got)
	}
}
