package advanced

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/daemon/state"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// Directory where terminusd places collected log bundles (same as SPA
// Settings -> Developer -> Logs).
const collectedLogsPath = "/Home/pod_logs"

// `olares-cli settings advanced collect-logs`
//
// Backed by POST /api/command/collectLogs (no body). user-service
// terminusd.controller.ts proxies to terminusd /command/collect-logs.
// The SPA sends X-Signature (device id); user-service getXSignatureByRequest
// falls back to X-Authorization, which the CLI supplies via the factory
// transport — same pattern as other /api/* settings calls.
//
// SPA: packages/app/src/stores/settings/terminusd.ts collect_logs(), then
// polls GET /api/system/status for collectingLogsState (see LogPage.vue).
//
// CLI mirrors `advanced status`: owner/admin only (soft preflight Gate), because
// polling uses the same /api/system/status surface.
func NewCollectLogsCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	var timeout time.Duration
	var pollInterval time.Duration

	cmd := &cobra.Command{
		Use:   "collect-logs",
		Short: "collect diagnostic logs and wait until the job finishes",
		Long: `Start the same log bundle job as Settings -> Developer -> Logs ("Collect") in the SPA.

The command POSTs /api/command/collectLogs, then polls GET /api/system/status for
collectingLogsState and collectingLogsError (same fields the SPA reads from
terminusdStore.olaresInfo) until the daemon reports completed or failed.

On success it prints the fixed output path under your Files volume. On failure it
prints the daemon error and exits non-zero.

Use --timeout when very large clusters need more than the default wait budget.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "collect diagnostic logs"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runCollectLogs(ctx, f, collectLogsOptions{
				output:       output,
				timeout:      timeout,
				pollInterval: pollInterval,
			}), "collect diagnostic logs")
		},
	}
	addOutputFlag(cmd, &output)
	cmd.Flags().DurationVar(&timeout, "timeout", 30*time.Minute,
		"maximum time to wait for collectingLogsState to reach completed or failed")
	cmd.Flags().DurationVar(&pollInterval, "poll-interval", 3*time.Second,
		"delay between GET /api/system/status polls while collection runs")
	return cmd
}

type collectLogsOptions struct {
	output       string
	timeout      time.Duration
	pollInterval time.Duration
}

type collectLogsResult struct {
	Path                string `json:"path"`
	CollectingLogsState string `json:"collectingLogsState"`
	Message             string `json:"message,omitempty"`
}

func runCollectLogs(ctx context.Context, f *cmdutil.Factory, opts collectLogsOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(opts.output)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	var accept json.RawMessage
	if err := doMutateEnvelope(ctx, pc.doer, http.MethodPost, "/api/command/collectLogs", nil, &accept); err != nil {
		return err
	}
	acceptMsg := decodeEnvelopeString(accept)

	pollCtx, cancel := context.WithTimeout(ctx, opts.timeout)
	defer cancel()

	if format == FormatTable {
		fmt.Fprintf(os.Stderr, "Waiting for log collection (timeout %s, poll every %s)...\n",
			opts.timeout.Round(time.Second), opts.pollInterval.Round(time.Second))
	}

	for {
		var st state.State
		if err := doGetEnvelope(pollCtx, pc.doer, "/api/system/status", &st); err != nil {
			return err
		}

		switch st.CollectingLogsState {
		case state.Completed:
			res := collectLogsResult{
				Path:                collectedLogsPath,
				CollectingLogsState: string(st.CollectingLogsState),
				Message:             acceptMsg,
			}
			switch format {
			case FormatJSON:
				return printJSON(os.Stdout, res)
			default:
				if acceptMsg != "" {
					fmt.Println(acceptMsg)
				}
				fmt.Printf("collectingLogsState: %s\n", st.CollectingLogsState)
				fmt.Printf("path: %s\n", res.Path)
				fmt.Println("(open from the Files app under Home -> pod_logs, same as the SPA)")
				return nil
			}
		case state.Failed:
			reason := strings.TrimSpace(st.CollectingLogsError)
			if reason == "" {
				reason = "(no message from daemon)"
			}
			return fmt.Errorf("log collection failed: %s", reason)
		}

		select {
		case <-pollCtx.Done():
			if pollCtx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("log collection timed out after %s (collectingLogsState did not reach completed/failed)", opts.timeout)
			}
			return pollCtx.Err()
		case <-time.After(opts.pollInterval):
		}
	}
}

func decodeEnvelopeString(raw json.RawMessage) string {
	s := strings.TrimSpace(string(raw))
	if s == "" {
		return ""
	}
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		return strings.TrimSpace(str)
	}
	return s
}
