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

// labelTLSReplica mirrors routecontrol's per-viewer TLS replica Secret label
// key. The owning constant lives in pkg/gateway/routecontrol (unexported, in a
// guarded file), so the value is restated here as the single string source for
// the mount guard; the two must stay byte-equal.
const labelTLSReplica = "gateway.olares.io/tls-replica"

// Admission deny error codes surfaced to the user (status message) and runbook.
const (
	codeTLSReplicaMountDenied     = "D2_TLS_REPLICA_MOUNT_DENIED"
	codeTLSReplicaLabelLookupFail = "D2_TLS_REPLICA_LABEL_LOOKUP_FAILED"
	codeTLSReplicaWebhookUnavail  = "D2_TLS_REPLICA_WEBHOOK_UNAVAILABLE"
)

// tls-replica mount-guard metric reason labels (denied counter). The set is
// exhaustive; webhook_unavailable is emitted by kube-apiserver via
// failurePolicy=Fail (handler never runs), declared here for completeness.
const (
	tlsReplicaDenyReasonNonD2Container    = "non_d2_container"
	tlsReplicaDenyReasonWebhookUnavail    = "webhook_unavailable"
	tlsReplicaDenyReasonLabelLookupFailed = "label_lookup_failed"
)

// tls-replica mount-guard allow result labels (validated counter).
const (
	tlsReplicaAllowResultD2        = "allow_d2"
	tlsReplicaAllowResultNoReplica = "allow_no_replica"
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
	prometheus.MustRegister(tlsReplicaMountDeniedTotal)
	prometheus.MustRegister(tlsReplicaMountValidatedTotal)
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

// recordTLSReplicaMountValidated increments the allow counter.
func recordTLSReplicaMountValidated(result string) {
	tlsReplicaMountValidatedTotal.WithLabelValues(result).Inc()
}

// hashCallerNamespace returns a non-PII, low-cardinality digest of a caller
// namespace (which embeds <app>-<user>) for use as a metric label value.
func hashCallerNamespace(ns string) string {
	sum := sha256.Sum256([]byte(ns))
	return hex.EncodeToString(sum[:8])
}

// ValidateTLSReplicaMount enforces that any volume referencing a tls-replica
// private-key Secret is mounted only by the platform-injected d2 sidecar.
//
// requirement: WI-T1-8 §2.2 — a tls-replica=true Secret may be consumed only by
// the olares-d2-sidecar container via the olares-d2-certs volume; any reference
// by another container (raw secret or projected source) is a cross-tenant
// private-key bypass and is denied.
//
// behavior: fail-closed — a Secret label lookup API error denies admission
// (private-key red line over availability). Pods with no tls-replica volume take
// an allow fast path. Returns (allowed, errorCode); errorCode is empty on allow.
func (wh *Webhook) ValidateTLSReplicaMount(ctx context.Context, pod *corev1.Pod, namespace string) (bool, string) {
	hasReplicaVolume := false

	for _, vol := range pod.Spec.Volumes {
		secretNames := referencedSecretNames(vol)
		if len(secretNames) == 0 {
			continue
		}
		isReplica := false
		for _, secretName := range secretNames {
			ok, err := wh.secretIsTLSReplica(ctx, namespace, secretName)
			if err != nil {
				recordTLSReplicaMountDenied(namespace, tlsReplicaDenyReasonLabelLookupFailed)
				return false, codeTLSReplicaLabelLookupFail
			}
			if ok {
				isReplica = true
			}
		}
		if !isReplica {
			continue
		}
		hasReplicaVolume = true

		if !isLegitD2ReplicaVolume(pod, vol.Name) {
			recordTLSReplicaMountDenied(namespace, tlsReplicaDenyReasonNonD2Container)
			return false, codeTLSReplicaMountDenied
		}
	}

	if hasReplicaVolume {
		recordTLSReplicaMountValidated(tlsReplicaAllowResultD2)
	} else {
		recordTLSReplicaMountValidated(tlsReplicaAllowResultNoReplica)
	}
	return true, ""
}

// referencedSecretNames collects every Secret name a volume references, covering
// both raw Secret volumes and projected sources.
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

// secretIsTLSReplica reports whether the named Secret carries tls-replica=true.
// A missing Secret is not a replica (skip); any other API error is returned so
// the caller can fail closed.
func (wh *Webhook) secretIsTLSReplica(ctx context.Context, namespace, name string) (bool, error) {
	secret, err := wh.kubeClient.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return secret.Labels[labelTLSReplica] == "true", nil
}

// isLegitD2ReplicaVolume reports whether a tls-replica-bearing volume is the
// platform-injected d2 cert volume mounted exclusively by the d2 sidecar
// (decision A: source type agnostic; discriminator is volume name + mounters).
func isLegitD2ReplicaVolume(pod *corev1.Pod, volumeName string) bool {
	if volumeName != constants.D2CertsVolumeName {
		return false
	}
	mounted := false
	for _, c := range append(append([]corev1.Container{}, pod.Spec.InitContainers...), pod.Spec.Containers...) {
		for _, vm := range c.VolumeMounts {
			if vm.Name != volumeName {
				continue
			}
			if c.Name != constants.D2SidecarContainerName {
				return false
			}
			mounted = true
		}
	}
	return mounted
}
