package pod

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewEventsCommand: `olares-cli cluster pod events <ns/name | name>
// [-n ns] [-o table|json]`.
//
// Calls SPA's getEvent
// (apps/packages/app/src/apps/controlPanelCommon/network/index.ts:443):
// `/api/v1/namespaces/<ns>/events`, then filters client-side to just
// the events whose involvedObject is the named pod. The filter is
// done after the fetch (rather than via fieldSelector) because
// kube-apiserver only allows a small set of fieldSelector keys for
// events and `involvedObject.name` isn't always one of them on older
// clusters; client-side filtering covers every server version.
//
// Event type + sort key + table renderer + URL builder all live in
// clusteropts/events.go so `cluster job events` and
// `cluster application status` share one definition.
func NewEventsCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace string
		limit     int
	)

	cmd := &cobra.Command{
		Use:   "events <ns/name | name>",
		Short: "show recent events for one pod",
		Long: `Show recent events on the active profile's cluster that target
this pod (involvedObject.kind=Pod, name=<pod>).

Identity follows the same "<namespace>/<name>" or "-n <ns> <name>"
convention as ` + "`cluster pod get`" + `. The server returns every event in
the namespace; we filter client-side to involvedObject.name=<pod> so
the output is always pod-scoped regardless of server version.

Sorted oldest-first by lastTimestamp (mirrors ` + "`kubectl describe pod`" + `'s
Events block) so reading top-to-bottom traces the pod's lifecycle.
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
		return fmt.Errorf("list events for pod %s/%s: %w", namespace, name, err)
	}

	// Filter to events whose involvedObject is the named pod. We
	// match on Kind=="Pod" so a service / deployment with the same
	// name in this namespace doesn't confuse the output.
	var filtered []clusteropts.Event
	for _, e := range resp.Items {
		if e.InvolvedObject.Kind == "Pod" && e.InvolvedObject.Name == name {
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
		fmt.Fprintf(os.Stderr, "no events for pod %s/%s\n", namespace, name)
		return nil
	}
	return clusteropts.RenderEventsTable(filtered, o.NoHeaders)
}
