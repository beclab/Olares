package apiserver

import (
	"net/http"

	"bytetrade.io/web3os/terminusd/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/shirou/gopsutil/disk"
	"k8s.io/klog/v2"
)

func (h *handlers) getMountedUsb(ctx *fiber.Ctx, mutate func(*disk.UsageStat) *disk.UsageStat) error {
	paths, err := utils.MountedUsbPath(ctx.Context())
	if err != nil {
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	var res []*disk.UsageStat
	for _, p := range paths {
		u, err := disk.UsageWithContext(ctx.Context(), p)
		if err != nil {
			klog.Error("get path usage error, ", err, ", ", p)
			return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
		}

		if mutate != nil {
			u = mutate(u)
		}

		res = append(res, u)
	}

	return h.OkJSON(ctx, "success", res)
}

func (h *handlers) GetMountedUsb(ctx *fiber.Ctx) error {
	return h.getMountedUsb(ctx, nil)
}

func (h *handlers) GetMountedUsbInCluster(ctx *fiber.Ctx) error {
	return h.getMountedUsb(ctx, func(us *disk.UsageStat) *disk.UsageStat {
		us.Path = nodePathToClusterPath(us.Path)
		return us
	})
}
