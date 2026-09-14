package enableoverlaygateway

import (
	"context"

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

// Execute registers the overlay gateway on this node: the wired NIC gets the
// alternative name the underlay network attaches to, and the desired state is
// recorded. No NetworkManager connection is touched, so the host keeps its
// address and routes throughout.
func (e *enableOverlayGateway) Execute(ctx context.Context, p any) (res any, err error) {
	dev, err := utils.EnsureOverlayParentAltname(ctx)
	if err != nil {
		klog.Errorf("overlay gateway enable: prepare overlay parent failed: %v", err)
		return nil, err
	}
	if err := utils.SetOverlayGatewayDesired(true); err != nil {
		klog.Errorf("overlay gateway enable: %v", err)
		return nil, err
	}
	klog.Infof("overlay gateway enabled on %s (%s)", dev, utils.OverlayParentAltname)
	return nil, nil
}
