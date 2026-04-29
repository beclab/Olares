package fan

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
	pkgfan "github.com/beclab/Olares/cli/pkg/dashboard/overview/fan"
)

// NewFanCommand assembles `dashboard overview fan` (root + live +
// curve). cf is the shared *pkgdashboard.CommonFlags pointer
// stored in the area's `common` package var; cobra's persistent-
// flag inheritance from the dashboard root mutates the pointed-at
// struct before any leaf RunE fires.
func NewFanCommand(f *cmdutil.Factory, cf *pkgdashboard.CommonFlags) *cobra.Command {
	common = cf
	cmd := &cobra.Command{
		Use:           "fan",
		Short:         "Sections envelope: live = real-time fan/temperature/power; curve = hardcoded fan-curve spec",
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
			return pkgfan.RunDefault(c.Context(), cli, common)
		},
	}
	cmd.AddCommand(newOverviewFanLiveCommand(f))
	cmd.AddCommand(newOverviewFanCurveCommand(f))
	return cmd
}
