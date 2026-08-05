package mesh

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestEvaluateEntranceProbeBypassReadyVacuous(t *testing.T) {
	targets := []EntranceProbeBypassTarget{{Namespace: "ns", SRRName: "app", ExpectedPaths: nil}}
	if !EvaluateEntranceProbeBypassReady(targets, func(EntranceProbeBypassTarget) bool { return false }) {
		t.Fatal("no probe paths must be Ready")
	}
}

func TestProbeBypassObjectsReadyRequiresUA(t *testing.T) {
	route := &unstructured.Unstructured{Object: map[string]any{
		"metadata": map[string]any{
			"annotations": map[string]any{ProbePathsAnnotation: "/healthz"},
		},
	}}
	polNoUA := &unstructured.Unstructured{Object: map[string]any{
		"metadata": map[string]any{"labels": map[string]any{}},
		"spec":     map[string]any{"lua": []any{map[string]any{"type": "Inline", "inline": "x"}}},
	}}
	if ProbeBypassObjectsReady(route, polNoUA, []string{"/healthz"}) {
		t.Fatal("missing UA label must fail")
	}
	pol := &unstructured.Unstructured{Object: map[string]any{
		"metadata": map[string]any{"labels": map[string]any{ProbeUARequiredLabel: "true"}},
		"spec":     map[string]any{"lua": []any{map[string]any{"type": "Inline", "inline": "user-agent"}}},
	}}
	if !ProbeBypassObjectsReady(route, pol, []string{"/healthz"}) {
		t.Fatal("expected ready")
	}
}
