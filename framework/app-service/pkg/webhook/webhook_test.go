package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
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
// CreateD2OffloaderPatch. The image is configured here so the image_unconfigured
// guard passes and derivation is what fails (no further client read needed).
func TestCreateD2OffloaderPatch_ViewerUnderiveSentinel(t *testing.T) {
	orig := d2SidecarImageDigest
	d2SidecarImageDigest = func() string { return "beclab/nginx@sha256:configured" }
	t.Cleanup(func() { d2SidecarImageDigest = orig })

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

// TC-T1-3-IDM: CreateD2OffloaderCallerPatch is idempotent -- a pod that already
// carries the d2 sidecar short-circuits to an empty patch. Bypass (a) runs after
// CreatePatch without its short-circuit, so this entry guard (M8) is mandatory.
func TestCreateD2OffloaderCallerPatch_IdempotentWhenD2ContainerExists(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1ObjectMeta("litellm-alice", "demo"),
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "app"},
				{Name: constants.D2SidecarContainerName},
			},
		},
	}
	raw, err := json.Marshal(pod)
	require.NoError(t, err)
	req := &admissionv1.AdmissionRequest{Object: runtime.RawExtension{Raw: raw}}

	wh := &Webhook{}
	patch, err := wh.CreateD2OffloaderCallerPatch(context.Background(), &pod, req, nil, uuid.New())
	require.NoError(t, err)

	var ops []map[string]any
	require.NoError(t, json.Unmarshal(patch, &ops))
	require.Len(t, ops, 0)
}

// TC-T1-3-R5: an unconfigured d2 image digest (WI-N1-IMG pending) fails open
// with ErrD2ImageUnconfigured before any API read, so the pod is admitted and
// no placeholder d2 (which would ImagePullBackOff) is injected.
func TestCreateD2OffloaderCallerPatch_ImageUnconfigured(t *testing.T) {
	orig := d2SidecarImageDigest
	d2SidecarImageDigest = func() string { return constants.D2SidecarImagePlaceholder }
	t.Cleanup(func() { d2SidecarImageDigest = orig })

	pod := corev1.Pod{
		ObjectMeta: metav1ObjectMeta("litellm-alice", "demo"),
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}}},
	}
	raw, err := json.Marshal(pod)
	require.NoError(t, err)
	req := &admissionv1.AdmissionRequest{Object: runtime.RawExtension{Raw: raw}}

	wh := &Webhook{}
	_, err = wh.CreateD2OffloaderCallerPatch(context.Background(), &pod, req, nil, uuid.New())
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrD2ImageUnconfigured))
	require.Equal(t, d2SkipReasonImageUnconfigured, ClassifyD2SkipReason(err))
}

// WI-NAT-4: server-mode CreateD2OffloaderPatch is symmetric to caller-mode -- an
// unconfigured d2 image digest fails open with ErrD2ImageUnconfigured after the
// hasD2Container short-circuit but before any API read, so the v3 shared-entrance
// pod is admitted and no placeholder d2 (which would ImagePullBackOff) is injected.
func TestCreateD2OffloaderPatch_ImageUnconfigured(t *testing.T) {
	orig := d2SidecarImageDigest
	d2SidecarImageDigest = func() string { return constants.D2SidecarImagePlaceholder }
	t.Cleanup(func() { d2SidecarImageDigest = orig })

	pod := corev1.Pod{
		ObjectMeta: metav1ObjectMeta("user-space-alice", "demo"),
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}}},
	}
	raw, err := json.Marshal(pod)
	require.NoError(t, err)
	req := &admissionv1.AdmissionRequest{Object: runtime.RawExtension{Raw: raw}}

	wh := &Webhook{}
	_, err = wh.CreateD2OffloaderPatch(context.Background(), &pod, req, nil, uuid.New())
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrD2ImageUnconfigured))
	require.Equal(t, d2SkipReasonImageUnconfigured, ClassifyD2SkipReason(err))
}

func TestD2ImageUnconfigured(t *testing.T) {
	require.True(t, d2ImageUnconfigured(constants.D2SidecarImagePlaceholder))
	require.True(t, d2ImageUnconfigured(""))
	require.True(t, d2ImageUnconfigured("  "))
	require.False(t, d2ImageUnconfigured(constants.D2SidecarImageDigest))
}

