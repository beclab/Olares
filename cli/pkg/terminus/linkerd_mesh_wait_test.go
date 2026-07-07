package terminus

import (
	"context"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestWaitAppGatewayMeshNP_bothPresent(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = netv1.AddToScheme(scheme)
	c := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(
			&netv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: appGatewayMeshNPName, Namespace: "linkerd"}},
			&netv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: appGatewayMeshNPName, Namespace: "os-gateway"}},
		).Build()
	ctx := context.Background()
	if err := waitAppGatewayMeshNP(ctx, c, time.Second, time.Millisecond); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestWaitAppGatewayMeshNP_oneSideMissing(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = netv1.AddToScheme(scheme)
	c := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(
			&netv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: appGatewayMeshNPName, Namespace: "linkerd"}},
		).Build()
	ctx := context.Background()
	err := waitAppGatewayMeshNP(ctx, c, 50*time.Millisecond, 10*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if got := err.Error(); !strings.Contains(got, "WaitAppGatewayMeshNP") || !strings.Contains(got, "os-gateway") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWaitAppGatewayMeshNP_notFoundThenPresent(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = netv1.AddToScheme(scheme)
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- waitAppGatewayMeshNP(ctx, c, 2*time.Second, 20*time.Millisecond)
	}()

	time.Sleep(30 * time.Millisecond)
	_ = c.Create(ctx, &netv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: appGatewayMeshNPName, Namespace: "linkerd"}})
	_ = c.Create(ctx, &netv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: appGatewayMeshNPName, Namespace: "os-gateway"}})

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected success after NP created, got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for waitAppGatewayMeshNP")
	}
}

func TestWaitLinkerdControlPlaneReady_errorMessage(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	ctx := context.Background()
	err := waitLinkerdControlPlaneReady(ctx, c, "linkerd", 20*time.Millisecond)
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); !strings.Contains(got, "WaitLinkerdControlPlaneReady") || !strings.Contains(got, "destination") {
		t.Fatalf("unexpected error: %v", err)
	}
}
