package webhook

import (
	"context"
	"fmt"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

const (
	testCallerNS      = "litellm-usera"
	testReplicaSecret = constants.D2SharedTLSSecretNamePrefix + "usera"
	testPlainSecret   = "app-config"
)

// replicaSecret returns a tls-replica=true Secret in the caller NS.
func replicaSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testReplicaSecret,
			Namespace: testCallerNS,
			Labels:    map[string]string{labelTLSReplica: "true"},
		},
	}
}

// plainSecret returns a non-replica Secret (no tls-replica label).
func plainSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testPlainSecret,
			Namespace: testCallerNS,
		},
	}
}

func rawSecretVolume(volName, secretName string) corev1.Volume {
	return corev1.Volume{
		Name: volName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{SecretName: secretName},
		},
	}
}

func projectedSecretVolume(volName, secretName string) corev1.Volume {
	return corev1.Volume{
		Name: volName,
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				Sources: []corev1.VolumeProjection{
					{Secret: &corev1.SecretProjection{
						LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
					}},
				},
			},
		},
	}
}

func container(name string, mountVolumes ...string) corev1.Container {
	c := corev1.Container{Name: name}
	for _, v := range mountVolumes {
		c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{Name: v})
	}
	return c
}

func deniedCount(reason string) float64 {
	return testutil.ToFloat64(tlsReplicaMountDeniedTotal.WithLabelValues(hashCallerNamespace(testCallerNS), reason))
}

func validatedCount(result string) float64 {
	return testutil.ToFloat64(tlsReplicaMountValidatedTotal.WithLabelValues(result))
}

// TC-T1-8-01: legitimate d2 pod (olares-d2-sidecar mounts the tls-replica secret
// via the olares-d2-certs volume) is allowed; allow_d2 counter increments.
func TestValidateTLSReplicaMount_LegitD2Allow(t *testing.T) {
	before := validatedCount(tlsReplicaAllowResultD2)
	wh := &Webhook{kubeClient: fake.NewSimpleClientset(replicaSecret())}
	pod := &corev1.Pod{Spec: corev1.PodSpec{
		Containers: []corev1.Container{container(constants.D2SidecarContainerName, constants.D2CertsVolumeName)},
		Volumes:    []corev1.Volume{rawSecretVolume(constants.D2CertsVolumeName, testReplicaSecret)},
	}}

	allowed, code := wh.ValidateTLSReplicaMount(context.Background(), pod, testCallerNS)
	require.True(t, allowed)
	require.Empty(t, code)
	require.Equal(t, before+1, validatedCount(tlsReplicaAllowResultD2))
}

// TC-T1-8-02: pod with no tls-replica volume takes the allow fast path;
// allow_no_replica counter increments.
func TestValidateTLSReplicaMount_NoReplicaFastPath(t *testing.T) {
	before := validatedCount(tlsReplicaAllowResultNoReplica)
	wh := &Webhook{kubeClient: fake.NewSimpleClientset(plainSecret())}
	pod := &corev1.Pod{Spec: corev1.PodSpec{
		Containers: []corev1.Container{container("app", "config")},
		Volumes:    []corev1.Volume{rawSecretVolume("config", testPlainSecret)},
	}}

	allowed, code := wh.ValidateTLSReplicaMount(context.Background(), pod, testCallerNS)
	require.True(t, allowed)
	require.Empty(t, code)
	require.Equal(t, before+1, validatedCount(tlsReplicaAllowResultNoReplica))
}

// TC-T1-8-03: tenant bypass pod (non-d2 container raw-mounts the tls-replica
// secret) is denied with reason=non_d2_container.
func TestValidateTLSReplicaMount_NonD2ContainerDeny(t *testing.T) {
	before := deniedCount(tlsReplicaDenyReasonNonD2Container)
	wh := &Webhook{kubeClient: fake.NewSimpleClientset(replicaSecret())}
	pod := &corev1.Pod{Spec: corev1.PodSpec{
		Containers: []corev1.Container{container("rogue", "stolen")},
		Volumes:    []corev1.Volume{rawSecretVolume("stolen", testReplicaSecret)},
	}}

	allowed, code := wh.ValidateTLSReplicaMount(context.Background(), pod, testCallerNS)
	require.False(t, allowed)
	require.Equal(t, codeTLSReplicaMountDenied, code)
	require.Equal(t, before+1, deniedCount(tlsReplicaDenyReasonNonD2Container))
}

// TC-T1-8-04 (decision A): the tls-replica secret mounted by the d2 sidecar but
// under a non-standard volume name (not olares-d2-certs) is denied. Under
// decision A the discriminator is volume-name + mounter (source type agnostic),
// diverging from frozen 详设 §2.2 which keyed on "projected vs raw secret".
func TestValidateTLSReplicaMount_D2ButWrongVolumeNameDeny(t *testing.T) {
	before := deniedCount(tlsReplicaDenyReasonNonD2Container)
	wh := &Webhook{kubeClient: fake.NewSimpleClientset(replicaSecret())}
	pod := &corev1.Pod{Spec: corev1.PodSpec{
		Containers: []corev1.Container{container(constants.D2SidecarContainerName, "rogue-certs")},
		Volumes:    []corev1.Volume{rawSecretVolume("rogue-certs", testReplicaSecret)},
	}}

	allowed, code := wh.ValidateTLSReplicaMount(context.Background(), pod, testCallerNS)
	require.False(t, allowed)
	require.Equal(t, codeTLSReplicaMountDenied, code)
	require.Equal(t, before+1, deniedCount(tlsReplicaDenyReasonNonD2Container))
}

