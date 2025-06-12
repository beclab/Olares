package apiserver

import (
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/gofiber/fiber/v2"
)

func (h *handlers) GetTerminusState(ctx *fiber.Ctx) error {
	return h.OkJSON(ctx, "success", state.CurrentState)
}
