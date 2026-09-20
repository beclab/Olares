package inventory

import (
	"context"

	"github.com/beclab/Olares/daemon/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Observation is what the cluster itself says about a node right now.
//
// It is deliberately smaller than Node: watching a reboot needs no addresses
// and no roles, and reading it from the master's own kube client keeps the
// watch off every node's user-facing endpoint, which would otherwise need a
// user credential to answer the question "did you restart".
type Observation struct {
	Ready bool

	// BootID is status.nodeInfo.bootID, the kernel's identifier for the boot
	// the node is currently on. A node that rebooted comes back with another
	// one, which is the only proof of a restart that does not rely on having
	// watched the machine the whole time it was gone.
	BootID string
}

// Observe reads the cluster's view using the daemon's shared kube client.
func Observe(ctx context.Context) (map[string]Observation, error) {
	client, err := utils.GetKubeClient()
	if err != nil {
		return nil, err
	}
	return (&Lister{Client: client}).Observe(ctx)
}

// Observe reports every node the cluster currently knows, keyed by node name.
// A node that has left the directory is absent rather than not-ready: telling
// the two apart is what lets a reboot see the node go away.
func (l *Lister) Observe(ctx context.Context) (map[string]Observation, error) {
	list, err := l.Client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	seen := make(map[string]Observation, len(list.Items))
	for i := range list.Items {
		n := &list.Items[i]
		seen[n.Name] = Observation{
			Ready:  utils.IsNodeReady(n),
			BootID: n.Status.NodeInfo.BootID,
		}
	}
	return seen, nil
}
