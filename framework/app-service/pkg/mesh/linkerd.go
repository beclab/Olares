package mesh

import (
	"context"
	"os"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	linkerdNamespace            = "linkerd"
	linkerdPKIGuardianDeploy    = "linkerd-pki-guardian"
	EntranceExtAuthPolicySuffix = "-entrance-ext-auth"
)

var (
	linkerdControlPlaneDeployments = []string{
		"linkerd-destination",
		"linkerd-identity",
		"linkerd-proxy-injector",
	}
	securityPolicyGVR = schema.GroupVersionResource{
		Group: "gateway.envoyproxy.io", Version: "v1alpha1", Resource: "securitypolicies",
	}
)

func linkerdLayer1Enabled() bool {
	v := os.Getenv("OLARES_LINKERD_LAYER1_ENABLED")
	return v == "" || v == "1" || v == "true" || v == "TRUE"
}

// IsLinkerdLayer1Ready reports whether core Linkerd control plane deployments are Available.
func IsLinkerdLayer1Ready(ctx context.Context, kube kubernetes.Interface) bool {
	if !linkerdLayer1Enabled() || kube == nil {
		return false
	}
	for _, name := range linkerdControlPlaneDeployments {
		dep, err := kube.AppsV1().Deployments(linkerdNamespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return false
			}
			klog.V(2).Infof("mesh: get deployment %s/%s failed: %v", linkerdNamespace, name, err)
			return false
		}
		if dep.Status.ReadyReplicas < 1 {
			return false
		}
	}
	guardian, err := kube.AppsV1().Deployments(linkerdNamespace).Get(ctx, linkerdPKIGuardianDeploy, metav1.GetOptions{})
	if err == nil && guardian.Status.ReadyReplicas < 1 {
		return false
	}
	return true
}

// EntranceExtAuthPolicyName returns the entrance SecurityPolicy object name.
func EntranceExtAuthPolicyName(srrName string) string {
	return srrName + EntranceExtAuthPolicySuffix
}

// HasEntranceExtAuthPolicy reports whether WI-ORD-ENT-EG-1 extAuth exists for an entrance SRR.
func HasEntranceExtAuthPolicy(ctx context.Context, ns, srrName string) bool {
	if ns == "" || srrName == "" {
		return false
	}
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return false
	}
	dc, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return false
	}
	_, err = dc.Resource(securityPolicyGVR).Namespace(ns).Get(ctx, EntranceExtAuthPolicyName(srrName), metav1.GetOptions{})
	return err == nil
}

// ShouldSkipEnvoySidecar is the general webhook gate for retiring olares-envoy-sidecar.
func ShouldSkipEnvoySidecar(ctx context.Context, kube kubernetes.Interface) bool {
	return IsLinkerdLayer1Ready(ctx, kube)
}

// ShouldSkipInboundEntranceSidecar skips inbound oes when Linkerd is ready and extAuth exists.
func ShouldSkipInboundEntranceSidecar(ctx context.Context, kube kubernetes.Interface, appNamespace, srrName string) bool {
	if !IsLinkerdLayer1Ready(ctx, kube) {
		return false
	}
	return HasEntranceExtAuthPolicy(ctx, appNamespace, srrName)
}

// EvaluateSkipOes is the pure L2-c gate (REF §3.9.5):
// LinkerdReady ∧ L2aExtAuthReady ∧ (¬HasProvider ∨ EgressAgentReady).
func EvaluateSkipOes(linkerdReady, extAuthReady, hasProvider, egressAgentReady bool) bool {
	if !linkerdReady || !extAuthReady {
		return false
	}
	if hasProvider && !egressAgentReady {
		return false
	}
	return true
}

// ShouldSkipOes combines Linkerd/extAuth cluster probes with provider/egress readiness.
func ShouldSkipOes(ctx context.Context, kube kubernetes.Interface, appNamespace, entranceSRRName string, hasProvider, egressAgentReady bool) bool {
	return EvaluateSkipOes(
		IsLinkerdLayer1Ready(ctx, kube),
		HasEntranceExtAuthPolicy(ctx, appNamespace, entranceSRRName),
		hasProvider,
		egressAgentReady,
	)
}
