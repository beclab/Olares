package os

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/beclab/Olares/cli/pkg/terminus"
	agwconfig "github.com/beclab/Olares/framework/app-gateway/pkg/config"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewCmdMaintainLinkerdPKI() *cobra.Command {
	var installerDir string
	cmd := &cobra.Command{
		Use:   "maintain-linkerd-pki",
		Short: "Check Linkerd identity issuer validity and rotate if under 6 months remain",
		RunE: func(cmd *cobra.Command, args []string) error {
			if installerDir == "" {
				installerDir = os.Getenv("OLARES_INSTALLER_DIR")
			}
			if installerDir != "" {
				initInstallAppGatewayLogger(installerDir)
			}
			config, err := ctrl.GetConfig()
			if err != nil {
				return err
			}
			c, err := client.New(config, client.Options{})
			if err != nil {
				return err
			}
			vendor := resolveVendorDirForLinkerdPKI(installerDir)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			if err := terminus.MaintainLinkerdPKI(ctx, c, agwconfig.LinkerdNamespace(), vendor); err != nil {
				return err
			}
			logger.Info("maintain-linkerd-pki completed")
			return nil
		},
	}
	cmd.Flags().StringVar(&installerDir, "installer-dir", "", "Olares installer directory (wizard/config/app-gateway-vendor)")
	return cmd
}

func resolveVendorDirForLinkerdPKI(installerDir string) string {
	if installerDir != "" {
		return filepath.Join(installerDir, "wizard", "config", "app-gateway-vendor")
	}
	if root := os.Getenv("OLARES_SOURCE_ROOT"); root != "" {
		return filepath.Join(root, "framework", "app-gateway")
	}
	return ""
}
