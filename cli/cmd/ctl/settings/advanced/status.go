package advanced

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

// `olares-cli settings advanced status`
//
// Backed by /api/system/status. The body is a BFL envelope around the
// TerminusStatus struct (terminusState, terminusdState, os_*, device_*,
// memory, disk, cpu_info, gpu_info, ...). The schema is large + tracks
// the daemon, so the table view shows the headline fields and points
// at --output json for the rest.
//
// The `terminusState` enum has friendly labels documented in the SPA's
// services/abstractions/mdns/service.ts; here we just print the raw
// state code (and let users `grep` it). The CLI is for tooling, not UI
// — friendly labels would drift away from upstream over time.
func NewStatusCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "show system / daemon status (Settings -> Advanced)",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runStatus(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runStatus(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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

	var raw json.RawMessage
	if err := doGetEnvelope(ctx, pc.doer, "/api/system/status", &raw); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		if len(raw) == 0 {
			_, err := fmt.Fprintln(os.Stdout, "{}")
			return err
		}
		var pretty interface{}
		if err := json.Unmarshal(raw, &pretty); err != nil {
			_, err := fmt.Fprintln(os.Stdout, string(raw))
			return err
		}
		return printJSON(os.Stdout, pretty)
	default:
		return renderStatusTable(os.Stdout, raw)
	}
}

// renderStatusTable surfaces a curated set of headline fields from
// TerminusStatus, then nudges users at --output json for the rest.
func renderStatusTable(w io.Writer, raw json.RawMessage) error {
	if len(raw) == 0 {
		_, err := fmt.Fprintln(w, "(empty status)")
		return err
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		_, err := fmt.Fprintln(w, string(raw))
		return err
	}

	headline := []string{
		"terminusState",
		"terminusdState",
		"installingState",
		"installingProgress",
		"uninstallingProgress",
		"terminusName",
		"terminusVersion",
		"device_name",
		"host_name",
		"os_type",
		"os_arch",
		"os_version",
		"cpu_info",
		"memory",
		"disk",
		"gpu_info",
		"updateTime",
	}
	printed := map[string]struct{}{}
	for _, k := range headline {
		v, ok := top[k]
		if !ok {
			continue
		}
		val, _ := scalarOrSummary(v)
		if _, err := fmt.Fprintf(w, "%-22s %s\n", k+":", val); err != nil {
			return err
		}
		printed[k] = struct{}{}
	}

	leftover := make([]string, 0)
	for k := range top {
		if _, done := printed[k]; done {
			continue
		}
		leftover = append(leftover, k)
	}
	sort.Strings(leftover)
	if len(leftover) > 0 {
		if _, err := fmt.Fprintf(w, "\nadditional fields available via --output json:\n  %s\n", joinComma(leftover)); err != nil {
			return err
		}
	}
	return nil
}

func scalarOrSummary(raw json.RawMessage) (string, bool) {
	if len(raw) == 0 {
		return "-", false
	}
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw), true
	}
	switch t := v.(type) {
	case nil:
		return "null", true
	case bool:
		if t {
			return "true", true
		}
		return "false", true
	case string:
		if t == "" {
			return "-", true
		}
		return t, true
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t)), true
		}
		return fmt.Sprintf("%g", t), true
	case []interface{}:
		return fmt.Sprintf("(array, %d)", len(t)), false
	case map[string]interface{}:
		return fmt.Sprintf("(object, %d keys)", len(t)), false
	default:
		return fmt.Sprintf("%v", t), false
	}
}

func joinComma(ss []string) string {
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += ", "
		}
		out += s
	}
	return out
}
