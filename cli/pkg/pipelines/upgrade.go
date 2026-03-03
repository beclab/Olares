package pipelines

import (
	"fmt"
	"path"

	"github.com/beclab/Olares/cli/pkg/upgrade"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/beclab/Olares/cli/version"
	"github.com/spf13/viper"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
	"github.com/beclab/Olares/cli/pkg/phase"
	"github.com/pkg/errors"
)

func UpgradeOlaresPipeline() error {
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

	targetVersion, err := utils.ParseOlaresVersionString(version.VERSION)
	if err != nil {
		return fmt.Errorf("error parsing target Olares version: %v", err)
	}

	if err := upgrade.Check(currentVersion, targetVersion); err != nil {
		return err
	}

	arg := common.NewArgument()
	arg.SetOlaresVersion(viper.GetString(common.FlagVersion))
	arg.SetConsoleLog("upgrade.log", true)
	arg.SetKubeVersion(phase.GetKubeType())

	runtime, err := common.NewKubeRuntime(*arg)
	if err != nil {
		return fmt.Errorf("error creating runtime: %v", err)
	}

	manifest := path.Join(runtime.GetInstallerDir(), "installation.manifest")
	runtime.Arg.SetManifest(manifest)

	upgradeModule := &upgrade.Module{
		TargetVersion: targetVersion,
	}

	p := &pipeline.Pipeline{
		Name:    "UpgradeOlares",
		Modules: []module.Module{upgradeModule},
		Runtime: runtime,
	}

	logger.Infof("Starting Olares upgrade from %s to %s...", currentVersion, targetVersion)
	if err := p.Start(); err != nil {
		return errors.Wrap(err, "upgrade failed")
	}

	logger.Info("Olares upgrade completed successfully!")
	return nil
}

func UpgradePreCheckPipeline() error {
	var arg = common.NewArgument()
	arg.SetConsoleLog("upgrade-precheck.log", true)

	runtime, err := common.NewKubeRuntime(*arg)
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
