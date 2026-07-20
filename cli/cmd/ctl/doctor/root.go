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
		Long: `Run Olares diagnostics that combine API surfaces.

Most doctor commands are read-only. Exceptions that mutate (when an
explicit flag is set) are documented on the subcommand help.`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(workload.NewDoctorImagesCommand(f))
	cmd.AddCommand(NewThirdLevelDomainCommand(f))
	return cmd
}
