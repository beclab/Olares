package overview

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// ----------------------------------------------------------------------------
// overview disk — sections (main + per-disk partitions)
// ----------------------------------------------------------------------------
// overview network — per-iface system-ifs table
// ----------------------------------------------------------------------------

func newOverviewNetworkCommand(f *cmdutil.Factory) *cobra.Command {
	var testConn bool
	cmd := &cobra.Command{
		Use:           "network",
		Short:         "Per-physical-NIC table from capi /system/ifs",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewNetwork(c.Context(), f, testConn)
		},
	}
	cmd.Flags().BoolVar(&testConn, "test-connectivity", true, "ask the BFF to probe internet/IPv6 connectivity per interface")
	return cmd
}

func runOverviewNetwork(ctx context.Context, f *cmdutil.Factory, testConn bool) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	r := &Runner{
		Flags:       common,
		Recommended: 60 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (Envelope, error) {
			ifs, err := fetchSystemIFS(ctx, c, testConn)
			if err != nil {
				return Envelope{Kind: KindOverviewNetwork}, err
			}
			items := make([]Item, 0, len(ifs))
			for i, it := range ifs {
				port := fmt.Sprintf("Port-%d", i+1)
				status := "down"
				if it.InternetConnected {
					status = "up"
				}
				raw := map[string]any{
					"port":           port,
					"iface":          it.Iface,
					"status":         status,
					"is_host_ip":     it.IsHostIp,
					"hostname":       it.Hostname,
					"method":         it.Method,
					"mtu":            it.MTU,
					"ip":             it.IP,
					"ipv4_mask":      it.IPv4Mask,
					"ipv4_gateway":   it.IPv4Gateway,
					"ipv4_dns":       it.IPv4DNS,
					"ipv6_address":   it.IPv6Address,
					"ipv6_gateway":   it.IPv6Gateway,
					"ipv6_dns":       it.IPv6DNS,
					"ipv4_connected": it.InternetConnected,
					"ipv6_connected": it.IPv6Connectivity,
					"tx_rate_raw":    it.TxRate,
					"rx_rate_raw":    it.RxRate,
				}
				disp := map[string]any{
					"port":         port,
					"iface":        it.Iface,
					"status":       status,
					"tx":           formatRateAny(it.TxRate),
					"rx":           formatRateAny(it.RxRate),
					"mtu":          fmt.Sprintf("%v", it.MTU),
					"method":       it.Method,
					"host":         it.Hostname,
					"ipv4":         it.IP,
					"ipv4_mask":    it.IPv4Mask,
					"ipv4_gateway": it.IPv4Gateway,
					"ipv4_dns":     it.IPv4DNS,
					"ipv6":         it.IPv6Address,
					"ipv6_gateway": it.IPv6Gateway,
					"ipv6_dns":     it.IPv6DNS,
				}
				items = append(items, Item{Raw: raw, Display: disp})
			}
			env := Envelope{
				Kind:  KindOverviewNetwork,
				Meta:  NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
				Items: HeadItems(items, common.Head),
			}
			env.Meta.RecommendedPollSeconds = 60
			if common.Output == OutputJSON {
				return env, nil
			}
			return env, writeNetworkTable(env)
		},
	}
	return r.Run(ctx)
}

func writeNetworkTable(env Envelope) error {
	cols := []TableColumn{
		{Header: "PORT", Get: func(it Item) string { return DisplayString(it, "port") }},
		{Header: "IFACE", Get: func(it Item) string { return DisplayString(it, "iface") }},
		{Header: "STATUS", Get: func(it Item) string { return DisplayString(it, "status") }},
		{Header: "TX", Get: func(it Item) string { return DisplayString(it, "tx") }},
		{Header: "RX", Get: func(it Item) string { return DisplayString(it, "rx") }},
		{Header: "MTU", Get: func(it Item) string { return DisplayString(it, "mtu") }},
		{Header: "METHOD", Get: func(it Item) string { return DisplayString(it, "method") }},
		{Header: "HOST", Get: func(it Item) string { return DisplayString(it, "host") }},
		{Header: "IPV4", Get: func(it Item) string { return DisplayString(it, "ipv4") }},
		{Header: "MASK", Get: func(it Item) string { return DisplayString(it, "ipv4_mask") }},
		{Header: "GW4", Get: func(it Item) string { return DisplayString(it, "ipv4_gateway") }},
		{Header: "DNS4", Get: func(it Item) string { return DisplayString(it, "ipv4_dns") }},
		{Header: "IPV6", Get: func(it Item) string { return DisplayString(it, "ipv6") }},
		{Header: "GW6", Get: func(it Item) string { return DisplayString(it, "ipv6_gateway") }},
		{Header: "DNS6", Get: func(it Item) string { return DisplayString(it, "ipv6_dns") }},
	}
	return WriteTable(os.Stdout, cols, env.Items)
}

// Helpers (sampleFloat / formatFloat / safeRatio / formatRateAny /
// toFloat / lastSampleFromRow) hoisted to cli/pkg/dashboard/numbers.go;
// the overview area re-exposes them via common.go bindings so leaf code
// keeps the lower-case names.
