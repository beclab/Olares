package connector

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/util"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

type UbuntuVersion string
type DebianVersion string

func (u UbuntuVersion) String() string {
	switch u {
	case Ubuntu20:
		return "20."
	case Ubuntu22:
		return "22."
	case Ubuntu24:
		return "24."
	}
	return ""
}

func (d DebianVersion) String() string {
	switch d {
	case Debian11:
		return "11"
	case Debian12:
		return "12"
	}
	return ""
}

const (
	Ubuntu20 UbuntuVersion = "20."
	Ubuntu22 UbuntuVersion = "22."
	Ubuntu24 UbuntuVersion = "24."

	Debian9  DebianVersion = "9"
	Debian10 DebianVersion = "10"
	Debian11 DebianVersion = "11"
	Debian12 DebianVersion = "12"
)

type Systems interface {
	IsSupport() error

	IsWindows() bool
	IsDarwin() bool
	IsWsl() bool
	IsPve() bool
	IsPveLxc() bool
	IsPveOrPveLxc() bool
	IsRaspbian() bool
	IsLinux() bool

	IsUbuntu() bool
	IsDebian() bool
	GetDebianVersionCode() string

	IsUbuntuVersionEqual(ver UbuntuVersion) bool
	IsDebianVersionEqual(ver DebianVersion) bool
	IsOsArchInvalid() bool

	SetHostname(v string)
	GetHostname() string
	GetOsType() string
	GetOsArch() string
	GetUsername() string
	GetHomeDir() string
	GetOsVersion() string
	GetPkgManager() string
	SetNATGateway(ip string)
	GetNATGateway() string

	GetOsPlatformFamily() string

	GetLocalIp() string

	CgroupCpuEnabled() bool
	CgroupMemoryEnabled() bool
	GetFsType() string
	GetDefaultZfsPrefixName() string
	GetTotalMemory() uint64

	Print()
	String() string
}

type SystemInfo struct {
	HostInfo   *HostInfo       `json:"host"`
	CpuInfo    *CpuInfo        `json:"cpu"`
	DiskInfo   *DiskInfo       `json:"disk"`
	MemoryInfo *MemoryInfo     `json:"memory"`
	FsInfo     *FileSystemInfo `json:"filesystem"`
	CgroupInfo *CgroupInfo     `json:"cgroup,omitempty"`
	LocalIp    string          `json:"local_ip"`
	NatGateway string          `json:"nat_gateway"`
	PkgManager string          `json:"pkg_manager"`
}

func (s *SystemInfo) IsSupport() error {
	if !s.IsLinux() && !s.IsDarwin() && !s.IsWindows() {
		return fmt.Errorf("unsupported os type '%s'", s.GetOsType())
	}

	if s.IsOsArchInvalid() {
		return fmt.Errorf("unsupported arch '%s'", s.GetOsArch())
	}

	//if !s.IsUbuntu() && !s.IsDebian() {
	//	return fmt.Errorf("unsupported os type '%s', exit ...", s.GetOsPlatformFamily())
	//}

	if s.IsUbuntu() {
		if !s.IsUbuntuVersionEqual(Ubuntu20) && !s.IsUbuntuVersionEqual(Ubuntu22) && !s.IsUbuntuVersionEqual(Ubuntu24) {
			return fmt.Errorf("unsupported ubuntu os version '%s'", s.GetOsVersion())
		}
	}

	if s.IsDebian() {
		if !s.IsDebianVersionEqual(Debian11) && !s.IsDebianVersionEqual(Debian12) {
			return fmt.Errorf("unsupported debian os version '%s'", s.GetOsVersion())
		}
	}

	return nil
}

func (s *SystemInfo) GetLocalIp() string {
	return s.LocalIp
}

func (s *SystemInfo) SetNATGateway(ip string) {
	s.NatGateway = ip
}
func (s *SystemInfo) GetNATGateway() string {
	return s.NatGateway
}

func (s *SystemInfo) GetHostname() string {
	return s.HostInfo.HostName
}

func (s *SystemInfo) SetHostname(v string) {
	s.HostInfo.HostName = v
}

func (s *SystemInfo) GetOsType() string {
	return s.HostInfo.OsType
}

