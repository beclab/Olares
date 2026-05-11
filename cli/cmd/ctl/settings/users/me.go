package users

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// NewMeCommand: `olares-cli settings users me [--refresh] [-o table|json]`
//
// Canonical "Settings -> Users -> me" entry point — a verb under the
// users area rather than a sibling of users / appearance / apps. Same
// behavior as `olares-cli profile whoami` and `olares-cli settings me
// whoami`: both are aliases that flow through pkg/whoami.Run, so they
// share output shape, cache write semantics, and the --refresh flag.
//
// We deliberately ship all three at once (rather than picking a winner)
// because each entry point matches a different mental model:
//
//   - profile whoami         → "where am I logged in / which profile is
//                              active" — same family as profile list,
//                              profile use.
//   - settings users me      → docs.olares.com/manual/olares/settings/
//                              UI mapping; you found Users in the SPA
//                              menu, the SPA shows you at the top.
//   - settings me whoami     → tucked under the Person dropdown in the
//                              SPA; CLI-friendly self-service tree.
//
// All three call the same code; the help text disambiguates which is
// which.
func NewMeCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		refresh   bool
		outputRaw string
	)
	cmd := &cobra.Command{
		Use:   "me",
		Short: "show the current profile's identity and role (alias for `profile whoami`)",
		Long: `Show the active profile's olaresId, friendly name, and role
("Owner" / "Admin" / "User") on the target Olares instance.

Same implementation as ` + "`olares-cli profile whoami`" + ` and
` + "`olares-cli settings me whoami`" + `. Mounted here so it discovers naturally
under the Settings -> Users tree (see the SPA's Settings UI).
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runMe(c.Context(), f, refresh, outputRaw)
		},
	}
	cmd.Flags().BoolVar(&refresh, "refresh", false, "force a fresh /api/backend/v1/user-info roundtrip and update the cached role")
	cmd.Flags().StringVarP(&outputRaw, "output", "o", "table", "output format: table, json")
	return cmd
}

func runMe(ctx context.Context, f *cmdutil.Factory, refresh bool, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if f == nil {
		return fmt.Errorf("internal error: settings users me not wired with cmdutil.Factory")
	}

	format, err := whoami.ParseOutput(outputRaw)
	if err != nil {
		return err
	}
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}
	hc, err := f.HTTPClient(ctx)
	if err != nil {
		return err
	}
	cfg, err := cliconfig.LoadMultiProfileConfig()
	if err != nil {
		return err
	}
	client := whoami.NewHTTPClient(hc, rp.DesktopURL, rp.OlaresID)
	return whoami.Run(ctx, client, cfg, rp.OlaresID, refresh, format, nil, os.Stdout)
}
