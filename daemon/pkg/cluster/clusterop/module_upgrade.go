package clusterop

import (
	"context"

	"fmt"

	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
	"k8s.io/klog/v2"
)

// TypeUpgrade moves every node to a new Olares version. It is declared here
// because this module is what answers for it.
const TypeUpgrade Type = "upgrade"

func init() {
	MustRegisterModule(upgradeModule{})
	// Starting an upgrade is the owner's decision, held to the same standard
	// as powering the cluster off. The route an upgrade normally arrives
	// through asks for the signature itself, but the generic create route
	// asks this registry, and a type absent from it is admitted on an access
	// token alone.
	MustRequireSignature(TypeUpgrade)
}

// upgradeModule is the cluster upgrade: every compute node is brought to the
// target release, then the plan that release describes is run stage by stage,
// compute nodes before the control node, with a barrier at every stage.
//
// It shares this package's record, store and single-operation lock with the
// power operations, which is the point: an upgrade must not begin while the
// cluster is rebooting, and a reboot must not begin half way through an
// upgrade. Two orchestrators would have to be taught that; one cannot forget.
type upgradeModule struct{}

// It deliberately does not implement NodeOperationModule, and the reasons are
// about the shape of the work rather than about the generic hop being
// unsuitable in general. Three of them, each enough on its own:
//
// The credential. Both ways through /command/cluster-operation end at a user
// credential — the owner's signature for a type that registered a requirement,
// an access token the master forwards otherwise. An upgrade has neither by the
// time it reaches a node: the run is detached from the request that authorized
// it, outlives any signature's permitted lifetime, and is resumed from disk by
// a daemon that was not running when the owner asked. What it carries instead
// is a per-operation secret held in the cluster; see UpgradeDeps.Auth.
//
// The shape. ExecuteNode is a function that returns, and the hop is one
// request answered once. A stage runs for minutes and is expected not to
// return at all on the node it is replacing olaresd on, so there is nothing to
// answer with and nobody left to answer.
//
// The record. Because a stage outlives the process running it, how it is going
// has to be on disk and readable afterwards by whichever daemon comes back.
// The generic hop has no status surface, and adding one would mean giving
// every node operation a lifecycle only this one needs.
//
// So an upgrade's node half has its own route, its own credential and its own
// per-stage record; see UpgradeStagePath and UpgradeStageRunner. What it does share is
// everything above the node: the module registry, the operation record, the
// single-operation lock and the recovery path.
var (
	_ OperationModule = upgradeModule{}
	_ ResumableModule = upgradeModule{}
	_ RetryableModule = upgradeModule{}
)

func (upgradeModule) Type() Type { return TypeUpgrade }

// Validate refuses anything but a whole-cluster upgrade. There is no such
// thing as upgrading one node: the plan's stages are defined over the cluster,
// and a node left on another version is the state an upgrade exists to end.
func (upgradeModule) Validate(req CreateRequest) error {
	switch req.Scope {
	case "", ScopeCluster:
		if req.Target != "" {
			return fmt.Errorf("an upgrade cannot name a target node")
		}
		return nil
	default:
		return fmt.Errorf("an upgrade is always %s-wide, not %q", ScopeCluster, req.Scope)
	}
}

// Phase is maintenance rather than a phase of its own: an upgrade is a period
// during which the cluster is being worked on and parts of it come and go,
// which is what the phase already means.
func (upgradeModule) Phase(Operation) (nodestatus.Phase, bool) {
	return nodestatus.PhaseMaintenance, true
}

func (upgradeModule) Run(ctx context.Context, rt Runtime, _ RunRequest) Outcome {
	m, id, ok := managerOf(rt)
	if !ok {
		// Only the manager's own Runtime carries the seams an upgrade is made
		// of. Anything else cannot reach a node, and says so rather than
		// reporting an upgrade it never performed.
		return Outcome{
			Status: StatusFailed,
			Code:   CodeUnsupportedOperation,
			Error:  "this runtime cannot carry out an upgrade",
		}
	}
	// The one place the upgrade seams are checked. Everything below this
	// point uses them without asking again, which is what keeps a missing
	// dependency from turning some later check into a silent no-op.
	if !m.deps.Upgrade.complete() {
		return Outcome{
			Status: StatusFailed,
			Code:   CodeUpgradeUnsupported,
			Error:  "this daemon was not built to orchestrate upgrades",
		}
	}
	// The credentials are deliberately unused: see UpgradeDeps.Auth for why an
	// upgrade cannot be carried by the signature that authorized it, and why
	// that is what lets a resumed run continue without one.
	return m.runUpgrade(ctx, id)
}

// RetryAfterFailure lets a failed upgrade be attempted again.
//
// An upgrade's request id is derived from the target version rather than
// generated, so that the watcher's retries and restarts rejoin the run they
// started instead of piling up concurrent upgrades. The cost of that is this:
// without an explicit retry the same id would be answered by the same failure
// for as long as the record exists, and the only way back would be deleting it
// off disk. Retrying is also the normal thing to want — most upgrade failures
// are a node that was busy, an image that would not pull, a service slow to
// come up — and every stage is written to be reentrant precisely so that
// running one again is safe.
//
// A cancelled upgrade is the exception, and the only one: it did not fail, it
// was stopped, and starting it again is the opposite of what was asked for.
// Removing the upgrade target normally stops the watcher as well, so the two
// halves of a cancel agree — but only while both happen. This is the half that
// does not depend on the other one having worked.
func (upgradeModule) RetryAfterFailure(op Operation) bool {
	return op.Code != CodeUpgradeCancelled
}

// ResumeInterrupted continues any upgrade that was still moving.
//
// An upgrade restarts olaresd on the node orchestrating it — installing the
// new daemon is one of the stages — so an interrupted record is the normal
// middle of a working upgrade rather than evidence that something went wrong.
// What makes continuing safe is that the record says which stages finished,
// and a resumed run walks past them; see Manager.stageSettled.
func (upgradeModule) ResumeInterrupted(Operation) bool { return true }

// Recover runs the rest of an upgrade that this daemon's predecessor started.
//
// It is the same run, not a repair of one: runUpgrade re-reads the plan and
// skips every stage already recorded as done, so what happens here is
// whatever had not happened yet.
func (upgradeModule) Recover(ctx context.Context, rt Runtime, op Operation) {
	m, id, ok := managerOf(rt)
	if !ok {
		return
	}
	if !m.deps.Upgrade.complete() {
		// Settled here rather than left alone. Declaring ResumeInterrupted
		// took this record off the path that would otherwise have failed it —
		// see recoverLoadedOperations, which skips the interrupted fallback
		// for exactly these — so returning without settling would leave it
		// running for ever, holding the cluster's single-operation lock and
		// blocking every later operation including a reboot.
		klog.Errorf("clusterop: cannot resume upgrade %s: this daemon does not orchestrate upgrades", id)
		if err := rt.Complete(failedWith(CodeUpgradeUnsupported,
			"this daemon cannot continue the upgrade it restarted in the middle of")); err != nil {
			klog.Errorf("clusterop: settle the unresumable upgrade %s: %v", id, err)
		}
		return
	}
	klog.Infof("clusterop: resuming upgrade %s after this daemon restarted", id)

	outcome := m.runUpgrade(ctx, id)
	if err := rt.Complete(outcome); err != nil {
		klog.Warningf("clusterop: settle the resumed upgrade %s: %v", op.ID, err)
	}
}
