package apps

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
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
// UpdateEnvItem is just `{envName, value}` — the SPA strips the
// definition fields and sends back only the changed values, all in one
// PUT (the upstream replaces the entire vector each call). To match
// that semantic safely from the CLI we:
//
//   - fetch the current env first with the same GET,
//   - merge user-supplied --var KEY=VALUE pairs on top, and
//   - PUT the merged set.
//
// Without the read-modify-write merge, a CLI user setting one variable
// would clobber every other variable the SPA had set previously.

func NewEnvCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "per-app environment variables (Settings -> Application -> Environment)",
		Long: `Inspect or change the environment variables an installed app sees.

Subcommands:
  get  <name>                         show current env vars              (Phase 3)
  set  <name> --var KEY=VAL...        update one or more env vars        (Phase 3)
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

The CLI does a read-modify-write to avoid clobbering values it doesn't
know about: it fetches the current vector, overlays the --var pairs you
pass, and PUTs the merged result back. Only variables the SPA flagged
as editable: true can be updated; the upstream rejects writes to system
fields with a 400.

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
	current, err := fetchAppEnv(ctx, pc.doer, appName)
	if err != nil {
		return err
	}
	merged := mergeEnvUpdates(current, updates)
	body := make([]map[string]string, 0, len(merged))
	for _, e := range merged {
		body = append(body, map[string]string{"envName": e.envName, "value": e.value})
	}
	path := "/api/env/apps/" + url.PathEscape(appName) + "/env"
	if err := doMutateEnvelope(ctx, pc.doer, "PUT", path, body, nil); err != nil {
		return err
	}
	keys := make([]string, 0, len(updates))
	for k := range updates {
		keys = append(keys, k)
	}
	fmt.Printf("Updated %d environment variable(s) on %q: %s\n", len(updates), appName, strings.Join(keys, ", "))
	return nil
}

// envPair is the wire shape we emit on PUT — matches UpdateEnvItem in
// constant/index.ts.
type envPair struct {
	envName string
	value   string
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

// mergeEnvUpdates takes the upstream's current BaseEnv vector, overlays
// the user's KEY=VALUE updates, and returns the full vector to PUT.
// New keys (not in `current`) are appended at the end so the SPA's
// stable order survives a CLI mutation.
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
