package os

import (
	"log"
	"os"
	"path/filepath"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/terminus"
	"github.com/spf13/cobra"
)

func NewCmdInstallAppGatewayStack() *cobra.Command {
	var installerDir string
	var kubeconfig string
	var withAppGatewayChart bool

	cmd := &cobra.Command{
		Use:   "install-app-gateway",
		Short: "Install Linkerd + Envoy Gateway + app-gateway (helm SDK)",
		Run: func(cmd *cobra.Command, args []string) {
			if installerDir == "" {
				_ = cmd.Usage()
				log.Fatal("error: --installer-dir is required")
			}

			_ = os.Setenv("OLARES_INSTALLER_DIR", filepath.Clean(installerDir))
			_ = os.Setenv("APP_GATEWAY_STACK_ENABLED", "1")
			if kubeconfig != "" {
				_ = os.Setenv("KUBECONFIG", kubeconfig)
			}

			runtime := &common.KubeRuntime{Arg: &common.Argument{}}
			kubeAction := common.KubeAction{KubeConf: &common.KubeConf{Arg: runtime.Arg}}

			if err := (&terminus.InstallAppGatewayVendor{KubeAction: kubeAction}).Execute(runtime); err != nil {
				log.Fatalf("error: install vendor: %v", err)
			}
			if err := (&terminus.WaitAppGatewayReady{KubeAction: kubeAction}).Execute(runtime); err != nil {
				log.Fatalf("error: wait ready: %v", err)
			}
			if withAppGatewayChart {
				if err := (&terminus.InstallAppGatewayChart{KubeAction: kubeAction}).Execute(runtime); err != nil {
					log.Fatalf("error: install app-gateway chart: %v", err)
				}
			}
		},
	}

	cmd.Flags().StringVar(&installerDir, "installer-dir", "", "Olares installer dist directory (e.g. .dist)")
	cmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "kubeconfig path (optional)")
	cmd.Flags().BoolVar(&withAppGatewayChart, "with-app-gateway-chart", true, "Install app-gateway Helm chart")

	return cmd
}
