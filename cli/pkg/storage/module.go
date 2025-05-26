package storage

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
)

type DeleteTmpModule struct {
	common.KubeModule
}

func (m *DeleteTmpModule) Init() {
	m.Name = "DeleteTmp"
}

type InitStorageModule struct {
	common.KubeModule
	Skip bool
}

func (m *InitStorageModule) IsSkip() bool {
	return m.Skip
}

func (m *InitStorageModule) Init() {
	m.Name = "InitStorage"

	mkStorageDir := &task.RemoteTask{
		Name:  "CreateStorageDir",
		Hosts: m.Runtime.GetAllHosts(),
		Prepare: &prepare.PrepareCollection{
			&CheckStorageVendor{},
		},
		Action:   new(MkStorageDir),
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		mkStorageDir,
	}
}

type RemoveMountModule struct {
	common.KubeModule
}

func (m *RemoveMountModule) Init() {
	m.Name = "DeleteS3Mount"

	downloadStorageCli := &task.RemoteTask{
		Name:  "DownloadStorageCli",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(CheckStorageVendor),
		},
		Action:   new(DownloadStorageCli),
		Parallel: false,
		Retry:    1,
	}

	unMountOSS := &task.RemoteTask{
		Name:  "UnMountOSS",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&CheckStorageType{
				StorageType: common.OSS,
			},
		},
		Action:   new(UnMountOSS),
		Parallel: false,
		Retry:    1,
	}

	unMountCOS := &task.RemoteTask{
		Name:  "UnMountCOS",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&CheckStorageType{
				StorageType: common.COS,
			},
		},
		Action:   new(UnMountCOS),
		Parallel: false,
		Retry:    1,
	}

	unMountS3 := &task.RemoteTask{
		Name:  "UnMountS3",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&CheckStorageType{
				StorageType: common.S3,
			},
		},
		Action:   new(UnMountS3),
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		downloadStorageCli,
		unMountOSS,
		unMountCOS,
		unMountS3,
	}
}

type RemoveStorageModule struct {
	common.KubeModule
}

func (m *RemoveStorageModule) Init() {
	m.Name = "RemoveStorage"

	stopMinio := &task.RemoteTask{
		Name:  "StopMinio",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(StopMinio),
		Parallel: false,
		Retry:    0,
	}

	stopMinioOperator := &task.RemoteTask{
		Name:  "StopMinioOperator",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(StopMinioOperator),
		Parallel: false,
		Retry:    0,
	}

	removeTerminusFiles := &task.RemoteTask{
		Name:  "RemoveOlaresFiles",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(RemoveTerminusFiles),
		Parallel: false,
		Retry:    0,
	}

	m.Tasks = []task.Interface{
		stopMinio,
		stopMinioOperator,
		removeTerminusFiles,
	}
}

type RemoveJuiceFSModule struct {
	common.KubeModule
}

func (m *RemoveJuiceFSModule) Init() {
	m.Name = "RemoveJuiceFS"

	stopJuiceFS := &task.RemoteTask{
		Name:  "StopJuiceFS",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(StopJuiceFS),
		Parallel: false,
		Retry:    0,
	}

	stopRedis := &task.RemoteTask{
		Name:  "StopRedis",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(StopRedis),
		Parallel: false,
		Retry:    0,
	}

	removeJuiceFSFiles := &task.RemoteTask{
		Name:  "RemoveJuiceFSFiles",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(RemoveJuiceFSFiles),
		Parallel: false,
		Retry:    0,
	}
	m.Tasks = []task.Interface{
		stopJuiceFS,
		stopRedis,
		removeJuiceFSFiles,
	}
}

type DeletePhaseFlagModule struct {
	common.KubeModule
	PhaseFile string
	BaseDir   string
}

func (m *DeletePhaseFlagModule) Init() {
	m.Name = "DeletePhaseFlag"

	deletePhaseFlagFile := &task.LocalTask{
		Name: "DeletePhaseFlag",
		Action: &DeletePhaseFlagFile{
			PhaseFile: m.PhaseFile,
			BaseDir:   m.BaseDir,
		},
	}

	m.Tasks = []task.Interface{
		deletePhaseFlagFile,
	}
}

type DeleteCacheModule struct {
	common.KubeModule
	BaseDir string
}

func (m *DeleteCacheModule) Init() {
	m.Name = "DeleteCaches"

	deleteCaches := &task.RemoteTask{
		Name:  "DeleteCaches",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Action: &DeleteCaches{
			BaseDir: m.BaseDir,
		},
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		deleteCaches,
	}
}

type DeleteUserDataModule struct {
	common.KubeModule
}

func (m *DeleteUserDataModule) Init() {
	m.Name = "DeleteUserData"

	deleteTerminusUserData := &task.RemoteTask{
		Name:     "DeleteUserData",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Action:   new(DeleteTerminusUserData),
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		deleteTerminusUserData,
	}
}

type DeleteTerminusDataModule struct {
	common.KubeModule
}

func (m *DeleteTerminusDataModule) Init() {
	m.Name = "DeleteOlaresData"

	deleteTerminusData := &task.RemoteTask{
		Name:     "DeleteOlaresData",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Action:   new(DeleteTerminusData),
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		deleteTerminusData,
	}
}
