package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/beclab/Olares/daemon/internel/client"
	"github.com/beclab/Olares/daemon/pkg/cluster/clusterop"
	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

var powerClaims interface {
	Consume(string, time.Time) error
	Forget(string) error
}

const powerExecutionUnavailable = "this node cannot carry out power commands yet"

func InstallPowerClaims(claims interface {
	Consume(string, time.Time) error
	Forget(string) error
}) {
	powerClaims = claims
}

// PostPowerNode powers the node it reaches. It is how the master carries out a
// cluster operation on a compute node, and it names no target: the request
// cannot be aimed at another machine, so reaching it grants no authority that
// the single-node power command on this node does not already grant.
//
// The authority it does need is the owner's signature bound to this operation.
// The master presents the same one it was given, and this node checks it
// against the request that arrived rather than trusting that the master
// checked. No access token crosses the hop; see clusterop.dispatchPower.
//
// Nothing here knows what reboot or shutdown means: the request goes to the
// module registered for its type, which is the same module the generic node
// endpoint would reach. This path exists because an older worker serves it
// and no other, so its request JSON, its statuses and its codes are fixed —
// see clusterop.PeerPath.
//
// What it will not do is reach any other module. Its request shape carries
// no params and it asks no module whether it accepts what arrived, so an
// operation added later would be carried out here without ever being
// validated. clusterop.ExecutePowerNode is what holds it to the two power
// operations that daemon implements itself.
func (h *Handlers) PostPowerNode(ctx *fiber.Ctx) error {
	var req clusterop.PeerRequest
	if err := ctx.BodyParser(&req); err != nil {
		return h.errPower(ctx, http.StatusBadRequest, clusterop.CodeInvalidRequest, "unable to parse body")
	}
	registry := nodeOperations
	if registry == nil {
		return h.errPower(ctx, http.StatusServiceUnavailable,
			clusterop.CodePowerUnsupported, powerExecutionUnavailable)
	}
	opType, err := registry.Parse(strings.TrimSpace(string(req.Type)))
	if err != nil {
		return h.errPower(ctx, http.StatusBadRequest, clusterop.CodeUnsupportedOperation,
			"this daemon does not perform that operation")
	}
	req.Type = opType
	nodeName, role, err := thisNodeInCluster(ctx.Context())
	if err != nil || nodeName == "" {
		klog.Error("resolve this node for a cluster power command: ", err)
		return h.errPower(ctx, http.StatusConflict,
			clusterop.CodeNodeIdentityUnknown, "this node's cluster identity is unavailable")
	}
	if role != inventory.RoleWorker {
		return h.errPower(ctx, http.StatusConflict,
			clusterop.CodePowerUnsupported, "the control node only powers itself through cluster orchestration")
	}
	binding, err := bindNodeRequest(ctx, registry, req, nodeName)
	if err != nil {
		return h.errBinding(ctx, err)
	}
	claim, spent, refusal := h.spendSignature(ctx, powerNodeEndpoint, binding)
	if !spent {
		return refusal
	}

	return h.runNodeOperation(ctx, powerNodeEndpoint, registry,
		clusterop.NodeRequest{PeerRequest: req}, claim)
}