func (s *SystemInfo) GetOsArch() string {
	return s.HostInfo.OsArch
}

func (s *SystemInfo) GetUsername() string {
	return s.HostInfo.CurrentUser
}

func (s *SystemInfo) GetHomeDir() string {
	return s.HostInfo.HomeDir
}

func (s *SystemInfo) IsOsArchInvalid() bool {
	return strings.EqualFold(s.HostInfo.OsArch, "")
}

func (s *SystemInfo) GetOsVersion() string {
	return s.HostInfo.OsVersion
}

func (s *SystemInfo) GetOsPlatformFamily() string {
	return s.HostInfo.OsPlatformFamily
}

func (s *SystemInfo) String() string {
	str, _ := json.Marshal(s)
	return string(str)
}

func (s *SystemInfo) IsWindows() bool {
	return s.HostInfo.OsType == common.Windows
}

func (s *SystemInfo) IsDarwin() bool {
	return s.HostInfo.OsType == common.Darwin
}

func (s *SystemInfo) IsPve() bool {
	return s.HostInfo.OsPlatform == common.PVE
}

func (s *SystemInfo) IsPveLxc() bool {
	return s.HostInfo.OsPlatform == common.PVE_LXC
}

func (s *SystemInfo) IsPveOrPveLxc() bool {
	return s.IsPve() || s.IsPveLxc()
}

func (s *SystemInfo) IsWsl() bool {
	return s.HostInfo.OsPlatform == common.WSL
}

func (s *SystemInfo) IsRaspbian() bool {
	return s.HostInfo.OsPlatform == common.Raspbian
}

func (s *SystemInfo) IsLinux() bool {
	return s.HostInfo.OsType == common.Linux
}

func (s *SystemInfo) IsUbuntu() bool {
	return s.HostInfo.OsPlatformFamily == common.Ubuntu
}

func (s *SystemInfo) IsDebian() bool {
	return s.HostInfo.OsPlatformFamily == common.Debian
}

func (s *SystemInfo) IsUbuntuVersionEqual(ver UbuntuVersion) bool {
	return strings.Contains(s.HostInfo.OsVersion, ver.String())
}

func (s *SystemInfo) IsDebianVersionEqual(ver DebianVersion) bool {
	return strings.Contains(s.HostInfo.OsVersion, ver.String())
}

func (s *SystemInfo) GetDebianVersionCode() string {
	if !s.IsDebian() {
		return ""
	}

	if strings.Contains(s.HostInfo.OsVersion, string(Debian12)) {
		return "bookworm"
	} else if strings.Contains(s.HostInfo.OsVersion, string(Debian11)) {
		return "bullseye"
	} else if strings.Contains(s.HostInfo.OsVersion, string(Debian10)) {
		return "buster"
	} else if strings.Contains(s.HostInfo.OsVersion, string(Debian9)) {
		return "stretch"
	} else {
		return "jessie"
	}
}

func (s *SystemInfo) CgroupCpuEnabled() bool {
	return s.CgroupInfo.CpuEnabled >= 1
}

func (s *SystemInfo) CgroupMemoryEnabled() bool {
	return s.CgroupInfo.MemoryEnabled >= 1
}

func (s *SystemInfo) GetFsType() string {
	return s.FsInfo.Type
}

func (s *SystemInfo) GetDefaultZfsPrefixName() string {
	return s.FsInfo.DefaultZfsPrefixName
}

func (s *SystemInfo) GetTotalMemory() uint64 {
	return s.MemoryInfo.Total
}

func (s *SystemInfo) GetPkgManager() string {
	return s.PkgManager
}

