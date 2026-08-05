package state

import (
	"context"

	"github.com/beclab/Olares/daemon/pkg/utils"
	"k8s.io/klog/v2"
)

// thisNodeName resolves the Kubernetes name of the node olaresd runs on. It is
// a variable so the pressure lookup can be exercised without a cluster.
var thisNodeName = func(ctx context.Context) (string, error) {
	client, err := utils.GetKubeClient()
	if err != nil {
		return "", err
	}
	name, _, _, err := utils.GetThisNodeName(ctx, client)
	return name, err
}

// refreshNodePressure picks this node's entries out of the cluster-wide
// pressure map. Returning nothing when the node cannot name itself is
// deliberate: reporting no pressure is wrong, but reporting somebody else's is
// worse.
func refreshNodePressure(ctx context.Context, pressures map[string][]utils.NodePressure) []utils.NodePressure {
	name, err := thisNodeName(ctx)
	if err != nil {
		klog.Warning("resolve this node name for pressure lookup error, ", err)
		return nil
	}
	return pressureForNode(pressures, name)
}

// pressureForNode looks the node up by its Kubernetes node name, which is the
// key GetNodesPressure builds the map with. The OS hostname is a different
// string whenever kubelet was given a name of its own.
func pressureForNode(pressures map[string][]utils.NodePressure, nodeName string) []utils.NodePressure {
	if nodeName == "" {
		return nil
	}
	return pressures[nodeName]
}
