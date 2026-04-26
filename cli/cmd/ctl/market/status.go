package market

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewCmdMarketStatus(f *cmdutil.Factory) *cobra.Command {
	opts := newMarketOptions(f)
	cmd := &cobra.Command{
		Use:     "status [app-name]",
		Aliases: []string{"stat", "st"},
		Short:   "Show runtime status of installed apps",
		Long: `Show runtime status of installed apps.

If an app name is provided, shows detailed status for that app only.
Without an app name, lists status of all installed apps.

Examples:
  olares-cli market status
  olares-cli market status myapp
  olares-cli market status -a
  olares-cli market status -o json`,
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
	// `status <app> --watch` is the "I forgot to pass --watch on install"
	// recovery path: poll until the app reaches a terminal state without
	// having to re-run status by hand. The flags are also accepted on
	// the all-apps form (no app name), but we explicitly reject that
	// combination in runStatusAll so the error message is actionable.
	opts.addWatchFlags(cmd)
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

// describeOtherSources renders a short summary of where the user does have
// installed apps, used when the active source filter has hidden everything.
// We list distinct source names verbatim when there are at most three of
// them (typical home cluster) and fall back to a count otherwise.
func describeOtherSources(rows []statusRow) string {
	seen := make(map[string]struct{}, len(rows))
	var sources []string
	for _, r := range rows {
		s := strings.TrimSpace(r.Source)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		sources = append(sources, s)
	}
	sort.Strings(sources)

	switch {
	case len(rows) == 1:
		return fmt.Sprintf("1 installed in %q", firstNonEmpty(sources, "another source"))
	case len(sources) == 0:
		return fmt.Sprintf("%d installed in other sources", len(rows))
	case len(sources) <= 3:
		quoted := make([]string, len(sources))
		for i, s := range sources {
			quoted[i] = fmt.Sprintf("%q", s)
		}
		return fmt.Sprintf("%d installed in %s", len(rows), strings.Join(quoted, ", "))
	default:
		return fmt.Sprintf("%d installed across %d other sources", len(rows), len(sources))
	}
}

func firstNonEmpty(values []string, fallback string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return fallback
}

func runStatusAll(opts *MarketOptions) error {
	if opts.Watch {
		// All-apps watch has no obvious terminal: every app may be in a
		// different lifecycle. We require the user to pin a specific
		// app so the wait condition is well-defined.
		return opts.failOp("status", "",
			fmt.Errorf("--watch requires an app name (e.g. 'olares-cli market status <app-name> --watch')"))
	}
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
			opts.info("Filtering installed apps by source '%s' (use -a for all sources)", source)
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
		// If the source filter hid everything, peek at the unfiltered set so
		// we can tell the user "you have N installs, just not in this
		// source" rather than implying nothing is installed at all.
		if source != "" {
			if allRows, parseErr := parseStatusRows(resp, "", true); parseErr == nil && len(allRows) > 0 {
				fmt.Fprintf(os.Stderr, "No installed apps in source '%s' (%s; run with -a to include them)\n",
					source, describeOtherSources(allRows))
				return nil
			}
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

func runStatusSingle(opts *MarketOptions, appName string) error {
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
			opts.info("Filtering installed apps by source '%s' (use -a for all sources)", source)
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

	// When the scoped scan misses, fall back to a global scan so an app
	// installed under a non-default source (e.g. cli/upload/studio) still
	// surfaces. We track whether this fallback fired so the renderer can
	// nudge the user about why the row's SOURCE column differs from the
	// filter they passed.
	fallbackHit := false
	if len(matches) == 0 {
		allRows, parseErr := parseStatusRows(resp, "", true)
		if parseErr == nil {
			for _, row := range allRows {
				if row.Name == appName {
					matches = append(matches, row)
				}
			}
			fallbackHit = len(matches) > 0
		}
	}

	if len(matches) == 0 {
		// Both scans came up empty — the app simply isn't installed.
		// The previous "not found in source 'X'" wording read like a
		// catalog/source-filter problem; this CTA points users at the
		// actual fix instead.
		if source != "" {
			return opts.failOp("status", appName,
				fmt.Errorf("app '%s' is not installed (run 'olares-cli market install %s' to install it)", appName, appName))
		}
		return opts.failOp("status", appName, fmt.Errorf("app '%s' is not installed", appName))
	}

	if fallbackHit && source != "" && matches[0].Source != source {
		opts.info("App is installed under source '%s' (not '%s'); showing that record.", matches[0].Source, source)
	}

	if opts.Watch {
		// Hand off to the watch loop. Watch always pins the first match
		// (status doesn't really make sense across multiple sources for
		// the same app name) and tracks its source so a row that lives
		// outside the default catalog still resolves correctly.
		return runStatusWatch(opts, mc, appName, matches[0])
	}

	if opts.Quiet {
		return nil
	}

	return renderStatusMatches(opts, matches)
}

// renderStatusMatches is the shared output renderer used by both the
// one-shot status path and the post-watch path. Behavior matches the
// pre-refactor code: JSON for `-o json` (single object unless `-a` and
// multiple matches), human-readable detail block otherwise.
func renderStatusMatches(opts *MarketOptions, matches []statusRow) error {
	if len(matches) == 0 {
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
		// parseStatusRows maps empty/0.00 to "-"; the watch path may
		// also synthesize rows with Progress unset, so suppress both.
		if match.Progress != "-" && match.Progress != "" {
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

// runStatusWatch polls the per-user market state until the row reaches a
// terminal classification (per watchStatus) or the deadline / Ctrl-C
// fires, then renders the latest known row through the same path the
// one-shot status command uses. Failure / timeout still render the row so
// JSON consumers see the structured state, but the process exits non-zero
// via errReported.
func runStatusWatch(opts *MarketOptions, mc *MarketClient, appName string, initial statusRow) error {
	if !opts.Quiet && !opts.isJSON() {
		opts.info("Watching '%s' (source '%s', current state '%s') until terminal state (timeout: %s)...",
			appName, initial.Source, initial.State, opts.WatchTimeout)
	}

	target := newWatchTarget(watchStatus, appName, initial.Source)
	finalRow, werr := waitForTerminal(context.Background(), mc, opts, target)

	rowToRender := &initial
	var fail *watchFailureError
	var to *watchTimeoutError
	switch {
	case werr == nil:
		rowToRender = &finalRow
	case errors.As(werr, &fail):
		rowToRender = &fail.row
	case errors.As(werr, &to):
		if to.last != nil {
			rowToRender = to.last
		}
	default:
		// Ctrl-C / context cancel: short-circuit through failOp so
		// users get the standard "operation failed" framing.
		return opts.failOp("status", appName, werr)
	}

	if !opts.Quiet {
		if err := renderStatusMatches(opts, []statusRow{*rowToRender}); err != nil {
			return err
		}
	}
	if werr != nil {
		// failOp would re-render an OperationResult on top of the row
		// we just printed; emit the watcher's message directly to
		// stderr instead so JSON callers still get a clean stdout
		// payload (the row) and humans see the failure detail.
		if !opts.Quiet {
			fmt.Fprintln(os.Stderr, werr.Error())
		}
		return errReported
	}
	return nil
}
