package apiserver

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/webhook"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// TC-A1-1 (handler level): when the d2 offloader prerequisite fails (here the
// viewer cannot be derived from a non-user-space namespace), injection fails
// open: the response stays Allowed with no d2 patch applied, instead of being
// rejected. The shared-entrance label is patched earlier in mutate and is
// unaffected by this path.
func TestInjectD2OffloaderFailOpen_ViewerUnderive(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: "kube-system", Name: "demo"},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}}},
	}
	raw, err := json.Marshal(pod)
	require.NoError(t, err)
	req := &admissionv1.AdmissionRequest{
		UID:       "uid-1",
		Namespace: pod.Namespace,
		Object:    runtime.RawExtension{Raw: raw},
	}
	resp := &admissionv1.AdmissionResponse{Allowed: true, UID: req.UID}

	h := &Handler{sidecarWebhook: &webhook.Webhook{}}
	require.NotPanics(t, func() {
		h.injectD2OffloaderFailOpen(context.Background(), resp, &pod, req, nil, uuid.New())
	})

	require.True(t, resp.Allowed)
	require.Nil(t, resp.Patch)
	require.Nil(t, resp.Result)
}
