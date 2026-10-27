package node

import (
	"log"

	"github.com/beclab/Olares/cli/cmd/config"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdJoinCommand() *cobra.Command {
	var opts pipelines.JoinCommandOptions

	cmd := &cobra.Command{
		Use:   "join-command",
		Short: "print the command that joins a worker node to this cluster",
		Long: `Print the self-contained command to run on a machine that should join this
cluster as a worker node.

The command is written to stdout and everything else to stderr, so it can be
captured directly. By default the SSH account workers will use to reach this
master is confirmed and its password verified interactively; --yes and
--master-ssh-password make the command non-interactive for use from a script.`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.JoinCommand(cmd.Context(), opts); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}

	flagSetter := config.NewFlagSetterFor(cmd)
	config.AddCDNServiceFlagBy(flagSetter)
	flagSetter.Add(common.FlagMasterHost, "", "", "LAN IPv4 address workers should use to reach this master, detected automatically if unset")
	flagSetter.Add(common.FlagMasterSSHUser, "", "", "SSH username workers should use to access this master")
	flagSetter.Add(common.FlagMasterSSHPort, "", 0, "SSH port workers should use to access this master, defaults to 22")
	// Also reachable as MASTER_SSH_PASSWORD, which keeps the password out of the
	// process list.
	flagSetter.Add(common.FlagMasterSSHPassword, "p", "",
		"SSH password workers should use, instead of being asked for it").WithAlias("password")

	// Registered directly rather than through the flag setter: these names are
	// too general to bind as global environment variables.
	cmd.Flags().BoolVarP(&opts.AssumeYes, "yes", "y", false,
		"accept the detected address and SSH account without asking, and skip the SSH check")
	cmd.Flags().BoolVarP(&opts.Quiet, "quiet", "q", false,
		"print only the worker command, with no explanation")

	return cmd
}
