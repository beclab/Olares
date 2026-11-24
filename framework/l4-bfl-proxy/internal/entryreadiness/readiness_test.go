package entryreadiness

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beclab/l4-bfl-proxy/internal/message"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func readySystemRoute(name, host string) *unstructured.Unstructured {
	reason := "Reconciled"
	if name == "sys-desktop" {
		reason = desktopAuthorizationReady
	}
	route := &unstructured.Unstructured{Object: map[string]any{
		"metadata": map[string]any{"namespace": "user-space-alice", "name": name, "generation": int64(4),
			"labels":      map[string]any{"gateway.olares.io/system-plane": "true", "app.kubernetes.io/managed-by": "app-service"},
			"annotations": map[string]any{"gateway.olares.io/extauth-user": "alice"},
		},
		"spec": map[string]any{"routeMode": "gateway", "entranceClass": "application", "hostPatterns": []any{host}},
		"status": map[string]any{"observedGeneration": int64(4), "httpRouteName": name, "conditions": []any{
			map[string]any{"type": "Ready", "status": "True", "reason": reason, "observedGeneration": int64(4)},
		}},
	}}
	route.SetGroupVersionKind(systemRouteRegistryGVK)
	return route
}

func TestEntryRoutesReadyRequiresBothObservedTenantRoutes(t *testing.T) {
	snapshot := &message.ProviderResources{}
	var reader client.Client
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	check := func(want string) {
		t.Helper()
		err := checker(snapshot, reader, readyGateway)(req)
		if want == "" && err != nil || want != "" && (err == nil || !strings.Contains(err.Error(), want)) {
			t.Fatalf("Entry readiness = %v, want %q", err, want)
		}
	}
	check("user snapshot not published")
	user := &message.UserInfo{Name: "alice", Namespace: "user-space-alice", SSL: &message.SSLConfig{Zone: "alice.example.org"}}
	snapshot.Store("test", &message.Resources{Users: []*message.UserInfo{user}})
	check("client material is unavailable")
	auth := readySystemRoute("sys-auth", "auth.alice.example.org")
	desktop := readySystemRoute("sys-desktop", "desktop.alice.example.org")
	snapshot.Store("test", &message.Resources{
		Users:     []*message.UserInfo{user, {IsEphemeral: true}},
		EntryMTLS: &message.EntryMTLSMaterial{ClientCert: "certificate", ClientKey: "key", CA: "ca"},
	})
	reader = fake.NewClientBuilder().WithObjects(auth).Build()
	check("sys-desktop")
	reader = fake.NewClientBuilder().WithObjects(auth, desktop).Build()
	check("")

	changes := []struct {
		name  string
		apply func(*unstructured.Unstructured)
	}{
		{"unobserved generation", func(route *unstructured.Unstructured) { route.SetGeneration(5) }},
		{"wrong host", func(route *unstructured.Unstructured) {
			_ = unstructured.SetNestedStringSlice(route.Object, []string{"desktop.bob.example.org"}, "spec", "hostPatterns")
		}},
		{"legacy direct", func(route *unstructured.Unstructured) {
			_ = unstructured.SetNestedField(route.Object, "direct", "spec", "routeMode")
		}},
		{"older app-service marked Desktop ready", func(route *unstructured.Unstructured) {
			_ = unstructured.SetNestedSlice(route.Object, []any{
				map[string]any{"type": "Ready", "status": "True", "reason": "Reconciled", "observedGeneration": int64(4)},
			}, "status", "conditions")
		}},
		{"unready backend", func(route *unstructured.Unstructured) {
			_ = unstructured.SetNestedSlice(route.Object, []any{
				map[string]any{"type": "Ready", "status": "False", "reason": "AwaitingDataPlane", "observedGeneration": int64(4)},
			}, "status", "conditions")
		}},
	}
	for _, change := range changes {
		t.Run(change.name, func(t *testing.T) {
			c := reader
			current := desktop.DeepCopy()
			if err := c.Get(t.Context(), client.ObjectKeyFromObject(desktop), current); err != nil {
				t.Fatal(err)
			}
			change.apply(current)
			if err := c.Update(t.Context(), current); err != nil {
				t.Fatal(err)
			}
			if err := checker(snapshot, reader, readyGateway)(req); err == nil {
				t.Fatal("invalid Desktop readiness passed")
			}
			reader = fake.NewClientBuilder().WithObjects(auth, desktop).Build()
		})
	}
}

func TestEntryGateAllowsBootstrapWithoutRegularTenants(t *testing.T) {
	snapshot := &message.ProviderResources{}
	snapshot.Store("test", &message.Resources{Users: []*message.UserInfo{{IsEphemeral: true}}})
	if err := Checker(snapshot, nil, Gateway{})(httptest.NewRequest(http.MethodGet, "/readyz", nil)); err != nil {
		t.Fatalf("bootstrap should not require a Desktop route yet: %v", err)
	}
}

func readyGateway(*message.EntryMTLSMaterial) (verifyEntry, func(), error) {
	return func(context.Context, string) error { return nil }, func() {}, nil
}
