package handlers

import (
	"net/http"

	"github.com/beclab/Olares/daemon/pkg/commands"
	umountnfs "github.com/beclab/Olares/daemon/pkg/commands/umount_nfs"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

type UmountNfsReq struct {
	Path string ``
}

func (h *Handlers) umountNfsInNode(ctx *fiber.Ctx, cmd commands.Interface, pathInNode string) error {
	_, err := cmd.Execute(ctx.Context(), &umountnfs.Param{
		MountPath: pathInNode,
	})

	if err != nil {
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	return h.OkJSON(ctx, "success to umount")
}

func (h *Handlers) PostUmountNfs(ctx *fiber.Ctx, cmd commands.Interface) error {
	var req UmountNfsReq
	if err := h.ParseBody(ctx, &req); err != nil {
		klog.Error("parse request error, ", err)
		return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
	}
	if req.Path == "" {
		return h.ErrJSON(ctx, http.StatusBadRequest, "ip is empty")
	}

	return h.umountNfsInNode(ctx, cmd, req.Path)
}

func (h *Handlers) PostUmountNfsInCluster(ctx *fiber.Ctx, cmd commands.Interface) error {
	var req UmountNfsReq
	if err := h.ParseBody(ctx, &req); err != nil {
		klog.Error("parse request error, ", err)
		return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
	}
	if req.Path == "" {
		return h.ErrJSON(ctx, http.StatusBadRequest, "ip is empty")
	}

	nodePath := clusterPathToNodePath(req.Path)

	return h.umountNfsInNode(ctx, cmd, nodePath)
}
