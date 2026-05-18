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
	"os"
	"path/filepath"
	"time"

	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// olares-linkerd-pki stores ca.key for issuer rotation; restrict Secret access via RBAC.
	linkerdPKISecretName           = "olares-linkerd-pki"
	linkerdIdentityIssuerSecret    = "linkerd-identity-issuer"
	linkerdIdentityDeployment      = "linkerd-identity"
	linkerdIdentityIssuerCrtKey    = "crt.pem"
	linkerdIdentityIssuerKeyKey    = "key.pem"
	linkerdIdentityTrustRootsCM    = "linkerd-identity-trust-roots"
	linkerdIdentityTrustRootsKey   = "ca-bundle.crt"

	linkerdIssuerLifetimeDays = 1095 // 3 years
	linkerdIssuerRotateThreshold   = 180 * 24 * time.Hour // 6 months

	linkerdPKICAKey       = "ca.key"
	linkerdPKICACrt       = "ca.crt"
	linkerdPKIIssuerKey   = "issuer.key"
	linkerdPKIIssuerCrt   = "issuer.crt"
	linkerdPKIMetadataKey = "metadata.json"
)

// linkerdPKIMaterial holds trust anchor and issuer PEM bytes for Linkerd Helm / Secrets.
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

func enrichLinkerdHelmValues(ctx context.Context, c client.Client, linkerdNS, vendorDir string, vals map[string]interface{}) error {
	if vals == nil {
		vals = map[string]interface{}{}
	}
	mat, err := loadOrCreateLinkerdPKI(ctx, c, linkerdNS, vendorDir)
	if err != nil {
		return err
	}
	if err := applyLinkerdPKIMaterialToHelmValues(vals, mat); err != nil {
		return err
	}
	if wipeInstallerLinkerdKeysEnabled() {
		wipeInstallerLinkerdPrivateKeys(vendorDir)
	}
	return nil
}

