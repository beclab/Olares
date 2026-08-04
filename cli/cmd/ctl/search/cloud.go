package search

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// cloudProvider describes a third-party cloud drive that search3 indexes
// under its own app partition (google_drive / dropbox). The CLI surfaces
// each as its own subcommand, mirroring TermiPass Desktop's independent
// search sources.
type cloudProvider struct {
	// use is the cobra Use name (e.g. "gdrive", "dropbox").
	use string
	// aliases are additional cobra aliases.
	aliases []string
	// title is the human label used in help and empty-result hints.
	title string
	// app is the search3 /api/search/init "app" field.
	app string
	// accountType is the /api/account/<type> segment used to probe whether
	// any integration account is bound (e.g. "google", "dropbox").
	accountType string
	// verb is the full CLI verb rendered in the version-gate error.
	verb string
	// reason is the parenthetical in the version-gate error.
	reason string
}

var (
	googleDriveProvider = cloudProvider{
		use:         "gdrive",
		aliases:     []string{"google-drive", "google"},
		title:       "Google Drive",
		app:         appGoogleDrive,
		accountType: "google",
		verb:        "search gdrive",
		reason:      "search3 app=google_drive index",
	}
	dropboxProvider = cloudProvider{
		use:         "dropbox",
		aliases:     nil,
		title:       "Dropbox",
		app:         appDropbox,
		accountType: "dropbox",
		verb:        "search dropbox",
		reason:      "search3 app=dropbox index",
	}
)

type cloudOptions struct {
	pagingOptions
	searchType string
}

func newGDriveCommand(f *cmdutil.Factory) *cobra.Command {
	return newCloudCommand(f, googleDriveProvider)
}

func newDropboxCommand(f *cmdutil.Factory) *cobra.Command {
	return newCloudCommand(f, dropboxProvider)
}

func newCloudCommand(f *cmdutil.Factory, p cloudProvider) *cobra.Command {
	o := &cloudOptions{}
	cmd := &cobra.Command{
		Use:     p.use + " <keyword>",
		Aliases: p.aliases,
		Short:   "Search " + p.title + " files (Olares >= 1.12.7)",
		Long: fmt.Sprintf(`Search the per-user search3 index for %s files.

Requires Olares >= %s. Uses the same session-based protocol as
`+"`search drive`"+` (/api/search/init + /more + /cancel), with app=%s.

Result locations are files-backend front-end paths (e.g. %s/<account>/...),
so they can be passed directly to `+"`olares-cli files ls`"+` / cat.

If no integration account is bound, the empty result prints a hint to
complete OAuth binding in LarePass → Settings → Integration.

Examples:
  olares-cli search %s report
  olares-cli search %s invoice --limit 50
  olares-cli search %s "design doc" --offset 20 -o json
`, p.title, searchSessionAppMinOlaresVersion, p.app, p.accountType, p.use, p.use, p.use),
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			keyword, err := parseKeyword(args)
			if err != nil {
				return err
			}
			return runCloudSearch(c.Context(), f, keyword, p, o)
		},
	}
	cmd.SilenceUsage = true
	cmd.Flags().StringVarP(&o.searchType, "type", "t", searchTypeAggregate,
		"search mode: aggregate, file_name")
	registerPagingFlags(cmd, &o.pagingOptions)
	return cmd
}

func runCloudSearch(ctx context.Context, f *cmdutil.Factory, keyword string, p cloudProvider, o *cloudOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := requireSessionAppBackendVersion(ctx, f, p.verb, p.reason); err != nil {
		return err
	}
	searchType, err := parseSearchType(o.searchType)
	if err != nil {
		return err
	}
	format, err := o.validate()
	if err != nil {
		return err
	}

	items, err := runSessionSearch(ctx, f, keyword, p.app, searchType, &o.pagingOptions)
	if err != nil {
		return err
	}

	if len(items) == 0 && format == FormatTable {
		return printCloudEmpty(ctx, f, p, os.Stdout)
	}
	return printSearchResults(format, items)
}

// accountMini is the slim IntegrationAccountMiniData shape returned by
// GET /api/account/<type>. Only Available matters for the empty-result hint.
type accountMini struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Available bool   `json:"available"`
}

// printCloudEmpty renders the empty-result message for a cloud search.
// Best-effort: if we can confirm no available integration account is bound,
// print a binding hint; any probe failure falls back to plain "no results".
func printCloudEmpty(ctx context.Context, f *cmdutil.Factory, p cloudProvider, w io.Writer) error {
	msg := cloudEmptyMessage(p, nil, true)
	if doer, err := newDoer(ctx, f); err == nil {
		accounts, probeErr := fetchIntegrationAccounts(ctx, doer, p.accountType)
		msg = cloudEmptyMessage(p, accounts, probeErr != nil)
	}
	_, err := fmt.Fprintln(w, msg)
	return err
}

// cloudEmptyMessage picks the empty-result copy. probeFailed=true means the
// account list could not be fetched, so we stay silent about bindings.
// Pure helper so tests can drive every branch without an HTTP client.
func cloudEmptyMessage(p cloudProvider, accounts []accountMini, probeFailed bool) string {
	if probeFailed {
		return "no results"
	}
	if hasAvailableAccount(accounts) {
		return "no results"
	}
	return fmt.Sprintf(
		"no results — no available %s integration account; bind one in LarePass → Settings → Integration (check with `olares-cli settings integration accounts list-by-type %s`)",
		p.title, p.accountType)
}

func hasAvailableAccount(accounts []accountMini) bool {
	for _, a := range accounts {
		if a.Available {
			return true
		}
	}
	return false
}

func fetchIntegrationAccounts(ctx context.Context, doer *whoami.HTTPClient, accountType string) ([]accountMini, error) {
	var rows []accountMini
	if err := doEnvelope(ctx, doer, "GET", "/api/account/"+accountType, nil, &rows); err != nil {
		return nil, err
	}
	return rows, nil
}
