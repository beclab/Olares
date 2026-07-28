package webhook

import (
	"encoding/json"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

// makeContainer is a small helper that builds a corev1.Container with the
// given limits/requests resource lists, parsing each value as a Quantity.
func makeContainer(name string, limits, requests map[string]string) corev1.Container {
	c := corev1.Container{Name: name}
	if limits != nil {
		c.Resources.Limits = corev1.ResourceList{}
		for k, v := range limits {
			c.Resources.Limits[corev1.ResourceName(k)] = resource.MustParse(v)
		}
	}
	if requests != nil {
		c.Resources.Requests = corev1.ResourceList{}
		for k, v := range requests {
			c.Resources.Requests[corev1.ResourceName(k)] = resource.MustParse(v)
		}
	}
	return c
}

func TestRemoveGpuResources_NilTemplate(t *testing.T) {
	if got := removeGpuResources(nil); got != nil {
		t.Fatalf("expected nil patches for nil template, got %v", got)
	}
}

func TestRemoveGpuResources_NoGpu_NoOp(t *testing.T) {
	tpl := &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				makeContainer("c0",
					map[string]string{"cpu": "2", "memory": "1Gi"},
					map[string]string{"cpu": "100m", "memory": "256Mi"},
				),
			},
		},
	}
	if got := removeGpuResources(tpl); len(got) != 0 {
		t.Fatalf("expected no patches for CPU-only container, got %d: %#v", len(got), got)
	}
}

func TestRemoveGpuResources_DropsNvidiaGpuAndGpumem(t *testing.T) {
	tpl := &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			RuntimeClassName: ptr.To("nvidia"),
			Containers: []corev1.Container{
				makeContainer("app",
					map[string]string{
						"cpu":                  "2",
						"memory":               "8Gi",
						constants.NvidiaGPU:    "1",
						constants.NvidiaGPUMem: "1024",
					},
					map[string]string{
						"cpu":                  "50m",
						"memory":               "512Mi",
						constants.NvidiaGPU:    "1",
						constants.NvidiaGPUMem: "1024",
					},
				),
			},
		},
	}

	patches := removeGpuResources(tpl)
	bytePatches, err := json.Marshal(patches)
	if err != nil {
		t.Fatalf("marshal patches: %v", err)
	}

	wantPaths := []string{
		"/spec/template/spec/containers/0/resources/limits/nvidia.com~1gpu",
		"/spec/template/spec/containers/0/resources/limits/nvidia.com~1gpumem",
		"/spec/template/spec/containers/0/resources/requests/nvidia.com~1gpu",
		"/spec/template/spec/containers/0/resources/requests/nvidia.com~1gpumem",
		"/spec/template/spec/runtimeClassName",
	}
	got := map[string]bool{}
	for _, p := range patches {
		if p.Op != constants.PatchOpRemove {
			t.Errorf("patch %s op=%s, want %s", p.Path, p.Op, constants.PatchOpRemove)
		}
		got[p.Path] = true
	}
	for _, want := range wantPaths {
		if !got[want] {
			t.Errorf("missing remove patch for %s; full patches=%s", want, string(bytePatches))
		}
	}
	if len(patches) != len(wantPaths) {
		t.Errorf("expected %d patches, got %d: %s", len(wantPaths), len(patches), string(bytePatches))
	}
}

func TestRemoveGpuResources_KeepsNonNvidiaRuntimeClass(t *testing.T) {
	tpl := &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			RuntimeClassName: ptr.To("kata"),
			Containers: []corev1.Container{
				makeContainer("c0", map[string]string{constants.NvidiaGPU: "1"}, nil),
			},
		},
	}

	patches := removeGpuResources(tpl)
	for _, p := range patches {
		if p.Path == runtimeClassPath {
			t.Fatalf("did not expect runtimeClassName remove patch for non-nvidia runtime class, got %+v", p)
		}
	}
}