// TC-T1-8-05: a Secret label lookup API error fails closed with
// reason=label_lookup_failed.
func TestValidateTLSReplicaMount_LabelLookupFailedDeny(t *testing.T) {
	before := deniedCount(tlsReplicaDenyReasonLabelLookupFailed)
	client := fake.NewSimpleClientset()
	client.PrependReactor("get", "secrets", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("apiserver unavailable")
	})
	wh := &Webhook{kubeClient: client}
	pod := &corev1.Pod{Spec: corev1.PodSpec{
		Containers: []corev1.Container{container("app", "vol")},
		Volumes:    []corev1.Volume{rawSecretVolume("vol", "some-secret")},
	}}

	allowed, code := wh.ValidateTLSReplicaMount(context.Background(), pod, testCallerNS)
	require.False(t, allowed)
	require.Equal(t, codeTLSReplicaLabelLookupFail, code)
	require.Equal(t, before+1, deniedCount(tlsReplicaDenyReasonLabelLookupFailed))
}

// TC-T1-8-06: the deny reason set is exhaustive — each frozen reason records its
// own series and never collapses into an "other" bucket.
func TestRecordTLSReplicaMountDenied_ReasonExhaustive(t *testing.T) {
	reasons := []string{
		tlsReplicaDenyReasonNonD2Container,
		tlsReplicaDenyReasonWebhookUnavail,
		tlsReplicaDenyReasonLabelLookupFailed,
	}
	for _, r := range reasons {
		before := deniedCount(r)
		recordTLSReplicaMountDenied(testCallerNS, r)
		require.Equal(t, before+1, deniedCount(r), "reason %s must increment its own series", r)
	}
	// An unknown reason must not create an "other" series; it is folded into the
	// non_d2_container bucket (conservative, fail-closed default).
	before := deniedCount(tlsReplicaDenyReasonNonD2Container)
	recordTLSReplicaMountDenied(testCallerNS, "bogus_reason")
	require.Equal(t, before+1, deniedCount(tlsReplicaDenyReasonNonD2Container))
}

// TC-T1-8-07: multi-container pod where only the d2 sidecar mounts the
// tls-replica volume (other containers mount unrelated volumes) is allowed.
func TestValidateTLSReplicaMount_MultiContainerOnlyD2Allow(t *testing.T) {
	before := validatedCount(tlsReplicaAllowResultD2)
	wh := &Webhook{kubeClient: fake.NewSimpleClientset(replicaSecret(), plainSecret())}
	pod := &corev1.Pod{Spec: corev1.PodSpec{
		Containers: []corev1.Container{
			container("app", "config"),
			container(constants.D2SidecarContainerName, constants.D2CertsVolumeName),
		},
		Volumes: []corev1.Volume{
			rawSecretVolume("config", testPlainSecret),
			rawSecretVolume(constants.D2CertsVolumeName, testReplicaSecret),
		},
	}}

	allowed, code := wh.ValidateTLSReplicaMount(context.Background(), pod, testCallerNS)
	require.True(t, allowed)
	require.Empty(t, code)
	require.Equal(t, before+1, validatedCount(tlsReplicaAllowResultD2))
}

// TC-T1-8-08: a projected volume whose source references the tls-replica secret,
// mounted by a non-d2 container, is denied (projected sources are expanded).
func TestValidateTLSReplicaMount_ProjectedSourceNonD2Deny(t *testing.T) {
	before := deniedCount(tlsReplicaDenyReasonNonD2Container)
	wh := &Webhook{kubeClient: fake.NewSimpleClientset(replicaSecret())}
	pod := &corev1.Pod{Spec: corev1.PodSpec{
		Containers: []corev1.Container{container("rogue", "proj")},
		Volumes:    []corev1.Volume{projectedSecretVolume("proj", testReplicaSecret)},
	}}

	allowed, code := wh.ValidateTLSReplicaMount(context.Background(), pod, testCallerNS)
	require.False(t, allowed)
	require.Equal(t, codeTLSReplicaMountDenied, code)
	require.Equal(t, before+1, deniedCount(tlsReplicaDenyReasonNonD2Container))
}

// initContainer mounting the tls-replica secret in a non-d2 container is denied
// (init containers are part of the mounters set).
func TestValidateTLSReplicaMount_InitContainerNonD2Deny(t *testing.T) {
	before := deniedCount(tlsReplicaDenyReasonNonD2Container)
	wh := &Webhook{kubeClient: fake.NewSimpleClientset(replicaSecret())}
	pod := &corev1.Pod{Spec: corev1.PodSpec{
		InitContainers: []corev1.Container{container("init-thief", constants.D2CertsVolumeName)},
		Containers:     []corev1.Container{container(constants.D2SidecarContainerName, constants.D2CertsVolumeName)},
		Volumes:        []corev1.Volume{rawSecretVolume(constants.D2CertsVolumeName, testReplicaSecret)},
	}}

	allowed, code := wh.ValidateTLSReplicaMount(context.Background(), pod, testCallerNS)
	require.False(t, allowed)
	require.Equal(t, codeTLSReplicaMountDenied, code)
	require.Equal(t, before+1, deniedCount(tlsReplicaDenyReasonNonD2Container))
}
