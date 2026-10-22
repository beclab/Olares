package app

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	netv1 "k8s.io/api/networking/v1"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"
)

func renderFrontend(t *testing.T) map[string]string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../.."))
	c := &chart.Chart{Metadata: &chart.Metadata{Name: "frontend", Version: "0.1.0", APIVersion: "v2"}}
	sources := map[string]string{
		"olares-app.yaml":                  "apps/.olares/config/cluster/deploy/os-frontend/olares-app.yaml",
		"olares-app-nginx.yaml":            "apps/.olares/config/cluster/deploy/os-frontend/olares-app-nginx.yaml",
		"user-service-backend-policy.yaml": "apps/.olares/config/cluster/deploy/os-frontend/user-service-backend-policy.yaml",
		"headscale.yaml":                   "framework/headscale/.olares/config/cluster/deploy/headscale.yaml",
	}
	for name, path := range sources {
		data, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		c.Templates = append(c.Templates, &chart.File{Name: "templates/" + name, Data: data})
	}
	vals, err := chartutil.ToRenderValues(c, map[string]interface{}{}, chartutil.ReleaseOptions{Name: "test", Namespace: "os-framework", IsInstall: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := engine.Render(c, vals)
	if err != nil {
		t.Fatal(err)
	}
	return rendered
}

func frontendConfigs(t *testing.T, rendered map[string]string) map[string]string {
	t.Helper()
	for _, doc := range strings.Split(rendered["frontend/templates/olares-app-nginx.yaml"], "\n---\n") {
		var cm struct {
			Data map[string]string `json:"data"`
		}
		if err := yaml.Unmarshal([]byte(doc), &cm); err != nil {
			t.Fatal(err)
		}
		if cm.Data["dashboard-control-hub.conf"] != "" {
			return cm.Data
		}
	}
	t.Fatal("routes ConfigMap not found")
	return nil
}

func TestFrontendProxyRoutesAndCredentials(t *testing.T) {
	configs := frontendConfigs(t, renderFrontend(t))
	for name, needles := range map[string][]string{
		"dashboard-control-hub.conf": {"/proxy/bfl/bfl;", "/proxy/middleware/middleware;"},
		"vault.conf":                 {"/proxy/vaultRpc/server;", "/proxy/bfl/bfl/;"},
		"settings.conf":              {"/proxy/secret/admin;"},
	} {
		for _, needle := range needles {
			if !strings.Contains(configs[name], needle) {
				t.Errorf("%s missing %s", name, needle)
			}
		}
	}
	if strings.Contains(configs["settings.conf"], "location /vault") || strings.Contains(configs["legacy-providers.conf"], "$legacy_vault") {
		t.Fatal("unused Settings Vault proxy must not be restored")
	}
	dashboard := configs["dashboard-control-hub.conf"]
	if strings.Contains(dashboard, "proxy_pass http://monitoring.user-space-") {
		t.Fatal("terminal/exec/watch must not use the per-user Provider")
	}
	if n := strings.Count(dashboard, "proxy_pass http://monitoring-server.os-framework:80;"); n != 9 {
		t.Fatalf("direct monitoring HTTP/stream routes=%d, want 9", n)
	}
	for _, identity := range []string{"dashboard", "control_hub"} {
		for _, marker := range []string{
			"auth_request /_" + identity + "_identity;",
			"proxy_set_header X-Bfl-User $" + identity + "_user;",
			"proxy_set_header Cookie \"auth_token=$" + identity + "_token\";",
		} {
			if !strings.Contains(dashboard, marker) {
				t.Errorf("missing verified stream identity: %s", marker)
			}
		}
	}
	for _, marker := range []string{"proxy_buffering off;", "proxy_read_timeout 3600s;", "proxy_send_timeout 3600s;", "proxy_set_header Sec-WebSocket-Protocol $http_sec_websocket_protocol;"} {
		if n := strings.Count(dashboard, marker); n != 9 {
			t.Errorf("stream setting %s occurs %d times, want 9", marker, n)
		}
	}
	for _, needle := range []string{"proxy_pass http://bfl.user-system-", "proxy_pass http://middleware.user-space-"} {
		if strings.Contains(dashboard, needle) {
			t.Errorf("ordinary HTTP retains %s", needle)
		}
	}
	for _, needle := range []string{"auth_request /_vault_identity;", "X-Provider-Proxy", ":28080"} {
		if strings.Contains(configs["vault.conf"], needle) {
			t.Errorf("Vault retains legacy provider: %s", needle)
		}
	}
	for _, name := range []string{"dashboard-control-hub.conf", "desktop.conf", "settings.conf"} {
		if strings.Contains(configs[name], "/proxy/monitoring/") {
			t.Errorf("%s still proxies monitoring data through user-service", name)
		}
	}
	for _, surface := range []string{"desktop", "settings"} {
		config := configs[surface+".conf"]
		for _, marker := range []string{"proxy_pass http://monitoring-server.os-framework:80;", "auth_request /_" + surface + "_identity;", "proxy_set_header X-Bfl-User $" + surface + "_user;"} {
			if !strings.Contains(config, marker) {
				t.Errorf("%s missing %s", surface, marker)
			}
		}
	}
	for _, marker := range []string{"$dashboard_token", "$control_hub_token", "proxy_http_version 1.1;"} {
		if !strings.Contains(dashboard, marker) {
			t.Errorf("WebSocket contract lost: %s", marker)
		}
	}
}

func renderedNetworkPolicy(t *testing.T, source, name string) netv1.NetworkPolicy {
	t.Helper()
	decoder := k8syaml.NewYAMLOrJSONDecoder(strings.NewReader(source), 4096)
	for {
		var np netv1.NetworkPolicy
		err := decoder.Decode(&np)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if np.Kind == "NetworkPolicy" && np.Name == name {
			return np
		}
	}
	t.Fatalf("NetworkPolicy %s missing", name)
	return netv1.NetworkPolicy{}
}

func TestFrontendBackendIngressIsScoped(t *testing.T) {
	rendered := renderFrontend(t)
	for _, tc := range []struct {
		template, name, namespace, app string
		ports                          []int32
	}{
		{"user-service-backend-policy.yaml", "user-service-secret-ingress", "os-protected", "infisical", []int32{8080}},
		{"headscale.yaml", "headscale-ingress", "os-network", "headscale", []int32{8080, 8000}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			np := renderedNetworkPolicy(t, rendered["frontend/templates/"+tc.template], tc.name)
			if np.Namespace != tc.namespace || np.Spec.PodSelector.MatchLabels["app"] != tc.app {
				t.Fatal("incorrect policy target")
			}
			if np.Labels["app.kubernetes.io/managed-by"] != "app-service" || np.Labels["app.kubernetes.io/component"] != "route-control" {
				t.Fatal("policy would be pruned")
			}
			found := 0
			for _, rule := range np.Spec.Ingress {
				for _, peer := range rule.From {
					if peer.NamespaceSelector == nil || peer.NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"] != "os-frontend" {
						continue
					}
					found++
					if len(rule.From) != 1 || peer.PodSelector == nil || peer.PodSelector.MatchLabels["app"] != "olares-app" || len(rule.Ports) != len(tc.ports) {
						t.Fatal("frontend ingress is not narrowly scoped")
					}
					for i, port := range rule.Ports {
						if port.Port.IntVal != tc.ports[i] {
							t.Fatalf("unexpected port: %v", port.Port)
						}
					}
				}
			}
			if found != 1 {
				t.Fatalf("frontend ingress rules=%d, want 1", found)
			}
		})
	}
}

