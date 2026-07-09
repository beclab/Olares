package terminus

import (
	"bytes"
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
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/core/logger"
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

	linkerdIdentitySecretsSyncTimeout  = 5 * time.Minute
	linkerdIdentitySecretsPollInterval = 5 * time.Second
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

func syncLinkerdIdentitySecrets(ctx context.Context, c client.Client, linkerdNS string, mat *linkerdPKIMaterial) (bool, error) {
	return waitSyncLinkerdIdentitySecrets(
		ctx, c, linkerdNS, mat,
		linkerdIdentitySecretsSyncTimeout,
		linkerdIdentitySecretsPollInterval,
	)
}

func waitSyncLinkerdIdentitySecrets(
	ctx context.Context,
	c client.Client,
	linkerdNS string,
	mat *linkerdPKIMaterial,
	timeout, pollInterval time.Duration,
) (bool, error) {
	start := time.Now()
	for {
		issuerReady, err := linkerdIdentityIssuerSecretExists(ctx, c, linkerdNS)
		if err != nil {
			logger.Errorf("sync linkerd identity secrets: check %s in namespace %s: %v", linkerdIdentityIssuerSecret, linkerdNS, err)
			return false, errors.Wrapf(err, "check %s", linkerdIdentityIssuerSecret)
		}
		trustReady, err := linkerdIdentityTrustRootsExists(ctx, c, linkerdNS)
		if err != nil {
			logger.Errorf("sync linkerd identity secrets: check %s in namespace %s: %v", linkerdIdentityTrustRootsCM, linkerdNS, err)
			return false, errors.Wrapf(err, "check %s", linkerdIdentityTrustRootsCM)
		}
		if issuerReady && trustReady {
			issuerChanged, err := patchLinkerdIdentityIssuerSecret(ctx, c, linkerdNS, mat)
			if err != nil {
				logger.Errorf("sync linkerd identity secrets: patch %s in namespace %s: %v", linkerdIdentityIssuerSecret, linkerdNS, err)
				return false, errors.Wrapf(err, "patch %s", linkerdIdentityIssuerSecret)
			}
			trustChanged, err := patchLinkerdTrustRootsConfigMap(ctx, c, linkerdNS, mat.CACrt)
			if err != nil {
				logger.Errorf("sync linkerd identity secrets: patch %s in namespace %s: %v", linkerdIdentityTrustRootsCM, linkerdNS, err)
				return false, errors.Wrapf(err, "patch %s", linkerdIdentityTrustRootsCM)
			}
			return issuerChanged || trustChanged, nil
		}
		var pending []string
		if !issuerReady {
			pending = append(pending, linkerdIdentityIssuerSecret)
		}
		if !trustReady {
			pending = append(pending, linkerdIdentityTrustRootsCM)
		}
		if time.Since(start) >= timeout {
			err := fmt.Errorf(
				"sync linkerd identity secrets: timed out after %s waiting for %s in namespace %s",
				timeout,
				strings.Join(pending, ", "),
				linkerdNS,
			)
			logger.Error(err)
			return false, err
		}
		select {
		case <-ctx.Done():
			logger.Errorf("sync linkerd identity secrets: context cancelled in namespace %s: %v", linkerdNS, ctx.Err())
			return false, ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}

func linkerdIdentityIssuerSecretExists(ctx context.Context, c client.Client, ns string) (bool, error) {
	var sec corev1.Secret
	err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: linkerdIdentityIssuerSecret}, &sec)
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func linkerdIdentityTrustRootsExists(ctx context.Context, c client.Client, ns string) (bool, error) {
	var cm corev1.ConfigMap
	err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: linkerdIdentityTrustRootsCM}, &cm)
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func patchLinkerdIdentityIssuerSecret(ctx context.Context, c client.Client, ns string, mat *linkerdPKIMaterial) (bool, error) {
	var sec corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: linkerdIdentityIssuerSecret}, &sec); err != nil {
		return false, err
	}
	if sec.Data == nil {
		sec.Data = map[string][]byte{}
	}
	if bytes.Equal(sec.Data[linkerdIdentityIssuerCrtKey], mat.IssuerCrt) &&
		bytes.Equal(sec.Data[linkerdIdentityIssuerKeyKey], mat.IssuerKey) {
		return false, nil
	}
	sec.Data[linkerdIdentityIssuerCrtKey] = mat.IssuerCrt
	sec.Data[linkerdIdentityIssuerKeyKey] = mat.IssuerKey
	return true, c.Update(ctx, &sec)
}

func patchLinkerdTrustRootsConfigMap(ctx context.Context, c client.Client, ns string, caCrt []byte) (bool, error) {
	var cm corev1.ConfigMap
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: linkerdIdentityTrustRootsCM}, &cm); err != nil {
		return false, err
	}
	if cm.Data == nil {
		cm.Data = map[string]string{}
	}
	desired := string(caCrt)
	if cm.Data[linkerdIdentityTrustRootsKey] == desired {
		return false, nil
	}
	cm.Data[linkerdIdentityTrustRootsKey] = desired
	return true, c.Update(ctx, &cm)
}

// restartLinkerdControlPlaneAfterPKISync rolls all Linkerd control-plane Deployments
// so sidecars re-read linkerd-identity-trust-roots after a CA or issuer change.
func restartLinkerdControlPlaneAfterPKISync(ctx context.Context, c client.Client, ns string) error {
	restartedAt := time.Now().UTC().Format(time.RFC3339)
	for _, name := range linkerdControlPlaneDeployments {
		if err := restartLinkerdDeployment(ctx, c, ns, name, restartedAt); err != nil {
			return errors.Wrapf(err, "restart %s", name)
		}
	}
	return nil
}

func restartLinkerdDeployment(ctx context.Context, c client.Client, ns, name, restartedAt string) error {
	var dep appsv1.Deployment
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, &dep); err != nil {
		return err
	}
	if dep.Spec.Template.Annotations == nil {
		dep.Spec.Template.Annotations = map[string]string{}
	}
	dep.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = restartedAt
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
