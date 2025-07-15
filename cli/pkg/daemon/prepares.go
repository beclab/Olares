package daemon

import (
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"net"
)

type HostnameNotResolvable struct {
	common.KubePrepare
}

func (p *HostnameNotResolvable) PreCheck(runtime connector.Runtime) (bool, error) {
	ips, _ := net.LookupIP(runtime.GetSystemInfo().GetHostname())
	for _, ip := range ips {
		if ip.To4() != nil && ip.To4().String() == runtime.GetSystemInfo().GetLocalIp() {
			return false, nil
		}
	}
	return true, nil
}
