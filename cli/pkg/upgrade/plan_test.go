package upgrade

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func namedTask(name string) task.Interface {
	return &task.LocalTask{Name: name}
}

// The flow is the same list for every version. A release changes what is in
// the stages, never which stages exist — that is what makes "which stages are
// there" a question with one answer instead of one per release.
func TestFlowIsTheSameShapeForEveryVersion(t *testing.T) {
	a, err := BuildPlan(semver.MustParse("1.12.7"))
	require.NoError(t, err)
	b, err := BuildPlan(semver.MustParse("1.12.4"))
	require.NoError(t, err)

	assert.Equal(t, a.StageNames(), b.StageNames())

	var flow []string
	for _, h := range UpgradeFlow {
		flow = append(flow, h.Name)
	}
	assert.Equal(t, flow, a.StageNames())
}

// Every stage of the flow is in the plan, including the ones this version puts
// nothing in. A plan whose stages come and go cannot be checked against the
// flow it is supposed to be an instance of.
func TestEmptyStagesStayInThePlan(t *testing.T) {
	plan, err := BuildPlan(semver.MustParse("1.12.7"))
	require.NoError(t, err)
	require.Len(t, plan.Stages, len(UpgradeFlow))

	notify, ok := plan.Stage(StageNotifyStart)
	require.True(t, ok)
	assert.True(t, notify.Empty(), "nothing contributes to notify-start today")
}

// The fanout comes from the stage, not from anything a task said.
func TestPlacementIsAPropertyOfTheStage(t *testing.T) {
	plan, err := BuildPlan(semver.MustParse("1.12.7"))
	require.NoError(t, err)

	for _, h := range plan.Stages {
		flowStage, ok := StageByName(h.Name)
		require.True(t, ok, "stage %s is not in the flow", h.Name)
		assert.Equal(t, flowStage.Placement, h.Placement)
		assert.Equal(t, flowStage.MaxParallel, h.MaxParallel)
		assert.Equal(t, flowStage.TimeoutSeconds, h.TimeoutSeconds)
		assert.True(t, h.Placement.Valid(), "stage %s has placement %q", h.Name, h.Placement)
	}
}

// Work a machine does to itself belongs to a node stage, and work against the
// cluster belongs to a control-node stage. This is the placement the weekly
// was about, so it is asserted rather than left to reading.
func TestPerMachineWorkIsPlacedOnNodeStages(t *testing.T) {
	plan, err := BuildPlan(semver.MustParse("1.12.7"))
	require.NoError(t, err)

	stageOf := map[string]string{}
	for _, h := range plan.Stages {
		for _, name := range h.Tasks {
			stageOf[name] = h.Name
		}
	}

	// /etc/containerd, /etc/hosts and the release markers are files on a
	// machine: every node writes its own.
	for _, name := range []string{
		"MigrateContainerdConfigV3", "RestartContainerd", "Update Hosts",
		"UpdateReleaseFile", "UpdatePreparedMarker", "UpdateInstalledMarker",
	} {
		stage, ok := stageOf[name]
		require.True(t, ok, "%s is in no stage", name)
		flowStage, _ := StageByName(stage)
		assert.True(t, flowStage.Placement.RunsOnWorkers(),
			"%s is in %s, which does not run on compute nodes", name, stage)
	}

	// Helm and kubectl talk to the cluster: the control node does that alone.
	for _, name := range []string{
		"UpgradeSystemComponents", "UpgradeUserComponents", "UpdateOlaresVersion",
	} {
		stage, ok := stageOf[name]
		require.True(t, ok, "%s is in no stage", name)
		flowStage, _ := StageByName(stage)
		assert.Equal(t, PlacementAdmin, flowStage.Placement,
			"%s is in %s, which is not control-node only", name, stage)
	}
}

// The GPU driver writes a marker that olaresd reads until the machine
// restarts, and the version flip has to land between the two. Splitting the
// driver and the reboot across stages has to keep that order.
func TestGPUDriverIsInstalledBeforeTheVersionFlipAndRebootedAfter(t *testing.T) {
	plan, err := BuildPlan(semver.MustParse("1.12.4"))
	require.NoError(t, err)

	pos := map[string]int{}
	for i, h := range plan.Stages {
		for _, name := range h.Tasks {
			pos[name] = i
		}
	}
	require.Contains(t, pos, "UpgradeGPUDriver")
	require.Contains(t, pos, "UpdateOlaresVersion")
	require.Contains(t, pos, "RebootIfNeeded")

	assert.Less(t, pos["UpgradeGPUDriver"], pos["UpdateOlaresVersion"])
	assert.Less(t, pos["UpdateOlaresVersion"], pos["RebootIfNeeded"])
}

