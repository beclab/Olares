package os

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/daemon/api"
	"github.com/beclab/Olares/cli/pkg/daemon/state"
)

// statusOptions holds the flags accepted by `olares-cli status`.
type statusOptions struct {
	endpoint string
	json     bool
	timeout  time.Duration
}

// NewCmdStatus returns the cobra command for `olares-cli status`.
//
// The command is a thin wrapper around the local olaresd daemon's
// /system/status endpoint. olaresd binds to 127.0.0.1:18088 and the
// endpoint is loopback-only on the daemon side, so the command must
// run on the same host as olaresd (typically the master node).
func NewCmdStatus() *cobra.Command {
	opts := &statusOptions{
		endpoint: api.DefaultEndpoint,
		timeout:  api.DefaultTimeout,
	}

	long := buildStatusLong()

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Print the current Olares system status reported by olaresd",
		Long:  long,
		Example: `  # Pretty-printed grouped report (default)
  olares-cli status

  # Raw JSON payload (forwarded verbatim from olaresd)
  olares-cli status --json | jq

  # Non-default daemon endpoint
  olares-cli status --endpoint http://127.0.0.1:18088 --timeout 10s`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := runStatus(cmd.Context(), cmd.OutOrStdout(), opts); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&opts.endpoint, "endpoint", opts.endpoint,
		"Base URL of the local olaresd daemon. Override only when olaresd binds to a non-default address.")
	cmd.Flags().BoolVar(&opts.json, "json", opts.json,
		"Print the raw JSON payload from olaresd (the data field), suitable for piping to tools like jq.")
	cmd.Flags().DurationVar(&opts.timeout, "timeout", opts.timeout,
		"Maximum time to wait for the olaresd response.")

	return cmd
}

func runStatus(ctx context.Context, out io.Writer, opts *statusOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}

	client := api.NewClient(opts.endpoint, opts.timeout)
	st, raw, err := client.GetSystemStatus(ctx)
	if err != nil {
		return err
	}

	if opts.json {
		// Pretty-print to keep the output friendly when not piped.
		var pretty interface{}
		if jerr := json.Unmarshal(raw, &pretty); jerr == nil {
			b, merr := json.MarshalIndent(pretty, "", "  ")
			if merr == nil {
				_, err = fmt.Fprintln(out, string(b))
				return err
			}
		}
		_, err = fmt.Fprintln(out, string(raw))
		return err
	}

	return printHumanReadable(out, st)
}

