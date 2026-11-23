package access

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestAutheliaStartupWaitsForTCPAfterLinkerd(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(file), "../../../..")
	for _, tc := range []struct {
		path, deployment, targets string
	}{
		{
			path:       "framework/authelia/.olares/config/user/helm-charts/auth/templates/auth_deploy.yaml",
			deployment: "authelia-deployment",
			targets:    "authelia-backend.user-system-{{ .Values.bfl.username }}:9091",
		},
		{
			path:       "apps/.olares/config/user/helm-charts/system-apps/templates/olares-app.yaml",
			deployment: "olares-app-deployment",
			targets: "authelia-backend.user-system-{{ .Values.bfl.username }}:9091," +
				"infisical-service.user-system-{{ .Values.bfl.username }}:8080," +
				"system-server.user-system-{{ .Values.bfl.username }}:80,nats.user-system-{{ .Values.bfl.username }}:4222",
		},
	} {
		t.Run(tc.deployment, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, tc.path))
			if err != nil {
				t.Fatal(err)
			}
			_, deployment, found := strings.Cut(string(data), "kind: Deployment\nmetadata:\n  name: "+tc.deployment+"\n")
			metadata, containers, hasInit := strings.Cut(deployment, "      initContainers:\n")
			init, _, hasApp := strings.Cut(containers, "\n      containers:")
			if !found || !hasInit || !hasApp {
				t.Fatal("could not locate the init containers in the deployment template")
			}
			for _, annotation := range []string{
				`linkerd.io/inject: enabled`,
				`config.linkerd.io/proxy-enable-native-sidecar: "true"`,
				`config.linkerd.io/default-inbound-policy: cluster-authenticated`,
				`traffic.olares.io/mtls-mode: Required`,
			} {
				if !strings.Contains(metadata, "\n        "+annotation+"\n") {
					t.Errorf("missing Pod annotation %s", annotation)
				}
			}
			var probes []struct {
				Name            string   `yaml:"name"`
				Image           string   `yaml:"image"`
				ImagePullPolicy string   `yaml:"imagePullPolicy"`
				Command         []string `yaml:"command"`
				Args            []string `yaml:"args"`
			}
			if err := yaml.Unmarshal([]byte(init), &probes); err != nil {
				t.Fatal(err)
			}
			if len(probes) == 0 {
				t.Fatal("no init containers found")
			}
			probe := probes[0]
			if probe.Name != "check-auth" || probe.Image != "owncloudci/wait-for:latest" ||
				probe.ImagePullPolicy != "IfNotPresent" || len(probe.Command) != 0 ||
				!reflect.DeepEqual(probe.Args, []string{"-it", tc.targets}) {
				t.Fatalf("unexpected check-auth startup contract: %+v", probe)
			}
		})
	}
}

func TestEntryMTLSIsScopedToTheDedicatedListener(t *testing.T) {
	channel := readMeshManifest(t, "l4-mtls-channel.yaml")
	proxy := readMeshManifest(t, "envoyproxy.yaml")["EnvoyProxy"]
	service := channel["Service"]
	ports := meshField(t, service, "spec", "ports").([]any)
	if len(ports) != 2 {
		t.Fatalf("unexpected Entry Service ports %v", ports)
	}
	for i, tc := range []struct {
		port                    int
		server, policy, section string
	}{
		{8443, "app-gateway-entry-mtls", "app-gateway-l4-entry-mtls", "l4-mtls"},
		{9444, "app-gateway-entry-authz", "app-gateway-l4-authz-mtls", "l4-authz"},
	} {
		server := channel["Server/"+tc.server]
		policy := channel["ClientTrafficPolicy/"+tc.policy]
		if meshField(t, ports[i], "port") != tc.port || meshField(t, ports[i], "targetPort") != tc.port ||
			meshField(t, server, "spec", "port") != tc.port || meshField(t, server, "spec", "proxyProtocol") != "TLS" ||
			meshField(t, server, "spec", "accessPolicy") != "all-unauthenticated" ||
			!reflect.DeepEqual(meshField(t, server, "spec", "podSelector", "matchLabels"), meshField(t, service, "spec", "selector")) {
			t.Fatalf("Entry Service and Linkerd exception disagree for %s", tc.section)
		}
		if meshField(t, policy, "spec", "targetRef", "sectionName") != tc.section ||
			meshField(t, policy, "spec", "tls", "clientValidation", "mode") != "RequireAndVerify" {
			t.Fatalf("Envoy must require L4 mTLS on %s", tc.section)
		}
		names := meshField(t, policy, "spec", "tls", "clientValidation", "subjectAltNames", "dnsNames").([]any)
		if len(names) != 1 || meshField(t, names[0], "type") != "Exact" || meshField(t, names[0], "value") != "l4-bfl-proxy.entry.client" {
			t.Fatalf("unexpected Entry client identities %v", names)
		}
	}
	gw := readMeshManifest(t, "gateway.yaml")["Gateway"]
	frontend := meshField(t, gw, "spec", "tls", "frontend", "perPort").([]any)
	if len(frontend) != 1 || meshField(t, frontend[0], "port") != 9444 ||
		meshField(t, frontend[0], "tls", "validation", "mode") != "AllowValidOnly" {
		t.Fatal("the authorization listener must require native frontend TLS validation")
	}
	if meshField(t, proxy, "spec", "provider", "kubernetes", "envoyService", "type") != "ClusterIP" {
		t.Fatal("the gateway must not provision a public LoadBalancer alongside L4")
	}
	annotations := meshField(t, proxy, "spec", "provider", "kubernetes", "envoyDeployment", "pod", "annotations").(map[string]any)
	if annotations["config.linkerd.io/default-inbound-policy"] != "cluster-authenticated" ||
		annotations["linkerd.io/inject"] != "enabled" ||
		annotations["config.linkerd.io/skip-inbound-ports"] != nil ||
		annotations["config.linkerd.io/opaque-ports"] != "19001" {
		t.Fatalf("unexpected AGW mesh posture %v", annotations)
	}
}

