package state

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/beclab/Olares/daemon/internel/watcher"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/nets"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"

	cpu "github.com/klauspost/cpuid/v2"
	"github.com/pbnjay/memory"
)

type state struct {
	TerminusdState      TerminusDState `json:"terminusdState"`
	TerminusState       TerminusState  `json:"terminusState"`
	TerminusName        *string        `json:"terminusName,omitempty"`
	TerminusVersion     *string        `json:"terminusVersion,omitempty"`
	InstalledTime       *int64         `json:"installedTime,omitempty"`
	InitializedTime     *int64         `json:"initializedTime,omitempty"`
	OlaresdVersion      *string        `json:"olaresdVersion,omitempty"`
	InstallFinishedTime *time.Time     `json:"-"`

	// sys info
	DeviceName *string `json:"device_name,omitempty"`
	HostName   *string `json:"host_name,omitempty"`
	OsType     string  `json:"os_type"`
	OsArch     string  `json:"os_arch"`
	OsInfo     string  `json:"os_info"`
	OsVersion  string  `json:"os_version"`
	CpuInfo    string  `json:"cpu_info"`
	GpuInfo    *string `json:"gpu_info,omitempty"`
	Memory     string  `json:"memory"`
	Disk       string  `json:"disk"`

	// network info
	WikiConnected  bool    `json:"wifiConnected"`
	WifiSSID       *string `json:"wifiSSID,omitempty"`
	WiredConnected bool    `json:"wiredConnected"`
	HostIP         string  `json:"hostIp"`
	ExternalIP     string  `json:"externalIp"`

	// installing / uninstalling / upgrading state
	InstallingState         ProcessingState `json:"installingState"`
	InstallingProgress      string          `json:"installingProgress"`
	InstallingProgressNum   int             `json:"-"`
	UninstallingState       ProcessingState `json:"uninstallingState"`
	UninstallingProgress    string          `json:"uninstallingProgress"`
	UninstallingProgressNum int             `json:"-"`
	UpgradingTarget         string          `json:"upgradingTarget"`
	UpgradingRetryNum       int             `json:"upgradingRetryNum"`
	UpgradingNextRetryAt    *time.Time      `json:"upgradingNextRetryAt,omitempty"`
	UpgradingState          ProcessingState `json:"upgradingState"`
	UpgradingStep           string          `json:"upgradingStep"`
	UpgradingProgress       string          `json:"upgradingProgress"`
	UpgradingProgressNum    int             `json:"-"`
	UpgradingError          string          `json:"upgradingError"`

	UpgradingDownloadState       ProcessingState `json:"upgradingDownloadState"`
	UpgradingDownloadStep        string          `json:"upgradingDownloadStep"`
	UpgradingDownloadProgress    string          `json:"upgradingDownloadProgress"`
	UpgradingDownloadProgressNum int             `json:"-"`
	UpgradingDownloadError       string          `json:"upgradingDownloadError"`

	CollectingLogsState ProcessingState `json:"collectingLogsState"`
	CollectingLogsError string          `json:"collectingLogsError"`

	DefaultFRPServer string `json:"defaultFrpServer"`
	FRPEnable        string `json:"frpEnable"`

	ContainerMode *string `json:"containerMode,omitempty"`

	Pressure []utils.NodePressure `json:"pressures,omitempty"`
}

var CurrentState state
var StateTrigger chan struct{}
var TerminusStateMu sync.Mutex

func init() {
	CurrentState.TerminusState = Checking
	CurrentState.TerminusdState = Initialize
	StateTrigger = make(chan struct{})
}

func (c *state) ChangeTerminusStateTo(s TerminusState) {
	TerminusStateMu.Lock()
	defer TerminusStateMu.Unlock()

	c.TerminusState = s
}

func bToGb(b uint64) string {
	return fmt.Sprintf("%d G", b/1024/1024/1024)
}

