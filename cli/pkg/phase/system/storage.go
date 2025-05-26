package system

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/storage"
	"os"
)

func InstallStoragePipeline(runtime *common.KubeRuntime) *pipeline.Pipeline {
	si := runtime.GetSystemInfo()
	if si.IsDarwin() || si.IsWindows() {
		logger.Infof("storage is supposed to be installed on Linux, no operation will be done on %s", si.GetOsType())
		os.Exit(0)
	}
	var modules []module.Module
	manifestMap, err := manifest.ReadAll(runtime.Arg.Manifest)
	if err != nil {
		logger.Fatal(err)
	}
	modules = []module.Module{
		&storage.ValidateModule{
			Skip: runtime.Arg.Storage.StorageType == common.ManagedMinIO,
		},
		&storage.InstallMinioModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: manifestMap,
				BaseDir:  runtime.GetBaseDir(), // l.runtime.Arg.BaseDir,
			},
			Skip: runtime.Arg.Storage.StorageType != common.ManagedMinIO,
		},
	}

	return &pipeline.Pipeline{
		Name:    "Install Storage",
		Modules: modules,
		Runtime: runtime,
	}
}
