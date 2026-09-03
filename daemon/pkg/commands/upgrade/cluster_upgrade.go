package upgrade

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/clusterop"
	"github.com/beclab/Olares/daemon/pkg/cluster/clusterstatus"
	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"k8s.io/klog/v2"
)

// clusterUpgradePoll is how often the watcher re-reads the operation it is
// following. The operation itself is driven by the orchestrator; this only
// moves the progress number the UI shows.
const clusterUpgradePoll = 5 * time.Second

// upgradeOrchestrator reports whether this upgrade has to be scheduled across
// other nodes, and returns the orchestrator to do it with.
//
// The order of the two questions is the point. Whether there are other nodes
// is asked first and from the node directory, and only then is an orchestrator
// looked for — because "there is no orchestrator" means two entirely different
// things depending on the answer. On a single machine it means nothing at all:
// there is nothing to schedule and the upgrade runs here, as it always has. On
// a cluster it means this daemon cannot upgrade the other nodes, and going
// ahead anyway would upgrade the control node alone and leave the rest of the
// cluster on the old version — the exact outcome the whole orchestrator exists
// to prevent, arrived at silently.
//
// For the same reason a node directory that cannot be read is an error rather
// than an assumption. Not knowing how many nodes there are is not evidence
// that there is one.
func upgradeOrchestrator(ctx context.Context) (*clusterop.Manager, bool, error) {
	nodes, err := inventory.List(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("cannot tell whether this Olares has other nodes to upgrade: %v", err)
	}
	return orchestratorFor(nodes, clusterop.Current())
}

// orchestratorFor is the decision on its own, so it can be checked without a
// cluster to read a directory from.
func orchestratorFor(nodes []inventory.Node, m *clusterop.Manager) (*clusterop.Manager, bool, error) {
	var self bool
	var others int
	for _, n := range nodes {
		if n.IsSelf && n.Role == inventory.RoleMaster {
			self = true
			continue
		}
		others++
	}
	// The only machine there is: this upgrade is its own, as it always has
	// been.
	if others == 0 {
		return nil, false, nil
	}

	// There are other nodes and the directory does not point at this machine
	// as the one that would schedule them. Refused for the same reason a
	// missing orchestrator is refused below, and it took reading this
	// function's own first paragraph to notice the two were not treated
	// alike: falling back to the local path here upgrades this machine and
	// leaves the rest of the cluster behind, which is the outcome being
	// guarded against, not a special case of a single node.
	if !self {
		return nil, false, fmt.Errorf(
			"this Olares has %d other node(s) but the node directory does not identify this machine"+
				" as the control node, so their upgrade cannot be scheduled from here", others)
	}

	if m == nil {
		return nil, false, fmt.Errorf(
			"this Olares has %d other node(s) but this daemon has no orchestrator to upgrade them with", others)
	}
	return m, true, nil
}

