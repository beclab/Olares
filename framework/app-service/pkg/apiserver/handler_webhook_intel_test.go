package apiserver

import (
	"reflect"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func nodeWithRegister(name, register string) corev1.Node {
	n := corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: name}}
	if register != "" {
		n.Annotations = map[string]string{constants.NodeIntelRegisterKey: register}
	}
	return n
}

func TestSelectIntelDriverResource(t *testing.T) {
	tests := []struct {
		name         string
		nodes        []corev1.Node
		wantKind     string
		wantResource string
		wantDistinct []string
	}{
		{
			name:         "igpu on i915",
			nodes:        []corev1.Node{nodeWithRegister("n0", "igpu,card0,i915,Intel® UHD Graphics,Xe,Tiger Lake,0")},
			wantKind:     utils.IntelGPUKindIntegrated,
			wantResource: constants.IntelIGPU,
			wantDistinct: []string{utils.IntelDriverI915},
		},
		{
			name:         "igpu on xe (driver differs from kind)",
			nodes:        []corev1.Node{nodeWithRegister("n0", "igpu,card0,xe,,,,0")},
			wantKind:     utils.IntelGPUKindIntegrated,
			wantResource: constants.IntelGPU,
			wantDistinct: []string{utils.IntelDriverXe},
		},
		{
			name:         "dgpu on i915 (driver differs from kind)",
			nodes:        []corev1.Node{nodeWithRegister("n0", "dgpu,card0,i915,,,,0")},
			wantKind:     utils.IntelGPUKindDiscrete,
			wantResource: constants.IntelIGPU,
			wantDistinct: []string{utils.IntelDriverI915},
		},
		{
			name: "only matching kind considered",
			nodes: []corev1.Node{
				nodeWithRegister("n0", "igpu,card0,i915,,,,0:dgpu,card1,xe,Intel® Arc™ B580 Graphics,Xe2,Battlemage,12884901888"),
			},
			wantKind:     utils.IntelGPUKindDiscrete,
			wantResource: constants.IntelGPU,
			wantDistinct: []string{utils.IntelDriverXe},
		},
		{
			name: "mixed cluster picks xe by priority",
			nodes: []corev1.Node{
				nodeWithRegister("n0", "igpu,card0,xe,,,,0"),
				nodeWithRegister("n1", "igpu,card0,i915,,,,0"),
			},
			wantKind:     utils.IntelGPUKindIntegrated,
			wantResource: constants.IntelGPU,
			wantDistinct: []string{utils.IntelDriverXe, utils.IntelDriverI915},
		},
		{
			name:         "no matching kind -> empty",
			nodes:        []corev1.Node{nodeWithRegister("n0", "dgpu,card0,xe,,,,0")},
			wantKind:     utils.IntelGPUKindIntegrated,
			wantResource: "",
			wantDistinct: nil,
		},
		{
			name:         "malformed annotation skipped -> empty",
			nodes:        []corev1.Node{nodeWithRegister("n0", "garbage,entry")},
			wantKind:     utils.IntelGPUKindIntegrated,
			wantResource: "",
			wantDistinct: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resource, distinct := selectIntelDriverResource(tc.nodes, tc.wantKind)
			if resource != tc.wantResource {
				t.Errorf("resource = %q, want %q", resource, tc.wantResource)
			}
			if !reflect.DeepEqual(distinct, tc.wantDistinct) {
				t.Errorf("distinct = %v, want %v", distinct, tc.wantDistinct)
			}
		})
	}
}
