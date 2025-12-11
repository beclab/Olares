package main

import (
	"integration/pkg/constant"
	"integration/pkg/hertz"
	"integration/pkg/utils"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var rootCmd = &cobra.Command{
	Use:   "integration",
	Short: "Integrated Account Services",
	Run: func(cmd *cobra.Command, args []string) {
		hertz.HertzServer()
	},
}

func init() {
	constant.InfisicalAddr = utils.GetenvOrDefault(constant.EnvInfisicalUrl, constant.InfisicalService)
	constant.Environment = utils.GetenvOrDefault(constant.EnvEnvironment, constant.DefaultEnvironment)
	constant.Workspace = utils.GetenvOrDefault(constant.EnvWorkspace, constant.DefaultWorkspace)

	rootCmd.SetVersionTemplate("Integration version {{printf \"%s\" .Version}}\n")
}

func main() {
	klog.InitFlags(nil)
	klog.Info("Starting integration server...")

	if err := rootCmd.Execute(); err != nil {
		klog.Fatal(err)
	}
}