func (s *SystemInfo) Print() {
	fmt.Printf("os info, all: %s\n", s.HostInfo.OsInfo)
	fmt.Printf("host info, user: %s, hostname: %s, hostid: %s, os: %s, platform: %s, platformfamily: %s, version: %s, arch: %s, localip: %s\n",
		s.HostInfo.CurrentUser, s.HostInfo.HostName, s.HostInfo.HostId,
		s.HostInfo.OsType, s.HostInfo.OsPlatform, s.HostInfo.OsPlatformFamily, s.HostInfo.OsVersion, s.HostInfo.OsArch, s.LocalIp)
	fmt.Printf("kernel info, version: %s\n", s.HostInfo.OsKernel)
	fmt.Printf("virtual info, role: %s, system: %s\n", s.HostInfo.VirtualizationRole, s.HostInfo.VirtualizationSystem)

	fmt.Printf("cpu info, model: %s, logical count: %d, physical count: %d\n",
		s.CpuInfo.CpuModel, s.CpuInfo.CpuLogicalCount, s.CpuInfo.CpuPhysicalCount)

	fmt.Printf("disk info, total: %s, free: %s\n", util.FormatBytes(int64(s.DiskInfo.Total)), util.FormatBytes(int64(s.DiskInfo.Free)))
	fmt.Printf("fs info, fs: %s, zfsmount: %s\n", s.FsInfo.Type, s.FsInfo.DefaultZfsPrefixName)
	fmt.Printf("mem info, total: %s, free: %s\n", util.FormatBytes(int64(s.MemoryInfo.Total)), util.FormatBytes(int64(s.MemoryInfo.Free)))
	fmt.Printf("cgroup info, cpu: %d, mem: %d\n", s.CgroupInfo.CpuEnabled, s.CgroupInfo.MemoryEnabled)
}

func GetSystemInfo() *SystemInfo {
	var si = new(SystemInfo)
	si.HostInfo = getHost()
	si.CpuInfo = getCpu()
	si.DiskInfo = getDisk()
	si.MemoryInfo = getMem()
	si.FsInfo = getFs()

	localIP, err := util.GetLocalIP()
	if err != nil {
		panic(errors.Wrap(err, "failed to get local ip"))
	}
	si.LocalIp = localIP.String()

	if si.IsLinux() {
		si.CgroupInfo = getCGroups()
	}

	switch si.GetOsPlatformFamily() {
	case common.Ubuntu, common.Debian:
		si.PkgManager = "apt-get"
	case common.Fedora:
		si.PkgManager = "dnf"
	case common.CentOs, common.RHEL:
		si.PkgManager = "yum"
	default:
		si.PkgManager = "apt-get"
	}

	return si
}

type HostInfo struct {
	HostName             string `json:"hostname"`
	HostId               string `json:"hostid"`
	OsType               string `json:"os_type"`
	OsPlatform           string `json:"os_platform"`
	OsPlatformFamily     string `json:"os_platform_family"`
	OsVersion            string `json:"os_version"`
	OsArch               string `json:"os_arch"`
	VirtualizationRole   string `json:"virtualization_role"`
	VirtualizationSystem string `json:"virtualization_system"`
	OsKernel             string `json:"os_kernel"`
	OsInfo               string `json:"os_info"`
	CurrentUser          string `json:"current_user"`
	HomeDir              string `json:"home_dir"`
}

func getHost() *HostInfo {
	hostInfo, err := host.Info()
	if err != nil {
		panic(err)
	}

	cmd := exec.Command("sh", "-c", "echo $(uname -a) |tr -d '\\n'")
	output, _ := cmd.Output()
	// if err != nil {
	// 	panic(err)
	// }

	var _osType = hostInfo.OS
	var _osPlatform = hostInfo.Platform
	var _osPlatformFamily = hostInfo.PlatformFamily
	var _osKernel = hostInfo.KernelVersion

	u, _ := user.Current()

	return &HostInfo{
		HostName:             hostInfo.Hostname,
		HostId:               hostInfo.HostID,
		OsType:               _osType,                                           // darwin linux windows
		OsPlatform:           formatOsPlatform(_osType, _osPlatform, _osKernel), // darwin linux wsl raspbian pve
		OsPlatformFamily:     formatOsPlatformFamily(_osPlatform, _osPlatformFamily),
		OsVersion:            hostInfo.PlatformVersion,
		OsArch:               ArchAlias(hostInfo.KernelArch),
		VirtualizationRole:   hostInfo.VirtualizationRole,
		VirtualizationSystem: hostInfo.VirtualizationSystem,
		OsKernel:             hostInfo.KernelVersion,
		OsInfo:               string(output),
		CurrentUser:          u.Username,
		HomeDir:              u.HomeDir,
	}
}