// bindNodeRequest establishes that the owner authorized this exact request on
// this exact machine. Both node endpoints go through it, so neither can end
// up checking less than the other.
//
// The target is checked against nodeName, which the caller resolved from the
// node directory rather than from the body: that is what keeps either
// endpoint from becoming a way to act on some other machine in the cluster.
//
// registry is the module set the signature is read against — the same one
// the operation will be carried out through, so the endpoint cannot accept a
// grant for an operation it would then be unable to find.
func bindNodeRequest(ctx *fiber.Ctx, registry *clusterop.ModuleRegistry, req clusterop.PeerRequest,
	nodeName string) (clusterop.Binding, error) {
	if req.Scope != clusterop.ScopeCluster && req.Scope != clusterop.ScopeNode ||
		(req.Scope == clusterop.ScopeCluster && req.Target != "") ||
		(req.Scope == clusterop.ScopeNode && req.Target != nodeName) {
		return clusterop.Binding{}, &clusterop.BindingError{
			Code: clusterop.CodeSignatureUnbound, Message: "the signature does not authorize this operation",
		}
	}
	binding, err := requireBindingIn(ctx, registry, clusterop.Binding{
		ClusterID: req.ClusterID,
		Type:      req.Type,
		RequestID: strings.TrimSpace(req.RequestID),
		Scope:     req.Scope,
		Target:    req.Target,
	})
	if err != nil {
		return clusterop.Binding{}, err
	}
	localClusterID, err := clusterIDOf(ctx.Context())
	if err != nil || localClusterID == "" || binding.ClusterID != localClusterID {
		return clusterop.Binding{}, &clusterop.BindingError{
			Code: clusterop.CodeSignatureMismatch, Message: "the signature authorizes a different operation",
		}
	}
	return binding, nil
}

// nodeEndpoint is how one node route reaches a module and what it says while
// doing so. The two routes differ here and only here. The power path keeps
// the sentences older callers already read, and reaches a module through the
// narrower helper: that endpoint asks no module whether it accepts the
// request, so it may only serve operations this daemon implements itself.
type nodeEndpoint struct {
	execute func(context.Context, *clusterop.ModuleRegistry, clusterop.NodeRequest) error

	persistenceGone string
	notRecorded     string
	alreadyUsed     string
	failureCode     string
	failureMessage  string
}

var powerNodeEndpoint = nodeEndpoint{
	execute:         clusterop.ExecutePowerNode,
	persistenceGone: "power request persistence is unavailable",
	notRecorded:     "power request could not be recorded",
	alreadyUsed:     "this power request was already used",
	failureCode:     clusterop.CodeHostPowerFailed,
	failureMessage:  "this node could not be powered",
}

var clusterOperationNodeEndpoint = nodeEndpoint{
	execute:         clusterop.ExecuteNode,
	persistenceGone: "cluster operation requests cannot be recorded on this node yet",
	notRecorded:     "the cluster operation request could not be recorded",
	alreadyUsed:     "this cluster operation request was already used",
	failureCode:     clusterop.CodeModuleFailed,
	failureMessage:  "this node could not carry out the operation",
}

// spendSignature records that this signature has now been used on this node.
// It returns the claim, which the caller gives back through releaseClaim if
// nothing was carried out after all.
//
// spent is false when the signature could not be spent, and refusal is then
// the reply for the caller. It is a separate result because a written
// response is a nil error: reading the error alone would let a replayed
// signature carry on into execution.
func (h *Handlers) spendSignature(ctx *fiber.Ctx, ep nodeEndpoint,
	binding clusterop.Binding) (claim string, spent bool, refusal error) {
	if powerClaims == nil {
		return "", false, h.errPower(ctx, http.StatusServiceUnavailable,
			clusterop.CodeStatePersistenceFailed, ep.persistenceGone)
	}
	claim = strings.Join([]string{
		ctx.Get(SIGNATURE_HEADER), binding.ClusterID, binding.RequestID,
	}, "\x00")
	if err := powerClaims.Consume(claim, time.UnixMilli(binding.ExpiresAt)); err != nil {
		if errors.Is(err, clusterop.ErrReplayConflict) {
			return "", false, h.errPower(ctx, http.StatusConflict,
				clusterop.CodeRequestInProgress, ep.alreadyUsed)
		}
		klog.Error("persist node operation claim: ", err)
		return "", false, h.errPower(ctx, http.StatusServiceUnavailable,
			clusterop.CodeStatePersistenceFailed, ep.notRecorded)
	}
	return claim, true, nil
}

