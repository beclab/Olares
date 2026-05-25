package handlers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func (h *Handlers) DisableOverlayGateway(ctx *fiber.Ctx) error {

	return h.ErrJSON(ctx, http.StatusOK, "success")
}
