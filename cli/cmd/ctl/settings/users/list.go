package users

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings users list`
//
// Wraps user-service's GET /api/users/v2 — chosen over /api/users so
// non-privileged callers see themselves rather than a 403. Server flow
// (users.controller.ts:71-93):
//
//   1. proxy app-service /app-service/v1/users (ListResult with all users)
//   2. resolve currentUser by olaresId.split('@')[0]
//   3. if currentUser is owner/admin → return as-is
//      else → return only currentUser
//
// Wire shape (after step 3, what NestJS sends back):
//
//	{ code: 200, data: [UserInfo, ...], totals: N }
//
// We unwrap manually here because the BFL envelope helper in
// settings/me/common.go expects code=0 with a message field. App-service
// uses code=200 and no message.
//
// Role: anyone authenticated. No PreflightRole — the v2 endpoint already
// degrades gracefully for normal users.
func NewListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list Olares users (Settings -> Users)",
		Long: `List Olares users via /api/users/v2.

Owner / admin callers see all users on the instance. Normal users see
only their own row (server-side filtering by user-service). The table
shows the four most useful columns; pass --output json to get the full
UserInfo struct including email, terminusName, memory_limit, cpu_limit,
last_login_time, and zone.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runList(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

// userInfo mirrors app-service's apiserver.UserInfo
// (framework/app-service/pkg/apiserver/handler_user.go:506). We define our
// own copy rather than importing the app-service type because:
//   - app-service is an internal cluster controller, not a stable public
//     library; pulling its types would drag in K8s + Prisma deps the CLI
//     doesn't need.
//   - The CLI only needs JSON-shape compatibility, not the Go-type
//     identity.
type userInfo struct {
	UID               string   `json:"uid"`
	Name              string   `json:"name"`
	DisplayName       string   `json:"display_name"`
	Description       string   `json:"description"`
	Email             string   `json:"email"`
	State             string   `json:"state"`
	LastLoginTime     *int64   `json:"last_login_time"`
	CreationTimestamp int64    `json:"creation_timestamp"`
	Avatar            string   `json:"avatar"`
	Zone              string   `json:"zone"`
	TerminusName      string   `json:"terminusName"`
	WizardComplete    bool     `json:"wizard_complete"`
	Roles             []string `json:"roles"`
	MemoryLimit       string   `json:"memory_limit"`
	CpuLimit          string   `json:"cpu_limit"`
}

func runList(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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

	var resp appServiceListResult[userInfo]
	if err := decodeListResult(ctx, pc.doer, "/api/users/v2", &resp); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, resp.Data)
	default:
		return renderUsersTable(os.Stdout, resp.Data)
	}
}

func renderUsersTable(w io.Writer, users []userInfo) error {
	if len(users) == 0 {
		_, err := fmt.Fprintln(w, "no users")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NAME\tDISPLAY NAME\tROLES\tSTATE\tCREATED"); err != nil {
		return err
	}
	for _, u := range users {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			nonEmpty(u.Name),
			nonEmpty(u.DisplayName),
			joinNonEmpty(u.Roles, ","),
			nonEmpty(u.State),
			fmtUserTime(u.CreationTimestamp),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func nonEmpty(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// joinNonEmpty joins []string with sep, falling back to "-" for empty
// slices. Lifted out so users get / users list both render the role
// list the same way.
func joinNonEmpty(ss []string, sep string) string {
	if len(ss) == 0 {
		return "-"
	}
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += sep
		}
		out += s
	}
	return out
}

// fmtUserTime renders epoch seconds as RFC3339 in local time. UserInfo's
// CreationTimestamp is the K8s creation time forwarded as Unix seconds.
func fmtUserTime(secs int64) string {
	if secs <= 0 {
		return "-"
	}
	return time.Unix(secs, 0).Local().Format(time.RFC3339)
}
