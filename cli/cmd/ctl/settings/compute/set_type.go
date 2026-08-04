package compute

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// validSupportTypes is the canonical DeviceSupportType set (compute.ts).
var validSupportTypes = []string{"Exclusive", "MemorySlice", "TimeSlice", "MemoryShared"}

func isValidSupportType(t string) bool {
	for _, v := range validSupportTypes {
		if v == t {
			return true
		}
	}
	return false
}

// `olares-cli settings compute set-type <node> <device> --type X`
//
// Backed by PUT /api/compute-resources/nodes/<node>/devices/<device>/support-type.
// Mirrors the SPA's ManageNodePage support-type dropdown: switching a device
// that has bound apps unbinds and STOPS those apps; the backend may also
// refuse with a bound-apps-stop-blocked outcome (apps that cannot be stopped).
//
// <node> and <device> are the NODE header and DEVICE-ID column from
// `olares-cli settings compute list`.
func newSetTypeCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		supportType string
		skipConfirm bool
		output      string
	)
	cmd := &cobra.Command{
		Use:   "set-type <node> <device>",
		Short: "switch a device's support type (may stop bound apps)",
		Long: fmt.Sprintf(`Switch an accelerator device's support type.

<node> and <device> are the NODE (shown in each node header) and DEVICE-ID
(first column) from "olares-cli settings compute list".

Valid --type values: %s.
The human labels shown by "list" are also accepted (e.g. "Memory Slicing"
== "MemorySlice"). Not every device supports every type — see the AVAILABLE
column in "list" for what a given device allows.

This mirrors the SPA's support-type dropdown. If the device currently has
bound apps, switching unbinds and STOPS those apps — you must type "yes" to
confirm (or pass --yes). The backend may refuse the switch when a bound app
cannot be stopped (bound-apps-stop-blocked); in that case nothing changes and
the command exits non-zero, listing the blocking app(s).
`, strings.Join(validSupportTypes, ", ")),
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "set compute support type"); err != nil {
				return err
			}
			if err := requireComputeBackendVersion(ctx, f); err != nil {
				return err
			}
			node := strings.TrimSpace(args[0])
			device := strings.TrimSpace(args[1])
			if node == "" || device == "" {
				return fmt.Errorf("both <node> and <device> are required")
			}
			if !c.Flags().Changed("type") {
				return fmt.Errorf("--type is required (one of: %s)", strings.Join(validSupportTypes, ", "))
			}
			target, ok := canonicalSupportType(supportType)
			if !ok {
				return fmt.Errorf("--type: invalid value %q (allowed: %s; labels like %q are also accepted)",
					supportType, strings.Join(validSupportTypes, ", "), acceleratorSupportTypeLabel("MemorySlice"))
			}
			format, err := parseFormat(output)
			if err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runSetType(ctx, f, node, device, target, skipConfirm, format), "set compute support type")
		},
	}
	cmd.Flags().StringVar(&supportType, "type", "", fmt.Sprintf("target support type: %s", strings.Join(validSupportTypes, " | ")))
	cmd.Flags().BoolVar(&skipConfirm, "yes", false, "skip interactive confirmation when bound apps would be stopped")
	addOutputFlag(cmd, &output)
	cmd.Flags().SortFlags = false
	return cmd
}

