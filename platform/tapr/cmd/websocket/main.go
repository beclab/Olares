package main

import (
	"bytetrade.io/web3os/tapr/cmd/websocket/app"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	klog.InitFlags(nil)
	var config *rest.Config

	if !app.DEBUG {
		config = ctrl.GetConfigOrDie()
	}

	cmd := &cobra.Command{
		Use:   "websocket-gateway",
		Short: "websocket gateway",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			server := &app.Server{
				KubeConfig: config,
			}

			err := server.Init()
			if err != nil {
				klog.Fatalln(err)
				panic(err)
			}

			server.ServerRun()

			klog.Info("websocket shutdown ")
		},
	}

	klog.Info("websocket starting ... ")

	if err := cmd.Execute(); err != nil {
		klog.Fatalln(err)
	}
}
