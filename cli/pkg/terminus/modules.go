package terminus

import (
	"fmt"
	"os"
	"path"
	"time"

	bootstrapos "bytetrade.io/web3os/installer/pkg/bootstrap/os"
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/container"
	containertemplates "bytetrade.io/web3os/installer/pkg/container/templates"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/etcd"
	etcdtemplates "bytetrade.io/web3os/installer/pkg/etcd/templates"
	"bytetrade.io/web3os/installer/pkg/k3s"
	k3stemplates "bytetrade.io/web3os/installer/pkg/k3s/templates"
	"bytetrade.io/web3os/installer/pkg/kubernetes"
	ktemplates "bytetrade.io/web3os/installer/pkg/kubernetes/templates"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/storage"
	storageemplates "bytetrade.io/web3os/installer/pkg/storage/templates"
	"bytetrade.io/web3os/installer/pkg/utils"
)

type InstallWizardDownloadModule struct {
	common.KubeModule
	Version        string
	DownloadCdnUrl string
}

func (m *InstallWizardDownloadModule) Init() {
	m.Name = "DownloadInstallWizard"
	download := &task.LocalTask{
		Name: "DownloadInstallWizard",
		Action: &Download{
			Version:        m.Version,
			DownloadCdnUrl: m.DownloadCdnUrl,
		},
		Retry: 1,
	}

	m.Tasks = []task.Interface{
		download,
	}
}

type PreparedModule struct {
	common.KubeModule
}

func (m *PreparedModule) Init() {
	m.Name = "PrepareFinished"

	prepareFinished := &task.LocalTask{
		Name:   "PrepareFinished",
		Action: &PrepareFinished{},
	}

	m.Tasks = []task.Interface{
		prepareFinished,
	}
}

type WriteReleaseFileModule struct {
	common.KubeModule
}

func (m *WriteReleaseFileModule) Init() {
	m.Name = "WriteReleaseFile"

	m.Tasks = []task.Interface{
		&task.LocalTask{
			Name:   "WriteReleaseFile",
			Action: new(WriteReleaseFile),
		},
	}
}

type RemoveReleaseFileModule struct {
	common.KubeModule
}

func (m *RemoveReleaseFileModule) Init() {
	m.Name = "RemoveReleaseFile"

	m.Tasks = []task.Interface{
		&task.LocalTask{
			Name:   "RemoveReleaseFile",
			Action: new(RemoveReleaseFile),
		},
	}
}

type CheckPreparedModule struct {
	common.KubeModule
	Force bool
}

func (m *CheckPreparedModule) Init() {
	m.Name = "CheckPrepared"

	checkPrepared := &task.LocalTask{
		Name:   "CheckPrepared",
		Action: &CheckPrepared{Force: m.Force},
	}

	m.Tasks = []task.Interface{
		checkPrepared,
	}
}

type CheckInstalledModule struct {
	common.KubeModule
	Force bool
}

func (m *CheckInstalledModule) Init() {
	m.Name = "CheckInstalled"

	checkPrepared := &task.LocalTask{
		Name:   "CheckInstalled",
		Action: &CheckInstalled{Force: m.Force},
	}

	m.Tasks = []task.Interface{
		checkPrepared,
	}
}

type InstallComponentsInClusterModule struct {
	common.KubeModule
}

type GetNATGatewayIPModule struct {
	common.KubeModule
}

func (m *GetNATGatewayIPModule) Init() {
	m.Name = "GetNATGatewayIP"

	getNATGatewayIP := &task.LocalTask{
		Name:   "GetNATGatewayIP",
		Action: new(GetNATGatewayIP),
	}

	m.Tasks = []task.Interface{
		getNATGatewayIP,
	}
}