func loadOrCreateLinkerdPKI(ctx context.Context, c client.Client, linkerdNS, vendorDir string) (*linkerdPKIMaterial, error) {
	if mat, ok, err := loadLinkerdPKISecret(ctx, c, linkerdNS); err != nil {
		return nil, err
	} else if ok {
		return mat, nil
	}
	certDir := filepath.Join(vendorDir, linkerdIdentityCertDirName)
	if linkerdIdentityCertsPresent(certDir) {
		mat, err := readLinkerdPKIDir(certDir)
		if err != nil {
			return nil, err
		}
		if err := writeLinkerdPKISecret(ctx, c, linkerdNS, mat, 1); err != nil {
			return nil, err
		}
		return mat, nil
	}
	if err := ensureLinkerdIdentityCerts(certDir); err != nil {
		return nil, err
	}
	mat, err := readLinkerdPKIDir(certDir)
	if err != nil {
		return nil, err
	}
	if err := writeLinkerdPKISecret(ctx, c, linkerdNS, mat, 1); err != nil {
		return nil, err
	}
	return mat, nil
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

func readLinkerdPKIDir(dir string) (*linkerdPKIMaterial, error) {
	read := func(name string) ([]byte, error) {
		return os.ReadFile(filepath.Join(dir, name))
	}
	caCrt, err := read("ca.crt")
	if err != nil {
		return nil, errors.Wrap(err, "read ca.crt")
	}
	caKey, err := read("ca.key")
	if err != nil {
		return nil, errors.Wrap(err, "read ca.key")
	}
	issuerCrt, err := read("issuer.crt")
	if err != nil {
		return nil, errors.Wrap(err, "read issuer.crt")
	}
	issuerKey, err := read("issuer.key")
	if err != nil {
		return nil, errors.Wrap(err, "read issuer.key")
	}
	return &linkerdPKIMaterial{
		CACrt: caCrt, CAKey: caKey, IssuerCrt: issuerCrt, IssuerKey: issuerKey,
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

func applyLinkerdPKIMaterialToHelmValues(vals map[string]interface{}, mat *linkerdPKIMaterial) error {
	vals["identityTrustAnchorsPEM"] = string(mat.CACrt)
	identity, _ := vals["identity"].(map[string]interface{})
	if identity == nil {
		identity = map[string]interface{}{}
	}
	issuer, _ := identity["issuer"].(map[string]interface{})
	if issuer == nil {
		issuer = map[string]interface{}{}
	}
	issuer["scheme"] = "linkerd.io/tls"
	tls, _ := issuer["tls"].(map[string]interface{})
	if tls == nil {
		tls = map[string]interface{}{}
	}
	tls["crtPEM"] = string(mat.IssuerCrt)
	tls["keyPEM"] = string(mat.IssuerKey)
	issuer["tls"] = tls
	identity["issuer"] = issuer
	vals["identity"] = identity
	return nil
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

func issuerNeedsRotation(issuerPEM []byte, now time.Time) (bool, time.Duration, error) {
	notAfter, err := certificateNotAfter(issuerPEM)
	if err != nil {
		return false, 0, err
	}
	remaining := notAfter.Sub(now)
	return remaining < linkerdIssuerRotateThreshold, remaining, nil
}

// MaintainLinkerdPKI rotates the identity issuer when remaining validity is under 6 months.
func MaintainLinkerdPKI(ctx context.Context, c client.Client, linkerdNS, _ string) error {
	mat, ok, err := loadLinkerdPKISecret(ctx, c, linkerdNS)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("olares-linkerd-pki secret not found; run install-app-gateway first")
	}
	need, remaining, err := issuerNeedsRotation(mat.IssuerCrt, time.Now().UTC())
	if err != nil {
		return err
	}
	if !need {
		logger.Infof("Linkerd issuer OK (remaining %s)", remaining.Round(time.Hour))
		return nil
	}
	logger.InfoInstallationProgress(fmt.Sprintf("Linkerd issuer expires in %s; rotating ...", remaining.Round(time.Hour)))

	rotated, err := rotateLinkerdIssuer(mat.CACrt, mat.CAKey)
	if err != nil {
		return errors.Wrap(err, "rotate linkerd issuer")
	}

	version := 1
	var sec corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{Namespace: linkerdNS, Name: linkerdPKISecretName}, &sec); err == nil {
		if metaBytes := sec.Data[linkerdPKIMetadataKey]; len(metaBytes) > 0 {
			var meta linkerdPKIMetadata
			if json.Unmarshal(metaBytes, &meta) == nil {
				version = meta.Version + 1
			}
		}
	}
	if err := writeLinkerdPKISecret(ctx, c, linkerdNS, rotated, version); err != nil {
		return err
	}
	if err := patchLinkerdIdentityIssuerSecret(ctx, c, linkerdNS, rotated); err != nil {
		return err
	}
	if err := restartLinkerdIdentity(ctx, c, linkerdNS); err != nil {
		return err
	}
	logger.InfoInstallationProgress("Linkerd identity issuer rotated successfully")
	return nil
}

func patchLinkerdIdentityIssuerSecret(ctx context.Context, c client.Client, ns string, mat *linkerdPKIMaterial) error {
	var sec corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: linkerdIdentityIssuerSecret}, &sec); err != nil {
		return errors.Wrap(err, "get linkerd-identity-issuer")
	}
	if sec.Data == nil {
		sec.Data = map[string][]byte{}
	}
	sec.Data[linkerdIdentityIssuerCrtKey] = mat.IssuerCrt
	sec.Data[linkerdIdentityIssuerKeyKey] = mat.IssuerKey
	return c.Update(ctx, &sec)
}

func restartLinkerdIdentity(ctx context.Context, c client.Client, ns string) error {
	var dep appsv1.Deployment
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: linkerdIdentityDeployment}, &dep); err != nil {
		return errors.Wrap(err, "get linkerd-identity deployment")
	}
	if dep.Spec.Template.Annotations == nil {
		dep.Spec.Template.Annotations = map[string]string{}
	}
	dep.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().UTC().Format(time.RFC3339)
	return c.Update(ctx, &dep)
}

func ensureLinkerdPKIGuardian(ctx context.Context, c client.Client, linkerdNS string) error {
	if os.Getenv("OLARES_LINKERD_PKI_GUARDIAN") == "0" {
		return nil
	}
	if err := ensureLinkerdPKIGuardianRBAC(ctx, c, linkerdNS); err != nil {
		return err
	}
	return applyLinkerdPKIGuardianCronJob(ctx, c, linkerdNS)
}

