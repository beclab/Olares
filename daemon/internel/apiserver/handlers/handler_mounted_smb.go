package handlers

import (
	"net/http"
	"slices"

	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/shirou/gopsutil/disk"
	"k8s.io/klog/v2"
)

type mountedSmbPathResponse struct {
	disk.UsageStat `json:",inline"`
	Invalid        bool   `json:"invalid"`
	Device         string `json:"device"`
}

// Deprecated: GetMountedSmb is deprecated, use GetMountedPathInCluster instead
func (h *Handlers) getMountedSmb(ctx *fiber.Ctx, mutate func(*disk.UsageStat) *disk.UsageStat) error {
	paths, err := utils.MountedSambaPath(ctx.Context())
	if err != nil {
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	klog.Info("mounted path, ", paths)

	var res []*mountedSmbPathResponse
	var mountedMountPoints []string
	for _, p := range paths {
		mountedMountPoints = append(mountedMountPoints, p.Path)
		u, err := disk.UsageWithContext(ctx.Context(), p.Path)
		if err != nil {
			klog.Error("get path usage error, ", err, ", ", p)
			u = &disk.UsageStat{Path: p.Path}
			p.Invalid = true
		}

		if mutate != nil {
			u = mutate(u)
		}

		res = append(res, &mountedSmbPathResponse{*u, p.Invalid, p.Device})
	}

	records, err := utils.LoadMountRecords(commands.MOUNT_RECORDS_FILE)
	if err != nil {
		klog.Warning("load mount records error, ", err)
	}
	for _, r := range records {
		if r.Type != utils.SMB {
			continue
		}
		if slices.Contains(mountedMountPoints, r.MountPoint) {
			continue
		}
		u := &disk.UsageStat{Path: r.MountPoint}
		if mutate != nil {
			u = mutate(u)
		}
		res = append(res, &mountedSmbPathResponse{*u, true, r.SmbPath})
	}

	return h.OkJSON(ctx, "success", res)
}

func (h *Handlers) GetMountedSmb(ctx *fiber.Ctx) error {
	return h.getMountedSmb(ctx, nil)
}

func (h *Handlers) GetMountedSmbInCluster(ctx *fiber.Ctx) error {
	return h.getMountedSmb(ctx, func(us *disk.UsageStat) *disk.UsageStat {
		us.Path = nodePathToClusterPath(us.Path)
		return us
	})
}