func GenerateTerminusComponentsModules(runtime connector.Runtime, manifestMap manifest.InstallationManifest) []module.Module {
	var modules []module.Module
	baseModules := []module.Module{
		&GetNATGatewayIPModule{},
		&InstallAccountModule{},
		&InstallSettingsModule{},
		&InstallOsSystemModule{},
		&InstallLauncherModule{},
		&InstallAppsModule{},
	}
	modules = append(modules, baseModules...)

	si := runtime.GetSystemInfo()
	switch {
	case si.IsDarwin():
	default:
		modules = append(
			modules,
			&InstallVeleroModule{
				ManifestModule: manifest.ManifestModule{
					Manifest: manifestMap,
					BaseDir:  runtime.GetBaseDir(),
				},
			})
	}

	modules = append(modules, &WelcomeModule{})

	return modules
}

type InstalledModule struct {
	common.KubeModule
}

func (m *InstalledModule) Init() {
	m.Name = "InstallFinished"

	installedState := &task.LocalTask{
		Name:   "InstallFinished",
		Action: &InstallFinished{},
	}

	m.Tasks = []task.Interface{
		installedState,
	}
}

type DeleteWizardFilesModule struct {
	common.KubeModule
}

func (d *DeleteWizardFilesModule) Init() {
	d.Name = "DeleteWizardFiles"

	deleteWizardFiles := &task.LocalTask{
		Name:   "DeleteWizardFiles",
		Action: &DeleteWizardFiles{},
	}

	d.Tasks = []task.Interface{
		deleteWizardFiles,
	}
}

var (
	systemdUnitDir        = "/etc/systemd/system/"
	k3sServiceName        = k3stemplates.K3sService.Name()
	kubeletServiceName    = ktemplates.KubeletService.Name()
	containerdServiceName = containertemplates.ContainerdService.Name()
	etcdServiceName       = etcdtemplates.ETCDService.Name()
	juiceFSServiceName    = storageemplates.JuicefsService.Name()
	redisServiceName      = storageemplates.RedisService.Name()
	minIOServiceName      = storageemplates.MinioService.Name()
)

func serviceExists(serviceName string) bool {
	return util.IsExist(path.Join(systemdUnitDir, serviceName))
}

type StopOlaresModule struct {
	common.KubeModule
	Timeout       time.Duration
	CheckInterval time.Duration
}

func (m *StopOlaresModule) Init() {
	m.Name = "StopOlares"

	newStopServiceTask := func(serviceName string) task.Interface {
		return &task.LocalTask{
			Name:  fmt.Sprintf("Stop %s", serviceName),
			Retry: 3,
			Action: &SystemctlCommand{
				Command:   "stop",
				UnitNames: []string{serviceName},
			},
		}
	}

	k3sServiceExists := serviceExists(k3sServiceName)
	kubeletServiceExists := serviceExists(kubeletServiceName)
	if k3sServiceExists {
		m.Tasks = append(m.Tasks, newStopServiceTask(k3sServiceName))
	}
	if kubeletServiceExists {
		m.Tasks = append(m.Tasks, newStopServiceTask(kubeletServiceName))
	}
	if serviceExists(containerdServiceName) {
		m.Tasks = append(m.Tasks, newStopServiceTask(containerdServiceName))
	}
	if k3sServiceExists || kubeletServiceExists {
		m.Tasks = append(m.Tasks,
			&task.LocalTask{
				Name: "KillContainers",
				Action: &container.KillContainerdProcess{
					Signal:        "TERM",
					Timeout:       m.Timeout,
					CheckInterval: m.CheckInterval,
				},
				Retry: 3,
			},
			&task.LocalTask{
				Name:   "ClearKubernetesMounts",
				Action: new(kubernetes.UmountKubelet),
				Retry:  3,
			},
			&task.LocalTask{
				Name:   "ClearKubernetesNetworkConfigs",
				Action: new(bootstrapos.ResetNetworkConfig),
				Retry:  3,
			},
		)
	}
	for _, service := range []string{
		etcdServiceName,
		juiceFSServiceName,
		redisServiceName,
		minIOServiceName,
	} {
		if serviceExists(service) {
			m.Tasks = append(m.Tasks, newStopServiceTask(service))
		}
	}
	if len(m.Tasks) == 0 {
		logger.Info("found no components of Olares, please install Olares first")
		logger.Info("exiting ...")
		os.Exit(0)
	} else if !k3sServiceExists && !kubeletServiceExists {
		logger.Warn("kubernetes service can not be found, it seems that Olares has not been installed yet")
		logger.Warn("will try to stop any other base components that we can find")
	}
}

