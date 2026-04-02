package handlers

import (
	"net/http"

	"github.com/beclab/Olares/daemon/pkg/commands"
	mountnfs "github.com/beclab/Olares/daemon/pkg/commands/mount_nfs"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

type MountNfsReq struct {
	Server     string `json:"server"`
	SharedPath string `json:"sharedPath"`
	MountPath  string `json:"mountPath"`
}

func (h *Handlers) PostMountNfsDriver(ctx *fiber.Ctx, cmd commands.Interface) error {
	var req MountNfsReq
	if err := h.ParseBody(ctx, &req); err != nil {
		klog.Error("parse request error, ", err)
		return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
	}

	_, err := cmd.Execute(ctx.Context(), &mountnfs.Param{
		MountBaseDir: commands.MOUNT_BASE_DIR,
		Server:       req.Server,
		NfsPath:      req.SharedPath,
		MountPath:    req.MountPath,
	})

	if err != nil {
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	return h.OkJSON(ctx, "success to mount")
}
