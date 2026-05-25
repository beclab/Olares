package overview

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
	pkgoverview "github.com/beclab/Olares/cli/pkg/dashboard/overview"

	"github.com/beclab/Olares/cli/cmd/ctl/dashboard/overview/disk"
	"github.com/beclab/Olares/cli/cmd/ctl/dashboard/overview/fan"
	"github.com/beclab/Olares/cli/cmd/ctl/dashboard/overview/gpu"
)

// `dashboard overview` — command tree assembly.
//
// Default action emits a sections envelope mirroring the SPA's overview
// page: { physical, user, ranking }. Each leaf below produces the same
// envelope shape it would emit standalone, so consumers can demux on
// `meta.kind` per section.
//
// Endpoint mapping (one helper per section + per leaf — see common.go):
//
//	overview (default)            — fan-out: physical + user + ranking
//	overview physical             — GET  /kapis/.../v1alpha3/cluster
//	overview user [<username>]    — GET  /kapis/.../v1alpha3/users/<u>
//	overview ranking              — workload-grain (FetchWorkloadsMetrics)
//	overview cpu                  — GET  /kapis/.../v1alpha3/nodes  (per-node multi-metric)
//	overview memory               — GET  /kapis/.../v1alpha3/nodes  (per-node, --mode physical|swap)
//	overview disk                 — sections: main + per-disk partitions
//	overview disk main            — GET  /kapis/.../v1alpha3/nodes  (per-disk metric)
//	overview disk partitions <d>  — GET  /kapis/.../v1alpha3/nodes  (per-partition metric)
//	overview pods                 — GET  /kapis/.../v1alpha3/nodes  (per-node count)
//	overview network              — GET  /capi/system/ifs           (per-iface system-ifs)
//	overview fan                  — sections: live + curve
//	overview fan live             — GET  /user-service/api/mdns/olares-one/cpu-gpu + graphics list
//	overview fan curve            — hardcoded fanCurveTable (helpers.go)
//	overview gpu list             — POST /hami/api/vgpu/v1/gpus
//	overview gpu tasks            — POST /hami/api/vgpu/v1/containers
//	overview gpu get <uuid>       — GET  /hami/api/vgpu/v1/gpu?uuid=...
//	overview gpu task <name> <uid>— GET  /hami/api/vgpu/v1/container?name=&podUid=
//
// Every leaf consumes CommonFlags (--output / --watch / --since / etc.)
// and returns one Envelope. Watch-able leaves wrap their fetch in Runner.

// NewOverviewCommand assembles the `dashboard overview` subtree. cf is
// the shared *pkgdashboard.CommonFlags pointer — once stored in the
// area's `common` package var, every leaf RunE reads through cobra's
// persistent-flag inheritance which mutates the pointed-at struct.
//
// The disk / fan / gpu subgroups live in their own Go subpackages so
// the directory tree mirrors the command tree. Wiring is strictly
// parent → child: this function imports `disk`, `fan`, `gpu`; none of
// them ever import `overview`.
func NewOverviewCommand(f *cmdutil.Factory, cf *pkgdashboard.CommonFlags) *cobra.Command {
	common = cf
	cmd := &cobra.Command{
		Use:   "overview",
		Short: "Sections envelope mirroring the SPA's overview page (physical / user / ranking)",
		Example: `  # Default — emit the three sections in parallel as a single envelope:
  olares-cli dashboard overview -o json

  # Just the workload-grain ranking:
  olares-cli dashboard overview ranking --sort desc`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				return unknownSubcommandRunE(c, args)
			}
			if err := common.Validate(); err != nil {
				return err
			}
			cli, err := prepareClient(c.Context(), f)
			if err != nil {
				return err
			}
			return pkgoverview.RunDefault(c.Context(), cli, common)
		},
	}

	cmd.AddCommand(newOverviewPhysicalCommand(f))
	cmd.AddCommand(newOverviewUserCommand(f))
	cmd.AddCommand(newOverviewRankingCommand(f))
	cmd.AddCommand(newOverviewCPUCommand(f))
	cmd.AddCommand(newOverviewMemoryCommand(f))
	cmd.AddCommand(disk.NewDiskCommand(f, cf))
	cmd.AddCommand(newOverviewPodsCommand(f))
	cmd.AddCommand(newOverviewNetworkCommand(f))
	cmd.AddCommand(fan.NewFanCommand(f, cf))
	cmd.AddCommand(gpu.NewGPUCommand(f, cf))
	return cmd
}
