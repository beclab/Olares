package routecontrol

import (
	"context"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/cluster"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestSyncCustomDomainTLS_createsSecret(t *testing.T) {
	s := testScheme(t)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "shop-example-com-domain-ssl-config",
			Namespace: "user-space-alice",
			Labels:    map[string]string{customDomainCertLabel: customDomainCertLabelValue},
		},
		Data: map[string]string{
			"zone": "shop.example.com",
			"cert": "CERT",
			"key":  "KEY",
		},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(cm).Build()
	if err := syncCustomDomainTLS(context.Background(), c, cm); err != nil {
		t.Fatal(err)
	}
	name := customDomainTLSName(cm.Name)
	sec := &corev1.Secret{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: name}, sec); err != nil {
		t.Fatal(err)
	}
	if sec.Labels[labelTLSCustomDomain] != "shop.example.com" {
		t.Errorf("domain label = %q", sec.Labels[labelTLSCustomDomain])
	}
}

func TestSyncCustomDomainTLS_incompleteGC(t *testing.T) {
	s := testScheme(t)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "shop-example-com-domain-ssl-config",
			Namespace: "user-space-alice",
			Labels:    map[string]string{customDomainCertLabel: customDomainCertLabelValue},
		},
		Data: map[string]string{"zone": "shop.example.com"},
	}
	name := customDomainTLSName(cm.Name)
	sec := desiredCustomDomainTLSSecret(name, "shop.example.com", cm.Namespace, "C", "K", "h")
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(cm, sec).Build()
	if err := syncCustomDomainTLS(context.Background(), c, cm); err != nil {
		t.Fatal(err)
	}
	err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: name}, &corev1.Secret{})
	if !apierrors.IsNotFound(err) {
		t.Fatalf("expected secret gc, got %v", err)
	}
}

func TestSyncCustomDomainTLS_bflConfigMapName(t *testing.T) {
	s := testScheme(t)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "shop.example.com-ssl-config",
			Namespace: "user-space-alice",
			Labels:    map[string]string{customDomainCertLabel: customDomainCertLabelValue},
		},
		Data: map[string]string{
			"zone": "shop.example.com",
			"cert": "CERT",
			"key":  "KEY",
		},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(cm).Build()
	if err := syncCustomDomainTLS(context.Background(), c, cm); err != nil {
		t.Fatal(err)
	}
	name := customDomainTLSName(cm.Name)
	sec := &corev1.Secret{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: name}, sec); err != nil {
		t.Fatal(err)
	}
	if sec.Labels[labelTLSCustomDomain] != "shop.example.com" {
		t.Errorf("domain label = %q", sec.Labels[labelTLSCustomDomain])
	}
}

func TestSyncCustomDomainTLS_idempotentHash(t *testing.T) {
	s := testScheme(t)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "shop-example-com-domain-ssl-config",
			Namespace: "user-space-alice",
			Labels:    map[string]string{customDomainCertLabel: customDomainCertLabelValue},
		},
		Data: map[string]string{
			"zone": "shop.example.com",
			"cert": "CERT",
			"key":  "KEY",
		},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(cm).Build()
	if err := syncCustomDomainTLS(context.Background(), c, cm); err != nil {
		t.Fatal(err)
	}
	name := customDomainTLSName(cm.Name)
	sec := &corev1.Secret{}
	key := types.NamespacedName{Namespace: defaultGatewayNS, Name: name}
	if err := c.Get(context.Background(), key, sec); err != nil {
		t.Fatal(err)
	}
	firstHash := sec.Annotations[annotationTLSContentHash]
	if err := syncCustomDomainTLS(context.Background(), c, cm); err != nil {
		t.Fatal(err)
	}
	if err := c.Get(context.Background(), key, sec); err != nil {
		t.Fatal(err)
	}
	if sec.Annotations[annotationTLSContentHash] != firstHash {
		t.Error("expected content hash unchanged on idempotent sync")
	}
}