func TestInternalAccessGatewayUsesActualEnvoyListenerPort(t *testing.T) {
	resources := readMeshManifest(t, "access-internal-gateway.yaml")
	listeners := meshField(t, resources["Gateway"], "spec", "listeners").([]any)
	ports := meshField(t, resources["Service"], "spec", "ports").([]any)
	if len(listeners) != 1 || len(ports) != 1 ||
		meshField(t, listeners[0], "port") != 8080 ||
		meshField(t, ports[0], "targetPort") != 8080 ||
		meshField(t, resources["Server"], "spec", "port") != 8080 {
		t.Fatal("internal access Gateway, Service, and Linkerd Server must agree on Envoy's unprivileged listener")
	}
}

func TestHTTPMeshAdmissionIsBoundAndFailClosed(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	data, err := os.Open(filepath.Join(filepath.Dir(file), "../../.olares/config/cluster/deploy/http-mesh-admission.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	defer data.Close()

	decoder := yaml.NewDecoder(data)
	policies, bindings := map[string]bool{}, map[string]bool{}
	for {
		var doc map[string]any
		if err := decoder.Decode(&doc); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		name := meshField(t, doc, "metadata", "name").(string)
		switch doc["kind"] {
		case "ValidatingAdmissionPolicy":
			if meshField(t, doc, "spec", "failurePolicy") != "Fail" {
				t.Fatalf("%s must fail closed", name)
			}
			policies[name] = true
		case "ValidatingAdmissionPolicyBinding":
			if meshField(t, doc, "spec", "policyName") != name ||
				!reflect.DeepEqual(meshField(t, doc, "spec", "validationActions"), []any{"Deny", "Audit"}) {
				t.Fatalf("unexpected enforcement for %s", name)
			}
			bindings[name] = true
		default:
			t.Fatalf("unexpected resource %v", doc["kind"])
		}
	}
	want := map[string]bool{"olares-http-mesh-pods": true, "olares-entry-mtls-client-validation": true, "olares-l4-authz-route-writer": true}
	if !reflect.DeepEqual(policies, want) || !reflect.DeepEqual(bindings, want) {
		t.Fatalf("missing admission policies or bindings: %v %v", policies, bindings)
	}
}

func readMeshManifest(t *testing.T, name string) map[string]map[string]any {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	data, err := os.Open(filepath.Join(filepath.Dir(file), "../../.olares/config/cluster/deploy", name))
	if err != nil {
		t.Fatal(err)
	}
	defer data.Close()
	decoder := yaml.NewDecoder(data)
	resources := make(map[string]map[string]any)
	for {
		var doc map[string]any
		if err := decoder.Decode(&doc); err == io.EOF {
			return resources
		} else if err != nil {
			t.Fatal(err)
		}
		kind := meshField(t, doc, "kind").(string)
		resources[kind] = doc
		resources[kind+"/"+meshField(t, doc, "metadata", "name").(string)] = doc
	}
}

func meshField(t *testing.T, object any, path ...string) any {
	t.Helper()
	for _, name := range path {
		fields, ok := object.(map[string]any)
		if !ok || fields[name] == nil {
			t.Fatalf("missing field %s in path %v", name, path)
		}
		object = fields[name]
	}
	return object
}

func TestL4ReadsOnlySystemEntryReadinessNames(t *testing.T) {
	resources := readMeshManifest(t, "l4-mtls-channel.yaml")
	role := resources["ClusterRole/olares-l4-system-entry-readiness"]
	rules := meshField(t, role, "rules").([]any)
	if len(rules) != 1 ||
		!reflect.DeepEqual(meshField(t, rules[0], "apiGroups"), []any{"gateway.olares.io"}) ||
		!reflect.DeepEqual(meshField(t, rules[0], "resources"), []any{"sharedrouteregistries"}) ||
		!reflect.DeepEqual(meshField(t, rules[0], "resourceNames"), []any{"sys-auth", "sys-desktop"}) ||
		!reflect.DeepEqual(meshField(t, rules[0], "verbs"), []any{"get"}) {
		t.Fatalf("L4 received broader access than required: %v", rules)
	}
	binding := resources["ClusterRoleBinding/olares-l4-system-entry-readiness"]
	subjects := meshField(t, binding, "subjects").([]any)
	if len(subjects) != 1 || meshField(t, subjects[0], "kind") != "ServiceAccount" ||
		meshField(t, subjects[0], "namespace") != "os-network" || meshField(t, subjects[0], "name") != "os-network-internal" ||
		meshField(t, binding, "roleRef", "name") != "olares-l4-system-entry-readiness" {
		t.Fatalf("unexpected role binding: %v", binding)
	}
}
