package terminus

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/core/logger"
	appsv1 "k8s.io/api/apps/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "olares-cli-terminus-test-log-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	logger.InitLog(dir, filepath.Join(dir, "console.log"), true)
	os.Exit(m.Run())
}

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
	err := waitLinkerdControlPlaneReadyWithPoll(ctx, c, "linkerd", 20*time.Millisecond, 5*time.Millisecond)
	if err == nil {
		t.Fatal("expected error")
	}
	got := err.Error()
	if !strings.Contains(got, "WaitLinkerdControlPlaneReady") || !strings.Contains(got, "destination") {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, linkerdPKIGuardianDeployment) {
		t.Fatalf("expected guardian in timeout error, got: %v", err)
	}
}

func TestLinkerdControlPlaneNotReady_guardianNotDeployed(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	ctx := context.Background()

	pending, err := linkerdControlPlaneNotReady(ctx, c, "linkerd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, item := range pending {
		if strings.Contains(item, linkerdPKIGuardianDeployment) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected guardian pending when not deployed, got %v", pending)
	}
}

func TestWaitLinkerdControlPlaneReady_guardianDelayedThenReady(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const ns = "linkerd"
	for _, name := range linkerdControlPlaneDeployments {
		dep := readyDeployment(name, ns)
		if err := c.Create(ctx, &dep); err != nil {
			t.Fatalf("create deployment %s: %v", name, err)
		}
	}

	done := make(chan error, 1)
	go func() {
		done <- waitLinkerdControlPlaneReadyWithPoll(ctx, c, ns, 2*time.Second, 20*time.Millisecond)
	}()

	time.Sleep(30 * time.Millisecond)
	guardian := readyDeployment(linkerdPKIGuardianDeployment, ns)
	if err := c.Create(ctx, &guardian); err != nil {
		t.Fatalf("create guardian deployment: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected success after guardian created, got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for waitLinkerdControlPlaneReady")
	}
}

func readyDeployment(name, ns string) appsv1.Deployment {
	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas: 1,
		},
	}
}
