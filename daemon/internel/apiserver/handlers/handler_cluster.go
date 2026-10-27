package handlers

import (
	"context"
	"net"
	"net/http"
	"strconv"

	"github.com/beclab/Olares/daemon/pkg/cluster/clusterstatus"
	"github.com/beclab/Olares/daemon/pkg/cluster/fanout"
	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

// clusterSummaryUnavailable is what a caller is told when the node directory
// cannot be read, which is the one fact the summary cannot be assembled
// without. The reason lives in the daemon log: the apiserver address, the
// certificate chain and the token are all in reach of the underlying error,
// and this route is served to any signed-in user.
const clusterSummaryUnavailable = "cluster status is unavailable"

// These are indirected so the summary can be exercised without a cluster and
// without a network.
var (
	clusterIDOf      = defaultClusterID
	remoteNodeStatus = fetchNodeStatus
)

// GetCluster serves the cluster-level summary: one health, one phase and one
// set of counts over every node the master knows about. It is the overview
// above the node directory, and it changes nothing.
func (h *Handlers) GetCluster(ctx *fiber.Ctx) error {
	collector := &clusterstatus.Collector{
		Nodes:       listClusterNodes,
		Identify:    clusterIdentity,
		Local:       localNodeStatus,
		Remote:      remoteNodeStatus,
		ActivePhase: activeClusterPhase,
	}

	// The caller's token is what each worker is asked with, and it is not
	// kept: it lives for this one aggregation, the same way the request does.
	summary, err := collector.Collect(ctx.Context(), ctx.Get(AUTH_HEADER))
	if err != nil {
		klog.Error("summarize cluster error, ", err)
		return h.ErrJSON(ctx, http.StatusInternalServerError, clusterSummaryUnavailable)
	}

	return h.OkJSON(ctx, "success", summary)
}

// clusterIdentity names the cluster. A field it cannot resolve is left empty
// and reported as a condition rather than failing the whole summary: a page
// that cannot show the identifier is still a page that shows the nodes.
func clusterIdentity(ctx context.Context) clusterstatus.Identity {
	snap, _ := currentStateSnapshot()

	id := clusterstatus.Identity{}
	if snap.TerminusName != nil {
		id.Name = *snap.TerminusName
	}
	if snap.TerminusVersion != nil {
		id.Version = *snap.TerminusVersion
	}

	clusterID, err := clusterIDOf(ctx)
	if err != nil {
		klog.Warning("resolve cluster id error, ", err)
	} else {
		id.ClusterID = clusterID
	}
	return id
}

// localNodeStatus is the control node's own view of itself, built from the
// same snapshot the node-local endpoint serves. It stays on this machine: the
// cluster's summary must not depend on the master being able to reach itself
// over the network.
func localNodeStatus(ctx context.Context) nodestatus.Status {
	// One snapshot for the whole answer: reading the live state per field
	// would let the refresh loop land in the middle and produce a report no
	// single observation ever supported.
	snap, observedAt := currentStateSnapshot()
	id := identityFrom(ctx, snap)

	return nodestatus.Build(
		id,
		snap,
		nodestatus.Detect(ctx, nodestatus.ProbeInput{Identity: id, State: snap}),
		observedAt,
	)
}

// activeClusterPhase is what a cluster operation in flight says the cluster is
// doing. The routes exist before the orchestrator is wired up, and reading the
// cluster is not an operation, so its absence leaves the phase alone.
func activeClusterPhase() (nodestatus.Phase, bool) {
	if clusterOperations == nil {
		return "", false
	}
	return clusterOperations.ActivePhase()
}

// fetchNodeStatus asks one worker for its own report. The address comes from
// the master's directory and never from a request.
func fetchNodeStatus(ctx context.Context, node inventory.Node, token string) (nodestatus.Status, error) {
	return nodestatus.Fetch(ctx, nodeStatusURL(node.IP), token)
}

func nodeStatusURL(ip string) string {
	return "http://" + net.JoinHostPort(ip, strconv.Itoa(fanout.OlaresdPort))
}

func defaultClusterID(ctx context.Context) (string, error) {
	client, err := utils.GetKubeClient()
	if err != nil {
		return "", err
	}
	return clusterstatus.ClusterID(ctx, client)
}
