package utils

import (
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const gib = int64(1 << 30)

func annotatedMeta(pairs ...string) metav1.ObjectMeta {
	annotations := map[string]string{}
	for i := 0; i+1 < len(pairs); i += 2 {
		annotations[pairs[i]] = pairs[i+1]
	}
	return metav1.ObjectMeta{Annotations: annotations}
}

func gpuMemResourceList(mib string) corev1.ResourceList {
	return corev1.ResourceList{nvidiaGPUMem: apiresource.MustParse(mib)}
}

// The two channels a chart can declare its GPU-memory quota through disagree on
// units, which is the whole reason the totals are normalized to bytes:
// nvidia.com/gpumem is a bare MiB count (HAMi's convention) while the pod
// annotation is an ordinary quantity, read the same way as the manifest's
// requiredGPUMemory it ends up backfilling.
func TestPodGPUMemoryBytesSources(t *testing.T) {
	tests := []struct {
		name     string
		meta     metav1.ObjectMeta
		requests corev1.ResourceList
		limits   corev1.ResourceList
		wantReq  int64
		wantLim  int64
	}{
		{
			name:     "nvidia gpumem resource is a MiB count",
			meta:     metav1.ObjectMeta{},
			requests: gpuMemResourceList("23552"),
			limits:   gpuMemResourceList("23552"),
			wantReq:  23 * gib,
			wantLim:  23 * gib,
		},
		{
			name: "annotation is a plain quantity",
			meta: annotatedMeta(
				constants.PodRequiredGPUMemory, "23Gi",
				constants.PodLimitedGPUMemory, "32Gi",
			),
			wantReq: 23 * gib,
			wantLim: 32 * gib,
		},
		{
			// A template rendered for amd-gpu may still carry the
			// nvidia.com/gpumem of its nvidia branch. Counting both would
			// double the app's declared demand.
			name: "annotation wins over the resource",
			meta: annotatedMeta(
				constants.PodRequiredGPUMemory, "23Gi",
				constants.PodLimitedGPUMemory, "23Gi",
			),
			requests: gpuMemResourceList("4096"),
			limits:   gpuMemResourceList("4096"),
			wantReq:  23 * gib,
			wantLim:  23 * gib,
		},
		{
			name:    "neither channel declares anything",
			meta:    metav1.ObjectMeta{},
			wantReq: 0,
			wantLim: 0,
		},
		{
			// A typo must not fail the install; it degrades to "no declared
			// demand", which is where such an app stood before the annotation
			// existed.
			name:    "unparseable annotation is ignored",
			meta:    annotatedMeta(constants.PodRequiredGPUMemory, "23 gigabytes"),
			wantReq: 0,
			wantLim: 0,
		},
		{
			name:    "non-positive annotation is ignored",
			meta:    annotatedMeta(constants.PodRequiredGPUMemory, "-1"),
			wantReq: 0,
			wantLim: 0,
		},
		{
			name:     "annotation on one side only leaves the other to the resource",
			meta:     annotatedMeta(constants.PodLimitedGPUMemory, "32Gi"),
			requests: gpuMemResourceList("23552"),
			limits:   gpuMemResourceList("23552"),
			wantReq:  23 * gib,
			wantLim:  32 * gib,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReq, gotLim := podGPUMemoryBytes(tt.meta, tt.requests, tt.limits)
			if gotReq.Value() != tt.wantReq {
				t.Errorf("request = %d bytes, want %d", gotReq.Value(), tt.wantReq)
			}
			if gotLim.Value() != tt.wantLim {
				t.Errorf("limit = %d bytes, want %d", gotLim.Value(), tt.wantLim)
			}
		})
	}
}

// The discrete AMD/Intel cards have no memory extended resource to carry the
// quota, so without the annotation channel a chart for them has no way to
// declare one and the auto-resource sentinel resolves to zero — the state that
// left a healthy card unbindable on resume.
func TestAnnotationIsTheOnlyChannelWithoutNvidiaGPUMem(t *testing.T) {
	meta := annotatedMeta(constants.PodRequiredGPUMemory, "23Gi")
	req, _ := podGPUMemoryBytes(meta, corev1.ResourceList{}, corev1.ResourceList{})
	if req.Value() != 23*gib {
		t.Fatalf("request = %d bytes, want %d", req.Value(), 23*gib)
	}

	bare, _ := podGPUMemoryBytes(metav1.ObjectMeta{}, corev1.ResourceList{}, corev1.ResourceList{})
	if bare.Value() != 0 {
		t.Fatalf("without the annotation the quota should be unknown, got %d bytes", bare.Value())
	}
}
