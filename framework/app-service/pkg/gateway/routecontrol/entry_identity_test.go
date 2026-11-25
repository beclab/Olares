package routecontrol

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strconv"
	"testing"
	"text/template"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"
)

func desktopEntrance() *srrv1alpha1.SharedRouteRegistry {
	return &srrv1alpha1.SharedRouteRegistry{ObjectMeta: metav1.ObjectMeta{Namespace: "user-space-alice", Name: "sys-desktop", UID: "desktop", Labels: map[string]string{ManagedByLabel: ManagedByValue, InstanceLabel: gateway.SystemPlaneInstance, gateway.LabelSystemPlane: gateway.LabelSystemPlaneValue}, Annotations: map[string]string{gateway.AnnotationExtAuthUser: "alice"}}, Spec: srrv1alpha1.SharedRouteRegistrySpec{EntranceClass: srrv1alpha1.EntranceClassApplication, HostPatterns: []string{"desktop.alice.example.org"}, Upstream: srrv1alpha1.UpstreamRef{ServiceName: "desktop", Port: 80}}}
}

// Render the shipped template, not a second copy of its Envoy configuration.
func tenantAuthBootstrap(t *testing.T, owner string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile("../../../../system-server/.olares/config/user/helm-charts/systemserver/templates/systemserver_deploy.yaml")
	if err != nil {
		t.Fatal(err)
	}
	tmpl, err := template.New("proxy").Funcs(template.FuncMap{"quote": strconv.Quote}).Parse(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	var rendered bytes.Buffer
	if err = tmpl.Execute(&rendered, map[string]any{"Values": map[string]any{"bfl": map[string]any{"username": owner}}}); err != nil {
		t.Fatal(err)
	}
	decoder := k8syaml.NewYAMLOrJSONDecoder(&rendered, 4096)
	for {
		var obj corev1.ConfigMap
		err = decoder.Decode(&obj)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if obj.Name == "systemserver-proxy-configs" {
			var result map[string]any
			if err = yaml.Unmarshal([]byte(obj.Data["envoy.yaml"]), &result); err != nil {
				t.Fatal(err)
			}
			return result
		}
	}
	t.Fatal("missing rendered auth proxy configuration")
	return nil
}

func TestTenantContextIsLimitedToAuthorizationHop(t *testing.T) {
	for _, owner := range []string{"alice", "bob"} {
		bootstrap := tenantAuthBootstrap(t, owner)
		listeners, _, _ := unstructured.NestedSlice(bootstrap, "static_resources", "listeners")
		found := false
		for _, raw := range listeners {
			l := raw.(map[string]any)
			chains := l["filter_chains"].([]any)
			filter := chains[0].(map[string]any)["filters"].([]any)[0].(map[string]any)
			if l["name"] != "listener_auth_backend_proxy" {
				if filter["name"] != "envoy.filters.network.tcp_proxy" {
					t.Fatal("unrelated native transport changed")
				}
				continue
			}
			found = true
			if filter["name"] != "envoy.filters.network.http_connection_manager" {
				t.Fatal("auth transport cannot bind tenant context")
			}
			h := filter["typed_config"].(map[string]any)
			if h["use_remote_address"] != false || h["skip_xff_append"] != true {
				t.Fatal("proxy would replace verified browser IP with mesh IP")
			}
			v := h["route_config"].(map[string]any)["virtual_hosts"].([]any)[0].(map[string]any)
			r := v["routes"].([]any)[0].(map[string]any)
			additions := r["request_headers_to_add"].([]any)
			header := additions[0].(map[string]any)
			if len(additions) != 1 || header["append_action"] != "OVERWRITE_IF_EXISTS_OR_ADD" || header["header"].(map[string]any)["key"] != "X-BFL-USER" || header["header"].(map[string]any)["value"] != owner {
				t.Fatalf("tenant is not namespace bound: %v", header)
			}
		}
		if !found {
			t.Fatal("missing authorization listener")
		}
	}
}

func TestEntryIdentityResponseAndIngressContract(t *testing.T) {
	legacy := desktopEntrance()
	legacy.Annotations[gateway.AnnotationStampBFLUser] = "bob"
	if stampedBFLUser(legacy) != "" {
		t.Fatal("legacy route annotation would overwrite the session user after authorization")
	}
	p := desiredApplicationExtAuthSecurityPolicy(desktopEntrance())
	if open, _, _ := unstructured.NestedBool(p.Object, "spec", "extAuth", "failOpen"); open {
		t.Fatal("authorization fails open")
	}
	headers, _, _ := unstructured.NestedStringSlice(p.Object, "spec", "extAuth", "http", "headersToBackend")
	for _, name := range []string{"x-bfl-user", "remote-user", "remote-groups", "remote-name", "remote-email"} {
		found := false
		for _, h := range headers {
			found = found || h == name
		}
		if !found {
			t.Fatal("verified response identity omitted", name)
		}
	}
	raw, err := os.ReadFile("../../../../app-gateway/.olares/config/cluster/deploy/l4-mtls-channel.yaml")
	if err != nil {
		t.Fatal(err)
	}
	dec := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(raw), 4096)
	policies := 0
	sections := map[string]bool{}
	for {
		var obj map[string]any
		err = dec.Decode(&obj)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if obj["kind"] != "ClientTrafficPolicy" {
			continue
		}
		policies++
		section, _, _ := unstructured.NestedString(obj, "spec", "targetRef", "sectionName")
		target, _, _ := unstructured.NestedString(obj, "spec", "targetRef", "name")
		if target != "app-gateway" || (section != "l4-mtls" && section != "l4-authz") || sections[section] {
			t.Fatal("northbound strip must target each Entry listener exactly once")
		}
		sections[section] = true
		h, _, _ := unstructured.NestedMap(obj, "spec", "headers", "earlyRequestHeaders")
		got, _ := json.Marshal(h)
		want, _ := json.Marshal(entryIdentityHeaderSettings())
		if !bytes.Equal(got, want) {
			t.Fatal("shipped listener does not strip forged identity before authorization")
		}
	}
	if policies != 2 {
		t.Fatal("both Entry listeners must enforce header sanitization")
	}
}

func TestEntryIdentityRemovalSetDoesNotFlapOrReintroduceClaims(t *testing.T) {
	h := entryIdentityHeaderSettings()
	names := h["remove"].([]any)
	for i, j := 0, len(names)-1; i < j; i, j = i+1, j-1 {
		names[i], names[j] = names[j], names[i]
	}
	if !validEntryIdentityHeaders(h) {
		t.Fatal("equivalent removal order changed readiness")
	}
	h["set"] = []any{map[string]any{"name": "Remote-User", "value": "admin"}}
	if validEntryIdentityHeaders(h) {
		t.Fatal("post-strip identity injection accepted")
	}
	delete(h, "set")
	h["remove"] = names[:len(names)-1]
	if validEntryIdentityHeaders(h) {
		t.Fatal("incomplete identity strip accepted")
	}
}

func TestEntryRejectsPolicyWithoutInternalCredentialRemoval(t *testing.T) {
	h := entryIdentityHeaderSettings()
	removed := []any{}
	for _, name := range h["remove"].([]any) {
		if name != "x-olares-caller-jwt" {
			removed = append(removed, name)
		}
	}
	h["remove"] = removed
	if validEntryIdentityHeaders(h) {
		t.Fatal("Entry is ready while permitting external caller JWT replay")
	}
}