func TestRemoveGpuResources_AmdResources(t *testing.T) {
	tpl := &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				makeContainer("c0",
					map[string]string{constants.AMDGPU: "1", constants.AMDAPU: "1"},
					nil,
				),
			},
		},
	}
	patches := removeGpuResources(tpl)
	want := map[string]bool{
		"/spec/template/spec/containers/0/resources/limits/amd.com~1gpu": false,
		"/spec/template/spec/containers/0/resources/limits/amd.com~1apu": false,
	}
	for _, p := range patches {
		if _, ok := want[p.Path]; ok {
			want[p.Path] = true
		}
	}
	for path, seen := range want {
		if !seen {
			t.Errorf("missing remove patch for %s", path)
		}
	}
}

func TestRemoveGpuResources_MultiContainer(t *testing.T) {
	tpl := &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				makeContainer("c0", map[string]string{constants.NvidiaGPU: "1"}, nil),
				makeContainer("sidecar", map[string]string{"cpu": "100m"}, nil),
				makeContainer("c2",
					map[string]string{constants.NvidiaGPU: "1", constants.NvidiaGPUMem: "1024"},
					map[string]string{constants.NvidiaGPU: "1"},
				),
			},
		},
	}
	patches := removeGpuResources(tpl)

	want := []string{
		"/spec/template/spec/containers/0/resources/limits/nvidia.com~1gpu",
		"/spec/template/spec/containers/2/resources/limits/nvidia.com~1gpu",
		"/spec/template/spec/containers/2/resources/limits/nvidia.com~1gpumem",
		"/spec/template/spec/containers/2/resources/requests/nvidia.com~1gpu",
	}
	got := map[string]bool{}
	for _, p := range patches {
		got[p.Path] = true
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("missing remove patch for %s; all=%+v", w, patches)
		}
	}
	// container index 1 has no GPU resources, must not be referenced.
	for _, p := range patches {
		if p.Path == "/spec/template/spec/containers/1/resources/limits/nvidia.com~1gpu" {
			t.Errorf("did not expect remove patch for sidecar container, got %+v", p)
		}
	}
}

func TestCreateCleanupPatchForDeployment_NoOp(t *testing.T) {
	tpl := &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{makeContainer("c0", map[string]string{"cpu": "1"}, nil)},
		},
	}
	b, err := CreateCleanupPatchForDeployment(tpl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b != nil {
		t.Fatalf("expected nil bytes when nothing to clean, got %s", string(b))
	}
}

func TestCreateCleanupPatchForDeployment_EmitsValidJSONPatch(t *testing.T) {
	tpl := &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			RuntimeClassName: ptr.To("nvidia"),
			Containers: []corev1.Container{
				makeContainer("c0",
					map[string]string{constants.NvidiaGPU: "1", constants.NvidiaGPUMem: "1024"},
					map[string]string{constants.NvidiaGPU: "1", constants.NvidiaGPUMem: "1024"},
				),
			},
		},
	}
	b, err := CreateCleanupPatchForDeployment(tpl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(b) == 0 {
		t.Fatalf("expected non-empty patch bytes")
	}
	var decoded []patchOp
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("patch bytes are not valid JSON: %v\n%s", err, string(b))
	}
	if len(decoded) == 0 {
		t.Fatalf("expected at least one patch op, got 0")
	}
	for _, p := range decoded {
		if p.Op != constants.PatchOpRemove {
			t.Errorf("op=%s, want %s; patch=%+v", p.Op, constants.PatchOpRemove, p)
		}
	}
}

func TestJSONPointerEscape(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"nvidia.com/gpu", "nvidia.com~1gpu"},
		{"a/b/c", "a~1b~1c"},
		{"with~tilde", "with~0tilde"},
		{"with~/slash", "with~0~1slash"},
		{"plain", "plain"},
	}
	for _, tc := range cases {
		if got := jsonPointerEscape(tc.in); got != tc.want {
			t.Errorf("jsonPointerEscape(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
