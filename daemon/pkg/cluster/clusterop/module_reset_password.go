package clusterop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
	sshpassword "github.com/beclab/Olares/daemon/pkg/commands/ssh_password"
	"k8s.io/klog/v2"
)

// TypeResetPassword is what a caller asks for and what goes on the wire to
// every node. It is declared here because this module is what answers for it.
const TypeResetPassword Type = "reset-password"

// The stages and node outcomes this module records. They are declared next to
// the operation rather than among the power step names, so a password reset
// never looks like a power command on the record.
const (
	stepResetPasswordWorkers = "reset-password-workers"
	stepResetPasswordMaster  = "reset-password-master"
	stepResetPasswordNode    = "reset-password-node"

	nodeResetSucceeded NodeStatus = "succeeded"

	codeResetPasswordFailed = "reset_password_failed"
)

func init() { MustRegisterModule(resetPasswordModule{}) }

// resetPasswordModule resets the SSH login password on the nodes the request
// names. It reuses the same node-local command the single-node /ssh-password
// endpoint runs, and reaches other nodes through the generic cluster-operation
// fan-out rather than the power endpoint.
//
// Unlike reboot and shutdown it finishes synchronously: the password is either
// set or it is not, so the record ends succeeded / failed / partially_failed
// and never at command_issued. It therefore implements no Recover — an
// interrupted run is settled by the generic framework as daemon_restarted.
type resetPasswordModule struct{}

// resetPasswordParams is this operation's own input. Nothing signs it at any
// hop, so Validate is the only judgement made about it.
type resetPasswordParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Seams the unit tests replace. Production keeps the same command and the
// same state check the single-node SSH password endpoint uses, and the same
// generic fan-out every non-power module reaches workers through.
var (
	newResetPasswordCommand = func() commands.Interface { return sshpassword.New() }

	validateResetPasswordOp = func(cmd commands.Interface) error {
		st, _ := state.Snapshot()
		return state.ValidateOp(st.TerminusState, cmd)
	}

	dispatchResetPassword = DispatchNodeOperation
)

func (resetPasswordModule) Type() Type { return TypeResetPassword }

func (resetPasswordModule) Validate(req CreateRequest) error {
	if err := validatePowerScope(req); err != nil {
		return err
	}
	_, err := parseResetPasswordParams(req.Params)
	return err
}

func (resetPasswordModule) Phase(Operation) (nodestatus.Phase, bool) {
	return nodestatus.PhaseMaintenance, true
}

func (resetPasswordModule) Run(ctx context.Context, rt Runtime, req RunRequest) Outcome {
	m, id, ok := managerOf(rt)
	if !ok {
		return Outcome{
			Status: StatusFailed,
			Code:   CodeUnsupportedOperation,
			Error:  "this runtime cannot carry out a reset-password operation",
		}
	}
	op, ok := rt.Operation()
	if !ok {
		return Outcome{
			Status: StatusFailed,
			Code:   CodeUnsupportedOperation,
			Error:  "the operation this run was started for is gone",
		}
	}
	params, err := parseResetPasswordParams(req.Params)
	if err != nil {
		return Outcome{Status: StatusFailed, Code: CodeModuleFailed, Error: err.Error()}
	}

	if op.Scope == ScopeNode {
		return runResetPasswordNode(ctx, m, rt, id, op, params, req.Creds)
	}
	return runResetPasswordCluster(ctx, m, rt, id, op, params, req.Creds)
}

// ExecuteNode sets the SSH password on the machine this daemon runs on,
// through the same command and the same state check as the single-node
// /ssh-password endpoint.
func (resetPasswordModule) ExecuteNode(ctx context.Context, req NodeRequest) error {
	params, err := parseResetPasswordParams(req.Params)
	if err != nil {
		return err
	}
	return applyResetPassword(ctx, params)
}

func parseResetPasswordParams(raw json.RawMessage) (resetPasswordParams, error) {
	var params resetPasswordParams
	if len(raw) == 0 {
		return params, errors.New("reset-password params require a password")
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return params, fmt.Errorf("reset-password params are not readable: %w", err)
	}
	if strings.TrimSpace(params.Password) == "" {
		return params, errors.New("reset-password params require a password")
	}
	return params, nil
}

func applyResetPassword(ctx context.Context, params resetPasswordParams) error {
	cmd := newResetPasswordCommand()
	if cmd == nil {
		return errors.New("this daemon does not perform that operation")
	}
	if err := validateResetPasswordOp(cmd); err != nil {
		return err
	}
	username := params.Username
	if username == "" {
		username = "olares"
	}
	_, err := cmd.Execute(ctx, &sshpassword.Param{
		Username: username,
		Password: params.Password,
	})
	return err
}

