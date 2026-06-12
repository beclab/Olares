package routecontrol

import (
	"context"
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
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	customDomainCertLabel      = "app.bytetrade.io/custom-domain-cert"
	customDomainCertLabelValue = "true"
	annotationTLSHostname      = "gateway.olares.io/tls-hostname"
)

var customDomainNameSanitizer = regexp.MustCompile(`[^a-z0-9-]+`)

// CustomDomainTLSReconciler syncs third-party custom-domain SSL ConfigMaps into
// gateway TLS Secrets. Platform URLs use EntranceTLSReconciler on zone-ssl-config
// instead; custom domains are identified by label app.bytetrade.io/custom-domain-cert.
type CustomDomainTLSReconciler struct {
	Client client.Client
}

// Reconcile copies custom domain cert material into a dedicated TLS Secret.
func (r *CustomDomainTLSReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	if r == nil || r.Client == nil {
		return reconcile.Result{}, nil
	}
	if _, ok := viewerFromUserSpaceNS(req.Namespace); !ok {
		return reconcile.Result{}, nil
	}

	cm := &corev1.ConfigMap{}
	if err := r.Client.Get(ctx, req.NamespacedName, cm); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, deleteCustomDomainTLSSecretByCMName(ctx, r.Client, req.Name)
		}
		return reconcile.Result{}, err
	}
	if cm.Labels[customDomainCertLabel] != customDomainCertLabelValue {
		return reconcile.Result{}, deleteCustomDomainTLSSecretByCMName(ctx, r.Client, cm.Name)
	}
	return reconcile.Result{}, syncCustomDomainTLS(ctx, r.Client, cm)
}

func syncCustomDomainTLS(ctx context.Context, c client.Client, cm *corev1.ConfigMap) error {
	domain := strings.TrimSpace(cm.Data["zone"])
	cert := strings.TrimSpace(cm.Data["cert"])
	key := strings.TrimSpace(cm.Data["key"])
	secretName := customDomainTLSName(cm.Name)
	if domain == "" || cert == "" || key == "" {
		klog.Warningf("custom domain CM %s/%s incomplete, gc secret", cm.Namespace, cm.Name)
		return deleteCustomDomainTLSSecret(ctx, c, secretName)
	}
	hash := tlsMaterialHash(cert, key)
	current := &corev1.Secret{}
	err := c.Get(ctx, types.NamespacedName{Namespace: defaultGatewayNS, Name: secretName}, current)
	switch {
	case apierrors.IsNotFound(err):
		return c.Create(ctx, desiredCustomDomainTLSSecret(secretName, domain, cm.Namespace, cert, key, hash))
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
	current.Labels[labelTLSCustomDomain] = domain
	if current.Annotations == nil {
		current.Annotations = map[string]string{}
	}
	current.Annotations[annotationTLSContentHash] = hash
	current.Annotations[annotationTLSSourceNS] = cm.Namespace
	current.Annotations[annotationTLSHostname] = domain
	current.Data = nil
	current.StringData = map[string]string{
		corev1.TLSCertKey:       cert,
		corev1.TLSPrivateKeyKey: key,
	}
	return c.Update(ctx, current)
}

func desiredCustomDomainTLSSecret(name, domain, sourceNS, cert, key, hash string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: defaultGatewayNS,
			Labels: map[string]string{
				ManagedByLabel:       ManagedByValue,
				labelTLSCustomDomain:   domain,
			},
			Annotations: map[string]string{
				annotationTLSContentHash: hash,
				annotationTLSSourceNS:  sourceNS,
				annotationTLSHostname:    domain,
			},
		},
		Type: corev1.SecretTypeTLS,
		StringData: map[string]string{
			corev1.TLSCertKey:       cert,
			corev1.TLSPrivateKeyKey: key,
		},
	}
}

func customDomainTLSName(cmName string) string {
	safe := strings.Trim(customDomainNameSanitizer.ReplaceAllString(strings.ToLower(cmName), "-"), "-")
	if safe == "" {
		safe = "domain"
	}
	return customDomainTLSPrefix + safe
}

func deleteCustomDomainTLSSecret(ctx context.Context, c client.Client, name string) error {
	sec := &corev1.Secret{}
	if err := c.Get(ctx, types.NamespacedName{Namespace: defaultGatewayNS, Name: name}, sec); err != nil {
		return client.IgnoreNotFound(err)
	}
	if sec.Labels[ManagedByLabel] != ManagedByValue {
		return nil
	}
	return client.IgnoreNotFound(c.Delete(ctx, sec))
}

func deleteCustomDomainTLSSecretByCMName(ctx context.Context, c client.Client, cmName string) error {
	return deleteCustomDomainTLSSecret(ctx, c, customDomainTLSName(cmName))
}

// SetupWithManager registers the reconciler for custom-domain cert ConfigMaps.
func (r *CustomDomainTLSReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r == nil {
		return nil
	}
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named("entrance-tls-custom").
		For(&corev1.ConfigMap{}, builder.WithPredicates(customDomainCertConfigMapPredicate())).
		Complete(r)
}

func customDomainCertConfigMapPredicate() predicate.Predicate {
	hasLabel := func(o client.Object) bool {
		if o == nil {
			return false
		}
		return o.GetLabels()[customDomainCertLabel] == customDomainCertLabelValue
	}
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool { return hasLabel(e.Object) },
		UpdateFunc: func(e event.UpdateEvent) bool {
			return hasLabel(e.ObjectNew) || hasLabel(e.ObjectOld)
		},
		DeleteFunc: func(e event.DeleteEvent) bool { return hasLabel(e.Object) },
	}
}
