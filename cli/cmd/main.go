package main

import (
	"fmt"
	"os"

	"github.com/beclab/Olares/cli/cmd/ctl"
)

func main() {
	cmd := ctl.NewDefaultCommand()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
