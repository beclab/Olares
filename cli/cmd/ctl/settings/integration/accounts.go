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
//   GET  /api/account/all                   → mini list                (list verb)
//   GET  /api/account/:type                 → mini list per type       (list-by-type verb)
//   POST /api/account/retrieve {name}       → single full record       (get verb)
//
// `GET /api/account/:type` upstream returns
// `[]IntegrationAccountMiniData` (same shape as `/all` but filtered to
// one type/name) — see account.service.ts:88-95
// `getIntegrationAccountByAccountType` → `IntegrationAccountMiniData[]`.
// The SPA hits a *different* endpoint for the full record:
// settings/src/stores/settings/integration.ts:166-174
// `getAccountFullData` → `POST /api/account/retrieve { name: key }`
// where `key = "integration-account:<type>:<name>"`. The CLI mirrors
// that split:
//
//   accounts list                          // GET  /api/account/all
//   accounts list-by-type <type>           // GET  /api/account/<type>     -> []mini
//   accounts get <type> <name>             // POST /api/account/retrieve   -> full
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
		Long: `Inspect and manage external integration accounts that Olares uses to
authenticate against third-party storage / identity providers.

Subcommands:
  list
  list-by-type <type>
  get          <type> <name>
  add          <type> [flags]
  delete       <type> [name]

The "add" verb covers the *direct* object-storage flows (awss3, tencent)
that don't need an OAuth/wallet redirect. The cookie store, OAuth flows
(Google Drive, Dropbox), and Olares-Space NFT cloud-binding stay in the
SPA — they need browser- or wallet-bound state that has no CLI surface.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newAccountsListCommand(f))
	cmd.AddCommand(newAccountsListByTypeCommand(f))
	cmd.AddCommand(newAccountsGetCommand(f))
	cmd.AddCommand(newAccountsAddCommand(f))
	cmd.AddCommand(newAccountsDeleteCommand(f))
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

func newAccountsListByTypeCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list-by-type <type>",
		Short: "list integration accounts of a single type",
		Long: `List the mini records of all integration accounts of the given type.

Use this when you have multiple accounts of the same provider (e.g. two
google drives) and want to see their names + availability before
calling "accounts get <type> <name>" for the full record.

This mirrors the SPA's getAccount(<type>) action, which feeds the
integration list view filtered to one provider. Pass <type> = "all" to
get every type — equivalent to plain "accounts list".
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runAccountsListByType(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func newAccountsGetCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get <type> <name>",
		Short: "show a single integration account's full record",
		Long: `Show a single integration account's full record (including raw_data).

Both positional args are required:
  <type>   account type (e.g. google, dropbox, awss3, tencent, space).
  <name>   account name — disambiguates when a user has multiple
           accounts of the same type.

If you only know the type and want every account of that type, use
"accounts list-by-type <type>" first to read the names, then come back
here for the full record.

This mirrors the SPA's getAccountFullData() action — internally hits
POST /api/account/retrieve { name: "integration-account:<type>:<name>" },
which is the only endpoint that returns the full IntegrationAccount
(including raw_data tokens). GET :type/:name only returns mini metadata.
`,
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			return runAccountsGet(c.Context(), f, args[0], args[1], output)
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

func runAccountsListByType(ctx context.Context, f *cmdutil.Factory, accountType, outputRaw string) error {
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

	var rows []accountMini
	if err := doGetEnvelope(ctx, pc.doer, "/api/account/"+accountType, &rows); err != nil {
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
	if name == "" {
		return fmt.Errorf("account name is required")
	}

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	// SPA settings/src/stores/settings/integration.ts:149-154 builds the
	// store key as `integration-account:<type>:<name>` and POSTs it to
	// /api/account/retrieve to get the full record (incl. raw_data).
	storeKey := "integration-account:" + accountType + ":" + name
	body := map[string]string{"name": storeKey}

	var account accountFull
	if err := doMutateEnvelope(ctx, pc.doer, "POST", "/api/account/retrieve", body, &account); err != nil {
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
