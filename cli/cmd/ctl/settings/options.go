package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// errReported is the canonical "the human already saw the error message,
// don't let cobra print anything else" sentinel — same convention as
// cli/cmd/ctl/market/options.go's errReported. RunE returns this from
// failOp helpers so the parent command suppresses its default "Error:"
// banner without losing the non-zero exit code.
var errReported = errors.New("(already reported)")

// SettingsOptions is the per-command shared option bag for the settings
// umbrella, mirroring cli/cmd/ctl/market/options.go's MarketOptions. Identity
// (--olares-id) and transport (--host) are intentionally absent here: the
// global --profile flag wired through cmdutil.Factory drives both, exactly
// the way `olares-cli files`/`market` resolve identity.
//
// Phase 0a only wires output flags, factory, and a thin prepare() that yields
// a ready-to-use SettingsClient pointed at the desktop ingress. Phase 0b adds
// role-tag plumbing on top of this struct (RoleRequired, NoRoleCheck) and a
// preflight helper. Per-command flags (e.g. --node, --app) get added to this
// struct as the matching verbs land in Phase 1+.
type SettingsOptions struct {
	factory *cmdutil.Factory

	// Output is the response renderer selector — "table" (default) or "json".
	// Reused across every read verb that doesn't have a more specific shape;
	// commands that do (e.g. backup snapshots) are free to add their own
	// flags but should still respect Output for top-level shape selection.
	Output    string
	Quiet     bool
	NoHeaders bool
}

// newSettingsOptions seeds SettingsOptions with the factory the parent
// command was constructed with. Default Output stays "table" to match the
// rest of the CLI (market/files use the same default).
func newSettingsOptions(f *cmdutil.Factory) *SettingsOptions {
	return &SettingsOptions{factory: f, Output: "table"}
}

func (o *SettingsOptions) isJSON() bool {
	return strings.EqualFold(strings.TrimSpace(o.Output), "json")
}

// info prints an informational/diagnostic line to stderr. Suppressed in
// JSON mode (so JSON-consuming scripts get clean stdout) and in --quiet
// mode (so exit-code-only callers get no chatter at all).
func (o *SettingsOptions) info(format string, args ...interface{}) {
	if o.Quiet || o.isJSON() {
		return
	}
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// addOutputFlags wires the standard output trio every read verb gets. Kept
// shape-compatible with MarketOptions.addOutputFlags so help text stays
// consistent across the CLI.
func (o *SettingsOptions) addOutputFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Output, "output", "o", "table", "output format: table, json")
	cmd.Flags().BoolVarP(&o.Quiet, "quiet", "q", false, "suppress output; exit code indicates success/failure")
	cmd.Flags().BoolVar(&o.NoHeaders, "no-headers", false, "omit table headers (useful for scripting)")
}

// prepare resolves the active profile and returns a ready-to-use
// SettingsClient pointed at the desktop ingress (https://desktop.<name>).
// Auth is handled transparently by the Factory's authTransport which injects
// X-Authorization on every request.
//
// Background context is fine here: ResolveProfile reads from the local
// credential store and HTTPClient builds the http.Client lazily. Per-call
// I/O context is set by the run* callers when invoking client methods.
func (o *SettingsOptions) prepare() (*SettingsClient, error) {
	if o.factory == nil {
		return nil, fmt.Errorf("internal error: settings options not wired with cmdutil.Factory")
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
	return NewSettingsClient(hc, rp), nil
}

// printJSON pretty-prints any value as indented JSON to stdout. Used by the
// read verbs when --output json is set; never used in table mode.
func (o *SettingsOptions) printJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}
