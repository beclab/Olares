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
	values, err := chartutil.ToRenderValues(c, map[string]any{"rootPath": "/olares", "bfl": map[string]any{"username": tenant}}, chartutil.ReleaseOptions{Namespace: namespace, Name: "nonce-test"}, nil)
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
			docs := renderNonceManifests(t, "user-space-"+tenant, tenant, "../../../system-server/.olares/config/user/helm-charts/systemserver/templates/nonce-mesh.yaml")
			for _, doc := range docs {
				meta := doc["metadata"].(map[string]any)
				if ns, ok := meta["namespace"]; ok && ns != "user-system-"+tenant {
					t.Fatal("cross-tenant resource")
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
			if !reflect.DeepEqual(nonce["rules"], []any{map[string]any{"matches": []any{map[string]any{"path": map[string]any{"type": "PathPrefix", "value": "/permission/v1alpha1/nonce"}}}}}) {
				t.Fatal("nonce variants may fall through to generic route")
			}
			identity := docs["MeshTLSAuthentication/systemserver-nonce"]["spec"].(map[string]any)
			if !reflect.DeepEqual(identity, map[string]any{"identityRefs": []any{map[string]any{"kind": "ServiceAccount", "name": "authelia-backend", "namespace": "os-framework"}}}) {
				t.Fatal("shared or cross-tenant identity can obtain nonce")
			}
			subjects := docs["ClusterRoleBinding/systemserver-nonce-reviewer-"+tenant]["subjects"]
			if !reflect.DeepEqual(subjects, []any{map[string]any{"kind": "ServiceAccount", "name": "bytetrade-sys-ops", "namespace": "user-system-" + tenant}}) {
				t.Fatal("reviewer bound across tenants")
			}
		})
	}
}

func TestAutheliaDedicatedLeastPrivilegeAndRotatingCredential(t *testing.T) {
	docs := renderNonceManifests(t, "os-framework", "alice",
		"../../../authelia/.olares/config/cluster/deploy/auth_backend_identity.yaml",
		"../../../authelia/.olares/config/cluster/deploy/auth_backend_deploy.yaml")
	spec := docs["Deployment/authelia-backend"]["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any)
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
	found := false
	for _, volume := range spec["volumes"].([]any) {
		v := volume.(map[string]any)
		if v["name"] != "nonce-credential" {
			continue
		}
		found = true
		projected := v["projected"].(map[string]any)
		token := projected["sources"].([]any)[0].(map[string]any)["serviceAccountToken"].(map[string]any)
		if projected["defaultMode"] != 256 || !reflect.DeepEqual(token, map[string]any{"audience": "olares-system-server-nonce", "expirationSeconds": 600, "path": "token"}) {
			t.Fatal("credential is not short-lived and audience-scoped")
		}
	}
	if !found {
		t.Fatal("missing projected credential")
	}
	for _, container := range spec["containers"].([]any) {
		c := container.(map[string]any)
		if c["name"] != "authelia" {
			continue
		}
		mounted := false
		for _, mount := range c["volumeMounts"].([]any) {
			m := mount.(map[string]any)
			if m["name"] == "nonce-credential" {
				mounted = m["readOnly"] == true && m["mountPath"] == "/var/run/secrets/olares/nonce" && m["subPath"] == nil
			}
		}
		if !mounted {
			t.Fatal("credential rotation mount is unsafe")
		}
	}
}
