package disk

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// ----------------------------------------------------------------------------

func NewDiskCommand(f *cmdutil.Factory, cf *pkgdashboard.CommonFlags) *cobra.Command {
	common = cf
	cmd := &cobra.Command{
		Use:   "disk",
		Short: "Sections envelope: main = per-disk table; partitions = per-device partition tables",
		Example: `  olares-cli dashboard overview disk -o json
  olares-cli dashboard overview disk main
  olares-cli dashboard overview disk partitions sda`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				return unknownSubcommandRunE(c, args)
			}
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewDiskDefault(c.Context(), f)
		},
	}
	cmd.AddCommand(newOverviewDiskMainCommand(f))
	cmd.AddCommand(newOverviewDiskPartitionsCommand(f))
	return cmd
}

func runOverviewDiskDefault(ctx context.Context, f *cmdutil.Factory) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	now := time.Now()

	mainEnv, mainErr := buildDiskMainEnvelope(ctx, c, now)
	partitionEnvs := map[string]Envelope{}

	if mainErr == nil {
		// One partitions section per device row in main.
		for _, it := range mainEnv.Items {
			device := DisplayString(it, "device")
			if device == "-" || device == "" {
				continue
			}
			env, err := buildDiskPartitionsEnvelope(ctx, c, device, now)
			if err != nil {
				env = Envelope{Kind: KindOverviewDiskPart}
				env.Meta.Error = err.Error()
				env.Meta.ErrorKind = ClassifyTransportErr(err)
			}
			env.Meta.FetchedAt = time.Now().In(common.Timezone.Time()).Format(time.RFC3339)
			partitionEnvs[device] = env
		}
	}

	sections := map[string]Envelope{
		"main": mainEnv,
	}
	if mainErr != nil {
		mainEnv.Kind = KindOverviewDiskMain
		mainEnv.Meta.Error = mainErr.Error()
		mainEnv.Meta.ErrorKind = ClassifyTransportErr(mainErr)
		sections["main"] = mainEnv
	}
	// Embed per-device partitions under a single envelope whose Sections
	// field is the device→partitions map. Lets consumers walk
	// sections.partitions.sda just like sections.main.
	partsEnv := Envelope{Kind: KindOverviewDiskPart, Sections: partitionEnvs}
	sections["partitions"] = partsEnv

	env := Envelope{
		Kind:     KindOverviewDisk,
		Meta:     NewMeta(time.Now().In(common.Timezone.Time()), c.OlaresID(), common.User),
		Sections: sections,
	}

	if common.Output == OutputJSON {
		return WriteJSON(os.Stdout, env)
	}
	// Table mode: render main, then for each device its partitions.
	fmt.Fprintln(os.Stdout, "== MAIN ==")
	if mainErr != nil {
		fmt.Fprintf(os.Stdout, "(error: %s)\n", mainErr)
	} else {
		_ = writeDiskMainTable(mainEnv)
	}
	for device, pEnv := range partitionEnvs {
		fmt.Fprintf(os.Stdout, "\n== PARTITIONS: %s ==\n", device)
		if pEnv.Meta.Error != "" {
			fmt.Fprintf(os.Stdout, "(error: %s)\n", pEnv.Meta.Error)
			continue
		}
		_ = writeDiskPartitionsTable(pEnv)
	}
	return nil
}