// A single-node cluster runs every stage on the one machine, in flow order,
// which is the sequence the upgrade has always been.
func TestAllTasksFollowsFlowOrder(t *testing.T) {
	all := AllTasks(getUpgraderByVersion(semver.MustParse("1.12.7")))
	require.NotEmpty(t, all)

	plan, err := BuildPlan(semver.MustParse("1.12.7"))
	require.NoError(t, err)

	var expected []string
	for _, h := range plan.Stages {
		expected = append(expected, h.Tasks...)
	}
	got := make([]string, 0, len(all))
	for _, t := range all {
		got = append(got, t.GetName())
	}
	assert.Equal(t, expected, got)
}

func TestTasksForStageResolvesTheRightStage(t *testing.T) {
	target := semver.MustParse("1.12.7")

	plan, tasks, err := TasksForStage(target, StagePreUpgradeNode)
	require.NoError(t, err)
	require.NotEmpty(t, tasks)

	stage, ok := plan.Stage(StagePreUpgradeNode)
	require.True(t, ok)
	assert.Len(t, tasks, len(stage.Tasks))
}

func TestTasksForStageRefusesAnUnknownStage(t *testing.T) {
	_, _, err := TasksForStage(semver.MustParse("1.12.7"), "no-such-stage")
	require.Error(t, err)
	assert.Contains(t, err.Error(), StageUpgradeCluster)
}

// Every stage a version can contribute to is a stage of the flow. The mapping
// lives in one place, so this is what stops it drifting from UpgradeFlow.
func TestEveryContributedStageIsInTheFlow(t *testing.T) {
	for name := range stageTasks(upgraderBase{}) {
		_, ok := StageByName(name)
		assert.True(t, ok, "stageTasks contributes to %q, which is not in UpgradeFlow", name)
	}
}

// stubUpgrader contributes to two stages and nothing else.
type stubUpgrader struct {
	upgraderBase
	pre     []task.Interface
	cluster []task.Interface
}

func (u stubUpgrader) PreUpgradeNode() []task.Interface           { return u.pre }
func (u stubUpgrader) PrepareForUpgrade() []task.Interface        { return nil }
func (u stubUpgrader) UpgradeSystemComponents() []task.Interface  { return u.cluster }
func (u stubUpgrader) ClearAppChartValues() []task.Interface      { return nil }
func (u stubUpgrader) ClearBFLChartValues() []task.Interface      { return nil }
func (u stubUpgrader) UpdateChartsInAppService() []task.Interface { return nil }
func (u stubUpgrader) UpgradeUserComponents() []task.Interface    { return nil }
func (u stubUpgrader) UpdateReleaseFile() []task.Interface        { return nil }
func (u stubUpgrader) UpdateOlaresVersion() []task.Interface      { return nil }
func (u stubUpgrader) PostUpgrade() []task.Interface              { return nil }

// A stage that takes nodes out of service needs a bound on how many at once,
// and every stage needs a bound on how long one node may spend in it.
func TestEveryStageIsBounded(t *testing.T) {
	for _, stage := range UpgradeFlow {
		if stage.TimeoutSeconds <= 0 {
			t.Errorf("stage %q has no timeout, so a node stuck in it would be waited on "+
				"for whatever the daemon's default happens to be", stage.Name)
		}
		if stage.MaxParallel < 0 {
			t.Errorf("stage %q allows %d nodes at a time", stage.Name, stage.MaxParallel)
		}
	}
}

// The check that the cluster came back runs after the reboot, not before it.
//
// A release that replaces a GPU driver leaves the device plugin running against
// a driver the kernel has not loaded, so it is legitimately not ready until the
// machine restarts. Asked before reboot-nodes, CheckSystemComponentsReady spends
// its retries on a cluster that is fine, fails, and takes the reboot with it —
// leaving the new driver installed and never loaded. This is the order the
// single-node flow had, where the reboot sat inside UpdateOlaresVersion ahead of
// the check.
func TestTheClusterIsVerifiedAfterTheRebootRatherThanBefore(t *testing.T) {
	var reboot, verify = -1, -1
	for i, s := range UpgradeFlow {
		switch s.Name {
		case StageRebootNodes:
			reboot = i
		case StageVerifyCluster:
			verify = i
		}
	}
	if reboot < 0 || verify < 0 {
		t.Fatalf("the flow is missing a stage: reboot=%d verify=%d", reboot, verify)
	}
	if verify < reboot {
		t.Errorf("%s runs at %d, before %s at %d", StageVerifyCluster, verify, StageRebootNodes, reboot)
	}
}

