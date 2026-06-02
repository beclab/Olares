package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
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

func TestClassifyD2SkipReason(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"nil", nil, "other"},
		{"snapshot", fmt.Errorf("ctx: %w", ErrD2SnapshotUnavailable), "snapshot_error"},
		{"viewer", fmt.Errorf("ctx: %w", ErrD2ViewerUnderive), "viewer_underive"},
		{"tls", fmt.Errorf("ctx: %w", ErrD2TLSSecretMissing), "tls_secret_missing"},
		{"other", errors.New("boom"), "other"},
		{"snapshot_wins", errors.Join(ErrD2SnapshotUnavailable, ErrD2ViewerUnderive), "snapshot_error"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.want, ClassifyD2SkipReason(c.err))
		})
	}
}

func TestRecordD2InjectSkipped(t *testing.T) {
	before := testutil.ToFloat64(d2InjectSkippedTotal.WithLabelValues("viewer_underive"))
	RecordD2InjectSkipped("viewer_underive")
	after := testutil.ToFloat64(d2InjectSkippedTotal.WithLabelValues("viewer_underive"))
	require.Equal(t, before+1, after)

	emptyBefore := testutil.ToFloat64(d2InjectSkippedTotal.WithLabelValues("other"))
	RecordD2InjectSkipped("")
	emptyAfter := testutil.ToFloat64(d2InjectSkippedTotal.WithLabelValues("other"))
	require.Equal(t, emptyBefore+1, emptyAfter)
}

func TestDeriveViewerFromPodNS_UnderiveSentinel(t *testing.T) {
	_, err := deriveViewerFromPodNS("kube-system")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrD2ViewerUnderive))

	_, err = deriveViewerFromPodNS("user-space-")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrD2ViewerUnderive))
}

// TC-A1-1: viewer derivation failure surfaces ErrD2ViewerUnderive from
// CreateD2OffloaderPatch (no client touched: derivation fails first).
func TestCreateD2OffloaderPatch_ViewerUnderiveSentinel(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1ObjectMeta("kube-system", "demo"),
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}}},
	}
	raw, err := json.Marshal(pod)
	require.NoError(t, err)
	req := &admissionv1.AdmissionRequest{Object: runtime.RawExtension{Raw: raw}}

	wh := &Webhook{}
	_, err = wh.CreateD2OffloaderPatch(context.Background(), &pod, req, nil, uuid.New())
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrD2ViewerUnderive))
}

// TC-A1-2: a server namespace with no per-viewer tls replica yields
// ErrD2TLSSecretMissing from resolveViewerAllowset.
func TestResolveViewerAllowset_TLSSecretMissingSentinel(t *testing.T) {
	wh := &Webhook{kubeClient: fake.NewSimpleClientset()}
	pod := &corev1.Pod{ObjectMeta: metav1ObjectMeta("ollamav3-shared", "demo")}

	_, err := wh.resolveViewerAllowset(context.Background(), pod)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrD2TLSSecretMissing))
}

func TestResolveViewerAllowset_FromReplicaSecret(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1ObjectMeta("ollamav3-shared", constants.D2SharedTLSSecretNamePrefix+"brucedai"),
	}
	wh := &Webhook{kubeClient: fake.NewSimpleClientset(secret)}
	pod := &corev1.Pod{ObjectMeta: metav1ObjectMeta("ollamav3-shared", "demo")}

	allowset, err := wh.resolveViewerAllowset(context.Background(), pod)
	require.NoError(t, err)
	require.Equal(t, []string{"brucedai"}, allowset)
}
