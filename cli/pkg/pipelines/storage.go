package pipelines

import (
	"fmt"
	"path"

	"github.com/beclab/Olares/cli/cmd/ctl/options"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/phase"
	"github.com/beclab/Olares/cli/pkg/phase/system"
	"github.com/pkg/errors"
)

func CliInstallStoragePipeline(opts *options.InstallStorageOptions) error {
	var terminusVersion, _ = phase.GetOlaresVersion()
	if terminusVersion != "" {
		return errors.New("Olares is already installed, please uninstall it first.")
	}

	arg := common.NewArgument()
	arg.SetBaseDir(opts.BaseDir)
	arg.SetOlaresVersion(opts.Version)
	arg.SetStorage(getStorageValueFromEnv())

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return fmt.Errorf("error creating runtime: %v", err)
	}

	manifest := path.Join(runtime.GetInstallerDir(), "installation.manifest")
	runtime.Arg.SetManifest(manifest)

	return system.InstallStoragePipeline(runtime).Start()
}
