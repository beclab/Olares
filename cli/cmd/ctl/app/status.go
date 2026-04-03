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

func NewCmdAppStatus() *cobra.Command {
	opts := &AppOptions{Output: "table"}
	cmd := &cobra.Command{
		Use:     "status [app-name]",
		Aliases: []string{"stat", "st"},
		Short:   "Show runtime status of installed apps",
		Long: `Show runtime status of installed apps.

If an app name is provided, shows detailed status for that app only.
Without an app name, lists status of all installed apps.

Examples:
  olares-cli app status
  olares-cli app status myapp
  olares-cli app status -a
  olares-cli app status -o json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return runStatusSingle(opts, args[0])
			}
			return runStatusAll(opts)
		},
	}
	opts.addCommonFlags(cmd)
	opts.addOutputFlags(cmd)
	opts.addAllSourcesFlag(cmd)
	return cmd
}

type statusRow struct {
	Name     string `json:"name"`
	State    string `json:"state"`
	OpType   string `json:"opType,omitempty"`
	Progress string `json:"progress,omitempty"`
	CfgType  string `json:"cfgType,omitempty"`
	Message  string `json:"message,omitempty"`
	Source   string `json:"source"`
}

func parseStatusRows(resp *APIResponse, source string, showAll bool) ([]statusRow, error) {
	var data MarketStateResponse
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse state data: %w", err)
	}

	if data.UserData == nil {
		return nil, nil
	}

	var rows []statusRow
	for sourceName, sourceData := range data.UserData.Sources {
		sourceName = strings.TrimSpace(sourceName)
		if sourceName == "" {
			continue
		}
		if sourceData == nil {
			continue
		}
		for _, appState := range sourceData.AppStateLatest {
			name := appState.Status.Name
			if name == "" {
				name = appState.Status.RawName
			}
			if name == "" {
				continue
			}
			if !showAll && sourceName != source {
				continue
			}
			progress := appState.Status.Progress
			if progress == "" || progress == "0.00" {
				progress = "-"
			}
			rows = append(rows, statusRow{
				Name:     name,
				State:    appState.Status.State,
				OpType:   appState.Status.OpType,
				Progress: progress,
				CfgType:  appState.Status.CfgType,
				Message:  appState.Status.Message,
				Source:   sourceName,
			})
		}
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Source == rows[j].Source {
			return rows[i].Name < rows[j].Name
		}
		return rows[i].Source < rows[j].Source
	})

	return rows, nil
}

func runStatusAll(opts *AppOptions) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("status", "", err)
	}

	ctx := context.Background()
	resp, err := mc.GetMarketState(ctx)
	if err != nil {
		return opts.failOp("status", "", fmt.Errorf("failed to get app status: %w", err))
	}

	source := ""
	if !opts.AllSources {
		source = resolveCatalogSource(opts)
		if strings.TrimSpace(opts.Source) == "" {
			opts.info("Using source: %s", source)
		}
	}

	rows, err := parseStatusRows(resp, source, opts.AllSources)
	if err != nil {
		return opts.failOp("status", "", err)
	}

	if opts.Quiet {
		return nil
	}

	if len(rows) == 0 {
		if opts.isJSON() {
			return opts.printJSON([]statusRow{})
		}
		if source != "" {
			fmt.Fprintf(os.Stderr, "No installed apps found in source '%s'\n", source)
		} else {
			fmt.Fprintln(os.Stderr, "No installed apps found")
		}
		return nil
	}

	if opts.isJSON() {
		return opts.printJSON(rows)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	if !opts.NoHeaders {
		fmt.Fprintln(w, "NAME\tSTATE\tOPERATION\tPROGRESS\tSOURCE")
	}
	for _, r := range rows {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", r.Name, r.State, r.OpType, r.Progress, r.Source)
	}
	w.Flush()
	return nil
}

func runStatusSingle(opts *AppOptions, appName string) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("status", appName, err)
	}

	ctx := context.Background()
	resp, err := mc.GetMarketState(ctx)
	if err != nil {
		return opts.failOp("status", appName, fmt.Errorf("failed to get app status: %w", err))
	}

	source := ""
	if !opts.AllSources {
		source = resolveCatalogSource(opts)
		if strings.TrimSpace(opts.Source) == "" {
			opts.info("Using source: %s", source)
		}
	}

	rows, err := parseStatusRows(resp, source, opts.AllSources)
	if err != nil {
		return opts.failOp("status", appName, err)
	}

	var matches []statusRow
	for _, row := range rows {
		if row.Name == appName {
			matches = append(matches, row)
		}
	}

	if len(matches) == 0 {
		allRows, parseErr := parseStatusRows(resp, "", true)
		if parseErr == nil {
			for _, row := range allRows {
				if row.Name == appName {
					matches = append(matches, row)
				}
			}
		}
	}

	if len(matches) == 0 {
		if source != "" {
			return opts.failOp("status", appName, fmt.Errorf("app '%s' not found in source '%s'", appName, source))
		}
		return opts.failOp("status", appName, fmt.Errorf("app '%s' not found", appName))
	}

	if opts.Quiet {
		return nil
	}

	if opts.isJSON() {
		if opts.AllSources && len(matches) > 1 {
			return opts.printJSON(matches)
		}
		return opts.printJSON(matches[0])
	}

	for idx, match := range matches {
		if idx > 0 {
			fmt.Println()
		}
		fmt.Printf("App:        %s\n", match.Name)
		fmt.Printf("Source:     %s\n", match.Source)
		fmt.Printf("State:      %s\n", match.State)
		if match.OpType != "" {
			fmt.Printf("Operation:  %s\n", match.OpType)
		}
		if match.Progress != "-" {
			fmt.Printf("Progress:   %s\n", match.Progress)
		}
		if match.Message != "" {
			fmt.Printf("Message:    %s\n", match.Message)
		}
		if !opts.AllSources {
			break
		}
	}
	return nil
}
