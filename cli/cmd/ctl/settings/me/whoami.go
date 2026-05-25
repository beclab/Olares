package me

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// NewWhoamiCommand: `olares-cli settings me whoami [--refresh] [-o table|json]`
//
// Third entry point into the same shared driver (pkg/whoami.Run) that
// `profile whoami` and `settings users me` use. Lives here so the SPA's
// Person dropdown -> "About me" mental model has a CLI surface that
// matches one-for-one. See cmd/ctl/settings/users/me.go for the
// rationale on shipping all three at once.
func NewWhoamiCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		refresh   bool
		outputRaw string
	)
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "show the current profile's identity and role (alias for `profile whoami`)",
		Long: `Show the active profile's olaresId, friendly name, and role
("Owner" / "Admin" / "User") on the target Olares instance.

Same implementation as ` + "`olares-cli profile whoami`" + ` and
` + "`olares-cli settings users me`" + `. Mounted here under the SPA's "Person /
About me" dropdown mental model.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runWhoami(c.Context(), f, refresh, outputRaw)
		},
	}
	cmd.Flags().BoolVar(&refresh, "refresh", false, "force a fresh /api/backend/v1/user-info roundtrip and update the cached role")
	cmd.Flags().StringVarP(&outputRaw, "output", "o", "table", "output format: table, json")
	return cmd
}

func runWhoami(ctx context.Context, f *cmdutil.Factory, refresh bool, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if f == nil {
		return fmt.Errorf("internal error: settings me whoami not wired with cmdutil.Factory")
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
