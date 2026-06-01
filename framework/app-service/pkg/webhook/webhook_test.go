package webhook

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestCreateD2OffloaderPatch_IdempotentWhenD2ContainerExists(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1ObjectMeta("user-space-alice", "demo"),
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "app"},
				{Name: constants.D2SidecarContainerName},
			},
		},
	}
	raw, err := json.Marshal(pod)
	require.NoError(t, err)
	req := &admissionv1.AdmissionRequest{
		Object: runtime.RawExtension{Raw: raw},
	}

	wh := &Webhook{}
	patch, err := wh.CreateD2OffloaderPatch(context.Background(), &pod, req, nil, uuid.New())
	require.NoError(t, err)

	var ops []map[string]any
	require.NoError(t, json.Unmarshal(patch, &ops))
	require.Len(t, ops, 0)
}

func TestDeriveViewerFromPodNS(t *testing.T) {
	viewer, err := deriveViewerFromPodNS("user-space-Alice")
	require.NoError(t, err)
	require.Equal(t, "alice", viewer)

	_, err = deriveViewerFromPodNS("kube-system")
	require.Error(t, err)
}

func TestHasD2Container(t *testing.T) {
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "app"},
				{Name: constants.D2SidecarContainerName},
			},
		},
	}
	require.True(t, hasD2Container(pod))

	pod.Spec.Containers = []corev1.Container{{Name: "app"}}
	require.False(t, hasD2Container(pod))
}

func metav1ObjectMeta(namespace, name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Namespace: namespace,
		Name:      name,
	}
}
