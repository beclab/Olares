package utils

import (
	"reflect"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestParseNodeIntelRegister(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    []IntelGPUEntry
		wantErr bool
	}{
		{
			name: "empty",
			in:   "",
			want: nil,
		},
		{
			name: "single igpu i915 (mem 0)",
			in:   "igpu,card0,i915,Intel® Iris® Xe Graphics,Xe,Tiger Lake,0",
			want: []IntelGPUEntry{{
				Kind: "igpu", Card: "card0", Driver: "i915",
				Name: "Intel® Iris® Xe Graphics", Architecture: "Xe", Codename: "Tiger Lake", Mem: 0,
			}},
		},
		{
			name: "multiple mixed with discrete mem",
			in:   "igpu,card0,i915,Intel® UHD Graphics,Xe,Alder Lake-P,0:dgpu,card1,xe,Intel® Arc™ B580 Graphics,Xe2,Battlemage,12884901888",
			want: []IntelGPUEntry{
				{Kind: "igpu", Card: "card0", Driver: "i915", Name: "Intel® UHD Graphics", Architecture: "Xe", Codename: "Alder Lake-P", Mem: 0},
				{Kind: "dgpu", Card: "card1", Driver: "xe", Name: "Intel® Arc™ B580 Graphics", Architecture: "Xe2", Codename: "Battlemage", Mem: 12884901888},
			},
		},
		{
			name: "empty metadata fields allowed",
			in:   "dgpu,card0,xe,,,,0",
			want: []IntelGPUEntry{{Kind: "dgpu", Card: "card0", Driver: "xe", Mem: 0}},
		},
		{
			name: "whitespace and trailing separators tolerated",
			in:   " igpu, card0, xe , Intel Graphics , Xe , Tiger Lake , 0 : ",
			want: []IntelGPUEntry{{Kind: "igpu", Card: "card0", Driver: "xe", Name: "Intel Graphics", Architecture: "Xe", Codename: "Tiger Lake", Mem: 0}},
		},
		{
			name:    "wrong field count",
			in:      "igpu,card0,i915",
			wantErr: true,
		},
		{
			name:    "unknown kind",
			in:      "vgpu,card0,i915,name,arch,code,0",
			wantErr: true,
		},
		{
			name:    "unknown driver",
			in:      "igpu,card0,amdgpu,name,arch,code,0",
			wantErr: true,
		},
		{
			name:    "non-numeric mem",
			in:      "dgpu,card0,xe,name,arch,code,lots",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseNodeIntelRegister(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (result %+v)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ParseNodeIntelRegister(%q) = %+v, want %+v", tc.in, got, tc.want)
			}
		})
	}
}

func TestIntelRegisterFromNode(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "n0",
			Annotations: map[string]string{constants.NodeIntelRegisterKey: "dgpu,card0,xe,Intel® Arc™ A770 Graphics,Xe-HPG,Alchemist,17179869184"},
		},
	}

	got, err := IntelRegisterFromNode(node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []IntelGPUEntry{{Kind: "dgpu", Card: "card0", Driver: "xe", Name: "Intel® Arc™ A770 Graphics", Architecture: "Xe-HPG", Codename: "Alchemist", Mem: 17179869184}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("IntelRegisterFromNode = %+v, want %+v", got, want)
	}

	// Absent annotation -> empty, no error.
	empty := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n1"}}
	got, err = IntelRegisterFromNode(empty)
	if err != nil || got != nil {
		t.Errorf("IntelRegisterFromNode(no annotation) = (%+v, %v), want (nil, nil)", got, err)
	}
}