func TestFrontendFilesDirectRoutesAndStreaming(t *testing.T) {
	configs := frontendConfigs(t, renderFrontend(t))
	for _, tc := range []struct {
		name     string
		routes   int
		identity string
	}{
		{"files.conf", 14, "files"}, {"share.conf", 3, ""}, {"settings.conf", 2, "settings"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := configs[tc.name]
			if strings.Contains(config, "files-provider-svc") || strings.Contains(config, "http://files:28080") {
				t.Fatal("Files still uses a Provider hop")
			}
			if n := strings.Count(config, "proxy_pass http://files-service.os-framework:80"); n != tc.routes {
				t.Fatalf("direct Files routes=%d, want %d", n, tc.routes)
			}
			if tc.identity != "" {
				for _, marker := range []string{"auth_request /_" + tc.identity + "_identity;", "proxy_set_header X-Bfl-User $" + tc.identity + "_user;", "proxy_set_header X-Authorization $" + tc.identity + "_token;"} {
					if !strings.Contains(config, marker) {
						t.Errorf("missing verified Files identity: %s", marker)
					}
				}
			} else {
				if strings.Contains(config, "auth_request ") {
					t.Fatal("public share transfer must not require a user-service login")
				}
				if !strings.Contains(config, "proxy_set_header X-Forwarded-Host $http_host;") {
					t.Fatal("share host must survive forwarding")
				}
			}
			for _, marker := range []string{"proxy_request_buffering off;", "proxy_buffering off;", "add_header Accept-Ranges bytes;", "client_max_body_size 4000M;"} {
				if !strings.Contains(config, marker) {
					t.Errorf("lost transfer setting: %s", marker)
				}
			}
		})
	}
	if !strings.Contains(configs["files.conf"], "http://files-service.os-framework:80/api/raw/drive/Home/Pictures/$1;") {
		t.Fatal("legacy picture URL rewrite must be retained")
	}
	if !strings.Contains(configs["share.conf"], "http://files-service.os-framework:80/videos;") {
		t.Fatal("share video URI must be retained")
	}
}

func TestFilesInnerNginxUploadLimit(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../.."))
	data, err := os.ReadFile(filepath.Join(root, "framework/files/.olares/config/cluster/deploy/files_deploy.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "client_max_body_size 100M;") {
		t.Fatal("inner Files Nginx still clips large transfers at 100M")
	}
	if !strings.Contains(string(data), "client_max_body_size 4000M;") {
		t.Fatal("Files upload limit must match the frontend transfer limit")
	}
}
