package gpu

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/edge"
	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/olaresclient"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// `olares-cli settings gpu list`
//
// Dispatches through olaresclient.ComputeOps.ListAccelerators, which targets
// the version-appropriate endpoint (1.12.5: GET /api/gpu/list; 1.12.6: GET
// /api/compute-resources). The two payloads have different shapes, so the
// renderer branches on the dispatched client's version.
func NewListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list GPU devices / per-app bindings",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "list GPUs"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runList(ctx, f, output), "list GPUs")
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

// --- legacy 1.12.5 shape (/api/gpu/list) ---

type gpuApp struct {
	AppName string `json:"appName"`
	Memory  int    `json:"memory,omitempty"`
}

type gpuInfo struct {
	NodeName  string   `json:"nodeName"`
	ID        string   `json:"id"`
	Count     int      `json:"count"`
	DevMem    int      `json:"devmem"`
	DevCore   int      `json:"devcore"`
	Type      string   `json:"type"`
	Mode      string   `json:"mode"`
	ShareMode string   `json:"sharemode"`
	Health    bool     `json:"health"`
	Apps      []gpuApp `json:"apps,omitempty"`
}

// --- 1.12.6 compute-resources shape (/api/compute-resources) ---

type computeBinding struct {
	AppID    string `json:"appId,omitempty"`
	AppName  string `json:"appName"`
	Owner    string `json:"owner,omitempty"`
	Mode     string `json:"mode,omitempty"`
	NodeName string `json:"nodeName,omitempty"`
	DeviceID string `json:"deviceId"`
	Memory   int64  `json:"memory,omitempty"`
}

type computeDevice struct {
	ID                   string           `json:"id"`
	NodeName             string           `json:"nodeName"`
	CardModel            string           `json:"cardModel"`
	Memory               int64            `json:"memory"`
	Health               string           `json:"health,omitempty"`
	SupportType          string           `json:"supportType"`
	AvailableSupportType []string         `json:"availableSupportTypes,omitempty"`
	Bindings             []computeBinding `json:"bindings,omitempty"`
}

type computeNode struct {
	NodeName string          `json:"nodeName"`
	GPUType  string          `json:"gpuType,omitempty"`
	Modes    []string        `json:"modes,omitempty"`
	Devices  []computeDevice `json:"devices,omitempty"`
}

func runList(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	doer, _, err := edge.New(ctx, f)
	if err != nil {
		return err
	}

	return f.WithOlaresClient(ctx, func(c olaresclient.OlaresClient) error {
		raw, err := c.ListAccelerators(ctx, doer)
		if err != nil {
			return err
		}
		if usesComputeResources(c.Version()) {
			var nodes []computeNode
			if err := decodeData(raw, &nodes); err != nil {
				return err
			}
			if format == FormatJSON {
				return printJSON(os.Stdout, nodes)
			}
			return renderComputeTable(os.Stdout, nodes)
		}
		var rows []gpuInfo
		if err := decodeData(raw, &rows); err != nil {
			return err
		}
		if format == FormatJSON {
			return printJSON(os.Stdout, rows)
		}
		return renderLegacyGPUTable(os.Stdout, rows)
	})
}

func renderLegacyGPUTable(w io.Writer, rows []gpuInfo) error {
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, "no GPUs (no compatible devices, or HAMI is not running)")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NODE\tID\tTYPE\tMODE\tHEALTHY\tMEM(MiB)\tCORE\tAPPS"); err != nil {
		return err
	}
	for _, r := range rows {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%d\t%d\t%d\n",
			nonEmpty(r.NodeName), nonEmpty(r.ID), nonEmpty(r.Type), nonEmpty(r.Mode),
			boolStr(r.Health), r.DevMem, r.DevCore, len(r.Apps),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

// deviceUsedMemory sums allocated memory (bytes) across a device's bindings,
// mirroring the SPA's deviceUsedMemory helper.
func deviceUsedMemory(d computeDevice) int64 {
	var sum int64
	for _, b := range d.Bindings {
		sum += b.Memory
	}
	return sum
}

func renderComputeTable(w io.Writer, nodes []computeNode) error {
	hasDevice := false
	for _, n := range nodes {
		if len(n.Devices) > 0 {
			hasDevice = true
			break
		}
	}
	if !hasDevice {
		_, err := fmt.Fprintln(w, "no compute devices (no compatible accelerators detected)")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NODE\tDEVICE\tMODEL\tSUPPORT-TYPE\tMEM\tUSED\tHEALTH\tAPPS"); err != nil {
		return err
	}
	for _, n := range nodes {
		if len(n.Devices) == 0 {
			if _, err := fmt.Fprintf(tw, "%s\t-\t-\t-\t-\t-\t-\t-\n", nonEmpty(n.NodeName)); err != nil {
				return err
			}
			continue
		}
		for _, d := range n.Devices {
			apps := make([]string, 0, len(d.Bindings))
			for _, b := range d.Bindings {
				apps = append(apps, b.AppName)
			}
			appCol := "-"
			if len(apps) > 0 {
				appCol = strings.Join(apps, ",")
			}
			if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				nonEmpty(n.NodeName), nonEmpty(d.ID), nonEmpty(d.CardModel),
				nonEmpty(d.SupportType), formatMem(d.Memory), formatMem(deviceUsedMemory(d)),
				nonEmpty(d.Health), appCol,
			); err != nil {
				return err
			}
		}
	}
	return tw.Flush()
}
