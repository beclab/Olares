package routecontrol

import (
	"context"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSyncMeshInCustomTLSReplica_createsAggregate(t *testing.T) {
	s := testScheme(t)
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "caller-alice", UID: "ns-uid-1"}}
	src := desiredCustomDomainTLSSecret(
		customDomainTLSPrefix+"shop", "shop.example.com",
		"user-space-alice", "CERT-SHOP", "KEY-SHOP", "hash-shop")
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(ns, src).Build()

	if err := syncMeshInCustomTLSReplica(context.Background(), c, "caller-alice",
		[]string{"abcd.alice.olares.com", "shop.example.com"}); err != nil {
		t.Fatal(err)
	}
	got := &corev1.Secret{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "caller-alice", Name: constants.MeshInCustomTLSSecretName,
	}, got); err != nil {
		t.Fatal(err)
	}
	if string(got.Data["shop.example.com.crt"]) != "CERT-SHOP" {
		t.Fatalf("crt = %q", got.Data["shop.example.com.crt"])
	}
	if string(got.Data["shop.example.com.key"]) != "KEY-SHOP" {
		t.Fatalf("key = %q", got.Data["shop.example.com.key"])
	}
	if _, ok := got.Data["abcd.alice.olares.com.crt"]; ok {
		t.Fatal("platform host must not appear in custom aggregate")
	}
	if got.Labels[labelTLSCustomReplica] != "true" {
		t.Fatalf("labels = %#v", got.Labels)
	}
}

func TestSyncMeshInCustomTLSReplica_removesStaleAndDeletesEmpty(t *testing.T) {
	s := testScheme(t)
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "caller-alice", UID: "ns-uid-1"}}
	src := desiredCustomDomainTLSSecret(
		customDomainTLSPrefix+"shop", "shop.example.com",
		"user-space-alice", "CERT-SHOP", "KEY-SHOP", "hash-shop")
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(ns, src).Build()
	if err := syncMeshInCustomTLSReplica(context.Background(), c, "caller-alice",
		[]string{"shop.example.com"}); err != nil {
		t.Fatal(err)
	}
	if err := syncMeshInCustomTLSReplica(context.Background(), c, "caller-alice", nil); err != nil {
		t.Fatal(err)
	}
	err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "caller-alice", Name: constants.MeshInCustomTLSSecretName,
	}, &corev1.Secret{})
	if !apierrors.IsNotFound(err) {
		t.Fatalf("expected delete, got %v", err)
	}
}

func TestSyncMeshInCustomTLSReplica_noopSameHash(t *testing.T) {
	s := testScheme(t)
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "caller-alice", UID: "ns-uid-1"}}
	src := desiredCustomDomainTLSSecret(
		customDomainTLSPrefix+"shop", "shop.example.com",
		"user-space-alice", "CERT-SHOP", "KEY-SHOP", "hash-shop")
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(ns, src).Build()
	if err := syncMeshInCustomTLSReplica(context.Background(), c, "caller-alice",
		[]string{"shop.example.com"}); err != nil {
		t.Fatal(err)
	}
	first := &corev1.Secret{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "caller-alice", Name: constants.MeshInCustomTLSSecretName,
	}, first); err != nil {
		t.Fatal(err)
	}
	rv := first.ResourceVersion
	if err := syncMeshInCustomTLSReplica(context.Background(), c, "caller-alice",
		[]string{"shop.example.com"}); err != nil {
		t.Fatal(err)
	}
	second := &corev1.Secret{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "caller-alice", Name: constants.MeshInCustomTLSSecretName,
	}, second); err != nil {
		t.Fatal(err)
	}
	if second.ResourceVersion != rv {
		t.Fatalf("noop should not bump resourceVersion: %s -> %s", rv, second.ResourceVersion)
	}
}

func TestCollectTLSHosts(t *testing.T) {
	got := collectTLSHosts([]SharedHostsTarget{
		{TLSHosts: []string{"a.example.com", "b.alice.olares.com"}},
		{TLSHosts: []string{"a.example.com"}},
	})
	if len(got) != 2 || got[0] != "a.example.com" || got[1] != "b.alice.olares.com" {
		t.Fatalf("got %#v", got)
	}
}
