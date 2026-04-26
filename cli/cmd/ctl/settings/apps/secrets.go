package apps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings apps secrets ...`
//
// Per-app secret store. The SPA's per-app Settings page reaches it via
// the desktop ingress at /admin/secret/<app> (NOT /api/...; this is a
// distinct upstream from user-service's /api/secret personal vault).
// The same X-Authorization Bearer token authenticates both paths,
// because the desktop ingress is the single trust boundary.
//
//	GET    /admin/secret/{app}                   list keys
//	POST   /admin/secret/{app}  body {Key,Value} create
//	PUT    /admin/secret/{app}  body {Key,Value} update
//	DELETE /admin/secret/{app}  body {Key}       delete
//
// Field casing is intentionally CapitalCase on the wire (matches the
// SPA store stores/settings/secret.ts:42-86 verbatim).
//
// Response shape: opaque to the SPA — store discards everything except
// the create/update success notification. We do the same: every verb
// returns nil on 2xx and lets the underlying Doer surface errors. We
// don't try to wrap a BFL envelope here because the upstream service
// hasn't been observed to use one consistently.

func NewSecretsCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secrets",
		Aliases: []string{"secret"},
		Short:   "per-app secret store (Settings -> Application -> Secrets)",
		Long: `Manage per-app secrets at /admin/secret/<app>.

Subcommands:
  list   <app>                                show stored keys           (Phase 3)
  set    <app> --key KEY (--value VAL | --value-stdin)  create OR update (Phase 3)
  delete <app> --key KEY [--yes]              remove a stored key        (Phase 3)
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newSecretsListCommand(f))
	cmd.AddCommand(newSecretsSetCommand(f))
	cmd.AddCommand(newSecretsDeleteCommand(f))
	return cmd
}

func newSecretsListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list <app>",
		Short: "list secret keys stored for an app",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runSecretsList(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runSecretsList(ctx context.Context, f *cmdutil.Factory, appName, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	appName = strings.TrimSpace(appName)
	if appName == "" {
		return fmt.Errorf("list requires an app name")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	path := "/admin/secret/" + url.PathEscape(appName)
	var raw json.RawMessage
	if err := pc.doer.DoJSON(ctx, "GET", path, nil, &raw); err != nil {
		return err
	}
	if format == FormatJSON {
		var v interface{}
		if len(raw) == 0 {
			return printJSON(os.Stdout, nil)
		}
		if err := json.Unmarshal(raw, &v); err != nil {
			return fmt.Errorf("decode secrets list: %w", err)
		}
		return printJSON(os.Stdout, v)
	}
	return renderSecretsTable(os.Stdout, raw)
}

// renderSecretsTable handles a couple of shapes upstream has shipped
// over time: a flat array of {Key,Value} or {key,value} objects, or an
// envelope with a "data" key wrapping the array. We pick whichever
// matches and print KEY (only) — values stay hidden in table mode for
// terminal-shoulder-surfing safety; --output json reveals the full
// payload exactly as the upstream returned it.
func renderSecretsTable(w io.Writer, raw json.RawMessage) error {
	keys := extractSecretKeys(raw)
	if len(keys) == 0 {
		fmt.Fprintln(w, "no secrets stored for this app")
		return nil
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "KEY")
	for _, k := range keys {
		fmt.Fprintln(tw, k)
	}
	return tw.Flush()
}

func extractSecretKeys(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	type secretEntry struct {
		Key      string `json:"Key"`
		KeyLower string `json:"key"`
	}
	tryArray := func(b []byte) []string {
		var arr []secretEntry
		if err := json.Unmarshal(b, &arr); err != nil {
			return nil
		}
		out := make([]string, 0, len(arr))
		for _, e := range arr {
			if e.Key != "" {
				out = append(out, e.Key)
			} else if e.KeyLower != "" {
				out = append(out, e.KeyLower)
			}
		}
		return out
	}
	if keys := tryArray(raw); keys != nil {
		return keys
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err == nil && len(env.Data) > 0 {
		if keys := tryArray(env.Data); keys != nil {
			return keys
		}
	}
	var asMap map[string]json.RawMessage
	if err := json.Unmarshal(raw, &asMap); err == nil && len(asMap) > 0 {
		out := make([]string, 0, len(asMap))
		for k := range asMap {
			out = append(out, k)
		}
		return out
	}
	return nil
}

func newSecretsSetCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		key         string
		value       string
		valueStdin  bool
		updateOnly  bool
		createOnly  bool
	)
	cmd := &cobra.Command{
		Use:   "set <app>",
		Short: "create or update an app secret",
		Long: `Create or update a single secret entry on an installed app.

Pass the value via --value (literal) or --value-stdin (read once from
stdin). For "create" semantics use --create-only (errors out if the key
exists), or --update-only for "update" (errors out if it doesn't).
Default is upsert: try POST first, fall back to PUT on 4xx that looks
like "already exists".

Examples:
  olares-cli settings apps secrets set my-app --key API_KEY --value abc123
  echo -n "abc123" | olares-cli settings apps secrets set my-app --key API_KEY --value-stdin
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runSecretsSet(c.Context(), f, args[0], key, value, valueStdin, createOnly, updateOnly)
		},
	}
	cmd.Flags().StringVar(&key, "key", "", "secret key (required)")
	cmd.Flags().StringVar(&value, "value", "", "secret value (use --value-stdin to read from stdin instead)")
	cmd.Flags().BoolVar(&valueStdin, "value-stdin", false, "read the value from stdin once")
	cmd.Flags().BoolVar(&createOnly, "create-only", false, "fail if the key already exists (POST only)")
	cmd.Flags().BoolVar(&updateOnly, "update-only", false, "fail if the key doesn't exist (PUT only)")
	_ = cmd.MarkFlagRequired("key")
	return cmd
}

func runSecretsSet(ctx context.Context, f *cmdutil.Factory, appName, key, value string, valueStdin, createOnly, updateOnly bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	appName = strings.TrimSpace(appName)
	key = strings.TrimSpace(key)
	if appName == "" {
		return fmt.Errorf("set requires an app name")
	}
	if key == "" {
		return fmt.Errorf("--key is required")
	}
	if createOnly && updateOnly {
		return fmt.Errorf("--create-only and --update-only are mutually exclusive")
	}
	if valueStdin {
		buf, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read --value-stdin: %w", err)
		}
		value = strings.TrimRight(string(buf), "\n\r")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	path := "/admin/secret/" + url.PathEscape(appName)
	body := map[string]string{"Key": key, "Value": value}
	method := "POST"
	switch {
	case updateOnly:
		method = "PUT"
	case createOnly:
		method = "POST"
	}
	if err := pc.doer.DoJSON(ctx, method, path, body, nil); err != nil {
		if !createOnly && !updateOnly && method == "POST" && looksLikeAlreadyExists(err) {
			if err := pc.doer.DoJSON(ctx, "PUT", path, body, nil); err != nil {
				return err
			}
			fmt.Printf("Updated secret %q on %q.\n", key, appName)
			return nil
		}
		return err
	}
	verb := "Set"
	if method == "PUT" {
		verb = "Updated"
	}
	fmt.Printf("%s secret %q on %q.\n", verb, key, appName)
	return nil
}

// looksLikeAlreadyExists is a best-effort sniff of the HTTP error
// message returned by the underlying Doer. The upstream phrases the
// "duplicate key" error a few different ways and there's no reliable
// status code (most variants come back as 4xx with the body string in
// the wrapped error) — we look for the most common substrings rather
// than parsing the body. False negatives degrade to a normal error
// surface, which is acceptable.
func looksLikeAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	for _, needle := range []string{"already exists", "already exist", "duplicate", "exist"} {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}

func newSecretsDeleteCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		key       string
		assumeYes bool
	)
	cmd := &cobra.Command{
		Use:     "delete <app>",
		Aliases: []string{"rm", "remove"},
		Short:   "delete an app secret by key",
		Args:    cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runSecretsDelete(c.Context(), f, args[0], key, assumeYes)
		},
	}
	cmd.Flags().StringVar(&key, "key", "", "secret key to delete (required)")
	cmd.Flags().BoolVar(&assumeYes, "yes", false, "skip the y/N prompt (required for non-TTY stdin)")
	_ = cmd.MarkFlagRequired("key")
	return cmd
}

func runSecretsDelete(ctx context.Context, f *cmdutil.Factory, appName, key string, assumeYes bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	appName = strings.TrimSpace(appName)
	key = strings.TrimSpace(key)
	if appName == "" {
		return fmt.Errorf("delete requires an app name")
	}
	if key == "" {
		return fmt.Errorf("--key is required")
	}
	if !assumeYes {
		if err := confirmSecretDelete(os.Stderr, os.Stdin, appName, key); err != nil {
			return err
		}
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	path := "/admin/secret/" + url.PathEscape(appName)
	body := map[string]string{"Key": key}
	if err := pc.doer.DoJSON(ctx, "DELETE", path, body, nil); err != nil {
		return err
	}
	fmt.Printf("Deleted secret %q on %q.\n", key, appName)
	return nil
}

// confirmSecretDelete is a thin local prompt rather than a generic
// helper because (a) it's the only destructive verb in `apps` and (b)
// the message wording is specific enough that a shared helper would
// hide too much.
func confirmSecretDelete(prompt io.Writer, in io.Reader, appName, key string) error {
	if f, ok := in.(*os.File); ok {
		fd := int(f.Fd())
		_ = fd
	}
	if _, err := fmt.Fprintf(prompt, "Delete secret %q on %q? [y/N]: ", key, appName); err != nil {
		return err
	}
	var line string
	if _, err := fmt.Fscanln(in, &line); err != nil && err.Error() != "unexpected newline" {
		return fmt.Errorf("read confirmation: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return nil
	default:
		return fmt.Errorf("aborted by user (pass --yes to skip the prompt)")
	}
}
