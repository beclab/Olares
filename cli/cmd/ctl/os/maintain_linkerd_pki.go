package os

import (
	"context"
	"os"
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
			// behavior: MaintainLinkerdPKI only reads the in-cluster olares-linkerd-pki
			// secret; the vendor dir is optional (empty when run from the guardian CronJob,
			// which has no installer on disk) and kept only for log init / compatibility.
			vendor := terminus.ResolveAppGatewayVendorDir(installerDir, "linkerd-values.yaml")
			initInstallAppGatewayLogger(installerDir)
			defer func() { _ = logger.Sync() }()

			config, err := ctrl.GetConfig()
			if err != nil {
				return err
			}
			c, err := client.New(config, client.Options{})
			if err != nil {
				return err
			}
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

