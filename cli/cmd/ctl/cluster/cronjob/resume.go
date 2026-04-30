package cronjob

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewResumeCommand: `olares-cli cluster cronjob resume
// <ns/name | name> [-n NS]`.
//
// PATCHes the CronJob with `{"spec":{"suspend":false}}` using the
// shared runToggle body (same merge-patch+json plumbing as
// `cronjob suspend`). No --yes flag because re-enabling a paused
// schedule is non-destructive — the controller catches up on the
// next tick.
func NewResumeCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var namespace string
	cmd := &cobra.Command{
		Use:   "resume <ns/name | name>",
		Short: "resume one CronJob (set spec.suspend=false)",
		Long: `Resume one CronJob — sets spec.suspend=false via merge-patch+json.

Equivalent to flipping the SPA's "Suspend" toggle off. Reversed by
` + "`cluster cronjob suspend`" + `.

No --yes flag: re-enabling a paused schedule is non-destructive
(the controller catches up on the next tick) so we don't gate it.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ns, name, err := splitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			// assumeYes=true at the call site — see runToggle's
			// "resume" branch (it doesn't prompt anyway, but keeps
			// the API symmetric).
			return runToggle(c.Context(), o, ns, name, false, true)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	o.AddOutputFlags(cmd)
	return cmd
}