func ensureLinkerdPKIGuardianRBAC(ctx context.Context, c client.Client, linkerdNS string) error {
	var err error
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Name: "linkerd-pki-guardian", Namespace: linkerdNS},
	}
	var existingSA corev1.ServiceAccount
	if err := c.Get(ctx, client.ObjectKeyFromObject(sa), &existingSA); apierrors.IsNotFound(err) {
		if err := c.Create(ctx, sa); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Name: "linkerd-pki-guardian", Namespace: linkerdNS},
		Rules: []rbacv1.PolicyRule{
			{APIGroups: []string{""}, Resources: []string{"secrets"}, Verbs: []string{"get", "update", "patch"}},
			{APIGroups: []string{"apps"}, Resources: []string{"deployments"}, Verbs: []string{"get", "update", "patch"}},
		},
	}
	var existingRole rbacv1.Role
	if err := c.Get(ctx, client.ObjectKeyFromObject(role), &existingRole); apierrors.IsNotFound(err) {
		if err := c.Create(ctx, role); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	binding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "linkerd-pki-guardian", Namespace: linkerdNS},
		Subjects: []rbacv1.Subject{{
			Kind: "ServiceAccount", Name: "linkerd-pki-guardian", Namespace: linkerdNS,
		}},
		RoleRef: rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "Role", Name: "linkerd-pki-guardian"},
	}
	var existingBinding rbacv1.RoleBinding
	err = c.Get(ctx, client.ObjectKeyFromObject(binding), &existingBinding)
	if apierrors.IsNotFound(err) {
		return c.Create(ctx, binding)
	}
	return err
}

func applyLinkerdPKIGuardianCronJob(ctx context.Context, c client.Client, linkerdNS string) error {
	image := os.Getenv("OLARES_CLI_JOB_IMAGE")
	if image == "" {
		logger.Info("OLARES_CLI_JOB_IMAGE unset; skip linkerd-pki-guardian CronJob (use: olares-cli maintain-linkerd-pki)")
		return nil
	}
	cron := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "linkerd-pki-guardian",
			Namespace: linkerdNS,
			Labels: map[string]string{
				"app.kubernetes.io/name":      "app-gateway",
				"app.kubernetes.io/component": "linkerd-pki-guardian",
			},
		},
		Spec: batchv1.CronJobSpec{
			Schedule: "0 0 1 1,7 *",
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app.kubernetes.io/name":      "app-gateway",
								"app.kubernetes.io/component": "linkerd-pki-guardian",
							},
						},
						Spec: corev1.PodSpec{
							RestartPolicy:      corev1.RestartPolicyOnFailure,
							ServiceAccountName: "linkerd-pki-guardian",
							Containers: []corev1.Container{{
								Name:            "guardian",
								Image:           image,
								ImagePullPolicy: corev1.PullIfNotPresent,
								Command:         []string{"olares-cli", "maintain-linkerd-pki"},
							}},
						},
					},
				},
			},
			ConcurrencyPolicy:          batchv1.ForbidConcurrent,
			SuccessfulJobsHistoryLimit: int32Ptr(2),
			FailedJobsHistoryLimit:     int32Ptr(3),
		},
	}
	var existing batchv1.CronJob
	err := c.Get(ctx, types.NamespacedName{Namespace: linkerdNS, Name: cron.Name}, &existing)
	if apierrors.IsNotFound(err) {
		return c.Create(ctx, cron)
	}
	if err != nil {
		return err
	}
	existing.Spec = cron.Spec
	existing.Labels = cron.Labels
	return c.Update(ctx, &existing)
}

func int32Ptr(v int32) *int32 { return &v }

func wipeInstallerLinkerdKeysEnabled() bool {
	v := os.Getenv("OLARES_LINKERD_PKI_WIPE_INSTALLER_KEYS")
	return v == "" || v == "1" || v == "true" || v == "TRUE"
}

func wipeInstallerLinkerdPrivateKeys(vendorDir string) {
	dir := filepath.Join(vendorDir, linkerdIdentityCertDirName)
	for _, name := range []string{"ca.key", "issuer.key"} {
		_ = os.Remove(filepath.Join(dir, name))
	}
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
