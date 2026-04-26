package profile

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/auth"
	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

type importOptions struct {
	commonCredFlags
	refreshToken string
	noSwitch     bool
}

// NewImportCommand: `olares-cli profile import --olares-id <id> --refresh-token <tok> [...]`
//
// Mode B: bootstrap an access_token from a user-supplied refresh_token by
// performing exactly ONE call to /api/refresh. This is the way to seed a
// profile when the user obtained their refresh token elsewhere (LarePass,
// wizard activation, manual extraction).
//
// Phase 1 does NOT use the stored refresh_token for background renewal —
// that's a Phase 2 deliverable. The same `auth.Refresh` HTTP call will be
// reused there, so the wire-format contract is locked in now.
func NewImportCommand() *cobra.Command {
	o := &importOptions{}
	cmd := &cobra.Command{
		Use:   "import",
		Short: "import a refresh token to bootstrap an access token (mode B)",
		Long: `Import an existing refresh_token (e.g. obtained via LarePass or the wizard
activation flow) and exchange it once for an access_token via /api/refresh.

The profile is auto-created on first import. Importing into an
already-authenticated profile is rejected; remove the profile first
(` + "`olares-cli profile remove <id>`" + `) and import again.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runImport(cmd.Context(), o)
		},
	}
	o.commonCredFlags.bind(cmd)
	cmd.Flags().StringVar(&o.refreshToken, "refresh-token", "", "refresh token to bootstrap (required)")
	cmd.Flags().BoolVar(&o.noSwitch, "no-switch", false, "do not change the current profile after a successful import (useful for scripts)")
	return cmd
}

func runImport(ctx context.Context, o *importOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if o.refreshToken == "" {
		return errors.New("--refresh-token is required")
	}
	_, _, authURL, err := o.commonCredFlags.validateAndDeriveAuthURL()
	if err != nil {
		return err
	}

	cfg, err := cliconfig.LoadMultiProfileConfig()
	if err != nil {
		return err
	}
	store := auth.NewTokenStore()
	profile, err := ensureProfileWritable(cfg, store, o.commonCredFlags, time.Now())
	if err != nil {
		return err
	}

	tok, err := auth.Refresh(ctx, auth.RefreshRequest{
		AuthURL:            authURL,
		RefreshToken:       o.refreshToken,
		InsecureSkipVerify: o.insecureSkipVerify,
	})
	if err != nil {
		return err
	}

	res, err := persistTokenAndProfile(cfg, store, profile, tok, !o.noSwitch)
	if err != nil {
		return err
	}

	fmt.Printf("imported credentials for %s (profile: %s)\n", o.olaresID, profile.DisplayName())
	printSwitchNotice(res, profile.DisplayName())
	printStorageNotice(profile.OlaresID)
	return nil
}
