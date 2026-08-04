package lpvpndns

import (
	"context"
	"net"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// detectPodCIDR mirrors corefile.go's primary probe: parse the cluster Pod
// CIDR from the kube-proxy KUBE-SERVICES chain. olaresd runs as root so the
// iptables read works on host. Falls back to defaultPodCIDR.
func detectPodCIDR(ctx context.Context, _ kubernetes.Interface) string {
	if cidr, ok := podCIDRFromIptables(ctx); ok {
		klog.V(2).Infof("lpvpndns: detectPodCIDR KUBE-SERVICES %s", cidr)
		return cidr
	}
	klog.V(2).Infof("lpvpndns: detectPodCIDR fallback %s", defaultPodCIDR)
	return defaultPodCIDR
}

func podCIDRFromIptables(ctx context.Context) (string, bool) {
	out, err := runCmd(ctx, "iptables", "-t", natTable, "-S", "KUBE-SERVICES")
	if err != nil {
		klog.V(4).Infof("lpvpndns: detectPodCIDR iptables unavailable: %v", err)
		return "", false
	}
	return parsePodCIDRFromKubeServicesIPTables(string(out))
}

// parsePodCIDRFromKubeServicesIPTables extracts the `! -s <cidr>` Pod CIDR from
// the KUBE-SERVICES masquerade rule. Pure function for unit testing.
func parsePodCIDRFromKubeServicesIPTables(output string) (string, bool) {
	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, "KUBE-SERVICES") || !strings.Contains(line, "KUBE-MARK-MASQ") {
			continue
		}
		fields := strings.Fields(line)
		for i := 0; i < len(fields)-2; i++ {
			if fields[i] != "!" || fields[i+1] != "-s" {
				continue
			}
			cidr := fields[i+2]
			if _, _, err := net.ParseCIDR(cidr); err == nil {
				return cidr, true
			}
		}
	}
	return "", false
}
