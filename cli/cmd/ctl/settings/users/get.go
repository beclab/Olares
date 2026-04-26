package users

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings users get <username>`
//
// Wraps user-service's GET /api/users/:username. user-service forwards to
// app-service /app-service/v1/users/<name>, which writes a single
// UserInfo struct directly to the response body (NO list wrapper, NO
// app-service ListResult, NO BFL envelope) — see
// framework/app-service/.../handler_user.go:422 (handleUser).
//
// As a result this verb does NOT use decodeListResult or doGetEnvelope —
// it just decodes UserInfo directly into the response body.
//
// Role: app-service does not gate handleUser server-side, so any
// authenticated user can call this for any username (including users
// other than themselves). We don't preflight — let the server be
// authoritative — and surface a clean 404 when the username doesn't
// exist.
func NewGetCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get <username>",
		Short: "show one user's record (Settings -> Users -> <name>)",
		Long: `Show one Olares user's record.

The username is the local part of the olaresId (e.g. "alice" for
"alice@olares.com"). Pass --output json for the full UserInfo struct
(uid, email, terminusName, last_login_time, memory_limit, cpu_limit,
zone, etc.).
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runGet(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runGet(ctx context.Context, f *cmdutil.Factory, username, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if username == "" {
		return fmt.Errorf("username is required")
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	var u userInfo
	path := "/api/users/" + username
	if err := pc.doer.DoJSON(ctx, "GET", path, nil, &u); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, u)
	default:
		return renderUserDetail(os.Stdout, u)
	}
}

// renderUserDetail prints a 2-column "Field: Value" view rather than the
// list table — single-record output reads better as labeled rows than as
// a 1-row table.
func renderUserDetail(w io.Writer, u userInfo) error {
	rows := [][2]string{
		{"Name", nonEmpty(u.Name)},
		{"Display Name", nonEmpty(u.DisplayName)},
		{"Description", nonEmpty(u.Description)},
		{"Email", nonEmpty(u.Email)},
		{"State", nonEmpty(u.State)},
		{"Roles", joinNonEmpty(u.Roles, ",")},
		{"Terminus Name", nonEmpty(u.TerminusName)},
		{"Zone", nonEmpty(u.Zone)},
		{"Wizard Complete", boolStr(u.WizardComplete)},
		{"Memory Limit", nonEmpty(u.MemoryLimit)},
		{"CPU Limit", nonEmpty(u.CpuLimit)},
		{"Created", fmtUserTime(u.CreationTimestamp)},
		{"Last Login", fmtUserTimePtr(u.LastLoginTime)},
		{"UID", nonEmpty(u.UID)},
	}
	for _, r := range rows {
		if _, err := fmt.Fprintf(w, "%-17s %s\n", r[0]+":", r[1]); err != nil {
			return err
		}
	}
	return nil
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func fmtUserTimePtr(p *int64) string {
	if p == nil || *p == 0 {
		return "-"
	}
	return fmtUserTime(*p)
}
