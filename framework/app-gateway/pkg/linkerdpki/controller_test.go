package linkerdpki

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-gateway/pkg/entrypki"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestColdStartRenewsBothExpiredEntryCertificatesWhileLinkerdIsBlocked(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	linkerdStarted := make(chan struct{})
	var signalLinkerdStart sync.Once
	var clientUpdates atomic.Int32
	c := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
		Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			if key.Namespace == "os-mesh" && key.Name == PKISecretName {
				signalLinkerdStart.Do(func() { close(linkerdStarted) })
				<-ctx.Done()
				return ctx.Err()
			}
			return c.Get(ctx, key, obj, opts...)
		},
		Update: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
			if obj.GetName() == entrypki.ClientSecretName && clientUpdates.Add(1) == 1 {
				return errors.New("temporary Entry Secret write failure")
			}
			return c.Update(ctx, obj, opts...)
		},
	}).Build()
	issuedAt := time.Now().UTC().Add(-366 * 24 * time.Hour)
	if err := entrypki.MaintainEntryMTLS(ctx, c, entrypki.DefaultNamespace, true, issuedAt); err != nil {
		t.Fatal(err)
	}
	get := func(ctx context.Context, name string) (*corev1.Secret, error) {
		var secret corev1.Secret
		err := c.Get(ctx, types.NamespacedName{Namespace: entrypki.DefaultNamespace, Name: name}, &secret)
		return &secret, err
	}
	leaves := []struct {
		name, certKey, privateKey, dns string
		usage                          x509.ExtKeyUsage
	}{
		{entrypki.SecretName, "tls.crt", "tls.key", "app-gateway-l4-mtls.os-gateway.svc.cluster.local", x509.ExtKeyUsageServerAuth},
		{entrypki.ClientSecretName, "client.crt", "client.key", "l4-bfl-proxy.entry.client", x509.ExtKeyUsageClientAuth},
	}
	baseline := make(map[string]*corev1.Secret, len(leaves))
	for _, leaf := range leaves {
		secret, err := get(ctx, leaf.name)
		if err != nil {
			t.Fatal(err)
		}
		pair, err := tls.X509KeyPair(secret.Data[leaf.certKey], secret.Data[leaf.privateKey])
		if err != nil {
			t.Fatal(err)
		}
		if !pair.Leaf.NotAfter.Before(time.Now()) {
			t.Fatalf("%s fixture must already be expired", leaf.name)
		}
		baseline[leaf.name] = secret
	}

	const interval = 24 * time.Hour
	controller := NewController(c, "os-mesh", interval, NewProbeState(interval), NewMetrics()).WithEntryMTLS(true, entrypki.DefaultNamespace)
	done := make(chan struct{})
	go func() {
		defer close(done)
		controller.Run(ctx)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("Guardian did not stop after cancellation")
		}
	})
	select {
	case <-linkerdStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("Linkerd reconcile did not start")
	}

	err := wait.PollUntilContextTimeout(ctx, 25*time.Millisecond, 12*time.Second, true, func(ctx context.Context) (bool, error) {
		for _, leaf := range leaves {
			secret, err := get(ctx, leaf.name)
			if err != nil {
				return false, err
			}
			if bytes.Equal(secret.Data[leaf.certKey], baseline[leaf.name].Data[leaf.certKey]) {
				return false, nil
			}
		}
		return true, nil
	})
	if err != nil {
		t.Fatalf("Entry must recover on retry while Linkerd is blocked: %v", err)
	}
	if got := clientUpdates.Load(); got != 2 {
		t.Fatalf("expected a failed client renewal and one successful retry, got %d attempts", got)
	}
	for _, leaf := range leaves {
		secret, err := get(ctx, leaf.name)
		if err != nil {
			t.Fatal(err)
		}
		ca := baseline[leaf.name].Data["ca.crt"]
		roots := x509.NewCertPool()
		if !bytes.Equal(secret.Data["ca.crt"], ca) || !roots.AppendCertsFromPEM(ca) {
			t.Fatalf("%s changed the trust anchor", leaf.name)
		}
		pair, err := tls.X509KeyPair(secret.Data[leaf.certKey], secret.Data[leaf.privateKey])
		if err != nil {
			t.Fatal(err)
		}
		if _, err := pair.Leaf.Verify(x509.VerifyOptions{Roots: roots, DNSName: leaf.dns, KeyUsages: []x509.ExtKeyUsage{leaf.usage}}); err != nil {
			t.Fatalf("renewed %s does not verify against the original CA: %v", leaf.name, err)
		}
	}
}

func TestEntryFailurePreservesLinkerdReadiness(t *testing.T) {
	const interval = 24 * time.Hour
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	mat, err := testCAAndIssuerMaterial(time.Now().UTC().Add(200 * 24 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	ns := DefaultLinkerdNamespace
	entryFailed := make(chan struct{}, 1)
	c := fake.NewClientBuilder().WithScheme(testScheme(t)).WithObjects(
		testPKISecret(ns, mat),
		testIdentityIssuerSecret(ns, mat, true),
		testIdentityDeployment(ns, issuerFingerprint(mat.IssuerCrt), "already-rolled"),
	).WithInterceptorFuncs(interceptor.Funcs{
		Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			if key.Namespace == entrypki.DefaultNamespace {
				select {
				case entryFailed <- struct{}{}:
				default:
				}
				return errors.New("Entry Secret unavailable")
			}
			return c.Get(ctx, key, obj, opts...)
		},
	}).Build()
	probes := NewProbeState(interval)
	controller := NewController(c, ns, interval, probes, NewMetrics()).WithEntryMTLS(true, entrypki.DefaultNamespace)
	done := make(chan struct{})
	go func() {
		defer close(done)
		controller.Run(ctx)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("Guardian did not stop after cancellation")
		}
	})
	select {
	case <-entryFailed:
	case <-time.After(2 * time.Second):
		t.Fatal("Entry reconcile did not start")
	}
	if err := wait.PollUntilContextTimeout(ctx, 25*time.Millisecond, 2*time.Second, true, func(context.Context) (bool, error) {
		recorder := httptest.NewRecorder()
		probes.ReadyHandler(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		return recorder.Code == http.StatusOK, nil
	}); err != nil {
		t.Fatalf("Linkerd readiness must remain independent of Entry: %v", err)
	}
}