// TC-T1-3-R2: clusterappref_empty is the caller-opt-in-but-empty-ref case,
// distinct from caller_viewer_unresolved (refs present, resolve to no owner).
// Any opted-in app with a non-empty ref makes the resolver proceed (false).
func TestCallerOptInRefsEmpty(t *testing.T) {
	cases := []struct {
		name string
		ns   string
		apps []v1alpha1.Application
		want bool
	}{
		{"opted_in_empty_ref", "litellm-userA", []v1alpha1.Application{callerApp("litellm", "litellm-userA", "")}, true},
		{"opted_in_with_ref", "litellm-userA", []v1alpha1.Application{callerApp("litellm", "litellm-userA", "ollama")}, false},
		{"mixed_one_ref_wins", "litellm-userA", []v1alpha1.Application{
			callerApp("a", "litellm-userA", ""),
			callerApp("b", "litellm-userA", "ollama"),
		}, false},
		{"no_opt_in_app", "litellm-userA", []v1alpha1.Application{clusterApp("ollama", "userA")}, false},
		{"empty_ns", "", []v1alpha1.Application{callerApp("a", "litellm-userA", "")}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.want, callerOptInRefsEmpty(c.ns, c.apps))
		})
	}
}

// TC-T1-3-R1..R5 / R4: ClassifyD2SkipReason exhaustively maps every caller-mode
// fail-open sentinel to its frozen reason label; no real sentinel degrades to
// "other" (only nil / uncategorized errors do). multi_ref_unsupported is a
// resolver warning (not an error) -- asserted in TC-T1-5-02.
func TestClassifyD2SkipReason_CallerSentinelsExhaustive(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{fmt.Errorf("ctx: %w", ErrD2SnapshotUnavailable), d2SkipReasonSnapshotError},
		{fmt.Errorf("ctx: %w", ErrD2ViewerUnderive), d2SkipReasonViewerUnderive},
		{fmt.Errorf("ctx: %w", ErrD2TLSSecretMissing), d2SkipReasonTLSSecretMissing},
		{fmt.Errorf("ctx: %w", ErrD2CallerViewerUnresolved), d2SkipReasonCallerViewerUnresolved},
		{fmt.Errorf("ctx: %w", ErrD2ClusterAppRefEmpty), d2SkipReasonClusterAppRefEmpty},
		{fmt.Errorf("ctx: %w", ErrD2ImageUnconfigured), d2SkipReasonImageUnconfigured},
	}
	for _, c := range cases {
		require.Equal(t, c.want, ClassifyD2SkipReason(c.err))
		require.NotEqual(t, d2SkipReasonOther, ClassifyD2SkipReason(c.err))
	}
}

func TestRecordD2InjectSucceeded(t *testing.T) {
	beforeA := testutil.ToFloat64(d2InjectSucceededTotal.WithLabelValues(D2InjectModeCaller, D2InjectScenarioA))
	RecordD2InjectSucceeded(D2InjectModeCaller, D2InjectScenarioA)
	require.Equal(t, beforeA+1, testutil.ToFloat64(d2InjectSucceededTotal.WithLabelValues(D2InjectModeCaller, D2InjectScenarioA)))

	beforeB := testutil.ToFloat64(d2InjectSucceededTotal.WithLabelValues(D2InjectModeCaller, D2InjectScenarioB))
	RecordD2InjectSucceeded(D2InjectModeCaller, D2InjectScenarioB)
	require.Equal(t, beforeB+1, testutil.ToFloat64(d2InjectSucceededTotal.WithLabelValues(D2InjectModeCaller, D2InjectScenarioB)))

	// WI-NAT-4: server-mode success is recorded under its own scenario label,
	// distinct from the caller-mode bypass A/B scenarios.
	beforeServer := testutil.ToFloat64(d2InjectSucceededTotal.WithLabelValues(D2InjectModeServer, D2InjectScenarioServerMain))
	RecordD2InjectSucceeded(D2InjectModeServer, D2InjectScenarioServerMain)
	require.Equal(t, beforeServer+1, testutil.ToFloat64(d2InjectSucceededTotal.WithLabelValues(D2InjectModeServer, D2InjectScenarioServerMain)))
}

// the default drain grace is added only when unset; an explicit
// caller value is respected (do not override business semantics).
func TestEnsureD2DrainGracePeriod(t *testing.T) {
	pod := &corev1.Pod{}
	ensureD2DrainGracePeriod(pod)
	require.NotNil(t, pod.Spec.TerminationGracePeriodSeconds)
	require.Equal(t, constants.D2DrainGracePeriodSeconds, *pod.Spec.TerminationGracePeriodSeconds)

	explicit := int64(10)
	pod2 := &corev1.Pod{Spec: corev1.PodSpec{TerminationGracePeriodSeconds: &explicit}}
	ensureD2DrainGracePeriod(pod2)
	require.Equal(t, int64(10), *pod2.Spec.TerminationGracePeriodSeconds)
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
