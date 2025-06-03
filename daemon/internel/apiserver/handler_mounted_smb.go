package apiserver

import (
	"net/http"

	"bytetrade.io/web3os/terminusd/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/shirou/gopsutil/disk"
	"k8s.io/klog/v2"
)

type mountedSmbPathResponse struct {
	disk.UsageStat `json:",inline"`
	Invalid        bool   `json:"invalid"`
	Device         string `json:"device"`
}

func (h *handlers) getMountedSmb(ctx *fiber.Ctx, mutate func(*disk.UsageStat) *disk.UsageStat) error {
	paths, err := utils.MountedSambaPath(ctx.Context())
	if err != nil {
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	klog.Info("mounted path, ", paths)

	var res []*mountedSmbPathResponse
	for _, p := range paths {
		u, err := disk.UsageWithContext(ctx.Context(), p.Path)
		if err != nil {
			klog.Error("get path usage error, ", err, ", ", p)
			return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
		}

		if mutate != nil {
			u = mutate(u)
		}

		res = append(res, &mountedSmbPathResponse{*u, p.Invalid, p.Device})
	}

	return h.OkJSON(ctx, "success", res)
}

func (h *handlers) GetMountedSmb(ctx *fiber.Ctx) error {
	return h.getMountedSmb(ctx, nil)
}

func (h *handlers) GetMountedSmbInCluster(ctx *fiber.Ctx) error {
	return h.getMountedSmb(ctx, func(us *disk.UsageStat) *disk.UsageStat {
		us.Path = nodePathToClusterPath(us.Path)
		return us
	})
}
