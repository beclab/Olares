package main

import (
	"github.com/beclab/Olares/cli/cmd/ctl"
	"os"
)

func main() {
	cmd := ctl.NewDefaultCommand()

	if err := cmd.Execute(); err != nil {
		// fmt.Println(err)
		os.Exit(1)
	}
}