type StartOlaresModule struct {
	common.KubeModule
}

func (m *StartOlaresModule) Init() {
	m.Name = "StartOlares"

	newStartServiceTask := func(serviceName string) task.Interface {
		return &task.LocalTask{
			Name:  fmt.Sprintf("Start %s", serviceName),
			Retry: 3,
			Delay: 15 * time.Second,
			Action: &SystemctlCommand{
				Command:   "start",
				UnitNames: []string{serviceName},
			},
		}
	}

	for _, service := range []string{
		minIOServiceName,
		redisServiceName,
		juiceFSServiceName,
		etcdServiceName,

		// backupETCDService is triggered by systemd.timer
		// and does not need to be manually started, it will only do an unnecessary backup
		// backupETCDServiceName,
		containerdServiceName,
	} {
		if serviceExists(service) {
			m.Tasks = append(m.Tasks, newStartServiceTask(service))
		}
	}
	k3sServiceExists := serviceExists(k3sServiceName)
	kubeletServiceExists := serviceExists(kubeletServiceName)
	if k3sServiceExists {
		m.Tasks = append(m.Tasks, newStartServiceTask(k3sServiceName))
	}
	if kubeletServiceExists {
		m.Tasks = append(m.Tasks, newStartServiceTask(kubeletServiceName))
	}
	if len(m.Tasks) == 0 {
		logger.Info("found no components of Olares, please install Olares first")
		logger.Info("exiting ...")
		os.Exit(0)
	} else if k3sServiceExists || kubeletServiceExists {
		m.Tasks = append(m.Tasks, &task.LocalTask{

			// when starting an already stopped Olares, which is the normal use case
			// it's very likely for this program to outrun kubelet
			// and retrieve a pod list from kube-apiserver that is a "snapshot" of the cluster state
			// when the kubernetes components was killed, before. i.e. the data will say all the pods are running while they're actually not
			// there's no easy way for us to detect whether the retrieved data represents the current cluster state
			// so we just wait some time for the kubelet to update the pod status before retrieving it from kube-apiserver
			// again, in normal cases, it takes much longer than this period for the pods to be back running again
			// thus no harm is caused
			Prepare: &prepare.InitialDelay{
				Duration: 30 * time.Second,
			},
			Name:   "EnsurePodsUpAndRunningAgain",
			Action: &CheckKeyPodsRunning{Node: m.Runtime.GetLocalHost().GetName()},
			Delay:  10 * time.Second,
			Retry:  60,
		})
	} else {
		logger.Warn("kubernetes service can not be found, it seems that Olares has not been installed yet")
		logger.Warn("will try to start any other base components that we can find")
	}

}

type ChangeIPModule struct {
	common.KubeModule
}

