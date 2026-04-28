package ctl

import (
	"fmt"

	"github.com/beclab/Olares/cli/cmd/config"
	"github.com/beclab/Olares/cli/cmd/ctl/amdgpu"
	"github.com/beclab/Olares/cli/cmd/ctl/app"
	"github.com/beclab/Olares/cli/cmd/ctl/disk"
	"github.com/beclab/Olares/cli/cmd/ctl/files"
	"github.com/beclab/Olares/cli/cmd/ctl/gpu"
	"github.com/beclab/Olares/cli/cmd/ctl/market"
	"github.com/beclab/Olares/cli/cmd/ctl/node"
	"github.com/beclab/Olares/cli/cmd/ctl/os"
	"github.com/beclab/Olares/cli/cmd/ctl/osinfo"
	"github.com/beclab/Olares/cli/cmd/ctl/profile"
	"github.com/beclab/Olares/cli/cmd/ctl/user"
	"github.com/beclab/Olares/cli/cmd/ctl/wizard"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewDefaultCommand() *cobra.Command {
	var showVendor bool
	// One Factory per process. Subcommands that need an authenticated HTTP
	// client (today: files; future: user/app/settings) all reach into this
	// same instance so credential resolution and HTTPClient construction are
	// memoized across verbs in the same invocation.
	factory := cmdutil.NewFactory()
	cobra.OnInitialize(func() {
		config.Init()
	})
	cmds := &cobra.Command{
		Use:               "olares-cli",
		Short:             "Olares Installer",
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
		Version:           version.VERSION,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.InheritedFlags())
			viper.BindPFlags(cmd.PersistentFlags())
			viper.BindPFlags(cmd.Flags())
		},
		Run: func(cmd *cobra.Command, args []string) {
			if showVendor {
				fmt.Println(version.VENDOR)
			} else {
				cmd.Usage()
			}
			return
		},
	}
	cmds.Flags().BoolVar(&showVendor, "vendor", false, "show the vendor type of olares-cli")
	// Persistent --profile flag binds straight onto the shared Factory so
	// that subcommands which use it (factory.ResolveProfile) automatically
	// honor the override without each having to re-declare the flag.
	cmds.PersistentFlags().StringVar(&factory.ProfileOverride, "profile", "",
		"olaresId of the profile to use (overrides the currently-selected one)")

	cmds.AddCommand(osinfo.NewCmdInfo())
	cmds.AddCommand(os.NewOSCommands()...)
	cmds.AddCommand(node.NewNodeCommand())
	cmds.AddCommand(gpu.NewCmdGpu())
	cmds.AddCommand(amdgpu.NewCmdAmdGpu())
	cmds.AddCommand(user.NewUserCommand())
	cmds.AddCommand(wizard.NewWizardCommand())
	cmds.AddCommand(disk.NewDiskCommand())
	cmds.AddCommand(market.NewMarketCommand(factory))
	cmds.AddCommand(app.NewAppCommand())
	cmds.AddCommand(profile.NewProfileCommand())
	cmds.AddCommand(files.NewFilesCommand(factory))

	return cmds
}
