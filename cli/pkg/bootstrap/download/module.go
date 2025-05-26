package download

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/task"
)

type PackageDownloadModule struct {
	common.KubeModule
	Manifest       string
	BaseDir        string
	DownloadCdnUrl string
}

func (i *PackageDownloadModule) Init() {
	i.Name = "PackageDownloadModule"

	download := &task.LocalTask{
		Name:   i.Name,
		Desc:   i.Desc,
		Action: &PackageDownload{Manifest: i.Manifest, BaseDir: i.BaseDir, DownloadCdnUrl: i.DownloadCdnUrl},
	}

	i.Tasks = []task.Interface{
		download,
	}
}

type CheckDownloadModule struct {
	common.KubeModule
	Manifest string
	BaseDir  string
}

func (i *CheckDownloadModule) Init() {
	i.Name = "CheckDownloadModule"

	check := &task.LocalTask{
		Name:   i.Name,
		Desc:   i.Desc,
		Action: &CheckDownload{PackageDownload{Manifest: i.Manifest, BaseDir: i.BaseDir}},
	}

	i.Tasks = []task.Interface{
		check,
	}
}
