package security

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
)

func renderAppServiceMeshManifests(t *testing.T, values map[string]any) (map[string]map[string]any, error) {
	t.Helper()
	return renderMeshManifests(t, values,
		"../../.olares/config/cluster/deploy/appservice_deploy.yaml",
		"../../.olares/config/cluster/deploy/appservice_api_mesh.yaml",
		"../../../market/.olares/config/cluster/deploy/market_deploy.yaml",
		"../../../market/.olares/config/cluster/deploy/market_backend_mesh.yaml",
		"../../../../apps/.olares/config/user/helm-charts/system-apps/templates/market-frontend-mesh.yaml",
	)
}

func renderMeshManifests(t *testing.T, values map[string]any, templates ...string) (map[string]map[string]any, error) {
	t.Helper()
	return renderMeshManifestsInNamespace(t, "os-framework", false, values, templates...)
}

func renderMeshManifestsInNamespace(t *testing.T, namespace string, onlyReleaseNamespace bool, values map[string]any, templates ...string) (map[string]map[string]any, error) {
	t.Helper()
	c := &chart.Chart{Metadata: &chart.Metadata{Name: "app-service-mesh-test"}}
	for _, file := range templates {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		c.Templates = append(c.Templates, &chart.File{Name: "templates/" + filepath.Base(file), Data: data})
	}
	values["rootPath"] = "/olares"
	input, err := chartutil.ToRenderValues(c, values, chartutil.ReleaseOptions{Name: "mesh-test", Namespace: namespace}, nil)
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
			if onlyReleaseNamespace && meta["namespace"] != namespace {
				continue
			}
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

	activeValues := meshValues(true, true)
	withoutMarketStrict, err := renderAppServiceMeshManifests(t, activeValues)
	if err != nil {
		t.Fatal(err)
	}
	if withoutMarketStrict["HTTPRoute/market-lifecycle-alice"] != nil ||
		withoutMarketStrict["AuthorizationPolicy/market-cluster-apps"] != nil ||
		withoutMarketStrict["AuthorizationPolicy/market-status-user"] == nil {
		t.Fatal("Market without strict ingress may use the existing status aggregation routes only")
	}
	for _, name := range []string{"settings-user-create-alice", "settings-user-create-status-alice", "settings-user-delete-alice"} {
		if withoutMarketStrict["HTTPRoute/"+name] != nil {
			t.Fatalf("user provisioning must be opt-in: %s", name)
		}
	}
	activeValues["appServiceAPIMesh"].(map[string]any)["userProvisioners"] = []string{"alice"}
	activeValues["marketLinkerd"] = true
	activeValues["marketBackendMTLS"] = map[string]any{"prepare": true, "activate": true, "tenants": []string{"alice", "bob"}}
	rollout, err := renderAppServiceMeshManifests(t, activeValues)
	if err != nil {
		t.Fatal(err)
	}
	if rollout["Server/market-backend"] == nil || rollout["Server/market-chart-repository"] == nil {
		t.Fatal("Market rollout needs strict ingress for both API and Chart")
	}
	for _, tenant := range []string{"alice", "bob"} {
		name := "market-lifecycle-" + tenant
		if rollout["HTTPRoute/"+name] != nil || rollout["AuthorizationPolicy/"+name] != nil {
			t.Fatalf("Market rollout must not grant lifecycle access: %s", name)
		}
	}
	activeValues["appServiceAPIMesh"].(map[string]any)["marketLifecycle"] = true
	activeValues["marketBackendMTLS"] = map[string]any{"prepare": true, "activate": false, "tenants": []string{"alice", "bob"}}
	if _, err := renderAppServiceMeshManifests(t, activeValues); err == nil {
		t.Fatal("Market lifecycle routes accepted unactivated Market ingress")
	}
	activeValues["marketBackendMTLS"] = map[string]any{"prepare": true, "activate": true, "tenants": []string{"alice"}}
	if _, err := renderAppServiceMeshManifests(t, activeValues); err == nil {
		t.Fatal("Market lifecycle routes accepted different ingress and API tenant inventories")
	}
	activeValues["marketBackendMTLS"] = map[string]any{"prepare": true, "activate": true, "tenants": []string{"bob", "alice"}}
	active, err := renderAppServiceMeshManifests(t, activeValues)
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
	for _, name := range []string{"settings-user-create-bob", "settings-user-create-status-bob", "settings-user-delete-bob"} {
		if active["HTTPRoute/"+name] != nil || active["AuthorizationPolicy/"+name] != nil {
			t.Fatalf("ordinary tenant received provisioning access: %s", name)
		}
	}
	for _, tc := range []struct{ name, namespace, account, method, path, userHeader, pathType string }{
		{"user-apps-alice", "user-system-alice", "bytetrade-sys-ops", "GET", "/app-service/v1/user-apps/alice", "", "Exact"},
		{"user-apps-bob", "user-system-bob", "bytetrade-sys-ops", "GET", "/app-service/v1/user-apps/bob", "", "Exact"},
		{"users-alice", "user-system-alice", "bytetrade-sys-ops", "GET", "/app-service/v1/users", "alice", "Exact"},
		{"users-bob", "user-system-bob", "bytetrade-sys-ops", "GET", "/app-service/v1/users", "bob", "Exact"},
		{"market-image-info", "os-framework", "market-sa", "POST", "/app-service/v1/apps/image-info", "", "Exact"},
		{"market-status-user", "os-framework", "market-sa", "GET", "/app-service/v1/user-info", "", "Exact"},
		{"market-status-user-resource", "os-framework", "market-sa", "GET", "/app-service/v1/user/resource", "", "Exact"},
		{"market-status-cluster-resource", "os-framework", "market-sa", "GET", "/app-service/v1/cluster/resource", "", "Exact"},
		{"market-cluster-apps", "os-framework", "market-sa", "GET", "/app-service/v1/all/apps", "", "Exact"},
		{"market-status-version", "os-framework", "market-sa", "GET", "/app-service/v1/terminus/version", "", "Exact"},
		{"settings-apps-alice", "user-system-alice", "bytetrade-sys-ops", "GET", "/app-service/v1/apps", "alice", "Exact"},
		{"settings-compute-alice", "user-system-alice", "bytetrade-sys-ops", "GET", "/app-service/v1/compute-resources", "alice", "Exact"},
		{"settings-system-envs-alice", "user-system-alice", "bytetrade-sys-ops", "GET", "/app-service/v1/systemenvs", "alice", "Exact"},
		{"settings-user-envs-alice", "user-system-alice", "bytetrade-sys-ops", "GET", "/app-service/v1/userenvs", "alice", "Exact"},
		{"settings-system-envs-bob", "user-system-bob", "bytetrade-sys-ops", "GET", "/app-service/v1/systemenvs", "bob", "Exact"},
		{"settings-user-envs-bob", "user-system-bob", "bytetrade-sys-ops", "GET", "/app-service/v1/userenvs", "bob", "Exact"},
		{"settings-user-bob", "user-system-bob", "bytetrade-sys-ops", "GET", "/app-service/v1/users/bob", "bob", "Exact"},
		{"settings-app-env-alice", "user-system-alice", "bytetrade-sys-ops", "GET", "^/app-service/v1/apps/[a-z0-9]([a-z0-9-]*[a-z0-9])?/env$", "alice", "RegularExpression"},
		{"settings-app-env-bob", "user-system-bob", "bytetrade-sys-ops", "GET", "^/app-service/v1/apps/[a-z0-9]([a-z0-9-]*[a-z0-9])?/env$", "bob", "RegularExpression"},
		{"control-hub-apps-alice", "user-system-alice", "bytetrade-sys-ops", "GET", "/app-service/v1/control-hub/apps", "alice", "Exact"},
		{"control-hub-apps-bob", "user-system-bob", "bytetrade-sys-ops", "GET", "/app-service/v1/control-hub/apps", "bob", "Exact"},
		{"settings-user-delete-alice", "user-system-alice", "bytetrade-sys-ops", "DELETE", "^/app-service/v1/users/[a-z0-9]([a-z0-9-]*[a-z0-9])?$", "alice", "RegularExpression"},
		{"settings-user-create-alice", "user-system-alice", "bytetrade-sys-ops", "POST", "/app-service/v1/users", "alice", "Exact"},
		{"settings-user-create-status-alice", "user-system-alice", "bytetrade-sys-ops", "GET", "^/app-service/v1/users/[a-z0-9]([a-z0-9-]*[a-z0-9])?/status$", "alice", "RegularExpression"},
		{"settings-app-entrance-reads-alice", "user-system-alice", "bytetrade-sys-ops", "GET", "^/app-service/v1/applications/[a-z0-9]([a-z0-9-]*[a-z0-9])?/(entrances|setup)$", "alice", "RegularExpression"},
		{"settings-app-entrance-reads-bob", "user-system-bob", "bytetrade-sys-ops", "GET", "^/app-service/v1/applications/[a-z0-9]([a-z0-9-]*[a-z0-9])?/(entrances|setup)$", "bob", "RegularExpression"},
		{"settings-app-permissions-alice", "user-system-alice", "bytetrade-sys-ops", "GET", "^/app-service/v1/perms/[a-z0-9]([a-z0-9-]*[a-z0-9])?$", "alice", "RegularExpression"},
		{"settings-app-permissions-bob", "user-system-bob", "bytetrade-sys-ops", "GET", "^/app-service/v1/perms/[a-z0-9]([a-z0-9-]*[a-z0-9])?$", "bob", "RegularExpression"},
		{"settings-app-providers-alice", "user-system-alice", "bytetrade-sys-ops", "GET", "^/app-service/v1/apps/provider-registry/[a-z0-9]([a-z0-9-]*[a-z0-9])?$", "alice", "RegularExpression"},
		{"settings-app-providers-bob", "user-system-bob", "bytetrade-sys-ops", "GET", "^/app-service/v1/apps/provider-registry/[a-z0-9]([a-z0-9-]*[a-z0-9])?$", "bob", "RegularExpression"},
		{"settings-attachments-bob", "user-system-bob", "bytetrade-sys-ops", "GET", "^/app-service/v1/applications/[a-z0-9][a-z0-9-]*/(attachments|attachment-capabilities|device-inventory)$", "bob", "RegularExpression"},
	} {
		route := active["HTTPRoute/"+tc.name]["spec"].(map[string]any)
		match := map[string]any{
			"path": map[string]any{"type": tc.pathType, "value": tc.path}, "method": tc.method,
		}
		if strings.HasPrefix(tc.name, "settings-attachments-") {
			pattern := regexp.MustCompile(tc.path)
			for _, path := range []string{"/app-service/v1/applications/olares-app/attachments", "/app-service/v1/applications/app-3/device-inventory", "/app-service/v1/applications/a/attachment-capabilities"} {
				if !pattern.MatchString(path) {
					t.Fatalf("Settings route misses supported path %s", path)
				}
			}
			for _, path := range []string{"/app-service/v1/applications/a/setup", "/app-service/v1/applications/a/attachments/extra", "/app-service/v1/applications/a/../a/attachments", "/app-service/v1/applications/a%2fb/attachments", "/app-service/v1/applications/a/attachments/", "/v1/applications/a/attachments"} {
				if pattern.MatchString(path) {
					t.Fatalf("Settings route admits unrelated path %s", path)
				}
			}
		}
		if strings.HasPrefix(tc.name, "settings-app-entrance-reads-") {
			pattern := regexp.MustCompile(tc.path)
			for _, path := range []string{"/app-service/v1/applications/olares-app/entrances", "/app-service/v1/applications/a/setup"} {
				if !pattern.MatchString(path) {
					t.Fatalf("entrance policy misses %s", path)
				}
			}
			for _, path := range []string{"/app-service/v1/applications/a/files/setup", "/app-service/v1/applications/a%2fb/setup", "/app-service/v1/applications/a/policy", "/app-service/v1/applications/a/setup/extra"} {
				if pattern.MatchString(path) {
					t.Fatalf("entrance policy admits %s", path)
				}
			}
		}
		if strings.HasPrefix(tc.name, "settings-app-permissions-") || strings.HasPrefix(tc.name, "settings-app-providers-") {
			prefix := "/app-service/v1/perms/"
			if strings.HasPrefix(tc.name, "settings-app-providers-") {
				prefix = "/app-service/v1/apps/provider-registry/"
			}
			pattern := regexp.MustCompile(tc.path)
			for _, app := range []string{"a", "olares-app", "app-3"} {
				if !pattern.MatchString(prefix + app) {
					t.Fatalf("Settings route %s misses app %s", tc.name, app)
				}
			}
			for _, suffix := range []string{"", "a/", "a-/b", "a%2fb", "../a", "provider-registry/type/group/version"} {
				if pattern.MatchString(prefix + suffix) {
					t.Fatalf("Settings route %s admits unrelated path %s", tc.name, prefix+suffix)
				}
			}
		}
		if tc.userHeader != "" {
			match["headers"] = []any{map[string]any{"name": "X-Bfl-User", "type": "Exact", "value": tc.userHeader}}
		}
		wantRules := []any{map[string]any{"matches": []any{match}}}
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
	for _, provisioners := range [][]string{{"outside"}, {"alice", "alice"}} {
		_, err := renderAppServiceMeshManifests(t, map[string]any{"appServiceAPIMesh": map[string]any{
			"prepare": true, "tenants": []string{"alice", "bob"}, "userProvisioners": provisioners,
		}})
		if err == nil {
			t.Fatalf("accepted invalid user provisioners %v", provisioners)
		}
	}
}

func TestControlHubMeshProtectsServiceTargetPort(t *testing.T) {
	const chart = "../../../../apps/.olares/config/user/helm-charts/system-apps/templates/"
	values := map[string]any{
		"bfl":       map[string]any{"username": "alice"},
		"userspace": map[string]any{"appCache": "/cache", "appData": "/app", "userData": "/user"},
	}
	render := func() (map[string]map[string]any, error) {
		return renderMeshManifestsInNamespace(t, "user-space-alice", true, values, chart+"olares-app.yaml", chart+"control-hub-frontend-mesh.yaml")
	}
	baseline, err := render()
	if err != nil {
		t.Fatal(err)
	}
	if baseline["Server/control-hub-frontend"] != nil {
		t.Fatal("Control Hub policy needs the verified gateway identity")
	}
	values["controlHubFrontendGatewayServiceAccount"] = "*"
	if _, err := render(); err == nil {
		t.Fatal("Control Hub accepted a wildcard gateway identity")
	}
	values["controlHubFrontendGatewayServiceAccount"] = "expected-entry"
	secured, err := render()
	if err != nil {
		t.Fatal(err)
	}
	service := secured["Service/control-hub-service"]["spec"].(map[string]any)
	port := service["ports"].([]any)[0].(map[string]any)
	server := secured["Server/control-hub-frontend"]["spec"].(map[string]any)
	if port["port"] != 80 || server["port"] != port["targetPort"] || server["port"] != 82 ||
		server["accessPolicy"] != "deny" || server["proxyProtocol"] != "HTTP/1" ||
		!reflect.DeepEqual(server["podSelector"], map[string]any{"matchLabels": service["selector"]}) {
		t.Fatalf("Control Hub policy does not protect the existing service target: %v / %v", service, server)
	}
	authn := secured["MeshTLSAuthentication/control-hub-frontend-entry"]["spec"].(map[string]any)
	authz := secured["AuthorizationPolicy/control-hub-frontend-entry"]["spec"].(map[string]any)
	if !reflect.DeepEqual(authn["identityRefs"], []any{map[string]any{"kind": "ServiceAccount", "namespace": "os-gateway", "name": "expected-entry"}}) ||
		!reflect.DeepEqual(authz["targetRef"], map[string]any{"group": "policy.linkerd.io", "kind": "Server", "name": "control-hub-frontend"}) ||
		!reflect.DeepEqual(authz["requiredAuthenticationRefs"], []any{map[string]any{"group": "policy.linkerd.io", "kind": "MeshTLSAuthentication", "name": "control-hub-frontend-entry"}}) {
		t.Fatalf("wrong Control Hub gateway authorization: %v / %v", authn, authz)
	}
}

func TestMarketMeshRequiresEnrollmentAndMatchingTenant(t *testing.T) {
	values := func(prepare, activate, inject bool, tenants ...string) map[string]any {
		return map[string]any{
			"marketLinkerd": inject,
			"marketBackendMTLS": map[string]any{
				"prepare": prepare, "activate": activate, "tenants": tenants,
			},
			"appServiceAPIMesh":                   map[string]any{"prepare": true, "activate": true, "tenants": []string{"alice", "bob"}},
			"bfl":                                 map[string]any{"username": "alice"},
			"marketFrontendGatewayServiceAccount": "expected-entry",
		}
	}
	missingReader := values(true, true, true, "alice")
	delete(missingReader, "appServiceAPIMesh")
	for _, input := range []map[string]any{
		missingReader,
		values(false, true, true, "alice"),
		values(true, true, false, "alice"),
		values(true, false, true),
		values(true, false, true, "alice", "alice"),
		values(true, false, true, "*"),
	} {
		if _, err := renderAppServiceMeshManifests(t, input); err == nil {
			t.Fatalf("accepted incomplete or unsafe config %v", input)
		}
	}
	var marketImage any
	for _, activate := range []bool{false, true} {
		resources, err := renderAppServiceMeshManifests(t, values(true, activate, true, "alice", "bob"))
		if err != nil {
			t.Fatal(err)
		}
		const enrollment = "traffic.olares.io/market-backend-mtls"
		server := resources["Server/market-backend"]["spec"].(map[string]any)
		if server["port"] != 8080 || server["accessPolicy"] != "deny" ||
			!reflect.DeepEqual(server["podSelector"], map[string]any{"matchLabels": map[string]any{"app": "appstore", enrollment: "required"}}) {
			t.Fatalf("wrong backend policy: %v", server)
		}
		template := resources["Deployment/market-deployment"]["spec"].(map[string]any)["template"].(map[string]any)
		pod := template["metadata"].(map[string]any)
		containers := template["spec"].(map[string]any)["containers"].([]any)
		for _, entry := range containers {
			container := entry.(map[string]any)
			if container["name"] == "appstore-backend" {
				wantPorts := []any{map[string]any{"containerPort": 8080}}
				if activate {
					wantPorts = append(wantPorts, map[string]any{"containerPort": 82})
				}
				if !reflect.DeepEqual(container["ports"], wantPorts) {
					t.Fatalf("Market listeners must be declared for Linkerd Server selection: %v", container["ports"])
				}
				if !activate {
					marketImage = container["image"]
				}
				if marketImage == nil || container["image"] != marketImage {
					t.Fatalf("Market mesh changed the backend image: %v", container["image"])
				}
			}
		}
		if marketImage == nil {
			t.Fatal("Market backend image is missing")
		}
		if (pod["labels"].(map[string]any)[enrollment] == "required") != activate ||
			(pod["annotations"].(map[string]any)["config.linkerd.io/default-inbound-policy"] == "deny") != activate {
			t.Fatalf("incorrect workload activation: %v", pod)
		}
		ports := resources["Service/appstore-svc"]["spec"].(map[string]any)["ports"].([]any)
		wantPorts := []any{map[string]any{"name": "appstore-backend", "protocol": "TCP", "port": 81, "targetPort": 8080}}
		if !activate {
			wantPorts = append(wantPorts, map[string]any{"name": "appstore-websocket", "protocol": "TCP", "port": 40010, "targetPort": 40010})
		}
		if !reflect.DeepEqual(ports, wantPorts) {
			t.Fatalf("unexpected Service port exposure: %v", ports)
		}
		if resources["ValidatingAdmissionPolicyBinding/olares-market-backend-mtls-pods"] == nil {
			t.Fatal("missing Market pod admission binding")
		}
		for _, tenant := range []string{"alice", "bob"} {
			name := "market-provider-" + tenant
			route := resources["HTTPRoute/"+name]["spec"].(map[string]any)
			wantMatch := map[string]any{
				"path":    map[string]any{"type": "PathPrefix", "value": "/app-store/api/v2"},
				"headers": []any{map[string]any{"name": "X-Bfl-User", "type": "Exact", "value": tenant}},
			}
			if !reflect.DeepEqual(route["rules"], []any{map[string]any{"matches": []any{wantMatch}}}) ||
				!reflect.DeepEqual(route["parentRefs"], []any{map[string]any{"group": "policy.linkerd.io", "kind": "Server", "name": "market-backend"}}) {
				t.Fatalf("tenant route is broader than its contract: %v", route)
			}
			authn := resources["MeshTLSAuthentication/"+name]["spec"].(map[string]any)
			wantIdentity := map[string]any{"kind": "ServiceAccount", "name": "bytetrade-sys-ops", "namespace": "user-system-" + tenant}
			if !reflect.DeepEqual(authn, map[string]any{"identityRefs": []any{wantIdentity}}) {
				t.Fatalf("incorrect tenant identity: %v", authn)
			}
			authz := resources["AuthorizationPolicy/"+name]["spec"].(map[string]any)
			wantPolicy := map[string]any{
				"targetRef":                  map[string]any{"group": "gateway.networking.k8s.io", "kind": "HTTPRoute", "name": name},
				"requiredAuthenticationRefs": []any{map[string]any{"group": "policy.linkerd.io", "kind": "MeshTLSAuthentication", "name": name}},
			}
			if !reflect.DeepEqual(authz, wantPolicy) {
				t.Fatalf("incorrect authorization policy: %v", authz)
			}
		}
		if frontend := resources["Server/market-frontend"]["spec"].(map[string]any); frontend["port"] != 90 || frontend["accessPolicy"] != "deny" {
			t.Fatalf("incorrect frontend boundary: %v", frontend)
		}
		authn := resources["MeshTLSAuthentication/market-frontend-entry"]["spec"].(map[string]any)
		if !reflect.DeepEqual(authn["identityRefs"], []any{map[string]any{"kind": "ServiceAccount", "name": "expected-entry", "namespace": "os-gateway"}}) {
			t.Fatalf("incorrect Entry identity: %v", authn)
		}
	}
}
