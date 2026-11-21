package upgrade

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/daemon/pkg/cluster/clusterop"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"k8s.io/klog/v2"
)

type olaresdStage func(ctx context.Context, req clusterop.UpgradeStageRequest) error

// olaresdStages are the stages before notify-start, which olaresd carries out
// itself. Each only fetches, checks or loads, and each is safe to repeat: the
// orchestrator dispatches a stage again after the olaresd running it restarts.
var olaresdStages = map[string]olaresdStage{
	clusterop.StageBootstrap: func(ctx context.Context, req clusterop.UpgradeStageRequest) error {
		return Bootstrap(ctx, bootstrapRequestOf(req))
	},
	clusterop.StageDownload:        phaseStage(NewDownloadWizard, NewDownloadSpaceCheck, NewDownloadComponent),
	clusterop.StagePrecheckNode:    phaseStage(NewImageImportSpaceCheck),
	clusterop.StagePrecheckCluster: phaseStage(NewClusterPreCheck),
	clusterop.StageImportImages:    phaseStage(NewImportImages),
}

// RunOlaresdStage runs req if it is a stage olaresd implements, and reports
// false for a stage that belongs to olares-cli.
func RunOlaresdStage(ctx context.Context, req clusterop.UpgradeStageRequest) (bool, error) {
	run, ok := olaresdStages[req.Stage]
	if !ok {
		return false, nil
	}
	return true, run(ctx, req)
}

func phaseStage(phases ...func() commands.Interface) olaresdStage {
	return func(ctx context.Context, req clusterop.UpgradeStageRequest) error {
		v, err := semver.NewVersion(req.Version)
		if err != nil {
			return fmt.Errorf("invalid target version %q: %v", req.Version, err)
		}
		target := state.UpgradeTarget{Version: *v, CliURL: req.CliURL, WizardURL: req.WizardURL, CDN: req.CDN}
		for _, newCMD := range phases {
			cmd := newCMD()
			klog.Infof("upgrade stage %s for %s: %s", req.Stage, req.Version, cmd.OperationName())
			if err := stagePhase(ctx, cmd, target); err != nil {
				return fmt.Errorf("%s: %w", cmd.OperationName(), err)
			}
		}
		return nil
	}
}

var stagePhase = runPhaseToCompletion

func appendReleaseCDN(args []string, cdn string) []string {
	if cdn == "" {
		cdn = commands.OLARES_CDN_SERVICE
	}
	if cdn != "" {
		return append(args, "--cdn-service", cdn)
	}
	return args
}

// runPhaseToCompletion executes one phase and waits for it, whether it reports
// completion straight away or through a progress channel.
func runPhaseToCompletion(ctx context.Context, cmd commands.Interface, target state.UpgradeTarget) error {
	res, err := cmd.Execute(ctx, target)
	if err != nil {
		return err
	}
	execution, ok := res.(ExecutionRes)
	if !ok {
		return fmt.Errorf("unexpected result type")
	}
	return AwaitExecution(ctx, execution, nil)
}
