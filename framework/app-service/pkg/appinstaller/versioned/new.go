package versioned

import (
	"context"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	v2 "github.com/beclab/Olares/framework/app-service/pkg/appinstaller/v2"
	"k8s.io/client-go/rest"
)

func NewHelmOps(ctx context.Context, kubeConfig *rest.Config, app *appcfg.ApplicationConfig, token string, options appinstaller.Opt) (ops appinstaller.HelmOpsInterface, err error) {
	if app.APIVersion == appcfg.V2 {
		ops, err = v2.NewHelmOps(ctx, kubeConfig, app, token, options)
	} else {
		ops, err = appinstaller.NewHelmOps(ctx, kubeConfig, app, token, options)
	}

	return ops, err
}
