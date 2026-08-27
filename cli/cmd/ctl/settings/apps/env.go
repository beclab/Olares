package apps

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings apps env ...`
//
// Per-app environment variables. Backed by user-service (or its
// upstream env service) at:
//
//	GET /api/env/apps/{appName}/env             -> BaseEnv[] (BFL envelope)
//	PUT /api/env/apps/{appName}/env  body: UpdateEnvItem[]
//
// UpdateEnvItem is just `{envName, value}`. The upstream patches the
// entries it finds by name and leaves the rest of the vector untouched,
// so the PUT carries only what the caller asked to change. Sending the
// whole vector back is worse than redundant: the handler rejects the
// entire request if any named variable is not editable, so a full-vector
// PUT fails on every app that declares a read-only variable.

func NewEnvCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "per-app environment variables (Settings -> Application -> Environment)",
		Long: `Inspect or change the environment variables an installed app sees.

Subcommands:
  get  <name>                         show current env vars
  set  <name> --var KEY=VAL...        update one or more env vars
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newEnvGetCommand(f))
	cmd.AddCommand(newEnvSetCommand(f))
	return cmd
}

// baseEnv mirrors apps/.../constant/index.ts:1028 BaseEnv. Most fields
// are metadata (default / type / required / regex) the SPA renders in
// its form widgets; we surface them in --output json and only render
// envName + current value in the table.
type baseEnv struct {
	EnvName       string      `json:"envName"`
	Value         string      `json:"value,omitempty"`
	Default       string      `json:"default,omitempty"`
	Editable      *bool       `json:"editable,omitempty"`
	Type          string      `json:"type,omitempty"`
	Required      *bool       `json:"required,omitempty"`
	Description   string      `json:"description,omitempty"`
	ApplyOnChange *bool       `json:"applyOnChange,omitempty"`
	ValueFrom     interface{} `json:"valueFrom,omitempty"`
}

func newEnvGetCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "list an installed app's environment variables",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runAppEnvGet(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runAppEnvGet(ctx context.Context, f *cmdutil.Factory, appName, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	appName = strings.TrimSpace(appName)
	if appName == "" {
		return fmt.Errorf("env get requires an app name")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	envs, err := fetchAppEnv(ctx, pc.doer, appName)
	if err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, envs)
	default:
		return renderAppEnvTable(os.Stdout, envs)
	}
}

func fetchAppEnv(ctx context.Context, d Doer, appName string) ([]baseEnv, error) {
	path := "/api/env/apps/" + url.PathEscape(appName) + "/env"
	var envs []baseEnv
	if err := doGetEnvelope(ctx, d, path, &envs); err != nil {
		return nil, err
	}
	return envs, nil
}

func renderAppEnvTable(w io.Writer, envs []baseEnv) error {
	if len(envs) == 0 {
		fmt.Fprintln(w, "no environment variables defined for this app")
		return nil
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tVALUE\tEDITABLE\tREQUIRED\tTYPE")
	for _, e := range envs {
		editable := "-"
		if e.Editable != nil {
			editable = boolStr(*e.Editable)
		}
		required := "-"
		if e.Required != nil {
			required = boolStr(*e.Required)
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			nonEmpty(e.EnvName),
			nonEmpty(e.Value),
			editable,
			required,
			nonEmpty(e.Type),
		)
	}
	return tw.Flush()
}

func newEnvSetCommand(f *cmdutil.Factory) *cobra.Command {
	var vars []string
	cmd := &cobra.Command{
		Use:   "set <name> --var KEY=VALUE [--var ...]",
		Short: "update one or more environment variables on an installed app",
		Long: `Update one or more environment variables on an installed app.

Only the variables you name are sent. The upstream patches the matching
entries and leaves the rest of the vector alone, so there is nothing to
merge client-side.

Two upstream rules to know:
  - only variables the app declares as editable can be written; naming a
    read-only one fails the whole request with a 400
  - a name the app does not declare is ignored rather than created, so
    this command reports which of your keys actually landed

Examples:
  olares-cli settings apps env set my-app \
    --var API_URL=https://api.example.com \
    --var LOG_LEVEL=debug

  # use a literal "=" inside the value
  olares-cli settings apps env set my-app --var "GREETING=hi=there"
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runAppEnvSet(c.Context(), f, args[0], vars)
		},
	}
	cmd.Flags().StringArrayVar(&vars, "var", nil, "KEY=VALUE pair (repeatable)")
	return cmd
}

func runAppEnvSet(ctx context.Context, f *cmdutil.Factory, appName string, vars []string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	appName = strings.TrimSpace(appName)
	if appName == "" {
		return fmt.Errorf("env set requires an app name")
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
	return runAppEnvSetWithDoer(ctx, pc.doer, appName, updates)
}

// runAppEnvSetWithDoer is the wire-level core of `apps env set`, split
// out so tests can assert the request shape without a live cluster.
func runAppEnvSetWithDoer(ctx context.Context, d Doer, appName string, updates map[string]string) error {
	body := envUpdateBody(updates)
	path := "/api/env/apps/" + url.PathEscape(appName) + "/env"
	var after []baseEnv
	if err := doMutateEnvelope(ctx, d, "PUT", path, body, &after); err != nil {
		return err
	}
	applied, ignored := splitAppliedEnvKeys(updates, after)
	if len(applied) > 0 {
		fmt.Printf("Updated %d environment variable(s) on %q: %s\n", len(applied), appName, strings.Join(applied, ", "))
	}
	if len(ignored) > 0 {
		return fmt.Errorf("%q does not declare %s, so %s ignored (this app's variables come from its chart, they cannot be created here)",
			appName, strings.Join(ignored, ", "), plural(len(ignored), "it was", "they were"))
	}
	return nil
}

// envUpdateBody emits UpdateEnvItem[] (constant/index.ts) sorted by name
// so a given set of updates always produces the same request.
func envUpdateBody(updates map[string]string) []map[string]string {
	names := make([]string, 0, len(updates))
	for name := range updates {
		names = append(names, name)
	}
	sort.Strings(names)
	body := make([]map[string]string, 0, len(names))
	for _, name := range names {
		body = append(body, map[string]string{"envName": name, "value": updates[name]})
	}
	return body
}

// splitAppliedEnvKeys sorts the requested names into those the app
// declares (the upstream echoes its full vector back) and those it does
// not, which the upstream drops without saying so.
func splitAppliedEnvKeys(updates map[string]string, after []baseEnv) (applied, ignored []string) {
	declared := make(map[string]bool, len(after))
	for _, e := range after {
		declared[e.EnvName] = true
	}
	for name := range updates {
		if declared[name] {
			applied = append(applied, name)
		} else {
			ignored = append(ignored, name)
		}
	}
	sort.Strings(applied)
	sort.Strings(ignored)
	return applied, ignored
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// parseVarFlags splits "KEY=VALUE" inputs into a map. Values may contain
// "=" themselves (we split on the first "=" only). Empty key is rejected.
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
