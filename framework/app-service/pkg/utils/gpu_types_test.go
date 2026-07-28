package utils

import (
	"reflect"
	"sort"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func nodeWithLabels(labels map[string]string) *corev1.Node {
	return &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n", Labels: labels}}
}

func sortedCopy(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}

func TestNodeSupportedGPUTypes(t *testing.T) {
	tests := []struct {
		name   string
		labels map[string]string
		want   []string
	}{
		{
			name:   "no labels -> cpu only (empty)",
			labels: nil,
			want:   []string{},
		},
		{
			name:   "legacy single-value label",
			labels: map[string]string{NodeGPUTypeLabel: NvidiaCardType},
			want:   []string{NvidiaCardType},
		},
		{
			name:   "legacy none label is cpu-only",
			labels: map[string]string{NodeGPUTypeLabel: "none"},
			want:   []string{},
		},
		{
			name: "multiple existence labels (Olares One: nvidia + intel)",
			labels: map[string]string{
				NodeGPUTypeLabelPrefix + NvidiaCardType: "true",
				NodeGPUTypeLabelPrefix + IntelType:      "true",
			},
			want: []string{NvidiaCardType, IntelType},
		},
		{
			name: "legacy + new labels are deduped",
			labels: map[string]string{
				NodeGPUTypeLabel:                        NvidiaCardType,
				NodeGPUTypeLabelPrefix + NvidiaCardType: "true",
				NodeGPUTypeLabelPrefix + AMDType:        "true",
			},
			want: []string{NvidiaCardType, AMDType},
		},
		{
			name:   "legacy single-value label is read verbatim (no alias folding)",
			labels: map[string]string{NodeGPUTypeLabel: AMDType},
			want:   []string{AMDType},
		},
		{
			name: "unrelated gpu.bytetrade.io keys are ignored",
			labels: map[string]string{
				"gpu.bytetrade.io/driver":               "570.x",
				"gpu.bytetrade.io/cuda-supported":       "true",
				NodeGPUTypeLabelPrefix + NvidiaCardType: "true",
			},
			want: []string{NvidiaCardType},
		},
		{
			name: "discrete gpu existence labels are still reported as supported",
			labels: map[string]string{
				NodeGPUTypeLabelPrefix + AMDGPUType:   "true",
				NodeGPUTypeLabelPrefix + IntelGPUType: "true",
			},
			want: []string{IntelGPUType, AMDGPUType},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NodeSupportedGPUTypes(nodeWithLabels(tt.labels))
			if !reflect.DeepEqual(sortedCopy(got), sortedCopy(tt.want)) {
				t.Fatalf("NodeSupportedGPUTypes = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAllGpuTypesFromNodesMultiMode(t *testing.T) {
	list := &corev1.NodeList{Items: []corev1.Node{
		// Olares One: nvidia + intel
		*nodeWithLabels(map[string]string{
			NodeGPUTypeLabelPrefix + NvidiaCardType: "true",
			NodeGPUTypeLabelPrefix + IntelType:      "true",
		}),
		// legacy single-value node (read verbatim, no alias folding)
		*nodeWithLabels(map[string]string{NodeGPUTypeLabel: AMDType}),
		// pure cpu node contributes nothing
		*nodeWithLabels(nil),
	}}
	got, err := GetAllGpuTypesFromNodes(list)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]struct{}{NvidiaCardType: {}, IntelType: {}, AMDType: {}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetAllGpuTypesFromNodes = %v, want %v", got, want)
	}
}
