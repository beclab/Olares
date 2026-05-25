package handlers

import (
	"net/http"

	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/gofiber/fiber/v2"
)

func (h *Handlers) EnableOverlayGateway(ctx *fiber.Ctx, cmd commands.Interface) error {
	s, err := h.getOverlayGatewayStatus(ctx)
	if err != nil {
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	if s.Disable {
		return h.ErrJSON(ctx, http.StatusBadRequest, s.DisableReason)
	}

	if s.Status != OverlayGatewayOff {
		return h.ErrJSON(ctx, http.StatusBadRequest, "overlay gateway is already enabled")
	}

	_, err = cmd.Execute(ctx.Context(), nil)
	if err != nil {
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}
	return h.ErrJSON(ctx, http.StatusOK, "success")
}
