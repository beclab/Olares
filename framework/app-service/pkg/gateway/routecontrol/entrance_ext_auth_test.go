package routecontrol

import (
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
)

func TestNeedsEntranceExtAuth(t *testing.T) {
	if needsEntranceExtAuth(nil) {
		t.Fatal("nil")
	}
	if !needsEntranceExtAuth(&srrv1alpha1.SharedRouteRegistry{
		Spec: srrv1alpha1.SharedRouteRegistrySpec{EntranceClass: srrv1alpha1.EntranceClassApplication},
	}) {
		t.Fatal("application class needs extAuth")
	}
	if needsEntranceExtAuth(&srrv1alpha1.SharedRouteRegistry{
		Spec: srrv1alpha1.SharedRouteRegistrySpec{EntranceClass: srrv1alpha1.EntranceClassShared},
	}) {
		t.Fatal("shared class uses JWT")
	}
}

func TestDesiredEntranceExtAuthPolicy(t *testing.T) {
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "app-demo-web", Namespace: "demo-alice"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			EntranceClass: srrv1alpha1.EntranceClassApplication,
			HostPatterns:  []string{"demo.alice.olares.com"},
			Upstream:      srrv1alpha1.UpstreamRef{ServiceName: "demo"},
		},
	}
	pol := desiredEntranceExtAuthPolicy(srr, "alice")
	if pol.GetName() != mesh.EntranceExtAuthPolicyName(srr.Name) {
		t.Fatalf("name = %q", pol.GetName())
	}
	spec, _, _ := unstructuredNestedMap(pol.Object, "spec")
	ext, ok := spec["extAuth"].(map[string]any)
	if !ok {
		t.Fatalf("missing extAuth: %#v", spec)
	}
	if failOpen, _ := ext["failOpen"].(bool); failOpen {
		t.Fatal("extAuth must be fail-closed")
	}
	httpCfg, _ := ext["http"].(map[string]any)
	if path, _ := httpCfg["pathOverride"].(string); path != autheliaVerifyPathOverride {
		t.Fatalf("pathOverride = %q, want %q", path, autheliaVerifyPathOverride)
	}
	if _, hasPath := httpCfg["path"]; hasPath {
		t.Fatal("must not set path when pathOverride is used")
	}
	headers, _ := ext["headersToExtAuth"].([]any)
	if len(headers) == 0 {
		t.Fatal("headersToExtAuth required")
	}
	if !strings.Contains(pol.GetLabels()["gateway.olares.io/auth-kind"], "entrance-ext-auth") {
		t.Fatalf("labels = %#v", pol.GetLabels())
	}
}

func unstructuredNestedMap(obj map[string]any, fields ...string) (map[string]any, bool, error) {
	cur := any(obj)
	for _, f := range fields {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false, nil
		}
		cur, ok = m[f]
		if !ok {
			return nil, false, nil
		}
	}
	out, ok := cur.(map[string]any)
	return out, ok, nil
}
