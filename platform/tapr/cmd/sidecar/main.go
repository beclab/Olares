package main

import (
	"flag"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	cmd := &cobra.Command{
		Use:   "tapr-sidecar-filters",
		Short: "tapr-sidecar-filters",
		Long:  `The TAPR sidecar filters provides the application runtime envoy sidecar filters plugin`,
		Run: func(cmd *cobra.Command, args []string) {

			klog.Info("tapr-sidecar shutdown ")
		},
	}

	klog.Info("tapr-sidecar starting ... ")

	if err := cmd.Execute(); err != nil {
		klog.Fatalln(err)
	}

}
