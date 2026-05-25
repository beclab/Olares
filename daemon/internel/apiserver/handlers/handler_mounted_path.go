package handlers

import (
	"net/http"

	"github.com/beclab/Olares/daemon/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/shirou/gopsutil/disk"
)

func (h *Handlers) getMountedPath(ctx *fiber.Ctx, mutate func(*disk.UsageStat) *disk.UsageStat) error {
	res, err := utils.GetMountedPathDetail(ctx.Context(), mutate)
	if err != nil {
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}
	return h.OkJSON(ctx, "success", res)
}

func (h *Handlers) GetMountedPath(ctx *fiber.Ctx) error {
	return h.getMountedPath(ctx, nil)
}

func (h *Handlers) GetMountedPathInCluster(ctx *fiber.Ctx) error {
	return h.getMountedPath(ctx, func(us *disk.UsageStat) *disk.UsageStat {
		path := nodePathToClusterPath(us.Path)
		if path == us.Path {
			// not in cluster path
			return nil
		}

		us.Path = path

		return us
	})
}
