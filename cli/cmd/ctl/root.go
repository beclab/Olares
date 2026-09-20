package ctl

import (
	"fmt"
	goOS "os"
	"sync"

	"github.com/beclab/Olares/cli/cmd/config"
	"github.com/beclab/Olares/cli/cmd/ctl/amdgpu"
	"github.com/beclab/Olares/cli/cmd/ctl/chart"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster"
	"github.com/beclab/Olares/cli/cmd/ctl/dashboard"
	"github.com/beclab/Olares/cli/cmd/ctl/disk"
	"github.com/beclab/Olares/cli/cmd/ctl/doctor"
	"github.com/beclab/Olares/cli/cmd/ctl/files"
	"github.com/beclab/Olares/cli/cmd/ctl/gpu"
	"github.com/beclab/Olares/cli/cmd/ctl/knowledge"
	"github.com/beclab/Olares/cli/cmd/ctl/market"
	"github.com/beclab/Olares/cli/cmd/ctl/node"
	"github.com/beclab/Olares/cli/cmd/ctl/os"
	"github.com/beclab/Olares/cli/cmd/ctl/osinfo"
	"github.com/beclab/Olares/cli/cmd/ctl/preinstall"
	"github.com/beclab/Olares/cli/cmd/ctl/profile"
	"github.com/beclab/Olares/cli/cmd/ctl/router"
	"github.com/beclab/Olares/cli/cmd/ctl/search"
	"github.com/beclab/Olares/cli/cmd/ctl/settings"
	"github.com/beclab/Olares/cli/cmd/ctl/skills"
	"github.com/beclab/Olares/cli/cmd/ctl/update"
	"github.com/beclab/Olares/cli/cmd/ctl/user"
	versioncmd "github.com/beclab/Olares/cli/cmd/ctl/version"
	"github.com/beclab/Olares/cli/cmd/ctl/wizard"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
	"github.com/beclab/Olares/cli/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const unknownVerbGroupAnnotation = "olares-cli/unknown-verb-group"

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
		// SilenceUsage on the root is what every subtree needs, and Cobra
		// reads it here: it prints usage only when neither the root nor the
		// command that ran has the flag set. Subtrees used to reach the
		// second of those through a PersistentPreRun of their own, which is
		// how a purely cosmetic setting ended up displacing the identity
		// import below — Cobra runs the nearest hook and nothing above it.
		SilenceUsage: true,
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

	// Keep the first line byte-for-byte identical to Cobra's default
	// ("olares-cli version <VERSION>") so existing parsers that read only
	// the first line / the third whitespace-delimited token keep working,
	// while appending the build metadata on following lines.
	cmds.SetVersionTemplate(fmt.Sprintf(
		"{{with .Name}}{{.}} {{end}}version {{.Version}}\nGit commit: %s\nBuild time: %s\n",
		version.GitCommit, version.BuildTime,
	))

	// Version-compat controls (--olares-version / --refresh-version) live on
	// the `profile` command tree, not here: backend version is a per-profile
	// property (cached in config.json, eagerly fetched at login). Other
	// command trees that branch on it (market, version-aware settings) read
	// that cache and auto-detect on demand; to override or force a refresh,
	// use the profile namespace (e.g. `profile list --refresh-version`).
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

	// Always-on: developer utilities (chart, preinstall) + remote/agent verbs
	// that go through control-hub.<terminus> via the active profile's token.
	cmds.AddCommand(chart.NewChartCommand())
	cmds.AddCommand(preinstall.NewPreinstallCommand())
	// The skill suite is compiled in, so these verbs read and write local
	// files only — nothing about them is host-side, and the npm distribution
	// is exactly where an agent needs them.
	cmds.AddCommand(skills.NewSkillsCommand())
	// `version` is the machine-readable form of --version, which stays byte
	// for byte as it is because the npm install wizard parses it.
	cmds.AddCommand(versioncmd.NewVersionCommand())
	// `update` is olares-cli updating itself, and is registered on every
	// channel including the Olares host — that host is where the confusion
	// with `upgrade` (which upgrades Olares OS) actually happens, so it is
	// where a verb that says so has to exist. It refuses to npm-install over
	// an OS bundle; see cmd/ctl/update.
	cmds.AddCommand(update.NewUpdateCommand())
	cmds.AddCommand(market.NewMarketCommand(factory))
	cmds.AddCommand(profile.NewProfileCommand(factory))
	cmds.AddCommand(knowledge.NewKnowledgeCommand(factory))
	cmds.AddCommand(files.NewFilesCommand(factory))
	cmds.AddCommand(doctor.NewDoctorCommand(factory))
	cmds.AddCommand(dashboard.NewDashboardCommand(factory))
	cmds.AddCommand(settings.NewSettingsCommand(factory))
	cmds.AddCommand(search.NewSearchCommand(factory))
	cmds.AddCommand(router.NewRouterCommand(factory))
	cmds.AddCommand(cluster.NewClusterCommand(factory))

	wireUnknownVerbRefusals(cmds)
	wireManagedIdentity(cmds)
	skipPreRunsForGroupHelp(cmds)
	return cmds
}

