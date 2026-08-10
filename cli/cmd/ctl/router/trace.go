package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router trace …` — the spans behind a call.
//
// GET /console/api/observability/traces
// GET /console/api/observability/traces/:trace_id
// GET /console/api/observability/capture-pref
// PUT /console/api/observability/capture-pref
//
// A trace is what an agent framework reports about one exchange: the model call,
// the tool calls around it, the retries. Router accepts them over OTLP and serves
// them back per person — including an admin, who sees their own and nobody
// else's. That is unlike the rest of this tree, and deliberate: a span can carry
// a prompt.
//
// Whether it does is a preference plus a policy. The preference is the person's;
// the policy is the deployment's, and a deployment that forbids capture wins.

type traceSummary struct {
	TraceID   string     `json:"trace_id"`
	StartTS   time.Time  `json:"start_ts"`
	EndTS     *time.Time `json:"end_ts,omitempty"`
	SpanCount int        `json:"span_count"`
	HasError  bool       `json:"has_error"`
}

type traceSpan struct {
	SpanID          string          `json:"span_id"`
	ParentSpanID    *string         `json:"parent_span_id,omitempty"`
	Name            string          `json:"name"`
	Kind            string          `json:"kind"`
	StartTS         time.Time       `json:"start_ts"`
	EndTS           *time.Time      `json:"end_ts,omitempty"`
	DurationMS      *int64          `json:"duration_ms,omitempty"`
	StatusCode      string          `json:"status_code"`
	StatusMessage   *string         `json:"status_message,omitempty"`
	Source          string          `json:"source"`
	Attributes      json.RawMessage `json:"attributes,omitempty"`
	ContentCaptured bool            `json:"content_captured"`
	Content         json.RawMessage `json:"content,omitempty"`
}

type capturePref struct {
	CaptureContent *bool  `json:"capture_content"`
	Effective      bool   `json:"effective"`
	Policy         string `json:"policy"`
}

func NewTraceCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trace",
		Short: "the spans an agent framework reported for a call",
		Long: `Read the traces sent to Router over OTLP.

A trace is one exchange as the calling software saw it — the model call, the tool
calls around it, the retries — which is the view "usage" cannot give: usage has
one row per model call and nothing about what surrounded it.

You see your own traces. So does an admin: a span can carry a prompt, so there is
no route to anybody else's, and an id you did not produce reads as not found.

Whether prompts and completions are stored is two decisions. "trace capture" is
yours; the deployment's policy can forbid it outright, in which case nothing you
set here changes anything and the command says so.

Subcommands:
  list                traces, newest first
  get <trace-id>      the spans in one trace
  capture             show or change whether content is stored
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newTraceListCommand(f))
	cmd.AddCommand(newTraceGetCommand(f))
	cmd.AddCommand(newTraceCaptureCommand(f))
	return cmd
}

func newTraceListCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		since  string
		until  string
		limit  int
		offset int
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "your traces, newest first",
		Long: `List the traces Router holds for you.

ERROR marks a trace with a failed span in it, which is the column to sort your
attention by: a trace is usually only worth opening when something in it went
wrong or took too long.

An empty list is ambiguous in one direction worth knowing — Router only has
traces that something sent it, so software that does not export OTLP produces
none no matter how much it calls.

Examples:
  olares-cli router trace list --since 24h
  olares-cli router trace list --limit 10 -o json
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runTraceList(c.Context(), f, since, until, limit, offset, output)
		},
	}
	cmd.Flags().StringVar(&since, "since", "", "traces starting at or after this time, or a span like 24h")
	cmd.Flags().StringVar(&until, "until", "", "traces starting before this time")
	cmd.Flags().IntVar(&limit, "limit", 50, "how many traces to return (1-1000)")
	cmd.Flags().IntVar(&offset, "offset", 0, "how many traces to skip")
	addOutputFlag(cmd, &output)
	return cmd
}

func runTraceList(ctx context.Context, f *cmdutil.Factory, since, until string, limit, offset int, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	q := url.Values{}
	if s := strings.TrimSpace(since); s != "" {
		when, perr := parseSinceOrInstant(s)
		if perr != nil {
			return fmt.Errorf("--since %w", perr)
		}
		q.Set("since", when.UTC().Format(time.RFC3339))
	}
	if s := strings.TrimSpace(until); s != "" {
		when, perr := parseSinceOrInstant(s)
		if perr != nil {
			return fmt.Errorf("--until %w", perr)
		}
		q.Set("until", when.UTC().Format(time.RFC3339))
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		q.Set("offset", strconv.Itoa(offset))
	}
	var env struct {
		Items  []traceSummary `json:"items"`
		Total  int            `json:"total"`
		Limit  int            `json:"limit"`
		Offset int            `json:"offset"`
	}
	path := consoleAPI + "/observability/traces"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	if err := pc.router.doJSON(ctx, "GET", path, nil, &env); err != nil {
		return traceRouteErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, env)
	}
	return renderTraceList(os.Stdout, env.Items, env.Total, env.Offset)
}

func renderTraceList(w io.Writer, items []traceSummary, total, offset int) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(w, "no traces. Router has only what something exported to it over OTLP, "+
			"so software that does not send traces produces none however much it calls.")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "STARTED\tTOOK\tSPANS\tERROR\tTRACE"); err != nil {
		return err
	}
	for i := range items {
		it := &items[i]
		took := "-"
		if it.EndTS != nil {
			took = fmt.Sprintf("%dms", it.EndTS.Sub(it.StartTS).Milliseconds())
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%d\t%s\t%s\n",
			it.StartTS.Local().Format("2006-01-02 15:04:05"), took, it.SpanCount,
			boolStr(it.HasError), it.TraceID); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	shown := offset + len(items)
	if total > shown {
		_, err := fmt.Fprintf(w, "\nshowing %d-%d of %d; --offset %d for the next page\n", offset+1, shown, total, shown)
		return err
	}
	return nil
}

func newTraceGetCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output      string
		withContent bool
	)
	cmd := &cobra.Command{
		Use:   "get <trace-id>",
		Short: "the spans in one trace",
		Long: `Show a trace as a tree of spans.

Indentation is the parent relation, so what called what is the shape of the
output rather than a column to read. A span's duration and its status are what
locate a problem; the attributes behind it are in the JSON form.

Prompts and completions are shown only with --content, and only when they were
captured at all — which needs both your preference and the deployment's policy to
have allowed it when the call happened. Enabling capture now does not fill in a
trace recorded before.

Examples:
  olares-cli router trace get 4bf92f3577b34da6a3ce929d0e0e4736
  olares-cli router trace get 4bf92f... --content
  olares-cli router trace get 4bf92f... -o json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runTraceGet(c.Context(), f, args[0], withContent, output)
		},
	}
	cmd.Flags().BoolVar(&withContent, "content", false, "include captured prompts and completions")
	addOutputFlag(cmd, &output)
	return cmd
}

func runTraceGet(ctx context.Context, f *cmdutil.Factory, traceID string, withContent bool, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	traceID = strings.TrimSpace(traceID)
	if traceID == "" {
		return fmt.Errorf("a trace id is required; `olares-cli router trace list` has them in the TRACE column")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	var env struct {
		TraceID string      `json:"trace_id"`
		Spans   []traceSpan `json:"spans"`
	}
	path := consoleAPI + "/observability/traces/" + url.PathEscape(traceID)
	if err := pc.router.doJSON(ctx, "GET", path, nil, &env); err != nil {
		return traceRouteErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, env)
	}
	return renderTrace(os.Stdout, env.Spans, withContent)
}

// renderTrace prints spans as a tree. Router returns them in start order, so a
// child always follows its parent and one pass over the slice is enough.
func renderTrace(w io.Writer, spans []traceSpan, withContent bool) error {
	if len(spans) == 0 {
		_, err := fmt.Fprintln(w, "this trace has no spans.")
		return err
	}
	depth := map[string]int{}
	captured := 0
	for i := range spans {
		s := &spans[i]
		d := 0
		if s.ParentSpanID != nil {
			if pd, ok := depth[*s.ParentSpanID]; ok {
				d = pd + 1
			}
		}
		depth[s.SpanID] = d

		took := "-"
		if s.DurationMS != nil {
			took = fmt.Sprintf("%dms", *s.DurationMS)
		}
		status := ""
		if !strings.EqualFold(s.StatusCode, "ok") && !strings.EqualFold(s.StatusCode, "unset") {
			status = "  " + s.StatusCode
			if s.StatusMessage != nil && *s.StatusMessage != "" {
				status += ": " + clip(*s.StatusMessage, 60)
			}
		}
		if _, err := fmt.Fprintf(w, "%s%s  %s%s\n",
			strings.Repeat("  ", d), nonEmpty(s.Name), took, status); err != nil {
			return err
		}
		if s.ContentCaptured {
			captured++
		}
		if withContent && len(s.Content) > 0 {
			if err := printJSON(w, json.RawMessage(s.Content)); err != nil {
				return err
			}
		}
	}
	if !withContent && captured > 0 {
		_, err := fmt.Fprintf(w, "\n%d span(s) carry captured prompts; --content shows them.\n", captured)
		return err
	}
	return nil
}

func newTraceCaptureCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		on     bool
		off    bool
	)
	cmd := &cobra.Command{
		Use:   "capture",
		Short: "show or change whether prompts are stored in your traces",
		Long: `Show, enable or disable capturing prompt and completion content.

Without a flag this reports the state: your preference, the deployment's policy,
and what the two add up to. --on and --off change the preference.

The policy decides what a preference can do. A deployment that forbids capture
refuses to turn it on; one that allows it leaves the choice to you; one that
defaults it off has you opt in. This is a privacy control, so the deployment
having the last word is the point rather than a limitation.

The change applies to what is recorded from now on. It does not add content to
traces already stored, and it does not remove it.

Examples:
  olares-cli router trace capture
  olares-cli router trace capture --on
  olares-cli router trace capture --off
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			if on && off {
				return fmt.Errorf("pass either --on or --off, not both")
			}
			var want *bool
			if on {
				v := true
				want = &v
			}
			if off {
				v := false
				want = &v
			}
			return runTraceCapture(c.Context(), f, want, output)
		},
	}
	cmd.Flags().BoolVar(&on, "on", false, "store prompts and completions in your traces")
	cmd.Flags().BoolVar(&off, "off", false, "stop storing them")
	addOutputFlag(cmd, &output)
	return cmd
}

func runTraceCapture(ctx context.Context, f *cmdutil.Factory, want *bool, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	var pref capturePref
	path := consoleAPI + "/observability/capture-pref"
	if want == nil {
		if err := pc.router.doJSON(ctx, "GET", path, nil, &pref); err != nil {
			return traceRouteErr(err)
		}
	} else {
		body := map[string]bool{"capture_content": *want}
		if err := pc.router.doJSON(ctx, "PUT", path, body, &pref); err != nil {
			return traceRouteErr(err)
		}
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, pref)
	}
	return renderCapturePref(os.Stdout, &pref)
}

// traceRouteErr explains an unmounted subtree. Router mounts these four routes
// only when observability is switched on, so the whole prefix is absent rather
// than empty on a deployment that did not enable it — and a bare 404 would read
// as a missing trace instead of a missing feature.
func traceRouteErr(err error) error {
	if err == nil {
		return nil
	}
	var re *RouterError
	if errors.As(err, &re) && re.Status == 404 && re.Code == "" {
		return fmt.Errorf("this Router has observability switched off, so it stores no traces. " +
			"Turn it on in the deployment's configuration; `olares-cli router usage` works either way, " +
			"with one row per model call rather than the spans around it")
	}
	return err
}

func renderCapturePref(w io.Writer, p *capturePref) error {
	yours := "not set"
	if p.CaptureContent != nil {
		yours = boolStr(*p.CaptureContent)
	}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	rows := [][2]string{
		{"YOUR CHOICE", yours},
		{"DEPLOYMENT POLICY", nonEmpty(p.Policy)},
		{"CONTENT STORED", boolStr(p.Effective)},
	}
	for _, r := range rows {
		if _, err := fmt.Fprintf(tw, "%s\t%s\n", r[0], r[1]); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	note := ""
	switch strings.ToLower(p.Policy) {
	case "forbid":
		note = "This deployment forbids capture, so nothing you set here has any effect."
	case "default_off":
		note = "This deployment leaves capture off unless you ask for it."
	case "allow":
		note = "This deployment allows capture and leaves the choice to you."
	}
	if note == "" {
		return nil
	}
	_, err := fmt.Fprintln(w, "\n"+note+" It applies to traces recorded from now on.")
	return err
}
