package market

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

var errReported = errors.New("(already reported)")

// MarketOptions is the per-command shared option bag. Identity (--user) and
// transport (--host, --kubeconfig) flags from the legacy `app` tree are gone:
// they are replaced by the global `--profile` flag wired through
// cmdutil.Factory, exactly the way `olares-cli files` resolves identity.
type MarketOptions struct {
	factory *cmdutil.Factory

	// Source still varies per command (some accept --source / -s; some don't),
	// so it stays a per-command flag.
	Source string

	Output    string
	Quiet     bool
	NoHeaders bool

	Version        string
	AllSources     bool
	Cascade        bool
	Category       string
	Envs           []string
	EntranceTitles []string
	DeleteData     bool
	Title          string

	// Watch-mode flags. Off by default so today's "fire and forget"
	// scripts keep their current exit semantics; opt in per invocation.
	// Defaults (15m / 2s) match the SPA's effective polling cadence and
	// give enough headroom for the slowest install paths (image pulls)
	// without inviting infinite hangs in CI.
	Watch         bool
	WatchTimeout  time.Duration
	WatchInterval time.Duration
}

// newMarketOptions seeds MarketOptions with the factory the parent command
// was constructed with. Default Output stays "table" to preserve current
// behavior across all subcommands.
func newMarketOptions(f *cmdutil.Factory) *MarketOptions {
	return &MarketOptions{factory: f, Output: "table"}
}

func (o *MarketOptions) isJSON() bool {
	return strings.EqualFold(strings.TrimSpace(o.Output), "json")
}

// info prints an informational message to stderr.
// Suppressed in JSON and quiet modes.
func (o *MarketOptions) info(format string, args ...interface{}) {
	if o.Quiet || o.isJSON() {
		return
	}
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func (o *MarketOptions) addSourceFlag(cmd *cobra.Command, desc string) {
	if desc == "" {
		desc = "market source id (auto-detected when omitted)"
	}
	cmd.Flags().StringVarP(&o.Source, "source", "s", "", desc)
}

// addCommonFlags wires the flags shared by source-aware commands. After
// dropping --user/--host/--kubeconfig (replaced by the global --profile),
// "common" effectively means just the source selector.
func (o *MarketOptions) addCommonFlags(cmd *cobra.Command) {
	o.addSourceFlag(cmd, "")
}

func (o *MarketOptions) addOutputFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Output, "output", "o", "table", "output format: table, json")
	cmd.Flags().BoolVarP(&o.Quiet, "quiet", "q", false, "suppress output; exit code indicates success/failure")
	cmd.Flags().BoolVar(&o.NoHeaders, "no-headers", false, "omit table headers (useful for scripting)")
}

func (o *MarketOptions) addVersionFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Version, "version", "", "app version (default: latest available)")
}

func (o *MarketOptions) addAllSourcesFlag(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&o.AllSources, "all-sources", "a", false, "include apps from all sources")
}

func (o *MarketOptions) addCascadeFlag(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.Cascade, "cascade", false, "apply to all sub-charts (for v2 multi-chart apps)")
}

func (o *MarketOptions) addEnvFlag(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&o.Envs, "env", nil, "set env var in KEY=VALUE format (repeatable)")
}

func (o *MarketOptions) addEntranceTitleFlag(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&o.EntranceTitles, "entrance-title", nil, "set cloned entrance title in NAME=TITLE format (repeatable)")
}

func (o *MarketOptions) addDeleteDataFlag(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.DeleteData, "delete-data", false, "delete persistent data when uninstalling")
}

func (o *MarketOptions) addTitleFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Title, "title", "", "display title for the cloned app instance")
}

// addWatchFlags exposes --watch / --watch-timeout / --watch-interval on
// every lifecycle-mutating verb. They are deliberately attached one-by-one
// (rather than baked into addCommonFlags) because read-only verbs like
// `list` / `get` / `status` have no use for them and we don't want them
// showing up in those help blurbs.
func (o *MarketOptions) addWatchFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&o.Watch, "watch", "w", false,
		"wait until the app reaches a terminal state (success or failure) before exiting")
	cmd.Flags().DurationVar(&o.WatchTimeout, "watch-timeout", 15*time.Minute,
		"maximum total time to wait when --watch is set (e.g. 15m, 1h)")
	cmd.Flags().DurationVar(&o.WatchInterval, "watch-interval", 2*time.Second,
		"polling interval when --watch is set (e.g. 2s, 5s)")
}

// prepare resolves the active profile and returns a ready-to-use MarketClient
// pointed at <MarketURL>/app-store/api/v2. Auth (X-Authorization injection +
// refresh-on-401 retry) is handled transparently by the Factory's
// refreshingTransport.
//
// Background context is fine here: ResolveProfile reads from the local
// credential store and HTTPClient builds the http.Client lazily; neither is a
// long-running call in practice. Per-call I/O context is set by the run*
// callers when invoking client methods.
func (o *MarketOptions) prepare() (*MarketClient, error) {
	if o.factory == nil {
		return nil, fmt.Errorf("internal error: market options not wired with cmdutil.Factory")
	}

	ctx := context.Background()
	rp, err := o.factory.ResolveProfile(ctx)
	if err != nil {
		return nil, err
	}
	hc, err := o.factory.HTTPClient(ctx)
	if err != nil {
		return nil, err
	}
	uploadHC, err := o.factory.HTTPClientWithoutTimeout(ctx)
	if err != nil {
		return nil, err
	}
	return NewMarketClient(hc, uploadHC, rp, strings.TrimSpace(o.Source)), nil
}

func (o *MarketOptions) failOp(op, app string, err error) error {
	if o.Quiet {
		return errReported
	}
	result := OperationResult{
		App:       app,
		Operation: op,
		Status:    "failed",
		Message:   err.Error(),
	}
	o.printResult(result)
	return errReported
}

func (o *MarketOptions) printJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func (o *MarketOptions) printResult(result OperationResult) {
	if o.Quiet {
		return
	}
	if o.isJSON() {
		if err := o.printJSON(result); err != nil {
			fmt.Fprintf(os.Stderr, "failed to encode JSON output: %v\n", err)
		}
		return
	}

	writer := os.Stdout
	appLabel := result.App
	if result.TargetApp != "" && result.TargetApp != result.App {
		appLabel = fmt.Sprintf("%s -> %s", result.App, result.TargetApp)
	}

	message := strings.TrimSpace(result.Message)
	if message == "" {
		switch result.Status {
		case "accepted":
			message = "request accepted"
		case "success":
			message = "completed successfully"
		case "failed":
			message = "request failed"
		default:
			message = "completed"
		}
	}

	if result.Status == "failed" {
		writer = os.Stderr
		fmt.Fprintf(writer, "%s '%s' failed: %s\n", result.Operation, appLabel, message)
	} else {
		fmt.Fprintf(writer, "%s '%s': %s\n", result.Operation, appLabel, message)
	}

	if result.Source != "" {
		fmt.Fprintf(writer, "  source: %s\n", result.Source)
	}
	if result.Version != "" {
		fmt.Fprintf(writer, "  version: %s\n", result.Version)
	}
	if result.State != "" {
		fmt.Fprintf(writer, "  state: %s\n", result.State)
	}
	if result.Progress != "" && result.Progress != "0.00" {
		fmt.Fprintf(writer, "  progress: %s\n", result.Progress)
	}
}
