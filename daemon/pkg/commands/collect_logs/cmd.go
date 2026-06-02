package collectlogs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/fanout"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"k8s.io/klog/v2"
)

// nodeLocalPath is the node-local endpoint the orchestrator fans out to.
const nodeLocalPath = "/command/collect-logs-node"

// collectLogs is the master-side orchestrator: it authorizes the request, then
// fans the collection out to every node's node-local endpoint and merges the
// per-node results into a single archive in the caller's Home.
type collectLogs struct {
	commands.Operation
	*commands.BaseCommand
}

var _ commands.Interface = &collectLogs{}

func New() commands.Interface {
	return &collectLogs{
		Operation: commands.Operation{
			Name: commands.CollectLogs,
		},
		BaseCommand: commands.NewBaseCommand(),
	}
}

func (i *collectLogs) Execute(ctx context.Context, p any) (res any, err error) {
	param, ok := p.(*Param)
	if !ok || param == nil {
		return nil, fmt.Errorf("collect-logs: invalid param type %T", p)
	}

	kubeClient, err := utils.GetKubeClient()
	if err != nil {
		return nil, fmt.Errorf("create k8s client error: %w", err)
	}
	dynamicClient, err := utils.GetDynamicClient()
	if err != nil {
		return nil, fmt.Errorf("create dynamic client error: %w", err)
	}

	// Authorize synchronously so scope violations are reported to the caller
	// (403) instead of being swallowed by the async orchestration.
	rs, err := authorize(ctx, kubeClient, dynamicClient, param)
	if err != nil {
		return nil, err
	}

	// Results land in the requesting user's own Home (JuiceFS-backed, visible
	// from every node), so any role can retrieve their result from Files.
	home, err := utils.GetUserspacePvcHostPath(ctx, rs.username, kubeClient)
	if err != nil {
		return nil, fmt.Errorf("get caller home path error: %w", err)
	}

	runID := newRunID()
	signature := param.CallerSignature
	// Forward the original (un-expanded) user scope; each node re-authorizes
	// under the same identity and collects only its local portion.
	nodeReqTemplate := *param

	state.CurrentState.CollectingLogsState = state.InProgress
	state.CurrentState.CollectingLogsError = ""

	go func() {
		startedAt := time.Now()
		var (
			errStr string
			ferr   error
		)
		defer func() {
			if ferr != nil {
				klog.Error(errStr)
				state.CurrentState.CollectingLogsState = state.Failed
				state.CurrentState.CollectingLogsError = errStr
			}
		}()

		bgCtx := context.Background()
		nodes, e := fanout.ListReadyNodes(bgCtx)
		if e != nil {
			ferr, errStr = e, fmt.Sprintf("list nodes error, %v", e)
			return
		}
		if len(nodes) == 0 {
			ferr = errors.New("no ready nodes found")
			errStr = ferr.Error()
			return
		}

		dispatcher := &fanout.Dispatcher{PeerPath: nodeLocalPath, Signature: signature}
		results := dispatcher.Run(bgCtx, nodes, func(t fanout.NodeTarget) any {
			req := &NodeRequest{Param: nodeReqTemplate, RunID: runID}
			return req
		})

		archivePath := finalArchivePath(home, runID)
		runDir := stagingRunDir(home, runID)
		if e := buildArchive(archivePath, runDir, runID, rs.username, startedAt, results); e != nil {
			ferr, errStr = e, fmt.Sprintf("build archive error, %v", e)
			return
		}
		if e := os.Chown(archivePath, 1000, 1000); e != nil {
			klog.Warningf("change archive owner error, %v", e)
		}
		if e := os.RemoveAll(runDir); e != nil {
			klog.Warningf("cleanup staging dir error, %v", e)
		}

		var okCount int
		var failed []string
		for _, r := range results {
			if r.Status == fanout.StatusOK {
				okCount++
			} else {
				failed = append(failed, fmt.Sprintf("%s %s", r.Node.Name, r.Status))
			}
		}
		if okCount == 0 {
			ferr = errors.New("all nodes failed to collect logs")
			errStr = ferr.Error()
			return
		}

		// Completed even with partial failures; see collect-report.json for
		// the authoritative per-node status.
		state.CurrentState.CollectingLogsState = state.Completed
		if len(failed) > 0 {
			state.CurrentState.CollectingLogsError = fmt.Sprintf("partial: %v", failed)
		} else {
			state.CurrentState.CollectingLogsError = ""
		}
		klog.Infof("collect log completed, archive=%s, ok=%d/%d", archivePath, okCount, len(results))
	}()

	return nil, nil
}