// wireManagedIdentity puts the platform-credential import in front of every
// command, whichever PersistentPreRun Cobra decides to run.
//
// Hanging it on the root's hook alone was wrong: Cobra walks up from the
// command that ran, executes the first persistent hook it finds and breaks.
// A subtree that declares one of its own — for years these did nothing but
// set SilenceUsage — therefore removed the identity from every verb beneath
// it, and `market list` in a fresh container reported that no profile was
// configured while the credential sat unread on its mount. cmd/main.go
// records the same trap for the skill-drift notice.
//
// Wrapping the declared hooks is enough to cover the tree: a command with no
// hook of its own runs an ancestor's, and the root declares one, so every
// path ends at something wrapped. Ordering matters twice — this runs before
// skipPreRunsForGroupHelp so `<group> help` still skips everything including
// the import, and the import runs before the wrapped body so a gate like
// files/nfs's version check has an identity to work with.
//
// It stays off the Factory's lazy chain because `profile list` reads
// config.json and the keychain directly, and only reaches for the Factory on
// --refresh-version: a profile imported there would be invisible in the one
// command most likely to go looking for it.
//
// The guard is per tree rather than per process: one tree serves one
// invocation, and a test that builds a second one is asking for a second
// container's worth of behavior, not a replay of the first.
func wireManagedIdentity(root *cobra.Command) {
	once := &sync.Once{}
	do := func(cmd *cobra.Command) {
		once.Do(func() { credential.ImportManagedCredential(cmd.Context()) })
	}
	var wire func(*cobra.Command)
	wire = func(cmd *cobra.Command) {
		if run := cmd.PersistentPreRun; run != nil {
			cmd.PersistentPreRun = func(current *cobra.Command, args []string) {
				do(current)
				run(current, args)
			}
		}
		if run := cmd.PersistentPreRunE; run != nil {
			cmd.PersistentPreRunE = func(current *cobra.Command, args []string) error {
				do(current)
				return run(current, args)
			}
		}
		for _, child := range cmd.Commands() {
			wire(child)
		}
	}
	wire(root)
}

func wireUnknownVerbRefusals(cmd *cobra.Command) {
	children := cmd.Commands()
	if len(children) > 0 && !cmd.Runnable() {
		cmd.Args = cmdutil.RefuseUnknownVerbGroupArgs
		cmd.RunE = cmdutil.RefuseUnknownVerb
		if cmd.Annotations == nil {
			cmd.Annotations = make(map[string]string)
		}
		cmd.Annotations[unknownVerbGroupAnnotation] = "true"
	}
	for _, child := range children {
		wireUnknownVerbRefusals(child)
	}
}

func skipPreRunsForGroupHelp(cmd *cobra.Command) {
	if run := cmd.PersistentPreRun; run != nil {
		cmd.PersistentPreRun = func(current *cobra.Command, args []string) {
			if !isGroupHelp(current, args) {
				run(current, args)
			}
		}
	}
	if run := cmd.PersistentPreRunE; run != nil {
		cmd.PersistentPreRunE = func(current *cobra.Command, args []string) error {
			if isGroupHelp(current, args) {
				return nil
			}
			return run(current, args)
		}
	}
	for _, child := range cmd.Commands() {
		skipPreRunsForGroupHelp(child)
	}
}

func isGroupHelp(cmd *cobra.Command, args []string) bool {
	return cmd.Annotations[unknownVerbGroupAnnotation] == "true" &&
		len(args) > 0 && args[0] == "help"
}
