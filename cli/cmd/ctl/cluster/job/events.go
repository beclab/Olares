package job

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/pod"
	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewEventsCommand: `olares-cli cluster job events <ns/name | name>
// [-n NS] [--limit N] [-o table|json]`.
//
// Same shape as `cluster pod events`, just with the involvedObject
// kind filter set to Job. Reuses pod.Event for the wire shape (events
// are corev1.Event regardless of the kind being targeted) so we don't
// duplicate the type.
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
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	path := fmt.Sprintf("/api/v1/namespaces/%s/events", url.PathEscape(namespace))
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}
	resp, err := clusterclient.GetK8sList[pod.Event](ctx, client, path)
	if err != nil {
		return fmt.Errorf("list events for job %s/%s: %w", namespace, name, err)
	}

	// Filter by Kind=Job + Name=<name> rather than fieldSelector to
	// stay portable across kube-apiserver versions (older clusters
	// don't expose involvedObject.name as a fieldSelector key).
	var filtered []pod.Event
	for _, e := range resp.Items {
		if e.InvolvedObject.Kind == "Job" && e.InvolvedObject.Name == name {
			filtered = append(filtered, e)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		return eventSortKey(filtered[i]).Before(eventSortKey(filtered[j]))
	})

	if o.IsJSON() {
		return o.PrintJSON(struct {
			Items []pod.Event `json:"items"`
		}{Items: filtered})
	}
	if len(filtered) == 0 {
		fmt.Fprintf(os.Stderr, "no events for job %s/%s\n", namespace, name)
		return nil
	}
	return renderEventsTable(filtered, o.NoHeaders)
}

func eventSortKey(e pod.Event) time.Time {
	for _, ts := range []string{e.LastTimestamp, e.FirstTimestamp, e.Metadata.CreationTimestamp} {
		if ts == "" {
			continue
		}
		if t, err := time.Parse(time.RFC3339, ts); err == nil {
			return t
		}
	}
	return time.Time{}
}

func renderEventsTable(events []pod.Event, noHeaders bool) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()
	if !noHeaders {
		fmt.Fprintln(w, "LAST SEEN\tTYPE\tREASON\tCOUNT\tFROM\tMESSAGE")
	}
	now := time.Now()
	for _, e := range events {
		count := 1
		if e.Count > 0 {
			count = e.Count
		}
		from := e.Source.Component
		if e.Source.Host != "" {
			if from != "" {
				from += "/" + e.Source.Host
			} else {
				from = e.Source.Host
			}
		}
		ts := e.LastTimestamp
		if ts == "" {
			ts = e.Metadata.CreationTimestamp
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
			clusteropts.Age(ts, now)+" ago",
			clusteropts.DashIfEmpty(e.Type),
			clusteropts.DashIfEmpty(e.Reason),
			count,
			clusteropts.DashIfEmpty(from),
			clusteropts.DashIfEmpty(e.Message),
		)
	}
	return nil
}
