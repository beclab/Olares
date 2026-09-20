package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router audit …` — who changed Router, and to what.
//
// GET /console/api/audit-logs
// GET /console/api/audit-logs/:id
//
// Every write to the management plane leaves an entry: a provider created, a
// credential rotated, a key issued, an application archived. The list is a
// slim view, and the before/after states — the part that says what a value
// actually became — are only in the single-entry read, which is why `audit get`
// exists at all.
//
// This is the management plane's history, not the model calls'. What was called
// and what it cost is `router usage`.

type auditEntry struct {
	ID            string          `json:"id"`
	CreatedAt     time.Time       `json:"created_at"`
	RequestID     *string         `json:"request_id,omitempty"`
	ActorUserID   *string         `json:"actor_user_id,omitempty"`
	ActorAppID    *string         `json:"actor_app_id,omitempty"`
	ActorBflName  string          `json:"actor_bfl_name"`
	Action        string          `json:"action"`
	TargetType    string          `json:"target_type"`
	TargetID      *string         `json:"target_id,omitempty"`
	RequestMethod string          `json:"request_method"`
	RequestPath   string          `json:"request_path"`
	ClientIP      *string         `json:"client_ip,omitempty"`
	StatusCode    int             `json:"status_code"`
	ErrorCode     *string         `json:"error_code,omitempty"`
	Note          *string         `json:"note,omitempty"`
	BeforeState   json.RawMessage `json:"before_state,omitempty"`
	AfterState    json.RawMessage `json:"after_state,omitempty"`
}

func NewAuditCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audit",
		Short: "who changed Router, and to what",
		Long: `Read the management plane's history.

Every write leaves an entry — a provider created, a credential rotated, a key
issued, a caller archived — with who did it and whether it worked. A failed
attempt is recorded too, which is what makes this useful for "who tried".

What a value became is not in the list. The before and after states live on the
single entry, so finding a change is "audit list" and understanding it is "audit
get <id>".

Model calls are elsewhere: "olares-cli router usage" is what was called and what
it cost.

Subcommands:
  list         changes, newest first
  get <id>     one change, with what it altered

Admin only.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newAuditListCommand(f))
	cmd.AddCommand(newAuditGetCommand(f))
	return cmd
}

func newAuditListCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output      string
		action      string
		actorRef    string
		targetType  string
		targetID    string
		statusClass string
		since       string
		until       string
		failedOnly  bool
		limit       int
		offset      int
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "changes to Router, newest first",
		Long: `List management-plane changes.

--action matches Router's own action names exactly, such as provider.create or
olares.market.install; the ACTION column is where to read them off.

--failed is the shortcut for everything that did not work, which is usually what
somebody is looking for. --status-class 4xx or 5xx narrows further.

Examples:
  olares-cli router audit list --since 24h
  olares-cli router audit list --actor pptest01 --failed
  olares-cli router audit list --action provider.update -o json
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			if failedOnly && strings.TrimSpace(statusClass) != "" {
				return fmt.Errorf("pass either --failed or --status-class, not both")
			}
			return runAuditList(c.Context(), f, auditFilter{
				Action:      action,
				ActorRef:    actorRef,
				TargetType:  targetType,
				TargetID:    targetID,
				StatusClass: statusClass,
				Since:       since,
				Until:       until,
				FailedOnly:  failedOnly,
				Limit:       limit,
				Offset:      offset,
			}, output)
		},
	}
	cmd.Flags().StringVar(&action, "action", "", "only this action, e.g. provider.create")
	cmd.Flags().StringVar(&actorRef, "actor", "", "only changes by this user, by name or id")
	cmd.Flags().StringVar(&targetType, "target-type", "", "only changes to this kind of thing, e.g. provider")
	cmd.Flags().StringVar(&targetID, "target-id", "", "only changes to this specific thing, by id")
	cmd.Flags().StringVar(&statusClass, "status-class", "", "2xx, 4xx or 5xx")
	cmd.Flags().StringVar(&since, "since", "", "changes at or after this time, or a span like 24h")
	cmd.Flags().StringVar(&until, "until", "", "changes before this time")
	cmd.Flags().BoolVar(&failedOnly, "failed", false, "only changes that were rejected")
	cmd.Flags().IntVar(&limit, "limit", 50, "how many entries to return (1-1000)")
	cmd.Flags().IntVar(&offset, "offset", 0, "how many entries to skip")
	addOutputFlag(cmd, &output)
	return cmd
}

type auditFilter struct {
	Action      string
	ActorRef    string
	TargetType  string
	TargetID    string
	StatusClass string
	Since       string
	Until       string
	FailedOnly  bool
	Limit       int
	Offset      int
}

func runAuditList(ctx context.Context, f *cmdutil.Factory, fl auditFilter, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	if s := strings.ToLower(strings.TrimSpace(fl.StatusClass)); s != "" && s != "2xx" && s != "4xx" && s != "5xx" {
		return fmt.Errorf("--status-class must be 2xx, 4xx or 5xx, not %q", fl.StatusClass)
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	q := url.Values{}
	if s := strings.TrimSpace(fl.Action); s != "" {
		q.Set("action", s)
	}
	if s := strings.TrimSpace(fl.ActorRef); s != "" {
		id, uerr := resolveUserID(ctx, pc, s)
		if uerr != nil {
			return uerr
		}
		q.Set("actor_user_id", id)
	}
	if s := strings.TrimSpace(fl.TargetType); s != "" {
		q.Set("target_type", s)
	}
	if s := strings.TrimSpace(fl.TargetID); s != "" {
		q.Set("target_id", s)
	}
	if s := strings.ToLower(strings.TrimSpace(fl.StatusClass)); s != "" {
		q.Set("status_class", s)
	}
	if s := strings.TrimSpace(fl.Since); s != "" {
		when, perr := parseSinceOrInstant(s)
		if perr != nil {
			return fmt.Errorf("--since %w", perr)
		}
		q.Set("since", when.UTC().Format(time.RFC3339))
	}
	if s := strings.TrimSpace(fl.Until); s != "" {
		when, perr := parseSinceOrInstant(s)
		if perr != nil {
			return fmt.Errorf("--until %w", perr)
		}
		q.Set("until", when.UTC().Format(time.RFC3339))
	}
	if fl.Limit > 0 {
		q.Set("limit", strconv.Itoa(fl.Limit))
	}
	if fl.Offset > 0 {
		q.Set("offset", strconv.Itoa(fl.Offset))
	}

	// Router has one status class per query, so "everything that failed" is two
	// requests joined here rather than a filter it offers.
	var (
		items []auditEntry
		total int
	)
	classes := []string{""}
	if fl.FailedOnly {
		classes = []string{"4xx", "5xx"}
	}
	for _, class := range classes {
		window := q
		if class != "" {
			window = cloneValues(q)
			window.Set("status_class", class)
		}
		var env page[auditEntry]
		if err := pc.router.doJSON(ctx, "GET", withQuery(epAuditLogs, window), nil, &env); err != nil {
			return err
		}
		items = append(items, env.Items...)
		total += env.Total
	}
	merged := len(classes) > 1
	if merged {
		sortEntriesNewestFirst(items)
		if len(items) > fl.Limit && fl.Limit > 0 {
			items = items[:fl.Limit]
		}
	}

	if format == FormatJSON {
		return printJSON(os.Stdout, map[string]any{"items": items, "total": total})
	}
	return renderAuditList(os.Stdout, items, total, fl.Offset, merged)
}

func cloneValues(in url.Values) url.Values {
	out := make(url.Values, len(in))
	for k, v := range in {
		cp := make([]string, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out
}

func sortEntriesNewestFirst(items []auditEntry) {
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j].CreatedAt.After(items[j-1].CreatedAt); j-- {
			items[j], items[j-1] = items[j-1], items[j]
		}
	}
}

func renderAuditList(w io.Writer, items []auditEntry, total, offset int, merged bool) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(w, "no changes match. Router records every management write, so an empty answer "+
			"means nothing was changed in this window rather than that recording is off.")
		return err
	}
	t := newTable(w, "WHEN", "WHO", "ACTION", "TARGET", "RESULT", "ID")
	for i := range items {
		it := &items[i]
		target := it.TargetType
		if it.TargetID != nil && *it.TargetID != "" {
			target = it.TargetType + " " + clip(*it.TargetID, 12)
		}
		result := strconv.Itoa(it.StatusCode)
		if it.ErrorCode != nil && *it.ErrorCode != "" {
			result += " " + *it.ErrorCode
		}
		t.row(
			it.CreatedAt.Local().Format("2006-01-02 15:04:05"),
			nonEmpty(clip(it.ActorBflName, 16)), nonEmpty(it.Action),
			nonEmpty(clip(target, 26)), clip(result, 34), it.ID)
	}
	if err := t.flush(); err != nil {
		return err
	}
	if merged {
		// Client-side rejections and server errors are two queries whose pages
		// cannot be walked as one, so offering a next page would hand out an
		// offset that means something different to each half.
		if _, err := fmt.Fprintf(w, "\n%d rejected changes in total; narrow with --since or "+
			"--status-class 4xx / 5xx to page through them.\n", total); err != nil {
			return err
		}
	} else if err := pageFooter(w, len(items), total, offset); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, "\n`olares-cli router audit get <id>` carries what the change altered.")
	return err
}

func newAuditGetCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "one change, with what it altered",
		Long: `Read one audit entry in full.

