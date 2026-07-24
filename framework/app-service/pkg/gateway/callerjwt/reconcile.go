package callerjwt

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	// JWTRefreshInterval re-issues caller JWT-SVIDs before MaxTTL expiry.
	JWTRefreshInterval = 30 * time.Minute
)

const (
	settingNeedsSharedAccess = "needsSharedAccess"
	settingSharedAppDeps     = "sharedAppDeps"
	settingClusterAppRef     = "clusterAppRef"
	managedByLabel           = "app.kubernetes.io/managed-by"
	managedByValue           = "app-service"
)

// IssuerReconciler issues caller JWT-SVID secrets for Shared consumer apps and
// maintains the cluster JWKS service (WI-OC-C2-01 L1-a).
type IssuerReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
	issuer *Issuer
}

// Reconcile ensures signing keys, JWKS service, and per-app JWT secrets.
func (r *IssuerReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	if r == nil || r.Client == nil {
		return reconcile.Result{}, nil
	}
	if err := r.ensureIssuer(ctx); err != nil {
		return reconcile.Result{}, err
	}
	if req.Namespace == JWKSServiceNamespace && req.Name == IssuerKeysSecretName {
		if err := r.reloadIssuer(ctx); err != nil {
			return reconcile.Result{}, err
		}
		if err := r.reconcileJWKSSurface(ctx); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}
	app := &appv1alpha1.Application{}
	if err := r.Client.Get(ctx, req.NamespacedName, app); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}
	if err := r.reconcileApplication(ctx, app); err != nil {
		klog.Errorf("callerjwt: reconcile application %s/%s failed: %v", app.Namespace, app.Name, err)
		return reconcile.Result{}, err
	}
	if applicationDeclaresSharedAccess(app) {
		return reconcile.Result{RequeueAfter: JWTRefreshInterval}, nil
	}
	return reconcile.Result{}, nil
}

func (r *IssuerReconciler) reconcileApplication(ctx context.Context, app *appv1alpha1.Application) error {
	if app == nil || app.Spec.Namespace == "" {
		return nil
	}
	if !applicationDeclaresSharedAccess(app) {
		return deleteAppJWTSecret(ctx, r.Client, app.Spec.Namespace)
	}
	token, err := r.issuer.Issue(issueRequestFromApplication(app))
	if err != nil {
		klog.Warningf("caller_jwt_issue_fail app=%s ns=%s err=%v", app.Spec.Name, app.Spec.Namespace, err)
		return err
	}
	if err := upsertAppJWTSecret(ctx, r.Client, r.Scheme, app, token); err != nil {
		return err
	}
	klog.V(1).Infof("caller jwt issued app=%s ns=%s", app.Spec.Name, app.Spec.Namespace)
	return nil
}

func (r *IssuerReconciler) ensureIssuer(ctx context.Context) error {
	if r.issuer != nil {
		return nil
	}
	ring, err := loadOrCreateKeyRing(ctx, r.Client)
	if err != nil {
		return err
	}
	issuer, err := NewIssuer(ring)
	if err != nil {
		return err
	}
	r.issuer = issuer
	return r.reconcileJWKSSurface(ctx)
}

func (r *IssuerReconciler) reconcileJWKSSurface(ctx context.Context) error {
	if err := r.reconcileJWKSService(ctx); err != nil {
		return err
	}
	if err := r.reconcileJWKSIngressNP(ctx); err != nil {
		return err
	}
	return r.reconcileJWKSTrust(ctx)
}

func (r *IssuerReconciler) reloadIssuer(ctx context.Context) error {
	ring, err := loadOrCreateKeyRing(ctx, r.Client)
	if err != nil {
		return err
	}
	issuer, err := NewIssuer(ring)
	if err != nil {
		return err
	}
	r.issuer = issuer
	return nil
}

