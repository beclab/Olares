package clusterstatus

import (
	"context"
	"sync"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
	"k8s.io/klog/v2"
)

const (
	// defaultTimeout bounds one node's status read. This is a page load, not
	// an operation: a node that has not answered in this long is reported as
	// unreachable, and the page says so while the rest of the cluster is
	// still described.
	defaultTimeout = 5 * time.Second

	// defaultParallel bounds how many nodes are dialled at once, so loading
	// the overview of a large cluster does not open a socket per node.
	defaultParallel = 8
)

// Collector assembles the cluster summary. Every dependency is injected: the
// aggregation is exercised without a cluster, a network or a clock.
type Collector struct {
	// Nodes is the master's node directory, NotReady and unaddressable
	// entries included. A node missing from it is a node nobody would know
	// to worry about, so a failure here fails the whole summary.
	Nodes func(ctx context.Context) ([]inventory.Node, error)

	// Identify names the cluster. It reports what it could resolve rather
	// than failing: a missing identifier is one field, not a blank page.
	Identify func(ctx context.Context) Identity

	// Local reads the control node's own status without leaving the
	// machine. The cluster's summary must not depend on this node being
	// able to reach itself over the network.
	Local func(ctx context.Context) nodestatus.Status

	// Remote reads one worker's status over its node-local endpoint.
	Remote func(ctx context.Context, node inventory.Node, token string) (nodestatus.Status, error)

	// ActivePhase is the phase of a cluster operation in flight, if any.
	ActivePhase func() (nodestatus.Phase, bool)

	Now      func() time.Time
	Timeout  time.Duration
	Parallel int
}

// Collect reads every node and summarizes them. token is the caller's access
// token, forwarded to each worker for the length of this one aggregation.
//
// The only error it returns is an unreadable directory. Everything else is
// reported inside the summary, because a page that cannot say which node is
// unreachable is worse than one that says the cluster is degraded.
func (c *Collector) Collect(ctx context.Context, token string) (Summary, error) {
	nodes, err := c.Nodes(ctx)
	if err != nil {
		return Summary{}, err
	}

	views := c.read(ctx, nodes, token)

	var override nodestatus.Phase
	if c.ActivePhase != nil {
		if phase, ok := c.ActivePhase(); ok {
			override = phase
		}
	}

	return Build(c.Identify(ctx), views, override, c.now()), nil
}

// read asks every node at once. Adding the nodes up one timeout at a time is
// how a two-node cluster's overview page becomes a minute long, and the node
// that makes it slow is exactly the one the user opened the page to look at.
func (c *Collector) read(ctx context.Context, nodes []inventory.Node, token string) []NodeView {
	views := make([]NodeView, len(nodes))
	sem := make(chan struct{}, c.parallel())

	var wg sync.WaitGroup
	for i := range nodes {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			views[i] = c.readOne(ctx, nodes[i], token)
		}(i)
	}
	wg.Wait()
	return views
}

func (c *Collector) readOne(ctx context.Context, n inventory.Node, token string) NodeView {
	// The address comes from the directory, so it survives the node not
	// answering. Everything else about the machine comes from the machine.
	view := NodeView{Name: n.NodeName, Role: n.Role, Ready: n.Ready, IP: n.IP}

	switch {
	case n.IsSelf:
		return apply(view, c.Local(ctx))
	case n.IP == "":
		return unreachableView(view, ConditionNodeUnaddressable)
	}

	callCtx, cancel := context.WithTimeout(ctx, c.timeout())
	defer cancel()
	status, err := c.Remote(callCtx, n, token)
	if err != nil {
		// The address, the port and whatever the transport made of them stay
		// here: this summary is served to any signed-in user.
		klog.Warningf("clusterstatus: read status of node %s: %v", n.NodeName, err)
		return unreachableView(view, "")
	}
	return apply(view, status)
}

// apply takes what a node said about itself. Connectivity is not among it: the
// node answered, which is the only evidence of reachability there is, and a
// node's own report of being online says nothing about the path to it.
func apply(view NodeView, status nodestatus.Status) NodeView {
	view.Health = status.Health
	view.Phase = status.Phase
	view.Connectivity = nodestatus.ConnectivityOnline
	if view.Health == "" {
		view.Health = nodestatus.HealthUnknown
	}
	if view.Phase == "" {
		view.Phase = nodestatus.PhaseUnknown
	}
	// The summary line the detail page expands. It is taken only from an
	// answer: a view built for a node that did not reply keeps these empty.
	view.DeviceType = status.DeviceType
	view.OlaresdVersion = status.OlaresdVersion
	view.CPU = status.CPU
	view.Memory = status.Memory
	view.Disk = status.Disk
	return view
}

// unreachableView is a node nothing could get an answer out of. Its health is
// unknown rather than degraded, and its connectivity is unreachable rather
// than offline: not answering is not proof that the machine is off.
func unreachableView(view NodeView, reason string) NodeView {
	view.Health = nodestatus.HealthUnknown
	view.Connectivity = nodestatus.ConnectivityUnreachable
	view.Phase = nodestatus.PhaseUnknown
	view.Reason = reason
	return view
}

func (c *Collector) timeout() time.Duration {
	if c.Timeout <= 0 {
		return defaultTimeout
	}
	return c.Timeout
}

func (c *Collector) parallel() int {
	if c.Parallel <= 0 {
		return defaultParallel
	}
	return c.Parallel
}

func (c *Collector) now() time.Time {
	if c.Now == nil {
		return time.Now()
	}
	return c.Now()
}
