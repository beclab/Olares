package terminus

import (
	"fmt"
	"net"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"

	"github.com/pkg/errors"
)

// CheckMasterReachable verifies that the master node's SSH endpoint is
// reachable before we attempt to run any remote command on it, so an
// unreachable master fails fast with an actionable message instead of a raw
// SSH/connector stack trace.
type CheckMasterReachable struct {
	common.KubeAction
}

func (a *CheckMasterReachable) Execute(runtime connector.Runtime) error {
	host := a.KubeConf.Arg.MasterHost
	if host == "" {
		return errors.New("master host is not provided")
	}
	port := a.KubeConf.Arg.MasterSSHPort
	if port == 0 {
		port = 22
	}
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return errors.Errorf(
			"failed to reach the master node at %s over SSH: %v\n"+
				"please make sure that:\n"+
				"  - the master node is powered on and reachable from this machine\n"+
				"  - the SSH service is running and listening on the port (default 22)\n"+
				"  - the --master-host / --master-ssh-port values are correct",
			addr, err)
	}
	_ = conn.Close()
	return nil
}
