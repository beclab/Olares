package routecontrol

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
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
	zoneSSLConfigMapName     = "zone-ssl-config"
	sharedEntranceTLSPrefix  = "shared-entrance-tls-"
	gatewayTLSSecretName     = "app-gateway-tls"
	annotationTLSContentHash = "gateway.olares.io/tls-content-hash"
	annotationTLSSourceNS    = "gateway.olares.io/tls-source-ns"
	labelTLSViewer           = "gateway.olares.io/tls-viewer"
	userSpacePrefix          = "user-space-"
)

var dns1123Label = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

// EntranceTLSReconciler syncs per-viewer platform certs from zone-ssl-config
// ConfigMaps into os-gateway/shared-entrance-tls-<viewer> Secrets.
type EntranceTLSReconciler struct {
	Client client.Client
}

// Reconcile copies cert/key from zone-ssl-config into the viewer TLS Secret.
// ConfigMap deletion or incomplete material garbage-collects the Secret.
func (r *EntranceTLSReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	if r == nil || r.Client == nil {
		return reconcile.Result{}, nil
	}
	viewer, ok := viewerFromUserSpaceNS(req.Namespace)
	if !ok {
		return reconcile.Result{}, nil
	}

	cm := &corev1.ConfigMap{}
	if err := r.Client.Get(ctx, req.NamespacedName, cm); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, deletePerViewerTLSSecret(ctx, r.Client, viewer)
		}
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, syncPerViewerTLS(ctx, r.Client, cm, viewer)
}

func syncPerViewerTLS(ctx context.Context, c client.Client, cm *corev1.ConfigMap, viewer string) error {
	if cm == nil || cm.Name != zoneSSLConfigMapName {
		return nil
	}
	// if cm.Data != nil && cm.Data["ephemeral"] == "true" {
	// 	return deletePerViewerTLSSecret(ctx, c, viewer)
	// }
	cert := strings.TrimSpace(cm.Data["cert"])
	key := strings.TrimSpace(cm.Data["key"])
	if cert == "" || key == "" {
		klog.Warningf("zone-ssl-config %s/%s missing cert or key, gc gateway TLS secret", cm.Namespace, cm.Name)
		return deletePerViewerTLSSecret(ctx, c, viewer)
	}
	hash := tlsMaterialHash(cert, key)
	secretName := sharedEntranceTLSName(viewer)

	current := &corev1.Secret{}
	err := c.Get(ctx, types.NamespacedName{Namespace: defaultGatewayNS, Name: secretName}, current)
	switch {
	case apierrors.IsNotFound(err):
		return c.Create(ctx, desiredPerViewerTLSSecret(viewer, cm.Namespace, cert, key, hash))
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
	current.Labels[labelTLSViewer] = viewer
	if current.Annotations == nil {
		current.Annotations = map[string]string{}
	}
	current.Annotations[annotationTLSContentHash] = hash
	current.Annotations[annotationTLSSourceNS] = cm.Namespace
	current.Data = nil
	current.StringData = map[string]string{
		corev1.TLSCertKey:       cert,
		corev1.TLSPrivateKeyKey: key,
	}
	return c.Update(ctx, current)
}

func desiredPerViewerTLSSecret(viewer, sourceNS, cert, key, hash string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sharedEntranceTLSName(viewer),
			Namespace: defaultGatewayNS,
			Labels: map[string]string{
				ManagedByLabel: ManagedByValue,
				labelTLSViewer: viewer,
			},
			Annotations: map[string]string{
				annotationTLSContentHash: hash,
				annotationTLSSourceNS:    sourceNS,
			},
		},
		Type: corev1.SecretTypeTLS,
		StringData: map[string]string{
			corev1.TLSCertKey:       cert,
			corev1.TLSPrivateKeyKey: key,
		},
	}
}

func deletePerViewerTLSSecret(ctx context.Context, c client.Client, viewer string) error {
	sec := &corev1.Secret{}
	key := types.NamespacedName{Namespace: defaultGatewayNS, Name: sharedEntranceTLSName(viewer)}
	if err := c.Get(ctx, key, sec); err != nil {
		return client.IgnoreNotFound(err)
	}
	if sec.Labels[ManagedByLabel] != ManagedByValue {
		return nil
	}
	return client.IgnoreNotFound(c.Delete(ctx, sec))
}

func sharedEntranceTLSName(viewer string) string {
	return sharedEntranceTLSPrefix + viewer
}

func viewerFromUserSpaceNS(ns string) (string, bool) {
	if !strings.HasPrefix(ns, userSpacePrefix) {
		return "", false
	}
	viewer := strings.TrimPrefix(ns, userSpacePrefix)
	if viewer == "" || !dns1123Label.MatchString(viewer) {
		return "", false
	}
	return viewer, true
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
