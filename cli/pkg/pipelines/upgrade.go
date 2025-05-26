package pipelines

import (
	"bytetrade.io/web3os/installer/pkg/upgrade"
	"bytetrade.io/web3os/installer/pkg/utils"
	"fmt"
	"os"
	"path"

	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/phase"
	"github.com/pkg/errors"
)

func UpgradeOlaresPipeline(opts *options.UpgradeOptions) error {
	currentVersionString, err := phase.GetOlaresVersion()
	if err != nil {
		return errors.Wrap(err, "failed to get current Olares version")
	}
	if currentVersionString == "" {
		return errors.New("Olares is not installed, please install it first")
	}
	currentVersion, err := utils.ParseOlaresVersionString(currentVersionString)
	if err != nil {
		return fmt.Errorf("error parsing current Olares version: %v", err)
	}

	// validate the expected version is non-empty before the NewArgument() call
	// as it will fall back to load the current olares release
	if opts.Version == "" {
		return errors.New("target version is required")
	}
	targetVersion, err := utils.ParseOlaresVersionString(opts.Version)
	if err != nil {
		return fmt.Errorf("error parsing target Olares version: %v", err)
	}

	if !targetVersion.GreaterThan(currentVersion) {
		fmt.Printf("current version is: %s, no need to upgrade to %s\n", currentVersion.String(), opts.Version)
		os.Exit(0)
	}

	arg := common.NewArgument()
	arg.SetBaseDir(opts.BaseDir)
	arg.SetOlaresVersion(opts.Version)
	arg.SetConsoleLog("upgrade.log", true)
	arg.SetKubeVersion(phase.GetKubeType())

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return fmt.Errorf("error creating runtime: %v", err)
	}

	manifest := path.Join(runtime.GetInstallerDir(), "installation.manifest")
	runtime.Arg.SetManifest(manifest)

	upgradeModule := &upgrade.UpgradeModule{
		CurrentVersion: currentVersion,
		TargetVersion:  targetVersion,
	}

	p := &pipeline.Pipeline{
		Name:    "UpgradeOlares",
		Modules: []module.Module{upgradeModule},
		Runtime: runtime,
	}

	logger.Infof("Starting Olares upgrade from %s to %s...", currentVersion, opts.Version)
	if err := p.Start(); err != nil {
		return errors.Wrap(err, "upgrade failed")
	}

	logger.Info("Olares upgrade completed successfully!")
	return nil
}

func UpgradePreCheckPipeline() error {
	var arg = common.NewArgument()
	arg.SetConsoleLog("upgrade-precheck.log", true)

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	p := &pipeline.Pipeline{
		Name: "UpgradePreCheck",
		Modules: []module.Module{
			&upgrade.PrecheckModule{},
		},
		Runtime: runtime,
	}
	return p.Start()

}
