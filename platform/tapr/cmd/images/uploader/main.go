package main

import (
	"flag"

	"bytetrade.io/web3os/tapr/cmd/images/uploader/app"
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
		Use:   "images-uploader",
		Short: "images uploader",
		Long:  `The images uploader provides the images uploading service of application runtime`,
		Run: func(cmd *cobra.Command, args []string) {
			s := &app.Server{
				KubeConfig: config,
			}

			s.ServerRun()

			klog.Info("images-uploader shutdown ")
		},
	}

	klog.Info("images-uploader starting ... ")

	if err := cmd.Execute(); err != nil {
		klog.Fatalln(err)
	}
}
