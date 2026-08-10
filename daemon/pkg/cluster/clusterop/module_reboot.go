package clusterop

import (
	"context"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/commands/reboot"
	"k8s.io/klog/v2"
)

// observationTimeout bounds one read of the cluster's own view while a
// reboot that outlived this daemon is being confirmed.
const observationTimeout = 10 * time.Second

// TypeReboot is what a caller asks for and what goes on the wire to every
// node. It is declared here because this module is what answers for it.
const TypeReboot Type = "reboot"

func init() { MustRegisterModule(rebootModule{}) }

// rebootModule is the cluster reboot: every compute node restarts, is watched
// back onto a new boot, and only then is the control node told to go.
//
// It is the one operation this daemon can prove happened. A machine that
// comes back on a boot it was not on when the command went out really went
// down, and that proof is what lets a reboot left at command_issued be
// confirmed by whichever daemon starts next.
type rebootModule struct{}

var rebootSpec = powerSpec{
	opType:           TypeReboot,
	capability:       nodestatus.CapPowerReboot,
	provesBootChange: true,
	awaitRestart:     true,
	grace:            func(t Timeouts) time.Duration { return t.Ready },
}

func (rebootModule) Type() Type { return TypeReboot }

func (rebootModule) Validate(req CreateRequest) error { return validatePowerScope(req) }

// Phase is what a reboot makes the cluster look like while it is happening,
// and it says nothing about whether it still is: that question has a
// different answer once the command has been issued and the machine has not
// gone down yet, and it belongs to the caller. See PhaseFor.
func (rebootModule) Phase(Operation) (nodestatus.Phase, bool) {
	return nodestatus.PhaseRestarting, true
}

func (rebootModule) Run(ctx context.Context, rt Runtime, req RunRequest) Outcome {
	return runPowerOperation(ctx, rt, req, rebootSpec)
}

// Recover confirms a control-node reboot this daemon can prove happened.
//
// The proof is the machine being on a different boot than the one recorded
// before the command was issued; olaresd restarting on the same boot proves
// nothing. Nothing is promoted until the control node is Ready on that new
// boot either, because a machine part way through coming up has not finished
// the operation it was asked for.
//
// The generic recovery framework (recovery.go) also calls this for a
// non-terminal record — one that stopped before it ever reached
// command_issued — so rebootChangedBoot's own status check is what makes
// this return immediately for that case rather than block: there is no
// command to confirm yet, and the framework's own MarkInterrupted fallback
// is what settles it once this returns. It never blocks on the cluster or
// the clock for anything but an outstanding command_issued reboot, and even
// that is bounded by Timeouts.Ready.
func (rebootModule) Recover(ctx context.Context, rt Runtime, op Operation) {
	m, _, ok := managerOf(rt)
	if !ok {
		return
	}
	// Confirming a reboot moves a stage, a node and the operation together,
	// which is one write or none. A runtime that cannot do that is not one
	// this can safely confirm through. In production this assertion always
	// succeeds: every caller of Recover — Manager.recoverLoadedOperations
	// and Manager.resume — builds it with newRecoveryRuntime. It exists so
	// a Runtime built some other way (a test, or a future caller) is
	// refused rather than silently trusted to behave like one.
	recovery, ok := rt.(recoveryRuntime)
	if !ok {
		klog.Warningf("clusterop: cannot confirm reboot %s: this runtime cannot settle it in one write", op.ID)
		return
	}
	boot, err := m.deps.HostBootID()
	if err != nil {
		// Without a boot id nothing is promoted. Reporting a reboot that may
		// not have happened is worse than leaving it at command_issued.
		klog.Warningf("clusterop: read this machine's boot id: %v", err)
	}
	if !rebootChangedBoot(&op, boot) {
		return
	}

	deadline := m.deps.Now().Add(m.deps.Timeouts.Ready)
	for m.deps.Now().Before(deadline) {
		observeCtx, cancel := context.WithTimeout(ctx, observationTimeout)
		observed, err := m.deps.Observe(observeCtx)
		cancel()
		if err == nil {
			current, ok := rt.Operation()
			if !ok || current.Status != StatusCommandIssued {
				return
			}
			if controlNodeReady(&current, boot, observed) {
				confirmReboot(recovery, &current)
				return
			}
		}
		if err := m.deps.Sleep(ctx, m.deps.Timeouts.Poll); err != nil {
			return
		}
	}
}

// ExecuteNode reboots the machine this daemon runs on, through the same
// execution point and the same state check as the single-node power endpoint.
func (rebootModule) ExecuteNode(ctx context.Context, _ NodeRequest) error {
	return PowerHost(ctx, TypeReboot)
}

// commandName and newCommand are how PowerHost reaches a reboot without
// knowing that a reboot is what it is doing.
func (rebootModule) commandName() string { return "reboot" }

func (rebootModule) newCommand() commands.Interface { return reboot.New() }

// rebootChangedBoot reports the one piece of evidence a control-node reboot
// leaves behind: the machine is on a boot other than the one it was told to
// leave. A missing id on either side is not evidence of anything.
func rebootChangedBoot(op *Operation, boot string) bool {
	return op.Type == TypeReboot && op.Status == StatusCommandIssued &&
		boot != "" && op.HostBootID != "" && boot != op.HostBootID
}

func controlNodeReady(op *Operation, boot string,
	observed map[string]inventory.Observation) bool {
	for _, node := range op.Nodes {
		if node.Role != inventory.RoleMaster {
			continue
		}
		obs, ok := observed[node.NodeName]
		return ok && obs.Ready && obs.BootID == boot
	}
	return false
}

// confirmReboot settles what the issued command turned out to have done: the
// control node's power stage succeeded, the control node itself restarted,
// and the operation ends succeeded. All three are asked for at once, because
// a record that had only some of them applied would describe a cluster that
// never existed.
func confirmReboot(rt recoveryRuntime, op *Operation) {
	s := settlement{outcome: Outcome{Status: StatusSucceeded}}
	for _, step := range op.Steps {
		if step.Name == StepMasterCommand && step.Status == StepCommandIssued {
			s.step = stepSettlement{name: StepMasterCommand, status: StepSucceeded}
			break
		}
	}
	for _, node := range op.Nodes {
		if node.Status != NodeCommandIssued || node.Role != inventory.RoleMaster {
			continue
		}
		s.nodes = append(s.nodes, nodeSettlement{
			name: node.NodeName, from: NodeCommandIssued, to: NodeRestarted,
		})
	}
	if err := rt.Settle(s); err != nil {
		klog.Warningf("clusterop: confirm the reboot: %v", err)
	}
}
