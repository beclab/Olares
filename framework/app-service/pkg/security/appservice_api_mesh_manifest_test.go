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

func renderAppServiceMeshManifests(t *testing.T, values map[string]any) (map[string]map[string]any, error) {
	t.Helper()
	c := &chart.Chart{Metadata: &chart.Metadata{Name: "app-service-mesh-test"}}
	for _, file := range []string{
		"../../.olares/config/cluster/deploy/appservice_deploy.yaml",
		"../../.olares/config/cluster/deploy/appservice_api_mesh.yaml",
		"../../../market/.olares/config/cluster/deploy/market_deploy.yaml",
	} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		c.Templates = append(c.Templates, &chart.File{Name: "templates/" + filepath.Base(file), Data: data})
	}
	values["rootPath"] = "/olares"
	input, err := chartutil.ToRenderValues(c, values, chartutil.ReleaseOptions{Name: "mesh-test", Namespace: "os-framework"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	files, err := engine.Render(c, input)
	if err != nil {
		return nil, err
	}
	resources := make(map[string]map[string]any)
	for _, contents := range files {
		decoder := yaml.NewDecoder(strings.NewReader(contents))
		for {
			var item map[string]any
			if err := decoder.Decode(&item); err == io.EOF {
				break
			} else if err != nil {
				t.Fatal(err)
			}
			if len(item) == 0 {
				continue
			}
			meta := item["metadata"].(map[string]any)
			key := item["kind"].(string) + "/" + meta["name"].(string)
			if _, ok := resources[key]; ok {
				t.Fatalf("duplicate rendered resource: %s", key)
			}
			resources[key] = item
		}
	}
	return resources, nil
}

func TestAppServiceAPIMeshRolloutAndPortBoundaries(t *testing.T) {
	meshValues := func(prepare, activate bool) map[string]any {
		return map[string]any{"appServiceAPIMesh": map[string]any{
			"prepare": prepare, "activate": activate, "tenants": []string{"alice", "bob"},
		}}
	}
	podMetadata := func(docs map[string]map[string]any, key string) map[string]any {
		t.Helper()
		return docs[key]["spec"].(map[string]any)["template"].(map[string]any)["metadata"].(map[string]any)
	}

	baseline, err := renderAppServiceMeshManifests(t, map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if baseline["Server/app-service-api"] != nil {
		t.Fatal("API mesh must be opt-in")
	}
	if _, err := renderAppServiceMeshManifests(t, meshValues(false, true)); err == nil {
		t.Fatal("activated API workload without prepared policies")
	}
	preparing := meshValues(true, false)
	preparing["marketLinkerd"] = true
	prepared, err := renderAppServiceMeshManifests(t, preparing)
	if err != nil {
		t.Fatal(err)
	}
	const app = "StatefulSet/app-service"
	const market = "Deployment/market-deployment"
	if !reflect.DeepEqual(podMetadata(prepared, app), podMetadata(baseline, app)) {
		t.Fatal("policy preparation changes the receiving workload")
	}
	if a := podMetadata(prepared, market)["annotations"].(map[string]any); a["linkerd.io/inject"] != "enabled" ||
		a["config.linkerd.io/proxy-enable-native-sidecar"] != "true" {
		t.Fatal("Market needs its proxy before its existing init checks")
	}

	active, err := renderAppServiceMeshManifests(t, meshValues(true, true))
	if err != nil {
		t.Fatal(err)
	}
	metadata := podMetadata(active, app)
	annotations := metadata["annotations"].(map[string]any)
	if annotations["linkerd.io/inject"] != "enabled" ||
		annotations["config.linkerd.io/proxy-enable-native-sidecar"] != "true" ||
		annotations["config.linkerd.io/default-inbound-policy"] != "deny" ||
		annotations["config.linkerd.io/skip-inbound-ports"] != "8433,8444" || annotations["config.linkerd.io/skip-outbound-ports"] != nil {
		t.Fatalf("unexpected port posture: %v", annotations)
	}
	if !reflect.DeepEqual(active["Service/app-service"]["spec"], baseline["Service/app-service"]["spec"]) {
		t.Fatal("activation changes the Webhook or JWKS Service")
	}
	server := active["Server/app-service-api"]["spec"].(map[string]any)
	wantSelector := map[string]any{"matchLabels": map[string]any{
		"tier": "app-service", "traffic.olares.io/app-service-api-mtls": "required",
	}}
	if server["accessPolicy"] != "deny" || server["port"] != 6755 ||
		!reflect.DeepEqual(server["podSelector"], wantSelector) ||
		metadata["labels"].(map[string]any)["traffic.olares.io/app-service-api-mtls"] != "required" {
		t.Fatalf("unexpected API server boundary: %v", server)
	}
	if active["ValidatingAdmissionPolicy/olares-app-service-api-mtls-pods"]["spec"].(map[string]any)["failurePolicy"] != "Fail" {
		t.Fatal("enrolled Pods must fail closed")
	}
	for _, tc := range []struct{ name, namespace, account, method, path string }{
		{"user-apps-alice", "user-system-alice", "bytetrade-sys-ops", "GET", "/app-service/v1/user-apps/alice"},
		{"user-apps-bob", "user-system-bob", "bytetrade-sys-ops", "GET", "/app-service/v1/user-apps/bob"},
		{"market-image-info", "os-framework", "market-sa", "POST", "/app-service/v1/apps/image-info"},
	} {
		route := active["HTTPRoute/"+tc.name]["spec"].(map[string]any)
		wantRules := []any{map[string]any{"matches": []any{map[string]any{
			"path": map[string]any{"type": "Exact", "value": tc.path}, "method": tc.method,
		}}}}
		wantParents := []any{map[string]any{"group": "policy.linkerd.io", "kind": "Server", "name": "app-service-api"}}
		if !reflect.DeepEqual(route["rules"], wantRules) || !reflect.DeepEqual(route["parentRefs"], wantParents) {
			t.Fatalf("incorrect request scope for %s: %v", tc.name, route)
		}
		wantIdentity := []any{map[string]any{"kind": "ServiceAccount", "namespace": tc.namespace, "name": tc.account}}
		if !reflect.DeepEqual(active["MeshTLSAuthentication/"+tc.name]["spec"].(map[string]any)["identityRefs"], wantIdentity) {
			t.Fatalf("incorrect tenant or identity for %s", tc.name)
		}
		authz := active["AuthorizationPolicy/"+tc.name]["spec"].(map[string]any)
		if !reflect.DeepEqual(authz["targetRef"], map[string]any{"group": "gateway.networking.k8s.io", "kind": "HTTPRoute", "name": tc.name}) ||
			!reflect.DeepEqual(authz["requiredAuthenticationRefs"], []any{map[string]any{
				"group": "policy.linkerd.io", "kind": "MeshTLSAuthentication", "name": tc.name,
			}}) {
			t.Fatalf("incorrect authorization references for %s: %v", tc.name, authz)
		}
	}
}

func TestAppServiceMeshRequiresExplicitValidTenants(t *testing.T) {
	for _, tenants := range [][]string{nil, {"other/tenant"}, {"tenant", "tenant"}} {
		_, err := renderAppServiceMeshManifests(t, map[string]any{"appServiceAPIMesh": map[string]any{
			"prepare": true, "tenants": tenants,
		}})
		if err == nil {
			t.Fatalf("accepted invalid tenant inventory %v", tenants)
		}
	}
}
