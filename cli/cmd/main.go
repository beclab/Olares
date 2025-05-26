package main

import (
	"os"
	"os/exec"

	"bytetrade.io/web3os/installer/cmd/ctl"
)

func main() {
	cmd := ctl.NewDefaultCommand()
	_ = exec.Command("/bin/bash", "-c", "ulimit -u 65535").Run()
	_ = exec.Command("/bin/bash", "-c", "ulimit -n 65535").Run()

	if err := cmd.Execute(); err != nil {
		// fmt.Println(err)
		os.Exit(1)
	}
}
