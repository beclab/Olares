package pipelines

import (
	"context"
	"fmt"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
	"github.com/beclab/Olares/cli/pkg/terminus"
)

func MasterInfoPipeline() error {
	arg := common.NewArgument()
	arg.SetConsoleLog("masterinfo.log", true)
	// TODO(ctx): plumb ctx in a follow-up; this entry point is not yet ctx-aware.
	_, err := probeMasterInfo(context.Background(), arg, false)
	return err
}

// probeMasterInfo connects to the master over SSH, checks that this node may
// join it, and reports what it found.
//
// It is the single implementation used by both `node masterinfo` and the worker
// join flow, so the two can never disagree about how the master's version,
// Kubernetes type, JuiceFS state or node names are determined, nor about which
// of those make a node ineligible.
//
// See terminus.CheckJoinEligibility for what bootstrapping changes.
func probeMasterInfo(ctx context.Context, arg *common.Argument, bootstrapping bool) (*terminus.MasterInfo, error) {
	if !arg.SystemInfo.IsLinux() {
		return nil, fmt.Errorf("only Linux nodes can be added to an Olares cluster")
	}
	if err := arg.MasterHostConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid master host config: %w", err)
	}

	runtime, err := common.NewKubeRuntime(*arg)
	if err != nil {
		return nil, fmt.Errorf("error creating runtime: %v", err)
	}

	var info terminus.MasterInfo
	p := &pipeline.Pipeline{
		Name: "Get Master Info",
		Modules: []module.Module{&terminus.GetMasterInfoModule{
			Print:         true,
			Out:           &info,
			Bootstrapping: bootstrapping,
		}},
		Runtime: runtime,
	}
	if err := p.Start(ctx); err != nil {
		return nil, err
	}
	return &info, nil
}
