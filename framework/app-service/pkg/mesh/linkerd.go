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
	// LinkerdNamespace is the RP Linkerd foundation install namespace (os-mesh),
	// not the upstream default "linkerd".
	LinkerdNamespace            = "os-mesh"
	linkerdNamespace            = LinkerdNamespace
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
// linkerd-pki-guardian is part of the readiness set: it keeps the identity issuer valid, so
// treating an unreachable guardian as ready would let Shared callers drop olares-envoy-sidecar
// before mesh auth can be trusted. Every lookup failure therefore fails closed.
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
	if err != nil {
		if !apierrors.IsNotFound(err) {
			klog.V(2).Infof("mesh: get deployment %s/%s failed: %v", linkerdNamespace, linkerdPKIGuardianDeploy, err)
		}
		return false
	}
	if guardian.Status.ReadyReplicas < 1 {
		return false
	}
	return true
}

// IsControlPlaneReady is the preferred name for mesh control-plane readiness checks.
func IsControlPlaneReady(ctx context.Context, kube kubernetes.Interface) bool {
	return IsLinkerdLayer1Ready(ctx, kube)
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

// ShouldSkipEnvoySidecar retires outbound/whole oes only when SteadyStateGate is Ready.
// Until the gate flips, admission must keep oes (fail-closed).
func ShouldSkipEnvoySidecar(ctx context.Context, kube kubernetes.Interface) bool {
	return steadyGateReadyOrOverride(ctx, kube)
}

// IsL4EdgePEPReady reports whether Track-A M-L4 is satisfied for oes skip gates.
// Order: env override (tests) → live l4 Deployment Ready (direct stop-inject for
// new installs / post-Track-A upgrade) → SteadyGate condition L4ProxyReady.
// Live probe avoids Accept chicken-egg (ZeroOesInventory required Accept before
// writing the condition that webhook used to wait on). Does not consult EG
// SecurityPolicy (EDGE path uses l4 as north-south PEP).
func IsL4EdgePEPReady(ctx context.Context, kube kubernetes.Interface) bool {
	if os.Getenv("OLARES_L4_EDGE_PEP_READY") == "1" {
		return true
	}
	if IsL4ProxyDeploymentReady(ctx, kube) {
		return true
	}
	st, err := LoadSteadyGate(ctx, kube)
	if err != nil {
		klog.V(2).Infof("mesh: IsL4EdgePEPReady load gate failed: %v", err)
		return false
	}
	if st == nil || st.Conditions == nil {
		return false
	}
	return st.Conditions[ConditionL4ProxyReady]
}

// ShouldSkipInboundEntranceSidecar skips inbound oes when M-L4 is true (l4 Ready).
// EDGE direct stop does not wait on Linkerd (E/W is mesh-in/Linkerd on Shared path).
func ShouldSkipInboundEntranceSidecar(ctx context.Context, kube kubernetes.Interface, appNamespace, srrName string) bool {
	_ = appNamespace
	_ = srrName
	return IsL4EdgePEPReady(ctx, kube)
}

// EvaluateSkipOes is the EDGE direct stop-inject gate:
// EdgePEPReady ∧ (¬HasProvider ∨ EgressAgentReady).
// linkerdReady is ignored (kept for call-site compatibility); Shared-caller
// E/W still uses EvaluateSkipOesForSharedCaller which does require Linkerd.
// edgePEPReady is M-L4 (l4 verify/302), not EG Entrance ExtAuth.
func EvaluateSkipOes(linkerdReady, edgePEPReady, hasProvider, egressAgentReady bool) bool {
	_ = linkerdReady
	if !edgePEPReady {
		return false
	}
	if hasProvider && !egressAgentReady {
		return false
	}
	return true
}

// ShouldSkipOes probes M-L4 and provider/egress readiness (no Linkerd gate).
func ShouldSkipOes(ctx context.Context, kube kubernetes.Interface, appNamespace, entranceSRRName string, hasProvider, egressAgentReady bool) bool {
	_ = appNamespace
	_ = entranceSRRName
	return EvaluateSkipOes(
		true,
		IsL4EdgePEPReady(ctx, kube),
		hasProvider,
		egressAgentReady,
	)
}

// EvaluateSkipOesForSharedCaller is the R1 Shared-caller gate:
// injectMeshIn ∧ LinkerdReady ∧ (¬HasProvider ∨ MeshOutReady).
// Unlike EvaluateSkipOes (L2-c), this does not require entrance extAuth and applies only
// when mesh-in will be injected.
func EvaluateSkipOesForSharedCaller(injectMeshIn, linkerdReady, hasProvider, injectMeshOut bool) bool {
	if !injectMeshIn || !linkerdReady {
		return false
	}
	if hasProvider && !injectMeshOut {
		return false
	}
	return true
}

// ShouldSkipOesForSharedCaller probes Linkerd readiness for the Shared-caller skip gate.
func ShouldSkipOesForSharedCaller(ctx context.Context, kube kubernetes.Interface, injectMeshIn, hasProvider, injectMeshOut bool) bool {
	return EvaluateSkipOesForSharedCaller(
		injectMeshIn,
		IsLinkerdLayer1Ready(ctx, kube),
		hasProvider,
		injectMeshOut,
	)
}
