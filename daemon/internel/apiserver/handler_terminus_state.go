package apiserver

import (
	"bytetrade.io/web3os/terminusd/pkg/cluster/state"
	"github.com/gofiber/fiber/v2"
)

func (h *handlers) GetTerminusState(ctx *fiber.Ctx) error {
	return h.OkJSON(ctx, "success", state.CurrentState)
}
