// Package clusteropts hosts the shared option bag used by every
// `cluster ...` verb (top-level + subpackages).
//
// It lives under internal/ for two reasons:
//
//   - The umbrella command (cmd/ctl/cluster/root.go) needs to register
//     subcommands declared in cmd/ctl/cluster/pod/, application/, ...
//     while those subpackages need ClusterOptions and Prepare(); putting
//     the options in the parent package would create an import cycle.
//
//   - Nothing outside the cluster tree should depend on cluster-internal
//     option scaffolding; the internal/ marker enforces that at the
//     compile level.
//
// API surface (exported names) intentionally tracks the SettingsOptions
// shape in cmd/ctl/settings/options.go so the two umbrellas read the
// same way side-by-side.
package clusteropts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// ErrReported is the canonical "the human already saw the error
// message, don't let cobra print anything else" sentinel — mirrors
// cli/cmd/ctl/{market,settings}/options.go errReported. RunE returns
// this from failOp helpers so the parent command suppresses its
// default "Error:" banner without losing the non-zero exit code.
//
// Exported (capital E) so subpackages (cluster/pod, cluster/application,
// ...) can return the same sentinel without re-declaring it.
var ErrReported = errors.New("(already reported)")

// ClusterOptions is the per-command shared option bag for the cluster
// umbrella, mirroring SettingsOptions in
// cli/cmd/ctl/settings/options.go. Identity (--olares-id) and
// transport (--host) are intentionally absent: the global --profile
// flag wired through cmdutil.Factory drives both, exactly the way
// `olares-cli files` / `market` / `settings` resolve identity.
//
// ClusterOptions wires output flags + the factory + a thin Prepare()
// that yields a ready-to-use clusterclient.Client pointed at
// https://control-hub.<terminus>. Per-command flags (e.g. --kind,
// --namespace, --label) hang off per-noun structs that compose this
// one.
//
// Exported (capital P) on Prepare so subpackages (cluster/pod,
// cluster/application, ...) can call it without poking into the
// `cluster` package's internals.
type ClusterOptions struct {
	factory *cmdutil.Factory

	// Output is the response renderer selector — "table" (default) or
	// "json". Reused across every read verb that doesn't have a more
	// specific shape; commands that do (e.g. `cluster pod yaml` which
	// emits YAML) are free to ignore it but should still respect
	// Quiet.
	Output    string
	Quiet     bool
	NoHeaders bool
}

// NewClusterOptions seeds ClusterOptions with the factory the parent
// command was constructed with. Default Output stays "table" to match
// the rest of the CLI.
//
// Exported so subpackages (cluster/pod, cluster/application, ...) can
// build their own per-noun ClusterOptions without re-implementing
// option construction.
func NewClusterOptions(f *cmdutil.Factory) *ClusterOptions {
	return &ClusterOptions{factory: f, Output: "table"}
}

// IsJSON reports whether --output is set to json (case-insensitive).
func (o *ClusterOptions) IsJSON() bool {
	return strings.EqualFold(strings.TrimSpace(o.Output), "json")
}

// Info prints an informational/diagnostic line to stderr. Suppressed
// in JSON and --quiet modes (so JSON-consuming scripts get clean
// stdout and exit-code-only callers get no chatter).
func (o *ClusterOptions) Info(format string, args ...interface{}) {
	if o.Quiet || o.IsJSON() {
		return
	}
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// AddOutputFlags wires the standard output trio every read verb gets.
// Shape-compatible with SettingsOptions.addOutputFlags +
// MarketOptions.addOutputFlags so help text stays consistent.
func (o *ClusterOptions) AddOutputFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Output, "output", "o", "table", "output format: table, json")
	cmd.Flags().BoolVarP(&o.Quiet, "quiet", "q", false, "suppress output; exit code indicates success/failure")
	cmd.Flags().BoolVar(&o.NoHeaders, "no-headers", false, "omit table headers (useful for scripting)")
}

// Prepare resolves the active profile and returns a ready-to-use
// clusterclient.Client pointed at the per-user ControlHub BFF
// (https://control-hub.<terminus>). Auth is handled transparently by
// the Factory's refreshingTransport which injects X-Authorization on
// every request and auto-rotates expired access_tokens via
// /api/refresh.
//
// Background context is fine here: ResolveProfile reads from the
// local credential store and HTTPClient builds the http.Client
// lazily. Per-call I/O context is set by the run* callers when
// invoking client methods.
func (o *ClusterOptions) Prepare() (*clusterclient.Client, error) {
	if o.factory == nil {
		return nil, fmt.Errorf("internal error: cluster options not wired with cmdutil.Factory")
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
	return clusterclient.NewClient(hc, rp), nil
}

// Factory exposes the underlying cmdutil.Factory so subcommands that
// need to ResolveProfile directly (e.g. context.go reads OlaresID for
// the cache lookup) can do so without re-threading the factory
// argument through every constructor.
func (o *ClusterOptions) Factory() *cmdutil.Factory { return o.factory }

// PrintJSON pretty-prints any value as indented JSON to stdout. Used
// by the read verbs when --output json is set; never used in table
// mode.
func (o *ClusterOptions) PrintJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}
