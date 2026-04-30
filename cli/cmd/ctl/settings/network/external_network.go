package network

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
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
// Only the read verb is in scope today; the matching write requires a
// JWS-signed device-id header on the BFL side and is owner-only.
func NewExternalNetworkCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "external-network",
		Short: "external-network switch (Settings -> Network)",
		Long: `Inspect the master "external network" switch. When disabled, the
reverse-proxy agent and DNS configuration are frozen — useful for
keeping an Olares isolated from the public internet.

Subcommands:
  get   show the current state

Out of scope until a JWS key sourcing path exists (owner-only):
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
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "show external-network state"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runExternalNetworkGet(ctx, f, output), "show external-network state")
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
