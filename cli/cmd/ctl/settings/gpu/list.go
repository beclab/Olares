package gpu

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings gpu list`
//
// Backed by user-service's /api/gpu/list, which proxies HAMI. The body
// is a BFL envelope around an array of GPUInfo:
//
//   {
//     "nodeName": "...", "id": "...", "type": "...",
//     "count": 1, "devmem": 12288, "devcore": 80,
//     "mode": "exclusive" | "shared",
//     "sharemode": "...",
//     "health": true,
//     "memoryAllocated": 0, "memoryAvailable": 0,
//     "apps": [{"appName": "...", "memory": <int>}],
//   }
//
// Phase 2 will add `mode set`, `assign`, `unassign`, `bulk-assign`.
func NewListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list GPUs / per-app assignment",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runList(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

type gpuApp struct {
	AppName string `json:"appName"`
	Memory  int    `json:"memory,omitempty"`
}

type gpuInfo struct {
	NodeName        string   `json:"nodeName"`
	ID              string   `json:"id"`
	Count           int      `json:"count"`
	DevMem          int      `json:"devmem"`
	DevCore         int      `json:"devcore"`
	Type            string   `json:"type"`
	Mode            string   `json:"mode"`
	ShareMode       string   `json:"sharemode"`
	Health          bool     `json:"health"`
	MemoryAvailable int      `json:"memoryAvailable,omitempty"`
	MemoryAllocated int      `json:"memoryAllocated,omitempty"`
	Apps            []gpuApp `json:"apps,omitempty"`
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

	var rows []gpuInfo
	if err := doGetEnvelope(ctx, pc.doer, "/api/gpu/list", &rows); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, rows)
	default:
		return renderGPUTable(os.Stdout, rows)
	}
}

func renderGPUTable(w io.Writer, rows []gpuInfo) error {
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, "no GPUs (no compatible devices, or HAMI is not running)")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NODE\tID\tTYPE\tMODE\tHEALTHY\tMEM(MiB)\tCORE\tAPPS"); err != nil {
		return err
	}
	for _, r := range rows {
		appCount := len(r.Apps)
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%d\t%d\t%d\n",
			nonEmpty(r.NodeName),
			nonEmpty(r.ID),
			nonEmpty(r.Type),
			nonEmpty(r.Mode),
			boolStr(r.Health),
			r.DevMem,
			r.DevCore,
			appCount,
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}
