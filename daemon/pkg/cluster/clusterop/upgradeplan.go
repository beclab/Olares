package clusterop

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
)

// Placement is where a stage's tasks run. The values match
// cli/pkg/upgrade.Placement, which is where they are decided: the daemon
// schedules stages, it does not classify tasks.
//
// It is the whole of the execution framework's input. Nothing above it — the
// stage sequence — knows how a node set is reached, and nothing below it — a
// task — knows which node set it landed on.
//
// It belongs to the upgrade rather than to this package generally, and is
// named for a third concept because the two obvious words are taken and both
// would mislead. Scope is an operation's breadth, cluster or one node, and
// says nothing about which nodes. Fanout is the transport in
// cluster/fanout — imported by this package's own upgradepeer.go — which is
// how a request reaches several nodes, not which several. A power operation
// has no equivalent: it always goes to the compute nodes and then the control
// node, so there is nothing to parameterise and no generic type to lift this
// into.
type Placement string

const (
	PlacementAdmin    Placement = "admin"
	PlacementWorkers  Placement = "workers"
	PlacementAllNodes Placement = "all-nodes"
)

func (f Placement) valid() bool {
	switch f {
	case PlacementAdmin, PlacementWorkers, PlacementAllNodes:
		return true
	default:
		return false
	}
}

func (f Placement) runsOnAdmin() bool   { return f == PlacementAdmin || f == PlacementAllNodes }
func (f Placement) runsOnWorkers() bool { return f == PlacementWorkers || f == PlacementAllNodes }

// Targets resolves a fanout against the cluster, in the order the nodes are
// acted on.
//
// Compute nodes come before the control node for the same reason they do in a
// power operation: the control node runs the orchestrator, and a stage that
// restarts olaresd there ends the process that would have scheduled the rest.
// Putting it last means everything else is already done when it happens.
func (f Placement) Targets(p plan) []inventory.Node {
	var targets []inventory.Node
	if f.runsOnWorkers() {
		targets = append(targets, p.workers...)
	}
	if f.runsOnAdmin() {
		targets = append(targets, p.master)
	}
	return targets
}

// UpgradeReadiness is a node's answer to "can you run a stage of this upgrade".
//
// It is asked over the upgrade's own node-local surface rather than read off
// the general node-status route, because that route requires a user
// credential and an upgrade has none: the owner's signature cannot cover an
// hour of work, so the only thing the orchestrator carries is the
// per-operation token. A readiness probe the orchestrator cannot make is a
// check that always fails, which is what this replaced.
type UpgradeReadiness struct {
	// Supported is whether this node declares it can run upgrade stages. An
	// olaresd from before stages existed serves no such route at all, and the
	// transport error that produces is the same answer.
	Supported bool `json:"supported"`

	// CLIVersion is the olares-cli this node would run a stage with, and it is
	// the whole of what establishes that a stage name means the same work here
	// as on the control node: both sides resolve the name against the plan
	// their own binary derives, and nothing afterwards compares the two.
	//
	// Empty means the node would not say, and is refused like a wrong answer.
	CLIVersion string `json:"cliVersion,omitempty"`
}

// UpgradeStage is one point in the upgrade flow: what it is called, where its
// tasks run, how many nodes at a time, and how long one node may take.
//
// The last two come from the plan rather than from this daemon's settings
// because both are properties of the work, and the work is described by the
// release being installed. A daemon can only hold one number for every stage of
// every version.
type UpgradeStage struct {
	Name      string    `json:"name"`
	Placement Placement `json:"placement"`

	// MaxParallel is how many nodes may be inside this stage at once, zero for
	// no limit. See cli/pkg/upgrade.Stage.MaxParallel.
	MaxParallel int `json:"maxParallel,omitempty"`

	// TimeoutSeconds is how long one node may take, zero for the daemon's
	// default. See Timeouts.Stage.
	TimeoutSeconds int `json:"timeoutSeconds,omitempty"`

	// AwaitRestart says the tasks in this stage may take the machine down, so
	// the stage is not over when olares-cli exits — it is over when every
	// node that went down is back.
	//
	// Without it a rebooting node reports the stage as succeeded a moment
	// before it stops answering, and the upgrade walks on: the next node is
	// told to reboot while the previous one is still coming up, and an
	// upgrade whose last node never returned still ends as succeeded.
	AwaitRestart bool `json:"awaitRestart,omitempty"`

	Desc  string   `json:"desc,omitempty"`
	Tasks []string `json:"tasks"`
}

// Bounded reports a stage that may only be given to some of its nodes at a
// time. It is also what decides whether a failure stops the stage being handed
// to nodes that have not started it: a bounded stage is bounded because running
// it costs the cluster something.
func (h UpgradeStage) Bounded() bool { return h.MaxParallel > 0 }

// Timeout is how long one node may take over this stage, falling back to the
// daemon's own bound when the plan does not say.
func (h UpgradeStage) Timeout(fallback time.Duration) time.Duration {
	if h.TimeoutSeconds <= 0 {
		return fallback
	}
	return time.Duration(h.TimeoutSeconds) * time.Second
}

// Empty reports a stage this version contributes nothing to.
//
// The flow is fixed, so most releases leave several of its stages empty. They
// stay in the plan and are recorded as skipped, which keeps the record a
// description of the whole flow rather than of the subset that happened to
// have work in it — and keeps "did nothing" distinguishable from "did
// something and it worked".
func (h UpgradeStage) Empty() bool { return len(h.Tasks) == 0 }

// UpgradePlan is what olares-cli says an upgrade to a given version consists
// of: the fixed flow, and what this version put in each stage.
//
// The daemon holds no flow of its own and does not interpret the tasks. Which
// stages exist is a property of the release being installed, and the binary
// that implements them is the one that answers.
type UpgradePlan struct {
	Version string         `json:"version"`
	Stages  []UpgradeStage `json:"stages"`
}

