package versioned

import (
	"context"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	v2 "github.com/beclab/Olares/framework/app-service/pkg/appinstaller/v2"
	v3 "github.com/beclab/Olares/framework/app-service/pkg/appinstaller/v3"
	"k8s.io/client-go/rest"
)

func NewHelmOps(ctx context.Context, kubeConfig *rest.Config, app *appcfg.ApplicationConfig, token string, options appinstaller.Opt) (ops appinstaller.HelmOpsInterface, err error) {
	switch app.APIVersion {
	case appcfg.V3:
		ops, err = v3.NewHelmOps(ctx, kubeConfig, app, token, options)
	case appcfg.V2:
		ops, err = v2.NewHelmOps(ctx, kubeConfig, app, token, options)
	default:
		ops, err = appinstaller.NewHelmOps(ctx, kubeConfig, app, token, options)
	}

	return ops, err
}
