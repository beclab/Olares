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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	zoneSSLConfigMapName      = "zone-ssl-config"
	labelGatewayRouteControl  = "gateway.olares.io/managed-by"
	gatewayRouteControlValue  = "routecontrol"
	labelTLSViewer            = "gateway.olares.io/tls-viewer"
	annotationTLSContentHash  = "gateway.olares.io/tls-content-hash"
	entranceTLSSecretPrefix   = "shared-entrance-tls-"
	userSpaceNamespacePrefix  = "user-space-"
)

// EntranceTLSReconciler syncs per-viewer TLS material from zone-ssl-config into
// the app-gateway namespace for in-cluster HTTPS termination.
type EntranceTLSReconciler struct {
	Client client.Client
}

// ReconcileConfigMap applies one user-space zone-ssl-config to the gateway Secret.
func (r *EntranceTLSReconciler) ReconcileConfigMap(ctx context.Context, cm *corev1.ConfigMap) error {
	if r == nil || r.Client == nil {
		return nil
	}
	_, err := reconcileEntranceTLS(ctx, r.Client, cm)
	return err
}

// reconcileEntranceTLS copies cert/key from zone-ssl-config into
// app-gateway/shared-entrance-tls-<viewer>. Returns true when a Secret write occurred.
//
// requirement: in-cluster HTTPS must reuse the edge wildcard cert from zone-ssl-config.
// behavior: idempotent Secret upsert keyed on content hash; skips incomplete CM data.
func reconcileEntranceTLS(ctx context.Context, c client.Client, cm *corev1.ConfigMap) (bool, error) {
	if cm == nil || cm.Name != zoneSSLConfigMapName {
		return false, nil
	}
	viewer, ok := viewerFromUserSpaceNamespace(cm.Namespace)
	if !ok {
		return false, nil
	}
	cert := strings.TrimSpace(cm.Data["cert"])
	key := strings.TrimSpace(cm.Data["key"])
	if cert == "" || key == "" {
		klog.Warningf("zone-ssl-config %s/%s missing cert or key, skip entrance TLS sync",
			cm.Namespace, cm.Name)
		return false, nil
	}

	desired := desiredEntranceTLSSecret(viewer, cert, key)
	hash := tlsMaterialHash(cert, key)
	desired.Annotations[annotationTLSContentHash] = hash

	current := &corev1.Secret{}
	err := c.Get(ctx, types.NamespacedName{
		Namespace: defaultGatewayNS,
		Name:      entranceTLSSecretName(viewer),
	}, current)
	switch {
	case apierrors.IsNotFound(err):
		return true, c.Create(ctx, desired)
	case err != nil:
		return false, err
	}
	if current.Annotations != nil && current.Annotations[annotationTLSContentHash] == hash {
		return false, nil
	}
	current.Type = corev1.SecretTypeTLS
	if current.Labels == nil {
		current.Labels = map[string]string{}
	}
	for k, v := range desired.Labels {
		current.Labels[k] = v
	}
	if current.Annotations == nil {
		current.Annotations = map[string]string{}
	}
	current.Annotations[annotationTLSContentHash] = hash
	current.Data = nil
	current.StringData = map[string]string{
		corev1.TLSCertKey:       cert,
		corev1.TLSPrivateKeyKey: key,
	}
	return true, c.Update(ctx, current)
}

// deleteEntranceTLSSecret removes the per-viewer TLS Secret when zone-ssl-config is gone.
func deleteEntranceTLSSecret(ctx context.Context, c client.Client, viewer string) error {
	if viewer == "" {
		return nil
	}
	secret := &corev1.Secret{}
	err := c.Get(ctx, types.NamespacedName{
		Namespace: defaultGatewayNS,
		Name:      entranceTLSSecretName(viewer),
	}, secret)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return client.IgnoreNotFound(c.Delete(ctx, secret))
}

func viewerFromUserSpaceNamespace(namespace string) (string, bool) {
	if !strings.HasPrefix(namespace, userSpaceNamespacePrefix) {
		return "", false
	}
	viewer := strings.TrimPrefix(namespace, userSpaceNamespacePrefix)
	viewer = strings.TrimSpace(viewer)
	if viewer == "" {
		return "", false
	}
	return viewer, true
}

func entranceTLSSecretName(viewer string) string {
	return entranceTLSSecretPrefix + viewer
}

func tlsMaterialHash(cert, key string) string {
	sum := sha256.Sum256([]byte(cert + "\n" + key))
	return hex.EncodeToString(sum[:])
}

func desiredEntranceTLSSecret(viewer, cert, key string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      entranceTLSSecretName(viewer),
			Namespace: defaultGatewayNS,
			Labels: map[string]string{
				ManagedByLabel:        ManagedByValue,
				labelGatewayRouteControl: gatewayRouteControlValue,
				labelTLSViewer:        viewer,
			},
			Annotations: map[string]string{},
		},
		Type: corev1.SecretTypeTLS,
		StringData: map[string]string{
			corev1.TLSCertKey:       cert,
			corev1.TLSPrivateKeyKey: key,
		},
	}
}
