package users

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// `olares-cli settings users get <username>`
//
// Wraps user-service's GET /api/users/:username. user-service's
// bfl/users.controller.ts:95 explicitly returns `data.data` from the
// upstream axios response — the upstream is app-service
// /app-service/v1/users/<name>, which writes a single UserInfo struct
// directly to the response body (see framework/app-service/.../
// handler_user.go:422 handleUser).
//
// Depending on whether NestJS' global response interceptor is active,
// the wire body shows up either as the raw UserInfo object or wrapped
// in an envelope `{code:200, data:{...}}`. decodeObjectResult
// (cli/cmd/ctl/settings/users/common.go) probes for a top-level `code`
// field and unwraps `data` accordingly, falling back to raw-body decode
// when no envelope is present, so we stay forward-compatible with
// whichever shape user-service settles on.
//
// Role: admin floor. app-service does not gate handleUser server-side
// (any authenticated user gets a successful response for any
// username), but the SPA's "Users" page is admin-only — non-admin
// users have no UI entry into per-user inspection. We mirror that
// here so the CLI surface matches the SPA. A 404 still wins when the
// username doesn't exist on the server.
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
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "get user"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runGet(ctx, f, args[0], output), "get user")
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
	path := userRecordPath(username)
	if err := decodeObjectResult(ctx, pc.Doer, path, &u); err != nil {
		return err
	}

	// Mirror SPA UserInfoPage: only surface a Wizard URL while the user has
	// not yet completed activation (wizard_complete=false) and is not in the
	// terminal Failed state. The URL itself comes from
	// GET /api/users/<name>/status (Address.Wizard), the same source the SPA
	// polls during create.
	wizardLookupErr := ""
	if !u.WizardComplete && !strings.EqualFold(strings.TrimSpace(u.State), "Failed") {
		statusPath := userStatusPath(username)
		var st accountModifyStatus
		if err := decodeObjectResult(ctx, pc.Doer, statusPath, &st); err != nil {
			// Non-fatal: keep printing the user record. Surface the reason
			// on stderr in table mode so the operator knows why the row is
			// "unavailable"; in JSON mode stay silent (the absent
			// wizard_url field already conveys it).
			wizardLookupErr = err.Error()
		} else if host := strings.TrimSpace(st.Address.Wizard); host != "" {
			u.WizardURL = "https://" + strings.TrimPrefix(strings.TrimPrefix(host, "https://"), "http://")
		}
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, u)
	default:
		if wizardLookupErr != "" {
			fmt.Fprintf(os.Stderr, "[get user] wizard URL lookup failed: %s\n", wizardLookupErr)
		}
		return renderUserDetail(os.Stdout, u)
	}
}

func userRecordPath(username string) string {
	return "/api/users/" + url.PathEscape(username)
}

func userStatusPath(username string) string {
	return userRecordPath(username) + "/status"
}

// renderUserDetail prints a 2-column "Field: Value" view rather than the
// list table — single-record output reads better as labeled rows than as
// a 1-row table.
//
// The Wizard URL row mirrors the SPA's UserInfoPage rules: only shown
// while the user is still in onboarding (wizard_complete=false and
// state!=Failed). Once the user activates, the URL is no longer relevant
// and the row is omitted entirely. If we should have a URL but the
// /status lookup yielded nothing, surface a hint rather than a dash so
// the operator knows to retry.
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
	}
	if !u.WizardComplete && !strings.EqualFold(strings.TrimSpace(u.State), "Failed") {
		if u.WizardURL != "" {
			rows = append(rows, [2]string{"Wizard URL", u.WizardURL})
		} else {
			rows = append(rows, [2]string{"Wizard URL", "(unavailable; retry after provisioning finishes)"})
		}
	}
	rows = append(rows,
		[2]string{"Memory Limit", nonEmpty(u.MemoryLimit)},
		[2]string{"CPU Limit", nonEmpty(u.CpuLimit)},
		[2]string{"Created", fmtUserTime(u.CreationTimestamp)},
		[2]string{"Last Login", fmtUserTimePtr(u.LastLoginTime)},
		[2]string{"UID", nonEmpty(u.UID)},
	)
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
