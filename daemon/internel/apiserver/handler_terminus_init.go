package apiserver

import (
	"net/http"

	"bytetrade.io/web3os/terminusd/pkg/commands"
	"bytetrade.io/web3os/terminusd/pkg/commands/install"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

type TerminusInitReq struct {
	Username string `json:"username" valid:"required"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Domain   string `json:"domain"`
}

func (h *handlers) PostTerminusInit(ctx *fiber.Ctx, cmd commands.Interface) error {
	var req TerminusInitReq
	if err := h.ParseBody(ctx, &req); err != nil {
		klog.Error("parse request error, ", err)
		return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
	}

	// run in background
	_, err := cmd.Execute(h.mainCtx, &install.Param{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		Domain:   req.Domain,
	})

	if err != nil {
		klog.Error("execute command error, ", err, ", ", cmd.OperationName().Stirng())
		return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
	}

	return h.OkJSON(ctx, "start to "+cmd.OperationName().Stirng())
}
