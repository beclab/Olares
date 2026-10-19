package webhook

import (
	"reflect"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

func TestAddGpuResourceLimits(t *testing.T) {
	mem := ptr.To("8192")
	tests := []struct {
		name            string
		gpu             GPUResources
		existingLimits  map[string]string
		runtimeClass    *string
		wantRuntime     *patchOp
		wantLimits      map[string]string
		wantLimitsPatch string
	}{
		{
			name:            "no resource injects nothing",
			gpu:             GPUResources{MemoryResource: constants.NvidiaGPUMem, Memory: mem, RuntimeClass: "nvidia"},
			wantLimitsPatch: "",
		},
		{
			name:            "device, memory and runtime class",
			gpu:             GPUResources{Resource: constants.NvidiaGPU, MemoryResource: constants.NvidiaGPUMem, Memory: mem, RuntimeClass: "nvidia"},
			wantRuntime:     &patchOp{Op: constants.PatchOpAdd, Path: runtimeClassPath, Value: "nvidia"},
			wantLimits:      map[string]string{constants.NvidiaGPU: "1", constants.NvidiaGPUMem: "8192"},
			wantLimitsPatch: constants.PatchOpAdd,
		},
		{
			name:            "existing runtime class is replaced",
			gpu:             GPUResources{Resource: constants.NvidiaGPU, RuntimeClass: "nvidia"},
			runtimeClass:    ptr.To("kata"),
			wantRuntime:     &patchOp{Op: constants.PatchOpReplace, Path: runtimeClassPath, Value: "nvidia"},
			wantLimits:      map[string]string{constants.NvidiaGPU: "1"},
			wantLimitsPatch: constants.PatchOpAdd,
		},
		{
			name:            "memory without a memory resource is dropped",
			gpu:             GPUResources{Resource: constants.AMDGPU, Memory: mem},
			wantLimits:      map[string]string{constants.AMDGPU: "1"},
			wantLimitsPatch: constants.PatchOpAdd,
		},
		{
			name:            "memory resource without an amount is dropped",
			gpu:             GPUResources{Resource: constants.NvidiaGPU, MemoryResource: constants.NvidiaGPUMem},
			wantLimits:      map[string]string{constants.NvidiaGPU: "1"},
			wantLimitsPatch: constants.PatchOpAdd,
		},
		{
			name:            "existing limits keep their own keys",
			gpu:             GPUResources{Resource: constants.NvidiaGPU, MemoryResource: constants.NvidiaGPUMem, Memory: mem},
			existingLimits:  map[string]string{"cpu": "2", constants.NvidiaGPUMem: "1024"},
			wantLimits:      map[string]string{"cpu": "2", constants.NvidiaGPU: "1", constants.NvidiaGPUMem: "8192"},
			wantLimitsPatch: constants.PatchOpReplace,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{
				RuntimeClassName: tt.runtimeClass,
				Containers:       []corev1.Container{makeContainer("app", tt.existingLimits, nil)},
			}}
			patches, err := addGpuResourceLimits(tpl, true, nil, tt.gpu)
			if err != nil {
				t.Fatalf("addGpuResourceLimits: %v", err)
			}
			var runtime, limits *patchOp
			for i := range patches {
				switch patches[i].Path {
				case runtimeClassPath:
					runtime = &patches[i]
				case "/spec/template/spec/containers/0/resources/limits":
					limits = &patches[i]
				}
			}
			if !reflect.DeepEqual(runtime, tt.wantRuntime) {
				t.Fatalf("runtime class patch = %+v, want %+v", runtime, tt.wantRuntime)
			}
			if tt.wantLimitsPatch == "" {
				if limits != nil {
					t.Fatalf("unexpected limits patch %+v", limits)
				}
				return
			}
			if limits == nil || limits.Op != tt.wantLimitsPatch {
				t.Fatalf("limits patch = %+v, want op %s", limits, tt.wantLimitsPatch)
			}
			got := map[string]string{}
			switch value := limits.Value.(type) {
			case map[string]interface{}:
				for k, v := range value {
					got[k] = v.(string)
				}
			case map[string]string:
				got = value
			}
			if !reflect.DeepEqual(got, tt.wantLimits) {
				t.Fatalf("limits = %v, want %v", got, tt.wantLimits)
			}
		})
	}
}
