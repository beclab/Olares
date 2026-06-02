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

func (h *Handlers) PostCollectLogs(ctx *fiber.Ctx, cmd commands.Interface) error {
	var param collectlogs.Param
	if len(ctx.Body()) > 0 {
		if err := h.ParseBody(ctx, &param); err != nil {
			return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
		}
	}

	c, ok := ctx.Context().UserValue(client.ClIENT_CONTEXT).(client.Client)
	if !ok {
		return h.ErrJSON(ctx, http.StatusForbidden, "client not found")
	}
	// Inject the verified caller identity; never trust a body-supplied value.
	param.CallerOlaresID = c.OlaresID()
	// Forward the verified signature so the orchestrator can authenticate to
	// each node's node-local endpoint as the same caller.
	if sig, ok := ctx.GetReqHeaders()[SIGNATURE_HEADER]; ok && len(sig) > 0 {
		param.CallerSignature = sig[0]
	}

	_, err := cmd.Execute(ctx.Context(), &param)
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

	return h.OkJSON(ctx, "success to exec command")
}
