package handlers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func (h *Handlers) EnableOverlayGateway(ctx *fiber.Ctx) error {
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

	return h.ErrJSON(ctx, http.StatusOK, "success")
}
