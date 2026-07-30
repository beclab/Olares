package linkerdpki

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Material holds trust anchor and issuer PEM bytes for the Linkerd PKI Secret.
type Material struct {
	CACrt     []byte
	CAKey     []byte
	IssuerCrt []byte
	IssuerKey []byte
}

type metadata struct {
	CANotAfter     time.Time `json:"caNotAfter"`
	IssuerNotAfter time.Time `json:"issuerNotAfter"`
	Version        int       `json:"version"`
}

func loadPKISecret(ctx context.Context, c client.Client, ns string) (*Material, bool, error) {
	var sec corev1.Secret
	err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: PKISecretName}, &sec)
	if apierrors.IsNotFound(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	mat, err := materialFromSecret(&sec)
	if err != nil {
		return nil, false, err
	}
	return mat, true, nil
}

func materialFromSecret(sec *corev1.Secret) (*Material, error) {
	req := []string{pkiCACrtKey, pkiCAKeyKey, pkiIssuerCrtKey, pkiIssuerKeyKey}
	for _, k := range req {
		if len(sec.Data[k]) == 0 {
			return nil, fmt.Errorf("secret %s missing %s", sec.Name, k)
		}
	}
	return &Material{
		CACrt:     sec.Data[pkiCACrtKey],
		CAKey:     sec.Data[pkiCAKeyKey],
		IssuerCrt: sec.Data[pkiIssuerCrtKey],
		IssuerKey: sec.Data[pkiIssuerKeyKey],
	}, nil
}

func writePKISecret(ctx context.Context, c client.Client, ns string, mat *Material, version int) error {
	if version < 1 {
		version = 1
	}
	meta, err := buildMetadata(mat, version)
	if err != nil {
		return err
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	desired := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PKISecretName,
			Namespace: ns,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "app-gateway",
				"app.kubernetes.io/component":  "linkerd-pki",
				"app.kubernetes.io/managed-by": "olares-cli",
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			pkiCACrtKey:     mat.CACrt,
			pkiCAKeyKey:     mat.CAKey,
			pkiIssuerCrtKey: mat.IssuerCrt,
			pkiIssuerKeyKey: mat.IssuerKey,
			pkiMetadataKey:  metaBytes,
		},
	}
	var existing corev1.Secret
	err = c.Get(ctx, types.NamespacedName{Namespace: ns, Name: PKISecretName}, &existing)
	if apierrors.IsNotFound(err) {
		return c.Create(ctx, desired)
	}
	if err != nil {
		return err
	}
	existing.Data = desired.Data
	existing.Labels = desired.Labels
	return c.Update(ctx, &existing)
}

func buildMetadata(mat *Material, version int) (metadata, error) {
	caEnd, err := certificateNotAfter(mat.CACrt)
	if err != nil {
		return metadata{}, fmt.Errorf("parse ca.crt: %w", err)
	}
	issuerEnd, err := certificateNotAfter(mat.IssuerCrt)
	if err != nil {
		return metadata{}, fmt.Errorf("parse issuer.crt: %w", err)
	}
	return metadata{
		CANotAfter:     caEnd,
		IssuerNotAfter: issuerEnd,
		Version:        version,
	}, nil
}
