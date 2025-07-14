package handlers

import (
	"net/http"

	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

func (h *Handlers) PostShutdown(ctx *fiber.Ctx, cmd commands.Interface) error {
	_, err := cmd.Execute(ctx.Context(), nil)
	if err != nil {
		klog.Error("execute command error, ", err, ", ", cmd.OperationName().Stirng())

		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	return h.OkJSON(ctx, "success to exec command")
}
