package routecontrol

import (
	"context"
	"testing"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestEnsureInClusterCallerIngressNP_creates(t *testing.T) {
	s := testScheme(t)
	c := fake.NewClientBuilder().WithScheme(s).Build()

	if err := EnsureInClusterCallerIngressNP(context.Background(), c); err != nil {
		t.Fatalf("EnsureInClusterCallerIngressNP: %v", err)
	}

	np := &networkingv1.NetworkPolicy{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: defaultGatewayNS,
		Name:      InClusterCallerIngressNPName,
	}, np); err != nil {
		t.Fatalf("get NP: %v", err)
	}
	if np.Labels[ManagedByLabel] != ManagedByValue {
		t.Fatalf("managed-by = %q", np.Labels[ManagedByLabel])
	}
	if len(np.Spec.Ingress) != 1 || len(np.Spec.Ingress[0].From) != 1 {
		t.Fatalf("ingress rules = %+v", np.Spec.Ingress)
	}
	expr := np.Spec.Ingress[0].From[0].NamespaceSelector.MatchExpressions
	if len(expr) != 1 || expr[0].Key != NamespaceOwnerLabel || expr[0].Operator != metav1.LabelSelectorOpExists {
		t.Fatalf("namespace selector = %+v", expr)
	}
	if len(np.Spec.PolicyTypes) != 1 || np.Spec.PolicyTypes[0] != networkingv1.PolicyTypeIngress {
		t.Fatalf("policy types = %+v", np.Spec.PolicyTypes)
	}
}

func TestEnsureInClusterCallerIngressNP_idempotent(t *testing.T) {
	s := testScheme(t)
	c := fake.NewClientBuilder().WithScheme(s).Build()
	ctx := context.Background()

	if err := EnsureInClusterCallerIngressNP(ctx, c); err != nil {
		t.Fatalf("first ensure: %v", err)
	}
	if err := EnsureInClusterCallerIngressNP(ctx, c); err != nil {
		t.Fatalf("second ensure: %v", err)
	}

	var list networkingv1.NetworkPolicyList
	if err := c.List(ctx, &list); err != nil {
		t.Fatalf("list NP: %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("NP count = %d, want 1", len(list.Items))
	}
}

func TestGatewayInClusterIngressNPReconciler_reconcile(t *testing.T) {
	s := testScheme(t)
	gw := &unstructured.Unstructured{}
	gw.SetGroupVersionKind(gatewayGVK)
	gw.SetName(defaultGatewayName)
	gw.SetNamespace(defaultGatewayNS)
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(gw).Build()

	r := &GatewayInClusterIngressNPReconciler{Client: c}
	if _, err := r.Reconcile(context.Background(), gatewayReconcileRequest()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	np := &networkingv1.NetworkPolicy{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: defaultGatewayNS,
		Name:      InClusterCallerIngressNPName,
	}, np); err != nil {
		t.Fatalf("NP not created: %v", err)
	}
}

func TestGatewayInClusterIngressNPReconciler_gatewayMissing(t *testing.T) {
	s := testScheme(t)
	c := fake.NewClientBuilder().WithScheme(s).Build()
	r := &GatewayInClusterIngressNPReconciler{Client: c}
	if _, err := r.Reconcile(context.Background(), gatewayReconcileRequest()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	var list networkingv1.NetworkPolicyList
	if err := c.List(context.Background(), &list); err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list.Items) != 0 {
		t.Fatalf("NP count = %d, want 0", len(list.Items))
	}
}

func gatewayReconcileRequest() reconcile.Request {
	return reconcile.Request{NamespacedName: types.NamespacedName{
		Namespace: defaultGatewayNS,
		Name:      defaultGatewayName,
	}}
}