// releaseClaim gives a spent signature back, for the one case that warrants
// it: the operation was attempted and failed, so the caller may retry the
// same request. A signature spent on work this node actually did — including
// asking a module to judge the request — is not given back. An empty claim
// is a no-op: types admitted without a signature never spent one.
func releaseClaim(claim string) {
	if claim == "" || powerClaims == nil {
		return
	}
	if err := powerClaims.Forget(claim); err != nil {
		klog.Error("release a failed node operation claim: ", err)
	}
}

// runNodeOperation hands the request to the module and reports what came
// back. Both node endpoints end here, so a module this build does not have
// and a module that fails are answered the same way however the master
// reached this node.
func (h *Handlers) runNodeOperation(ctx *fiber.Ctx, ep nodeEndpoint,
	registry *clusterop.ModuleRegistry, req clusterop.NodeRequest, claim string) error {
	// The command outlives the request: the process that would carry the
	// request context is about to go away with the machine.
	runCtx := h.mainCtx
	if runCtx == nil {
		runCtx = context.Background()
	}

	klog.Infof("cluster operation %s: %s on this node", req.OperationID, req.Type)
	if err := ep.execute(runCtx, registry, req); err != nil {
		releaseClaim(claim)
		// The detail belongs in this node's log. What goes back is the stable
		// code and the fixed sentence that came with it.
		klog.Error("carry out a cluster operation on this node, ", err)
		var pe *clusterop.PowerError
		if errors.As(err, &pe) {
			return h.errPower(ctx, powerStatus(pe.Code), pe.Code, pe.Message)
		}
		return h.errPower(ctx, http.StatusInternalServerError, ep.failureCode, ep.failureMessage)
	}

	return h.OkJSON(ctx, "success", fiber.Map{"accepted": req.Type})
}

// powerStatus maps a refusal to the status that describes it. Unsupported is a
// conflict rather than a client error: the request was well formed and this
// node simply cannot do it.
func powerStatus(code string) int {
	switch code {
	case clusterop.CodePowerUnsupported:
		return http.StatusConflict
	case clusterop.CodeUnsupportedOperation:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func (h *Handlers) errPower(ctx *fiber.Ctx, status int, code, message string) error {
	return h.ErrJSON(ctx, status, message, fiber.Map{"reason": code})
}

func (h *Handlers) errBinding(ctx *fiber.Ctx, err error) error {
	var be *clusterop.BindingError
	if errors.As(err, &be) {
		return h.errPower(ctx, http.StatusForbidden, be.Code, be.Message)
	}
	return h.errPower(ctx, http.StatusForbidden, clusterop.CodeSignatureUnbound, "request is forbidden")
}

// requireBinding checks the verified signature against the request that
// arrived, reading it against the modules built into this daemon.
func requireBinding(ctx *fiber.Ctx, want clusterop.Binding) (clusterop.Binding, error) {
	return requireBindingIn(ctx, clusterop.DefaultRegistry(), want)
}

// requireBindingIn is requireBinding against an explicit module set. want
// describes the request; the signature says whether the owner authorized
// that one in particular and until when. The set matters because the
// signature names an operation, and an operation this set does not hold is
// one the reader cannot confirm the owner meant.
func requireBindingIn(ctx *fiber.Ctx, registry *clusterop.ModuleRegistry,
	want clusterop.Binding) (clusterop.Binding, error) {
	if want.RequestID == "" {
		return clusterop.Binding{}, &clusterop.BindingError{
			Code:    clusterop.CodeSignatureUnbound,
			Message: "the request names no operation for the signature to authorize",
		}
	}
	signed, ok := ctx.Context().UserValue(client.ClIENT_CONTEXT).(client.SignedClient)
	if !ok || signed == nil {
		return clusterop.Binding{}, &clusterop.BindingError{
			Code:    clusterop.CodeSignatureUnbound,
			Message: "the signature does not authorize this operation",
		}
	}
	binding, err := clusterop.ParseBindingIn(registry, signed.SignedBody())
	if err != nil {
		return clusterop.Binding{}, err
	}
	if err := binding.Authorizes(want, time.Now()); err != nil {
		return clusterop.Binding{}, err
	}
	return binding, nil
}
