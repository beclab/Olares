package mesh

import (
	"context"

	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/security"
)

// EnsureAppGatewayMeshNetworkPolicies creates/updates the static mesh NPs that
// admit in-cluster-caller / shared / gateway peers to Linkerd CP (os-mesh) and
// admit os-mesh into os-gateway. Chart YAML alone is not enough on brownfield
// clusters where others-np deny-all would otherwise block identity issuance.
func EnsureAppGatewayMeshNetworkPolicies(ctx context.Context, c client.Client) error {
	if c == nil {
		return nil
	}
	for _, desired := range []*networkingv1.NetworkPolicy{
		security.NewAppGatewayMeshNPOsMesh(),
		security.NewAppGatewayMeshNPOsGateway(),
	} {
		if err := ensureNetworkPolicy(ctx, c, desired); err != nil {
			return err
		}
	}
	return nil
}

func ensureNetworkPolicy(ctx context.Context, c client.Client, desired *networkingv1.NetworkPolicy) error {
	if desired == nil {
		return nil
	}
	key := types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}
	current := &networkingv1.NetworkPolicy{}
	err := c.Get(ctx, key, current)
	switch {
	case apierrors.IsNotFound(err):
		if err := c.Create(ctx, desired); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				klog.Errorf("mesh-xport: create NP %s/%s failed: %v", desired.Namespace, desired.Name, err)
				return err
			}
			// Concurrent Ensure* (Application + SharedRoute) raced Create; refresh and Update.
			if err := c.Get(ctx, key, current); err != nil {
				klog.Errorf("mesh-xport: get NP %s/%s after AlreadyExists failed: %v", desired.Namespace, desired.Name, err)
				return err
			}
		} else {
			klog.Infof("mesh-xport: created NP %s/%s", desired.Namespace, desired.Name)
			return nil
		}
	case err != nil:
		klog.Errorf("mesh-xport: get NP %s/%s failed: %v", desired.Namespace, desired.Name, err)
		return err
	}
	current.Spec = desired.Spec
	if current.Labels == nil {
		current.Labels = map[string]string{}
	}
	for k, v := range desired.Labels {
		current.Labels[k] = v
	}
	if err := c.Update(ctx, current); err != nil {
		klog.Errorf("mesh-xport: update NP %s/%s failed: %v", desired.Namespace, desired.Name, err)
		return err
	}
	return nil
}
