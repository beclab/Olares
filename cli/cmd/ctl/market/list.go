package market

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewCmdMarketList(f *cmdutil.Factory) *cobra.Command {
	opts := newMarketOptions(f)
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List apps from market sources (catalog by default; --mine for the user's apps)",
		Long: `List apps from market sources.

By default this browses the catalog: the CLI auto-selects a source from market
settings (use -s to override, -a to include every source).

Pass --mine (-m) to instead list the active profile's apps — the same set
the Market UI's "My Terminus" tab shows. "My apps" is broader than
"completed installs": in-flight rows (pending / downloading / installing /
*Canceling / *CancelFailed), post-install transitional rows (upgrading /
resuming / stopping / applyingEnv / uninstalling) and post-install
failures (upgradeFailed / stopFailed / resumeFailed / applyEnvFailed /
uninstallFailed) all surface. Only the SPA's six "uninstalled" states
(uninstalled / installFailed / installingCanceled / downloadFailed /
downloadingCanceled / pendingCanceled) are filtered out, matching the
Market UI exactly.

"Mine" mode differences vs catalog browse:

  - Source scope defaults to "all sources" (no -a needed); pass -s to narrow.
  - Output adds a STATE column showing the live row state from /market/state.
  - Title / version / categories are best-effort enriched from /market/data;
    locally-uploaded charts that no longer appear in the catalog still show
    up but may render with blank title / version.

Examples:
  olares-cli market list
  olares-cli market list -s market.olares
  olares-cli market list -a
  olares-cli market list -c AI
  olares-cli market list -o json
  olares-cli market list --mine
  olares-cli market list --mine -s cli
  olares-cli market list --mine -c AI -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(opts)
		},
	}
	opts.addCommonFlags(cmd)
	opts.addOutputFlags(cmd)
	opts.addAllSourcesFlag(cmd)
	opts.addMineFlag(cmd)
	cmd.Flags().StringVarP(&opts.Category, "category", "c", "", "filter by category")
	return cmd
}

func NewCmdMarketCategories(f *cmdutil.Factory) *cobra.Command {
	opts := newMarketOptions(f)
	cmd := &cobra.Command{
		Use:     "categories",
		Aliases: []string{"cats"},
		Short:   "List available app categories",
		Long: `List app categories with counts from market sources.

Examples:
  olares-cli market categories
  olares-cli market categories -a
  olares-cli market categories -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListCategories(opts)
		},
	}
	opts.addCommonFlags(cmd)
	opts.addOutputFlags(cmd)
	opts.addAllSourcesFlag(cmd)
	return cmd
}

func fetchApps(mc *MarketClient, source string, showAll bool) ([]AppDisplayInfo, error) {
	ctx := context.Background()
	resp, err := mc.GetMarketData(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get market data: %w", err)
	}

	var data MarketDataResponse
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse market data: %w", err)
	}

	if data.UserData == nil {
		return nil, nil
	}

	var apps []AppDisplayInfo
	for sourceName, sourceData := range data.UserData.Sources {
		if !showAll && sourceName != source {
			continue
		}
		if sourceData == nil {
			continue
		}
		for _, item := range sourceData.AppInfoLatest {
			info := extractAppDisplayInfo(item, sourceName)
			if info != nil {
				apps = append(apps, *info)
			}
		}
	}

	sort.Slice(apps, func(i, j int) bool {
		if apps[i].Source == apps[j].Source {
			if apps[i].Name == apps[j].Name {
				return apps[i].Version < apps[j].Version
			}
			return apps[i].Name < apps[j].Name
		}
		return apps[i].Source < apps[j].Source
	})

	return apps, nil
}

func runList(opts *MarketOptions) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("list", "", err)
	}

	if opts.Mine {
		return runListInstalled(opts, mc)
	}

	source := ""
	if !opts.AllSources {
		source = resolveCatalogSource(opts)
		if strings.TrimSpace(opts.Source) == "" {
			opts.info("Using source: %s", source)
		}
	}

	apps, err := fetchApps(mc, source, opts.AllSources)
	if err != nil {
		return opts.failOp("list", "", err)
	}

	category := strings.TrimSpace(opts.Category)
	if category != "" {
		apps = filterByCategory(apps, category)
	}

	if opts.Quiet {
		return nil
	}

	if opts.isJSON() {
		return opts.printJSON(apps)
	}

	if len(apps) == 0 {
		if category != "" {
			fmt.Fprintf(os.Stderr, "No apps found in category '%s'\n", category)
		} else if source != "" {
			fmt.Fprintf(os.Stderr, "No apps found in source '%s'\n", source)
		} else {
			fmt.Fprintln(os.Stderr, "No apps found")
		}
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	if !opts.NoHeaders {
		fmt.Fprintln(w, "NAME\tTITLE\tVERSION\tSOURCE\tCATEGORIES")
	}
	for _, a := range apps {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			a.Name, a.Title, a.Version, a.Source, strings.Join(a.Categories, ", "))
	}
	w.Flush()

	if !opts.NoHeaders {
		fmt.Fprintf(os.Stderr, "\nTotal: %d app(s)\n", len(apps))
	}
	return nil
}

