package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings integration accounts ...`
//
// Backed by user-service's /api/account/* (account.controller.ts):
//
//   GET  /api/account/all                   → mini list  (list verb)
//   GET  /api/account/:type/:name           → single full record (get verb)
//
// The mini list omits raw_data/tokens; the per-account GET returns the
// full IntegrationAccount including raw_data. We never print raw_data
// in --output table, but pass --output json to get it back.
//
// The cookie store and Olares-Space NFT cloud-binding flows live behind
// separate routes that need browser/wallet context, so they stay out
// of CLI scope.
func NewAccountsCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accounts",
		Short: "external integration accounts (S3 / Dropbox / Drive / Tencent / Space)",
		Long: `Inspect external integration accounts that Olares uses to authenticate
against third-party storage / identity providers.

Subcommands:
  list                       list all integration accounts
  get <type> [name]          show a single account (full data)

The cookie store and Olares-Space NFT cloud-binding flows stay in the
SPA — they need browser- or wallet-bound state that has no CLI surface.

Subcommands landing in Phase 2:
  create, delete
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newAccountsListCommand(f))
	cmd.AddCommand(newAccountsGetCommand(f))
	return cmd
}

func newAccountsListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all integration accounts",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runAccountsList(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func newAccountsGetCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get <type> [name]",
		Short: "show a single integration account",
		Long: `Show a single integration account's full record (including raw_data).

The first positional arg is the account type (e.g. google, dropbox, awss3,
tencent, space). The second positional arg is the optional account name —
the SPA uses it to disambiguate when a single user has multiple
accounts of the same type.
`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(c *cobra.Command, args []string) error {
			name := ""
			if len(args) > 1 {
				name = args[1]
			}
			return runAccountsGet(c.Context(), f, args[0], name, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

// accountMini mirrors user-service's IntegrationAccountMiniData.
type accountMini struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	ExpiresAt int64  `json:"expires_at"`
	Available bool   `json:"available"`
	CreateAt  int64  `json:"create_at"`
	Data      string `json:"data,omitempty"`
}

// accountFull mirrors IntegrationAccount + raw_data, but kept loose
// because raw_data subclass shape varies by provider (Google adds
// scope/id_token, Space adds userid, AWSS3 adds endpoint/bucket, ...).
type accountFull struct {
	Name    string          `json:"name"`
	Type    string          `json:"type"`
	RawData json.RawMessage `json:"raw_data"`
}

func runAccountsList(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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

	var rows []accountMini
	if err := doGetEnvelope(ctx, pc.doer, "/api/account/all", &rows); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, rows)
	default:
		return renderAccountsTable(os.Stdout, rows)
	}
}

func runAccountsGet(ctx context.Context, f *cmdutil.Factory, accountType, name, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	if accountType == "" {
		return fmt.Errorf("account type is required")
	}

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	path := "/api/account/" + accountType
	if name != "" {
		path += "/" + name
	}

	var account accountFull
	if err := doGetEnvelope(ctx, pc.doer, path, &account); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, account)
	default:
		return renderAccountDetail(os.Stdout, account)
	}
}

func renderAccountsTable(w io.Writer, rows []accountMini) error {
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, "no integration accounts")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "TYPE\tNAME\tAVAILABLE\tEXPIRES\tCREATED"); err != nil {
		return err
	}
	for _, r := range rows {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			nonEmpty(r.Type),
			nonEmpty(r.Name),
			boolStr(r.Available),
			fmtMillis(r.ExpiresAt),
			fmtMillis(r.CreateAt),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func renderAccountDetail(w io.Writer, a accountFull) error {
	if _, err := fmt.Fprintf(w, "Type:        %s\n", nonEmpty(a.Type)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Name:        %s\n", nonEmpty(a.Name)); err != nil {
		return err
	}
	if len(a.RawData) == 0 {
		_, err := fmt.Fprintln(w, "Raw Data:    (none)")
		return err
	}
	var pretty json.RawMessage = a.RawData
	indent, err := json.MarshalIndent(pretty, "  ", "  ")
	if err != nil {
		// fall back to compact data
		if _, err := fmt.Fprintf(w, "Raw Data:    %s\n", string(a.RawData)); err != nil {
			return err
		}
		return nil
	}
	if _, err := fmt.Fprintln(w, "Raw Data:"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %s\n", string(indent)); err != nil {
		return err
	}
	return nil
}

// fmtMillis renders a unix-millis timestamp as RFC3339, or "-" when
// missing/zero. Account records use millisecond timestamps.
func fmtMillis(ms int64) string {
	if ms <= 0 {
		return "-"
	}
	return time.UnixMilli(ms).UTC().Format(time.RFC3339)
}