// And on a release that really does replace a driver, the tasks land in that
// order: install, flip the version, reboot, then check.
func TestAGPUReleaseInstallsRebootsThenChecks(t *testing.T) {
	plan, err := BuildPlan(semver.MustParse("1.12.7-20260513"))
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	at := func(task string) int {
		for i, s := range plan.Stages {
			for _, got := range s.Tasks {
				if got == task {
					return i
				}
			}
		}
		t.Fatalf("no stage carries %q", task)
		return -1
	}
	driver, version := at("UpgradeGPUDriver"), at("UpdateOlaresVersion")
	reboot, check := at("RebootIfNeeded"), at("EnsureSystemComponentsUpAndRunningAgain")

	if !(driver <= version && version < reboot && reboot < check) {
		t.Errorf("stages ran driver=%d version=%d reboot=%d check=%d, want the check last",
			driver, version, reboot, check)
	}
}

// Whatever restarts the control plane is followed, in the same stage and with
// nothing in between, by the wait for the apiserver to answer again.
//
// The two halves of the regeneration restart different things in different
// places — k3s on every node, kubeadm on the control node — and they are
// placed in different stages accordingly. What has to hold for both is this
// adjacency: a stage boundary between the restart and the wait would let the
// orchestrator start the next stage's work against an apiserver that is still
// coming back, and a task inserted between them would run against it directly.
// The wait is kept inside the same helper as the restart so that moving the
// work cannot leave it behind.
func TestTheApiserverWaitImmediatelyFollowsTheRestart(t *testing.T) {
	plan, err := BuildPlan(semver.MustParse("1.12.7-20260824"))
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}

	// Whichever kube type this build resolves, the last regeneration task is
	// the one that restarts something.
	restart := "RestartK3sService"
	if _, ok := taskIndex(plan, restart); !ok {
		restart = "RegenerateK8sFilesWithKubeadm"
	}
	restartStage, restartPos := taskPosition(t, plan, restart)
	waitStage, waitPos := taskPosition(t, plan, "WaitForKubeAPIServerUp")

	if waitStage != restartStage {
		t.Fatalf("%s is in stage %q and the wait in %q; a stage boundary between them lets"+
			" the next stage begin against an apiserver that is still restarting",
			restart, plan.Stages[restartStage].Name, plan.Stages[waitStage].Name)
	}
	if waitPos != restartPos+1 {
		t.Errorf("stage %q runs %v; the wait has to come straight after %s, and nothing"+
			" may be scheduled in between", plan.Stages[waitStage].Name,
			plan.Stages[waitStage].Tasks, restart)
	}
}

// The kubeadm regeneration re-runs kubeadm init, which writes control-plane
// manifests and certificates. On a compute node that is not a regeneration but
// a promotion, so it may only ever be scheduled somewhere the control node
// runs alone.
func TestTheKubeadmRegenerationIsNeverScheduledOnAComputeNode(t *testing.T) {
	plan, err := BuildPlan(semver.MustParse("1.12.7-20260824"))
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	for _, only := range []string{"RegenerateKubeadmConfig", "RegenerateK8sFilesWithKubeadm"} {
		i, ok := taskIndex(plan, only)
		if !ok {
			continue // this build resolved k3s, where the task does not exist
		}
		if got := plan.Stages[i].Placement; got != PlacementAdmin {
			t.Errorf("%s is scheduled %q in stage %q, want %q",
				only, got, plan.Stages[i].Name, PlacementAdmin)
		}
	}
}

func taskPosition(t *testing.T, plan Plan, task string) (stage, pos int) {
	t.Helper()
	for i, s := range plan.Stages {
		for j, got := range s.Tasks {
			if got == task {
				return i, j
			}
		}
	}
	t.Fatalf("no stage carries %q", task)
	return -1, -1
}

func taskIndex(plan Plan, task string) (int, bool) {
	for i, s := range plan.Stages {
		for _, got := range s.Tasks {
			if got == task {
				return i, true
			}
		}
	}
	return 0, false
}
