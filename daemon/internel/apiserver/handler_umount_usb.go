package apiserver

import (
	"net/http"

	"bytetrade.io/web3os/terminusd/pkg/commands"
	umountusb "bytetrade.io/web3os/terminusd/pkg/commands/umount_usb"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

type UmountReq struct {
	Path string ``
}

func (h *handlers) umountUsbInNode(ctx *fiber.Ctx, cmd commands.Interface, pathInNode string) error {
	_, err := cmd.Execute(ctx.Context(), &umountusb.Param{
		Path: pathInNode,
	})

	if err != nil {
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	return h.OkJSON(ctx, "success to umount")
}

func (h *handlers) PostUmountUsb(ctx *fiber.Ctx, cmd commands.Interface) error {
	var req UmountReq
	if err := h.ParseBody(ctx, &req); err != nil {
		klog.Error("parse request error, ", err)
		return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
	}
	if req.Path == "" {
		return h.ErrJSON(ctx, http.StatusBadRequest, "ip is empty")
	}

	return h.umountUsbInNode(ctx, cmd, req.Path)
}

func (h *handlers) PostUmountUsbInCluster(ctx *fiber.Ctx, cmd commands.Interface) error {
	var req UmountReq
	if err := h.ParseBody(ctx, &req); err != nil {
		klog.Error("parse request error, ", err)
		return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
	}
	if req.Path == "" {
		return h.ErrJSON(ctx, http.StatusBadRequest, "ip is empty")
	}

	nodePath := clusterPathToNodePath(req.Path)

	return h.umountUsbInNode(ctx, cmd, nodePath)
}
