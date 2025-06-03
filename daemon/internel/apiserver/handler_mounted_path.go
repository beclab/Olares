package apiserver

import (
	"net/http"

	"bytetrade.io/web3os/terminusd/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/shirou/gopsutil/disk"
	"k8s.io/klog/v2"
)

type mountedPath struct {
	disk.UsageStat `json:",inline"`
	Type           string `json:"type"`
	Invalid        bool   `json:"invalid"`
	IDSerial       string `json:"id_serial"`
	IDSerialShort  string `json:"id_serial_short"`
	PartitionUUID  string `json:"partition_uuid"`
	Device         string `json:"device"`
	ReadOnly       bool   `json:"read_only"`
}

func (h *handlers) getMountedPath(ctx *fiber.Ctx, mutate func(*disk.UsageStat) *disk.UsageStat) error {
	paths, err := utils.MountedPath(ctx.Context())
	if err != nil {
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	klog.Info("mounted path, ", paths)

	var res []*mountedPath
	for _, p := range paths {
		u, err := disk.UsageWithContext(ctx.Context(), p.Path)
		if err != nil {
			klog.Error("get path usage error, ", err, ", ", p)
			// return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())

			u = &disk.UsageStat{Path: p.Path}
			p.Invalid = true
		}

		if mutate != nil {
			u = mutate(u)
		}

		res = append(res, &mountedPath{
			*u,
			string(p.Type),
			p.Invalid,
			p.IDSerial,
			p.IDSerialShort,
			p.PartitionUUID,
			p.Device,
			p.ReadOnly,
		})
	}

	return h.OkJSON(ctx, "success", res)
}

func (h *handlers) GetMountedPath(ctx *fiber.Ctx) error {
	return h.getMountedPath(ctx, nil)
}

func (h *handlers) GetMountedPathInCluster(ctx *fiber.Ctx) error {
	return h.getMountedPath(ctx, func(us *disk.UsageStat) *disk.UsageStat {
		us.Path = nodePathToClusterPath(us.Path)
		return us
	})
}
