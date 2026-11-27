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

func renderNonceManifests(t *testing.T, namespace, tenant string, paths ...string) map[string]map[string]any {
	t.Helper()
	c := &chart.Chart{Metadata: &chart.Metadata{Name: "nonce-contract"}}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		c.Templates = append(c.Templates, &chart.File{Name: "templates/" + filepath.Base(path), Data: data})
	}
	values, err := chartutil.ToRenderValues(c, map[string]any{"rootPath": "/olares", "autheliaBackendImage": "beclab/auth:0.2.63-1.12.8-50b54e7c", "bfl": map[string]any{"username": tenant}}, chartutil.ReleaseOptions{Namespace: namespace, Name: "nonce-test"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	files, err := engine.Render(c, values)
	if err != nil {
		t.Fatal(err)
	}
	docs := map[string]map[string]any{}
	for _, text := range files {
		d := yaml.NewDecoder(strings.NewReader(text))
		for {
			var doc map[string]any
			err := d.Decode(&doc)
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(doc) == 0 {
				continue
			}
			meta := doc["metadata"].(map[string]any)
			key := doc["kind"].(string) + "/" + meta["name"].(string)
			if doc["kind"] == "Role" || doc["kind"] == "RoleBinding" {
				key = doc["kind"].(string) + "/" + meta["namespace"].(string) + "/" + meta["name"].(string)
			}
			if docs[key] != nil {
				t.Fatal("duplicate resource", key)
			}
			docs[key] = doc
		}
	}
	return docs
}

func TestNonceMeshIdentityAndTenantBinding(t *testing.T) {
	for _, tenant := range []string{"alice", "bob"} {
		t.Run(tenant, func(t *testing.T) {
			docs := renderNonceManifests(t, "user-space-"+tenant, tenant, "../../../system-server/.olares/config/user/helm-charts/systemserver/templates/nonce-mesh.yaml",
				"../../../system-server/.olares/config/user/helm-charts/systemserver/templates/systemserver_deploy.yaml")
			for _, doc := range docs {
				meta := doc["metadata"].(map[string]any)
				if ns, ok := meta["namespace"]; ok && ns != "user-system-"+tenant {
					t.Fatal("cross-tenant resource")
				}
			}
			pod := docs["Deployment/system-server"]["spec"].(map[string]any)["template"].(map[string]any)
			ann := pod["metadata"].(map[string]any)["annotations"].(map[string]any)
			if ann["config.linkerd.io/default-inbound-policy"] != "deny" || ann["config.linkerd.io/proxy-enable-native-sidecar"] != "true" {
				t.Fatal("system-server deployment can fail open")
			}
			if ann["linkerd.io/inject"] != "enabled" || ann["traffic.olares.io/auth-proxy-context-version"] != "1" {
				t.Fatal("fresh tenant would not enroll its tenant-bound Auth proxy")
			}
			declared := map[any]bool{}
			for _, container := range pod["spec"].(map[string]any)["containers"].([]any) {
				c := container.(map[string]any)
				for _, port := range c["ports"].([]any) {
					declared[port.(map[string]any)["containerPort"]] = true
				}
				if c["name"] == "system-server" {
					if c["image"] == "beclab/system-server:0.1.37" || !strings.Contains(c["image"].(string), "@sha256:") {
						t.Fatal("fresh tenant would use the legacy nonce handler or a mutable image tag")
					}
					enrolled := false
					for _, env := range c["env"].([]any) {
						e := env.(map[string]any)
						if e["name"] == "NONCE_AUTHORIZATION_MODE" && e["value"] == "linkerd" {
							enrolled = true
						}
					}
					if !enrolled {
						t.Fatal("mesh-only handler not explicitly enrolled")
					}
				}
			}
			for key, doc := range docs {
				if strings.HasPrefix(key, "Server/") && !declared[doc["spec"].(map[string]any)["port"]] {
					t.Fatalf("Linkerd Server port not declared on Pod: %s", key)
				}
			}
			server := docs["Server/systemserver-api"]["spec"].(map[string]any)
			if server["accessPolicy"] != "deny" || server["port"] != 80 || server["proxyProtocol"] != "HTTP/1" {
				t.Fatal("nonce port does not fail closed")
			}
			if !reflect.DeepEqual(server["podSelector"], map[string]any{"matchLabels": map[string]any{"app": "systemserver"}}) {
				t.Fatal("wrong protected workload")
			}
			for _, name := range []string{"systemserver-nonce", "systemserver-api-existing"} {
				route := docs["HTTPRoute/"+name]["spec"].(map[string]any)
				if !reflect.DeepEqual(route["parentRefs"], []any{map[string]any{"group": "policy.linkerd.io", "kind": "Server", "name": "systemserver-api"}}) {
					t.Fatal("route detached from protected port")
				}
				authz := docs["AuthorizationPolicy/"+name]["spec"].(map[string]any)
				if !reflect.DeepEqual(authz["targetRef"], map[string]any{"group": "gateway.networking.k8s.io", "kind": "HTTPRoute", "name": name}) {
					t.Fatal("server-wide authorization can bypass nonce restriction")
				}
				if !reflect.DeepEqual(authz["requiredAuthenticationRefs"], []any{map[string]any{"group": "policy.linkerd.io", "kind": "MeshTLSAuthentication", "name": name}}) {
					t.Fatal("identity requirement missing")
				}
			}
			nonce := docs["HTTPRoute/systemserver-nonce"]["spec"].(map[string]any)
			if !reflect.DeepEqual(nonce["rules"], []any{map[string]any{"matches": []any{map[string]any{"path": map[string]any{"type": "Exact", "value": "/permission/v1alpha1/nonce"}, "method": "GET"}}}}) {
				t.Fatal("nonce variants may fall through to generic route")
			}
			identity := docs["MeshTLSAuthentication/systemserver-nonce"]["spec"].(map[string]any)
			if !reflect.DeepEqual(identity, map[string]any{"identityRefs": []any{map[string]any{"kind": "ServiceAccount", "name": "authelia-backend", "namespace": "os-framework"}}}) {
				t.Fatal("shared or cross-tenant identity can obtain nonce")
			}
			if docs["ClusterRoleBinding/systemserver-nonce-reviewer-"+tenant] != nil {
				t.Fatal("obsolete TokenReview permission retained")
			}
			existing := docs["HTTPRoute/systemserver-api-existing"]["spec"].(map[string]any)
			if !reflect.DeepEqual(existing["rules"], []any{map[string]any{"matches": []any{
				map[string]any{"path": map[string]any{"type": "PathPrefix", "value": "/permission/v2alpha1"}},
				map[string]any{"path": map[string]any{"type": "PathPrefix", "value": "/provider/v2alpha1"}},
			}}}) {
				t.Fatal("a fallback route can authorize nonce when its policy is missing")
			}
			for _, port := range []string{"28080", "6379", "27017", "5432", "4222", "9091", "3010", "4568", "8080", "3000"} {
				server := docs["Server/systemserver-existing-"+port]["spec"].(map[string]any)
				if server["accessPolicy"] != "cluster-authenticated" {
					t.Fatal("other listener compatibility lost")
				}
			}

		})
	}
}

func TestTenantNATSMeshKeepsServerFirstTrafficIntercepted(t *testing.T) {
	for _, tenant := range []string{"alice", "bob"} {
		t.Run(tenant, func(t *testing.T) {
			namespace := "user-system-" + tenant
			docs := renderNonceManifests(t, namespace, tenant,
				"../../../system-server/.olares/config/user/helm-charts/systemserver/templates/nonce-mesh.yaml",
				"../../../system-server/.olares/config/user/helm-charts/systemserver/templates/systemserver_deploy.yaml",
				"../../../../platform/tapr/.olares/config/user/helm-charts/nats/templates/deployment.yaml")
			service := docs["Service/nats"]
			ports := service["spec"].(map[string]any)["ports"].([]any)
			want := []any{map[string]any{"name": "nats", "protocol": "TCP", "appProtocol": "linkerd.io/opaque", "port": 4222, "targetPort": 4222}}
			if service["metadata"].(map[string]any)["namespace"] != namespace || !reflect.DeepEqual(ports, want) {
				t.Fatalf("unexpected tenant NATS Service: %v", service)
			}
			server := docs["Server/systemserver-existing-4222"]["spec"].(map[string]any)
			if server["proxyProtocol"] != "opaque" || server["port"] != 4222 || server["accessPolicy"] != "cluster-authenticated" {
				t.Fatalf("NATS requires authenticated opaque TCP: %v", server)
			}
			nonce := docs["Server/systemserver-api"]["spec"].(map[string]any)
			if nonce["proxyProtocol"] != "HTTP/1" || nonce["accessPolicy"] != "deny" {
				t.Fatalf("NATS configuration changed strict nonce boundary: %v", nonce)
			}
		})
	}
}

func TestAutheliaDedicatedLeastPrivilegeWithoutNonceToken(t *testing.T) {
	docs := renderNonceManifests(t, "os-framework", "alice",
		"../../../authelia/.olares/config/cluster/deploy/auth_backend_identity.yaml",
		"../../../authelia/.olares/config/cluster/deploy/auth_backend_deploy.yaml")
	spec := docs["Deployment/authelia-backend"]["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any)
	pod := docs["Deployment/authelia-backend"]["spec"].(map[string]any)["template"].(map[string]any)
	if pod["metadata"].(map[string]any)["annotations"].(map[string]any)["config.linkerd.io/proxy-enable-native-sidecar"] != "true" {
		t.Fatal("database init can start before its mesh proxy")
	}
	for _, container := range spec["containers"].([]any) {
		c := container.(map[string]any)
		if c["name"] == "authelia" && c["image"] != "beclab/auth:0.2.63-1.12.8-50b54e7c" {
			t.Fatal("identity migration changed the retained auth image")
		}
	}

	if spec["serviceAccountName"] != "authelia-backend" || spec["serviceAccount"] != "authelia-backend" {
		t.Fatal("Authelia retains shared administrator identity")
	}
	role := docs["ClusterRole/authelia-backend"]
	want := []any{
		map[string]any{"apiGroups": []any{"app.bytetrade.io"}, "resources": []any{"applications"}, "verbs": []any{"list"}},
		map[string]any{"apiGroups": []any{"iam.kubesphere.io"}, "resources": []any{"users"}, "verbs": []any{"list"}},
		map[string]any{"apiGroups": []any{""}, "resources": []any{"nodes"}, "verbs": []any{"list"}},
		map[string]any{"apiGroups": []any{"authentication.k8s.io"}, "resources": []any{"tokenreviews"}, "verbs": []any{"create"}},
	}
	if !reflect.DeepEqual(role["rules"], want) {
		t.Fatal("Authelia RBAC expanded beyond source call inventory")
	}
	if docs["ClusterRoleBinding/authelia-backend"]["roleRef"].(map[string]any)["name"] != "authelia-backend" {
		t.Fatal("shared or administrator binding")
	}
	for namespace, name := range map[string]string{"os-platform": "lldap-credentials", "kubesphere-system": "redis-secret"} {
		rules := docs["Role/"+namespace+"/authelia-backend"]["rules"]
		if !reflect.DeepEqual(rules, []any{map[string]any{"apiGroups": []any{""}, "resources": []any{"secrets"}, "resourceNames": []any{name}, "verbs": []any{"get"}}}) {
			t.Fatal("Secret access is not restricted to the source inventory")
		}
		binding := docs["RoleBinding/"+namespace+"/authelia-backend"]
		if !reflect.DeepEqual(binding["subjects"], []any{map[string]any{"kind": "ServiceAccount", "name": "authelia-backend", "namespace": "os-framework"}}) {
			t.Fatal("Secret role bound to wrong identity")
		}
	}
	for _, volume := range spec["volumes"].([]any) {
		v := volume.(map[string]any)
		if v["name"] == "nonce-credential" {
			t.Fatal("Authelia still requires a new nonce token")
		}
	}
}
