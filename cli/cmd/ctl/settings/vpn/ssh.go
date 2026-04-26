package vpn

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings vpn ssh ...`
//
// SSH-over-Headscale toggle. Mirrors the SPA's Settings -> VPN ->
// Allow SSH switch (apps/.../stores/settings/acl.ts:67-93). Three
// endpoints behind one cobra parent:
//
//	GET  /api/acl/ssh/status              -> { state, allow_ssh }
//	POST /api/acl/ssh/enable     body {}  -> success / failure
//	POST /api/acl/ssh/disable    body {}  -> success / failure
//
// All three sit on user-service's bfl/acl.controller.ts and forward to
// the per-Olares ACL CRD. The status read returns the unwrapped inner
// shape (`state`, `allow_ssh`) — user-service strips the BFL envelope
// before returning, same pattern as public-domain-policy.
//
// Role: SPA gates the switch on isAdmin; we don't hard-gate here
// because the server is authoritative — a normal user calling this
// will get a 403 with the WrapPermissionErr CTA.

func NewSSHCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh",
		Short: "SSH-over-Headscale toggle (Settings -> VPN -> Allow SSH)",
		Long: `Inspect or change whether SSH connections are permitted across the
Headscale mesh. Mirrors the "Allow SSH" switch on the SPA's VPN page.

Subcommands:
  status     show current ACL state + allow_ssh flag
  enable     permit SSH across the mesh
  disable    block SSH across the mesh
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newSSHStatusCommand(f))
	cmd.AddCommand(newSSHEnableCommand(f))
	cmd.AddCommand(newSSHDisableCommand(f))
	return cmd
}

// sshStatus is the inner shape user-service returns after stripping the
// BFL envelope. The `state` field is a free-form upstream label
// (typically "ok" or an error string when the ACL CRD hasn't reconciled
// yet); `allow_ssh` is the actual toggle value.
type sshStatus struct {
	State    string `json:"state"`
	AllowSSH bool   `json:"allow_ssh"`
}

func newSSHStatusCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "show current SSH-over-Headscale state",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runSSHStatus(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runSSHStatus(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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
	var resp sshStatus
	if err := pc.doer.DoJSON(ctx, "GET", "/api/acl/ssh/status", nil, &resp); err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, resp)
	default:
		return renderSSHStatus(os.Stdout, resp)
	}
}

func renderSSHStatus(w io.Writer, s sshStatus) error {
	rows := [][2]string{
		{"State", nonEmpty(s.State)},
		{"Allow SSH", boolStr(s.AllowSSH)},
	}
	for _, r := range rows {
		if _, err := fmt.Fprintf(w, "%-12s %s\n", r[0]+":", r[1]); err != nil {
			return err
		}
	}
	return nil
}

func newSSHEnableCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "enable",
		Short: "permit SSH across the Headscale mesh",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runSSHToggle(c.Context(), f, true)
		},
	}
}

func newSSHDisableCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: "block SSH across the Headscale mesh",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runSSHToggle(c.Context(), f, false)
		},
	}
}

func runSSHToggle(ctx context.Context, f *cmdutil.Factory, enable bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	if err := doSSHToggleViaDoer(ctx, pc.doer, enable); err != nil {
		return err
	}
	verb := "disable"
	if enable {
		verb = "enable"
	}
	fmt.Fprintf(os.Stdout, "%sd SSH-over-Headscale\n", verb)
	return nil
}

// doSSHToggleViaDoer is the wire-only core of `vpn ssh enable|disable`.
// SPA sends an explicit empty {} — we match it for consistency with the
// SPA's request shape and to keep any defensive middleware in front of
// user-service happy with a well-formed body.
func doSSHToggleViaDoer(ctx context.Context, d Doer, enable bool) error {
	verb := "disable"
	if enable {
		verb = "enable"
	}
	path := "/api/acl/ssh/" + verb
	return d.DoJSON(ctx, "POST", path, struct{}{}, nil)
}
