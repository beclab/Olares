package handlers

import (
	"github.com/beclab/Olares/daemon/pkg/utils"
	"github.com/gofiber/fiber/v2"
)

func (h *Handlers) CheckDefaultSSHPwd(ctx *fiber.Ctx) error {
	isDefault := utils.IsDefaultSSHPassword()
	return h.OkJSON(ctx, "", fiber.Map{"owner_ssh_is_default": isDefault})
}
