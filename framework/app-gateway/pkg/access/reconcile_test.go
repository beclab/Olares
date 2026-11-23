package access

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func samplePolicy() *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "access.olares.io/v1alpha1",
		"kind":       "AppAccessPolicy",
		"metadata":   map[string]any{"name": "p1", "namespace": "os-app-access"},
		"spec": map[string]any{
			"targetNamespace": "app-b",
			"desiredState":    "Active",
			"routeMode":       "Direct",
			"serverPort":      int64(8080),
			"sourceServiceAccount": map[string]any{
				"namespace": "app-a", "name": "sa",
			},
			"rules": []any{
				map[string]any{"method": "GET", "path": "/state", "pathType": "Exact"},
			},
		},
	}}
}

func TestReconcileAccessPolicyDisabledNoWrite(t *testing.T) {
	scheme := runtime.NewScheme()
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	if err := ReconcileAccessPolicy(context.Background(), c, samplePolicy(), false); err != nil {
		t.Fatal(err)
	}
}

func TestPolicyFromUnstructuredPrefersTargetNamespace(t *testing.T) {
	p, err := PolicyFromUnstructured(samplePolicy())
	if err != nil {
		t.Fatal(err)
	}
	if p.Namespace != "app-b" {
		t.Fatalf("namespace=%s", p.Namespace)
	}
	if p.ServerPort != 8080 || p.SourceSAName != "sa" {
		t.Fatalf("spec=%+v", p)
	}
}
