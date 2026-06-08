package ctl

import (
	"fmt"
	goOS "os"

	"github.com/beclab/Olares/cli/cmd/config"
	"github.com/beclab/Olares/cli/cmd/ctl/amdgpu"
	"github.com/beclab/Olares/cli/cmd/ctl/chart"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster"
	"github.com/beclab/Olares/cli/cmd/ctl/dashboard"
	"github.com/beclab/Olares/cli/cmd/ctl/disk"
	"github.com/beclab/Olares/cli/cmd/ctl/doctor"
	"github.com/beclab/Olares/cli/cmd/ctl/files"
	"github.com/beclab/Olares/cli/cmd/ctl/gpu"
	"github.com/beclab/Olares/cli/cmd/ctl/market"
	"github.com/beclab/Olares/cli/cmd/ctl/node"
	"github.com/beclab/Olares/cli/cmd/ctl/os"
	"github.com/beclab/Olares/cli/cmd/ctl/osinfo"
	"github.com/beclab/Olares/cli/cmd/ctl/profile"
	"github.com/beclab/Olares/cli/cmd/ctl/settings"
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
	// client (market, profile, files, dashboard, settings, cluster) all
	// reach into this same instance so credential resolution and HTTPClient
	// construction are memoized across verbs in the same invocation.
	factory := cmdutil.NewFactory()
	cobra.OnInitialize(func() {
		config.Init()
	})
	cmds := &cobra.Command{
		Use:               "olares-cli",
		Short:             "Olares Installer",
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
		Version:           version.VERSION,
		// SilenceErrors: cmd/main.go prints the error once on non-zero exit; without
		// this, Cobra also prints to stderr and users see duplicate "Error:" lines.
		SilenceErrors: true,
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
		},
	}
	cmds.Flags().BoolVar(&showVendor, "vendor", false, "show the vendor type of olares-cli")

	// The --olares-version override lives on the `profile` command tree, not
	// here: backend version is a per-profile property (cached in config.json,
	// detected at login). Other command trees that branch on it (market,
	// version-aware settings) read that cache and auto-detect on demand; to
	// override it, use the profile namespace, and to re-detect on demand use
	// `profile whoami --refresh` / `profile list --refresh`.
	// Identity is single-source: whichever profile `olares-cli profile use`
	// (or the most recent `profile login` / `profile import`) selected. There
	// is intentionally no per-invocation `--profile` override — agents and
	// scripts must commit to one role up-front rather than silently hopping
	// identities mid-pipeline. To target a different profile, run
	// `olares-cli profile use <name>` first.

	// OLARES_CLI_REMOTE_ONLY=1 hides host-side verbs (install, upgrade, node,
	// os, gpu, disk, wizard, user, osinfo, amdgpu) that require an Olares host
	// filesystem (~/.olares/versions/<v>/...) laid down by the install wizard.
	// The npm distribution sets this from its Node shim (cli/npm/bin/olares-cli.js)
	// so `npx @olares/cli` never exposes those verbs to remote/agent users.
	// The host-bundled binary at /usr/local/bin/olares-cli leaves the env unset
	// and behaves as before — all verbs registered.
	remoteOnly := goOS.Getenv("OLARES_CLI_REMOTE_ONLY") == "1"

	if !remoteOnly {
		cmds.AddCommand(osinfo.NewCmdInfo())
		cmds.AddCommand(os.NewOSCommands()...)
		cmds.AddCommand(node.NewNodeCommand())
		cmds.AddCommand(gpu.NewCmdGpu())
		cmds.AddCommand(amdgpu.NewCmdAmdGpu())
		cmds.AddCommand(user.NewUserCommand())
		cmds.AddCommand(wizard.NewWizardCommand())
		cmds.AddCommand(disk.NewDiskCommand())
	}

	// Always-on: developer utilities (chart) + remote/agent verbs that go
	// through control-hub.<terminus> via the active profile's token.
	cmds.AddCommand(chart.NewChartCommand())
	cmds.AddCommand(market.NewMarketCommand(factory))
	cmds.AddCommand(profile.NewProfileCommand(factory))
	cmds.AddCommand(files.NewFilesCommand(factory))
	cmds.AddCommand(doctor.NewDoctorCommand(factory))
	cmds.AddCommand(dashboard.NewDashboardCommand(factory))
	cmds.AddCommand(settings.NewSettingsCommand(factory))
	cmds.AddCommand(cluster.NewClusterCommand(factory))

	return cmds
}
