package ctl

import (
	"github.com/beclab/Olares/cli/cmd/ctl/gpu"
	"github.com/beclab/Olares/cli/cmd/ctl/node"
	"github.com/beclab/Olares/cli/cmd/ctl/os"
	"github.com/beclab/Olares/cli/cmd/ctl/osinfo"
	"github.com/beclab/Olares/cli/version"
	"github.com/spf13/cobra"
)

func NewDefaultCommand() *cobra.Command {
	cmds := &cobra.Command{
		Use:               "olares-cli",
		Short:             "Olares Installer",
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
		Version:           version.VERSION,
	}

	cmds.AddCommand(osinfo.NewCmdInfo())
	cmds.AddCommand(os.NewOSCommands()...)
	cmds.AddCommand(node.NewNodeCommand())
	cmds.AddCommand(gpu.NewCmdGpu())

	return cmds
}