func runResetPasswordCluster(ctx context.Context, m *Manager, rt Runtime, id string,
	op Operation, params resetPasswordParams, creds Credentials) Outcome {
	if err := rt.StartStep(StepPrecheck); err != nil {
		return outcomeStatePersistenceFailed
	}
	nodes, err := m.deps.Inventory(ctx)
	if err != nil {
		reason := suppress(CodeInventoryUnavailable, "read the node directory for reset-password", err)
		_ = rt.FinishStep(StepPrecheck, StepFailed, CodeInventoryUnavailable, reason)
		return failedWith(CodeInventoryUnavailable, reason)
	}
	p, code, err := splitCluster(nodes)
	if err != nil {
		_ = rt.FinishStep(StepPrecheck, StepFailed, code, err.Error())
		return failedWith(code, err.Error())
	}
	results := make([]NodeResult, 0, 1+len(p.workers))
	results = append(results, NodeResult{NodeName: p.master.NodeName, Role: p.master.Role, Status: NodePending})
	for _, w := range p.workers {
		results = append(results, NodeResult{NodeName: w.NodeName, Role: w.Role, Status: NodePending})
	}
	if err := rt.InitNodes(results); err != nil {
		return outcomeStatePersistenceFailed
	}
	if err := rt.FinishStep(StepPrecheck, StepSucceeded, "", ""); err != nil {
		return outcomeStatePersistenceFailed
	}

	workerOK := true
	if len(p.workers) > 0 {
		var failure *Outcome
		workerOK, failure = dispatchResetPasswordWorkers(ctx, m, rt, id, op, p.workers, params, creds)
		if failure != nil {
			return *failure
		}
	}

	if err := rt.StartStep(stepResetPasswordMaster); err != nil {
		return outcomeStatePersistenceFailed
	}
	if err := applyResetPassword(ctx, params); err != nil {
		reason := suppress(codeResetPasswordFailed, "reset the password on the control node", err)
		_ = rt.UpdateNode(p.master.NodeName, func(n *NodeResult) {
			done := m.deps.Now()
			n.Status = NodeFailed
			n.Code = codeResetPasswordFailed
			n.Error = reason
			n.FinishedAt = &done
		})
		_ = rt.FinishStep(stepResetPasswordMaster, StepFailed, codeResetPasswordFailed, reason)
		if workerOK && len(p.workers) > 0 {
			return settledWith(StatusPartiallyFailed, codeResetPasswordFailed,
				"the password was reset on some nodes but not the control node")
		}
		return settledWith(StatusFailed, codeResetPasswordFailed, reason)
	}
	_ = rt.UpdateNode(p.master.NodeName, func(n *NodeResult) {
		done := m.deps.Now()
		n.Status = nodeResetSucceeded
		n.FinishedAt = &done
	})
	if err := rt.FinishStep(stepResetPasswordMaster, StepSucceeded, "", ""); err != nil {
		return outcomeStatePersistenceFailed
	}
	if !workerOK {
		return settledWith(StatusPartiallyFailed, codeResetPasswordFailed,
			"the password was reset on the control node but not on every compute node")
	}
	return Outcome{Status: StatusSucceeded}
}

func runResetPasswordNode(ctx context.Context, m *Manager, rt Runtime, id string,
	op Operation, params resetPasswordParams, creds Credentials) Outcome {
	if err := rt.StartStep(StepPrecheck); err != nil {
		return outcomeStatePersistenceFailed
	}
	nodes, err := m.deps.Inventory(ctx)
	if err != nil {
		reason := suppress(CodeInventoryUnavailable, "read the node directory for reset-password", err)
		_ = rt.FinishStep(StepPrecheck, StepFailed, CodeInventoryUnavailable, reason)
		return failedWith(CodeInventoryUnavailable, reason)
	}
	var node inventory.Node
	for _, candidate := range nodes {
		if candidate.NodeName == op.Target {
			node = candidate
			break
		}
	}
	if node.NodeName == "" {
		const reason = "the node directory could not identify this node"
		_ = rt.FinishStep(StepPrecheck, StepFailed, CodeNodeIdentityUnknown, reason)
		return failedWith(CodeNodeIdentityUnknown, reason)
	}
	if err := rt.InitNodes([]NodeResult{{
		NodeName: node.NodeName, Role: node.Role, Status: NodePending,
	}}); err != nil {
		return outcomeStatePersistenceFailed
	}
	if err := rt.FinishStep(StepPrecheck, StepSucceeded, "", ""); err != nil {
		return outcomeStatePersistenceFailed
	}

	if err := rt.StartStep(stepResetPasswordNode); err != nil {
		return outcomeStatePersistenceFailed
	}

	if node.Role == inventory.RoleMaster || node.IsSelf {
		if err := applyResetPassword(ctx, params); err != nil {
			reason := suppress(codeResetPasswordFailed, "reset the password on this node", err)
			_ = rt.UpdateNode(node.NodeName, markNodeFailed(m, codeResetPasswordFailed, reason))
			_ = rt.FinishStep(stepResetPasswordNode, StepFailed, codeResetPasswordFailed, reason)
			return settledWith(StatusFailed, codeResetPasswordFailed, reason)
		}
		_ = rt.UpdateNode(node.NodeName, markNodeSucceeded(m))
		if err := rt.FinishStep(stepResetPasswordNode, StepSucceeded, "", ""); err != nil {
			return outcomeStatePersistenceFailed
		}
		return Outcome{Status: StatusSucceeded}
	}

	outcomes := dispatchResetPassword(ctx, []inventory.Node{node}, NodeRequest{
		PeerRequest: PeerRequest{
			Type:        TypeResetPassword,
			OperationID: id,
			RequestID:   op.RequestID,
			Scope:       op.Scope,
			Target:      op.Target,
			ClusterID:   op.ClusterID,
		},
		Params: reqParams(params),
	}, creds)
	if len(outcomes) != 1 || outcomes[0].Code != "" {
		code, reason := dispatchFailure(outcomes, node.NodeName)
		_ = rt.UpdateNode(node.NodeName, markNodeFailed(m, code, reason))
		_ = rt.FinishStep(stepResetPasswordNode, StepFailed, code, reason)
		return settledWith(StatusFailed, code, reason)
	}
	_ = rt.UpdateNode(node.NodeName, markNodeSucceeded(m))
	if err := rt.FinishStep(stepResetPasswordNode, StepSucceeded, "", ""); err != nil {
		return outcomeStatePersistenceFailed
	}
	return Outcome{Status: StatusSucceeded}
}

