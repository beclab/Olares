package handlers

import (
	"errors"
	"net/http"

	"github.com/beclab/Olares/daemon/internel/client"
	"github.com/beclab/Olares/daemon/pkg/commands"
	collectlogs "github.com/beclab/Olares/daemon/pkg/commands/collect_logs"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

// PostCollectLogsNode is the node-local executor endpoint invoked by the master
// orchestrator. It runs on every node (no RequireMaster) and collects only this
// node's logs into the shared staging directory.
func (h *Handlers) PostCollectLogsNode(ctx *fiber.Ctx, cmd commands.Interface) error {
	var req collectlogs.NodeRequest
	if err := h.ParseBody(ctx, &req); err != nil {
		return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
	}

	c, ok := ctx.Context().UserValue(client.ClIENT_CONTEXT).(client.Client)
	if !ok {
		return h.ErrJSON(ctx, http.StatusForbidden, "client not found")
	}
	req.CallerOlaresID = c.OlaresID()

	res, err := cmd.Execute(ctx.Context(), &req)
	if err != nil {
		var denied *collectlogs.ScopeDeniedError
		switch {
		case errors.As(err, &denied):
			return h.ErrJSON(ctx, http.StatusForbidden, "requested scope exceeds caller permission", denied.Denied)
		case errors.Is(err, collectlogs.ErrNothingRequested):
			return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
		default:
			klog.Error("execute command error, ", err, ", ", cmd.OperationName().Stirng())
			return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
		}
	}

	return h.OkJSON(ctx, "success to collect node logs", res)
}