func TestCustomDomainTLSReconciler_Reconcile_cmDeleted(t *testing.T) {
	s := testScheme(t)
	cmName := "shop-example-com-domain-ssl-config"
	sec := desiredCustomDomainTLSSecret(customDomainTLSName(cmName), "shop.example.com", "user-space-alice", "C", "K", "h")
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(sec).Build()
	r := &CustomDomainTLSReconciler{Client: c}
	if _, err := r.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Namespace: "user-space-alice", Name: cmName},
	}); err != nil {
		t.Fatal(err)
	}
	err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: sec.Name}, &corev1.Secret{})
	if !apierrors.IsNotFound(err) {
		t.Fatalf("secret should be deleted when CM missing: %v", err)
	}
}

func TestCustomDomainTLSReconciler_Reconcile_labelRemoved(t *testing.T) {
	s := testScheme(t)
	cmName := "shop-example-com-domain-ssl-config"
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: cmName, Namespace: "user-space-alice"},
		Data:       map[string]string{"zone": "shop.example.com", "cert": "C", "key": "K"},
	}
	sec := desiredCustomDomainTLSSecret(customDomainTLSName(cmName), "shop.example.com", "user-space-alice", "C", "K", "h")
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(cm, sec).Build()
	r := &CustomDomainTLSReconciler{Client: c}
	if _, err := r.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Namespace: "user-space-alice", Name: cmName},
	}); err != nil {
		t.Fatal(err)
	}
	err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: sec.Name}, &corev1.Secret{})
	if !apierrors.IsNotFound(err) {
		t.Fatalf("secret should be gc when label removed: %v", err)
	}
}

func TestCustomDomainTLSReconciler_Reconcile_nonUserSpaceNS(t *testing.T) {
	s := testScheme(t)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "shop-example-com-domain-ssl-config",
			Namespace: "kube-system",
			Labels:    map[string]string{customDomainCertLabel: customDomainCertLabelValue},
		},
		Data: map[string]string{"zone": "shop.example.com", "cert": "C", "key": "K"},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(cm).Build()
	r := &CustomDomainTLSReconciler{Client: c}
	if _, err := r.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Namespace: "kube-system", Name: cm.Name},
	}); err != nil {
		t.Fatal(err)
	}
	var secList corev1.SecretList
	if err := c.List(context.Background(), &secList); err != nil {
		t.Fatal(err)
	}
	if len(secList.Items) != 0 {
		t.Fatalf("expected no secrets, got %d", len(secList.Items))
	}
}

func TestDeleteCustomDomainTLSSecret_skipsUnmanaged(t *testing.T) {
	s := testScheme(t)
	name := customDomainTLSName("shop-example-com-domain-ssl-config")
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: defaultGatewayNS},
		Type:       corev1.SecretTypeTLS,
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(sec).Build()
	if err := deleteCustomDomainTLSSecret(context.Background(), c, name); err != nil {
		t.Fatal(err)
	}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: name}, &corev1.Secret{}); err != nil {
		t.Fatalf("unmanaged secret should not be deleted: %v", err)
	}
}

func TestApplyGatewayHTTPSListeners_customDomainListener(t *testing.T) {
	cluster.PrimePlatformDomainForTest("olares.com")
	t.Cleanup(cluster.ResetPlatformDomainForTest)
	s := gatewayScheme(t)
	customSec := desiredCustomDomainTLSSecret("shared-entrance-tls-custom-shop", "shop.example.com", "user-space-alice", "C", "K", "h")
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(baseGateway(), customSec).Build()
	custom, err := listCustomDomainTLSSecrets(context.Background(), c)
	if err != nil || len(custom) != 1 {
		t.Fatalf("list custom: %v len=%d", err, len(custom))
	}
	if err := applyGatewayHTTPSListeners(context.Background(), c, nil, custom, "olares.com"); err != nil {
		t.Fatal(err)
	}
	updated := &unstructured.Unstructured{}
	updated.SetGroupVersionKind(gatewayGVK)
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: defaultGatewayNS, Name: defaultGatewayName}, updated); err != nil {
		t.Fatal(err)
	}
	l := findListener(updated, "https-custom-shop-example-com")
	if l == nil {
		t.Fatal("custom listener missing")
	}
	host, _, _ := unstructured.NestedString(l, "hostname")
	if host != "shop.example.com" {
		t.Errorf("hostname = %q", host)
	}
}
