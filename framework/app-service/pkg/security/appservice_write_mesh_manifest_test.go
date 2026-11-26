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
	lifecycleOnly := render(true, false)
	envOnly := render(false, true)
	for _, tenant := range tenants {
		t.Run(tenant, func(t *testing.T) {
			lifecycle := "market-lifecycle-" + tenant
			for _, kind := range []string{"HTTPRoute", "AuthorizationPolicy", "MeshTLSAuthentication"} {
				if prepared[kind+"/"+lifecycle] != nil || envOnly[kind+"/"+lifecycle] != nil {
					t.Fatalf("lifecycle authorization was generated before lifecycle activation: %s", kind)
				}
				if active[kind+"/market-lifecycle-unregistered"] != nil {
					t.Fatalf("lifecycle authorization was generated for an unregistered tenant: %s", kind)
				}
				name := kind + "/settings-system-envs-save-" + tenant
				if prepared[name] != nil || lifecycleOnly[name] != nil {
					t.Fatalf("environment writes must be independently enabled: %s", name)
				}
			}

			const appPath = `^/app-service/v1/apps/[a-z0-9]([a-z0-9-]*[a-z0-9])?`
			operations := []struct {
				method, action string
				query          []any
			}{
				{method: "POST", action: "install"},
				{method: "GET", action: "status"},
				{method: "POST", action: "cancel", query: []any{map[string]any{"type": "Exact", "name": "type", "value": "operate"}}},
			}
			var matches []any
			for _, operation := range operations {
				expected := appPath + "/" + operation.action + "$"
				path := regexp.MustCompile(expected)
				for _, name := range []string{"router", "a", "app-3"} {
					if !path.MatchString("/app-service/v1/apps/" + name + "/" + operation.action) {
						t.Fatalf("%s route misses app %s", operation.action, name)
					}
				}
				for _, invalid := range []string{"a/uninstall", "a-/" + operation.action, "a%2fb/" + operation.action, "a/" + operation.action + "/extra", "a/../b/" + operation.action} {
					if path.MatchString("/app-service/v1/apps/" + invalid) {
						t.Fatalf("%s route admits unrelated path %s", operation.action, invalid)
					}
				}
				match := appServiceTenantMatch(tenant, operation.method, "RegularExpression", expected)
				if operation.query != nil {
					match["queryParams"] = operation.query
				}
				matches = append(matches, match)
			}
			checkAppServiceCapability(t, active, lifecycle, "os-framework", "market-sa", matches)
			checkAppServiceCapability(t, envOnly, "settings-system-envs-save-"+tenant, "user-system-"+tenant, "bytetrade-sys-ops",
				[]any{appServiceTenantMatch(tenant, "PUT", "Exact", "/app-service/v1/systemenvs")})
		})
	}
}

func appServiceTenantMatch(tenant, method, pathType, path string) map[string]any {
	return map[string]any{
		"method":  method,
		"path":    map[string]any{"type": pathType, "value": path},
		"headers": []any{map[string]any{"type": "Exact", "name": "X-Bfl-User", "value": tenant}},
	}
}

func checkAppServiceCapability(t *testing.T, resources map[string]map[string]any, name, namespace, account string, matches []any) {
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

	if !reflect.DeepEqual(route["rules"], []any{map[string]any{"matches": matches}}) ||
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
