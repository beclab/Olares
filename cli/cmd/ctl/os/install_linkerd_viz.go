package os

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/beclab/Olares/cli/pkg/terminus"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/cli"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
			vendor := resolveVendorDirForLinkerdViz(installerDir)
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
	cmd.Flags().StringVar(&installerDir, "installer-dir", "", "Olares installer dir (wizard/config/app-gateway-vendor)")
	cmd.Flags().StringVar(&prometheusURL, "prometheus-url", "", "Platform Prometheus base URL (default: KubeSphere prometheus-k8s)")
	return cmd
}

func resolveVendorDirForLinkerdViz(installerDir string) string {
	if installerDir != "" {
		return filepath.Join(installerDir, "wizard", "config", "app-gateway-vendor")
	}
	if root := os.Getenv("OLARES_SOURCE_ROOT"); root != "" {
		return filepath.Join(root, "framework", "app-gateway", ".olares", "config", "app-gateway-vendor")
	}
	return ""
}
