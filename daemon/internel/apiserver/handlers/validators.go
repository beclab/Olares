package handlers

import (
	"net/http"

	"github.com/beclab/Olares/daemon/internel/client"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"github.com/gofiber/fiber/v2"
)

func (h *Handlers) itsMe(ctx *fiber.Ctx, user string) error {
	u, ok := ctx.Context().UserValue(client.USER_CONTEXT).(*utils.ValidToken)
	if !ok || u == nil {
		return h.ErrJSON(ctx, http.StatusForbidden, "user data not found in context")
	}
	if u.Username != user {
		return h.ErrJSON(ctx, http.StatusForbidden, "operation is only allowed for the user themselves")
	}
	return nil
}