func (m *ChangeIPModule) Init() {
	m.Name = "ChangeIP"

	prepared, _ := m.PipelineCache.GetMustBool(common.CachePreparedState)
	if !prepared {
		logger.Info("the Olares OS is not prepared, will only try to update /etc/hosts")
	}
	m.Tasks = []task.Interface{
		&task.LocalTask{
			Name:   "RemoveFileAttributesForUpdating",
			Action: new(precheck.RemoveWSLChattr),
		},
		&task.LocalTask{
			Name:   "UpdateHosts",
			Action: new(UpdateKubeKeyHosts),
		},
	}

	installed, _ := m.PipelineCache.GetMustBool(common.CacheInstalledState)
	if !installed && prepared {
		logger.Info("the Olares OS is not installed, will only try to update prepared base components")
	}

	if installed {
		stopKubeTask := &task.LocalTask{
			Name:  "StopKubernetes",
			Retry: 3,
		}
		stopKubeAction := &SystemctlCommand{
			Command: "stop",
		}
		if m.KubeConf.Arg.Kubetype == common.K3s {
			stopKubeAction.UnitNames = []string{"k3s"}
		} else {
			stopKubeAction.UnitNames = []string{"kubelet"}
		}
		if util.IsExist("/etc/systemd/system/etcd.service") {
			stopKubeAction.UnitNames = append(stopKubeAction.UnitNames, "etcd", "backup-etcd")
		}
		stopKubeTask.Action = stopKubeAction
		m.Tasks = append(m.Tasks, stopKubeTask)
	}

	if prepared {
		m.addStorageTasks()
	}
	if installed {
		m.addEtcdTasks()
		m.addKubernetesTasks()
		m.addRestartTasks()
	}
	if m.KubeConf.Arg.MasterHost != "" {
		m.Tasks = append(m.Tasks,
			&task.LocalTask{
				Name:   "SaveMasterHostConfig",
				Action: new(SaveMasterHostConfig),
			})
	}
	m.Tasks = append(m.Tasks,
		&task.LocalTask{
			Name:   "RewriteFileAttributes",
			Action: new(precheck.AddWSLChattr),
		})
}

func (m *ChangeIPModule) addStorageTasks() {
	var storageComponents []string
	juiceFSExists := util.IsExist(storage.JuiceFsServiceFile)
	if juiceFSExists {
		storageComponents = append(storageComponents, "juicefs")
	} else {
		logger.Info("JuiceFS is not installed, no storage component needs IP reconfiguration.")
		return
	}
	redisExists := util.IsExist(storage.RedisServiceFile)
	if redisExists {
		storageComponents = append(storageComponents, "redis-server")
	}
	minioExists := util.IsExist(storage.MinioServiceFile)
	if minioExists {
		storageComponents = append(storageComponents, "minio")
	}
	m.Tasks = append(m.Tasks,
		&task.LocalTask{
			Name: "StopStorageComponents",
			Action: &SystemctlCommand{
				Command:   "stop",
				UnitNames: storageComponents,
			},
			Retry: 3,
		},

		&task.RemoteTask{
			Name:   "GetRedisConfig",
			Hosts:  m.Runtime.GetHostsByRole(common.Master),
			Action: new(storage.GetOrSetRedisConfig),
			Retry:  1,
		},
	)
	if redisExists {
		m.Tasks = append(m.Tasks,
			&task.LocalTask{
				Name:   "ReGenerateRedisService",
				Action: new(storage.GenerateRedisService),
			},
			&task.LocalTask{
				Name:   "EnableRedisService",
				Action: new(storage.EnableRedisService),
				Retry:  3,
			},
			&task.LocalTask{
				Name:   "CheckRedisState",
				Action: new(storage.CheckRedisServiceState),
				Retry:  20,
			},
		)
	}

	if minioExists {
		m.Tasks = append(m.Tasks,
			&task.LocalTask{
				Name:   "GetOrSetMinIOPassword",
				Action: new(storage.GetOrSetMinIOPassword),
			},
			&task.LocalTask{
				Name:   "ReConfigureMinio",
				Action: new(storage.ConfigMinio),
			},
			&task.LocalTask{
				Name:   "EnableMinioService",
				Action: new(storage.EnableMinio),
			},
			&task.LocalTask{
				Name:   "CheckMinioState",
				Action: new(storage.CheckMinioState),
				Retry:  20,
			},
			&task.LocalTask{
				Name:   "ConfigJuiceFSMetaDB",
				Action: new(storage.ConfigJuiceFsMetaDB),
			},
		)
	}

	m.Tasks = append(m.Tasks,
		&task.LocalTask{
			Name:   "EnableJuiceFsService",
			Action: new(storage.EnableJuiceFsService),
		},

		&task.LocalTask{
			Name:   "CheckJuiceFsState",
			Action: new(storage.CheckJuiceFsState),
			Retry:  20,
		},
	)
}

