package apiserver

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/cluster"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
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

// TC-T1-3-IDM (handler): injectD2OffloaderCallerFailOpen admits and applies the
// (empty) patch on the M8 idempotent short-circuit without panicking on a bare
// Webhook, and records a caller-mode success. The deeper happy path (live
// snapshot + per-viewer secrets) is integration-level (N5-T1 e2e) because the
// Webhook dynamicClient is a concrete clientset with no assignable fake.
func TestInjectD2OffloaderCallerFailOpen_IdempotentSuccess(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: "litellm-alice", Name: "demo"},
		Spec: corev1.PodSpec{Containers: []corev1.Container{
			{Name: "app"},
			{Name: constants.D2SidecarContainerName},
		}},
	}
	raw, err := json.Marshal(pod)
	require.NoError(t, err)
	req := &admissionv1.AdmissionRequest{UID: "uid-1", Namespace: pod.Namespace, Object: runtime.RawExtension{Raw: raw}}
	resp := &admissionv1.AdmissionResponse{Allowed: true, UID: req.UID}

	h := &Handler{sidecarWebhook: &webhook.Webhook{}}
	require.NotPanics(t, func() {
		h.injectD2OffloaderCallerFailOpen(context.Background(), resp, &pod, req, nil, uuid.New(), webhook.D2InjectScenarioA)
	})

	require.True(t, resp.Allowed)
	require.Nil(t, resp.Result)
	require.NotNil(t, resp.PatchType)
}

// TC-LITE-D2: the server-mode d2 offloader is gated on meshProfile. On lite the
// EG :443 listener terminates TLS, so a shared-entrance server pod must NOT get
// the d2 offloader even when the in-cluster gateway is enabled (spec §2.2). Full
// (or absent) meshProfile keeps the existing inject decision.
func TestServerD2InjectEnabled_MeshProfileGate(t *testing.T) {
	cases := []struct {
		name string
		snap cluster.Snapshot
		want bool
	}{
		{name: "full + gateway enabled injects", snap: cluster.Snapshot{InClusterGatewayEnabled: true, MeshProfile: cluster.MeshProfileFull}, want: true},
		{name: "lite + gateway enabled skips", snap: cluster.Snapshot{InClusterGatewayEnabled: true, MeshProfile: cluster.MeshProfileLite}, want: false},
		{name: "absent meshProfile defaults full injects", snap: cluster.Snapshot{InClusterGatewayEnabled: true, MeshProfile: ""}, want: true},
		{name: "gateway disabled skips", snap: cluster.Snapshot{InClusterGatewayEnabled: false, MeshProfile: cluster.MeshProfileFull}, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, serverD2InjectEnabled(tc.snap))
		})
	}
}
