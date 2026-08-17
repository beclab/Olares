package pipelines

import (
	"context"
	"fmt"
	"path"

	"github.com/beclab/Olares/cli/pkg/upgrade"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/beclab/Olares/cli/version"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
	"github.com/beclab/Olares/cli/pkg/phase"
	"github.com/pkg/errors"
)

func UpgradeOlaresPipeline(ctx context.Context) error {
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

	runtime, err := upgradeRuntime()
	if err != nil {
		return err
	}

	p := &pipeline.Pipeline{
		Name:    "UpgradeOlares",
		Modules: []module.Module{&upgrade.Module{TargetVersion: targetVersion}},
		Runtime: runtime,
	}

	logger.Infof("Starting Olares upgrade from %s to %s...", currentVersion, targetVersion)
	if err := p.Start(ctx); err != nil {
		return errors.Wrap(err, "upgrade failed")
	}

	logger.Info("Olares upgrade completed successfully!")
	return nil
}

// UpgradeOlaresStagePipeline runs one stage of the upgrade flow on this node.
//
// It deliberately does not repeat upgrade.Check. Viability is a property of
// the upgrade as a whole, decided once by the orchestrator before any node was
// touched; asking again here would fail every stage that runs after the
// control node flipped the version, because by then this node is already
// "current" and Check reports there is nothing to upgrade.
func UpgradeOlaresStagePipeline(ctx context.Context, stage string) error {
	if stage == "" {
		return errors.New("no upgrade stage to run")
	}
	targetVersion, err := utils.ParseOlaresVersionString(version.VERSION)
	if err != nil {
		return fmt.Errorf("error parsing target Olares version: %v", err)
	}

	runtime, err := upgradeRuntime()
	if err != nil {
		return err
	}

	p := &pipeline.Pipeline{
		Name: "UpgradeOlares/" + stage,
		Modules: []module.Module{&upgrade.Module{
			TargetVersion: targetVersion,
			Stage:         stage,
		}},
		Runtime: runtime,
	}

	logger.Infof("Starting upgrade stage %s (target %s) on this node...", stage, targetVersion)
	if err := p.Start(ctx); err != nil {
		return errors.Wrapf(err, "upgrade stage %s failed", stage)
	}

	logger.Infof("Upgrade stage %s completed on this node", stage)
	return nil
}

// upgradeRuntime is the environment every upgrade run needs, whole or by
// stage. Per-stage runs share one upgrade.log so that whatever tails it sees the
// upgrade rather than one stage of it.
func upgradeRuntime() (*common.KubeRuntime, error) {
	arg := common.NewArgument()
	arg.SetOlaresVersion(version.VERSION)
	arg.SetConsoleLog("upgrade.log", true)
	arg.SetKubeVersion(phase.GetKubeType())

	runtime, err := common.NewKubeRuntime(*arg)
	if err != nil {
		return nil, fmt.Errorf("error creating runtime: %v", err)
	}
	runtime.Arg.SetManifest(path.Join(runtime.GetInstallerDir(), "installation.manifest"))
	return runtime, nil
}

func UpgradePreCheckPipeline(ctx context.Context) error {
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
	return p.Start(ctx)

}
