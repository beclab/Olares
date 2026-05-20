package terminus

import (
	"context"

	"github.com/beclab/Olares/cli/pkg/utils"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

// UpgradeEnvoyGatewayHelmWait installs/upgrades the Envoy Gateway control-plane chart with Helm --wait.
// App-gateway vendor install only; implemented in utils/app_gateway_helm.go without changing UpgradeCharts.
func UpgradeEnvoyGatewayHelmWait(ctx context.Context, actionConfig *action.Configuration, settings *cli.EnvSettings,
	appName, chartName, repoURL, namespace string, vals map[string]interface{}, reuseValue bool) error {
	return utils.UpgradeChartsSkipCRDsWait(ctx, actionConfig, settings, appName, chartName, repoURL, namespace, vals, reuseValue)
}
