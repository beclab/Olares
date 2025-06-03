package apiserver

import (
	"net/http"

	"bytetrade.io/web3os/terminusd/pkg/commands"
	changehost "bytetrade.io/web3os/terminusd/pkg/commands/change_host"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

type ChangeHostReq struct {
	IP string `json:"ip"`
}

func (h *handlers) PostChangeHost(ctx *fiber.Ctx, cmd commands.Interface) error {
	var req ChangeHostReq
	if err := h.ParseBody(ctx, &req); err != nil {
		klog.Error("parse request error, ", err)
		return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
	}
	if req.IP == "" {
		return h.ErrJSON(ctx, http.StatusBadRequest, "ip is empty")
	}

	_, err := cmd.Execute(ctx.Context(), &changehost.Param{IP: req.IP})
	if err != nil {
		return h.ErrJSON(ctx, http.StatusServiceUnavailable, err.Error())
	}

	return h.OkJSON(ctx, "success to change host ip")
}
