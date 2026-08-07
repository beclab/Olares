package profile

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

// NewUseCommand: `olares-cli profile use <name|->`
//
// `name` may be a profile alias (Name) or its OlaresID. The literal `-`
// switches back to the previous profile (a la `cd -`), and is rejected when
// PreviousProfile is unset.
func NewUseCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "use <name|->",
		Short: "switch the current profile (use `-` to switch back to the previous one)",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runUse(args[0])
		},
	}
}

func runUse(key string) error {
	cfg, err := cliconfig.LoadMultiProfileConfig()
	if err != nil {
		return err
	}
	target, err := cfg.SetCurrent(key)
	if err != nil {
		return err
	}
	if err := cliconfig.SaveMultiProfileConfig(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	fmt.Printf("switched to profile %s (%s)\n", target.DisplayName(), target.OlaresID)
	return nil
}
