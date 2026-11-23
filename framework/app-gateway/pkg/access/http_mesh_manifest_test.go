package access

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestEntryMTLSIsScopedToTheDedicatedListener(t *testing.T) {
	channel := readMeshManifest(t, "l4-mtls-channel.yaml")
	proxy := readMeshManifest(t, "envoyproxy.yaml")["EnvoyProxy"]
	server := channel["Server"]
	service := channel["Service"]
	policy := channel["ClientTrafficPolicy"]

	ports := meshField(t, service, "spec", "ports").([]any)
	if len(ports) != 1 || meshField(t, ports[0], "port") != 8443 || meshField(t, ports[0], "targetPort") != 8443 {
		t.Fatalf("unexpected Entry Service ports %v", ports)
	}
	if meshField(t, server, "spec", "port") != 8443 || meshField(t, server, "spec", "proxyProtocol") != "TLS" ||
		meshField(t, server, "spec", "accessPolicy") != "all-unauthenticated" ||
		!reflect.DeepEqual(meshField(t, server, "spec", "podSelector", "matchLabels"), meshField(t, service, "spec", "selector")) {
		t.Fatal("Linkerd's Entry exception must select exactly the Entry Service pods and the TLS listener")
	}
	if meshField(t, policy, "spec", "targetRef", "sectionName") != "l4-mtls" ||
		meshField(t, policy, "spec", "tls", "clientValidation", "mode") != "RequireAndVerify" {
		t.Fatal("Envoy must authenticate L4 clients on the dedicated Entry listener")
	}
	names := meshField(t, policy, "spec", "tls", "clientValidation", "subjectAltNames", "dnsNames").([]any)
	if len(names) != 1 || meshField(t, names[0], "type") != "Exact" || meshField(t, names[0], "value") != "l4-bfl-proxy.entry.client" {
		t.Fatalf("unexpected Entry client identities %v", names)
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
	want := map[string]bool{"olares-http-mesh-pods": true, "olares-entry-mtls-client-validation": true}
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
		resources[meshField(t, doc, "kind").(string)] = doc
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
