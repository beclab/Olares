package clusteropts

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"text/tabwriter"
	"time"
)

// Event is the minimal corev1.Event view every cluster verb that lists
// events shares (`cluster pod events`, `cluster job events`,
// `cluster application status`). Fields are limited to what those
// renderers actually consume — adding more is fine when a verb needs
// it, but the wire shape is corev1.Event so any field name from
// kube-apiserver is fair game.
//
// LastTimestamp is the wall-clock the event was last observed; type
// is "Normal" / "Warning"; reason+message together describe what
// happened.
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

// EventSortTime returns the wall-clock to use when ordering events.
// Prefers LastTimestamp; falls back to FirstTimestamp and then
// CreationTimestamp for events the controller hasn't refreshed yet.
// Returns the zero time when nothing is parseable.
func EventSortTime(e Event) time.Time {
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

// SortEventsByLastTimestamp orders events oldest-first using
// EventSortTime as the key. Stable so events with identical
// timestamps keep their server-side order.
func SortEventsByLastTimestamp(events []Event) {
	sort.SliceStable(events, func(i, j int) bool {
		return EventSortTime(events[i]).Before(EventSortTime(events[j]))
	})
}

// RenderEventsTable writes a kubectl-style events table to stdout
// matching the column set used by `cluster pod events` and
// `cluster job events` (LAST SEEN / TYPE / REASON / COUNT / FROM /
// MESSAGE). Callers that want a different column layout (e.g.
// `cluster application status`'s nested table that swaps in OBJECT)
// should render their own — this helper is the canonical pod/job
// shape.
func RenderEventsTable(events []Event, noHeaders bool) error {
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
		// Avoid rendering literal "- ago" when the age is unknown:
		// keep the bare "-" so the column matches every other table.
		lastSeen := Age(ts, now)
		if lastSeen != "-" {
			lastSeen += " ago"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
			lastSeen,
			DashIfEmpty(e.Type),
			DashIfEmpty(e.Reason),
			count,
			DashIfEmpty(from),
			DashIfEmpty(e.Message),
		)
	}
	return nil
}

// EventsListPath builds the K8s-native namespace-scoped events
// list URL with an optional `limit=N` query param. Three callers
// (`cluster pod events`, `cluster job events`, `cluster application
// status`) share this verbatim, so it lives here rather than in
// each callsite.
//
// `limit <= 0` omits the query string entirely (lets the server's
// default page size apply).
func EventsListPath(namespace string, limit int) string {
	p := fmt.Sprintf("/api/v1/namespaces/%s/events", url.PathEscape(namespace))
	if limit > 0 {
		p += "?limit=" + strconv.Itoa(limit)
	}
	return p
}
