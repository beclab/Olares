package routecontrol

import (
	"context"

	"github.com/beclab/Olares/framework/app-service/pkg/security"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	// InClusterCallerIngressNPName admits in-cluster user namespaces to the
	// app-gateway data plane. Lite line: always reconciled when the Gateway
	// exists; Full caller opt-in gating is a separate future WI.
	InClusterCallerIngressNPName = "app-gateway-incluster-caller-ingress-np"
	RouteControlComponentLabel   = "app.kubernetes.io/component"
	RouteControlComponentValue   = "route-control"
	NamespaceOwnerLabel          = "bytetrade.io/ns-owner"
)

// GatewayInClusterIngressNPReconciler ensures app-gateway-incluster-caller-
// ingress-np exists in the gateway namespace so user-space callers can reach
// app-gateway-data without patching others-np.
type GatewayInClusterIngressNPReconciler struct {
	Client client.Client
}

// Reconcile writes the in-cluster caller ingress NetworkPolicy when the parent
// Gateway is present.
func (r *GatewayInClusterIngressNPReconciler) Reconcile(ctx context.Context, _ reconcile.Request) (reconcile.Result, error) {
	if r == nil || r.Client == nil {
		return reconcile.Result{}, nil
	}
	gw := &unstructured.Unstructured{}
	gw.SetGroupVersionKind(gatewayGVK)
	if err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: defaultGatewayNS,
		Name:      defaultGatewayName,
	}, gw); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}
	return reconcile.Result{}, EnsureInClusterCallerIngressNP(ctx, r.Client)
}

// EnsureInClusterCallerIngressNP creates or updates the managed ingress NP.
func EnsureInClusterCallerIngressNP(ctx context.Context, c client.Client) error {
	desired := desiredInClusterCallerIngressNP()
	current := &networkingv1.NetworkPolicy{}
	key := types.NamespacedName{Namespace: defaultGatewayNS, Name: InClusterCallerIngressNPName}
	err := c.Get(ctx, key, current)
	switch {
	case apierrors.IsNotFound(err):
		return c.Create(ctx, desired)
	case err != nil:
		return err
	}
	current.Spec = desired.Spec
	if current.Labels == nil {
		current.Labels = map[string]string{}
	}
	current.Labels[ManagedByLabel] = ManagedByValue
	current.Labels[RouteControlComponentLabel] = RouteControlComponentValue
	return c.Update(ctx, current)
}

func desiredInClusterCallerIngressNP() *networkingv1.NetworkPolicy {
	existsOp := metav1.LabelSelectorOpExists
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      InClusterCallerIngressNPName,
			Namespace: defaultGatewayNS,
			Labels: map[string]string{
				ManagedByLabel:             ManagedByValue,
				RouteControlComponentLabel: RouteControlComponentValue,
			},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
			PodSelector: metav1.LabelSelector{},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      NamespaceOwnerLabel,
										Operator: existsOp,
									},
								},
							},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      "bytetrade.io/ns-type",
										Operator: metav1.LabelSelectorOpIn,
										Values:   []string{"system"},
									},
								},
							},
						},
					},
				},
				{
					From: security.NodeTunnelRule(),
				},
			},
		},
	}
}

// SetupWithManager watches the app-gateway Gateway in os-gateway.
func (r *GatewayInClusterIngressNPReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r == nil {
		return nil
	}
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	onlyAppGateway := predicate.NewPredicateFuncs(func(o client.Object) bool {
		return o.GetNamespace() == defaultGatewayNS && o.GetName() == defaultGatewayName
	})
	gw := &unstructured.Unstructured{}
	gw.SetGroupVersionKind(gatewayGVK)
	return ctrl.NewControllerManagedBy(mgr).
		Named("gateway-incluster-ingress-np").
		For(gw, builder.WithPredicates(onlyAppGateway)).
		Complete(r)
}
