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
	for _, kind := range []string{"HTTPRoute", "AuthorizationPolicy", "MeshTLSAuthentication"} {
		if name := kind + "/market-chart-admins"; prepared[name] != nil || envOnly[name] != nil {
			t.Fatalf("owner/admin list authorization was generated before Market lifecycle activation: %s", name)
		}
	}
	checkAppServiceCapability(t, active, "market-chart-admins", "os-framework", "market-sa", []any{map[string]any{
		"method": "GET",
		"path":   map[string]any{"type": "Exact", "value": "/app-service/v1/users/admins"},
	}})
	for _, tenant := range tenants {
		t.Run(tenant, func(t *testing.T) {
			lifecycle := "market-lifecycle-" + tenant
			runtime := "market-runtime-" + tenant
			for _, kind := range []string{"HTTPRoute", "AuthorizationPolicy", "MeshTLSAuthentication"} {
				if prepared[kind+"/"+lifecycle] != nil || envOnly[kind+"/"+lifecycle] != nil {
					t.Fatalf("lifecycle authorization was generated before lifecycle activation: %s", kind)
				}
				if prepared[kind+"/"+runtime] != nil || envOnly[kind+"/"+runtime] != nil {
					t.Fatalf("runtime authorization was generated before lifecycle activation: %s", kind)
				}
				if active[kind+"/market-lifecycle-unregistered"] != nil || active[kind+"/market-runtime-unregistered"] != nil {
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
			checkAppServiceCapability(t, active, runtime, "os-framework", "market-sa", []any{
				appServiceTenantMatch(tenant, "POST", "RegularExpression", appPath+"/suspend$"),
				appServiceTenantMatch(tenant, "POST", "RegularExpression", appPath+"/resume$"),
			})
			checkAppServiceCapability(t, envOnly, "settings-system-envs-save-"+tenant, "user-system-"+tenant, "bytetrade-sys-ops",
				[]any{appServiceTenantMatch(tenant, "PUT", "Exact", "/app-service/v1/systemenvs")})
		})
	}
}

func TestAppServiceAppStatusMeshAccess(t *testing.T) {
	const statusPath = `^/app-service/v1/apps/[a-z0-9]([a-z0-9-]*[a-z0-9])?/status$`
	for _, mode := range []struct {
		name            string
		marketLifecycle bool
	}{
		{name: "tenant proxy only"},
		{name: "strict Market and tenant proxy", marketLifecycle: true},
	} {
		t.Run(mode.name, func(t *testing.T) {
			tenants := []string{"alice", "bob"}
			resources, err := renderAppServiceMeshManifests(t, map[string]any{
				"appServiceAPIMesh": map[string]any{"prepare": true, "activate": true, "tenants": tenants, "marketLifecycle": mode.marketLifecycle},
				"marketBackendMTLS": map[string]any{"prepare": mode.marketLifecycle, "activate": mode.marketLifecycle, "tenants": tenants},
				"marketLinkerd":     mode.marketLifecycle,
			})
			if err != nil {
				t.Fatal(err)
			}
			var marketIdentity []map[string]any
			if mode.marketLifecycle {
				marketIdentity = append(marketIdentity, map[string]any{"kind": "ServiceAccount", "namespace": "os-framework", "name": "market-sa"})
			}
			for _, tenant := range tenants {
				name := "market-app-progress-" + tenant
				match := appServiceTenantMatch(tenant, "GET", "RegularExpression", statusPath)
				checkAppServiceCapability(t, resources, name, "user-system-"+tenant, "bytetrade-sys-ops", []any{match}, marketIdentity...)
				for key, resource := range resources {
					if resource["kind"] != "HTTPRoute" || key == "HTTPRoute/"+name {
						continue
					}
					rules, _ := resource["spec"].(map[string]any)["rules"].([]any)
					for _, rule := range rules {
						matches, _ := rule.(map[string]any)["matches"].([]any)
						for _, otherMatch := range matches {
							if reflect.DeepEqual(otherMatch, match) {
								t.Fatalf("%s competes with %s for the same tenant status request", key, name)
							}
						}
					}
				}
			}
			for _, kind := range []string{"HTTPRoute", "MeshTLSAuthentication", "AuthorizationPolicy"} {
				if resources[kind+"/market-app-progress-unregistered"] != nil {
					t.Fatalf("%s grants status access to an unregistered tenant", kind)
				}
			}
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

func checkAppServiceCapability(t *testing.T, resources map[string]map[string]any, name, namespace, account string, matches []any, additionalIdentities ...map[string]any) {
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
	wantIdentities := []any{map[string]any{"kind": "ServiceAccount", "namespace": namespace, "name": account}}
	for _, identity := range additionalIdentities {
		wantIdentities = append(wantIdentities, identity)
	}
	if !reflect.DeepEqual(spec("MeshTLSAuthentication"), map[string]any{"identityRefs": wantIdentities}) {
		t.Fatalf("wrong authenticated workload for %s: %v", name, spec("MeshTLSAuthentication"))
	}
	authorization := spec("AuthorizationPolicy")
	if !reflect.DeepEqual(authorization["targetRef"], map[string]any{"group": "gateway.networking.k8s.io", "kind": "HTTPRoute", "name": name}) ||
		!reflect.DeepEqual(authorization["requiredAuthenticationRefs"], []any{map[string]any{"group": "policy.linkerd.io", "kind": "MeshTLSAuthentication", "name": name}}) {
		t.Fatalf("wrong authorization for %s: %v", name, authorization)
	}
}