func CheckCurrentStatus(ctx context.Context) error {
	TerminusStateMu.Lock()
	name, err := utils.GetOlaresNameFromReleaseFile()
	if err != nil {
		klog.Error("get olares name from release file error, ", err)
	} else {
		CurrentState.TerminusName = &name
	}

	var currentTerminusState TerminusState = CurrentState.TerminusState
	defer func() {
		CurrentState.TerminusState = currentTerminusState
		TerminusStateMu.Unlock()
		klog.Info("current state: ", CurrentState.TerminusState)
	}()

	utils.ForceMountHdd(ctx)

	// set default value
	if CurrentState.TerminusVersion == nil {
		CurrentState.TerminusVersion = &commands.INSTALLED_VERSION
	}
	CurrentState.DefaultFRPServer = os.Getenv("FRP_SERVER")
	CurrentState.FRPEnable = os.Getenv("FRP_ENABLE")
	container := os.Getenv("CONTAINER_MODE")
	if container != "" {
		CurrentState.ContainerMode = pointer.String(container)
	}

	// get machine info
	osType, osInfo, osArch, osVersion, _, err := GetMachineInfo(ctx)
	if err != nil {
		klog.Error("get machine info from terminus cli error, ", err)
		return err
	}

	diskSize, err := utils.GetDiskSize()
	if err != nil {
		return err
	}

	gpu, err := utils.GetGpuInfo()
	if err != nil {
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	CurrentState.OsArch = osArch
	CurrentState.OsInfo = osInfo
	CurrentState.OsVersion = osVersion
	CurrentState.OsType = osType
	CurrentState.DeviceName = utils.GetDeviceName()
	CurrentState.CpuInfo = cpu.CPU.BrandName
	CurrentState.Memory = bToGb(memory.TotalMemory())
	CurrentState.Disk = diskSize
	CurrentState.GpuInfo = gpu
	CurrentState.HostName = &hostname

	// get network info
	ips, err := nets.GetInternalIpv4Addr()
	if err != nil {
		currentTerminusState = NetworkNotReady
		return err
	}

	if _, err := utils.ManagedAllDevices(ctx); err != nil {
		klog.Error("managed all devices error, ", err)
	}

	devices, err := utils.GetAllDevice(ctx)
	if err != nil {
		klog.Error(err)
	}

	// clear value
	CurrentState.WikiConnected = false
	CurrentState.WifiSSID = nil
	CurrentState.WiredConnected = false
	for _, i := range ips {
		if devices != nil {
			if d, ok := devices[i.Iface.Name]; ok {
				switch d.Type {
				case "wifi":
					CurrentState.WikiConnected = true
					CurrentState.WifiSSID = &d.Connection
				case "ethernet":
					CurrentState.WiredConnected = true
				}
				continue
			}
		}

		// for macos
		if strings.HasPrefix(i.Iface.Name, "wl") {
			CurrentState.WikiConnected = true
			CurrentState.WifiSSID = utils.WifiName()
		}

		if strings.HasPrefix(i.Iface.Name, "en") || strings.HasPrefix(i.Iface.Name, "eth") {
			CurrentState.WiredConnected = true
		}

	}

	hostIp, err := nets.GetHostIp()
	if err != nil {
		return err
	}

	fix := func() error {
		if len(ips) > 0 {
			hostIp = ips[0].IP
			err = nets.FixHostIP(hostIp)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if !slices.ContainsFunc(ips, func(i *nets.NetInterface) bool { return i.IP == hostIp }) {
		// wrong host ip
		klog.Warningf("host ip %s not in internal ips, try to fix it", hostIp)
		if err = fix(); err != nil {
			return err
		}
	}

	if conflict, err := nets.ConflictDomainIpInHostsFile(hostname); err != nil {
		return err
	} else if conflict {
		klog.Warningf("domain %s conflict with internal ip, try to fix it", hostname)
		if err = fix(); err != nil {
			return err
		}
	}

	CurrentState.HostIP = hostIp
	CurrentState.ExternalIP = nets.GetMyExternalIPAddr()

	// get olares state

	if shutdown, err := IsSystemShuttingdown(); err != nil {
		return err
	} else if shutdown {
		currentTerminusState = Shutdown
		return nil
	}

	if CurrentState.TerminusState != IPChanging {
		// if olaresd restarted, retain ip changing state
		if ipChanging, err := IsIpChangeRunning(); err != nil {
			if err == ErrChangeIpFailed {
				klog.Error("check state, change ip process failed")
				currentTerminusState = IPChangeFailed
				return nil
			}

			return err
		} else if ipChanging {
			klog.Warning("ip changing command is still running, state mismatch")
			return nil
		} else {
			// device reboot, ip change failed file still exists
			if _, err = os.Stat(commands.PREV_IP_CHANGE_FAILED); err == nil {
				// resume ip change failed state
				currentTerminusState = IPChangeFailed
				return nil
			} else if CurrentState.TerminusState == IPChanging {
				// ip changing not in process, however the current state is mismatch
				// set temporary state to checking
				currentTerminusState = Checking
			}
		}
	}

	// if ip changing in progress, waiting for it finished anyway
	if currentTerminusState == IPChanging {
		return nil
	}

	if installing, err := IsTerminusInstalling(); err != nil {
		if err == ErrInstallFailed {
			currentTerminusState = InstallFailed
			return nil
		}

		return err
	} else if installing {
		currentTerminusState = Installing
		return nil
	}

	if installed, err := IsTerminusInstalled(); err != nil {
		return err
	} else if !installed {
		if k3srunning, err := IsK3SRunning(ctx); err != nil {
			return err
		} else if k3srunning {
			currentTerminusState = SystemError
			return nil
		}

		currentTerminusState = NotInstalled
		CurrentState.InstallingProgress = ""
		CurrentState.InstallingState = ""
		CurrentState.InstalledTime = nil
		CurrentState.InitializedTime = nil

	} else {
		currentTerminusState = Uninitialized
	}

	if changed, err := utils.IsIpChanged(ctx, CurrentState.TerminusState != NotInstalled); err != nil {
		return err
	} else if changed {
		currentTerminusState = InvalidIpAddress
		return nil
	}

	if currentTerminusState == NotInstalled {
		return nil
	}

	kubeClient, err := utils.GetKubeClient()
	if err != nil {
		currentTerminusState = SystemError
		return err
	}
	dynamicClient, err := utils.GetDynamicClient()
	if err != nil {
		currentTerminusState = SystemError
		return err
	}

	pressure, err := utils.GetNodesPressure(ctx, kubeClient)
	if err != nil {
		klog.Error("get nodes pressure error, ", err)
	} else {
		// update node pressure of current node
		if p, ok := pressure[*CurrentState.HostName]; ok && len(p) == 0 {
			CurrentState.Pressure = p
		} else {
			CurrentState.Pressure = nil
		}
	}

	if CurrentState.InstallFinishedTime != nil {
		CurrentState.InstalledTime = pointer.Int64(CurrentState.InstallFinishedTime.Unix())
	} else {
		CurrentState.InstalledTime, err = utils.GetTerminusInstalledTime(ctx, dynamicClient, kubeClient)
		if err != nil {
			klog.Error(err)
		}
	}

	// only set system state to Upgrading if actual upgrade should be in progress
	// (not during download phase)
	upgradeTarget, err := GetOlaresUpgradeTarget()
	if err != nil {
		return fmt.Errorf("error getting Olares upgrade target: %v", err.Error())
	}
	if upgradeTarget != nil && upgradeTarget.Downloaded && !upgradeTarget.DownloadOnly {
		currentTerminusState = Upgrading
		return nil
	}

	if tmsrunning, err := utils.IsTerminusRunning(ctx, kubeClient); err != nil {
		currentTerminusState = SystemError
		return err
	} else if tmsrunning {
		currentTerminusState = Uninitialized

		terminusName, err := utils.GetAdminUserTerminusName(ctx, dynamicClient)
		if err != nil {
			klog.Error("get user olares name error, ", err)
		} else {
			CurrentState.TerminusName = &terminusName
		}

		terminusVerion, err := utils.GetTerminusVersion(ctx, dynamicClient)
		if err != nil {
			klog.Error("get olares version error, ", err)
		} else {
			CurrentState.TerminusVersion = terminusVerion
		}

		inited, failed, err := utils.IsTerminusInitialized(ctx, dynamicClient)
		if err != nil {
			currentTerminusState = SystemError
			klog.Error("check olares initialized error, ", err)

			// check status error, report state as system error
			return nil
		}

		if inited {
			currentTerminusState = TerminusRunning
			CurrentState.InitializedTime, err = utils.GetTerminusInitializedTime(ctx, kubeClient)
			if err != nil {
				klog.Error(err)
			}

			restarting, err := utils.SystemStartLessThan(3 * time.Minute) // uptime less then 3 minutes
			if err != nil {
				return err
			}

			// if the uptime is less than 3m, the pods running state maybe are fake.
			// waiting for the kublet checking the all pods state to update the real state
			if restarting {
				currentTerminusState = Restarting
				return nil
			}

			return nil
		}

		if failed {
			currentTerminusState = InitializeFailed
			return nil
		}

		initing, err := utils.IsTerminusInitializing(ctx, dynamicClient)
		if err != nil {
			return err
		}

		if initing {
			currentTerminusState = Initializing
			return nil
		}

		return nil
	} else {
		// some key pods are abnormal
		if CurrentState.InstallFinishedTime != nil && time.Since(*CurrentState.InstallFinishedTime) < 5*time.Minute {
			currentTerminusState = Installing
			return nil
		}

		restarting, err := utils.SystemStartLessThan(10 * time.Minute) // uptime less then 10 minutes
		if err != nil {
			return err
		}

		if restarting {
			currentTerminusState = Restarting
			return nil
		}

		currentTerminusState = SystemError
	}

	return nil
}

func WatchStatus(ctx context.Context, watchers []watcher.Watcher, postWatch func()) {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		run := func() {
			err := CheckCurrentStatus(ctx)
			if err != nil {
				klog.Warning("check olares status error, ", err)
			}

			for _, w := range watchers {
				w.Watch(ctx)
			}

			postWatch()
		}

		for {
			select {
			case <-StateTrigger:
				run()
			case <-ticker.C:
				run()
			case <-ctx.Done():
				return
			}
		}
	}()
}
