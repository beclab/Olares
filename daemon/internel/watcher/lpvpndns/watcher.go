// Package lpvpndns maintains a master-only deterministic SNAT chain so that
// node / hostNetwork tailscale Pod DNS queries to CoreDNS are sourced from the
// CoreDNS vpn view whitelist IP, making them hit the vpn view and resolve to
// Tailnet IPs.
package lpvpndns

import (
	"context"
	"net"

	"github.com/beclab/Olares/daemon/internel/watcher"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"k8s.io/klog/v2"
)

var _ watcher.Watcher = &lpvpnDNSWatcher{}

type lpvpnDNSWatcher struct{}

func NewWatcher() *lpvpnDNSWatcher {
	return &lpvpnDNSWatcher{}
}

// Watch reconciles the SNAT chain once per tick.
//  1. master hard gate: non-master / role lookup failure -> teardown + return.
//  2. running gate: not TerminusRunning -> skip (do NOT teardown, transient).
//  3. vpnIP: master node InternalIP (same field corefile.go writes into vpn view).
//  4. detect podCIDR.
//  5. reconcile.
func (w *lpvpnDNSWatcher) Watch(ctx context.Context) {
	kube, err := utils.GetKubeClient()
	if err != nil {
		klog.Warning("lpvpndns: get kube client error, ", err)
		return
	}

	// ① master hard gate (fail-closed): never default to master.
	_, nodeIP, nodeRole, err := utils.GetThisNodeName(ctx, kube)
	if err != nil {
		klog.Warningf("lpvpndns: get node role failed: %v, teardown", err)
		teardown(ctx)
		return
	}
	if nodeRole != "master" {
		klog.V(4).Infof("lpvpndns: not master (role=%q), teardown", nodeRole)
		teardown(ctx)
		return
	}

	// ② running gate: transient non-running states must not flap the chain.
	if state.CurrentState.TerminusState != state.TerminusRunning {
		return
	}

	// ③ vpnIP = master node InternalIP
	vpnIP := nodeIP
	if net.ParseIP(vpnIP) == nil {
		klog.Warningf("lpvpndns: no valid vpn IP from node InternalIP=%q, teardown", nodeIP)
		teardown(ctx)
		return
	}

	// ④ detect podCIDR (SNAT anchor, drift-resistant to CoreDNS Pod IP changes).
	podCIDR := detectPodCIDR(ctx, kube)

	// ⑤ converge.
	reconcile(ctx, vpnIP, podCIDR)
}
