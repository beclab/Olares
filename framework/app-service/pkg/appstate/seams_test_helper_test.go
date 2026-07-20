package appstate

import (
	"context"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/testutil"

	"k8s.io/client-go/rest"
)

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