// printHumanReadable renders the State as grouped sections, each with
// padded labels and inline annotations for state values. The width
// is chosen so that values line up when printed in a typical 80-col
// terminal without truncating common content.
func printHumanReadable(out io.Writer, s *state.State) error {
	const labelWidth = 20

	w := &writer{out: out, label: labelWidth}

	w.section("Olares")
	w.kv("State", string(s.TerminusState), s.TerminusState.Describe())
	w.kv("Olaresd state", string(s.TerminusdState), "")
	w.kv("Name", strPtr(s.TerminusName), "")
	w.kv("Version", strPtr(s.TerminusVersion), "")
	w.kv("Olaresd version", strPtr(s.OlaresdVersion), "")
	w.kv("Installed at", formatEpoch(s.InstalledTime), "")
	w.kv("Initialized at", formatEpoch(s.InitializedTime), "")

	w.section("System")
	w.kv("Device", strPtr(s.DeviceName), "")
	w.kv("Hostname", strPtr(s.HostName), "")
	w.kv("OS", joinOSInfo(s.OsType, s.OsArch, s.OsInfo), "")
	w.kv("OS version", s.OsVersion, "")
	w.kv("CPU", s.CpuInfo, "")
	w.kv("Memory", s.Memory, "")
	w.kv("Disk", s.Disk, "")
	w.kv("GPU", strPtr(s.GpuInfo), "")

	w.section("Network")
	w.kv("Wired", yesNo(s.WiredConnected), "")
	w.kv("Wi-Fi", yesNo(s.WifiConnected), "")
	w.kv("Wi-Fi SSID", strPtr(s.WifiSSID), "")
	w.kv("Host IP", s.HostIP, "")
	w.kv("External IP", s.ExternalIP, "")

	w.section("Install / Uninstall")
	w.kv("Installing", string(s.InstallingState), s.InstallingProgress)
	w.kv("Uninstalling", string(s.UninstallingState), s.UninstallingProgress)

	w.section("Upgrade")
	w.kv("Target", s.UpgradingTarget, "")
	w.kv("State", string(s.UpgradingState), s.UpgradingProgress)
	w.kv("Step", s.UpgradingStep, "")
	w.kv("Last error", s.UpgradingError, "")
	w.kv("Download state", string(s.UpgradingDownloadState), s.UpgradingDownloadProgress)
	w.kv("Download step", s.UpgradingDownloadStep, "")
	w.kv("Download error", s.UpgradingDownloadError, "")
	if s.UpgradingRetryNum > 0 {
		w.kv("Retry count", fmt.Sprintf("%d", s.UpgradingRetryNum), "")
	}
	if s.UpgradingNextRetryAt != nil {
		w.kv("Next retry at", s.UpgradingNextRetryAt.Local().Format(time.RFC3339), "")
	}

	w.section("Logs collection")
	w.kv("State", string(s.CollectingLogsState), s.CollectingLogsError)

	w.section("Pressures")
	if len(s.Pressure) == 0 {
		w.line("(none)")
	} else {
		for _, p := range s.Pressure {
			w.kv(p.Type, p.Message, "")
		}
	}

	w.section("Other")
	w.kv("FRP enabled", s.FRPEnable, "")
	w.kv("FRP server", s.DefaultFRPServer, "")
	w.kv("Container mode", strPtr(s.ContainerMode), "")

	return w.err
}

// writer collects formatting helpers in one place so the section
// printer above stays declarative.
type writer struct {
	out   io.Writer
	label int
	err   error
}

func (w *writer) writef(format string, a ...interface{}) {
	if w.err != nil {
		return
	}
	_, w.err = fmt.Fprintf(w.out, format, a...)
}

func (w *writer) section(name string) {
	w.writef("\n%s\n", name)
}

func (w *writer) line(s string) {
	w.writef("  %s\n", s)
}

func (w *writer) kv(key, value, note string) {
	if value == "" {
		value = "-"
	}
	if note != "" {
		w.writef("  %-*s %s   (%s)\n", w.label, key, value, note)
	} else {
		w.writef("  %-*s %s\n", w.label, key, value)
	}
}

func strPtr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func formatEpoch(p *int64) string {
	if p == nil || *p == 0 {
		return ""
	}
	return time.Unix(*p, 0).Local().Format("2006-01-02 15:04:05 -0700")
}

func joinOSInfo(parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, " ")
}

// buildStatusLong constructs the cobra Long help text. State
// descriptions are sourced from state.TerminusState.Describe() so
// that adding a new state here automatically updates the help.
func buildStatusLong() string {
	var b strings.Builder
	b.WriteString(`Print the current Olares system status reported by the local olaresd daemon.

This command sends an HTTP GET to the daemon's /system/status endpoint
(default: http://127.0.0.1:18088/system/status). The endpoint is
loopback-only on the daemon side, so this command must run on the same
host as olaresd (typically the master node).

The default output is a grouped, human-readable report:

  Olares                installation lifecycle, version, names, key timestamps
  System                hardware and OS facts about the host
  Network               connectivity and IP addresses
  Install / Uninstall   progress of in-flight install or uninstall
  Upgrade               progress of in-flight upgrade (download + install phases)
  Logs collection       state of the most recent log collection job
  Pressures             active kubelet node pressure conditions, if any
  Other                 FRP, container mode, etc.

Pass --json to get the raw daemon payload instead, useful for scripting.

Olares system states (TerminusState):

`)

	for _, s := range state.AllTerminusStates() {
		fmt.Fprintf(&b, "  %-20s %s\n", string(s), s.Describe())
	}

	return b.String()
}
