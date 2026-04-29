package gpu

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func newOverviewGPUTaskDetailFullCommand(f *cmdutil.Factory) *cobra.Command {
	var sharemode string
	cmd := &cobra.Command{
		Use:           "task-detail <name> <pod-uid>",
		Short:         "Per-task detail page (info + gauges + trends; SPA Overview2/GPU/TasksDetails)",
		Args:          cobra.ExactArgs(2),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewGPUTaskDetailFull(c.Context(), f, args[0], args[1], sharemode)
		},
	}
	cmd.Flags().StringVar(&sharemode, "sharemode", "", `task share mode ("0"=App exclusive, "1"=Memory slicing, "2"=Time slicing). When "2", allocation gauges are skipped to match the SPA.`)
	return cmd
}

func runOverviewGPUTaskDetailFull(ctx context.Context, f *cmdutil.Factory, name, podUID, sharemode string) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	r := &Runner{
		Flags:       common,
		Recommended: 30 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (Envelope, error) {
			start, end, since := resolveGPUDetailWindow(now, taskDetailDefaultSince)
			env, err := buildGPUTaskDetailFullEnvelope(ctx, c, name, podUID, sharemode, start, end, since)
			if err != nil {
				return env, err
			}
			if common.Output == OutputJSON {
				return env, nil
			}
			return env, writeGPUDetailFullTable(env)
		},
	}
	return r.Run(ctx)
}

// ----------------------------------------------------------------------------
// Table renderer (shared by gpu detail / task-detail)
// ----------------------------------------------------------------------------

// writeGPUDetailFullTable emits a three-block table view:
//
//	== Detail ==        — flat key/value list (the basic info card).
//	== Gauges ==        — N rows with title / value / unit / percent / used-total.
//	== Trends ==        — for each trend, a label header + (timestamp, value)
//	                      rows truncated to --head 16 by default.
//
// The output is intentionally text-only (no ANSI), so the bash smoke
// script can grep / awk through it.
func writeGPUDetailFullTable(env Envelope) error {
	out := os.Stdout
	if env.Meta.Empty {
		fmt.Fprintf(out, "(empty: %s", env.Meta.EmptyReason)
		if env.Meta.Note != "" {
			fmt.Fprintf(out, "; note: %s", env.Meta.Note)
		}
		fmt.Fprintln(out, ")")
		return nil
	}
	if env.Meta.Note != "" {
		fmt.Fprintf(os.Stderr, "(advisory) %s\n", env.Meta.Note)
	}
	if env.Meta.Window != nil {
		fmt.Fprintf(out, "Window: start=%s end=%s step=%s",
			env.Meta.Window.Start, env.Meta.Window.End, env.Meta.Window.Step)
		if env.Meta.Window.Since != "" {
			fmt.Fprintf(out, " since=%s", env.Meta.Window.Since)
		}
		fmt.Fprintln(out)
	}

	// Detail section.
	fmt.Fprintln(out, "\n== Detail ==")
	if dEnv, ok := env.Sections["detail"]; ok && len(dEnv.Items) > 0 {
		writeKeyValueTable(out, dEnv.Items[0])
	} else {
		fmt.Fprintln(out, "-")
	}

	// Gauges section.
	fmt.Fprintln(out, "\n== Gauges ==")
	if gEnv, ok := env.Sections["gauges"]; ok {
		cols := []TableColumn{
			{Header: "KEY", Get: func(it Item) string { return DisplayString(it, "key") }},
			{Header: "TITLE", Get: func(it Item) string { return DisplayString(it, "title") }},
			{Header: "VALUE", Get: func(it Item) string { return DisplayString(it, "value") }},
			{Header: "UNIT", Get: func(it Item) string { return DisplayString(it, "unit") }},
			{Header: "PERCENT", Get: func(it Item) string { return DisplayString(it, "percent") }},
			{Header: "USED/TOTAL", Get: func(it Item) string { return DisplayString(it, "used_total") }},
		}
		_ = WriteTable(out, cols, gEnv.Items)
	}

	// Trends section.
	fmt.Fprintln(out, "\n== Trends ==")
	if tEnv, ok := env.Sections["trends"]; ok {
		head := common.Head
		if head <= 0 {
			head = 16 // SPA renders ~16 buckets in the chart
		}
		for _, it := range tEnv.Items {
			title := DisplayString(it, "title")
			unit := DisplayString(it, "unit")
			fmt.Fprintf(out, "\n-- %s (%s) --\n", title, unit)
			if errStr, ok := it.Raw["error"].(string); ok && errStr != "" {
				fmt.Fprintf(out, "(error: %s)\n", errStr)
				continue
			}
			lines, _ := it.Raw["lines"].([]map[string]any)
			if len(lines) == 0 {
				fmt.Fprintln(out, "-")
				continue
			}
			for _, ln := range lines {
				label, _ := ln["label"].(string)
				fmt.Fprintf(out, "  %s:\n", label)
				points, _ := ln["points"].([]map[string]any)
				if len(points) == 0 {
					fmt.Fprintln(out, "    -")
					continue
				}
				rendered := points
				if head > 0 && head < len(points) {
					rendered = points[:head]
				}
				for _, p := range rendered {
					ts, _ := p["timestamp"].(string)
					v := p["value"]
					fmt.Fprintf(out, "    %s\t%v\n", ts, v)
				}
				if len(points) > len(rendered) {
					fmt.Fprintf(out, "    ... (%d more rows; pass --head 0 for full)\n", len(points)-len(rendered))
				}
			}
		}
	}

	if len(env.Meta.Warnings) > 0 {
		fmt.Fprintln(out, "\n== Warnings ==")
		for _, w := range env.Meta.Warnings {
			fmt.Fprintf(out, "- %s\n", w)
		}
	}
	return nil
}

// writeKeyValueTable renders one Item as a vertical key/value table.
// Sorted lexicographically so output is deterministic across runs.
func writeKeyValueTable(out *os.File, it Item) {
	if it.Display == nil {
		fmt.Fprintln(out, "-")
		return
	}
	keys := make([]string, 0, len(it.Display))
	for k := range it.Display {
		keys = append(keys, k)
	}
	sortStrings(keys)
	for _, k := range keys {
		v := it.Display[k]
		fmt.Fprintf(out, "%s\t%v\n", k, v)
	}
}

// sortStrings is a tiny helper to avoid pulling sort into this file's
// imports just for one site.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}
