package os

import (
	"log"
	"os"
	"path/filepath"

	"github.com/beclab/Olares/cli/pkg/common"
	cc "github.com/beclab/Olares/cli/pkg/core/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/terminus"
	"github.com/spf13/cobra"
)

// initInstallAppGatewayLogger matches connector.BaseRuntime logging used during full install.
func initInstallAppGatewayLogger(installerDir string) {
	logDir := filepath.Join(installerDir, "logs", "app-gateway-install")
	consoleLog := filepath.Join(logDir, cc.InstallLogFile)
	logger.InitLog(logDir, consoleLog, true)
}

func NewCmdInstallAppGatewayStack() *cobra.Command {
	var installerDir string
	var kubeconfig string
	var withAppGatewayChart bool
	var chartOnly bool

	cmd := &cobra.Command{
		Use:   "install-app-gateway",
		Short: "Re-install unified ingress stack (Linkerd + Envoy Gateway + app-gateway with EG data-plane mesh); normal installs use olares-cli os install",
		Run: func(cmd *cobra.Command, args []string) {
			if installerDir == "" {
				_ = cmd.Usage()
				log.Fatal("error: --installer-dir is required")
			}

			installerDir = filepath.Clean(installerDir)
			_ = os.Setenv("OLARES_INSTALLER_DIR", installerDir)
			_ = os.Setenv("APP_GATEWAY_STACK_ENABLED", "1")
			if kubeconfig != "" {
				_ = os.Setenv("KUBECONFIG", kubeconfig)
			}

			initInstallAppGatewayLogger(installerDir)
			defer func() { _ = logger.Sync() }()

			runtime := &common.KubeRuntime{Arg: &common.Argument{}}
			kubeAction := common.KubeAction{KubeConf: &common.KubeConf{Arg: runtime.Arg}}

			if chartOnly {
				if err := (&terminus.InstallAppGatewayChart{KubeAction: kubeAction}).Execute(runtime); err != nil {
					log.Fatalf("error: install app-gateway chart: %v", err)
				}
				if err := (&terminus.WaitAppGatewayDataPlaneMeshed{KubeAction: kubeAction}).Execute(runtime); err != nil {
					log.Fatalf("error: wait EG data-plane mesh: %v", err)
				}
				return
			}

			if err := (&terminus.InstallAppGatewayVendor{KubeAction: kubeAction}).Execute(runtime); err != nil {
				log.Fatalf("error: install vendor: %v", err)
			}
			if withAppGatewayChart {
				if err := (&terminus.InstallAppGatewayChart{KubeAction: kubeAction}).Execute(runtime); err != nil {
					log.Fatalf("error: install app-gateway chart: %v", err)
				}
				if err := (&terminus.WaitAppGatewayDataPlaneMeshed{KubeAction: kubeAction}).Execute(runtime); err != nil {
					log.Fatalf("error: wait EG data-plane mesh: %v", err)
				}
			}
		},
	}

	cmd.Flags().StringVar(&installerDir, "installer-dir", "", "Olares installer dist directory (e.g. .dist)")
	cmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "kubeconfig path (optional; default ~/.kube/config — on Olares node usually copied from /etc/rancher/k3s/k3s.yaml)")
	cmd.Flags().BoolVar(&withAppGatewayChart, "with-app-gateway-chart", true, "Install app-gateway Helm chart (after vendor; ignored with --chart-only)")
	cmd.Flags().BoolVar(&chartOnly, "chart-only", false, "Install only the app-gateway Helm chart (vendor stack must already be installed)")

	return cmd
}
