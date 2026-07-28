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
	// Take the config lock + re-read inside the mutation so a concurrent
	// background location write (during a long-running command) can't clobber
	// the current-profile switch, and vice-versa.
	var target *cliconfig.ProfileConfig
	if err := cliconfig.UpdateLocked(func(cfg *cliconfig.MultiProfileConfig) error {
		t, err := cfg.SetCurrent(key)
		if err != nil {
			return err
		}
		target = t
		return nil
	}); err != nil {
		return err
	}
	fmt.Printf("switched to profile %s (%s)\n", target.DisplayName(), target.OlaresID)
	return nil
}
