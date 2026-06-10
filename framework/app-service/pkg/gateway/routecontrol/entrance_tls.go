package routecontrol

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	// zoneSSLConfigMapName is the platform TLS source maintained per user-space.
	zoneSSLConfigMapName = "zone-ssl-config"
	// gatewayTLSSecretName is the single HTTPS-listener cert the app-gateway
	// chart pins via certificateRefs. Lite: one shared platform cert.
	gatewayTLSSecretName = "app-gateway-tls"
	// annotationTLSContentHash makes the Secret upsert idempotent.
	annotationTLSContentHash = "gateway.olares.io/tls-content-hash"
)

// EntranceTLSReconciler syncs the platform cert from a zone-ssl-config
// ConfigMap into the single app-gateway/app-gateway-tls Secret that backs the
// HTTPS listener. Lite: one shared cert, no per-viewer Secrets or listeners.
type EntranceTLSReconciler struct {
	Client client.Client
}

// Reconcile copies cert/key from the observed zone-ssl-config ConfigMap into
// the gateway TLS Secret. A deleted ConfigMap leaves the last good Secret in
// place so the listener does not lose its certificate.
func (r *EntranceTLSReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	if r == nil || r.Client == nil {
		return reconcile.Result{}, nil
	}
	cm := &corev1.ConfigMap{}
	if err := r.Client.Get(ctx, req.NamespacedName, cm); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}
	return reconcile.Result{}, syncGatewayTLS(ctx, r.Client, cm)
}

func syncGatewayTLS(ctx context.Context, c client.Client, cm *corev1.ConfigMap) error {
	if cm == nil || cm.Name != zoneSSLConfigMapName {
		return nil
	}
	cert := strings.TrimSpace(cm.Data["cert"])
	key := strings.TrimSpace(cm.Data["key"])
	if cert == "" || key == "" {
		klog.Warningf("zone-ssl-config %s/%s missing cert or key, skip gateway TLS sync", cm.Namespace, cm.Name)
		return nil
	}
	hash := tlsMaterialHash(cert, key)

	current := &corev1.Secret{}
	err := c.Get(ctx, types.NamespacedName{Namespace: defaultGatewayNS, Name: gatewayTLSSecretName}, current)
	switch {
	case apierrors.IsNotFound(err):
		return c.Create(ctx, desiredGatewayTLSSecret(cert, key, hash))
	case err != nil:
		return err
	}
	if current.Annotations != nil && current.Annotations[annotationTLSContentHash] == hash {
		return nil
	}
	current.Type = corev1.SecretTypeTLS
	if current.Labels == nil {
		current.Labels = map[string]string{}
	}
	current.Labels[ManagedByLabel] = ManagedByValue
	if current.Annotations == nil {
		current.Annotations = map[string]string{}
	}
	current.Annotations[annotationTLSContentHash] = hash
	current.Data = nil
	current.StringData = map[string]string{
		corev1.TLSCertKey:       cert,
		corev1.TLSPrivateKeyKey: key,
	}
	return c.Update(ctx, current)
}

func desiredGatewayTLSSecret(cert, key, hash string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        gatewayTLSSecretName,
			Namespace:   defaultGatewayNS,
			Labels:      map[string]string{ManagedByLabel: ManagedByValue},
			Annotations: map[string]string{annotationTLSContentHash: hash},
		},
		Type: corev1.SecretTypeTLS,
		StringData: map[string]string{
			corev1.TLSCertKey:       cert,
			corev1.TLSPrivateKeyKey: key,
		},
	}
}

func tlsMaterialHash(cert, key string) string {
	sum := sha256.Sum256([]byte(cert + "\n" + key))
	return hex.EncodeToString(sum[:])
}

// SetupWithManager registers the reconciler against zone-ssl-config ConfigMaps.
func (r *EntranceTLSReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r == nil {
		return nil
	}
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	onlyZoneSSL := predicate.NewPredicateFuncs(func(o client.Object) bool {
		return o.GetName() == zoneSSLConfigMapName
	})
	return ctrl.NewControllerManagedBy(mgr).
		Named("entrance-tls").
		For(&corev1.ConfigMap{}, builder.WithPredicates(onlyZoneSSL)).
		Complete(r)
}
