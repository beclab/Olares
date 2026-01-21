package ctl

import (
	"github.com/beclab/Olares/cli/cmd/ctl/amdgpu"
	"github.com/beclab/Olares/cli/cmd/ctl/disk"
	"github.com/beclab/Olares/cli/cmd/ctl/gpu"
	"github.com/beclab/Olares/cli/cmd/ctl/node"
	"github.com/beclab/Olares/cli/cmd/ctl/os"
	"github.com/beclab/Olares/cli/cmd/ctl/osinfo"
	"github.com/beclab/Olares/cli/cmd/ctl/user"
	"github.com/beclab/Olares/cli/version"
	"github.com/spf13/cobra"
)

func NewDefaultCommand() *cobra.Command {
	var showVendor bool
	cmds := &cobra.Command{
		Use:               "olares-cli",
		Short:             "Olares Installer",
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
		Version:           version.VERSION,
		Run: func(cmd *cobra.Command, args []string) {
			if showVendor {
				println(version.VENDOR)
			} else {
				cmd.Usage()
			}
			return
		},
	}
	cmds.Flags().BoolVar(&showVendor, "vendor", false, "show the vendor type of olares-cli")

	cmds.AddCommand(osinfo.NewCmdInfo())
	cmds.AddCommand(os.NewOSCommands()...)
	cmds.AddCommand(node.NewNodeCommand())
	cmds.AddCommand(gpu.NewCmdGpu())
	cmds.AddCommand(amdgpu.NewCmdAmdGpu())
	cmds.AddCommand(user.NewUserCommand())
	cmds.AddCommand(disk.NewDiskCommand())

	return cmds
}
