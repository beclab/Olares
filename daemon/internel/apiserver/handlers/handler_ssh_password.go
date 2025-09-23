package handlers

import (
	"net/http"

	"github.com/beclab/Olares/daemon/pkg/commands"
	sshpassword "github.com/beclab/Olares/daemon/pkg/commands/ssh_password"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

type SSHPasswordReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handlers) PostSSHPassword(ctx *fiber.Ctx, cmd commands.Interface) error {
	var req SSHPasswordReq
	if err := h.ParseBody(ctx, &req); err != nil {
		klog.Error("parse request error, ", err)

		return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
	}

	_, err := cmd.Execute(ctx.Context(), &sshpassword.Param{
		Username: req.Username,
		Password: req.Password,
	})

	if err != nil {
		klog.Error("execute command error, ", err, ", ", cmd.OperationName())

		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	return h.OkJSON(ctx, "success to set ssh password")
}
