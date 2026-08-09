package dev

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

// NewNodeCommand assembles `olares-cli dev node` — where the ssh
// transport's coordinates are stored, once, per profile.
func NewNodeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Configure how `dev push --transport ssh` reaches your Olares node",
		Long: `Store the SSH coordinates of the Olares node for the active profile.

The settings live in ~/.olares-cli/config.json next to the profile they
belong to, so switching profiles switches nodes with them.

There is deliberately no --password flag. A password would have to be
written to that file or into the keychain, and the keychain is reserved
for the tokens that gate access to the instance itself. Use a key or an
SSH agent.
`,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("unknown command %q for %q", args[0], c.CommandPath())
			}
			return c.Help()
		},
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newNodeSetCommand())
	cmd.AddCommand(newNodeShowCommand())
	cmd.AddCommand(newNodeClearCommand())
	return cmd
}

func newNodeSetCommand() *cobra.Command {
	var (
		address    string
		user       string
		port       int
		privateKey string
	)
	cmd := &cobra.Command{
		Use:   "set",
		Short: "set the node address / user / key for the active profile",
		Example: `  olares-cli dev node set --address 192.168.178.42 --user ronny
  olares-cli dev node set --address olares.lan --user root --private-key ~/.ssh/id_ed25519`,
		RunE: func(c *cobra.Command, args []string) error {
			if strings.TrimSpace(address) == "" {
				return fmt.Errorf("--address is required")
			}
			if strings.TrimSpace(user) == "" {
				return fmt.Errorf("--user is required")
			}
			cfg, err := cliconfig.LoadMultiProfileConfig()
			if err != nil {
				return err
			}
			current := cfg.Current()
			if current == nil {
				return fmt.Errorf("no profile selected; run `olares-cli profile login <olares-id>` first")
			}
			current.DevNode = &cliconfig.DevNodeConfig{
				Address:        strings.TrimSpace(address),
				Port:           port,
				User:           strings.TrimSpace(user),
				PrivateKeyPath: strings.TrimSpace(privateKey),
			}
			if err := cliconfig.SaveMultiProfileConfig(cfg); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "node for %s set to %s@%s:%d\n",
				current.DisplayName(), current.DevNode.User, current.DevNode.Address, portOrDefault(current.DevNode.Port))
			return nil
		},
	}
	cmd.Flags().StringVar(&address, "address", "", "node hostname or IP (REQUIRED)")
	cmd.Flags().StringVar(&user, "user", "", "SSH user; must be able to reach containerd (root or a sudoer) (REQUIRED)")
	cmd.Flags().IntVar(&port, "port", 22, "SSH port")
	cmd.Flags().StringVar(&privateKey, "private-key", "", "path to an SSH private key; empty uses the agent (SSH_AUTH_SOCK)")
	return cmd
}

func newNodeShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "print the node configured for the active profile",
		RunE: func(c *cobra.Command, args []string) error {
			cfg, err := cliconfig.LoadMultiProfileConfig()
			if err != nil {
				return err
			}
			current := cfg.Current()
			if current == nil {
				return fmt.Errorf("no profile selected; run `olares-cli profile login <olares-id>` first")
			}
			if current.DevNode == nil {
				return fmt.Errorf("no node configured for %s; run `olares-cli dev node set --address <host> --user <user>`",
					current.DisplayName())
			}
			tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			defer tw.Flush()
			fmt.Fprintf(tw, "Profile:\t%s\n", current.DisplayName())
			fmt.Fprintf(tw, "Address:\t%s\n", current.DevNode.Address)
			fmt.Fprintf(tw, "Port:\t%d\n", portOrDefault(current.DevNode.Port))
			fmt.Fprintf(tw, "User:\t%s\n", current.DevNode.User)
			key := current.DevNode.PrivateKeyPath
			if key == "" {
				key = "(ssh agent)"
			}
			fmt.Fprintf(tw, "Private key:\t%s\n", key)
			return nil
		},
	}
}

func newNodeClearCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "forget the node configured for the active profile",
		RunE: func(c *cobra.Command, args []string) error {
			cfg, err := cliconfig.LoadMultiProfileConfig()
			if err != nil {
				return err
			}
			current := cfg.Current()
			if current == nil {
				return fmt.Errorf("no profile selected")
			}
			if current.DevNode == nil {
				fmt.Fprintf(os.Stdout, "no node configured for %s — nothing to clear\n", current.DisplayName())
				return nil
			}
			current.DevNode = nil
			if err := cliconfig.SaveMultiProfileConfig(cfg); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "node cleared for %s\n", current.DisplayName())
			return nil
		},
	}
}

// currentDevNode returns the active profile's node config, or nil when
// none is set. A missing profile is not an error here: the transport
// selector reports "no node configured" either way, and `dev push
// --transport local` must keep working before anyone has logged in.
func currentDevNode() *cliconfig.DevNodeConfig {
	cfg, err := cliconfig.LoadMultiProfileConfig()
	if err != nil {
		return nil
	}
	current := cfg.Current()
	if current == nil {
		return nil
	}
	return current.DevNode
}

func portOrDefault(p int) int {
	if p == 0 {
		return 22
	}
	return p
}
