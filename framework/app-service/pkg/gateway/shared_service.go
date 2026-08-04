package gateway

import (
	"context"
	"fmt"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NamespaceSharedLabel marks namespaces created for v2 shared subcharts.
const NamespaceSharedLabel = "bytetrade.io/ns-shared"

// ResolveSharedEntranceService finds the Kubernetes Service backing a
// sharedEntrance host. It checks app.Spec.Namespace first, then any namespace
// labelled NamespaceSharedLabel (v2 multi-chart layout).
func ResolveSharedEntranceService(ctx context.Context, c client.Client, app *appv1alpha1.Application, serviceName string) (*corev1.Service, error) {
	if app == nil || app.Spec.Namespace == "" {
		return nil, fmt.Errorf("application or spec.namespace is empty")
	}
	if serviceName == "" {
		return nil, fmt.Errorf("shared entrance service name is empty")
	}

	svc := &corev1.Service{}
	if err := c.Get(ctx, client.ObjectKey{Namespace: app.Spec.Namespace, Name: serviceName}, svc); err == nil {
		return svc, nil
	} else if !apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("get backing service %s/%s: %w", app.Spec.Namespace, serviceName, err)
	}

	var nsList corev1.NamespaceList
	if err := c.List(ctx, &nsList, client.MatchingLabels(map[string]string{NamespaceSharedLabel: "true"})); err != nil {
		return nil, fmt.Errorf("list shared namespaces: %w", err)
	}
	for i := range nsList.Items {
		ns := nsList.Items[i].Name
		if err := c.Get(ctx, client.ObjectKey{Namespace: ns, Name: serviceName}, svc); err == nil {
			return svc, nil
		} else if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("get backing service %s/%s: %w", ns, serviceName, err)
		}
	}
	return nil, fmt.Errorf("backing service %q not found in %s or namespaces with %s=true",
		serviceName, app.Spec.Namespace, NamespaceSharedLabel)
}
