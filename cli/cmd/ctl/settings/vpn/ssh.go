package vpn

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
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
// the per-Olares ACL CRD which BFL exposes at /bfl/settings/v1alpha1/
// headscale/{ssh/acl,enable/ssh,disable/ssh}.
//
// Wire shape: user-service forwards BFL's response verbatim, so the
// status read comes back wrapped in the BFL envelope
// `{code, message, data: {state, allow_ssh}}`. The SPA hides that
// because boot/axios.ts installs a global response interceptor that
// strips `data.data` before handing the body to the store; CLI callers
// don't get that for free, so decodeSSHStatus has to unwrap the
// envelope before we can render `state` / `allow_ssh`. It also accepts
// the already-unwrapped shape defensively in case a future user-service
// version starts stripping the envelope for this path (the comment in
// vpn/common.go used to assert that, but the live behavior is the
// opposite — see KNOWN_ISSUES for the back-history).
//
// Role: SPA gates the switch on isAdmin (VPNPage.vue:20-32). We mirror
// that here with a soft preflight; the server stays authoritative, so a
// non-admin caller still gets a friendly hint instead of a bare 403.

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

// sshStatus is the inner shape BFL returns for the SSH ACL read
// (handle_headscale.go's `SshAcl`). It is normally wrapped in the BFL
// envelope by `response.Success` on the server side and forwarded
// verbatim by user-service — see decodeSSHStatus for the unwrapping
// logic. The `state` field is a free-form upstream label (typically
// "applied" or an error string when the ACL CRD hasn't reconciled yet);
// `allow_ssh` is the actual toggle value.
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
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "show SSH-over-Headscale state"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runSSHStatus(ctx, f, output), "show SSH-over-Headscale state")
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
	var raw json.RawMessage
	if err := pc.doer.DoJSON(ctx, "GET", "/api/acl/ssh/status", nil, &raw); err != nil {
		return err
	}
	resp, err := decodeSSHStatus(raw)
	if err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, resp)
	default:
		return renderSSHStatus(os.Stdout, resp)
	}
}

// decodeSSHStatus is the wire-shape adapter for `GET /api/acl/ssh/status`.
// The endpoint currently returns the BFL envelope
// `{code, message, data: {state, allow_ssh}}` verbatim through
// user-service; the SPA only sees the inner shape because its
// boot/axios.ts response interceptor strips `data.data` globally
// (apps/packages/app/src/boot/axios.ts:163). Without that interceptor
// the CLI used to silently render `state: ""` / `allow_ssh: no` and
// callers saw no change after `vpn ssh enable|disable`.
//
// We tolerate both the wrapped and unwrapped shapes so a future
// user-service rewrite that does strip the envelope (the way it
// already does for /api/launcher-public-domain-access-policy) doesn't
// flip this command back into "always empty" mode.
func decodeSSHStatus(raw json.RawMessage) (sshStatus, error) {
	var out sshStatus
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return out, nil
	}
	var env bflEnvelope
	if err := json.Unmarshal(trimmed, &env); err == nil && envelopeLooksWrapped(trimmed, env) {
		switch env.Code {
		case 0, 200:
		default:
			msg := strings.TrimSpace(env.Message)
			if msg == "" {
				msg = fmt.Sprintf("server returned code=%d", env.Code)
			}
			return out, fmt.Errorf("GET /api/acl/ssh/status: %s", msg)
		}
		if len(env.Data) == 0 || string(bytes.TrimSpace(env.Data)) == "null" {
			return out, nil
		}
		if err := json.Unmarshal(env.Data, &out); err != nil {
			return out, fmt.Errorf("decode acl ssh status data: %w", err)
		}
		return out, nil
	}
	if err := json.Unmarshal(trimmed, &out); err != nil {
		return out, fmt.Errorf("decode acl ssh status body: %w", err)
	}
	return out, nil
}

// envelopeLooksWrapped reports whether `body` matches the BFL envelope
// shape `{code, message, data}` rather than the inner sshStatus shape
// `{state, allow_ssh}`. Both targets json.Unmarshal cleanly into a
// bflEnvelope (extra keys are ignored, missing keys default to zero),
// so we look at the body itself: the envelope always carries a
// top-level `code` key (response.Success / response.HandleError both
// emit it unconditionally), while the inner shape never does. Treat
// the presence of either `code` or `data` as proof of an envelope so
// we still surface non-success codes on error responses where the
// upstream omits `data`.
func envelopeLooksWrapped(body []byte, env bflEnvelope) bool {
	if len(env.Data) > 0 {
		return true
	}
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(body, &probe); err != nil {
		return false
	}
	if _, ok := probe["data"]; ok {
		return true
	}
	_, ok := probe["code"]
	return ok
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
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "enable SSH over mesh"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runSSHToggle(ctx, f, true), "enable SSH over mesh")
		},
	}
}

func newSSHDisableCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: "block SSH across the Headscale mesh",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "disable SSH over mesh"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runSSHToggle(ctx, f, false), "disable SSH over mesh")
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
