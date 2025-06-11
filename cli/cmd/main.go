package main

import (
	"os"
	"os/exec"

	"github.com/beclab/Olares/cli/cmd/ctl"
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
