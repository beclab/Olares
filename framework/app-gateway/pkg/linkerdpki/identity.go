package linkerdpki

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func issuerFingerprint(issuerCrt []byte) string {
	sum := sha256.Sum256(issuerCrt)
	return fmt.Sprintf("%x", sum)
}

// ensureIdentityRollout restarts linkerd-identity when its recorded issuer
// fingerprint does not match the vault. No-op when already matched.
func ensureIdentityRollout(ctx context.Context, c client.Client, ns string, issuerCrt []byte) error {
	var dep appsv1.Deployment
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: identityDeployment}, &dep); err != nil {
		slog.Error("linkerd pki get identity deployment failed",
			"op", "maintain", "namespace", ns, "error", err)
		return fmt.Errorf("get linkerd-identity deployment: %w", err)
	}
	want := issuerFingerprint(issuerCrt)
	if dep.Annotations[identityIssuerFingerprintAnnotation] == want {
		return nil
	}
	if dep.Annotations == nil {
		dep.Annotations = map[string]string{}
	}
	dep.Annotations[identityIssuerFingerprintAnnotation] = want
	if dep.Spec.Template.Annotations == nil {
		dep.Spec.Template.Annotations = map[string]string{}
	}
	dep.Spec.Template.Annotations[restartedAtAnnotation] = time.Now().UTC().Format(time.RFC3339)
	if err := c.Update(ctx, &dep); err != nil {
		slog.Error("linkerd pki identity rollout update failed",
			"op", "maintain", "namespace", ns, "error", err)
		return fmt.Errorf("rollout linkerd-identity: %w", err)
	}
	slog.Info("linkerd identity rollout scheduled",
		"op", "maintain", "namespace", ns, "issuer_fingerprint", want)
	return nil
}
