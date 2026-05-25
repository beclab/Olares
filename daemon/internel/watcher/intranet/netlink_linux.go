//go:build linux
// +build linux

package intranet

import (
	"fmt"
	"os/exec"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"k8s.io/klog/v2"
)

func getPodNeighborInfo(podIp string) (mac, iface string, err error) {
	// family: unix.AF_INET for IPv4, unix.AF_INET6 for IPv6
	neighs, err := netlink.NeighList(0, unix.AF_INET) // 0 => all links
	if err != nil {
		klog.Error("list neighbor error, ", err)
		return
	}

	for _, n := range neighs {
		if n.IP.String() == podIp && n.State == netlink.NUD_REACHABLE {
			mac = n.HardwareAddr.String()
			if mac == "<nil>" {
				mac = ""
			}

			if link, err := netlink.LinkByIndex(n.LinkIndex); err == nil {
				iface = link.Attrs().Name
			}

			return
		}
	}

	// try to refresh neighbor table
	go func() {
		cmd := exec.Command("ping", "-c", "3", podIp)
		err := cmd.Run()
		if err != nil {
			klog.Error("ping pod ip to refresh neighbor table error, ", err)
			return
		}
	}()

	return "", "", fmt.Errorf("not found pod neighbor info for ip %s", podIp)
}
