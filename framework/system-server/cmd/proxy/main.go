package main

import (
	"os"

	"bytetrade.io/web3os/system-server/pkg/serviceproxy/v2alpha1"
	"k8s.io/component-base/cli"
)

func main() {
	command := v2alpha1.NewKubeRBACProxyCommand()
	code := cli.Run(command)
	os.Exit(code)
}