// upgradeAcrossCluster hands the upgrade to the orchestrator and follows it.
//
// The owner is not asked again. They already authorized this upgrade when they
// created the upgrade target through the signed route, and this is the same
// upgrade being carried out — inventing a second confirmation here would mean
// the watcher had to hold a signature, which is exactly what an hour-long,
// restart-crossing operation cannot do.
//
// The request id is derived from the target version rather than generated, so
// the watcher's own retries and restarts rejoin the operation they started
// instead of being refused as a second concurrent upgrade.
func (i *upgrade) upgradeAcrossCluster(ctx context.Context, m *clusterop.Manager,
	target state.UpgradeTarget) (ExecutionRes, error) {
	owner, err := upgradeOwner()
	if err != nil {
		return nil, err
	}
	clusterID, err := currentClusterID(ctx)
	if err != nil {
		return nil, err
	}

	op, err := m.Create(ctx, clusterop.CreateRequest{
		Type:      clusterop.TypeUpgrade,
		RequestID: upgradeRequestID(target.Version.Original()),
		Scope:     clusterop.ScopeCluster,
		ClusterID: clusterID,
		Owner:     owner,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start the cluster upgrade: %v", err)
	}
	klog.Infof("upgrading this cluster to %s as operation %s", target.Version.Original(), op.ID)

	// The one thing the flat progress fields cannot carry: where to read the
	// per-stage, per-node detail. See state.State.UpgradingOperationID.
	state.CurrentState.UpgradingOperationID = op.ID

	if op.Status.Terminal() {
		return newExecutionRes(true, nil), terminalUpgradeError(op)
	}

	progressChan := make(chan int, 100)
	go i.followClusterUpgrade(ctx, m, op.ID, progressChan)
	return newExecutionRes(false, progressChan), nil
}

// upgradeRequestID is the operation an upgrade to this version belongs to.
//
// It is derived rather than generated so that the watcher's retries and
// restarts rejoin the run they started, and so that anything else that needs
// to find that run — cancelling it, following it from the UI — can name it
// without having been told an id.
func upgradeRequestID(version string) string { return "olares-upgrade-" + version }

// StopClusterUpgrade asks the orchestrated upgrade of this version to stop.
//
// Removing the upgrade target is what stops the watcher, but on a cluster the
// watcher is not the thing doing the work: the orchestrator is, it outlives
// any one phase, and left alone it would carry on dispatching stages to nodes
// after the operator had cancelled. This is the other half of that cancel.
//
// It is best-effort by design. A node has no orchestrator, an upgrade that
// already settled has nothing to stop, and neither is a reason to refuse to
// remove the target.
func StopClusterUpgrade(version string) {
	m := clusterop.Current()
	if m == nil || strings.TrimSpace(version) == "" {
		return
	}
	op, ok := m.GetByRequest(upgradeRequestID(version))
	if !ok {
		return
	}
	if m.RequestStop(op.ID) {
		klog.Infof("cluster upgrade %s will stop after the stage it is running", op.ID)
	}
}

// StopActiveClusterUpgrade asks the in-flight cluster upgrade to stop when
// the caller no longer has the version that names it.
//
// The signed cancel reads the target, stops that version's run, then deletes
// the file. Anyone who only deletes the file — a script, a crashed writer, the
// watcher seeing the file vanish on its next tick — leaves the orchestrator
// going, and a daemon that restarts afterwards resumes it. The watcher already
// treats a missing target as "not upgrading"; this is the other half of that.
func StopActiveClusterUpgrade() {
	m := clusterop.Current()
	if m == nil {
		return
	}
	op, ok := m.ActiveOfType(clusterop.TypeUpgrade)
	if !ok {
		return
	}
	if m.RequestStop(op.ID) {
		klog.Infof("cluster upgrade %s will stop after the stage it is running: the upgrade target is gone", op.ID)
	}
}

// ClusterOutcome is what became of the orchestrated upgrade to some
// version: enough for the watcher to tell an upgrade that is over from one that
// merely flipped the version CR on its way through, without the watcher having
// to know what an operation record is.
type ClusterOutcome struct {
	// ID names the record to read the per-stage, per-node detail from.
	ID string
	// Settled is whether the orchestrator is finished with it, either way.
	Settled bool
	// Succeeded is whether every stage ran on every node it was meant to.
	// Only meaningful once Settled.
	Succeeded bool
	// Error is the reviewed sentence the record settled on, empty on success.
	Error string
}

// ClusterUpgradeOutcome reports what became of the orchestrated upgrade to this
// version, and whether there was one at all. A single node has no orchestrator
// and no record, which is not a failure — it is the other flow, and the second
// return is how a caller tells the two apart.
func ClusterUpgradeOutcome(version string) (ClusterOutcome, bool) {
	m := clusterop.Current()
	if m == nil || strings.TrimSpace(version) == "" {
		return ClusterOutcome{}, false
	}
	op, ok := m.GetByRequest(upgradeRequestID(version))
	if !ok {
		return ClusterOutcome{}, false
	}
	return ClusterOutcome{
		ID:        op.ID,
		Settled:   op.Status.Terminal(),
		Succeeded: op.Status == clusterop.StatusSucceeded,
		Error:     op.Error,
	}, true
}

// followClusterUpgrade turns the operation record into the progress number the
// rest of the upgrade watcher already reports.
func (i *upgrade) followClusterUpgrade(ctx context.Context, m *clusterop.Manager,
	id string, progressChan chan<- int) {
	defer close(progressChan)

	ticker := time.NewTicker(clusterUpgradePoll)
	defer ticker.Stop()

	var last int
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		op, ok := m.Get(id)
		if !ok {
			return
		}
		if p := upgradeProgressOf(op); p > last {
			last = p
			progressChan <- p
		}
		if op.Status.Terminal() {
			if err := terminalUpgradeError(op); err != nil {
				klog.Errorf("cluster upgrade %s: %v", id, err)
				return
			}
			if last < commands.ProgressNumFinished {
				progressChan <- commands.ProgressNumFinished
			}
			return
		}
	}
}

// upgradeProgressOf is the share of the operation's steps that have settled.
// It is deliberately coarse: the steps are the only thing the control node
// actually knows, and inventing a finer number from the log of whichever node
// happens to be running would describe one machine as if it were the cluster.
func upgradeProgressOf(op clusterop.Operation) int {
	if len(op.Steps) == 0 {
		return 0
	}
	var done int
	for _, s := range op.Steps {
		switch s.Status {
		case clusterop.StepSucceeded, clusterop.StepSkipped:
			done++
		}
	}
	// Capped below completion: only a terminal operation reports finished,
	// because the last step succeeding is not the same as the upgrade having
	// succeeded until the orchestrator says so.
	if p := done * 99 / len(op.Steps); p < commands.ProgressNumFinished {
		return p
	}
	return commands.ProgressNumFinished - 1
}

func terminalUpgradeError(op clusterop.Operation) error {
	if op.Status == clusterop.StatusSucceeded {
		return nil
	}
	if op.Error != "" {
		return fmt.Errorf("cluster upgrade %s: %s", op.Status, op.Error)
	}
	return fmt.Errorf("cluster upgrade %s", op.Status)
}

// upgradeOwner names the Olares this upgrade belongs to. It is the same
// identity the power operations record, resolved the same way.
func upgradeOwner() (string, error) {
	if name, err := utils.GetOlaresNameFromReleaseFile(); err == nil {
		if name = strings.TrimSpace(name); name != "" {
			return name, nil
		}
	}
	st, _ := state.Snapshot()
	if st.TerminusName != nil {
		if name := strings.TrimSpace(*st.TerminusName); name != "" {
			return name, nil
		}
	}
	return "", fmt.Errorf("cannot determine the owner of this Olares")
}

func currentClusterID(ctx context.Context) (string, error) {
	client, err := utils.GetKubeClient()
	if err != nil {
		return "", err
	}
	return clusterstatus.ClusterID(ctx, client)
}
