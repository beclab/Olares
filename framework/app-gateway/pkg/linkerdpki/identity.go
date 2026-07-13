package linkerdpki

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func patchIdentityIssuerSecret(ctx context.Context, c client.Client, ns string, mat *Material) error {
	var sec corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: identityIssuerSecret}, &sec); err != nil {
		return fmt.Errorf("get linkerd-identity-issuer: %w", err)
	}
	if sec.Data == nil {
		sec.Data = map[string][]byte{}
	}
	sec.Data[identityIssuerCrtKey] = mat.IssuerCrt
	sec.Data[identityIssuerKeyKey] = mat.IssuerKey
	return c.Update(ctx, &sec)
}

func restartIdentity(ctx context.Context, c client.Client, ns string) error {
	var dep appsv1.Deployment
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: identityDeployment}, &dep); err != nil {
		return fmt.Errorf("get linkerd-identity deployment: %w", err)
	}
	if dep.Spec.Template.Annotations == nil {
		dep.Spec.Template.Annotations = map[string]string{}
	}
	dep.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().UTC().Format(time.RFC3339)
	return c.Update(ctx, &dep)
}
