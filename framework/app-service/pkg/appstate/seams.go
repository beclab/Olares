package appstate

import (
	"context"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller/versioned"

	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

// newHelmOps and getKubeConfig are indirections over the concrete helm-ops
// factory and the controller-runtime kube config getter so tests can inject
// fakes without standing up a real cluster or helm backend.
var (
	newHelmOps = func(ctx context.Context, kubeConfig *rest.Config, app *appcfg.ApplicationConfig, token string, options appinstaller.Opt) (appinstaller.HelmOpsInterface, error) {
		return versioned.NewHelmOps(ctx, kubeConfig, app, token, options)
	}

	getKubeConfig = ctrl.GetConfig
)