func (m *ChangeIPModule) addEtcdTasks() {
	if !util.IsExist("/etc/systemd/system/etcd.service") {
		return
	}
	m.Tasks = append(m.Tasks,
		&task.RemoteTask{
			Name:   "GetETCDStatus",
			Action: new(etcd.GetStatus),
			Hosts:  m.Runtime.GetHostsByRole(common.ETCD),
		},
		&task.RemoteTask{
			Name:   "GenerateETCDAccessAddress",
			Hosts:  m.Runtime.GetHostsByRole(common.ETCD),
			Action: new(etcd.GenerateAccessAddress),
		},
		&task.LocalTask{
			Name:   "PrepareETCDFiles",
			Action: new(PrepareFilesForETCDIPChange),
		},
		&task.LocalTask{
			Name:   "RegenerateETCDCerts",
			Action: new(etcd.GenerateCerts),
		},
		&task.RemoteTask{
			Name:   "SyncETCDCerts",
			Action: new(etcd.SyncCertsFile),
			Hosts:  m.Runtime.GetHostsByRole(common.ETCD),
		},
		&task.RemoteTask{
			Name:   "RefreshETCDConfig",
			Action: new(etcd.RefreshConfig),
			Hosts:  m.Runtime.GetHostsByRole(common.ETCD),
		},
		&task.RemoteTask{
			Name:   "RestartETCD",
			Action: new(etcd.RestartETCD),
			Hosts:  m.Runtime.GetHostsByRole(common.ETCD),
		},
		&task.RemoteTask{
			Name:   "ETCDHealthCheck",
			Action: new(etcd.HealthCheck),
			Hosts:  m.Runtime.GetHostsByRole(common.ETCD),
			Retry:  20,
		},
		&task.RemoteTask{
			Name:   "RefreshBackupETCDScript",
			Action: new(etcd.BackupETCD),
			Hosts:  m.Runtime.GetHostsByRole(common.ETCD),
		})
}

func (m *ChangeIPModule) addKubernetesTasks() {
	if m.KubeConf.Arg.Kubetype == common.K3s {
		cluster := k3s.NewK3sStatus()
		m.PipelineCache.GetOrSet(common.ClusterStatus, cluster)
		if !m.Runtime.GetLocalHost().IsRole(common.Master) {
			m.Tasks = append(m.Tasks,
				&task.RemoteTask{
					Name:     "GetClusterStatus(k3s)",
					Desc:     "Get k3s cluster status",
					Hosts:    m.Runtime.GetHostsByRole(common.Master),
					Action:   new(k3s.GetClusterStatus),
					Parallel: false,
				},
			)
		}
		m.Tasks = append(m.Tasks,
			&task.LocalTask{
				Name:   "RegenerateK3sService",
				Action: new(k3s.GenerateK3sService),
			},
			&task.LocalTask{
				Name:   "RegenerateK3sServiceEnv",
				Action: new(k3s.GenerateK3sServiceEnv),
			},
			&task.LocalTask{
				Name:   "EnableK3sService",
				Desc:   "Enable k3s service",
				Action: new(k3s.EnableK3sService),
			},
		)
	} else {
		m.Tasks = append(m.Tasks,
			&task.LocalTask{
				Name:   "RegenerateKubeletServiceEnv",
				Action: new(kubernetes.GenerateKubeletEnv),
			})
		// worker node, no need to reconfigure control-plane components
		if !util.IsExist("/etc/kubernetes/manifests/kube-apiserver.yaml") {
			m.Tasks = append(m.Tasks,
				&task.LocalTask{
					Name: "RestartKubelet",
					Action: &SystemctlCommand{
						Command:             "start",
						UnitNames:           []string{"kubelet"},
						DaemonReloadPreExec: true,
					},
				})
			return
		}
		m.Tasks = append(m.Tasks,
			&task.LocalTask{
				Name:   "PrepareK8sFiles",
				Action: new(PrepareFilesForK8sIPChange),
			},
			&task.RemoteTask{
				Name: "RegenerateKubeadmConfig",
				Action: &kubernetes.GenerateKubeadmConfig{
					IsInitConfiguration:     true,
					WithSecurityEnhancement: m.KubeConf.Arg.SecurityEnhancement,
				},
				Hosts: m.Runtime.GetHostsByRole(common.Master),
			},
			&task.LocalTask{
				Name:   "RegenerateK8sFilesWithKubeadm",
				Action: new(RegenerateFilesForK8sIPChange),
			},
			&task.LocalTask{
				Name:   "CopyNewKubeConfig",
				Action: new(kubernetes.CopyKubeConfigForControlPlane),
			},
		)
	}
}

