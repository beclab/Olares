package overview

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// RunNetwork is the cmd-side entry point for `dashboard overview
// network`. Owns the watch-aware Runner so the cmd-side leaf stays a
// thin shell. testConn maps to --test-connectivity (asks the BFF to
// probe internet/IPv6 connectivity per interface).
func RunNetwork(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, testConn bool) error {
	r := &pkgdashboard.Runner{
		Flags:       cf,
		Recommended: 60 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (pkgdashboard.Envelope, error) {
			env, err := BuildNetworkEnvelope(ctx, c, cf, testConn, now)
			if err != nil {
				return env, err
			}
			if cf.Output == pkgdashboard.OutputJSON {
				return env, nil
			}
			return env, WriteNetworkTable(os.Stdout, env)
		},
	}
	return r.Run(ctx)
}

// BuildNetworkEnvelope fans out to /capi/system/ifs and emits one
// Item per physical NIC. Synthesises the Port-N column locally
// because the SPA renders the row index that way and agents would
// otherwise need a second join. Exported so the package's _test.go
// can drive it against a stub upstream without going through the
// watch loop.
func BuildNetworkEnvelope(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, testConn bool, now time.Time) (pkgdashboard.Envelope, error) {
	ifs, err := pkgdashboard.FetchSystemIFS(ctx, c, testConn)
	if err != nil {
		return pkgdashboard.Envelope{Kind: pkgdashboard.KindOverviewNetwork}, err
	}
	items := make([]pkgdashboard.Item, 0, len(ifs))
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
		items = append(items, pkgdashboard.Item{Raw: raw, Display: disp})
	}
	env := pkgdashboard.Envelope{
		Kind:  pkgdashboard.KindOverviewNetwork,
		Meta:  pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User),
		Items: pkgdashboard.HeadItems(items, cf.Head),
	}
	env.Meta.RecommendedPollSeconds = 60
	return env, nil
}

// WriteNetworkTable renders env.Items as the SPA-aligned 15-column
// per-NIC table. Column order is pinned: agents who scrape the
// rendered scrollback need the IPv4 / IPv6 sub-fields to land in
// the same positions across releases.
func WriteNetworkTable(w io.Writer, env pkgdashboard.Envelope) error {
	cols := []pkgdashboard.TableColumn{
		{Header: "PORT", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "port") }},
		{Header: "IFACE", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "iface") }},
		{Header: "STATUS", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "status") }},
		{Header: "TX", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "tx") }},
		{Header: "RX", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "rx") }},
		{Header: "MTU", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "mtu") }},
		{Header: "METHOD", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "method") }},
		{Header: "HOST", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "host") }},
		{Header: "IPV4", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "ipv4") }},
		{Header: "MASK", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "ipv4_mask") }},
		{Header: "GW4", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "ipv4_gateway") }},
		{Header: "DNS4", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "ipv4_dns") }},
		{Header: "IPV6", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "ipv6") }},
		{Header: "GW6", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "ipv6_gateway") }},
		{Header: "DNS6", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "ipv6_dns") }},
	}
	return pkgdashboard.WriteTable(w, cols, env.Items)
}
