package webhook

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"

	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// labelTLSReplica mirrors constants.LabelTLSReplica (viewer platform TLS replica).
const labelTLSReplica = constants.LabelTLSReplica

// labelTLSCustomReplica mirrors constants.LabelTLSCustomReplica (third-party
// aggregate Secret). Must stay byte-equal with routecontrol writers.
const labelTLSCustomReplica = constants.LabelTLSCustomReplica

// Admission deny error codes surfaced to the user (status message) and runbook.
const (
	codeTLSReplicaMountDenied     = "MESH_IN_TLS_REPLICA_MOUNT_DENIED"
	codeTLSReplicaLabelLookupFail = "MESH_IN_TLS_REPLICA_LABEL_LOOKUP_FAILED"
	codeTLSReplicaWebhookUnavail  = "MESH_IN_TLS_REPLICA_WEBHOOK_UNAVAILABLE"
)

// tls-replica mount-guard metric reason labels (denied counter). The set is
// exhaustive; webhook_unavailable is emitted by kube-apiserver via
// failurePolicy=Fail (handler never runs), declared here for completeness.
const (
	tlsReplicaDenyReasonNonD2Container    = "non_mesh_in_container"
	tlsReplicaDenyReasonWebhookUnavail    = "webhook_unavailable"
	tlsReplicaDenyReasonLabelLookupFailed = "label_lookup_failed"
)

// tls-replica mount-guard allow result labels (validated counter).
const (
	tlsReplicaAllowResultD2        = "allow_mesh_in"
	tlsReplicaAllowResultNoReplica = "allow_no_replica"
)

type tlsProtectedKind int

const (
	tlsProtectedNone tlsProtectedKind = iota
	tlsProtectedViewer
	tlsProtectedCustom
)

var tlsReplicaMountDeniedTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "app_service_d2_tls_replica_mount_denied_total",
		Help: "cross-tenant tls-replica private-key mount admission denials by reason",
	},
	[]string{"ns", "reason"},
)

var tlsReplicaMountValidatedTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "app_service_d2_tls_replica_mount_validated_total",
		Help: "tls-replica mount admission allows (legit d2 mount / no replica fast path)",
	},
	[]string{"result"},
)

func init() {
	prometheus.MustRegister(tlsReplicaMountDeniedTotal, tlsReplicaMountValidatedTotal)
}

// recordTLSReplicaMountDenied increments the deny counter. The reason switch is
// exhaustive over the three frozen reasons; an unknown reason is rejected rather
// than collapsed into an "other" bucket (详设 §6.1 exhaustive contract).
func recordTLSReplicaMountDenied(ns, reason string) {
	switch reason {
	case tlsReplicaDenyReasonNonD2Container,
		tlsReplicaDenyReasonWebhookUnavail,
		tlsReplicaDenyReasonLabelLookupFailed:
	default:
		reason = tlsReplicaDenyReasonNonD2Container
	}
	tlsReplicaMountDeniedTotal.WithLabelValues(hashCallerNamespace(ns), reason).Inc()
}

func recordTLSReplicaMountValidated(result string) {
	tlsReplicaMountValidatedTotal.WithLabelValues(result).Inc()
}

func hashCallerNamespace(ns string) string {
	sum := sha256.Sum256([]byte(ns))
	return hex.EncodeToString(sum[:8])
}

// ValidateTLSReplicaMount enforces that private-key TLS Secrets projected for
// mesh-in are mounted only by the platform-injected mesh-in agent:
//   - tls-replica            → volume olares-mesh-in-certs
//   - tls-custom-replica     → volume olares-mesh-in-custom-certs
//
// behavior: fail-closed on Secret label lookup errors. Pods with no protected
// TLS volume take an allow fast path. Returns (allowed, errorCode).
func (wh *Webhook) ValidateTLSReplicaMount(ctx context.Context, pod *corev1.Pod, namespace string) (bool, string) {
	hasProtectedVolume := false

	for _, vol := range pod.Spec.Volumes {
		secretNames := referencedSecretNames(vol)
		if len(secretNames) == 0 {
			continue
		}
		kind := tlsProtectedNone
		for _, secretName := range secretNames {
			k, err := wh.secretTLSProtectedKind(ctx, namespace, secretName)
			if err != nil {
				recordTLSReplicaMountDenied(namespace, tlsReplicaDenyReasonLabelLookupFailed)
				return false, codeTLSReplicaLabelLookupFail
			}
			if k != tlsProtectedNone {
				kind = k
				break
			}
		}
		if kind == tlsProtectedNone {
			continue
		}
		hasProtectedVolume = true

		expectedVol := constants.MeshInCertsVolumeName
		if kind == tlsProtectedCustom {
			expectedVol = constants.MeshInCustomCertsVolumeName
		}
		if !isLegitMeshInProtectedVolume(pod, vol.Name, expectedVol) {
			recordTLSReplicaMountDenied(namespace, tlsReplicaDenyReasonNonD2Container)
			return false, codeTLSReplicaMountDenied
		}
	}

	if hasProtectedVolume {
		recordTLSReplicaMountValidated(tlsReplicaAllowResultD2)
	} else {
		recordTLSReplicaMountValidated(tlsReplicaAllowResultNoReplica)
	}
	return true, ""
}

func referencedSecretNames(vol corev1.Volume) []string {
	var names []string
	if vol.Secret != nil && vol.Secret.SecretName != "" {
		names = append(names, vol.Secret.SecretName)
	}
	if vol.Projected != nil {
		for _, src := range vol.Projected.Sources {
			if src.Secret != nil && src.Secret.Name != "" {
				names = append(names, src.Secret.Name)
			}
		}
	}
	return names
}

// secretTLSProtectedKind classifies Secrets that carry mesh-in private-key
// material. Missing Secret → none (optional mounts at Pod start).
func (wh *Webhook) secretTLSProtectedKind(ctx context.Context, namespace, name string) (tlsProtectedKind, error) {
	secret, err := wh.kubeClient.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return tlsProtectedNone, nil
		}
		return tlsProtectedNone, err
	}
	if secret.Labels[labelTLSCustomReplica] == "true" {
		return tlsProtectedCustom, nil
	}
	if secret.Labels[labelTLSReplica] == "true" {
		return tlsProtectedViewer, nil
	}
	return tlsProtectedNone, nil
}

// isLegitMeshInProtectedVolume requires the volume name to match the platform
// cert volume for that Secret kind and that only mesh-in-agent mounts it.
func isLegitMeshInProtectedVolume(pod *corev1.Pod, volumeName, expectedVolumeName string) bool {
	if volumeName != expectedVolumeName {
		return false
	}
	mounted := false
	for _, c := range append(append([]corev1.Container{}, pod.Spec.InitContainers...), pod.Spec.Containers...) {
		for _, vm := range c.VolumeMounts {
			if vm.Name != volumeName {
				continue
			}
			if c.Name != constants.MeshInAgentContainerName {
				return false
			}
			mounted = true
		}
	}
	return mounted
}

// isLegitMeshInReplicaVolume is the viewer-cert helper kept for existing tests.
func isLegitMeshInReplicaVolume(pod *corev1.Pod, volumeName string) bool {
	return isLegitMeshInProtectedVolume(pod, volumeName, constants.MeshInCertsVolumeName)
}
