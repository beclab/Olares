package applications

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
	pkgapps "github.com/beclab/Olares/cli/pkg/dashboard/applications"
)

// ----------------------------------------------------------------------------
// `dashboard applications` — workload-grain table (single leaf, no subverbs).
// ----------------------------------------------------------------------------
//
// Default action: workload-grain table — same data source as `overview
// ranking` (FetchWorkloadsMetrics) with the `state` and `pods` columns
// added. The SPA's Applications2/IndexPage renders this view; we mirror
// the row shape so consumers can join `applications` and `overview
// ranking` on the (app, namespace) tuple.
//
// The deprecated `applications list / users / containers / pods`
// leaves are gone: `list` is now the default action; `users` was
// admin-only and rarely useful (the same data shows up in `overview
// user --user`); `containers` was a stub that nobody depended on;
// `pods` duplicated `kubectl get pods -n <ns>` and never carried
// first-class agent semantics.
//
// All business logic lives in cli/pkg/dashboard/applications/. This
// file is pure cobra wiring: register the leaf, bind its private
// flags, and forward to pkgapps.RunList from RunE. Per the
// thin-cmd / pkg-business split documented in
// cli/skills/olares-dashboard/SKILL.md.

func NewApplicationsCommand(f *cmdutil.Factory, cf *pkgdashboard.CommonFlags) *cobra.Command {
	common = cf
	var sortDir, sortBy string
	cmd := &cobra.Command{
		Use:           "applications",
		Aliases:       []string{"apps"},
		Short:         "Workload-grain application table (mirrors the SPA's Applications page)",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				return unknownSubcommandRunE(c, args)
			}
			if err := common.Validate(); err != nil {
				return err
			}
			cli, err := prepareClient(c.Context(), f)
			if err != nil {
				return err
			}
			return pkgapps.RunList(c.Context(), cli, common, sortBy, sortDir)
		},
	}
	cmd.Flags().StringVar(&sortDir, "sort", "desc", "sort direction (asc or desc)")
	cmd.Flags().StringVar(&sortBy, "sort-by", "cpu", "sort key: cpu | memory | net_in | net_out")
	return cmd
}