func (m *ChangeIPModule) addRestartTasks() {
	restartPodsTasks := []task.Interface{
		&task.LocalTask{
			Name:   "RestartAllPods",
			Action: &DeleteAllPods{Node: m.Runtime.GetLocalHost().GetName()},
			Retry:  5,
			Delay:  15 * time.Second,
		},
	}
	if !utils.IsExist(storage.JuiceFsServiceFile) {
		restartPodsTasks = []task.Interface{
			&task.LocalTask{
				Name:   "RestartPodsUsingHostIP",
				Action: new(DeletePodsUsingHostIP),
				Delay:  3 * time.Second,
				Retry:  5,
			},
			// the app controllers in k3s have significant delay detecting state change
			// and there's a gap between a pod being deleted and recreated
			&task.LocalTask{
				Name:   "WaitForPodsUsingHostIPRecreate",
				Action: new(WaitForPodsUsingHostIPRecreate),
				Delay:  15 * time.Second,
				Retry:  20,
			},
		}
	}
	m.Tasks = append(m.Tasks,
		&task.LocalTask{
			Name:   "WaitForKubeAPIServerUp",
			Action: new(precheck.GetKubernetesNodesStatus),
			Retry:  20,
		})
	m.Tasks = append(m.Tasks, restartPodsTasks...)

	m.Tasks = append(m.Tasks, &task.LocalTask{
		Name:   "EnsurePodsUpAndRunningAgain",
		Action: &CheckKeyPodsRunning{Node: m.Runtime.GetLocalHost().GetName()},
		Delay:  10 * time.Second,
		Retry:  60,
	})
}

type ChangeHostIPModule struct {
	common.KubeModule
}

func (m *ChangeHostIPModule) Init() {
	m.Name = "ChangeHostIP"

	m.Tasks = append(m.Tasks,
		&task.LocalTask{
			Name:   "CheckOlaresStateInHost",
			Action: new(CheckTerminusStateInHost),
		},
		&task.LocalTask{
			Name:   "GetNATGatewayIP",
			Action: new(GetNATGatewayIP),
		},
		&task.LocalTask{
			Name:   "UpdateNATGatewayIPForUser",
			Action: new(UpdateNATGatewayForUser),
		},
	)
}

type GetMasterInfoModule struct {
	common.KubeModule
	Print bool
}

func (m *GetMasterInfoModule) Init() {
	m.Name = "GetMasterInfo"
	m.Tasks = append(m.Tasks,
		&task.RemoteTask{
			Name:   "GetMasterInfo",
			Action: &GetMasterInfo{Print: m.Print},
			Hosts:  m.Runtime.GetHostsByRole(common.Master),
		},
		&task.LocalTask{
			Name:   "AddNodePrecheck",
			Action: new(AddNodePrecheck),
		},
	)
}

type SaveMasterHostConfigModule struct {
	common.KubeModule
}

func (m *SaveMasterHostConfigModule) Init() {
	m.Name = "SaveMasterHostConfig"
	m.Tasks = append(m.Tasks,
		&task.LocalTask{
			Name:   "SaveMasterHostConfig",
			Action: new(SaveMasterHostConfig),
		})
}
