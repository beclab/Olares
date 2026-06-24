package compute

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// `olares-cli settings compute list`
//
// Backed by /api/compute-resources, which returns an array of nodes, each
// with its accelerator devices and per-device app bindings. The table view
// groups by node and mirrors the SPA Accelerator page (node summary +
// per-device usage + bound apps). The DEVICE-ID column and the per-node
// header are the two values `compute set-type <node> <device>` takes.
func newListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list nodes, accelerator devices (with DEVICE-ID) and per-app bindings",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "list compute resources"); err != nil {
				return err
			}
			if err := requireComputeBackendVersion(ctx, f); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runList(ctx, f, output), "list compute resources")
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runList(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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

	var nodes []computeNode
	if err := doGetEnvelope(ctx, pc.doer, "/api/compute-resources", &nodes); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, nodes)
	default:
		return renderResources(os.Stdout, nodes)
	}
}

// renderResources groups output by node, mirroring the SPA Accelerator page:
// a node header with the dedicated-VRAM / shared-RAM summary, then a table of
// its devices (DEVICE-ID, support type, used/total memory, bound apps).
func renderResources(w io.Writer, nodes []computeNode) error {
	if len(nodes) == 0 {
		_, err := fmt.Fprintln(w, "no accelerator found (connect a GPU or other accelerator to your cluster to get started)")
		return err
	}
	for i, node := range nodes {
		if i > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		header := nodeHeaderLine(node)
		if _, err := fmt.Fprintln(w, header); err != nil {
			return err
		}
		if summary := nodeSummaryLine(node); summary != "" {
			if _, err := fmt.Fprintf(w, "  %s\n", summary); err != nil {
				return err
			}
		}
		if len(node.Devices) == 0 {
			if _, err := fmt.Fprintln(w, "  no devices"); err != nil {
				return err
			}
			continue
		}
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		if _, err := fmt.Fprintln(tw, "  DEVICE-ID\tNAME\tMODE\tSUPPORT-TYPE\tAVAILABLE\tMEM(Gi)\tHEALTH\tAPPS"); err != nil {
			return err
		}
		for _, d := range node.Devices {
			if _, err := fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				nonEmpty(d.ID),
				nonEmpty(d.deviceDisplayName()),
				nonEmpty(computeModeTitle(d.Mode)),
				nonEmpty(d.SupportType),
				deviceAvailableCell(d),
				fmt.Sprintf("%s / %s", computeMemoryValue(d.effectiveUsedMemory()), computeMemoryValue(d.Memory)),
				nonEmpty(d.Health),
				deviceAppsCell(d),
			); err != nil {
				return err
			}
		}
		if err := tw.Flush(); err != nil {
			return err
		}
	}
	return nil
}

func nodeHeaderLine(node computeNode) string {
	var b strings.Builder
	b.WriteString("Node: ")
	b.WriteString(nonEmpty(node.NodeName))
	badges := make([]string, 0, len(node.GpuTypes))
	for _, m := range node.GpuTypes {
		if t := computeModeTitle(m); t != "" {
			badges = append(badges, t)
		}
	}
	if len(badges) > 0 {
		b.WriteString("  [")
		b.WriteString(strings.Join(badges, ", "))
		b.WriteString("]")
	}
	if node.Health != "" {
		b.WriteString("  health=")
		b.WriteString(node.Health)
	}
	return b.String()
}

// nodeSummaryLine mirrors summaryTiles: Dedicated VRAM = sum of VRAM-mode
// device memory; Shared RAM = sum of non-VRAM, non-CPU device memory.
func nodeSummaryLine(node computeNode) string {
	var vramTotal, sharedTotal int64
	for _, d := range node.Devices {
		total := d.Memory
		if normalizeComputeMode(d.Mode) == "cpu" {
			continue
		}
		if isVramComputeMode(d.Mode) {
			vramTotal += total
		} else {
			sharedTotal += total
		}
	}
	tiles := make([]string, 0, 2)
	if vramTotal > 0 {
		tiles = append(tiles, "Dedicated VRAM: "+formatComputeMemory(vramTotal))
	}
	if sharedTotal > 0 {
		tiles = append(tiles, "Shared RAM: "+formatComputeMemory(sharedTotal))
	}
	return strings.Join(tiles, "  ")
}

// deviceAvailableCell lists the support types this device can switch to
// (device.availableSupportTypes, enum values) — the valid `set-type --type`
// arguments for this card. Not every device supports every type.
func deviceAvailableCell(d computeDevice) string {
	if len(d.AvailableSupportTypes) == 0 {
		return "-"
	}
	return strings.Join(d.AvailableSupportTypes, ",")
}

func deviceAppsCell(d computeDevice) string {
	if len(d.Bindings) == 0 {
		return "no app bound"
	}
	names := make([]string, 0, len(d.Bindings))
	for _, b := range d.Bindings {
		names = append(names, nonEmpty(b.AppName))
	}
	return strings.Join(names, ", ")
}
