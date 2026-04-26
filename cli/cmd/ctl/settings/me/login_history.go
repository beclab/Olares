package me

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/olares"
)

// `olares-cli settings me login-history`
//
// Wraps user-service's GET /api/users/:username/login-records, which is a
// straight pass-through to BFL IAM:
//   /bfl/iam/v1alpha1/users/:username/login-records
// (framework/bfl/.../handler.go:154 - handleListUserLoginRecords). LLDAP
// stores the underlying records.
//
// Wire shape:
//   user-service strips the BFL envelope server-side (users.controller.ts
//   returns `data.data`), so this endpoint is one of the few that does NOT
//   go through doGetEnvelope. The body we receive is api.ListResult:
//
//     { "items": [LoginRecord, ...], "totals": <int> }
//
//   Each LoginRecord (model.go:35-42):
//     type      string  ("Token")
//     success   bool
//     sourceIP  string
//     user_agent string
//     reason    string
//     login_time *int64 (unix seconds, may be missing on old rows)
//
// Username derivation:
//   The SPA hits /api/users/<username>/login-records using the OlaresID's
//   local part (alice@olares.com -> alice), which BFL then validates
//   against the LLDAP user list. We do the same — derived once from the
//   resolved profile, never asked from the user, so it can't drift.
//
// Role: any authenticated user can read their own login history. BFL's
// handler doesn't inherently forbid reading another user's records, but
// the SPA only ever asks for the current user's, and we follow suit; if
// later phases need cross-user reads we'll add that as an admin-gated
// `settings users login-history <user>` verb under settings/users/.
func NewLoginHistoryCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		limit  int
	)
	cmd := &cobra.Command{
		Use:   "login-history",
		Short: "list recent login attempts for the current user (Settings -> Person -> Login History)",
		Long: `Show recent login attempts (success and failure) for the active
profile's user, sourced from BFL IAM via /api/users/<username>/login-records.

The Olares server returns records in reverse-chronological order; the
table shows the newest first. --limit caps the number of rows shown
(default 50). Pass --output json for the full list including any rows
beyond the cap, with raw unix timestamps for downstream parsing.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runLoginHistory(c.Context(), f, output, limit)
		},
	}
	addOutputFlag(cmd, &output)
	cmd.Flags().IntVar(&limit, "limit", 50, "max rows to print in table mode (0 means no cap; ignored in json mode)")
	return cmd
}

// loginRecord matches framework/bfl/.../v1alpha1/model.go's LoginRecord
// field-for-field. Pointer LoginTime mirrors the source — old rows can
// legitimately be nil, and we want to show "-" rather than "1970-01-01"
// when that happens.
type loginRecord struct {
	Type      string `json:"type"`
	Success   bool   `json:"success"`
	SourceIP  string `json:"sourceIP"`
	UserAgent string `json:"user_agent"`
	Reason    string `json:"reason"`
	LoginTime *int64 `json:"login_time"`
}

// loginHistoryResp is what reaches the wire after user-service unwraps
// the BFL envelope. NewListResult-shaped: items + totals.
type loginHistoryResp struct {
	Items  []loginRecord `json:"items"`
	Totals int           `json:"totals"`
}

func runLoginHistory(ctx context.Context, f *cmdutil.Factory, outputRaw string, limit int) error {
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

	id, err := olares.ParseID(pc.profile.OlaresID)
	if err != nil {
		return fmt.Errorf("derive username from profile olaresId: %w", err)
	}
	path := "/api/users/" + id.Local() + "/login-records"

	// Direct DoJSON: user-service returns the unwrapped IAM list body
	// here, NOT a BFL { code, message, data } envelope (see file header).
	var resp loginHistoryResp
	if err := pc.doer.DoJSON(ctx, "GET", path, nil, &resp); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, resp)
	default:
		return renderLoginHistoryTable(os.Stdout, resp, limit)
	}
}

// renderLoginHistoryTable formats the records as an ASCII table aligned
// via tabwriter. Columns are chosen to match what the SPA's
// LoginHistoryPage shows in column order, minus user_agent (typically
// noisy and not actionable from the CLI).
func renderLoginHistoryTable(w io.Writer, resp loginHistoryResp, limit int) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "TIME\tSTATUS\tSOURCE IP\tREASON"); err != nil {
		return err
	}
	rows := resp.Items
	if limit > 0 && len(rows) > limit {
		rows = rows[:limit]
	}
	for _, r := range rows {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			fmtLoginTime(r.LoginTime),
			loginStatus(r.Success),
			nonEmpty(r.SourceIP),
			singleLine(nonEmpty(r.Reason)),
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	if limit > 0 && resp.Totals > limit {
		fmt.Fprintf(w, "\n(showing %d of %d total rows; pass --limit 0 or --output json for the rest)\n",
			limit, resp.Totals)
	}
	return nil
}

func fmtLoginTime(p *int64) string {
	if p == nil || *p == 0 {
		return "-"
	}
	return time.Unix(*p, 0).Local().Format(time.RFC3339)
}

func loginStatus(ok bool) string {
	if ok {
		return "success"
	}
	return "failure"
}

func singleLine(s string) string {
	// Reasons can include CR/LF (e.g. authelia error stacks). Squash to
	// keep the table from breaking; users who want the raw text can use
	// --output json.
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > 80 {
		s = s[:77] + "..."
	}
	return s
}