// dispatchResetPasswordWorkers fans the password out to every compute node.
// A failure on any of them is recorded, but the control node is still tried:
// leaving the control node on the old password when workers already moved is
// worse than a partial outcome the caller can act on.
func dispatchResetPasswordWorkers(ctx context.Context, m *Manager, rt Runtime, id string,
	op Operation, workers []inventory.Node, params resetPasswordParams,
	creds Credentials) (allOK bool, failure *Outcome) {
	if err := rt.StartStep(stepResetPasswordWorkers); err != nil {
		stopped := outcomeStatePersistenceFailed
		return false, &stopped
	}
	at := m.deps.Now()
	for _, w := range workers {
		name := w.NodeName
		_ = rt.UpdateNode(name, func(n *NodeResult) { n.StartedAt = &at })
	}
	if !m.canContinue(id) {
		stopped := outcomeStatePersistenceFailed
		return false, &stopped
	}

	outcomes := dispatchResetPassword(ctx, workers, NodeRequest{
		PeerRequest: PeerRequest{
			Type:        TypeResetPassword,
			OperationID: id,
			RequestID:   op.RequestID,
			Scope:       op.Scope,
			Target:      op.Target,
			ClusterID:   op.ClusterID,
		},
		Params: reqParams(params),
	}, creds)

	accepted := 0
	for _, outcome := range outcomes {
		if outcome.Code == "" {
			accepted++
			_ = rt.UpdateNode(outcome.NodeName, markNodeSucceeded(m))
			continue
		}
		reason := suppress(outcome.Code, "dispatch reset-password to node "+outcome.NodeName,
			errors.New(outcome.Err))
		_ = rt.UpdateNode(outcome.NodeName, markNodeFailed(m, outcome.Code, reason))
	}
	if !m.canContinue(id) {
		stopped := outcomeStatePersistenceFailed
		return false, &stopped
	}
	if accepted == len(workers) {
		if err := rt.FinishStep(stepResetPasswordWorkers, StepSucceeded, "", ""); err != nil {
			stopped := outcomeStatePersistenceFailed
			return false, &stopped
		}
		return true, nil
	}
	const reason = "one or more nodes could not reset the password"
	_ = rt.FinishStep(stepResetPasswordWorkers, StepFailed, codeResetPasswordFailed, reason)
	return false, nil
}

func reqParams(params resetPasswordParams) json.RawMessage {
	raw, err := json.Marshal(params)
	if err != nil {
		// params came from JSON and only carries two strings; a marshal
		// failure here would be a programming error rather than input.
		klog.Errorf("clusterop: marshal reset-password params: %v", err)
		return nil
	}
	return raw
}

func markNodeSucceeded(m *Manager) func(*NodeResult) {
	return func(n *NodeResult) {
		done := m.deps.Now()
		n.Status = nodeResetSucceeded
		n.FinishedAt = &done
	}
}

func markNodeFailed(m *Manager, code, reason string) func(*NodeResult) {
	return func(n *NodeResult) {
		done := m.deps.Now()
		n.Status = NodeFailed
		n.Code = code
		n.Error = reason
		n.FinishedAt = &done
	}
}

func dispatchFailure(outcomes []DispatchOutcome, nodeName string) (code, reason string) {
	if len(outcomes) == 0 {
		return CodeDispatchFailed, reasonFor(CodeDispatchFailed)
	}
	outcome := outcomes[0]
	code = outcome.Code
	if code == "" {
		code = CodeDispatchFailed
	}
	var detail error
	if outcome.Err != "" {
		detail = errors.New(outcome.Err)
	}
	return code, suppress(code, "dispatch reset-password to node "+nodeName, detail)
}
