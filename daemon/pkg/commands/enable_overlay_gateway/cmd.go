package enableoverlaygateway

import (
	"context"
	"os"
	"os/exec"

	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"k8s.io/klog/v2"
)

type enableOverlayGateway struct {
	commands.Operation
}

var _ commands.Interface = &enableOverlayGateway{}

func New() commands.Interface {
	return &enableOverlayGateway{
		Operation: commands.Operation{
			Name: commands.EnableOverlayGateway,
		},
	}
}

func (e *enableOverlayGateway) Execute(ctx context.Context, p any) (res any, err error) {
	// turn on the CNI-DHCP service
	cmd := exec.CommandContext(ctx, "systemctl", "enable", "--now", "cni-dhcp.service")
	cmd.Env = os.Environ()
	_, err = cmd.Output()
	if err != nil {
		klog.Errorf("overlay gateway enable: enable cni-dhcp.service failed: %v", err)
		return nil, err
	}

	// create the bridge connection
	err = utils.CreateBridgeConnection(ctx)
	utils.NotifyNetworkChanged()
	if err != nil {
		klog.Errorf("overlay gateway enable: create bridge connection failed: %v", err)
		return nil, err
	}

	return nil, nil
}
