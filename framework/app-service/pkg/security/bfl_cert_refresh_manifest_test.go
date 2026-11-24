package security

import (
	"io"
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

func TestBFLAPIMTLSActivatesOnlyForItsTenantAndDefaultsToDeny(t *testing.T) {
	const ns = "user-space-example"
	const file = "templates/bfl-api-mtls.yaml"
	data, err := os.ReadFile(filepath.Join("../../..", "bfl/.olares/config/launcher", file))
	if err != nil {
		t.Fatal(err)
	}
	c := &chart.Chart{Metadata: &chart.Metadata{Name: "launcher"}, Templates: []*chart.File{{Name: file, Data: data}}}
	render := func(namespace string, enabled bool) (string, error) {
		values, err := chartutil.ToRenderValues(c,
			map[string]any{"bfl": map[string]any{"username": "example", "enforceAPIMesh": enabled}},
			chartutil.ReleaseOptions{Name: "launcher", Namespace: namespace}, nil)
		if err != nil {
			return "", err
		}
		files, err := engine.Render(c, values)
		return strings.TrimSpace(files["launcher/"+file]), err
	}
	if doc, err := render(ns, false); err != nil || doc != "" {
		t.Fatalf("BFL policy must be opt-in: %s %v", doc, err)
	}
	if _, err := render("user-space-other", true); err == nil {
		t.Fatal("refusing to pair a BFL API with another tenant's identities")
	}
	doc, err := render(ns, true)
	if err != nil {
		t.Fatal(err)
	}
	docs := make(map[string]map[string]any)
	decoder := yaml.NewDecoder(strings.NewReader(doc))
	for {
		var m map[string]any
		if err := decoder.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		meta := m["metadata"].(map[string]any)
		if meta["namespace"] != ns {
			t.Fatalf("unexpected policy namespace: %v", meta)
		}
		key := m["kind"].(string) + "/" + meta["name"].(string)
		if _, exists := docs[key]; exists {
			t.Fatalf("duplicate BFL policy: %s", key)
		}
		docs[key] = m["spec"].(map[string]any)
	}
	server := docs["Server/bfl-api"]
	if len(docs) != 31 {
		t.Fatalf("expected one Server and ten scoped API routes with their authorizations, got %d resources", len(docs))
	}
	wantSelector := map[string]any{"matchLabels": map[string]any{"tier": "bfl", "traffic.olares.io/bfl-api-mtls": "required"}}
	if server["port"] != 8080 || server["accessPolicy"] != "deny" ||
		!reflect.DeepEqual(server["podSelector"], wantSelector) {
		t.Fatalf("unexpected BFL Server boundary: %v", server)
	}
	for _, tc := range []struct {
		name, path, method string
		callers            []string
	}{
		{"frontend-modules", "/bfl/settings/v1alpha1", "", []string{ns + "/system-frontend", ns + "/bfl-wizard"}},
		{"frontend-backend", "/bfl/backend/v1/terminus-info", "GET", []string{ns + "/system-frontend", ns + "/bfl-wizard"}},
		{"frontend-apps", "/bfl/backend/v1/myapps", "POST", []string{ns + "/system-frontend", ns + "/bfl-wizard"}},
		{"user-info", "/bfl/backend/v1/user-info", "GET", []string{ns + "/system-frontend", ns + "/bfl-wizard", "os-framework/os-internal"}},
		{"public-info", "/bfl/info/v1/olares-info", "OPTIONS", []string{ns + "/system-frontend", ns + "/bfl-wizard", ns + "/bfl-login", "os-framework/bfl-vault"}},
		{"login-cluster", "/bfl/monitor/v1alpha1/cluster", "GET", []string{ns + "/system-frontend", ns + "/bfl-wizard", ns + "/bfl-login"}},
		{"login-password", "/bfl/iam/v1alpha1/users", "PUT", []string{ns + "/system-frontend", ns + "/bfl-wizard", ns + "/bfl-login"}},
		{"vault-binding", "/bfl/settings/v1alpha1/binding-zone", "POST", []string{ns + "/system-frontend", ns + "/bfl-wizard", "os-framework/bfl-vault"}},
		{"tapr-app-policy", "/bfl/settings/v1alpha1/applications", "POST", []string{ns + "/system-frontend", ns + "/bfl-wizard", "user-system-example/tapr-images"}},
		{"cert-refresh", "/bfl/backend/v1/re-download-cert", "GET", []string{ns + "/bfl-cert-refresh"}},
	} {
		name := "bfl-api-" + tc.name
		var callers []string
		for _, ref := range docs["MeshTLSAuthentication/"+name]["identityRefs"].([]any) {
			r := ref.(map[string]any)
			if r["kind"] != "ServiceAccount" {
				t.Fatalf("unexpected identity type on %s: %v", name, r)
			}
			callers = append(callers, r["namespace"].(string)+"/"+r["name"].(string))
		}
		if !reflect.DeepEqual(callers, tc.callers) {
			t.Fatalf("unexpected callers for %s: %v", name, callers)
		}
		matched := false
		wantParent := []any{map[string]any{"group": "policy.linkerd.io", "kind": "Server", "name": "bfl-api"}}
		if !reflect.DeepEqual(docs["HTTPRoute/"+name]["parentRefs"], wantParent) {
			t.Fatalf("route %s must attach to the BFL Server", name)
		}
		for _, rule := range docs["HTTPRoute/"+name]["rules"].([]any) {
			for _, match := range rule.(map[string]any)["matches"].([]any) {
				m := match.(map[string]any)
				path := m["path"].(map[string]any)
				method, _ := m["method"].(string)
				matched = matched || (path["value"] == tc.path && method == tc.method)
				if path["value"] == "/" || path["value"] == "/bfl" || path["value"] == "/bfl/backend/v1" ||
					(path["value"] == "/bfl/backend/v1/re-download-cert" && (tc.name != "cert-refresh" || path["type"] != "Exact")) {
					t.Fatalf("authorization route %s would include unrelated API requests: %v", name, m)
				}
			}
		}
		if !matched {
			t.Fatalf("missing BFL route %s %s for %s", tc.method, tc.path, name)
		}
		wantTarget := map[string]any{"group": "gateway.networking.k8s.io", "kind": "HTTPRoute", "name": name}
		wantRefs := []any{map[string]any{"group": "policy.linkerd.io", "kind": "MeshTLSAuthentication", "name": name}}
		authz := docs["AuthorizationPolicy/"+name]
		if !reflect.DeepEqual(authz["targetRef"], wantTarget) || !reflect.DeepEqual(authz["requiredAuthenticationRefs"], wantRefs) {
			t.Fatalf("unexpected route authorization %s: %v", name, authz)
		}
	}
}
