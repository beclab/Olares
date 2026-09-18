package advanced

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/userenv"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
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
// Both are GET (list) + PUT (upsert of the entries in the body). Per the
// SPA's SystemEnvironmentPage.vue rules, system entries that the upstream
// has flagged with editable: false are read-only; the upstream rejects
// PUTs that try to change them. We don't pre-validate that locally
// because the editable flag isn't always populated and we'd rather
// surface the upstream error than block a legitimate write.
//
// Per-app env is at `settings apps env get|set <name>` — this command
// is the *system-wide* surface, not the per-app one.

func NewEnvCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "system + user environment variables (Settings -> Advanced -> Env)",
		Long: `Inspect the system-wide / user-wide environment variables
Olares injects into apps. The corresponding per-app surface lives at
"olares-cli settings apps env get <name>".

Subcommands:
  system list
  user   list
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newEnvScopeCommand(f, "system", userenv.SystemEnvsPath))
	cmd.AddCommand(newEnvScopeCommand(f, "user", userenv.UserEnvsPath))
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

// newEnvListCommand returns the read verb for either scope. The SPA's
// System Environment Variables page is reachable by every user (the
// menu entry is not gated on isAdmin) — normal users just see system
// rows with editable=false. We mirror that here: no soft preflight,
// the server stays authoritative.
func newEnvListCommand(f *cmdutil.Factory, scope, basePath string) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("list %s environment variables", scope),
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()
			verb := fmt.Sprintf("list %s env", scope)
			return preflight.Wrap(ctx, f, runEnvList(ctx, f, basePath, output), verb)
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
	envs, err := userenv.List(ctx, pc.doer, path)
	if err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, envs)
	}
	return renderEnvTable(os.Stdout, envs)
}

func renderEnvTable(w io.Writer, envs []userenv.Entry) error {
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

Only the variables you name are sent; the upstream leaves every other
one untouched. It rejects the whole batch if any named variable is one
the SPA flags as editable: false, or if it does not exist yet — this
verb updates existing variables and does not create them.

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
			ctx := c.Context()
			verb := fmt.Sprintf("set %s env", scope)
			// System env is admin-only in the SPA: rows are rendered
			// editable=false for normal users, so the matching
			// scripted verb gets a soft admin gate. User env is
			// per-user (every authenticated caller can write their
			// own bag) and therefore stays ungated.
			if scope == "system" {
				if err := preflight.Gate(ctx, f, whoami.RoleAdmin, verb); err != nil {
					return err
				}
			}
			return preflight.Wrap(ctx, f, runEnvSet(ctx, f, basePath, scope, vars), verb)
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
	if err := userenv.SetValues(ctx, pc.doer, path, updates); err != nil {
		return err
	}
	keys := make([]string, 0, len(updates))
	for k := range updates {
		keys = append(keys, k)
	}
	fmt.Printf("Updated %d %s environment variable(s): %s\n", len(updates), scope, strings.Join(keys, ", "))
	return nil
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
