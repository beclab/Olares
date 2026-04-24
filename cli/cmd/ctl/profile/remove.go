package profile

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/auth"
	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

// NewRemoveCommand: `olares-cli profile remove <name>`
//
// Removes the profile entry AND its stored token in one shot. There is no
// separate `auth logout` — `profile remove` is the canonical way to invalidate
// local credentials. If the removed profile was the current one, the current
// pointer falls back to PreviousProfile (when still valid) or to the first
// remaining profile.
//
// Token deletion failures are reported but don't stop config save: a
// dangling token entry is harmless (it'll just be stale) and we'd rather
// have a consistent config.json than abort halfway.
func NewRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "delete a profile and its stored token",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runRemove(args[0])
		},
	}
}

func runRemove(key string) error {
	cfg, err := cliconfig.LoadMultiProfileConfig()
	if err != nil {
		return err
	}
	removed, ok := cfg.Remove(key)
	if !ok {
		return fmt.Errorf("profile %q not found", key)
	}
	if err := cliconfig.SaveMultiProfileConfig(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	store, err := auth.NewFileStore()
	if err != nil {
		return err
	}
	if err := store.Delete(removed.OlaresID); err != nil && !errors.Is(err, auth.ErrTokenNotFound) {
		// Non-fatal: config is already updated.
		fmt.Printf("warning: failed to clear stored token for %s: %v\n", removed.OlaresID, err)
	}

	fmt.Printf("removed profile %s (%s)\n", removed.DisplayName(), removed.OlaresID)
	if cfg.CurrentProfile != "" {
		fmt.Printf("current profile is now: %s\n", cfg.CurrentProfile)
	} else if len(cfg.Profiles) == 0 {
		fmt.Println("no profiles remain.")
	}
	return nil
}
