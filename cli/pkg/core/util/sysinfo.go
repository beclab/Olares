package util

import (
	"context"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	_ "github.com/shirou/gopsutil/v4/net"
)

func GetHost() ([]string, error) {
	hostInfo, err := host.Info()
	if err != nil {
		return nil, err
	}

	var res = make([]string, 0, 9)
	res = append(res,
		hostInfo.Hostname, hostInfo.HostID, hostInfo.OS,
		hostInfo.Platform, hostInfo.PlatformFamily, hostInfo.PlatformVersion,
		hostInfo.KernelArch, hostInfo.VirtualizationRole, hostInfo.VirtualizationSystem,
		hostInfo.KernelVersion)

	return res, nil
}

func GetCpu() (string, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cpuInfo, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return "", 0, 0, err
	}
	if len(cpuInfo) == 0 {
		return "", 0, 0, fmt.Errorf("cpu info is empty")
	}

	cpuLogicalCount, _ := cpu.CountsWithContext(ctx, true)
	cpuPhysicalCount, _ := cpu.CountsWithContext(ctx, false)

	return cpuInfo[0].ModelName, cpuLogicalCount, cpuPhysicalCount, nil
}

func GetFs() (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// todo support other fs type
	var fsType = "overlayfs"
	var zfsPrefixName = ""

	ps, err := disk.PartitionsWithContext(ctx, true)
	if err != nil {
		return "", "", err
	}

	if ps == nil || len(ps) == 0 {
		return "", "", fmt.Errorf("partitions state is empty")
	}

	for _, p := range ps {
		if p.Mountpoint == "/var/lib" && p.Fstype == "zfs" {
			fsType = "zfs"
			zfsPrefixName = p.Device
			break
		}
	}

	return fsType, zfsPrefixName, nil
}

func GetPs() ([]disk.PartitionStat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ps, err := disk.PartitionsWithContext(ctx, true)
	if err != nil {
		return nil, err
	}

	return ps, nil
}

func GetDisk() (uint64, uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	usageInfo, err := disk.UsageWithContext(ctx, "/")
	if err != nil {
		return 0, 0, err
	}

	return usageInfo.Total, usageInfo.Free, nil
}

func GetMem() (uint64, uint64, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return 0, 0, err
	}

	return memInfo.Total, memInfo.Free, nil
}

func GetNet() {
	// ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// defer cancel()

	// ifInfo, _ := net.InterfacesWithContext(ctx)
	// fmt.Printf("ifinfo %s\n", ifInfo.String())

	// iocsInfo, _ := net.IOCountersWithContext(ctx, true)
	// for _, iocInfo := range iocsInfo {
	// 	fmt.Printf("iocInfo: %v\n", iocInfo)
	// }
}
