// Package clusterstatus builds the cluster-level view the master olaresd
// serves: one health, one phase and one set of counts over every node in the
// directory.
//
// It answers in the same vocabulary a single node answers in
// (pkg/cluster/nodestatus), because the difference between "this node is
// degraded" and "the cluster is degraded" should be the scope of the sentence
// and not a second set of words to map between.
//
// Three rules shape the aggregation. A fact nothing confirmed is unknown
// rather than assumed: a cluster is healthy only when every node said so. A
// node that did not answer is unreachable and never offline — nothing here
// watched a machine go down. And a failing node keeps its own condition, so
// the page can name it instead of saying that something, somewhere, is wrong.
package clusterstatus

import (
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
)

// Condition types. They are part of the contract with user-service and
// TermiPass: the message may change, these may not.
const (
	// ConditionNodeUnreachable marks a node whose olaresd did not answer.
	// What went wrong on the way there is a fact about the cluster's inside
	// — an address, a port, a certificate — and stays in the daemon log.
	ConditionNodeUnreachable = "NodeUnreachable"

	// ConditionNodeUnaddressable marks a directory entry with no internal
	// address. It is distinct from unreachable: nothing was ever dialled.
	ConditionNodeUnaddressable = "NodeUnaddressable"

	// ConditionNodeNotReady marks a node Kubernetes itself considers down.
	ConditionNodeNotReady = "NodeNotReady"

	// ConditionNodeDegraded marks a node that answered and said it is in
	// trouble.
	ConditionNodeDegraded = "NodeDegraded"

	// ConditionNodeStatusUnknown marks a node that answered without being
	// able to say how it is doing.
	ConditionNodeStatusUnknown = "NodeStatusUnknown"

	// ConditionClusterInventoryEmpty marks a directory with nothing in it,
	// which is a cluster this daemon could not describe rather than one
	// with no nodes.
	ConditionClusterInventoryEmpty = "ClusterInventoryEmpty"

	// ConditionClusterIdentityUnavailable marks a cluster whose identifier
	// could not be read. The rest of the summary is still served.
	ConditionClusterIdentityUnavailable = "ClusterIdentityUnavailable"
)

// conditionMessages are fixed sentences. A message assembled from an error
// would put whatever the transport said in front of a signed-in user, and this
// route is served to every one of them.
var conditionMessages = map[string]string{
	ConditionNodeUnreachable:            "the node's olaresd did not answer",
	ConditionNodeUnaddressable:          "the node has no internal address to reach it on",
	ConditionNodeNotReady:               "Kubernetes reports this node as NotReady",
	ConditionNodeDegraded:               "the node reports itself as degraded",
	ConditionNodeStatusUnknown:          "the node could not determine its own health",
	ConditionClusterInventoryEmpty:      "the node directory is empty",
	ConditionClusterIdentityUnavailable: "the cluster identifier could not be read",
}

// Condition is one structured fact behind the summarized health. Node names
// the node it is about, and is empty for a fact about the cluster itself.
type Condition struct {
	Type    string `json:"type"`
	Status  bool   `json:"status"`
	Node    string `json:"node,omitempty"`
	Message string `json:"message,omitempty"`
}

// NodeCounts is the node summary shown on the overview card.
//
// Healthy, Degraded and Unknown are health buckets and add up to Total.
// Unreachable counts a different axis — connectivity — so a node that did not
// answer is counted there and in Unknown: not answering is why its health
// could not be established.
type NodeCounts struct {
	Total       int `json:"total"`
	Healthy     int `json:"healthy"`
	Degraded    int `json:"degraded"`
	Unknown     int `json:"unknown"`
	Unreachable int `json:"unreachable"`
}

// Identity is what the cluster is called. ClusterID is empty when it could not
// be read, which the summary reports as a condition rather than papering over.
type Identity struct {
	ClusterID string
	Name      string
	Version   string
}

// Summary is the cluster-level response body.
type Summary struct {
	ClusterID string `json:"clusterId"`
	Name      string `json:"name"`

	Health nodestatus.Health `json:"health"`

	// Connectivity describes the link the caller just used: user-service to
	// this master. In an answered request it is online by construction —
	// the other values are for the caller that failed to get one.
	Connectivity nodestatus.Connectivity `json:"connectivity"`

	Phase         nodestatus.Phase `json:"phase"`
	OlaresVersion string           `json:"olaresVersion"`

	Nodes NodeCounts `json:"nodes"`

	// NodeList names the nodes behind those counts. "two nodes are degraded"
	// is never the end of the question, and the page's next move is always to
	// say which two.
	NodeList []NodeEntry `json:"nodeList"`

	// UpdatedAt is when this summary was assembled. It describes the
	// aggregation and not the state underneath it: how fresh one node's
	// answer is stays on that node's own status, under observedAt.
	UpdatedAt time.Time `json:"updatedAt"`

	Conditions []Condition `json:"conditions"`
}

