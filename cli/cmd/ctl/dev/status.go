package dev

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/workload"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewStatusCommand: `olares-cli dev status [-n NS] [--json]`.
func NewStatusCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		namespace string
		jsonOut   bool
	)
	cmd := &cobra.Command{
		Use:   "status",
		Short: "list workloads currently running a dev image",
		Long: `Report every workload carrying a ` + workload.PreviousImagesAnnotation + `
annotation — i.e. everything ` + "`dev deploy`" + ` (or
` + "`cluster workload set-image`" + `) has repointed and not yet reverted.

This is the first thing to check when a workload misbehaves after a dev
session. Side-loaded images are especially worth tracking: they exist
only in the node's image store, so an image prune can delete one out
from under a running workload, and with imagePullPolicy: IfNotPresent
there is no registry to recover it from. The symptom is a pod that will
not start with nothing obviously wrong in the chart; this command is
what connects it back to a dev override.

MISSING marks an annotation entry whose container is no longer in the
pod template — usually a chart upgrade that renamed it. Those cannot be
reverted automatically.
`,
		RunE: func(c *cobra.Command, args []string) error {
			overrides, err := workload.ListDevOverrides(c.Context(), f,
				workload.OutputMode{JSON: jsonOut, Quiet: jsonOut}, namespace)
			if err != nil {
				return err
			}
			if jsonOut {
				return printJSON(overrides)
			}
			if len(overrides) == 0 {
				fmt.Fprintln(os.Stdout, "no workloads are running a dev image")
				return nil
			}
			tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			defer tw.Flush()
			fmt.Fprintln(tw, "NAMESPACE\tKIND\tNAME\tCONTAINER\tCURRENT\tORIGINAL")
			for _, o := range overrides {
				current := o.Current
				if o.Missing {
					current = "MISSING"
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
					o.Namespace, o.Kind, o.Name, dashIfEmpty(o.Container),
					dashIfEmpty(current), dashIfEmpty(o.Previous))
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "restrict the scan to one namespace")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit the overrides as JSON")
	return cmd
}

func printJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

var _ = context.Background
