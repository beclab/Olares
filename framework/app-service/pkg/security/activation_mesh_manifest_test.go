package security

import (
	"reflect"
	"strconv"
	"testing"
)

func TestTenantActivationMesh(t *testing.T) {
	const systemserver = "../../../system-server/.olares/config/user/helm-charts/systemserver/templates/"
	const launcher = "../../../bfl/.olares/config/launcher/templates/"
	const vault = "../../../vault/.olares/config/cluster/deploy/vault_rpc_mesh.yaml"

	for _, tenant := range []string{"alice", "bob"} {
		t.Run(tenant, func(t *testing.T) {
			space, system := "user-space-"+tenant, "user-system-"+tenant
			values := map[string]any{"bfl": map[string]any{"username": tenant, "prepareAPIMesh": true}}
			docs, err := renderMeshManifestsInNamespace(t, space, false, values,
				systemserver+"activation-mesh.yaml", systemserver+"nonce-mesh.yaml",
				launcher+"bfl-vault-rpc-mtls.yaml", vault)
			if err != nil {
				t.Fatal(err)
			}
			spec := func(kind, name, ns string) map[string]any {
				t.Helper()
				doc := docs[kind+"/"+name]
				if doc == nil || doc["metadata"].(map[string]any)["namespace"] != ns {
					t.Fatalf("missing %s %s/%s", kind, ns, name)
				}
				return doc["spec"].(map[string]any)
			}
			for _, port := range []int{28080, 9091, 3000} {
				server := spec("Server", "systemserver-existing-"+strconv.Itoa(port), system)
				if server["accessPolicy"] != "deny" || server["port"] != port {
					t.Fatalf("listener %d is not default deny: %v", port, server)
				}
			}
			if got := spec("Server", "vault-public-rpc", "os-framework"); got["port"] != 3000 || got["accessPolicy"] != "deny" {
				t.Fatalf("Vault public RPC listener was broadened: %v", got)
			}
			identity := func(name, ns string) map[string]any {
				return map[string]any{"kind": "ServiceAccount", "name": name, "namespace": ns}
			}
			for _, rule := range []struct {
				name, ns, route, parent string
				identities              []any
			}{
				{"activation-password", system, "activation-password", "systemserver-existing-28080", []any{identity("bytetrade-controller", space)}},
				{"activation-login", system, "activation-login", "systemserver-existing-9091", []any{
					identity("envoy-os-gateway-app-gateway-f5ccc039", "os-gateway"),
					identity("bfl-login", space), identity("system-frontend", space), identity("bfl-wizard", space),
				}},
				{"activation-vault", system, "activation-vault", "systemserver-existing-3000", []any{
					identity("bfl-wizard", space), identity("system-frontend", space),
				}},
				{"vault-rpc-" + tenant, "os-framework", "vault-public-rpc", "vault-public-rpc", []any{identity("bytetrade-sys-ops", system)}},
			} {
				if refs := spec("MeshTLSAuthentication", rule.name, rule.ns)["identityRefs"]; !reflect.DeepEqual(refs, rule.identities) {
					t.Fatalf("%s granted another identity: %v", rule.name, refs)
				}
				target := map[string]any{"group": "gateway.networking.k8s.io", "kind": "HTTPRoute", "name": rule.route}
				authz := spec("AuthorizationPolicy", rule.name, rule.ns)
				if !reflect.DeepEqual(authz["targetRef"], target) || !reflect.DeepEqual(authz["requiredAuthenticationRefs"],
					[]any{map[string]any{"group": "policy.linkerd.io", "kind": "MeshTLSAuthentication", "name": rule.name}}) {
					t.Fatalf("unexpected route authorization: %v", authz)
				}
				if got := spec("HTTPRoute", rule.route, rule.ns)["parentRefs"]; !reflect.DeepEqual(got,
					[]any{map[string]any{"group": "policy.linkerd.io", "kind": "Server", "name": rule.parent}}) {
					t.Fatalf("invalid route parent: %v", got)
				}
			}
			wantPassword := []any{map[string]any{"matches": []any{map[string]any{
				"path": map[string]any{"type": "Exact", "value": "/api/reset/" + tenant + "/password"}, "method": "POST",
			}}}}
			if got := spec("HTTPRoute", "activation-password", system)["rules"]; !reflect.DeepEqual(got, wantPassword) {
				t.Fatalf("password grant was broadened: %v", got)
			}
			guard := spec("HTTPRoute", "activation-password-deny", system)
			if got := guard["rules"]; !reflect.DeepEqual(got, []any{map[string]any{"matches": []any{map[string]any{
				"path": map[string]any{"type": "PathPrefix", "value": "/api/reset"},
			}}}}) || docs["AuthorizationPolicy/activation-password-deny"] != nil {
				t.Fatalf("unrecognized reset requests fall back to generic provider access: %v", got)
			}
			if provider := spec("MeshTLSAuthentication", "activation-provider", system)["identities"]; !reflect.DeepEqual(provider,
				[]any{"*.serviceaccount.identity.os-mesh.cluster.local"}) {
				t.Fatalf("provider compatibility route failed to require Linkerd mTLS: %v", provider)
			}
			if _, err := renderMeshManifestsInNamespace(t, "user-space-other", false, values, systemserver+"activation-mesh.yaml"); err == nil {
				t.Fatal("a launcher can enroll another tenant's activation routes")
			}
		})
	}
}

func TestWizardChartInjectsAuthenticatedInboundMeshWithoutChangingImage(t *testing.T) {
	const template = "../../../../apps/.olares/config/user/helm-charts/wizard/templates/wizard_deploy.yaml"
	docs, err := renderMeshManifestsInNamespace(t, "user-space-alice", false, map[string]any{
		"bfl": map[string]any{"username": "alice"}, "userspace": map[string]any{"userData": "/olares"},
	}, template)
	if err != nil {
		t.Fatal(err)
	}
	pod := docs["Deployment/wizard"]["spec"].(map[string]any)["template"].(map[string]any)
	annotations := pod["metadata"].(map[string]any)["annotations"].(map[string]any)
	spec := pod["spec"].(map[string]any)
	if annotations["linkerd.io/inject"] != "enabled" ||
		annotations["config.linkerd.io/default-inbound-policy"] != "cluster-authenticated" ||
		annotations["config.linkerd.io/proxy-enable-native-sidecar"] != "true" ||
		spec["serviceAccountName"] != "bfl-wizard" {
		t.Fatalf("Wizard does not declare its activation mesh posture: %v", annotations)
	}
	container := spec["containers"].([]any)[0].(map[string]any)
	if container["image"] != "beclab/wizard:v1.13.1" {
		t.Fatalf("unexpected Wizard business image: %v", container["image"])
	}
}