func formatOsPlatform(osType, osPlatform, osKernel string) string {
	if osType == common.Darwin {
		return common.Darwin
	}

	if osPlatform == common.Raspbian {
		return common.Raspbian
	}

	if strings.Contains(osKernel, "pve") {
		if lxc := os.Getenv("container"); lxc == "lxc" {
			return common.PVE_LXC
		}
		return common.PVE
	}

	if strings.Contains(osKernel, "-WSL") {
		return common.WSL
	}

	return common.Linux
}

func formatOsPlatformFamily(osPlatform, osPlatformFamily string) string {
	if osPlatform == common.Darwin {
		return common.Darwin
	}

	if osPlatform == common.Raspbian {
		return osPlatformFamily
	}

	return osPlatform
}

type CpuInfo struct {
	CpuModel         string `json:"cpu_model"`
	CpuLogicalCount  int    `json:"cpu_logical_count"`
	CpuPhysicalCount int    `json:"cpu_physical_count"`
}

func getCpu() *CpuInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cpuInfo, _ := cpu.InfoWithContext(ctx)
	cpuLogicalCount, _ := cpu.CountsWithContext(ctx, true)
	cpuPhysicalCount, _ := cpu.CountsWithContext(ctx, false)

	var cpuModel = ""
	if cpuInfo != nil && len(cpuInfo) > 0 {
		cpuModel = cpuInfo[0].ModelName
	}

	return &CpuInfo{
		CpuModel:         cpuModel,
		CpuLogicalCount:  cpuLogicalCount,
		CpuPhysicalCount: cpuPhysicalCount,
	}
}

type DiskInfo struct {
	Total uint64 `json:"total"`
	Free  uint64 `json:"free"`
}

func getDisk() *DiskInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	usageInfo, _ := disk.UsageWithContext(ctx, "/")

	var total uint64
	var free uint64
	if usageInfo != nil {
		total = usageInfo.Total
		free = usageInfo.Free
	}

	return &DiskInfo{
		Total: total,
		Free:  free,
	}
}

type MemoryInfo struct {
	Total uint64 `json:"total"`
	Free  uint64 `json:"free"`
}

func getMem() *MemoryInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	memInfo, _ := mem.VirtualMemoryWithContext(ctx)

	var total uint64
	var free uint64
	if memInfo != nil {
		total = memInfo.Total
		free = memInfo.Free
	}

	return &MemoryInfo{
		Total: total,
		Free:  free,
	}
}

type FileSystemInfo struct {
	Type                 string `json:"type"`
	DefaultZfsPrefixName string `json:"default_zfs_prefix_name"`
}

func getFs() *FileSystemInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var fsType = "overlayfs"
	var zfsPrefixName = ""

	ps, _ := disk.PartitionsWithContext(ctx, true)
	if ps != nil && len(ps) > 0 {
		for _, p := range ps {
			if p.Mountpoint == "/var/lib" && p.Fstype == "zfs" {
				fsType = "zfs"
				zfsPrefixName = p.Device
				break
			}
		}
	}

	return &FileSystemInfo{
		Type:                 fsType,
		DefaultZfsPrefixName: zfsPrefixName,
	}
}

type CgroupInfo struct {
	CpuEnabled    int `json:"cpu_enabled"`
	MemoryEnabled int `json:"memory_enabled"`
}

func getCGroups() *CgroupInfo {

	file, err := os.Open("/proc/cgroups")
	if err != nil {
		return nil
	}
	defer file.Close()

	var cpuEnabled int64
	var memoryEnabled int64

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		switch fields[0] {
		case "cpu":
			cpuEnabled, _ = strconv.ParseInt(fields[3], 10, 64)
		case "memory":
			memoryEnabled, _ = strconv.ParseInt(fields[3], 10, 64)
		default:
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return &CgroupInfo{
		CpuEnabled:    int(cpuEnabled),
		MemoryEnabled: int(memoryEnabled),
	}
}

func ArchAlias(arch string) string {
	switch arch {
	case "aarch64", "armv7l", "arm64", "arm":
		return "arm64"
	case "x86_64", "amd64":
		fallthrough
	case "ppc64le":
		fallthrough
	case "s390x":
		return "amd64"
	default:
		return ""
	}
}
