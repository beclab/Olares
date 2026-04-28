package me

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
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
// Phase 1 / 2 verbs:
//
//   list      -> `olares-cli settings me sso list`        (Phase 1)
//   revoke    -> `olares-cli settings me sso revoke <id>` (Phase 2)
//
// Backend: user-service/src/device2.controller.ts. The list handler
// (@Get('/sso')) proxies Authelia (authelia-backend-svc:9091/api/usertokens)
// and merges in any matching TermiPass device record. Body shape after
// the BFL envelope unwrap:
//
//   data: [
//     {
//       sso: {
//         expireTime, createTime, tokenType, username, uninitialized,
//         authLevel,                                  // string, e.g. "one_factor" / "two_factor"
//         firstFactorTimestamp, secondFactorTimestamp
//       },
//       termiPass?: { sso: <token-id>, ...TermiPassDeviceInfo }
//     },
//     ...
//   ]
//
// Wire-shape note (KI-2 in KNOWN_ISSUES.md, fixed 2026-04-28):
//   - authLevel: string (not int as the CLI used to declare). user-service
//     transparently forwards Authelia's raw.authLevel which has always
//     been a string.
//   - firstFactorTimestamp / secondFactorTimestamp: number (not string
//     as the CLI used to declare). Modeled here as int64 to match
//     ExpireTime / CreateTime — they share the epoch-seconds semantics.
//
// The id used by `revoke` is `entry.termiPass.sso` (only tokens with a
// bound TermiPass device can be revoked from the SPA either; the SPA
// hides the delete icon when termiPass is missing).
//
// Role: any authenticated user can list / revoke their own tokens; no
// PreflightRole check.

func NewSSOCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sso",
		Short: "SSO authorization tokens (Settings -> Person -> SSOToken)",
		Long: `Inspect SSO authorization tokens issued to this Olares user. These
JWTs back the OIDC/SSO sessions Authelia hands out to third-party clients
and to TermiPass devices bound to this account.

Subcommands:
  list                            list active SSO tokens          (Phase 1)
  revoke <id>                     revoke a TermiPass-bound token  (Phase 2)
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newSSOListCommand(f))
	cmd.AddCommand(newSSORevokeCommand(f))
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
// settings/src/stores/settings/admin.ts (and the server-side
// projection in device2.controller.ts:161-173). Only fields the SPA's
// SSOToken.vue actually shows are surfaced in the table; the rest go
// through to JSON output unchanged.
type ssoToken struct {
	ExpireTime            int64  `json:"expireTime"`
	CreateTime            int64  `json:"createTime"`
	TokenType             string `json:"tokenType"`
	Username              string `json:"username"`
	Uninitialized         string `json:"uninitialized"`
	AuthLevel             string `json:"authLevel"`             // string, e.g. "one_factor" / "two_factor", mirrors Authelia raw payload
	FirstFactorTimestamp  int64  `json:"firstFactorTimestamp"`  // epoch seconds, mirrors Authelia raw.firstFactorTimestamp (number)
	SecondFactorTimestamp int64  `json:"secondFactorTimestamp"` // epoch seconds, mirrors Authelia raw.secondFactorTimestamp (number)
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
	if _, err := fmt.Fprintln(tw, "ID\tUSERNAME\tTYPE\tCREATED\tEXPIRES\tAUTH LEVEL\tTERMIPASS"); err != nil {
		return err
	}
	for _, e := range entries {
		id, hasTP := termiPassSSOID(e.TermiPass)
		tp := "-"
		idCell := "-"
		if hasTP {
			tp = "yes"
			if id != "" {
				idCell = id
			}
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			idCell,
			nonEmpty(e.SSO.Username),
			nonEmpty(e.SSO.TokenType),
			fmtSSOTime(e.SSO.CreateTime),
			fmtSSOTime(e.SSO.ExpireTime),
			nonEmpty(e.SSO.AuthLevel),
			tp,
		); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "\nuse `olares-cli settings me sso revoke <id>` to revoke a TermiPass-bound token (only rows with a non-`-` ID can be revoked)"); err != nil {
		return err
	}
	return nil
}

// termiPassSSOID extracts the `termiPass.sso` field — that's the
// Authelia-issued token id the revoke endpoint expects (mirrors
// `${tokenStore.url}/api/device/sso/${token.termiPass.sso}` from
// stores/settings/admin.ts revoke_token).
func termiPassSSOID(tp map[string]interface{}) (string, bool) {
	if len(tp) == 0 {
		return "", false
	}
	raw, ok := tp["sso"]
	if !ok {
		return "", true
	}
	if s, ok := raw.(string); ok {
		return s, true
	}
	return fmt.Sprint(raw), true
}

// `olares-cli settings me sso revoke <id>`
//
// The <id> is the value rendered in the ID column of `me sso list` — it
// originates from `termiPass.sso` on the list response. Tokens without a
// bound TermiPass device cannot be revoked through this endpoint (the
// SPA's SSOToken.vue hides the delete icon for those rows; the server
// returns 404 if you try anyway).
func newSSORevokeCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke <id>",
		Short: "revoke a TermiPass-bound SSO authorization token",
		Long: `Revoke a single SSO authorization token by id.

The id comes from the ID column of "settings me sso list" (which itself
mirrors the SPA's Settings -> Person -> SSOToken page). Only tokens with
a bound TermiPass device can be revoked through this endpoint; rows that
list "-" in the TERMIPASS column have no revocable id.

The current Olares CLI session is unaffected by revoking other devices'
tokens, but revoking a token bound to your active TermiPass session may
log it out — re-authenticate with the TermiPass app afterwards if so.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runSSORevoke(c.Context(), f, args[0])
		},
	}
	return cmd
}

func runSSORevoke(ctx context.Context, f *cmdutil.Factory, tokenID string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	tokenID = strings.TrimSpace(tokenID)
	if tokenID == "" || tokenID == "-" {
		return fmt.Errorf("revoke requires a non-empty token id (run `olares-cli settings me sso list` and copy the ID column)")
	}

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	path := "/api/device/sso/" + url.PathEscape(tokenID)
	if err := pc.doer.DoJSON(ctx, "DELETE", path, nil, nil); err != nil {
		return err
	}
	fmt.Printf("Revoked SSO token %q.\n", tokenID)
	return nil
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
