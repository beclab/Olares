package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestAddToSchemeRegistersTypes(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme: %v", err)
	}
	if !scheme.Recognizes(GroupVersion.WithKind("SharedRouteRegistry")) {
		t.Error("SharedRouteRegistry not registered")
	}
	if !scheme.Recognizes(GroupVersion.WithKind("SharedRouteRegistryList")) {
		t.Error("SharedRouteRegistryList not registered")
	}
}

func TestDeepCopyRoundTrip(t *testing.T) {
	in := &SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-demo", Namespace: "demo-shared"},
		Spec: SharedRouteRegistrySpec{
			RouteMode:     RouteModeGateway,
			EntranceClass: EntranceClassShared,
			HostPatterns:  []string{"a.example.com", "b.example.com"},
			Upstream:      UpstreamRef{ServiceName: "demo", Port: 8080},
		},
		Status: SharedRouteRegistryStatus{
			Conditions: []metav1.Condition{{Type: "Ready", Status: metav1.ConditionTrue}},
		},
	}
	out := in.DeepCopy()
	if out == in {
		t.Fatal("DeepCopy returned same pointer")
	}
	out.Spec.HostPatterns[0] = "mutated"
	if in.Spec.HostPatterns[0] == "mutated" {
		t.Error("HostPatterns slice not deep-copied")
	}
	out.Spec.EntranceClass = EntranceClassApplication
	if in.Spec.EntranceClass != EntranceClassShared {
		t.Errorf("EntranceClass changed on source object: %q", in.Spec.EntranceClass)
	}
	out.Status.Conditions[0].Type = "mutated"
	if in.Status.Conditions[0].Type == "mutated" {
		t.Error("Conditions slice not deep-copied")
	}
}
