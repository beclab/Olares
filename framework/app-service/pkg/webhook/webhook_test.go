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

func nsWithLabels(name string, labels map[string]string) *corev1.Namespace {
	return &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name, Labels: labels}}
}

func podWithLabels(namespace string, labels map[string]string) *corev1.Pod {
	return &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: "demo", Labels: labels}}
}

// TC-A2-1..A2-4 and TC-A2-6: viewer is the namespace ns-owner when present,
// otherwise the user-space-/user-system- prefix segment, otherwise the pod
// owner label (lowercased); all-absent (and <app>-<user> without an owner)
// yields an error. The pilot value "brucedai" equals that cluster's owner; the
// assertion is on the derived viewer matching the source, not the literal.
func TestDeriveViewerFromPodNS(t *testing.T) {
	cases := []struct {
		name       string
		ns         *corev1.Namespace
		podLabels  map[string]string
		wantViewer string
		wantErr    bool
	}{
		{
			name:       "ns_owner_label",
			ns:         nsWithLabels("ollamav3-shared", map[string]string{"bytetrade.io/ns-owner": "brucedai"}),
			wantViewer: "brucedai",
		},
		{
			name:       "user_space_prefix",
			ns:         nsWithLabels("user-space-alice", nil),
			wantViewer: "alice",
		},
		{
			name:       "pod_owner_label_fallback",
			ns:         nsWithLabels("litellm-team", nil),
			podLabels:  map[string]string{constants.ApplicationOwnerLabel: "Bob"},
			wantViewer: "bob",
		},
		{
			name:    "all_absent",
			ns:      nsWithLabels("kube-system", nil),
			wantErr: true,
		},
		{
			name:    "app_user_namespace_without_owner",
			ns:      nsWithLabels("litellm-brucedai", nil),
			wantErr: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			wh := &Webhook{kubeClient: fake.NewSimpleClientset(c.ns)}
			pod := podWithLabels(c.ns.Name, c.podLabels)
			viewer, err := wh.deriveViewerFromPodNS(context.Background(), pod)
			if c.wantErr {
				require.Error(t, err)
				require.True(t, errors.Is(err, ErrD2ViewerUnderive))
				return
			}
			require.NoError(t, err)
			require.Equal(t, c.wantViewer, viewer)
		})
	}
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

// TC-A2-5: after method-ization the all-absent branch still returns an error
// matched by errors.Is(err, ErrD2ViewerUnderive), so WI-AIA-1 keeps classifying
// reason=viewer_underive instead of degrading to other.
func TestDeriveViewerFromPodNS_UnderiveSentinel(t *testing.T) {
	wh := &Webhook{kubeClient: fake.NewSimpleClientset(nsWithLabels("kube-system", nil))}
	_, err := wh.deriveViewerFromPodNS(context.Background(), podWithLabels("kube-system", nil))
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrD2ViewerUnderive))

	wh = &Webhook{kubeClient: fake.NewSimpleClientset(nsWithLabels("user-space-", nil))}
	_, err = wh.deriveViewerFromPodNS(context.Background(), podWithLabels("user-space-", nil))
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

	wh := &Webhook{kubeClient: fake.NewSimpleClientset(nsWithLabels("kube-system", nil))}
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
