package network

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings network external-network ...`
//
// Backed by user-service's /api/external-network, which forwards
// /bfl/settings/v1alpha1/external-network. The body is BFL's
// ExternalNetworkSwitchView:
//
//   {
//     "spec":   { "disabled": <bool> },
//     "status": { "phase": "...", "message": "...", "updatedAt": "RFC3339" }
//   }
//
// Phase 1 ships GET; Phase 4 will add `set --enable / --disable`,
// which is owner-only on the BFL side (BFL returns 400 to non-owners).
func NewExternalNetworkCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "external-network",
		Short: "external-network switch (Settings -> Network)",
		Long: `Inspect or change the master "external network" switch. When disabled,
the reverse-proxy agent and DNS configuration are frozen — useful for
keeping an Olares isolated from the public internet.

Subcommands:
  get   show the current state                            (Phase 1)

Subcommands landing in Phase 4 (owner-only):
  set   --enable / --disable
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newExternalNetworkGetCommand(f))
	return cmd
}

func newExternalNetworkGetCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "show the external-network switch state",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runExternalNetworkGet(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

type externalNetworkSpec struct {
	Disabled bool `json:"disabled"`
}

type externalNetworkStatus struct {
	Phase     string `json:"phase"`
	Message   string `json:"message"`
	UpdatedAt string `json:"updatedAt"`
}

type externalNetworkView struct {
	Spec   externalNetworkSpec   `json:"spec"`
	Status externalNetworkStatus `json:"status"`
}

func runExternalNetworkGet(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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

	var view externalNetworkView
	if err := doGetEnvelope(ctx, pc.doer, "/api/external-network", &view); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, view)
	default:
		return renderExternalNetwork(os.Stdout, view)
	}
}

func renderExternalNetwork(w io.Writer, v externalNetworkView) error {
	state := "enabled"
	if v.Spec.Disabled {
		state = "disabled"
	}
	rows := [][2]string{
		{"State", state},
		{"Phase", nonEmpty(v.Status.Phase)},
		{"Message", nonEmpty(v.Status.Message)},
		{"Updated At", nonEmpty(v.Status.UpdatedAt)},
	}
	for _, r := range rows {
		if _, err := fmt.Fprintf(w, "%-12s %s\n", r[0]+":", r[1]); err != nil {
			return err
		}
	}
	return nil
}
