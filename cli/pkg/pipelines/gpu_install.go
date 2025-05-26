package pipelines

import (
	"path"

	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/gpu"
	"bytetrade.io/web3os/installer/pkg/manifest"
)

func InstallGpuDrivers(opt *options.InstallGpuOptions) error {
	arg := common.NewArgument()
	arg.SetOlaresVersion(opt.Version)
	arg.SetCudaVersion(opt.Cuda)
	arg.SetBaseDir(opt.BaseDir)
	arg.SetConsoleLog("gpuinstall.log", true)
	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	manifestFile := path.Join(runtime.GetInstallerDir(), "installation.manifest")

	runtime.Arg.SetManifest(manifestFile)

	manifestMap, err := manifest.ReadAll(runtime.Arg.Manifest)
	if err != nil {
		logger.Fatal(err)
	}

	p := &pipeline.Pipeline{
		Name: "InstallGpuDrivers",
		Modules: []module.Module{
			&gpu.InstallDriversModule{
				ManifestModule: manifest.ManifestModule{
					Manifest: manifestMap,
					BaseDir:  runtime.Arg.BaseDir,
				},
				FailOnNoInstallation: true,
			},
			&gpu.InstallContainerToolkitModule{
				ManifestModule: manifest.ManifestModule{
					Manifest: manifestMap,
					BaseDir:  runtime.Arg.BaseDir,
				},
			},
			&gpu.RestartContainerdModule{},
			&gpu.NodeLabelingModule{},
		},
		Runtime: runtime,
	}

	return p.Start()

}
