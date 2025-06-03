package apiserver

import (
	"net/http"

	"bytetrade.io/web3os/terminusd/pkg/commands"
	mountsmb "bytetrade.io/web3os/terminusd/pkg/commands/mount_smb"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

type MountReq struct {
	SmbPath  string `json:"smbPath"`
	User     string `json:"user"`
	Password string `json:"password"`
}

func (h *handlers) PostMountSambaDriver(ctx *fiber.Ctx, cmd commands.Interface) error {
	var req MountReq
	if err := h.ParseBody(ctx, &req); err != nil {
		klog.Error("parse request error, ", err)
		return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
	}

	_, err := cmd.Execute(ctx.Context(), &mountsmb.Param{
		MountBaseDir: commands.MOUNT_BASE_DIR,
		SmbPath:      req.SmbPath,
		User:         req.User,
		Password:     req.Password,
	})

	if err != nil {
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	return h.OkJSON(ctx, "success to mount")
}
