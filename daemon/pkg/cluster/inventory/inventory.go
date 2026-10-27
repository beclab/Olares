// Package inventory is the master olaresd's node directory: the full list of
// Kubernetes nodes with the facts needed to address each node's olaresd.
//
// Nothing here is filtered out. A NotReady node and a node with no internal IP
// are both first-class entries: dropping them would make a node vanish from the
// UI exactly when the user needs to see that it is in trouble. Consumers that
// only want Ready, addressable nodes (for example collect-logs) filter after
// List; fanout itself does not discover nodes.
package inventory

import (
	"context"
	"os"
	"strings"

	"github.com/beclab/Olares/daemon/pkg/nets"
	"github.com/beclab/Olares/daemon/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// Role is the Kubernetes role of a node. It is deliberately separate from the
// device type: what a node is and what it does are different questions.
type Role string

const (
	RoleMaster Role = "master"
	RoleWorker Role = "worker"

	// RoleUnknown is used when a node cannot determine its own role, e.g. a
	// node-local status call made while the cluster is unreachable.
	RoleUnknown Role = "unknown"
)

// hostnameLabel is set by kubelet on every node it registers, and holds the
// machine hostname even when the node was registered under another name.
const hostnameLabel = "kubernetes.io/hostname"

// Node is one entry of the node directory.
type Node struct {
	NodeName string `json:"nodeName"`
	Role     Role   `json:"role"`

	// IP is the internal address other nodes reach this node's olaresd on.
	// Empty when the node has no internal address yet; the entry is still
	// returned so callers can report the node as unaddressable.
	IP string `json:"ip"`

	Ready  bool `json:"ready"`
	IsSelf bool `json:"isSelf"`
}

// Lister reads the directory. Every dependency is injected so the directory
// can be exercised without a cluster or a host.
type Lister struct {
	Client kubernetes.Interface

	// HostIPs resolves the addresses of the machine olaresd runs on, used to
	// tag the local node. Defaults to nets.LookupHostIps.
	HostIPs func() ([]string, error)

	// SelfNodeName resolves the Kubernetes name of this machine's node, used
	// when no address matches. Defaults to selfNodeName.
	SelfNodeName func() (string, error)
}

// List returns every node in the cluster.
func (l *Lister) List(ctx context.Context) ([]Node, error) {
	list, err := l.Client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	nodes := make([]Node, 0, len(list.Items))
	for i := range list.Items {
		n := &list.Items[i]
		nodes = append(nodes, Node{
			NodeName: n.Name,
			Role:     roleOf(n),
			IP:       internalIP(n),
			Ready:    utils.IsNodeReady(n),
		})
	}

	l.markSelf(nodes, list.Items)
	return nodes, nil
}

// markSelf tags the entry describing the machine olaresd runs on.
//
// Addresses come first because they are what the fan-out will dial. Names are
// the fallback, and they are the only thing left for the node an operator most
// wants to see: one whose kubelet is down has no usable address, and matching
// on addresses alone would drop it out of its own view of the cluster.
func (l *Lister) markSelf(nodes []Node, items []corev1.Node) {
	hostIPs := l.HostIPs
	if hostIPs == nil {
		hostIPs = nets.LookupHostIps
	}
	selfIPs := map[string]struct{}{}
	if ips, err := hostIPs(); err != nil {
		// Losing self-detection degrades the directory; it does not invalidate
		// it, so the nodes are still returned.
		klog.Warningf("inventory: lookup host ips error: %v", err)
	} else {
		for _, ip := range ips {
			selfIPs[ip] = struct{}{}
		}
	}

	for i := range nodes {
		if nodes[i].IP == "" {
			continue
		}
		if _, ok := selfIPs[nodes[i].IP]; ok {
			nodes[i].IsSelf = true
			return
		}
	}

	selfName := l.SelfNodeName
	if selfName == nil {
		selfName = selfNodeName
	}
	name, err := selfName()
	if err != nil {
		klog.Warningf("inventory: resolve this node's name error: %v", err)
		return
	}
	if name == "" {
		return
	}
	for i := range nodes {
		if strings.EqualFold(nodes[i].NodeName, name) ||
			strings.EqualFold(items[i].Labels[hostnameLabel], name) {
			nodes[i].IsSelf = true
			return
		}
	}
}

// selfNodeName reports the Kubernetes name of this machine's node. NODE_NAME
// wins when olaresd was given one, since a containerized daemon's hostname is
// the container's rather than the node's; otherwise kubelet's default, the
// machine hostname, is the name to look for.
func selfNodeName() (string, error) {
	if name := strings.TrimSpace(os.Getenv("NODE_NAME")); name != "" {
		return name, nil
	}
	return os.Hostname()
}

// List reads the directory using the daemon's shared kube client.
func List(ctx context.Context) ([]Node, error) {
	client, err := utils.GetKubeClient()
	if err != nil {
		return nil, err
	}
	return (&Lister{Client: client}).List(ctx)
}

func roleOf(n *corev1.Node) Role {
	if utils.IsMasterNode(n) {
		return RoleMaster
	}
	return RoleWorker
}

func internalIP(n *corev1.Node) string {
	for _, addr := range n.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			return addr.Address
		}
	}
	return ""
}
