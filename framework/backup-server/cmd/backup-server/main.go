package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"olares.com/backup-server/cmd/backup-server/apiserver"
	"olares.com/backup-server/cmd/backup-server/controller"
)

var rootCommand = cobra.Command{
	Use: "backup-server",
}

func completionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion",
		Short: "Generate the autocompletion script for the specified shell",
	}
}

func init() {
	completion := completionCommand()
	completion.Hidden = true
	rootCommand.AddCommand(completion)

	rootCommand.AddCommand(apiserver.NewAPIServerCommand())
	rootCommand.AddCommand(controller.NewControllerCommand())

}

func main() {
	if err := rootCommand.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
