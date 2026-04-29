package advanced

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings advanced env ...`
//
// System and user-level environment variables. Mirrors the SPA's
// Settings -> Developer -> System Environment Variables and User Env
// pages (apps/.../api/settings/env.ts:18-35). Two scopes:
//
//	system  /api/env/systemenvs
//	user    /api/env/userenvs
//
// Both are GET (list) + PUT (replace-with-merged-vector). Per the SPA's
// SystemEnvironmentPage.vue rules, system entries that the upstream has
// flagged with editable: false are read-only; the upstream rejects PUTs
// that try to change them. We don't pre-validate that locally because
// the editable flag isn't always populated and we'd rather surface the
// upstream error than block a legitimate write.
//
// Per-app env is at `settings apps env get|set <name>` — this command
// is the *system-wide* surface, not the per-app one.

func NewEnvCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "system + user environment variables (Settings -> Advanced -> Env)",
		Long: `Inspect or change the system-wide / user-wide environment variables
Olares injects into apps. The corresponding per-app surface lives at
"olares-cli settings apps env get|set <name>".

Subcommands:
  system list
  system set --var KEY=VALUE [--var ...]
  user   list
  user   set --var KEY=VALUE [--var ...]

NOTE: "set" requires --var KEY=VALUE flags (repeatable). Bare positional
"KEY=VALUE" is NOT accepted — Cobra would treat the first token as a
sub-verb and report "unknown command".
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newEnvScopeCommand(f, "system", "/api/env/systemenvs"))
	cmd.AddCommand(newEnvScopeCommand(f, "user", "/api/env/userenvs"))
	return cmd
}

func newEnvScopeCommand(f *cmdutil.Factory, scope, basePath string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   scope,
		Short: fmt.Sprintf("%s environment variables (%s)", scope, basePath),
	}
	cmd.AddCommand(newEnvListCommand(f, scope, basePath))
	cmd.AddCommand(newEnvSetCommand(f, scope, basePath))
	return cmd
}

// baseEnv mirrors apps/.../constant/index.ts:1028 BaseEnv. We share
// only the subset we render in the table; --output json marshalls the
// full BFL inner shape verbatim via the json.RawMessage path in
// runEnvList.
type baseEnv struct {
	EnvName     string `json:"envName"`
	Value       string `json:"value,omitempty"`
	Default     string `json:"default,omitempty"`
	Editable    *bool  `json:"editable,omitempty"`
	Type        string `json:"type,omitempty"`
	Required    *bool  `json:"required,omitempty"`
	Description string `json:"description,omitempty"`
}

func newEnvListCommand(f *cmdutil.Factory, scope, basePath string) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("list %s environment variables", scope),
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runEnvList(c.Context(), f, basePath, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runEnvList(ctx context.Context, f *cmdutil.Factory, path, outputRaw string) error {
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
	envs, err := fetchEnv(ctx, pc.doer, path)
	if err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, envs)
	}
	return renderEnvTable(os.Stdout, envs)
}

func fetchEnv(ctx context.Context, d Doer, path string) ([]baseEnv, error) {
	var envs []baseEnv
	if err := doGetEnvelope(ctx, d, path, &envs); err != nil {
		return nil, err
	}
	return envs, nil
}

func renderEnvTable(w io.Writer, envs []baseEnv) error {
	if len(envs) == 0 {
		fmt.Fprintln(w, "no environment variables defined")
		return nil
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tVALUE\tEDITABLE\tTYPE")
	for _, e := range envs {
		editable := "-"
		if e.Editable != nil {
			editable = boolStr(*e.Editable)
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			nonEmpty(e.EnvName),
			nonEmpty(e.Value),
			editable,
			nonEmpty(e.Type),
		)
	}
	return tw.Flush()
}

func newEnvSetCommand(f *cmdutil.Factory, scope, basePath string) *cobra.Command {
	var vars []string
	cmd := &cobra.Command{
		Use:   "set --var KEY=VALUE [--var ...]",
		Short: fmt.Sprintf("update one or more %s env vars (use --var KEY=VALUE; positional KEY=VALUE is NOT accepted)", scope),
		Long: `Update one or more environment variables.

Argument shape: pass each variable as --var KEY=VALUE (repeatable);
the bare form "set KEY=VALUE" is NOT accepted — Cobra would treat the
first positional token as a sub-verb and reject it as "unknown command".

The CLI does a read-modify-write to avoid clobbering values it doesn't
know about: it fetches the current vector, overlays the --var pairs
you pass, and PUTs the merged result back. The upstream rejects writes
to system fields the SPA flags as editable: false.

Examples:
  olares-cli settings advanced env user   set --var FOO=bar
  olares-cli settings advanced env system set --var FOO=bar --var BAZ=qux
`,
		// Args: explicit ArbitraryArgs + a runtime guard so we can give
		// a friendlier error than Cobra's default "unknown command FOO=bar"
		// when a user types `set KEY=VAL` instead of `set --var KEY=VAL`.
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("positional %q is not accepted; use --var %s", args[0], args[0])
			}
			return runEnvSet(c.Context(), f, basePath, scope, vars)
		},
	}
	cmd.Flags().StringArrayVar(&vars, "var", nil, "KEY=VALUE pair (repeatable; required)")
	return cmd
}

func runEnvSet(ctx context.Context, f *cmdutil.Factory, path, scope string, vars []string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	updates, err := parseVarFlags(vars)
	if err != nil {
		return err
	}
	if len(updates) == 0 {
		return fmt.Errorf("env set requires at least one --var KEY=VALUE flag")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	current, err := fetchEnv(ctx, pc.doer, path)
	if err != nil {
		return err
	}
	merged := mergeEnvUpdates(current, updates)
	body := make([]map[string]string, 0, len(merged))
	for _, e := range merged {
		body = append(body, map[string]string{"envName": e.envName, "value": e.value})
	}
	if err := doMutateEnvelope(ctx, pc.doer, "PUT", path, body, nil); err != nil {
		return err
	}
	keys := make([]string, 0, len(updates))
	for k := range updates {
		keys = append(keys, k)
	}
	fmt.Printf("Updated %d %s environment variable(s): %s\n", len(updates), scope, strings.Join(keys, ", "))
	return nil
}

type envPair struct {
	envName string
	value   string
}

func parseVarFlags(raw []string) (map[string]string, error) {
	out := make(map[string]string, len(raw))
	for _, item := range raw {
		idx := strings.IndexByte(item, '=')
		if idx <= 0 {
			return nil, fmt.Errorf("invalid --var %q (expected KEY=VALUE)", item)
		}
		key := strings.TrimSpace(item[:idx])
		val := item[idx+1:]
		if key == "" {
			return nil, fmt.Errorf("invalid --var %q (empty key)", item)
		}
		out[key] = val
	}
	return out, nil
}

func mergeEnvUpdates(current []baseEnv, updates map[string]string) []envPair {
	out := make([]envPair, 0, len(current)+len(updates))
	seen := make(map[string]bool, len(current))
	for _, e := range current {
		v := e.Value
		if up, ok := updates[e.EnvName]; ok {
			v = up
		}
		out = append(out, envPair{envName: e.EnvName, value: v})
		seen[e.EnvName] = true
	}
	for k, v := range updates {
		if !seen[k] {
			out = append(out, envPair{envName: k, value: v})
		}
	}
	return out
}
