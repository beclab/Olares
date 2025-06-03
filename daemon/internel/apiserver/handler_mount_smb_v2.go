package apiserver

import (
	"net/http"
	"slices"
	"strings"

	"bytetrade.io/web3os/terminusd/pkg/commands"
	mountsmb "bytetrade.io/web3os/terminusd/pkg/commands/mount_smb"
	"bytetrade.io/web3os/terminusd/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

type ListSmbResponse struct {
	Path    string `json:"path"`
	Mounted bool   `json:"mounted"`
}

func (h *handlers) PostMountSambaDriverV2(ctx *fiber.Ctx, cmd commands.Interface) error {
	var req MountReq
	if err := h.ParseBody(ctx, &req); err != nil {
		klog.Error("parse request error, ", err)
		return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
	}

	// parse smb path
	parsedSmbPath := strings.TrimPrefix(req.SmbPath, "//")
	if parsedSmbPath == "" {
		return h.ErrJSON(ctx, http.StatusBadRequest, "invalid smb path")
	}

	parsedToken := strings.Split(parsedSmbPath, "/")
	smbServer, smbPath := parsedToken[0], ""

	if len(parsedToken) > 1 {
		smbPath = strings.Join(parsedToken[1:], "/")
	}

	sharedName, err := utils.ListSambaSharenames(ctx.Context(), smbServer, req.User, req.Password)
	if err != nil {
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	if len(sharedName) == 0 {
		klog.Error("smb server has not shared points")
		return h.ErrJSON(ctx, http.StatusBadRequest, "smb server has not shared points")
	}

	// if request smb path only has the server address or smb path is not a valid shared name,
	// response the shared smb path list
	if smbPath == "" || !slices.ContainsFunc(sharedName, func(s string) bool {
		return s == strings.Split(smbPath, "/")[0] // allow mounting subpath
	}) {
		mountedPath, err := utils.MountedPath(ctx.Context())
		klog.Info(mountedPath)
		if err != nil {
			klog.Error("get mounted path error, ", err)
			return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
		}

		var sharedSmbPath []ListSmbResponse
		for _, name := range sharedName {
			sharedPath := strings.Join([]string{"/", smbServer, name}, "/")
			mounted := slices.ContainsFunc(mountedPath, utils.SubpathOfMountedPath(sharedPath))

			sharedSmbPath = append(sharedSmbPath, ListSmbResponse{
				Path:    sharedPath,
				Mounted: mounted,
			})
		}

		return h.NeedChoiceJSON(ctx, "Choose a valid share path", sharedSmbPath)
	}

	_, err = cmd.Execute(ctx.Context(), &mountsmb.Param{
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
