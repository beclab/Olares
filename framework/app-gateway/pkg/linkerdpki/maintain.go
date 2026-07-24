package linkerdpki

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MaintainLinkerdPKI performs one level-triggered check/rotate cycle. It reads
// the in-cluster olares-linkerd-pki Secret only and rotates the identity issuer
// when its remaining validity drops below IssuerRotateThreshold.
func MaintainLinkerdPKI(ctx context.Context, c client.Client, linkerdNS string) error {
	mat, ok, err := loadPKISecret(ctx, c, linkerdNS)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("olares-linkerd-pki secret not found; run install-app-gateway first")
	}
	need, remaining, err := IssuerNeedsRotation(mat.IssuerCrt, time.Now().UTC())
	if err != nil {
		return err
	}
	if !need {
		slog.Info("linkerd issuer ok", "remaining_hours", remaining.Round(time.Hour).Hours())
		return nil
	}
	slog.Info("linkerd issuer needs rotation", "remaining_hours", remaining.Round(time.Hour).Hours())

	rotated, err := rotateIssuer(mat.CACrt, mat.CAKey)
	if err != nil {
		return fmt.Errorf("rotate linkerd issuer: %w", err)
	}

	version := 1
	var sec corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{Namespace: linkerdNS, Name: PKISecretName}, &sec); err == nil {
		if metaBytes := sec.Data[pkiMetadataKey]; len(metaBytes) > 0 {
			var meta metadata
			if json.Unmarshal(metaBytes, &meta) == nil {
				version = meta.Version + 1
			}
		}
	}
	if err := writePKISecret(ctx, c, linkerdNS, rotated, version); err != nil {
		return err
	}
	if err := patchIdentityIssuerSecret(ctx, c, linkerdNS, rotated); err != nil {
		return err
	}
	if err := restartIdentity(ctx, c, linkerdNS); err != nil {
		return err
	}
	slog.Info("linkerd identity issuer rotated", "version", version)
	return nil
}
