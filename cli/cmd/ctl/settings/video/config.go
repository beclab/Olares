package video

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings video config get`
//
// Backed by /api/files/video/config (a Jellyfin-style encoding blob).
// The schema is large (hardware acceleration, encoding scheme,
// transcoding settings, audio transcoding, encoding quality, others)
// and tracks the upstream Jellyfin config — so we keep the CLI honest
// by emitting the JSON tree as-is. The default --output table prints
// the keys at the top level with their value/sub-object summaries; pass
// --output json to script against the full structure.
func NewConfigCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "video / Jellyfin encoding config",
		Long: `Read the video encoding configuration that Olares ships through
Jellyfin (hardware accel, encoding scheme, transcoding, ...).

Subcommands:
  get   show the current config

Out of scope for now:
  set   write a new config blob (atomic replace)
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newConfigGetCommand(f))
	return cmd
}

func newConfigGetCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "show the current video encoding config",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runConfigGet(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runConfigGet(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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

	raw, err := doGetEnvelopeRaw(ctx, pc.doer, "/api/files/video/config")
	if err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSONRaw(os.Stdout, raw)
	default:
		return renderConfigTable(os.Stdout, raw)
	}
}

// renderConfigTable surfaces the top-level keys of the (provider-versioned)
// JSON tree. For scalar leaves we print the value verbatim; for nested
// objects/arrays we render `(object)` / `(array, N)` and tell the reader
// to use --output json to inspect.
func renderConfigTable(w io.Writer, raw json.RawMessage) error {
	if len(raw) == 0 {
		_, err := fmt.Fprintln(w, "(empty config)")
		return err
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		_, e := fmt.Fprintln(w, string(raw))
		return e
	}
	keys := make([]string, 0, len(top))
	for k := range top {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		summary, err := summarizeJSON(top[k])
		if err != nil {
			summary = string(top[k])
		}
		if _, err := fmt.Fprintf(w, "%-32s %s\n", k+":", summary); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w, "\nuse --output json for the full tree"); err != nil {
		return err
	}
	return nil
}

func summarizeJSON(raw json.RawMessage) (string, error) {
	if len(raw) == 0 {
		return "-", nil
	}
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return "", err
	}
	switch t := v.(type) {
	case nil:
		return "null", nil
	case bool:
		if t {
			return "true", nil
		}
		return "false", nil
	case string:
		return fmt.Sprintf("%q", t), nil
	case float64:
		// json numbers come back as float64; preserve int formatting when possible
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t)), nil
		}
		return fmt.Sprintf("%g", t), nil
	case []interface{}:
		return fmt.Sprintf("(array, %d)", len(t)), nil
	case map[string]interface{}:
		return fmt.Sprintf("(object, %d keys)", len(t)), nil
	default:
		return fmt.Sprintf("%v", t), nil
	}
}

var _ = io.Discard // keep io import even when only used via os.Stdout