This is where the before and after states are: the list cannot carry them, so a
change you want to understand rather than merely find is read here.

Credentials do not appear. Router records that a credential changed and not what
it changed to, so an entry cannot be used to recover a secret — which also means
it cannot tell you which secret is in place.

Examples:
  olares-cli router audit get 0f3a1c62-...
  olares-cli router audit get 0f3a1c62-... -o json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runAuditGet(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runAuditGet(ctx context.Context, f *cmdutil.Factory, id, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("an audit entry id is required; `olares-cli router audit list` has them in the ID column")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	var entry auditEntry
	if err := pc.router.doJSON(ctx, "GET", epAuditLog(id), nil, &entry); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, entry)
	}
	return renderAuditEntry(os.Stdout, &entry)
}

func renderAuditEntry(w io.Writer, e *auditEntry) error {
	t := newTable(w)
	t.row("WHEN", e.CreatedAt.Local().Format("2006-01-02 15:04:05"))
	t.row("WHO", nonEmpty(e.ActorBflName))
	t.row("ACTION", nonEmpty(e.Action))
	t.row("TARGET", nonEmpty(strings.TrimSpace(e.TargetType+" "+derefOr(e.TargetID, ""))))
	t.row("REQUEST", nonEmpty(e.RequestMethod+" "+e.RequestPath))
	t.row("RESULT", strconv.Itoa(e.StatusCode))
	if e.ErrorCode != nil && *e.ErrorCode != "" {
		t.row("ERROR", *e.ErrorCode)
	}
	if e.ClientIP != nil && *e.ClientIP != "" {
		t.row("FROM", *e.ClientIP)
	}
	if e.Note != nil && *e.Note != "" {
		t.row("NOTE", *e.Note)
	}
	if err := t.flush(); err != nil {
		return err
	}
	if err := printState(w, "BEFORE", e.BeforeState); err != nil {
		return err
	}
	// A rejected write changed nothing, and Router records the refusal it sent
	// back in the same field. Calling that "after" would describe the resource as
	// having become an error envelope.
	after := "AFTER"
	if e.StatusCode >= 400 {
		after = "REFUSAL"
	}
	return printState(w, after, e.AfterState)
}

// printState shows a recorded state, and says so when there is none. An absent
// state is meaningful — a create has no before, a rejected write has no after —
// and printing nothing at all would read as a rendering gap.
func printState(w io.Writer, label string, raw json.RawMessage) error {
	if len(raw) == 0 || string(raw) == "null" {
		_, err := fmt.Fprintf(w, "\n%s  (none recorded)\n", label)
		return err
	}
	var pretty any
	if err := json.Unmarshal(raw, &pretty); err != nil {
		_, werr := fmt.Fprintf(w, "\n%s\n%s\n", label, string(raw))
		return werr
	}
	if _, err := fmt.Fprintf(w, "\n%s\n", label); err != nil {
		return err
	}
	return printJSON(w, pretty)
}
