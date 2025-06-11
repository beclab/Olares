package os

import (
	"log"

	"github.com/beclab/Olares/cli/cmd/ctl/options"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdPrecheck() *cobra.Command {
	o := options.NewPreCheckOptions()
	cmd := &cobra.Command{
		Use:   "precheck",
		Short: "precheck the installation compatibility of the system",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.StartPreCheckPipeline(o); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	o.AddFlags(cmd)
	return cmd
}
