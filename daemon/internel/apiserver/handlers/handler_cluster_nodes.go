package handlers

import (
	"net/http"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

// listClusterNodes is indirected so the handler can be exercised without a
// cluster.
var listClusterNodes = inventory.List

// nodeDirectoryUnavailable is what a caller is told when the directory cannot
// be read. The reason lives in the daemon log: the apiserver address, the
// certificate chain and the token are all in reach of the underlying error,
// and this route is served to any signed-in user.
const nodeDirectoryUnavailable = "cluster node directory is unavailable"

// GetClusterNodes serves the node directory. It is the master's answer and the
// only place a caller may learn a node's address from: a node the master
// cannot see is a node nobody should be told to talk to.
func (h *Handlers) GetClusterNodes(ctx *fiber.Ctx) error {
	nodes, err := listClusterNodes(ctx.Context())
	if err != nil {
		klog.Error("list cluster nodes error, ", err)
		return h.ErrJSON(ctx, http.StatusInternalServerError, nodeDirectoryUnavailable)
	}

	return h.OkJSON(ctx, "success", nodes)
}
