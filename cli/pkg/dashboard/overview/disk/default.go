package disk

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// RunDefault is the cmd-side entry point for `dashboard overview
// disk` without a subverb. Emits a sections envelope: one `main`
// section + one `partitions` section whose Sections map keys by
// device. Per-device partition fetches are sequential because the
// device list is itself derived from the main fetch — fan-out
// concurrency would be a wash and complicate error fan-in.
//
// One-shot only: the per-section recommended cadences differ
// (lsblk vs SMART) and a unified --watch-interval would lie to
// consumers. Mirrors the overview-default policy.
func RunDefault(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags) error {
	now := time.Now()
	env := BuildSectionsEnvelope(ctx, c, cf, now)
	if cf.Output == pkgdashboard.OutputJSON {
		return pkgdashboard.WriteJSON(os.Stdout, env)
	}
	return WriteSectionsTable(os.Stdout, env)
}

// BuildSectionsEnvelope is the per-iteration aggregator. main goes
// first (its row list seeds the per-device partition fetches);
// partition errors land on the per-device section's Meta.Error
// rather than aborting the whole envelope so a transient lsblk
// failure on one device doesn't blank the table for the other
// devices.
func BuildSectionsEnvelope(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, now time.Time) pkgdashboard.Envelope {
	mainEnv, mainErr := BuildMainEnvelope(ctx, c, cf, now)
	partitionEnvs := map[string]pkgdashboard.Envelope{}

	if mainErr == nil {
		for _, it := range mainEnv.Items {
			device := pkgdashboard.DisplayString(it, "device")
			if device == "-" || device == "" {
				continue
			}
			pEnv, err := BuildPartitionsEnvelope(ctx, c, cf, device, now)
			if err != nil {
				pEnv = pkgdashboard.Envelope{Kind: pkgdashboard.KindOverviewDiskPart}
				pEnv.Meta.Error = err.Error()
				pEnv.Meta.ErrorKind = pkgdashboard.ClassifyTransportErr(err)
			}
			pEnv.Meta.FetchedAt = time.Now().In(cf.Timezone.Time()).Format(time.RFC3339)
			partitionEnvs[device] = pEnv
		}
	}

	sections := map[string]pkgdashboard.Envelope{
		"main": mainEnv,
	}
	if mainErr != nil {
		mainEnv.Kind = pkgdashboard.KindOverviewDiskMain
		mainEnv.Meta.Error = mainErr.Error()
		mainEnv.Meta.ErrorKind = pkgdashboard.ClassifyTransportErr(mainErr)
		sections["main"] = mainEnv
	}
	// Embed per-device partitions under a single envelope whose
	// Sections field is the device→partitions map. Lets consumers
	// walk sections.partitions.sda just like sections.main.
	sections["partitions"] = pkgdashboard.Envelope{
		Kind:     pkgdashboard.KindOverviewDiskPart,
		Sections: partitionEnvs,
	}
	return pkgdashboard.Envelope{
		Kind:     pkgdashboard.KindOverviewDisk,
		Meta:     pkgdashboard.NewMeta(time.Now().In(cf.Timezone.Time()), c.OlaresID(), cf.User),
		Sections: sections,
	}
}

// WriteSectionsTable renders the sections envelope to a single
// scrollback chunk. Layout:
//
//	== MAIN ==
//	<main table>
//
//	== PARTITIONS: <device> ==
//	<lsblk subtree table>
//
// Per-device partitions are emitted in alphabetical-device order so
// the human view stays deterministic across runs (the upstream map
// has no ordering guarantee).
func WriteSectionsTable(w io.Writer, env pkgdashboard.Envelope) error {
	mainEnv, hasMain := env.Sections["main"]
	partsEnv, hasParts := env.Sections["partitions"]

	fmt.Fprintln(w, "== MAIN ==")
	switch {
	case !hasMain:
		fmt.Fprintln(w, "(missing)")
	case mainEnv.Meta.Error != "":
		fmt.Fprintf(w, "(error: %s)\n", mainEnv.Meta.Error)
	default:
		if err := WriteMainTable(w, mainEnv); err != nil {
			return err
		}
	}
	if !hasParts {
		return nil
	}
	devices := make([]string, 0, len(partsEnv.Sections))
	for d := range partsEnv.Sections {
		devices = append(devices, d)
	}
	sort.Strings(devices)
	for _, device := range devices {
		pEnv := partsEnv.Sections[device]
		fmt.Fprintf(w, "\n== PARTITIONS: %s ==\n", device)
		if pEnv.Meta.Error != "" {
			fmt.Fprintf(w, "(error: %s)\n", pEnv.Meta.Error)
			continue
		}
		if err := WritePartitionsTable(w, pEnv); err != nil {
			return err
		}
	}
	return nil
}
