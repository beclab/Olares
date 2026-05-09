package job

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewEventsCommand: `olares-cli cluster job events <ns/name | name>
// [-n NS] [--limit N] [-o table|json]`.
//
// Same shape as `cluster pod events`, just with the involvedObject
// kind filter set to Job. Reuses clusteropts.Event for the wire shape
// (events are corev1.Event regardless of the kind being targeted) so
// the type, sort key, table renderer, and URL builder are all shared.
func NewEventsCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace string
		limit     int
	)
	cmd := &cobra.Command{
		Use:   "events <ns/name | name>",
		Short: "show recent events for one Job",
		Long: `Show recent events on the active profile's cluster that target
this Job (involvedObject.kind=Job, name=<job>).

Identity follows the same "<namespace>/<name>" or "-n <ns> <name>"
convention as ` + "`cluster job get`" + `. The server returns every event in
the namespace; we filter client-side to involvedObject.name=<job> so
the output is always job-scoped regardless of server version.

Sorted oldest-first by lastTimestamp (mirrors ` + "`kubectl describe job`" + `'s
Events block) so reading top-to-bottom traces the Job's lifecycle.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ns, name, err := clusteropts.SplitNsName(namespace, args[0])
			if err != nil {
				return err
			}
			return runEvents(c.Context(), o, ns, name, limit)
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional argument is a bare name)")
	cmd.Flags().IntVar(&limit, "limit", 200, "max events to fetch from the namespace before client-side filtering")
	o.AddOutputFlags(cmd)
	return cmd
}

func runEvents(ctx context.Context, o *clusteropts.ClusterOptions, namespace, name string, limit int) error {
	if ctx == nil {
		ctx = context.Background()
	}
	client, err := o.Prepare()
	if err != nil {
		return err
	}
	resp, err := clusterclient.GetK8sList[clusteropts.Event](ctx, client, clusteropts.EventsListPath(namespace, limit))
	if err != nil {
		return fmt.Errorf("list events for job %s/%s: %w", namespace, name, err)
	}

	// Filter by Kind=Job + Name=<name> rather than fieldSelector to
	// stay portable across kube-apiserver versions (older clusters
	// don't expose involvedObject.name as a fieldSelector key).
	var filtered []clusteropts.Event
	for _, e := range resp.Items {
		if e.InvolvedObject.Kind == "Job" && e.InvolvedObject.Name == name {
			filtered = append(filtered, e)
		}
	}
	clusteropts.SortEventsByLastTimestamp(filtered)

	if o.IsJSON() {
		return o.PrintJSON(struct {
			Items []clusteropts.Event `json:"items"`
		}{Items: filtered})
	}
	if o.Quiet {
		return nil
	}
	if len(filtered) == 0 {
		fmt.Fprintf(os.Stderr, "no events for job %s/%s\n", namespace, name)
		return nil
	}
	return clusteropts.RenderEventsTable(filtered, o.NoHeaders)
}
