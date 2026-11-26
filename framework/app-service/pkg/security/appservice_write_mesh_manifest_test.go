package security

import (
	"reflect"
	"regexp"
	"testing"
)

func TestAppServiceMeshScopedWrites(t *testing.T) {
	tenants := []string{"alice", "bob"}
	values := map[string]any{
		"appServiceAPIMesh": map[string]any{"prepare": true, "activate": true, "tenants": tenants},
		"marketBackendMTLS": map[string]any{"prepare": true, "activate": true, "tenants": tenants},
		"marketLinkerd":     true,
	}
	render := func(lifecycle, systemEnvWrites bool) map[string]map[string]any {
		t.Helper()
		values["appServiceAPIMesh"].(map[string]any)["marketLifecycle"] = lifecycle
		values["appServiceAPIMesh"].(map[string]any)["systemEnvWrites"] = systemEnvWrites
		resources, err := renderAppServiceMeshManifests(t, values)
		if err != nil {
			t.Fatal(err)
		}
		return resources
	}
	prepared, active := render(false, false), render(true, true)
	cancelOnly := render(true, false)
	envOnly := render(false, true)
	for _, tenant := range tenants {
		t.Run(tenant, func(t *testing.T) {
			cancel := "market-app-cancel-" + tenant
			for _, kind := range []string{"HTTPRoute", "AuthorizationPolicy", "MeshTLSAuthentication"} {
				if prepared[kind+"/"+cancel] != nil || envOnly[kind+"/"+cancel] != nil {
					t.Fatalf("cancel authorization was generated before lifecycle activation: %s", kind)
				}
				if active[kind+"/market-app-cancel-unregistered"] != nil {
					t.Fatalf("cancel authorization was generated for an unregistered tenant: %s", kind)
				}
				name := kind + "/settings-system-envs-save-" + tenant
				if prepared[name] != nil || cancelOnly[name] != nil {
					t.Fatalf("environment writes must be independently enabled: %s", name)
				}
			}

			const cancelPath = `^/app-service/v1/apps/[a-z0-9]([a-z0-9-]*[a-z0-9])?/cancel$`
			path := regexp.MustCompile(cancelPath)
			for _, name := range []string{"router", "a", "app-3"} {
				if !path.MatchString("/app-service/v1/apps/" + name + "/cancel") {
					t.Fatalf("cancel route does not match app %s", name)
				}
			}
			for _, suffix := range []string{"a/uninstall", "a/install", "a-/cancel", "a%2fb/cancel", "a/cancel/extra", "a/../b/cancel"} {
				if path.MatchString("/app-service/v1/apps/" + suffix) {
					t.Fatalf("cancel route admits unrelated path %s", suffix)
				}
			}
			checkAppServiceWrite(t, active, cancel, "os-framework", "market-sa", tenant, "POST", "RegularExpression", cancelPath,
				[]any{map[string]any{"type": "Exact", "name": "type", "value": "operate"}})
			checkAppServiceWrite(t, envOnly, "settings-system-envs-save-"+tenant, "user-system-"+tenant,
				"bytetrade-sys-ops", tenant, "PUT", "Exact", "/app-service/v1/systemenvs", nil)
		})
	}
}

func checkAppServiceWrite(t *testing.T, resources map[string]map[string]any, name, namespace, account, tenant, method, pathType, path string, query []any) {
	t.Helper()
	spec := func(kind string) map[string]any {
		t.Helper()
		value, ok := resources[kind+"/"+name]["spec"].(map[string]any)
		if !ok {
			t.Fatalf("missing %s/%s", kind, name)
		}
		return value
	}
	route := spec("HTTPRoute")
	match := map[string]any{
		"method":  method,
		"path":    map[string]any{"type": pathType, "value": path},
		"headers": []any{map[string]any{"type": "Exact", "name": "X-Bfl-User", "value": tenant}},
	}
	if query != nil {
		match["queryParams"] = query
	}
	if !reflect.DeepEqual(route["rules"], []any{map[string]any{"matches": []any{match}}}) ||
		!reflect.DeepEqual(route["parentRefs"], []any{map[string]any{"group": "policy.linkerd.io", "kind": "Server", "name": "app-service-api"}}) {
		t.Fatalf("wrong HTTP scope for %s: %v", name, route)
	}
	wantIdentity := map[string]any{"kind": "ServiceAccount", "namespace": namespace, "name": account}
	if !reflect.DeepEqual(spec("MeshTLSAuthentication"), map[string]any{"identityRefs": []any{wantIdentity}}) {
		t.Fatalf("wrong authenticated workload for %s: %v", name, spec("MeshTLSAuthentication"))
	}
	authorization := spec("AuthorizationPolicy")
	if !reflect.DeepEqual(authorization["targetRef"], map[string]any{"group": "gateway.networking.k8s.io", "kind": "HTTPRoute", "name": name}) ||
		!reflect.DeepEqual(authorization["requiredAuthenticationRefs"], []any{map[string]any{"group": "policy.linkerd.io", "kind": "MeshTLSAuthentication", "name": name}}) {
		t.Fatalf("wrong authorization for %s: %v", name, authorization)
	}
}