// NodeView is one node's contribution to the summary: what the directory says
// about it, and what it said about itself.
type NodeView struct {
	Name         string
	Role         inventory.Role
	Ready        bool
	Health       nodestatus.Health
	Connectivity nodestatus.Connectivity
	Phase        nodestatus.Phase

	// IP is the internal address out of the master's directory. It is known
	// whether or not the node answered, which is why it is the one field an
	// unreachable node still contributes.
	IP string

	// What the node said about itself. All of it is empty for a node that did
	// not answer: the alternative is a page describing the hardware of a
	// machine nothing can currently reach.
	DeviceType     string
	OlaresdVersion string
	CPU            string
	Memory         string
	Disk           string

	// Reason names the condition when the caller knows something more
	// specific than the fields above imply, e.g. a node that was never
	// dialled because the directory has no address for it. Empty lets the
	// condition be derived.
	Reason string
}

// NodeEntry is one node as the overview serves it: enough to name the node, say
// how it is doing and show the summary line the detail page expands.
//
// Memory and Disk are the strings the node formatted, not byte counts — see
// nodestatus.Status. There is no node id: nothing in this cluster issues one,
// and a name that looked like an id would be relied upon as stable.
type NodeEntry struct {
	Name  string         `json:"name"`
	Role  inventory.Role `json:"role"`
	Ready bool           `json:"ready"`

	Health       nodestatus.Health       `json:"health"`
	Connectivity nodestatus.Connectivity `json:"connectivity"`
	Phase        nodestatus.Phase        `json:"phase"`

	IP string `json:"ip"`

	DeviceType     string `json:"deviceType"`
	OlaresdVersion string `json:"olaresdVersion"`
	CPU            string `json:"cpu"`
	Memory         string `json:"memory"`
	Disk           string `json:"disk"`
}

func entryFor(n NodeView) NodeEntry {
	return NodeEntry{
		Name:           n.Name,
		Role:           n.Role,
		Ready:          n.Ready,
		Health:         n.Health,
		Connectivity:   n.Connectivity,
		Phase:          n.Phase,
		IP:             n.IP,
		DeviceType:     n.DeviceType,
		OlaresdVersion: n.OlaresdVersion,
		CPU:            n.CPU,
		Memory:         n.Memory,
		Disk:           n.Disk,
	}
}

// Build assembles the cluster summary.
//
// phaseOverride is the phase of a cluster operation in flight; empty leaves
// the phase to the control node. It deliberately cannot touch health: a
// cluster rebooting because somebody asked it to is not a cluster in trouble.
func Build(id Identity, nodes []NodeView, phaseOverride nodestatus.Phase, at time.Time) Summary {
	s := Summary{
		ClusterID:     id.ClusterID,
		Name:          id.Name,
		OlaresVersion: id.Version,
		Connectivity:  nodestatus.ConnectivityOnline,
		Phase:         nodestatus.PhaseUnknown,
		UpdatedAt:     at,
		Conditions:    []Condition{},
		Nodes:         NodeCounts{Total: len(nodes)},
		NodeList:      make([]NodeEntry, 0, len(nodes)),
	}
	for _, n := range nodes {
		s.NodeList = append(s.NodeList, entryFor(n))
	}

	if id.ClusterID == "" {
		s.Conditions = append(s.Conditions, condition(ConditionClusterIdentityUnavailable, ""))
	}
	if len(nodes) == 0 {
		s.Health = nodestatus.HealthUnknown
		s.Conditions = append(s.Conditions, condition(ConditionClusterInventoryEmpty, ""))
		return s
	}

	trouble, uncertain := false, false
	for _, n := range nodes {
		switch n.Health {
		case nodestatus.HealthHealthy:
			s.Nodes.Healthy++
		case nodestatus.HealthDegraded:
			s.Nodes.Degraded++
		default:
			s.Nodes.Unknown++
		}
		if n.Connectivity == nodestatus.ConnectivityUnreachable {
			s.Nodes.Unreachable++
		}
		if n.Role == inventory.RoleMaster {
			s.Phase = n.Phase
		}

		reason := reasonFor(n)
		if reason == "" {
			continue
		}
		s.Conditions = append(s.Conditions, condition(reason, n.Name))
		if reason == ConditionNodeStatusUnknown {
			uncertain = true
		} else {
			trouble = true
		}
	}

	switch {
	case trouble:
		s.Health = nodestatus.HealthDegraded
	case uncertain:
		s.Health = nodestatus.HealthUnknown
	default:
		s.Health = nodestatus.HealthHealthy
	}

	if phaseOverride != "" {
		s.Phase = phaseOverride
	}
	return s
}

// reasonFor names the one condition a node contributes, or "" when the node is
// fine. One per node keeps the list as long as the number of things wrong,
// rather than as long as the number of ways to describe them.
func reasonFor(n NodeView) string {
	if n.Reason != "" {
		return n.Reason
	}
	switch {
	case n.Connectivity == nodestatus.ConnectivityUnreachable:
		return ConditionNodeUnreachable
	case !n.Ready:
		return ConditionNodeNotReady
	case n.Health == nodestatus.HealthDegraded:
		return ConditionNodeDegraded
	case n.Health != nodestatus.HealthHealthy:
		return ConditionNodeStatusUnknown
	default:
		return ""
	}
}

func condition(ty, node string) Condition {
	return Condition{Type: ty, Status: true, Node: node, Message: conditionMessages[ty]}
}
