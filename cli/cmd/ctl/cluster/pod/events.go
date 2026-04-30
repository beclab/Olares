package pod

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

// Event is the minimal corev1.Event view we care about for the
// rendered table. lastTimestamp is the wall-clock the event was last
// observed; type is "Normal" / "Warning"; reason+message together
// describe what happened.
type Event struct {
	Metadata struct {
		Name              string `json:"name"`
		Namespace         string `json:"namespace,omitempty"`
		CreationTimestamp string `json:"creationTimestamp,omitempty"`
	} `json:"metadata"`
	InvolvedObject struct {
		Kind      string `json:"kind,omitempty"`
		Namespace string `json:"namespace,omitempty"`
		Name      string `json:"name,omitempty"`
		UID       string `json:"uid,omitempty"`
	} `json:"involvedObject"`
	Reason         string `json:"reason,omitempty"`
	Message        string `json:"message,omitempty"`
	Type           string `json:"type,omitempty"`
	Count          int    `json:"count,omitempty"`
	FirstTimestamp string `json:"firstTimestamp,omitempty"`
	LastTimestamp  string `json:"lastTimestamp,omitempty"`
	Source         struct {
		Component string `json:"component,omitempty"`
		Host      string `json:"host,omitempty"`
	} `json:"source,omitempty"`
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

	resp, err := clusterclient.GetK8sList[Event](ctx, client, path)
	if err != nil {
		return fmt.Errorf("list events for pod %s/%s: %w", namespace, name, err)
	}

	// Filter to events whose involvedObject is the named pod. We
	// match on Kind=="Pod" so a service / deployment with the same
	// name in this namespace doesn't confuse the output.
	var filtered []Event
	for _, e := range resp.Items {
		if e.InvolvedObject.Kind == "Pod" && e.InvolvedObject.Name == name {
			filtered = append(filtered, e)
		}
	}
	sortEventsByLastTimestamp(filtered)

	if o.IsJSON() {
		return o.PrintJSON(struct {
			Items []Event `json:"items"`
		}{Items: filtered})
	}

	if len(filtered) == 0 {
		fmt.Fprintf(os.Stderr, "no events for pod %s/%s\n", namespace, name)
		return nil
	}
	return renderEventsTable(filtered, o.NoHeaders)
}

// sortEventsByLastTimestamp orders events oldest-first using
// lastTimestamp (preferred) with creationTimestamp as a fallback for
// events the controller hasn't refreshed yet.
func sortEventsByLastTimestamp(events []Event) {
	sort.SliceStable(events, func(i, j int) bool {
		ti := eventSortKey(events[i])
		tj := eventSortKey(events[j])
		return ti.Before(tj)
	})
}

func eventSortKey(e Event) time.Time {
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

func renderEventsTable(events []Event, noHeaders bool) error {
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
