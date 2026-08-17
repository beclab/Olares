package upgrade

// Placement is where a stage's tasks run. It is a property of the stage, never of
// a task: the same task placed in a different stage runs somewhere else, and
// nothing about the task changes.
type Placement string

const (
	// PlacementAdmin runs on the control node alone.
	PlacementAdmin Placement = "admin"

	// PlacementWorkers runs on every compute node and not on the control node.
	PlacementWorkers Placement = "workers"

	// PlacementAllNodes runs on every node, compute nodes before the control
	// node.
	PlacementAllNodes Placement = "all-nodes"
)

func (f Placement) Valid() bool {
	switch f {
	case PlacementAdmin, PlacementWorkers, PlacementAllNodes:
		return true
	default:
		return false
	}
}

// RunsOnWorkers reports whether compute nodes take part in this stage.
func (f Placement) RunsOnWorkers() bool { return f == PlacementWorkers || f == PlacementAllNodes }

// Stage names. These are the points in the upgrade flow that tasks can be
// placed into, and there are only these: adding a task means choosing one of
// them, not inventing a place in a list.
const (
	// StageNotifyStart tells the components that need to know an upgrade is
	// beginning..
	StageNotifyStart = "notify-start"

	// StageReplaceBinaries is the nesting-doll replacement of olaresd and
	// olares-cli on every machine..
	StageReplaceBinaries = "replace-binaries"

	// StageImportImages loads the images this upgrade needs onto every
	// machine..
	StageImportImages = "import-images"

	// StagePreUpgradeNode is the pre-upgrade script on every node..
	StagePreUpgradeNode = "pre-upgrade-node"

	// StagePreUpgradeAdmin is the pre-upgrade work that only the control node
	// does: KubeSphere manifests, CRDs, prometheus rules — everything that
	// runs against the apiserver before the upgrade proper. It is where most
	// of what a release does before upgrading ends up.
	StagePreUpgradeAdmin = "pre-upgrade-admin"

	// StageUpgradeCluster is the cluster upgrade itself, which Kubernetes
	// carries out from the control node..
	StageUpgradeCluster = "upgrade-cluster"

	// StagePostUpgradeNode is the post-upgrade script on every node..
	StagePostUpgradeNode = "post-upgrade-node"

	// StagePostUpgradeAdmin is the post-upgrade script on the control node.
	StagePostUpgradeAdmin = "post-upgrade-admin"

	// StageRebootNodes restarts the machines that need it, after everything
	// else has been recorded. It comes last because the GPU driver upgrade
	// reboots the machines whose driver changed and has to do it after the
	// version has been flipped: the marker olares-cli writes before rebooting
	// is what makes olaresd report upgrading-and-rebooting rather than
	// briefly reporting success.
	StageRebootNodes = "reboot-nodes"

	// StageNotifyDone tells the components that the upgrade is over..
	StageNotifyDone = "notify-done"
)

// Stage is one point in the flow: a name, where its tasks run, and how.
type Stage struct {
	Name      string    `json:"name"`
	Placement Placement `json:"placement"`

	// MaxParallel is how many nodes may be inside this stage at once. Zero
	// means no limit.
	//
	// It is a count rather than a serial flag because the thing being
	// expressed is how much of the cluster this stage may take away at a time,
	// and that is not a yes-or-no on anything bigger than a couple of nodes.
	// One is the strictest answer and the right one for work that costs the
	// cluster something while it happens — replacing a container runtime,
	// installing a GPU driver, rebooting — but on twenty nodes "one at a
	// time" is an afternoon and "all at once" is an outage.
	//
	// A limit also changes what happens after a failure: a bounded stage stops
	// giving the work to new nodes, because the reason it was bounded is that
	// running it costs something. An unbounded one has already given it to
	// everyone.
	MaxParallel int `json:"maxParallel,omitempty"`

	// TimeoutSeconds is how long one node may take over this stage. Zero
	// leaves it to the daemon's default.
	//
	// It belongs to the flow rather than to olaresd because the flow is where
	// the work is known. The daemon can only pick one number for every stage,
	// and there is no such number: notify-start does nothing and
	// upgrade-cluster upgrades every chart in the cluster and then waits for
	// the pods. A bound loose enough for the second is not a bound at all for
	// the first.
	TimeoutSeconds int `json:"timeoutSeconds,omitempty"`

	// AwaitRestart says the tasks here may take the machine down, so the
	// stage is not over when olares-cli exits — it is over when every node
	// that went down is back and Ready.
	//
	// Rebooting is the case: olares-cli issues the reboot and returns
	// immediately, a moment before the machine stops answering. Without this
	// the orchestrator reads that as the node being finished, and a stage
	// limited to one node at a time reboots the next one while the previous
	// is still coming up — which is exactly what MaxParallel was set to
	// prevent.
	AwaitRestart bool `json:"awaitRestart,omitempty"`

	// Desc is what this point in the flow is for, shown to whoever is
	// watching an upgrade rather than reading this file.
	Desc string `json:"desc,omitempty"`
}

