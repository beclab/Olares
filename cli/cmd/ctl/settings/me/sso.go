package me

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

// `olares-cli settings me sso ...`
//
// SSO authorization tokens are the JWTs Authelia issued for this user
// when third-party clients authenticated through Olares' OIDC. The SPA
// surfaces them under Person -> SSOToken; we mirror the same view here
// so users can audit / clean up sessions from the CLI.
//
// Phase 1 ships read only:
//
//   list      -> `olares-cli settings me sso list`
//
// Phase 2 will add:
//
//   revoke    -> `olares-cli settings me sso revoke <token>`
//
// Backend: user-service/src/device2.controller.ts:141 (@Get('/sso')).
// The handler proxies Authelia (authelia-backend-svc:9091/api/usertokens)
// and merges in any matching TermiPass device record. Body shape after
// the BFL envelope unwrap:
//
//   data: [
//     {
//       sso: {
//         expireTime, createTime, tokenType, username, uninitialized,
//         authLevel, firstFactorTimestamp, secondFactorTimestamp
//       },
//       termiPass?: { ...TermiPassDeviceInfo }
//     },
//     ...
//   ]
//
// Role: any authenticated user can list their own tokens; no
// PreflightRole check.

func NewSSOCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sso",
		Short: "SSO authorization tokens (Settings -> Person -> SSOToken)",
		Long: `Inspect SSO authorization tokens issued to this Olares user. These
JWTs back the OIDC/SSO sessions Authelia hands out to third-party clients
and to TermiPass devices bound to this account.

Subcommands:
  list   show all current tokens (Phase 1, read-only)
  revoke revoke a specific token by id  (Phase 2)
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newSSOListCommand(f))
	return cmd
}

// `olares-cli settings me sso list`
func newSSOListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list active SSO authorization tokens for the current user",
		Long: `List SSO authorization tokens (JWTs) currently held for this Olares user.
Each row shows when the token was issued / expires, the username
embedded in the token, the auth-level Authelia recorded, and whether
a TermiPass device is bound to it.

Pass --output json for the full per-token record (including the raw
firstFactorTimestamp / secondFactorTimestamp pair and the bound device
fields the table elides).
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runSSOList(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

// ssoToken mirrors the SSOToken interface in
// apps/packages/app/src/stores/settings/admin.ts (and the server-side
// projection in device2.controller.ts:161-173). Only fields the SPA's
// SSOToken.vue actually shows are surfaced in the table; the rest go
// through to JSON output unchanged.
type ssoToken struct {
	ExpireTime            int64  `json:"expireTime"`
	CreateTime            int64  `json:"createTime"`
	TokenType             string `json:"tokenType"`
	Username              string `json:"username"`
	Uninitialized         string `json:"uninitialized"`
	AuthLevel             int    `json:"authLevel"`
	FirstFactorTimestamp  string `json:"firstFactorTimestamp"`
	SecondFactorTimestamp string `json:"secondFactorTimestamp"`
}

// ssoEntry is one element of the /api/device/sso response array.
// `termiPass` is omitted by the server when no TermiPass device is bound,
// so it has to be a pointer / be allowed to be a missing JSON key.
type ssoEntry struct {
	SSO       ssoToken               `json:"sso"`
	TermiPass map[string]interface{} `json:"termiPass,omitempty"`
}

func runSSOList(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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

	var entries []ssoEntry
	if err := doGetEnvelope(ctx, pc.doer, "/api/device/sso", &entries); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, entries)
	default:
		return renderSSOTable(os.Stdout, entries)
	}
}

func renderSSOTable(w io.Writer, entries []ssoEntry) error {
	if len(entries) == 0 {
		_, err := fmt.Fprintln(w, "no SSO tokens found")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "USERNAME\tTYPE\tCREATED\tEXPIRES\tAUTH LEVEL\tTERMIPASS"); err != nil {
		return err
	}
	for _, e := range entries {
		tp := "-"
		if len(e.TermiPass) > 0 {
			tp = "yes"
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\t%s\n",
			nonEmpty(e.SSO.Username),
			nonEmpty(e.SSO.TokenType),
			fmtSSOTime(e.SSO.CreateTime),
			fmtSSOTime(e.SSO.ExpireTime),
			e.SSO.AuthLevel,
			tp,
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

// fmtSSOTime renders epoch seconds in the user's local timezone, matching
// the SPA's date formatting expectations. Zero / negative values render
// as "-" rather than "1970-01-01" so the table stays readable.
func fmtSSOTime(secs int64) string {
	if secs <= 0 {
		return "-"
	}
	return time.Unix(secs, 0).Local().Format(time.RFC3339)
}
