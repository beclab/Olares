package callerjwt

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
)

// reconcileJWKSIngressNP ensures a NetworkPolicy so Envoy Gateway pods in
// os-gateway can reach the JWKS listen port on app-service pods.
func (r *IssuerReconciler) reconcileJWKSIngressNP(ctx context.Context) error {
	if r == nil || r.Client == nil {
		return nil
	}
	_, targetPort := JWKSListenPort()
	desired := desiredJWKSIngressNP(targetPort)

	current := &networkingv1.NetworkPolicy{}
	key := types.NamespacedName{Name: JWKSIngressNPName, Namespace: JWKSServiceNamespace}
	err := r.Client.Get(ctx, key, current)
	switch {
	case apierrors.IsNotFound(err):
		if err := r.Client.Create(ctx, desired); err != nil {
			klog.Errorf("callerjwt: create JWKS ingress NetworkPolicy %s/%s failed: %v",
				JWKSServiceNamespace, JWKSIngressNPName, err)
			return fmt.Errorf("create JWKS ingress NetworkPolicy: %w", err)
		}
		return nil
	case err != nil:
		klog.Errorf("callerjwt: get JWKS ingress NetworkPolicy %s/%s failed: %v",
			JWKSServiceNamespace, JWKSIngressNPName, err)
		return fmt.Errorf("get JWKS ingress NetworkPolicy: %w", err)
	default:
		current.Spec = desired.Spec
		if current.Labels == nil {
			current.Labels = map[string]string{}
		}
		for k, v := range desired.Labels {
			current.Labels[k] = v
		}
		if err := r.Client.Update(ctx, current); err != nil {
			klog.Errorf("callerjwt: update JWKS ingress NetworkPolicy %s/%s failed: %v",
				JWKSServiceNamespace, JWKSIngressNPName, err)
			return fmt.Errorf("update JWKS ingress NetworkPolicy: %w", err)
		}
		return nil
	}
}

func desiredJWKSIngressNP(targetPort int32) *networkingv1.NetworkPolicy {
	proto := corev1.ProtocolTCP
	port := intstr.FromInt(int(targetPort))
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      JWKSIngressNPName,
			Namespace: JWKSServiceNamespace,
			Labels: map[string]string{
				managedByLabel:          managedByValue,
				managedByComponentLabel: JWKSIngressNPComponentValue,
			},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					jwksAppServiceSelectorKey: jwksAppServiceSelectorValue,
				},
			},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									corev1.LabelMetadataName: JWKSIngressNPFromNamespace,
								},
							},
						},
					},
					Ports: []networkingv1.NetworkPolicyPort{
						{
							Protocol: &proto,
							Port:     &port,
						},
					},
				},
			},
		},
	}
}
