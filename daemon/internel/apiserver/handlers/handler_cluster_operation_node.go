package handlers

import (
	"net/http"
	"strings"

	"github.com/beclab/Olares/daemon/pkg/cluster/clusterop"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

// nodeOperations is the module set this node can carry an operation out
// through, left nil until the daemon installs it in main. A test binary
// therefore holds nothing that can act on the machine running it: a node
// route reached without a module set of its own refuses rather than acts.
var nodeOperations *clusterop.ModuleRegistry

// InstallNodeOperations gives this node the ability to carry out the
// node-local half of a cluster operation. The daemon calls it once at
// startup with the modules built into this build; nothing else does.
func InstallNodeOperations(registry *clusterop.ModuleRegistry) {
	nodeOperations = registry
}

const nodeExecutionUnavailable = "this node cannot carry out cluster operations yet"

// PostClusterOperationNode carries out the node-local half of a cluster
// operation, whichever operation it is. Only a daemon new enough to hold
// this package's module registry serves it, which is why reboot and shutdown
// are still dispatched to the power endpoint instead: an older worker has
// that one and nothing else.
//
// Types that require an owner signature are guarded exactly as the power
// endpoint is — the owner's signature, bound to this operation and checked
// here rather than trusted from the master. Types that do not were admitted
// with an access token on the route; they still act only on the machine they
// reached, so reaching them grants no authority over any other node.
//
// The request carries the module's params. They are not covered by the
// owner's signature: what the owner signed is the operation, the request id,
// the cluster and the scope, and nothing under params is part of that. A
// module must therefore treat params as input from whoever could reach this
// route, and Validate is where it says what it will accept — which is why it
// is asked before anything is carried out, and after a signature that got
// the request this far has already been spent when one was required.
func (h *Handlers) PostClusterOperationNode(ctx *fiber.Ctx) error {
	var req clusterop.NodeRequest
	if err := ctx.BodyParser(&req); err != nil {
		return h.errPower(ctx, http.StatusBadRequest, clusterop.CodeInvalidRequest, "unable to parse body")
	}
	registry := nodeOperations
	if registry == nil {
		return h.errPower(ctx, http.StatusServiceUnavailable,
			clusterop.CodePowerUnsupported, nodeExecutionUnavailable)
	}
	// Resolved against the set this node actually holds, which is the same
	// set the operation will be carried out through. Asking a wider one
	// would accept a type here and fail to find it a moment later.
	opType, err := registry.Parse(strings.TrimSpace(string(req.Type)))
	if err != nil {
		return h.errPower(ctx, http.StatusBadRequest, clusterop.CodeUnsupportedOperation,
			"this daemon does not perform that operation")
	}
	req.Type = opType

	// Every role answers here, control node included. Which machines an
	// operation may act on is the module's to decide, and deciding it once
	// for all modules at the door would answer for operations this build
	// does not yet contain.
	nodeName, _, err := thisNodeInCluster(ctx.Context())
	if err != nil || nodeName == "" {
		klog.Error("resolve this node for a cluster operation: ", err)
		return h.errPower(ctx, http.StatusConflict,
			clusterop.CodeNodeIdentityUnknown, "this node's cluster identity is unavailable")
	}

	module, ok := registry.Lookup(opType)
	if !ok {
		// Parse only accepts a type this registry holds, so reaching here
		// means it lost the module in between.
		return h.errPower(ctx, http.StatusBadRequest, clusterop.CodeUnsupportedOperation,
			"this daemon does not perform that operation")
	}

	var (
		signed clusterop.PeerRequest
		claim  string
	)
	if clusterop.RequiresSignature(opType) {
		binding, err := bindNodeRequest(ctx, registry, req.PeerRequest, nodeName)
		if err != nil {
			return h.errBinding(ctx, err)
		}

		// Spent before the module is asked anything, and not given back if the
		// module then refuses. Judging a request is work this node performs on
		// the strength of the signature, and a caller who could present the same
		// one repeatedly would have an unlimited way to drive somebody else's
		// code with params nothing signed.
		var spent bool
		var unspent error
		claim, spent, unspent = h.spendSignature(ctx, clusterOperationNodeEndpoint, binding)
		if !spent {
			return unspent
		}

		// Every field the module judges the request by comes from the binding
		// rather than from the body. The two agree — the binding was checked
		// against the body to get here — but only one of them the owner signed,
		// and they are not character for character the same: the request id was
		// compared after trimming, so a body may still carry a padded one.
		signed = clusterop.PeerRequest{
			Type:        binding.Type,
			RequestID:   binding.RequestID,
			Scope:       binding.Scope,
			Target:      binding.Target,
			ClusterID:   binding.ClusterID,
			OperationID: req.OperationID,
		}
	} else {
		if err := authorizeNodeRequestScope(ctx, req.PeerRequest, nodeName); err != nil {
			return h.errBinding(ctx, err)
		}
		signed = clusterop.PeerRequest{
			Type:        opType,
			RequestID:   strings.TrimSpace(req.RequestID),
			Scope:       req.Scope,
			Target:      req.Target,
			ClusterID:   req.ClusterID,
			OperationID: req.OperationID,
		}
	}

	// The same boundary the master's own Create asks a module through: a
	// module that cannot answer is not a module that refused, and neither
	// end of the operation may decide that differently from the other.
	refused, judged := clusterop.SafeValidate(module, clusterop.CreateRequest{
		Type:      signed.Type,
		RequestID: signed.RequestID,
		Scope:     signed.Scope,
		Target:    signed.Target,
		ClusterID: signed.ClusterID,
		Owner:     ownerOf(ctx),
		Params:    req.Params,
	})
	if !judged {
		return h.errPower(ctx, http.StatusInternalServerError,
			clusterOperationNodeEndpoint.failureCode, clusterOperationNodeEndpoint.failureMessage)
	}
	if refused != nil {
		// The module's own sentence is detail: it is text written outside
		// this package, about params this node did not check, and it goes
		// to the log rather than to the caller.
		klog.Errorf("cluster operation %s refused a node request: %v", opType, refused)
		return h.errPower(ctx, http.StatusBadRequest, clusterop.CodeInvalidRequest,
			"this node cannot carry out the operation as asked")
	}

	// Carried out against the same fields it was judged by, so nothing the
	// module was shown can differ from what it is then asked to act on.
	return h.runNodeOperation(ctx, clusterOperationNodeEndpoint, registry,
		clusterop.NodeRequest{PeerRequest: signed, Params: req.Params}, claim)
}

// authorizeNodeRequestScope checks the scope/target and cluster id of a
// node request that was not admitted by an operation-bound signature. The
// target is still compared to nodeName from the directory, so this hop
// cannot be aimed at another machine.
func authorizeNodeRequestScope(ctx *fiber.Ctx, req clusterop.PeerRequest, nodeName string) error {
	if req.Scope != clusterop.ScopeCluster && req.Scope != clusterop.ScopeNode ||
		(req.Scope == clusterop.ScopeCluster && req.Target != "") ||
		(req.Scope == clusterop.ScopeNode && req.Target != nodeName) {
		return &clusterop.BindingError{
			Code: clusterop.CodeSignatureUnbound, Message: "the signature does not authorize this operation",
		}
	}
	localClusterID, err := clusterIDOf(ctx.Context())
	if err != nil || localClusterID == "" || req.ClusterID != localClusterID {
		return &clusterop.BindingError{
			Code: clusterop.CodeSignatureMismatch, Message: "the signature authorizes a different operation",
		}
	}
	return nil
}
