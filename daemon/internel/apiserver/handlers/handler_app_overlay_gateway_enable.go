package handlers

import (
	"net/http"

	"github.com/beclab/Olares/daemon/pkg/commands"
	enableappoverlaygateway "github.com/beclab/Olares/daemon/pkg/commands/enable_app_overlay_gateway"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

type EnableAppOverlayGatewayReq struct {
	AppID string `json:"app_id"`
	User  string `json:"user"`
}

func (h *Handlers) EnableAppOverlayGateway(ctx *fiber.Ctx, cmd commands.Interface) error {
	var req EnableAppOverlayGatewayReq
	if err := h.ParseBody(ctx, &req); err != nil {
		klog.Error("parse request error, ", err)

		return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
	}

	s, err := h.getOverlayGatewayStatus(ctx.Context())
	if err != nil {
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	if s.Disable {
		return h.ErrJSON(ctx, http.StatusBadRequest, s.DisableReason)
	}

	if s.Status == OverlayGatewayOff {
		return h.ErrJSON(ctx, http.StatusBadRequest, "overlay gateway is disabled, please enable it first")
	}

	_, err = cmd.Execute(ctx.Context(), &enableappoverlaygateway.Param{
		AppID: req.AppID,
		User:  req.User,
	})
	if err != nil {
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	return h.ErrJSON(ctx, http.StatusOK, "success")
}
