package terminus

import (
	"context"
	"os"
	"path/filepath"

	agwconfig "github.com/beclab/Olares/framework/app-gateway/pkg/config"

	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/utils"
	"helm.sh/helm/v3/pkg/cli"
	netv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const appGatewayMeshNPName = "app-gateway-mesh-np"

func applyLinkerdMeshBootstrapNP(ctx context.Context, c client.Client, settings *cli.EnvSettings, deployDir string) error {
	reconciled, err := linkerdMeshNPReconciledByAppService(ctx, c)
	if err != nil {
		return err
	}
	if reconciled {
		logger.Info("app-gateway-mesh-np already present; skip bootstrap NetworkPolicy apply")
		return nil
	}
	path := filepath.Join(deployDir, "app-gateway-mesh-np-bootstrap.yaml")
	if st, err := os.Stat(path); err != nil || st.IsDir() {
		logger.Info("linkerd mesh bootstrap NetworkPolicy manifest not found; skip")
		return nil
	}
	logger.InfoInstallationProgress("Applying Linkerd mesh bootstrap NetworkPolicy ...")
	return utils.KubectlApplyFile(ctx, settings, path)
}

func linkerdMeshNPReconciledByAppService(ctx context.Context, c client.Client) (bool, error) {
	for _, ns := range []string{agwconfig.LinkerdNamespace(), agwconfig.Namespace()} {
		var np netv1.NetworkPolicy
		if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: appGatewayMeshNPName}, &np); err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
	}
	return true, nil
}
