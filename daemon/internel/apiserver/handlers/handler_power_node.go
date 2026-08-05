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

// powerThisNode is the execution seam, left nil until the daemon installs the
// real one in main. A test binary therefore has nothing that can power the
// machine running it: a route reached without a fake refuses rather than
// reboots.
var powerThisNode func(context.Context, clusterop.Type) error
var powerClaims interface {
	Claim(string) error
	Release(string) error
	Complete(string) error
	Completed(string) (bool, error)
}

const powerExecutionUnavailable = "this node cannot carry out power commands yet"

// InstallPowerExecution gives this node the ability to power its own machine.
// The daemon calls it once at startup; nothing else does.
func InstallPowerExecution(power func(context.Context, clusterop.Type) error) {
	powerThisNode = power
}

func InstallPowerClaims(claims interface {
	Claim(string) error
	Release(string) error
	Complete(string) error
	Completed(string) (bool, error)
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
func (h *Handlers) PostPowerNode(ctx *fiber.Ctx) error {
	var req clusterop.PeerRequest
	if err := ctx.BodyParser(&req); err != nil {
		return h.errPower(ctx, http.StatusBadRequest, clusterop.CodeInvalidRequest, "unable to parse body")
	}
	opType, err := clusterop.ParseType(strings.TrimSpace(string(req.Type)))
	if err != nil {
		return h.errPower(ctx, http.StatusBadRequest, clusterop.CodeUnsupportedOperation,
			"this daemon does not perform that operation")
	}
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
	binding, err := requireBinding(ctx, clusterop.Binding{
		Type:      opType,
		RequestID: strings.TrimSpace(req.RequestID),
		Scope:     clusterop.ScopeCluster,
	})
	if err != nil {
		return h.errBinding(ctx, err)
	}

	return h.executePower(ctx, opType, req.OperationID, binding)
}

type directNodePowerRequest struct {
	Type      string `json:"type"`
	RequestID string `json:"requestId"`
}

// PostPowerThisNode handles an owner-approved operation aimed directly at the
// node serving the request. The signed target must match this node's cluster
// name, so routing a valid request to another node cannot power it.
func (h *Handlers) PostPowerThisNode(ctx *fiber.Ctx) error {
	var req directNodePowerRequest
	if err := ctx.BodyParser(&req); err != nil {
		return h.errPower(ctx, http.StatusBadRequest, clusterop.CodeInvalidRequest, "unable to parse body")
	}
	opType, err := clusterop.ParseType(strings.TrimSpace(req.Type))
	if err != nil {
		return h.errPower(ctx, http.StatusBadRequest, clusterop.CodeUnsupportedOperation,
			"this daemon does not perform that operation")
	}
	nodeName, role, err := thisNodeInCluster(ctx.Context())
	if err != nil || nodeName == "" {
		klog.Error("resolve this node for a power command: ", err)
		return h.errPower(ctx, http.StatusConflict,
			clusterop.CodeNodeIdentityUnknown, "this node's cluster identity is unavailable")
	}
	if opType == clusterop.TypeShutdown && role != inventory.RoleWorker {
		return h.errPower(ctx, http.StatusConflict,
			clusterop.CodePowerUnsupported, "the control node cannot be shut down by a node operation")
	}
	binding, err := requireBinding(ctx, clusterop.Binding{
		Type:      opType,
		RequestID: strings.TrimSpace(req.RequestID),
		Scope:     clusterop.ScopeNode,
		Target:    nodeName,
	})
	if err != nil {
		return h.errBinding(ctx, err)
	}

	return h.executePower(ctx, opType, req.RequestID, binding)
}

func (h *Handlers) executePower(ctx *fiber.Ctx, opType clusterop.Type, operationID string,
	binding clusterop.Binding) error {
	if powerThisNode == nil {
		return h.errPower(ctx, http.StatusServiceUnavailable,
			clusterop.CodePowerUnsupported, powerExecutionUnavailable)
	}
	if powerClaims == nil {
		return h.errPower(ctx, http.StatusServiceUnavailable,
			clusterop.CodeStatePersistenceFailed, "power request persistence is unavailable")
	}
	claimKey := strings.Join([]string{
		ownerOf(ctx), string(binding.Type), binding.RequestID, binding.Scope, binding.Target,
	}, "\x00")
	if err := powerClaims.Claim(claimKey); err != nil {
		if errors.Is(err, clusterop.ErrClaimExists) {
			completed, stateErr := powerClaims.Completed(claimKey)
			if stateErr != nil {
				klog.Error("read power request claim: ", stateErr)
				return h.errPower(ctx, http.StatusServiceUnavailable,
					clusterop.CodeStatePersistenceFailed, "power request state is unavailable")
			}
			if completed {
				return h.OkJSON(ctx, "success", fiber.Map{"accepted": opType})
			}
			return h.errPower(ctx, http.StatusConflict,
				clusterop.CodeRequestInProgress, "this power request is still in progress")
		}
		klog.Error("persist power request claim: ", err)
		return h.errPower(ctx, http.StatusServiceUnavailable,
			clusterop.CodeStatePersistenceFailed, "power request could not be recorded")
	}

	// The command outlives the request: the process that would carry the
	// request context is about to go away with the machine.
	runCtx := h.mainCtx
	if runCtx == nil {
		runCtx = context.Background()
	}

	klog.Infof("power operation %s: %s this node", operationID, opType)
	if err := powerThisNode(runCtx, opType); err != nil {
		if releaseErr := powerClaims.Release(claimKey); releaseErr != nil {
			klog.Error("release failed power request claim: ", releaseErr)
		}
		// The detail belongs in this node's log. What goes back is the stable
		// code and the fixed sentence that came with it.
		klog.Error("power this node error, ", err)
		var pe *clusterop.PowerError
		if errors.As(err, &pe) {
			return h.errPower(ctx, powerStatus(pe.Code), pe.Code, pe.Message)
		}
		return h.errPower(ctx, http.StatusInternalServerError,
			clusterop.CodeHostPowerFailed, "this node could not be powered")
	}
	if err := powerClaims.Complete(claimKey); err != nil {
		klog.Error("complete power request claim: ", err)
		return h.errPower(ctx, http.StatusServiceUnavailable,
			clusterop.CodeStatePersistenceFailed, "power request completion could not be recorded")
	}

	return h.OkJSON(ctx, "success", fiber.Map{"accepted": opType})
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
// arrived. want describes the request; the signature says whether the owner
// authorized that one in particular and until when.
func requireBinding(ctx *fiber.Ctx, want clusterop.Binding) (clusterop.Binding, error) {
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
	binding, err := clusterop.ParseBinding(signed.SignedBody())
	if err != nil {
		return clusterop.Binding{}, err
	}
	if err := binding.Authorizes(want, time.Now()); err != nil {
		return clusterop.Binding{}, err
	}
	return binding, nil
}
