package overview

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"

	"github.com/beclab/Olares/cli/cmd/ctl/dashboard/overview/disk"
	"github.com/beclab/Olares/cli/cmd/ctl/dashboard/overview/fan"
	"github.com/beclab/Olares/cli/cmd/ctl/dashboard/overview/gpu"
)

// ----------------------------------------------------------------------------
// `dashboard overview` — command tree assembly
// ----------------------------------------------------------------------------
//
// Default action emits a sections envelope mirroring the SPA's overview
// page: { physical, user, ranking }. Each leaf below produces the same
// envelope shape it would emit standalone, so consumers can demux on
// `meta.kind` per section.
//
// Endpoint mapping (one helper per section + per leaf — see helpers.go):
//
//	overview (default)            — fan-out: physical + user + ranking
//	overview physical             — GET  /kapis/.../v1alpha3/cluster
//	overview user [<username>]    — GET  /kapis/.../v1alpha3/users/<u>
//	overview ranking              — workload-grain (fetchWorkloadsMetrics)
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
// The disk / fan / gpu subgroups live in their own Go subpackages so the
// directory tree mirrors the command tree (settings precedent). Wiring is
// strictly parent → child: this function imports `disk`, `fan`, `gpu`;
// none of them ever import `overview`.
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
			return runOverviewDefault(c.Context(), f)
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

// runOverviewDefault is the aggregate action. Fans out the three SECTIONS
// in parallel; per-section failures populate Meta.Error on that section
// without aborting the whole envelope. Mirrors the SPA's "partial degradation
// is fine, surface it" behaviour on the overview page.
func runOverviewDefault(ctx context.Context, f *cmdutil.Factory) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	now := time.Now()
	type sectionResult struct {
		key string
		env Envelope
	}
	results := make(chan sectionResult, 3)

	go func() {
		env, err := buildPhysicalEnvelope(ctx, c, now)
		if err != nil {
			env.Kind = KindOverviewPhysical
			env.Meta.Error = err.Error()
			env.Meta.ErrorKind = ClassifyTransportErr(err)
		}
		results <- sectionResult{"physical", env}
	}()

	go func() {
		env, err := buildUserEnvelope(ctx, c, common.User, now)
		if err != nil {
			env.Kind = KindOverviewUser
			env.Meta.Error = err.Error()
			env.Meta.ErrorKind = ClassifyTransportErr(err)
		}
		results <- sectionResult{"user", env}
	}()

	go func() {
		env, err := buildRankingEnvelope(ctx, c, common.User, "desc", now)
		if err != nil {
			env.Kind = KindOverviewRanking
			env.Meta.Error = err.Error()
			env.Meta.ErrorKind = ClassifyTransportErr(err)
		}
		results <- sectionResult{"ranking", env}
	}()

	out := map[string]Envelope{}
	for i := 0; i < 3; i++ {
		r := <-results
		r.env.Meta.FetchedAt = time.Now().In(common.Timezone.Time()).Format(time.RFC3339)
		out[r.key] = r.env
	}

	env := Envelope{
		Kind:     KindOverview,
		Meta:     NewMeta(time.Now().In(common.Timezone.Time()), c.OlaresID(), common.User),
		Sections: out,
	}

	if common.Output == OutputJSON {
		return WriteJSON(os.Stdout, env)
	}
	// Table mode: render each section's table back-to-back, separated
	// by section banners. Lets a human eyeball the three views in one
	// shot.
	return writeOverviewSectionsTable(env)
}

// writeOverviewSectionsTable lays out the three sections back-to-back in
// table mode. Section banners use a leading "==" so a human scanning the
// scrollback can locate them by simple pattern match.
func writeOverviewSectionsTable(env Envelope) error {
	for _, key := range []string{"physical", "user", "ranking"} {
		section, ok := env.Sections[key]
		if !ok {
			continue
		}
		fmt.Fprintf(os.Stdout, "== %s ==\n", strings.ToUpper(key))
		if section.Meta.Error != "" {
			fmt.Fprintf(os.Stdout, "(error: %s)\n\n", section.Meta.Error)
			continue
		}
		switch section.Kind {
		case KindOverviewPhysical:
			if err := writePhysicalTable(section); err != nil {
				return err
			}
		case KindOverviewUser:
			if err := writeUserTable(section); err != nil {
				return err
			}
		case KindOverviewRanking:
			if err := writeRankingTable(section); err != nil {
				return err
			}
		}
		fmt.Fprintln(os.Stdout)
	}
	return nil
}

// ----------------------------------------------------------------------------
// overview physical — 9-row cluster metric table
// ----------------------------------------------------------------------------

// physicalMetric is one row of the SPA's Physical Resources panel. Columns:
// metric / value / unit / utilisation / detail. Names mirror the SPA's
// rendering conventions.
type physicalMetric struct {
	Key         string  // canonical metric key (cpu / memory / disk / pods / net_in / net_out)
	Label       string  // human-friendly metric name shown in column 1
	Value       float64 // headline numeric value (used / running)
	Total       float64 // total / quota
	Unit        string  // SPA unit suffix
	Utilisation float64 // 0..1 ratio
	Detail      string  // free-form detail string (used by net rows)
}