// fetchInstalledApps queries /market/state for the active profile's "My
// Terminus" rows (i.e. every per-user state row except the 6 SPA
// `uninstalledAppStates` — see notInstalledStates in types.go) and
// best-effort enriches each row with title / version / categories from
// /market/data. The catalog fetch is best-effort because a locally
// uploaded chart that has since been deleted from its source can still
// appear in the user's state but no longer in the catalog; we don't
// want that to drop it from the listing.
//
// The function is still named "fetchInstalledApps" rather than
// "fetchMyApps" because internally we use "installed" as shorthand for
// "present in the per-user state map and not in the SPA's
// uninstalledAppStates denylist" — which intentionally includes
// in-flight install rows. `--mine` is just the UX-friendly spelling.
//
// `source` and `showAll` follow the same semantics as fetchApps: when
// showAll is true every source is included; otherwise only the row's
// source name must equal `source`.
func fetchInstalledApps(mc *MarketClient, source string, showAll bool) ([]AppDisplayInfo, error) {
	ctx := context.Background()

	stateResp, err := mc.GetMarketState(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user's market state: %w", err)
	}

	var stateData MarketStateResponse
	if err := json.Unmarshal(stateResp.Data, &stateData); err != nil {
		return nil, fmt.Errorf("failed to parse user's market state: %w", err)
	}

	if stateData.UserData == nil {
		return nil, nil
	}

	catalog := buildCatalogLookup(mc, ctx)

	var apps []AppDisplayInfo
	for sourceName, sourceData := range stateData.UserData.Sources {
		sourceName = strings.TrimSpace(sourceName)
		if sourceName == "" || sourceData == nil {
			continue
		}
		if !showAll && sourceName != source {
			continue
		}
		for _, appState := range sourceData.AppStateLatest {
			if !isInstalledState(appState.Status.State) {
				continue
			}
			name := appState.Status.Name
			if name == "" {
				name = appState.Status.RawName
			}
			if name == "" {
				continue
			}

			entry := AppDisplayInfo{
				Name: name,
				// Version comes from AppStateLatest.Version, which is
				// the version recorded on this per-user state row —
				// the chart the user picked to install / upgrade
				// to (regardless of whether the operation has
				// completed). Falling back to the catalog's "latest
				// available" would silently lie whenever the user is
				// behind upstream, so we deliberately leave Version
				// empty if the state row didn't carry one (older
				// backends, mid-pending rows, edge cases).
				Version: strings.TrimSpace(appState.Version),
				Title:   appState.Status.Title,
				Source:  sourceName,
				State:   appState.Status.State,
			}
			// Catalog enrichment is restricted to title / categories —
			// fields the state row does not carry. Version stays the
			// version on the user's state row above, even if it
			// disagrees with the catalog's latest.
			//
			// Clones get their own per-instance `name` (e.g.
			// `windowsefe992`) but the catalog only knows the source
			// app (`windows`), so a name-based lookup would miss
			// every clone and they'd render with blank categories.
			// `rawAppName`, when non-empty, IS the source app name
			// (see framework/app-service/pkg/utils/app/app.go
			// `GetRawAppName`), so use it as the catalog lookup key
			// whenever present, falling back to `name` for normal
			// (non-clone) installs.
			lookupName := strings.TrimSpace(appState.Status.RawName)
			if lookupName == "" {
				lookupName = name
			}
			if cat, ok := catalog[sourceName+"/"+lookupName]; ok {
				if entry.Title == "" {
					entry.Title = cat.Title
				}
				entry.Categories = cat.Categories
			}
			apps = append(apps, entry)
		}
	}

	sort.Slice(apps, func(i, j int) bool {
		if apps[i].Source == apps[j].Source {
			return apps[i].Name < apps[j].Name
		}
		return apps[i].Source < apps[j].Source
	})

	return apps, nil
}

