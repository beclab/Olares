package appstate

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/testutil"

	"k8s.io/client-go/rest"
)

// TestMain pins the package-level newHelmOps / getKubeConfig seams to
// fail-fast stubs for the entire test binary. Production defaults call
// ctrl.GetConfig + helm.InitConfig, both of which silently fall back to
// the developer's ~/.kube/config and then BLOCK on real cluster traffic
// (helm action.Init). That makes any test path that walks through
// install_failure_cleanup.go (e.g. installing_app.go's toInstallFailed
// after a validation rejection) flaky and environment-sensitive: present-
// but-unreachable kubeconfig hangs for the test's whole timeout instead
// of failing fast and letting cleanup move on. Tests that need real
// fakes call injectHelmOps to override these on top.
func TestMain(m *testing.M) {
	newHelmOps = func(ctx context.Context, kubeConfig *rest.Config, app *appcfg.ApplicationConfig, token string, options appinstaller.Opt) (appinstaller.HelmOpsInterface, error) {
		return nil, errors.New("appstate test: newHelmOps not injected")
	}
	getKubeConfig = func() (*rest.Config, error) {
		return nil, errors.New("appstate test: getKubeConfig not injected")
	}
	os.Exit(m.Run())
}

// injectHelmOps overrides the package-level newHelmOps/getKubeConfig seams to
// return the supplied fake, restoring them when the test ends. Tests using it
// must not run in parallel.
func injectHelmOps(t *testing.T, f *testutil.FakeHelmOps) {
	t.Helper()
	origNew := newHelmOps
	origCfg := getKubeConfig
	newHelmOps = func(ctx context.Context, kubeConfig *rest.Config, app *appcfg.ApplicationConfig, token string, options appinstaller.Opt) (appinstaller.HelmOpsInterface, error) {
		return f, nil
	}
	getKubeConfig = func() (*rest.Config, error) { return &rest.Config{}, nil }
	t.Cleanup(func() {
		newHelmOps = origNew
		getKubeConfig = origCfg
	})
}

// injectHelmOpsError makes the newHelmOps seam fail to construct.
func injectHelmOpsError(t *testing.T, err error) {
	t.Helper()
	origNew := newHelmOps
	origCfg := getKubeConfig
	newHelmOps = func(ctx context.Context, kubeConfig *rest.Config, app *appcfg.ApplicationConfig, token string, options appinstaller.Opt) (appinstaller.HelmOpsInterface, error) {
		return nil, err
	}
	getKubeConfig = func() (*rest.Config, error) { return &rest.Config{}, nil }
	t.Cleanup(func() {
		newHelmOps = origNew
		getKubeConfig = origCfg
	})
}

// fakeDeps returns a Deps value wired to the supplied FakeHelmOps and a
// stub KubeConfig. It is the per-instance counterpart to injectHelmOps:
// methods that moved off the package-level newHelmOps/getKubeConfig seams
// onto Deps (forceDeleteApp, scaleOrPatchSuspend/Resume, ...) read
// p.deps.NewHelmOps / p.deps.KubeConfig, so unit tests on those methods
// must populate baseStatefulApp.deps explicitly.
func fakeDeps(f *testutil.FakeHelmOps) Deps {
	return Deps{
		KubeConfig: func() (*rest.Config, error) { return &rest.Config{}, nil },
		NewHelmOps: func(ctx context.Context, kubeConfig *rest.Config, app *appcfg.ApplicationConfig, token string, options appinstaller.Opt) (appinstaller.HelmOpsInterface, error) {
			return f, nil
		},
	}
}
