package os

import (
	"log"

	"github.com/beclab/Olares/cli/cmd/ctl/options"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdChangeIP() *cobra.Command {
	o := options.NewChangeIPOptions()
	cmd := &cobra.Command{
		Use:   "change-ip",
		Short: "change The IP address of Olares OS",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.ChangeIPPipeline(o); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	o.AddFlags(cmd)
	return cmd
}
