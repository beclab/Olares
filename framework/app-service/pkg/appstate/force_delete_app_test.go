package appstate

import (
	"context"
	"errors"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

func TestForceDeleteAppSystemEmptyConfig(t *testing.T) {
	am := testutil.NewAppManager("nginx",
		testutil.WithSource("system"),
		testutil.WithConfigJSON(""),
		testutil.WithState(appsv1.Uninstalling),
	)
	c := testutil.NewFakeClient(am)
	b := &baseStatefulApp{manager: am, client: c}

	if err := b.forceDeleteApp(context.TODO()); err != nil {
		t.Fatalf("forceDeleteApp: %v", err)
	}
	got := getAM(t, b, "nginx")
	if got.Status.State != appsv1.Uninstalled {
		t.Errorf("state=%q want Uninstalled", got.Status.State)
	}
}

func TestForceDeleteAppNormalUninstalls(t *testing.T) {
	cfg := &appcfg.ApplicationConfig{AppName: "nginx", Namespace: "nginx-alice", OwnerName: "alice"}
	am := testutil.NewAppManager("nginx",
		testutil.WithNamespace("nginx-alice"),
		testutil.WithConfig(t, cfg),
		testutil.WithState(appsv1.Uninstalling),
	)
	c := testutil.NewFakeClient(am)
	b := &baseStatefulApp{manager: am, client: c}

	fake := testutil.NewFakeHelmOps()
	injectHelmOps(t, fake)

	if err := b.forceDeleteApp(context.TODO()); err != nil {
		t.Fatalf("forceDeleteApp: %v", err)
	}
	if fake.CallCount("Uninstall") != 1 {
		t.Errorf("Uninstall called %d times, want 1", fake.CallCount("Uninstall"))
	}
	if got := getAM(t, b, "nginx"); got.Status.State != appsv1.Uninstalled {
		t.Errorf("state=%q want Uninstalled", got.Status.State)
	}
}

func TestForceDeleteAppToleratesNotFound(t *testing.T) {
	cfg := &appcfg.ApplicationConfig{AppName: "nginx", Namespace: "nginx-alice", OwnerName: "alice"}
	am := testutil.NewAppManager("nginx",
		testutil.WithNamespace("nginx-alice"),
		testutil.WithConfig(t, cfg),
		testutil.WithState(appsv1.Uninstalling),
	)
	c := testutil.NewFakeClient(am)
	b := &baseStatefulApp{manager: am, client: c}

	fake := testutil.NewFakeHelmOps()
	fake.UninstallErr = errors.New("release: not found")
	injectHelmOps(t, fake)

	if err := b.forceDeleteApp(context.TODO()); err != nil {
		t.Fatalf("forceDeleteApp should tolerate not-found: %v", err)
	}
	if got := getAM(t, b, "nginx"); got.Status.State != appsv1.Uninstalled {
		t.Errorf("state=%q want Uninstalled", got.Status.State)
	}
}

func TestForceDeleteAppPropagatesHelmError(t *testing.T) {
	cfg := &appcfg.ApplicationConfig{AppName: "nginx", Namespace: "nginx-alice", OwnerName: "alice"}
	am := testutil.NewAppManager("nginx",
		testutil.WithNamespace("nginx-alice"),
		testutil.WithConfig(t, cfg),
		testutil.WithState(appsv1.Uninstalling),
	)
	c := testutil.NewFakeClient(am)
	b := &baseStatefulApp{manager: am, client: c}

	fake := testutil.NewFakeHelmOps()
	fake.UninstallErr = errors.New("helm boom")
	injectHelmOps(t, fake)

	if err := b.forceDeleteApp(context.TODO()); err == nil {
		t.Fatal("forceDeleteApp should propagate non-not-found helm error")
	}
	if got := getAM(t, b, "nginx"); got.Status.State == appsv1.Uninstalled {
		t.Error("state should not be Uninstalled when uninstall failed")
	}
}
