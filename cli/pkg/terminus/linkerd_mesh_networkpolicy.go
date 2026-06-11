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

// appGatewayMeshNPName matches app-service security.AppGatewayMeshNPName (long-term mesh ingress).
const appGatewayMeshNPName = "app-gateway-mesh-np"

// applyLinkerdMeshNetworkPolicies applies a bootstrap NetworkPolicy until app-service reconciles
// app-gateway-mesh-np. security-controller deletes the bootstrap NP once the managed policy exists.
func applyLinkerdMeshNetworkPolicies(ctx context.Context, c client.Client, settings *cli.EnvSettings, vendorDir string) error {
	reconciled, err := linkerdMeshNPReconciledByAppService(ctx, c)
	if err != nil {
		return err
	}
	if reconciled {
		logger.Info("app-gateway-mesh-np already present; skip bootstrap linkerd mesh NetworkPolicy apply")
		return nil
	}
	path := linkerdMeshNetworkPolicyManifest(vendorDir)
	if path == "" {
		logger.Info("linkerd mesh NetworkPolicy manifest not found in vendor; skip bootstrap apply")
		return nil
	}
	logger.InfoInstallationProgress("Applying Linkerd mesh bootstrap NetworkPolicy (until app-service reconciles app-gateway-mesh-np) ...")
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

func linkerdMeshNetworkPolicyManifest(vendorDir string) string {
	for _, dir := range appGatewayVendorDirCandidates(vendorDir) {
		p := filepath.Join(dir, "network-policies", "linkerd-mesh-ingress.yaml")
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}
