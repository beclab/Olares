package handlers

import (
	"context"

	clistate "github.com/beclab/Olares/cli/pkg/daemon/state"
	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

// currentStateSnapshot is indirected so the handler can be exercised without a
// running state loop.
var currentStateSnapshot = state.Snapshot

// GetNodeStatus serves what this one node knows about itself. Every node
// answers it, including the master, so a caller needs a single protocol for
// the whole cluster.
func (h *Handlers) GetNodeStatus(ctx *fiber.Ctx) error {
	return h.OkJSON(ctx, "success", localNodeStatus(ctx.Context()))
}

// identityFrom combines the two halves of a node's identity: the part only the
// cluster knows (name, role) and the part the host knows on its own (hostname,
// device). A node cut off from the apiserver still reports the second half
// instead of reporting nothing.
func identityFrom(ctx context.Context, snap clistate.State) nodestatus.Identity {
	id := nodestatus.Identity{
		Role:       inventory.RoleUnknown,
		DeviceType: nodestatus.DeviceType(snap.DeviceName),
	}
	if snap.HostName != nil {
		id.Hostname = *snap.HostName
	}

	name, role, err := thisNodeInCluster(ctx)
	if err != nil {
		klog.Warning("resolve this node in the cluster error, ", err)
		return id
	}
	id.NodeName, id.Role = name, role
	return id
}

// thisNodeInCluster is indirected so both this handler and the master gate can
// be exercised without a cluster.
var thisNodeInCluster = func(ctx context.Context) (string, inventory.Role, error) {
	client, err := utils.GetKubeClient()
	if err != nil {
		return "", "", err
	}
	name, _, role, err := utils.GetThisNodeName(ctx, client)
	if err != nil {
		return "", "", err
	}
	return name, inventory.Role(role), nil
}
