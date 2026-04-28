package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func NewCmdAppList() *cobra.Command {
	opts := &AppOptions{Output: "table"}
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List apps from market sources",
		Long: `List available apps from market sources.

By default the CLI auto-selects a source from market settings. Use -s to choose
one explicitly, or -a to include all sources.

Examples:
  olares-cli app list
  olares-cli app list -s market.olares
  olares-cli app list -a
  olares-cli app list -c AI
  olares-cli app list -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(opts)
		},
	}
	opts.addCommonFlags(cmd)
	opts.addOutputFlags(cmd)
	opts.addAllSourcesFlag(cmd)
	cmd.Flags().StringVarP(&opts.Category, "category", "c", "", "filter by category")
	return cmd
}

func NewCmdAppCategories() *cobra.Command {
	opts := &AppOptions{Output: "table"}
	cmd := &cobra.Command{
		Use:     "categories",
		Aliases: []string{"cats"},
		Short:   "List available app categories",
		Long: `List app categories with counts from market sources.

Examples:
  olares-cli app categories
  olares-cli app categories -a
  olares-cli app categories -o json`,
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

func runList(opts *AppOptions) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("list", "", err)
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

func runListCategories(opts *AppOptions) error {
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
