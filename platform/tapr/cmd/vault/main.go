package main

import (
	"flag"

	"bytetrade.io/web3os/tapr/cmd/vault/app"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	config := ctrl.GetConfigOrDie()

	cmd := &cobra.Command{
		Use:   "secret-vault",
		Short: "secret vault",
		Long:  `The secret vault provides the secret service of application runtime`,
		Run: func(cmd *cobra.Command, args []string) {
			s := &app.Server{
				KubeConfig: config,
			}

			err := s.Init()
			if err != nil {
				panic(err)
			}

			s.ServerRun()

			klog.Info("secret-vault shutdown ")
		},
	}

	klog.Info("secret-vault starting ... ")

	if err := cmd.Execute(); err != nil {
		klog.Fatalln(err)
	}
}
