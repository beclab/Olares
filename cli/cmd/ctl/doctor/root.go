// Package doctor hosts diagnostic checks that combine multiple Olares API
// surfaces without changing cluster or settings state.
package doctor

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/workload"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewDoctorCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run read-only Olares diagnostics",
		Long: `Run read-only diagnostics that combine Olares API surfaces.

Doctor commands do not mutate cluster, settings, market, or files state.`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(workload.NewDoctorImagesCommand(f))
	return cmd
}
