package apiserver

import (
	"net/http"

	"bytetrade.io/web3os/terminusd/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/shirou/gopsutil/disk"
	"k8s.io/klog/v2"
)

func (h *handlers) getMountedHdd(ctx *fiber.Ctx, mutate func(*disk.UsageStat) *disk.UsageStat) error {
	paths, err := utils.MountedHddPath(ctx.Context())
	if err != nil {
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	klog.Info("mounted path, ", paths)

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

func (h *handlers) GetMountedHdd(ctx *fiber.Ctx) error {
	return h.getMountedHdd(ctx, nil)
}

func (h *handlers) GetMountedHddInCluster(ctx *fiber.Ctx) error {
	return h.getMountedHdd(ctx, func(us *disk.UsageStat) *disk.UsageStat {
		us.Path = nodePathToClusterPath(us.Path)
		return us
	})
}
