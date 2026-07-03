package terminus

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	linkerdPKISecretName        = "olares-linkerd-pki"
	linkerdIdentityIssuerSecret = "linkerd-identity-issuer"
	linkerdIdentityDeployment   = "linkerd-identity"
	linkerdIdentityIssuerCrtKey = "crt.pem"
	linkerdIdentityIssuerKeyKey = "key.pem"
	linkerdIdentityTrustRootsCM = "linkerd-identity-trust-roots"
	linkerdIdentityTrustRootsKey = "ca-bundle.crt"

	linkerdPKICAKey       = "ca.key"
	linkerdPKICACrt       = "ca.crt"
	linkerdPKIIssuerKey   = "issuer.key"
	linkerdPKIIssuerCrt   = "issuer.crt"
	linkerdPKIMetadataKey = "metadata.json"

	linkerdIssuerLifetimeDays = 1095
	linkerdCALifetimeDays      = 10950
)

type linkerdPKIMaterial struct {
	CACrt     []byte
	CAKey     []byte
	IssuerCrt []byte
	IssuerKey []byte
}

type linkerdPKIMetadata struct {
	CANotAfter     time.Time `json:"caNotAfter"`
	IssuerNotAfter time.Time `json:"issuerNotAfter"`
	Version        int       `json:"version"`
}

func prepareLinkerdPKI(ctx context.Context, c client.Client, linkerdNS string) (*linkerdPKIMaterial, error) {
	if err := ensureLinkerdNamespace(ctx, c, linkerdNS); err != nil {
		return nil, err
	}
	mat, ok, err := loadLinkerdPKISecret(ctx, c, linkerdNS)
	if err != nil {
		return nil, err
	}
	if ok {
		return mat, nil
	}
	mat, err = generateInitialLinkerdPKIMaterial()
	if err != nil {
		return nil, errors.Wrap(err, "generate linkerd pki material")
	}
	if err := writeLinkerdPKISecret(ctx, c, linkerdNS, mat, 1); err != nil {
		return nil, err
	}
	return mat, nil
}

func generateInitialLinkerdPKIMaterial() (*linkerdPKIMaterial, error) {
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	caNotBefore := time.Now().UTC().Add(-time.Hour)
	caNotAfter := caNotBefore.Add(linkerdCALifetimeDays * 24 * time.Hour)
	caSerial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}
	caTemplate := &x509.Certificate{
		SerialNumber:          caSerial,
		Subject:               pkix.Name{CommonName: "root.linkerd.cluster.local"},
		NotBefore:             caNotBefore,
		NotAfter:              caNotAfter,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, err
	}
	caCrtPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})
	caKeyPEM, err := marshalECPrivateKey(caKey)
	if err != nil {
		return nil, err
	}
	return rotateLinkerdIssuer(caCrtPEM, caKeyPEM)
}

func ensureLinkerdNamespace(ctx context.Context, c client.Client, ns string) error {
	desired := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
			Labels: map[string]string{
				"linkerd.io/is-control-plane": "true",
				"config.linkerd.io/admission-webhooks": "disabled",
				"linkerd.io/control-plane-ns": ns,
			},
		},
	}
	var existing corev1.Namespace
	err := c.Get(ctx, types.NamespacedName{Name: ns}, &existing)
	if apierrors.IsNotFound(err) {
		return c.Create(ctx, desired)
	}
	return err
}

func loadLinkerdPKISecret(ctx context.Context, c client.Client, ns string) (*linkerdPKIMaterial, bool, error) {
	var sec corev1.Secret
	err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: linkerdPKISecretName}, &sec)
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

func materialFromSecret(sec *corev1.Secret) (*linkerdPKIMaterial, error) {
	req := []string{linkerdPKICACrt, linkerdPKICAKey, linkerdPKIIssuerCrt, linkerdPKIIssuerKey}
	for _, k := range req {
		if len(sec.Data[k]) == 0 {
			return nil, fmt.Errorf("secret %s missing %s", sec.Name, k)
		}
	}
	return &linkerdPKIMaterial{
		CACrt:     sec.Data[linkerdPKICACrt],
		CAKey:     sec.Data[linkerdPKICAKey],
		IssuerCrt: sec.Data[linkerdPKIIssuerCrt],
		IssuerKey: sec.Data[linkerdPKIIssuerKey],
	}, nil
}

