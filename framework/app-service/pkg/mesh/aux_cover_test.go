package mesh

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestEvaluateEntranceAuxCoveredVacuous(t *testing.T) {
	targets := []EntranceAuxTarget{{
		Namespace: "ns", SRRName: "app", HTTPRoute: "app",
		Needs: EntranceAuxNeeds{},
	}}
	if !EvaluateEntranceAuxCovered(targets, func(EntranceAuxTarget) bool { return false }) {
		t.Fatal("vacuous needs must be Ready without callback success")
	}
}

func TestEvaluateEntranceAuxCoveredWSMissing(t *testing.T) {
	targets := []EntranceAuxTarget{{
		Namespace: "ns", SRRName: "app", HTTPRoute: "app",
		Needs: EntranceAuxNeeds{WS: true},
	}}
	if EvaluateEntranceAuxCovered(targets, func(EntranceAuxTarget) bool { return false }) {
		t.Fatal("ws declared without route must be False")
	}
}

func TestImagesRouteBypassesExtAuth(t *testing.T) {
	good := &unstructured.Unstructured{Object: map[string]any{
		"metadata": map[string]any{
			"labels":      map[string]any{"gateway.olares.io/auth-kind": AuthKindEntranceImages},
			"annotations": map[string]any{AuxExtAuthBypassAnnotation: "true"},
		},
		"spec": map[string]any{
			"rules": []any{
				map[string]any{
					"matches": []any{
						map[string]any{"path": map[string]any{"type": "PathPrefix", "value": AuxImagesPathPrefix}},
					},
					"backendRefs": []any{
						map[string]any{"name": AuxImagesServiceName, "port": int64(8080), "namespace": "user-system-alice"},
					},
				},
			},
		},
	}}
	if !ImagesRouteBypassesExtAuth(good) {
		t.Fatal("expected /images/upload ExtAuth bypass ready")
	}
	noBypass := good.DeepCopy()
	noBypass.SetAnnotations(map[string]string{})
	if ImagesRouteBypassesExtAuth(noBypass) {
		t.Fatal("missing bypass annotation must fail (ExtAuth not excluded)")
	}
}

func TestAuxTargetReadyFromRoutesWS(t *testing.T) {
	main := &unstructured.Unstructured{Object: map[string]any{
		"spec": map[string]any{
			"rules": []any{
				map[string]any{
					"matches": []any{
						map[string]any{"path": map[string]any{"type": "PathPrefix", "value": "/ws"}},
					},
					"backendRefs": []any{
						map[string]any{"name": "demo-deenvy-aux", "port": int64(40010)},
					},
				},
			},
		},
	}}
	tOk := EntranceAuxTarget{Needs: EntranceAuxNeeds{WS: true}}
	if !AuxTargetReadyFromRoutes(tOk, main, nil) {
		t.Fatal("ws rule should be Ready")
	}
	tMiss := EntranceAuxTarget{Needs: EntranceAuxNeeds{WS: true}}
	if AuxTargetReadyFromRoutes(tMiss, &unstructured.Unstructured{Object: map[string]any{}}, nil) {
		t.Fatal("missing ws rule must be False")
	}
}

func TestParseAuxCapabilitiesAnnotation(t *testing.T) {
	n := ParseAuxCapabilitiesAnnotation("ws,upload,images")
	if !n.WS || !n.Upload || !n.Images {
		t.Fatalf("got %#v", n)
	}
}
