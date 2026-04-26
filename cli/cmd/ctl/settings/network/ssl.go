package network

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings network ssl ...`
//
// Backed by user-service's /api/ssl/task-state, which forwards
// /bfl/settings/v1alpha1/ssl/task-state. The body is a BFL envelope
// around `{state: <int>}`. The state code maps as the SPA does it
// (see Mobile/connect/activate/CheckNetwork.vue):
//
//   1  Pending
//   2  Running
//   3  Failed
//   4  Succeeded
//   5  CheckL4Proxy
//   6  CheckFrpAgent
//   7  GenerateCert
//   8  ConfigureIngressHTTPs
//   9  CheckTunnel
//
// Phase 1 ships GET; Phase 4 will add `enable` (POST /api/ssl/enable).
func NewSSLCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssl",
		Short: "HTTPS provisioning task state (Settings -> Network)",
		Long: `Inspect the running HTTPS / SSL provisioning task — the same state
machine that drives the activation wizard's "checking certificate"
step.

Subcommands:
  status   show the current ssl task state                (Phase 1)

Subcommands landing in Phase 4:
  enable   re-trigger HTTPS provisioning (POST /api/ssl/enable)
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newSSLStatusCommand(f))
	return cmd
}

func newSSLStatusCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "show the current SSL task state",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runSSLStatus(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

type sslTaskState struct {
	State int `json:"state"`
}

// sslStateName preserves the SPA's exact mapping. Anything outside the
// table prints with a numeric label rather than guessing.
func sslStateName(state int) string {
	switch state {
	case 1:
		return "Pending"
	case 2:
		return "Running"
	case 3:
		return "Failed"
	case 4:
		return "Succeeded"
	case 5:
		return "CheckL4Proxy"
	case 6:
		return "CheckFrpAgent"
	case 7:
		return "GenerateCert"
	case 8:
		return "ConfigureIngressHTTPs"
	case 9:
		return "CheckTunnel"
	default:
		return fmt.Sprintf("Unknown(%d)", state)
	}
}

func runSSLStatus(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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

	var s sslTaskState
	if err := doGetEnvelope(ctx, pc.doer, "/api/ssl/task-state", &s); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, s)
	default:
		return renderSSLStatus(os.Stdout, s)
	}
}

func renderSSLStatus(w io.Writer, s sslTaskState) error {
	rows := [][2]string{
		{"State", sslStateName(s.State)},
		{"Code", fmt.Sprintf("%d", s.State)},
	}
	for _, r := range rows {
		if _, err := fmt.Fprintf(w, "%-7s %s\n", r[0]+":", r[1]); err != nil {
			return err
		}
	}
	return nil
}