// buildCatalogLookup pre-aggregates /market/data into a (source/name) →
// AppDisplayInfo map so fetchInstalledApps can enrich each per-user
// state row in O(1). Returns an empty map (never nil) on any failure:
// enrichment is best-effort and we'd rather render rows with blank
// title/version than fail the whole listing because the catalog call
// hiccupped.
func buildCatalogLookup(mc *MarketClient, ctx context.Context) map[string]AppDisplayInfo {
	lookup := map[string]AppDisplayInfo{}
	dataResp, err := mc.GetMarketData(ctx)
	if err != nil {
		return lookup
	}
	var data MarketDataResponse
	if err := json.Unmarshal(dataResp.Data, &data); err != nil {
		return lookup
	}
	if data.UserData == nil {
		return lookup
	}
	for sourceName, sourceData := range data.UserData.Sources {
		if sourceData == nil {
			continue
		}
		for _, item := range sourceData.AppInfoLatest {
			info := extractAppDisplayInfo(item, sourceName)
			if info == nil {
				continue
			}
			lookup[sourceName+"/"+info.Name] = *info
		}
	}
	return lookup
}

// runListInstalled implements `market list --mine`. Source scope
// defaults to "all sources" (so the user sees every one of their apps
// without remembering -a); passing -s narrows to that source. -a stays
// a no-op when nothing else pins the scope. The function name is kept
// as runListInstalled (rather than runListMine) because internally
// "installed" is shorthand for "present in the per-user state map and
// not hidden by the SPA's `uninstalledAppStates`" — which is the same
// set the Market UI's "My Terminus" tab shows. `--mine` is just the
// UX-friendly spelling of that set.
func runListInstalled(opts *MarketOptions, mc *MarketClient) error {
	source := strings.TrimSpace(opts.Source)
	showAll := opts.AllSources || source == ""

	if !showAll {
		opts.info("Filtering my apps by source '%s' (use -a or omit -s for all sources)", source)
	}

	apps, err := fetchInstalledApps(mc, source, showAll)
	if err != nil {
		return opts.failOp("list", "", err)
	}

	category := strings.TrimSpace(opts.Category)
	if category != "" {
		apps = filterByCategory(apps, category)
	}

	if opts.Quiet {
		return nil
	}

	if opts.isJSON() {
		if apps == nil {
			apps = []AppDisplayInfo{}
		}
		return opts.printJSON(apps)
	}

	if len(apps) == 0 {
		switch {
		case category != "":
			fmt.Fprintf(os.Stderr, "No apps in your Market in category '%s'\n", category)
		case !showAll:
			fmt.Fprintf(os.Stderr, "No apps in your Market in source '%s'\n", source)
		default:
			fmt.Fprintln(os.Stderr, "No apps in your Market")
		}
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	if !opts.NoHeaders {
		fmt.Fprintln(w, "NAME\tTITLE\tVERSION\tSTATE\tSOURCE\tCATEGORIES")
	}
	for _, a := range apps {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			a.Name, a.Title, a.Version, a.State, a.Source, strings.Join(a.Categories, ", "))
	}
	w.Flush()

	if !opts.NoHeaders {
		fmt.Fprintf(os.Stderr, "\nTotal: %d app(s) in your Market\n", len(apps))
	}
	return nil
}

func runListCategories(opts *MarketOptions) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("categories", "", err)
	}

	source := ""
	if !opts.AllSources {
		source = resolveCatalogSource(opts)
		if strings.TrimSpace(opts.Source) == "" {
			opts.info("Using source: %s", source)
		}
	}

	apps, err := fetchApps(mc, source, opts.AllSources)
	if err != nil {
		return opts.failOp("categories", "", err)
	}

	counts := map[string]int{}
	for _, a := range apps {
		for _, c := range a.Categories {
			counts[c]++
		}
	}

	if opts.Quiet {
		return nil
	}

	if opts.isJSON() {
		return opts.printJSON(counts)
	}

	if len(counts) == 0 {
		fmt.Fprintln(os.Stderr, "No categories found")
		return nil
	}

	type catRow struct {
		Name  string
		Count int
	}
	var rows []catRow
	for name, count := range counts {
		rows = append(rows, catRow{name, count})
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Name < rows[j].Name
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	if !opts.NoHeaders {
		fmt.Fprintln(w, "CATEGORY\tAPPS")
	}
	for _, r := range rows {
		fmt.Fprintf(w, "%s\t%d\n", r.Name, r.Count)
	}
	w.Flush()
	return nil
}

func filterByCategory(apps []AppDisplayInfo, category string) []AppDisplayInfo {
	lower := strings.ToLower(category)
	var result []AppDisplayInfo
	for _, a := range apps {
		for _, c := range a.Categories {
			if strings.ToLower(c) == lower {
				result = append(result, a)
				break
			}
		}
	}
	return result
}
