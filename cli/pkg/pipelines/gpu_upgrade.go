package pipelines

import (
	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/gpu"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/utils"
	"fmt"
	"io"
	"os/exec"
	"path"
	"strings"
)

func UpgradeGpuDrivers(opt *options.InstallGpuOptions) error {
	arg := common.NewArgument()
	arg.SetOlaresVersion(opt.Version)
	arg.SetCudaVersion(opt.Cuda)
	arg.SetBaseDir(opt.BaseDir)
	arg.SetConsoleLog("gpuupgrade.log", true)
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
		Name: "UpgradeGpuDrivers",
		Modules: []module.Module{
			&gpu.ExitIfNoDriverUpgradeNeededModule{},
			&gpu.UninstallCudaModule{},
			&gpu.InstallDriversModule{
				ManifestModule: manifest.ManifestModule{
					Manifest: manifestMap,
					BaseDir:  runtime.Arg.BaseDir,
				},
				FailOnNoInstallation:      true,
				SkipNVMLCheckAfterInstall: true,
			},
			&gpu.InstallContainerToolkitModule{
				ManifestModule: manifest.ManifestModule{
					Manifest: manifestMap,
					BaseDir:  runtime.Arg.BaseDir,
				},
				// when nvidia driver is just upgraded
				// nvidia-smi will fail to execute
				SkipCudaCheck: true,
			},
			&gpu.RestartContainerdModule{},
		},
		Runtime: runtime,
	}

	if err := p.Start(); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("The GPU driver has been upgraded, for it to work properly, the machine needs to be rebooted.")
	reader, err := utils.GetBufIOReaderOfTerminalInput()
	if err != nil {
		return nil
	}
	for {
		fmt.Printf("Reboot now? [yes/no]: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("failed to read user input for reboot confirmation: %v", err)
		}
		input = strings.ToLower(strings.TrimSpace(input))
		if input == "" {
			continue
		}
		if strings.HasPrefix("yes", input) {
			output, err := exec.Command("reboot").CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to reboot: %v", err)
			}
			fmt.Println(string(output))
			return nil
		} else if strings.HasPrefix("no", input) {
			return nil
		}
	}

}
