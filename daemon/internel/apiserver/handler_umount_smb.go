package apiserver

import (
	"net/http"

	"bytetrade.io/web3os/terminusd/pkg/commands"
	umountsmb "bytetrade.io/web3os/terminusd/pkg/commands/umount_smb"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

type UmountSmbReq struct {
	Path string ``
}

func (h *handlers) umountSmbInNode(ctx *fiber.Ctx, cmd commands.Interface, pathInNode string) error {
	_, err := cmd.Execute(ctx.Context(), &umountsmb.Param{
		MountPath: pathInNode,
	})

	if err != nil {
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	return h.OkJSON(ctx, "success to umount")
}

func (h *handlers) PostUmountSmb(ctx *fiber.Ctx, cmd commands.Interface) error {
	var req UmountSmbReq
	if err := h.ParseBody(ctx, &req); err != nil {
		klog.Error("parse request error, ", err)
		return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
	}
	if req.Path == "" {
		return h.ErrJSON(ctx, http.StatusBadRequest, "ip is empty")
	}

	return h.umountSmbInNode(ctx, cmd, req.Path)
}

func (h *handlers) PostUmountSmbInCluster(ctx *fiber.Ctx, cmd commands.Interface) error {
	var req UmountSmbReq
	if err := h.ParseBody(ctx, &req); err != nil {
		klog.Error("parse request error, ", err)
		return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
	}
	if req.Path == "" {
		return h.ErrJSON(ctx, http.StatusBadRequest, "ip is empty")
	}

	nodePath := clusterPathToNodePath(req.Path)

	return h.umountSmbInNode(ctx, cmd, nodePath)
}
