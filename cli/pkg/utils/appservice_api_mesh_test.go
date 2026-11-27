package utils

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func TestPrepareAppServiceAPIValues(t *testing.T) {
	tenants := []AppServiceAPITenant{{"bob", "normal"}, {"alice", "owner"}, {"carol", "admin"}}
	previous := map[string]any{
		"appServiceAPIMesh": map[string]any{"activate": true, "marketLifecycle": true, "systemEnvWrites": true, "tenants": []any{"alice"}, "userProvisioners": []any{"bob", "carol", "deleted"}},
		"marketBackendMTLS": map[string]any{"prepare": true, "activate": true, "tenants": []any{"alice"}},
		"marketLinkerd":     true,
	}
	fresh := map[string]any{}
	if err := prepareAppServiceAPIValues(fresh, nil, tenants); err != nil {
		t.Fatal(err)
	}
	wantFresh := map[string]any{"prepare": true,
		"tenants": []string{"alice", "bob", "carol"}, "userProvisioners": []string{"alice"}}
	if !reflect.DeepEqual(fresh["appServiceAPIMesh"], wantFresh) || fresh["marketBackendMTLS"] != nil {
		t.Fatalf("preparation activated a listener or promoted a tenant: %v", fresh)
	}

	upgrade := map[string]any{}
	if err := prepareAppServiceAPIValues(upgrade, previous, tenants); err != nil {
		t.Fatal(err)
	}
	wantAPI := map[string]any{"prepare": true, "activate": true, "marketLifecycle": true, "systemEnvWrites": true,
		"marketLifecycleTenants": []string{"alice"}, "systemEnvWriterTenants": []string{"alice"},
		"tenants": []string{"alice", "bob", "carol"}, "userProvisioners": []string{"alice", "carol"}}
	wantMarket := map[string]any{"prepare": true, "activate": true, "tenants": []string{"alice", "bob", "carol"}}
	if !reflect.DeepEqual(upgrade["appServiceAPIMesh"], wantAPI) ||
		!reflect.DeepEqual(upgrade["marketBackendMTLS"], wantMarket) || upgrade["marketLinkerd"] != true {
		t.Fatalf("upgrade drifted from the deployed security posture: %v", upgrade)
	}
	if old := previous["marketBackendMTLS"].(map[string]any)["tenants"]; !reflect.DeepEqual(old, []any{"alice"}) {
		t.Fatal("preparation mutated the deployed release values")
	}
}

func TestInvalidAppServiceRolloutLeavesValuesUntouchedAndCanRetry(t *testing.T) {
	for _, tenants := range [][]AppServiceAPITenant{
		nil, {{"bob", "normal"}}, {{"alice", "owner"}, {"alice", "normal"}},
		{{"alice", "owner"}, {"../mallory", "normal"}}, {{"alice", "owner"}, {"bob", "owner"}},
	} {
		values := map[string]any{"unrelated": true}
		if err := prepareAppServiceAPIValues(values, nil, tenants); err == nil {
			t.Fatalf("invalid inventory accepted: %v", tenants)
		}
		if !reflect.DeepEqual(values, map[string]any{"unrelated": true}) {
			t.Fatalf("failed preparation partially applied values: %v", values)
		}
		if err := prepareAppServiceAPIValues(values, nil, []AppServiceAPITenant{{"alice", "owner"}, {"bob", "normal"}}); err != nil {
			t.Fatalf("valid retry failed: %v", err)
		}
	}
}

func TestAppServicePolicyPreparationRetriesMissingAuthorization(t *testing.T) {
	const ns = "os-framework"
	object := func(version, kind, name string, fields map[string]any) *unstructured.Unstructured {
		obj := &unstructured.Unstructured{Object: fields}
		obj.SetAPIVersion(version)
		obj.SetKind(kind)
		obj.SetName(name)
		obj.SetNamespace(ns)
		return obj
	}
	server := object("policy.linkerd.io/v1beta3", "Server", "app-service-api", map[string]any{
		"spec": map[string]any{"port": int64(6755), "accessPolicy": "deny",
			"podSelector": map[string]any{"matchLabels": map[string]any{
				"tier": "app-service", "traffic.olares.io/app-service-api-mtls": "required",
			}},
		},
	})
	objects := []runtime.Object{server}
	var delayed *unstructured.Unstructured
	tenants := []AppServiceAPITenant{{"alice", "owner"}, {"bob", "normal"}}
	for _, name := range []string{"market-image-info", "users-alice", "user-apps-alice", "settings-user-create-alice", "users-bob", "user-apps-bob"} {
		objects = append(objects, object("gateway.networking.k8s.io/v1", "HTTPRoute", name, map[string]any{
			"status": map[string]any{"parents": []any{map[string]any{
				"controllerName": "linkerd.io/policy-controller",
				"parentRef":      map[string]any{"group": "policy.linkerd.io", "kind": "Server", "name": "app-service-api"},
				"conditions":     []any{map[string]any{"type": "Accepted", "status": "True"}},
			}}},
		}), object("policy.linkerd.io/v1alpha1", "MeshTLSAuthentication", name, map[string]any{}))
		authorization := object("policy.linkerd.io/v1alpha1", "AuthorizationPolicy", name, map[string]any{})
		if name == "user-apps-bob" {
			delayed = authorization
		} else {
			objects = append(objects, authorization)
		}
	}
	client := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), objects...)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()
	if err := waitForPreparedAppServiceAPI(ctx, client, tenants); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("missing tenant authorization did not stop rollout: %v", err)
	}
	policy := client.Resource(schema.GroupVersionResource{Group: "policy.linkerd.io", Version: "v1alpha1", Resource: "authorizationpolicies"}).Namespace(ns)
	if _, err := policy.Create(t.Context(), delayed, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := waitForPreparedAppServiceAPI(t.Context(), client, tenants); err != nil {
		t.Fatalf("preparation never recovered after policy appeared: %v", err)
	}
}
