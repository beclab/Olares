package os

import (
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdPrintInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Print Olares info",
		Run: func(cmd *cobra.Command, args []string) {
			pipelines.PrintTerminusInfo()
		},
	}
	return cmd
}
