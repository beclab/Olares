package pipelines

import (
	"path"

	"github.com/beclab/Olares/cli/cmd/ctl/options"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
	"github.com/beclab/Olares/cli/pkg/gpu"
	"github.com/beclab/Olares/cli/pkg/manifest"
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