// Validate rejects a plan this daemon cannot schedule. It runs before any node
// is touched, because a plan with a stage nobody can execute is better
// discovered while the cluster is still whole.
func (p UpgradePlan) Validate() error {
	if strings.TrimSpace(p.Version) == "" {
		return fmt.Errorf("upgrade plan names no version")
	}
	if len(p.Stages) == 0 {
		return fmt.Errorf("upgrade plan for %s has no stages", p.Version)
	}
	seen := make(map[string]bool, len(p.Stages))
	for _, h := range p.Stages {
		if strings.TrimSpace(h.Name) == "" {
			return fmt.Errorf("upgrade plan for %s has an unnamed stage", p.Version)
		}
		if seen[h.Name] {
			return fmt.Errorf("upgrade plan for %s repeats stage %q", p.Version, h.Name)
		}
		seen[h.Name] = true
		if !h.Placement.valid() {
			return fmt.Errorf("stage %s has unknown fanout %q", h.Name, h.Placement)
		}
		if h.MaxParallel < 0 {
			return fmt.Errorf("stage %s allows %d nodes at a time", h.Name, h.MaxParallel)
		}
		if h.TimeoutSeconds < 0 {
			return fmt.Errorf("stage %s has a negative timeout", h.Name)
		}
	}
	return nil
}

// HasStage reports whether name is a stage of this plan. It is how the
// orchestrator tells the flow's own stages from the steps around them — the
// precheck, the plan read, the readiness check, the node prepare — none of
// which change the cluster.
func (p UpgradePlan) HasStage(name string) bool {
	for _, h := range p.Stages {
		if h.Name == name {
			return true
		}
	}
	return false
}

// StageNodePrepare brings a node's own binaries up to the target version: the
// packages, the new olares-cli, and the new olaresd.
//
// It is deliberately not a stage of the plan. A plan stage is resolved by the
// node's olares-cli against the plan that binary derives, and the whole point
// of this one is that the node does not have that binary yet — it would
// resolve the stage against the previous version's plan. So the one piece of
// work that cannot be expressed as a plan stage is the one that installs the
// thing which understands plans.
//
// In the written flow this is step b, "olaresd 和 olares-cli 进行套娃替换（多节点）",
// and it stays the orchestrator's own: olaresd performs it, not olares-cli.
const StageNodePrepare = "node-prepare"

// NodePrepareStage is the synthetic stage that runs it.
//
// Its fanout is the compute nodes only. The control node already has the
// target version — the upgrade watcher installed the new olares-cli and the
// new olaresd there before this orchestrator existed to run, which is
// precisely why there is an orchestrator on this machine and not the old one.
func NodePrepareStage() UpgradeStage {
	return UpgradeStage{
		Name:      StageNodePrepare,
		Placement: PlacementWorkers,
		// Unbounded: fetching a release is the one part of an upgrade that
		// costs the cluster nothing, and on a large cluster doing it one node
		// at a time is most of the wall clock for no benefit.
		MaxParallel: 0,
		// Generous, because this is the download: a release is gigabytes and
		// the node may be fetching it over the slowest link in the cluster.
		TimeoutSeconds: 90 * 60,
		Desc:           "fetch and install the target release on each compute node",
		Tasks:          []string{"DownloadTargetVersion", "InstallOlaresCLI", "InstallOlaresd", "ImportImages"},
	}
}

// UpgradeDeps are the side effects an upgrade has, and nothing else does.
//
// They are grouped so that "can this daemon orchestrate an upgrade" is one
// question with one answer. See Deps.Upgrade.
type UpgradeDeps struct {
	// Plan reads what an upgrade to the version this node is holding consists
	// of. The daemon asks olares-cli rather than deciding: the plan is a
	// property of the binary that implements the tasks.
	Plan func(ctx context.Context) (UpgradePlan, error)

	// Local runs stages on this machine. See UpgradeStageRunner.
	Local UpgradeStageRunner

	// Auth is the credential a worker checks a stage against.
	//
	// It is deliberately not the owner's signature. That signature is bound
	// to one request with a lifetime measured in minutes — see
	// MaxSignatureLifetime, and the ceiling is there for a reason — while an
	// upgrade runs for an hour and has to survive this daemon restarting in
	// the middle of it, because installing the new olaresd is part of the
	// work. A signature cannot cover either. The owner still authorizes
	// creating the operation; what authorizes the hops is a per-operation
	// secret held in the cluster, which both sides can read for as long as
	// the operation lasts and which a resumed run can pick up again.
	Auth func(ctx context.Context, operationID string) (string, error)

	// Start and Status are the same two operations against another node, over
	// HTTP. They are separate seams from Deps.Dispatch because a stage is not
	// accepted-and-forgotten: it runs for minutes and is polled, where a
	// power command is answered once and ends the machine.
	Start  func(ctx context.Context, node inventory.Node, req UpgradeStageRequest, token string) (UpgradeStageState, error)
	Status func(ctx context.Context, node inventory.Node, operationID, stageName, token string) (UpgradeStageState, error)

	// Readiness asks one node whether it can run a stage of this upgrade, over
	// the same token-guarded surface as the other two. See UpgradeReadiness.
	Readiness func(ctx context.Context, node inventory.Node, operationID, token string) (UpgradeReadiness, error)
}

// complete reports whether every seam an upgrade needs is present. A partly
// wired set is refused the same way a missing one is: the difference is not
// something a run part way through a cluster should discover.
func (d *UpgradeDeps) complete() bool {
	return d != nil && d.Plan != nil && d.Local != nil && d.Auth != nil &&
		d.Start != nil && d.Status != nil && d.Readiness != nil
}