// UpgradeFlow is the upgrade, as a fixed sequence of stages.
//
// It is written here once and does not vary by version. A release adds tasks
// to these stages; it does not add, reorder or rename them. That is the whole
// point of the abstraction: a flat task list that every contributor appends
// to has no shape anyone can reason about, and the shape is what says when
// work on one node has to finish before work on another begins.
//
// Every boundary between two stages is a barrier: the next stage starts on no
// node until this one has finished on every node it was scheduled on.
// The timeouts are per node and deliberately generous: they are there to end a
// stage that has stopped making progress, not to hurry one along. Where a stage
// waits on the cluster rather than on itself, the bound is the task's own
// retry budget plus room — post-upgrade-admin waits sixty times at fifteen
// seconds for the system components, so anything under fifteen minutes would
// cut off a stage that was working.
var UpgradeFlow = []Stage{
	{
		Name: StageNotifyStart, Placement: PlacementAdmin,
		TimeoutSeconds: 5 * 60,
		Desc:           "tell user-service and l4-gateway an upgrade is starting",
	},
	{
		Name: StageReplaceBinaries, Placement: PlacementAllNodes, MaxParallel: 1,
		TimeoutSeconds: 30 * 60,
		Desc:           "replace olaresd and olares-cli on each machine",
	},
	{
		Name: StageImportImages, Placement: PlacementAllNodes,
		TimeoutSeconds: 30 * 60,
		Desc:           "import the images this upgrade needs on each machine",
	},
	{
		Name: StagePreUpgradeNode, Placement: PlacementAllNodes, MaxParallel: 1,
		TimeoutSeconds: 20 * 60,
		Desc:           "pre-upgrade work each machine does to itself",
	},
	{
		Name: StagePreUpgradeAdmin, Placement: PlacementAdmin,
		TimeoutSeconds: 30 * 60,
		Desc:           "pre-upgrade work the control node does to the cluster",
	},
	{
		Name: StageUpgradeCluster, Placement: PlacementAdmin,
		TimeoutSeconds: 90 * 60,
		Desc:           "upgrade the Olares cluster",
	},
	{
		Name: StagePostUpgradeNode, Placement: PlacementAllNodes,
		// Generous because this is where the GPU driver install lands, and a
		// DKMS build against the running kernel does not fit in ten minutes.
		TimeoutSeconds: 45 * 60,
		Desc:           "post-upgrade work each machine does to itself",
	},
	{
		Name: StagePostUpgradeAdmin, Placement: PlacementAdmin,
		TimeoutSeconds: 45 * 60,
		Desc:           "post-upgrade work the control node does to the cluster",
	},
	{
		Name: StageRebootNodes, Placement: PlacementAllNodes, MaxParallel: 1,
		TimeoutSeconds: 20 * 60, AwaitRestart: true,
		Desc: "restart the machines whose upgrade requires it",
	},
	{
		Name: StageNotifyDone, Placement: PlacementAdmin,
		TimeoutSeconds: 5 * 60,
		Desc:           "tell l4-gateway the upgrade is over",
	},
}

// StageByName finds a stage of the flow.
func StageByName(name string) (Stage, bool) {
	for _, h := range UpgradeFlow {
		if h.Name == name {
			return h, true
		}
	}
	return Stage{}, false
}
