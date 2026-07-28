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

// olaresUID/olaresGID own user-facing files under a user's Home; the daemon
// runs as root, so results it writes must be chowned back to the user.
const (
	olaresUID = 1000
	olaresGID = 1000
)

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

	// Authorize synchronously so scope violations are reported to the caller
	// (403) instead of being swallowed by the async orchestration.
	rs, err := authorize(ctx, kubeClient, param)
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
	authToken := param.CallerToken
	// Forward the original (un-expanded) user scope; each node re-authorizes
	// under the same identity and collects only its local portion.
	nodeReqTemplate := *param

	startedAt := time.Now()
	relPath := finalArchiveRelPath(runID)
	fileName := archiveFileName(runID)

	// Register the per-runID task before returning so an immediate status poll
	// finds it. Runs are isolated by runID (staging + archive paths), so any
	// number can proceed concurrently.
	registerTask(&TaskStatus{
		RunID:     runID,
		Caller:    rs.username,
		State:     string(state.InProgress),
		File:      fileName,
		Path:      relPath,
		StartedAt: startedAt,
	})

	// Global singleton kept for backward compatibility; under concurrency it
	// reflects the most recent transition, while the per-runID task is
	// authoritative.
	state.CurrentState.CollectingLogsState = state.InProgress
	state.CurrentState.CollectingLogsError = ""

	go func() {
		var (
			errStr  string
			ferr    error
			results []fanout.NodeResult
			partial string
			ok      bool
		)
		defer func() {
			finishedAt := time.Now()
			nodes := nodeSummaries(results)
			switch {
			case ferr != nil:
				klog.Error(errStr)
				state.CurrentState.CollectingLogsState = state.Failed
				state.CurrentState.CollectingLogsError = errStr
				updateTask(runID, func(t *TaskStatus) {
					t.State = string(state.Failed)
					t.Error = errStr
					t.FinishedAt = &finishedAt
					t.Nodes = nodes
				})
			case ok:
				state.CurrentState.CollectingLogsState = state.Completed
				state.CurrentState.CollectingLogsError = partial
				updateTask(runID, func(t *TaskStatus) {
					t.State = string(state.Completed)
					t.Error = partial
					t.FinishedAt = &finishedAt
					t.Nodes = nodes
				})
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

		dispatcher := &fanout.Dispatcher{PeerPath: nodeLocalPath, AuthToken: authToken}
		results = dispatcher.Run(bgCtx, nodes, func(t fanout.NodeTarget) any {
			req := &NodeRequest{Param: nodeReqTemplate, RunID: runID}
			return req
		})

		archivePath := finalArchivePath(home, runID)
		runDir := stagingRunDir(home, runID)
		if e := buildArchive(archivePath, runDir, runID, rs.username, startedAt, results); e != nil {
			ferr, errStr = e, fmt.Sprintf("build archive error, %v", e)
			return
		}
		if e := os.Chown(archivePath, olaresUID, olaresGID); e != nil {
			klog.Warningf("change archive owner error, %v", e)
		}
		// The pod_logs dir is created by root via MkdirAll; chown it so the
		// user can manage their archives in Files.
		if e := os.Chown(podLogsDir(home), olaresUID, olaresGID); e != nil {
			klog.Warningf("change pod_logs dir owner error, %v", e)
		}
		if e := os.RemoveAll(runDir); e != nil {
			klog.Warningf("cleanup staging dir error, %v", e)
		}
		pruneOldArchives(podLogsDir(home), archiveRetention)

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
		if len(failed) > 0 {
			partial = fmt.Sprintf("partial: %v", failed)
		}
		ok = true
		klog.Infof("collect log completed, archive=%s, ok=%d/%d", archivePath, okCount, len(results))
	}()

	return &CollectResult{
		RunID: runID,
		File:  fileName,
		Path:  relPath,
	}, nil
}

// nodeSummaries condenses fan-out results into the per-node view stored on a
// task.
func nodeSummaries(results []fanout.NodeResult) []TaskNodeStatus {
	if len(results) == 0 {
		return nil
	}
	out := make([]TaskNodeStatus, 0, len(results))
	for _, r := range results {
		out = append(out, TaskNodeStatus{
			Name:   r.Node.Name,
			Status: string(r.Status),
			Err:    r.Err,
		})
	}
	return out
}
