package gpu

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/edge"
	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/pkg/cliutil"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/olaresclient"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// validSupportTypes is the set the backend (and SPA) recognize for a device's
// support type. Nvidia cards use Exclusive|MemorySlice|TimeSlice; non-Nvidia
// single-virtual-device nodes use Exclusive|MemoryShared.
var validSupportTypes = []string{"Exclusive", "MemorySlice", "TimeSlice", "MemoryShared"}

func isValidSupportType(s string) bool {
	for _, v := range validSupportTypes {
		if v == s {
			return true
		}
	}
	return false
}

// NewSupportTypeCommand is the `settings gpu support-type` parent (Olares
// 1.12.6+); today it carries the single `set` verb.
func NewSupportTypeCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "support-type",
		Short: "manage a device's support type (Olares 1.12.6+)",
		Long: `Inspect and change a compute device's support type.

Subcommands:
  set <node> <device> <type>   switch a device's support type

Valid types: Exclusive | MemorySlice | TimeSlice | MemoryShared.`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newSupportTypeSetCommand(f))
	return cmd
}

// `olares-cli settings gpu support-type set <node> <device> <type>`
//
// Switches a device's support type via
// olaresclient.ComputeOps.SwitchSupportType (PUT
// /api/compute-resources/nodes/<node>/devices/<device>/support-type). The
// backend reports the outcome via a `status` discriminator
// (switched / unchanged / bound-apps-stop-blocked) it keeps at HTTP 200, so we
// branch on that rather than on an HTTP error.
func newSupportTypeSetCommand(f *cmdutil.Factory) *cobra.Command {
	var assumeYes bool
	cmd := &cobra.Command{
		Use:   "set <node> <device> <type>",
		Short: "switch a device's support type",
		Long: `Switch a compute device's support type.

Valid types: Exclusive | MemorySlice | TimeSlice | MemoryShared.

Changing the support type may stop apps currently bound to the device, so this
prompts for confirmation by default; pass --yes to skip it. If apps that must be
stopped first are blocking the switch, the command reports them and exits
non-zero without changing anything.

Example:
  olares-cli settings gpu support-type set node-1 GPU-abc123 TimeSlice --yes`,
		Args: cobra.ExactArgs(3),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "switch device support type"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runSupportTypeSet(ctx, f, args[0], args[1], args[2], assumeYes), "switch device support type")
		},
	}
	cmd.Flags().BoolVar(&assumeYes, "yes", false, "skip the confirmation prompt (required for non-TTY automation)")
	return cmd
}

type supportTypeAppRef struct {
	AppName string `json:"appName"`
	Owner   string `json:"owner,omitempty"`
	State   string `json:"state,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

type supportTypeResult struct {
	Status      string              `json:"status"`
	StoppedApps []supportTypeAppRef `json:"stoppedApps,omitempty"`
	BlockedApps []supportTypeAppRef `json:"blockedApps,omitempty"`
}

func runSupportTypeSet(ctx context.Context, f *cmdutil.Factory, node, deviceID, supportType string, assumeYes bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	supportType = strings.TrimSpace(supportType)
	if !isValidSupportType(supportType) {
		return fmt.Errorf("invalid support type %q (allowed: %s)", supportType, strings.Join(validSupportTypes, ", "))
	}
	doer, _, err := edge.New(ctx, f)
	if err != nil {
		return err
	}

	if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin,
		fmt.Sprintf("Switch device %q on node %q to %q? Apps bound to this device may be stopped.", deviceID, node, supportType),
		assumeYes); err != nil {
		return err
	}

	return f.WithOlaresClient(ctx, func(c olaresclient.OlaresClient) error {
		raw, err := c.SwitchSupportType(ctx, doer, node, deviceID, supportType)
		if err != nil {
			return err
		}
		var res supportTypeResult
		if err := decodeData(raw, &res); err != nil {
			return err
		}
		switch res.Status {
		case "switched":
			fmt.Fprintf(os.Stdout, "switched device %s to %s\n", deviceID, supportType)
			if len(res.StoppedApps) > 0 {
				fmt.Fprintf(os.Stdout, "stopped %d app(s) bound to the device:\n", len(res.StoppedApps))
				for _, a := range res.StoppedApps {
					fmt.Fprintf(os.Stdout, "  - %s\n", appRefLabel(a))
				}
			}
			return nil
		case "unchanged":
			fmt.Fprintf(os.Stdout, "device %s already uses %s; nothing to do\n", deviceID, supportType)
			return nil
		case "bound-apps-stop-blocked":
			var b strings.Builder
			fmt.Fprintf(&b, "cannot switch device %s to %s: stop the following bound app(s) first", deviceID, supportType)
			for _, a := range res.BlockedApps {
				fmt.Fprintf(&b, "\n  - %s", appRefLabel(a))
			}
			return fmt.Errorf("%s", b.String())
		default:
			return fmt.Errorf("switch device %s: unexpected status %q from backend", deviceID, res.Status)
		}
	})
}

func appRefLabel(a supportTypeAppRef) string {
	label := a.AppName
	if a.Owner != "" {
		label += " (" + a.Owner + ")"
	}
	if a.Reason != "" {
		label += ": " + a.Reason
	} else if a.State != "" {
		label += " [" + a.State + "]"
	}
	return label
}