func runSetType(ctx context.Context, f *cmdutil.Factory, node, device, target string, skipConfirm bool, format Format) error {
	if ctx == nil {
		ctx = context.Background()
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	// Inspect the device first (same data the SPA dropdown is built from):
	// short-circuit a no-op switch, validate against availableSupportTypes,
	// and learn the bound apps so we can warn before stopping them.
	var nodes []computeNode
	if err := doGetEnvelope(ctx, pc.doer, "/api/compute-resources", &nodes); err != nil {
		return err
	}
	dev := findDevice(nodes, node, device)
	if dev == nil {
		return fmt.Errorf("device %q not found on node %q (see `olares-cli settings compute list`)", device, node)
	}
	if dev.SupportType == target {
		return printSetTypeOutcome(format, node, device, target, &supportTypeResult{Status: "unchanged"})
	}
	if len(dev.AvailableSupportTypes) > 0 && !containsStr(dev.AvailableSupportTypes, target) {
		return fmt.Errorf("device %q does not support %q; available: %s",
			device, target, strings.Join(dev.AvailableSupportTypes, ", "))
	}

	boundApps := make([]string, 0, len(dev.Bindings))
	for _, b := range dev.Bindings {
		boundApps = append(boundApps, b.AppName)
	}
	if len(boundApps) > 0 && !skipConfirm {
		fmt.Fprintf(os.Stderr,
			"Switching device %q to %q will unbind and STOP the following app(s):\n  %s\n"+
				"Type 'yes' to continue: ",
			device, acceleratorSupportTypeLabel(target), strings.Join(boundApps, ", "))
		line, rerr := bufio.NewReader(os.Stdin).ReadString('\n')
		if rerr != nil {
			return rerr
		}
		if strings.TrimSpace(line) != "yes" {
			return fmt.Errorf("aborting support-type switch: confirmation was not yes")
		}
	}

	path := fmt.Sprintf("/api/compute-resources/nodes/%s/devices/%s/support-type",
		url.PathEscape(node), url.PathEscape(device))
	var raw json.RawMessage
	if err := doMutateEnvelope(ctx, pc.doer, "PUT", path, map[string]string{"supportType": target}, &raw); err != nil {
		return err
	}
	result := parseSupportTypeResult(raw)

	if result.Status == "bound-apps-stop-blocked" {
		// Mode was NOT switched; surface the blocking apps and fail.
		return blockedError(result.BlockedApps)
	}
	return printSetTypeOutcome(format, node, device, target, result)
}

// parseSupportTypeResult decodes the discriminated wire shape: the inner data
// is either a supportTypeResult directly, or wrapped as
// {type: 'computeDeviceSwitchBlocked', Data: {...}}.
func parseSupportTypeResult(raw json.RawMessage) *supportTypeResult {
	if len(raw) == 0 {
		return &supportTypeResult{Status: "switched"}
	}
	var env supportTypeEnvelope
	if err := json.Unmarshal(raw, &env); err == nil && env.Type == "computeDeviceSwitchBlocked" && env.Data != nil {
		return env.Data
	}
	var res supportTypeResult
	if err := json.Unmarshal(raw, &res); err == nil && res.Status != "" {
		return &res
	}
	return &supportTypeResult{Status: "switched"}
}

func blockedError(blocked []computeAppRef) error {
	var b strings.Builder
	b.WriteString("unable to switch support type: the following bound app(s) could not be stopped:")
	if len(blocked) == 0 {
		b.WriteString(" (no detail returned)")
		return fmt.Errorf("%s", b.String())
	}
	for _, a := range blocked {
		reason := strings.TrimSpace(a.Reason)
		if reason == "" {
			reason = "unable to switch mode"
		}
		b.WriteString(fmt.Sprintf("\n  - %s: %s", nonEmpty(a.AppName), reason))
	}
	return fmt.Errorf("%s", b.String())
}

func printSetTypeOutcome(format Format, node, device, target string, result *supportTypeResult) error {
	stopped := make([]string, 0, len(result.StoppedApps))
	for _, a := range result.StoppedApps {
		stopped = append(stopped, a.AppName)
	}
	switch format {
	case FormatJSON:
		out := map[string]interface{}{
			"node":        node,
			"device":      device,
			"supportType": target,
			"status":      result.Status,
		}
		if len(stopped) > 0 {
			out["stoppedApps"] = stopped
		}
		return printJSON(os.Stdout, out)
	default:
		switch result.Status {
		case "unchanged":
			_, err := fmt.Fprintf(os.Stdout, "device %q on node %q already uses %q; no change\n",
				device, node, acceleratorSupportTypeLabel(target))
			return err
		default:
			msg := fmt.Sprintf("switched device %q on node %q to %q", device, node, acceleratorSupportTypeLabel(target))
			if len(stopped) > 0 {
				msg += fmt.Sprintf(" (stopped app(s): %s)", strings.Join(stopped, ", "))
			}
			_, err := fmt.Fprintln(os.Stdout, msg)
			return err
		}
	}
}

func findDevice(nodes []computeNode, node, device string) *computeDevice {
	for ni := range nodes {
		if nodes[ni].NodeName != node {
			continue
		}
		for di := range nodes[ni].Devices {
			if nodes[ni].Devices[di].ID == device {
				return &nodes[ni].Devices[di]
			}
		}
	}
	return nil
}

func containsStr(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