func (r *IssuerReconciler) reconcileJWKSService(ctx context.Context) error {
	if r.issuer == nil {
		return nil
	}
	portName, targetPort := JWKSListenPort()
	desired := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      JWKSServiceName,
			Namespace: JWKSServiceNamespace,
			Labels: map[string]string{
				managedByLabel: managedByValue,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"tier": "app-service",
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "https",
					Port:       jwksServicePort,
					TargetPort: intstr.FromString(portName),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
	_ = targetPort

	current := &corev1.Service{}
	key := types.NamespacedName{Name: JWKSServiceName, Namespace: JWKSServiceNamespace}
	err := r.Client.Get(ctx, key, current)
	switch {
	case apierrors.IsNotFound(err):
		return r.Client.Create(ctx, desired)
	case err != nil:
		return err
	default:
		current.Spec.Selector = desired.Spec.Selector
		current.Spec.Ports = desired.Spec.Ports
		if current.Labels == nil {
			current.Labels = map[string]string{}
		}
		current.Labels[managedByLabel] = managedByValue
		return r.Client.Update(ctx, current)
	}
}

// Issuer returns the reconciler signing issuer after keys are loaded.
func (r *IssuerReconciler) Issuer() *Issuer {
	if r == nil {
		return nil
	}
	return r.issuer
}

// SetupWithManager registers the issuer reconciler and JWKS HTTPS server.
// Do not touch the API/cache here: controller-runtime cache is not started
// until mgr.Start, and eager ensureIssuer caused CrashLoopBackOff (P8).
func (r *IssuerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r == nil {
		return nil
	}
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	if r.Scheme == nil {
		r.Scheme = mgr.GetScheme()
	}
	if err := mgr.Add(&JWKSServer{Reconciler: r}); err != nil {
		return fmt.Errorf("add jwks server: %w", err)
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named("caller-jwt-issuer").
		For(&appv1alpha1.Application{}).
		Complete(r)
}

func applicationDeclaresSharedAccess(app *appv1alpha1.Application) bool {
	if app == nil || app.Spec.Settings == nil {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(app.Spec.Settings[settingNeedsSharedAccess]), "true") {
		return true
	}
	if strings.TrimSpace(app.Spec.Settings[settingSharedAppDeps]) != "" {
		return true
	}
	return strings.TrimSpace(app.Spec.Settings[settingClusterAppRef]) != ""
}

func issueRequestFromApplication(app *appv1alpha1.Application) IssueRequest {
	sa := strings.TrimSpace(app.Spec.Name)
	if app.Spec.Settings != nil {
		if v := strings.TrimSpace(app.Spec.Settings["serviceAccountName"]); v != "" {
			sa = v
		}
	}
	viewer := strings.TrimSpace(app.Spec.Owner)
	// olares.entrance is the caller's entrance name for per-entrance policy.
	// Do not map sharedAppDeps (callee refs) into this claim.
	return IssueRequest{
		Namespace:          app.Spec.Namespace,
		ServiceAccountName: sa,
		AppRef:             app.Spec.Name,
		Viewer:             viewer,
	}
}

func upsertAppJWTSecret(ctx context.Context, c client.Client, scheme *runtime.Scheme, app *appv1alpha1.Application, token string) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AppJWTSecretName,
			Namespace: app.Spec.Namespace,
			Labels: map[string]string{
				managedByLabel: managedByValue,
				"app":          app.Spec.Name,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			AppJWTSecretDataKey: []byte(token),
		},
	}
	if scheme != nil && app.Namespace == app.Spec.Namespace {
		if err := ctrl.SetControllerReference(app, secret, scheme); err != nil {
			return err
		}
	}
	current := &corev1.Secret{}
	key := types.NamespacedName{Name: AppJWTSecretName, Namespace: app.Spec.Namespace}
	err := c.Get(ctx, key, current)
	switch {
	case apierrors.IsNotFound(err):
		return c.Create(ctx, secret)
	case err != nil:
		return err
	default:
		current.Data = secret.Data
		current.Labels = secret.Labels
		return c.Update(ctx, current)
	}
}

func deleteAppJWTSecret(ctx context.Context, c client.Client, namespace string) error {
	secret := &corev1.Secret{}
	key := types.NamespacedName{Name: AppJWTSecretName, Namespace: namespace}
	if err := c.Get(ctx, key, secret); err != nil {
		return client.IgnoreNotFound(err)
	}
	return client.IgnoreNotFound(c.Delete(ctx, secret))
}

func loadOrCreateKeyRing(ctx context.Context, c client.Client) (KeyRing, error) {
	secret := &corev1.Secret{}
	key := types.NamespacedName{Name: IssuerKeysSecretName, Namespace: JWKSServiceNamespace}
	err := c.Get(ctx, key, secret)
	if apierrors.IsNotFound(err) {
		return createIssuerKeysSecret(ctx, c)
	}
	if err != nil {
		return KeyRing{}, err
	}
	return keyRingFromSecret(secret)
}

func createIssuerKeysSecret(ctx context.Context, c client.Client) (KeyRing, error) {
	active, err := GenerateKeyPair()
	if err != nil {
		return KeyRing{}, err
	}
	prev, err := GenerateKeyPair()
	if err != nil {
		return KeyRing{}, err
	}
	ring := KeyRing{Active: active, Previous: &prev}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      IssuerKeysSecretName,
			Namespace: JWKSServiceNamespace,
			Labels: map[string]string{
				managedByLabel: managedByValue,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			SigningKeyPEM:    encodePrivateKeyPEM(active),
			SigningKeyIDKey:  []byte(active.KID),
			PreviousKeyPEM:   encodePrivateKeyPEM(prev),
			PreviousKeyIDKey: []byte(prev.KID),
		},
	}
	if err := c.Create(ctx, secret); err != nil {
		return KeyRing{}, err
	}
	return ring, nil
}

func keyRingFromSecret(secret *corev1.Secret) (KeyRing, error) {
	if secret == nil || secret.Data == nil {
		return KeyRing{}, fmt.Errorf("callerjwt: issuer keys secret is empty")
	}
	active, err := KeyPairFromPEM(secret.Data[SigningKeyPEM], string(secret.Data[SigningKeyIDKey]))
	if err != nil {
		return KeyRing{}, err
	}
	ring := KeyRing{Active: active}
	if pemBytes := secret.Data[PreviousKeyPEM]; len(pemBytes) > 0 {
		prev, err := KeyPairFromPEM(pemBytes, string(secret.Data[PreviousKeyIDKey]))
		if err != nil {
			return KeyRing{}, err
		}
		ring.Previous = &prev
	}
	return ring, nil
}

func encodePrivateKeyPEM(pair KeyPair) []byte {
	if pair.PrivateKey == nil {
		return nil
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(pair.PrivateKey),
	})
}
