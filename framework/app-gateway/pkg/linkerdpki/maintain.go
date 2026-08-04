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

// MaintainLinkerdPKI checks whether the vault issuer needs rotation, then
// always syncs linkerd-identity-issuer and restarts linkerd-identity when
// it still lags the vault.
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
	if need {
		slog.Info("linkerd issuer needs rotation", "remaining_hours", remaining.Round(time.Hour).Hours())
		rotated, err := rotateIssuer(mat.CACrt, mat.CAKey)
		if err != nil {
			slog.Error("linkerd pki rotate issuer failed",
				"op", "maintain", "namespace", linkerdNS, "error", err)
			return fmt.Errorf("rotate linkerd issuer: %w", err)
		}
		version := nextPKIVersion(ctx, c, linkerdNS)
		if err := writePKISecret(ctx, c, linkerdNS, rotated, version); err != nil {
			slog.Error("linkerd pki write vault failed",
				"op", "maintain", "namespace", linkerdNS, "error", err)
			return err
		}
		mat = rotated
		slog.Info("linkerd pki vault rotated", "op", "maintain", "namespace", linkerdNS, "version", version)
	} else {
		slog.Info("linkerd issuer ok", "remaining_hours", remaining.Round(time.Hour).Hours())
	}

	if _, err := syncIdentityIssuerSecret(ctx, c, linkerdNS, mat); err != nil {
		slog.Error("linkerd pki sync identity issuer failed",
			"op", "maintain", "namespace", linkerdNS, "error", err)
		return fmt.Errorf("sync identity issuer: %w", err)
	}
	if err := ensureIdentityRollout(ctx, c, linkerdNS, mat.IssuerCrt); err != nil {
		return err
	}
	return nil
}

func nextPKIVersion(ctx context.Context, c client.Client, ns string) int {
	version := 1
	var sec corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: PKISecretName}, &sec); err != nil {
		return version
	}
	if metaBytes := sec.Data[pkiMetadataKey]; len(metaBytes) > 0 {
		var meta metadata
		if json.Unmarshal(metaBytes, &meta) == nil {
			version = meta.Version + 1
		}
	}
	return version
}
