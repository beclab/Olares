package os

import (
	"context"
	"os"
	"path/filepath"
	"time"

	cc "github.com/beclab/Olares/cli/pkg/core/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/terminus"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/cli"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func initLinkerdVizLogger(installerDir string) {
	logBase := installerDir
	if logBase == "" {
		logBase = os.Getenv("OLARES_INSTALLER_DIR")
	}
	if logBase == "" {
		logBase = filepath.Join(os.TempDir(), "olares-linkerd-viz-install")
	}
	logDir := filepath.Join(logBase, cc.LogsDir, "linkerd-viz-install")
	consoleLog := filepath.Join(logDir, cc.InstallLogFile)
	logger.InitLog(logDir, consoleLog, true)
}

func NewCmdInstallLinkerdViz() *cobra.Command {
	var installerDir string
	var prometheusURL string
	cmd := &cobra.Command{
		Use:   "install-linkerd-viz",
		Short: "Optionally install linkerd-viz (no bundled Prometheus; uses platform Prometheus)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if installerDir == "" {
				installerDir = os.Getenv("OLARES_INSTALLER_DIR")
			}
			if prometheusURL == "" {
				prometheusURL = os.Getenv("OLARES_LINKERD_PROMETHEUS_URL")
			}
			initLinkerdVizLogger(installerDir)
			defer func() { _ = logger.Sync() }()

			vendor := terminus.ResolveAppGatewayVendorDir(installerDir, terminus.LinkerdVizValuesFileName)
			if vendor == "" {
				return terminus.ErrLinkerdVizVendorNotFound
			}

			config, err := ctrl.GetConfig()
			if err != nil {
				return err
			}
			c, err := client.New(config, client.Options{})
			if err != nil {
				return err
			}
			settings := cli.New()
			if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
				settings.KubeConfig = kubeconfig
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
			defer cancel()
			return terminus.InstallLinkerdViz(ctx, c, settings, vendor, prometheusURL)
		},
	}
	cmd.Flags().StringVar(&installerDir, "installer-dir", "", "Installer root (.dist) or app-gateway-vendor directory")
	cmd.Flags().StringVar(&prometheusURL, "prometheus-url", "", "Platform Prometheus base URL (default: KubeSphere prometheus-k8s)")
	return cmd
}