func writeLinkerdPKISecret(ctx context.Context, c client.Client, ns string, mat *linkerdPKIMaterial, version int) error {
	if version < 1 {
		version = 1
	}
	meta, err := buildLinkerdPKIMetadata(mat, version)
	if err != nil {
		return err
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	desired := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      linkerdPKISecretName,
			Namespace: ns,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "app-gateway",
				"app.kubernetes.io/component":  "linkerd-pki",
				"app.kubernetes.io/managed-by": "olares-cli",
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			linkerdPKICACrt:       mat.CACrt,
			linkerdPKICAKey:       mat.CAKey,
			linkerdPKIIssuerCrt:   mat.IssuerCrt,
			linkerdPKIIssuerKey:   mat.IssuerKey,
			linkerdPKIMetadataKey: metaBytes,
		},
	}
	var existing corev1.Secret
	err = c.Get(ctx, types.NamespacedName{Namespace: ns, Name: linkerdPKISecretName}, &existing)
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

func buildLinkerdPKIMetadata(mat *linkerdPKIMaterial, version int) (linkerdPKIMetadata, error) {
	caEnd, err := certificateNotAfter(mat.CACrt)
	if err != nil {
		return linkerdPKIMetadata{}, errors.Wrap(err, "parse ca.crt")
	}
	issuerEnd, err := certificateNotAfter(mat.IssuerCrt)
	if err != nil {
		return linkerdPKIMetadata{}, errors.Wrap(err, "parse issuer.crt")
	}
	return linkerdPKIMetadata{
		CANotAfter:     caEnd,
		IssuerNotAfter: issuerEnd,
		Version:        version,
	}, nil
}

func certificateNotAfter(pemBytes []byte) (time.Time, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return time.Time{}, errors.New("invalid PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, err
	}
	return cert.NotAfter, nil
}

func syncLinkerdIdentitySecrets(ctx context.Context, c client.Client, linkerdNS string, mat *linkerdPKIMaterial) error {
	if err := patchLinkerdIdentityIssuerSecret(ctx, c, linkerdNS, mat); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	return patchLinkerdTrustRootsConfigMap(ctx, c, linkerdNS, mat.CACrt)
}

func patchLinkerdIdentityIssuerSecret(ctx context.Context, c client.Client, ns string, mat *linkerdPKIMaterial) error {
	var sec corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: linkerdIdentityIssuerSecret}, &sec); err != nil {
		return err
	}
	if sec.Data == nil {
		sec.Data = map[string][]byte{}
	}
	sec.Data[linkerdIdentityIssuerCrtKey] = mat.IssuerCrt
	sec.Data[linkerdIdentityIssuerKeyKey] = mat.IssuerKey
	return c.Update(ctx, &sec)
}

func patchLinkerdTrustRootsConfigMap(ctx context.Context, c client.Client, ns string, caCrt []byte) error {
	var cm corev1.ConfigMap
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: linkerdIdentityTrustRootsCM}, &cm); err != nil {
		return err
	}
	if cm.Data == nil {
		cm.Data = map[string]string{}
	}
	cm.Data[linkerdIdentityTrustRootsKey] = string(caCrt)
	return c.Update(ctx, &cm)
}

func restartLinkerdIdentityIfNeeded(ctx context.Context, c client.Client, ns string) error {
	var dep appsv1.Deployment
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: linkerdIdentityDeployment}, &dep); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if dep.Spec.Template.Annotations == nil {
		dep.Spec.Template.Annotations = map[string]string{}
	}
	dep.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().UTC().Format(time.RFC3339)
	return c.Update(ctx, &dep)
}

func rotateLinkerdIssuer(caCrtPEM, caKeyPEM []byte) (*linkerdPKIMaterial, error) {
	caCert, err := parseCertificate(caCrtPEM)
	if err != nil {
		return nil, errors.Wrap(err, "parse ca.crt")
	}
	caKey, err := parseECPrivateKey(caKeyPEM)
	if err != nil {
		return nil, errors.Wrap(err, "parse ca.key")
	}
	issuerKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}
	notBefore := time.Now().UTC().Add(-time.Hour)
	notAfter := notBefore.Add(linkerdIssuerLifetimeDays * 24 * time.Hour)
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "identity.linkerd.cluster.local"},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            0,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, caCert, &issuerKey.PublicKey, caKey)
	if err != nil {
		return nil, err
	}
	issuerCrtPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	issuerKeyPEM, err := marshalECPrivateKey(issuerKey)
	if err != nil {
		return nil, err
	}
	return &linkerdPKIMaterial{
		CACrt: caCrtPEM, CAKey: caKeyPEM,
		IssuerCrt: issuerCrtPEM, IssuerKey: issuerKeyPEM,
	}, nil
}

func parseCertificate(pemBytes []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("invalid certificate PEM")
	}
	return x509.ParseCertificate(block.Bytes)
}

func parseECPrivateKey(pemBytes []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("invalid private key PEM")
	}
	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ec, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("ca.key is not ECDSA")
	}
	return ec, nil
}

func marshalECPrivateKey(key *ecdsa.PrivateKey) ([]byte, error) {
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der}), nil
}
