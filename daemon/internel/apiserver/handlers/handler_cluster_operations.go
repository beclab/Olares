package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/beclab/Olares/daemon/internel/client"
	"github.com/beclab/Olares/daemon/pkg/cluster/clusterop"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

// clusterOperationManager is the master's orchestrator, as this package needs
// it. It is an interface so the routes can be driven without a cluster.
type clusterOperationManager interface {
	Create(ctx context.Context, req clusterop.CreateRequest) (clusterop.Operation, error)
	Get(id string) (clusterop.Operation, bool)
	GetByRequest(requestID string) (clusterop.Operation, bool)

	// ActivePhase is what the operation in flight, if any, makes the cluster
	// as a whole be doing. The cluster summary reads it; ok is false when
	// nothing is in flight, which leaves the summary's own phase alone.
	ActivePhase() (nodestatus.Phase, bool)
}

// clusterOperations is set once the daemon has a state directory to record
// operations in. Until then the routes exist and refuse to act.
var clusterOperations clusterOperationManager

const orchestratorUnavailable = "cluster operations are not available on this node yet"

// InitClusterOperations opens the operation records in dir and publishes the
// orchestrator over them. Records written before a restart are read back:
// anything still moving when olaresd stopped is settled as failed rather than
// left holding the cluster's single-operation lock, and a control-node reboot
// left at command_issued is confirmed if the machine is on a different boot.
//
// deps carries every side effect the orchestrator may have, and is a parameter
// rather than something built here so that no test can install one wired to a
// real cluster and a real power command by accident.
//
// What this daemon can be asked to do is settled here too. The module set and
// the signature-requirement set are closed before the orchestrator is built
// over them, so what the routes accept, which types demand an owner signature
// on create, what the owner's signature can be read against, and what the
// manager can carry out are decided once — and a module registering itself
// after the daemon started serving is refused rather than becoming an
// operation half of this process knows about. The same module set is what
// main gives the node routes; see InstallNodeOperations.
func InitClusterOperations(dir string, deps clusterop.Deps) error {
	// Closed first, and whatever happens next: a daemon that could not open
	// its records still serves the routes, and the set they answer from must
	// not stay open just because there was nowhere to write to. Freezing is
	// idempotent, so a second call is not a failure.
	registry := clusterop.DefaultRegistry()
	registry.Freeze()
	clusterop.DefaultSignatureRequirementRegistry().Freeze()

	store, err := clusterop.NewStore(dir)
	if err != nil {
		return err
	}
	claims, err := clusterop.NewReplayGuard(filepath.Join(dir, "power-replays"))
	if err != nil {
		return err
	}
	deps.Store = store

	// The manager is built over the set that was just closed, rather than
	// looking it up again: a module that got in while the records were being
	// read would be one the manager has and no other reader of the same
	// registry ever saw.
	m, err := clusterop.NewManagerWithRegistry(deps, registry)
	if err != nil {
		return err
	}
	clusterOperations = m
	// Also published for the upgrade watcher, which reaches the orchestrator
	// from inside a command rather than from a route. See clusterop.Publish.
	clusterop.Publish(m)
	InstallPowerClaims(claims)
	klog.Info("cluster operations recorded in ", store.Dir())
	return nil
}

type createClusterOperationRequest struct {
	Type      string `json:"type"`
	RequestID string `json:"requestId"`
	Scope     string `json:"scope"`
	Target    string `json:"target"`
	ClusterID string `json:"clusterId"`

	// Params is the module's input, carried through untouched. It is
	// deliberately outside what the owner signs: the signature binds the
	// operation, the request id, the cluster and the scope, and nothing
	// under here is part of that. Whoever can reach this route can choose
	// these bytes freely, so the module that reads them is the one that
	// has to say what it will accept — see OperationModule.Validate. They
	// are also never written down; only a digest of them is, so that a
	// retried request id can be told apart from a changed one.
	Params json.RawMessage `json:"params,omitempty"`
}

// PostClusterOperation starts a cluster-wide power operation and answers with
// the record. It returns before anything has been powered: the operation runs
// on for minutes, and the caller follows it by id.
func (h *Handlers) PostClusterOperation(ctx *fiber.Ctx) error {
	if clusterOperations == nil {
		return h.ErrJSON(ctx, http.StatusServiceUnavailable, orchestratorUnavailable)
	}

	var req createClusterOperationRequest
	if err := ctx.BodyParser(&req); err != nil {
		return h.ErrJSON(ctx, http.StatusBadRequest, "unable to parse body")
	}
	opType, err := clusterop.ParseType(strings.TrimSpace(req.Type))
	if err != nil {
		return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
	}
	if strings.TrimSpace(req.RequestID) == "" {
		return h.ErrJSON(ctx, http.StatusBadRequest, clusterop.ErrRequestIDRequired.Error())
	}
	if req.Scope != clusterop.ScopeCluster && req.Scope != clusterop.ScopeNode ||
		(req.Scope == clusterop.ScopeCluster && req.Target != "") ||
		(req.Scope == clusterop.ScopeNode && req.Target == "") {
		return h.errBinding(ctx, &clusterop.BindingError{
			Code: clusterop.CodeSignatureUnbound, Message: "the signature does not authorize this operation",
		})
	}

	// Types that require a signature bind it to this exact operation, not
	// "anything dangerous for the next twenty minutes". Without the check, a
	// signature captured from any other owner-only route would power the
	// cluster off. Types that do not require one were already admitted by an
	// access token on the route.
	if clusterop.RequiresSignature(opType) {
		if _, err := requireBinding(ctx, clusterop.Binding{
			ClusterID: req.ClusterID,
			Type:      opType,
			RequestID: strings.TrimSpace(req.RequestID),
			Scope:     req.Scope,
			Target:    req.Target,
		}); err != nil {
			return h.errBinding(ctx, err)
		}
	}
	localClusterID, err := clusterIDOf(ctx.Context())
	if err != nil || localClusterID == "" || localClusterID != req.ClusterID {
		return h.errBinding(ctx, &clusterop.BindingError{
			Code: clusterop.CodeSignatureMismatch, Message: "the signature authorizes a different operation",
		})
	}

	owner := ownerOf(ctx)
	if owner == "" {
		return h.ErrJSON(ctx, http.StatusForbidden, "cannot determine who asked for this operation")
	}

	op, err := clusterOperations.Create(ctx.Context(), clusterop.CreateRequest{
		Type:      opType,
		RequestID: strings.TrimSpace(req.RequestID),
		Scope:     req.Scope,
		Target:    req.Target,
		ClusterID: req.ClusterID,
		Owner:     owner,
		Params:    req.Params,
		Creds: clusterop.Credentials{
			// The token stays available for types that fan out without a
			// signature; signature-bound types still present only the JWS to
			// workers — see clusterop.DispatchNodeOperation.
			//
			// Both are copied because Create hands them to a run that outlives
			// this request. ctx.Get returns a string aliasing the fasthttp
			// header buffer, and that buffer is reused by the next request on
			// this connection the moment the handler returns: what the run
			// would otherwise present to a worker is whatever arrived after it.
			Token:     strings.Clone(ctx.Get(AUTH_HEADER)),
			Signature: strings.Clone(ctx.Get(SIGNATURE_HEADER)),
		},
	})
	if err != nil {
		var requestConflict *clusterop.RequestConflictError
		var conflict *clusterop.ConflictError
		var refused *clusterop.ModuleValidationError
		switch {
		case errors.As(err, &refused):
			// The module refused the request before anything was recorded.
			// Its sentence is about params nothing signed and is written
			// outside this package, so it goes to the log and the caller is
			// told only that the request is not one this daemon can carry
			// out as asked.
			klog.Error("cluster operation refused the request, ", err)
			return h.ErrJSON(ctx, http.StatusBadRequest,
				"this cluster cannot carry out the operation as asked")
		case errors.As(err, &requestConflict):
			return h.ErrJSON(ctx, http.StatusConflict, requestConflict.Error(), fiber.Map{
				"requestId":           requestConflict.RequestID,
				"existingOperationId": requestConflict.ExistingID,
			})
		case errors.As(err, &conflict):
			return h.ErrJSON(ctx, http.StatusConflict, conflict.Error(), fiber.Map{
				"activeOperationId": conflict.ActiveID,
				"activeType":        conflict.ActiveType,
			})
		case errors.Is(err, clusterop.ErrRequestIDRequired), errors.Is(err, clusterop.ErrOwnerRequired):
			return h.ErrJSON(ctx, http.StatusBadRequest, err.Error())
		default:
			klog.Error("create cluster operation error, ", err)
			return h.ErrJSON(ctx, http.StatusInternalServerError, "failed to start the cluster operation")
		}
	}

	return h.OkJSON(ctx, "success", op)
}

// GetClusterOperationByRequest serves the operation bound to a caller request id.
func (h *Handlers) GetClusterOperationByRequest(ctx *fiber.Ctx) error {
	if clusterOperations == nil {
		return h.ErrJSON(ctx, http.StatusServiceUnavailable, orchestratorUnavailable)
	}
	requestID, err := url.PathUnescape(ctx.Params("requestId"))
	if err != nil {
		return h.ErrJSON(ctx, http.StatusBadRequest, "invalid requestId")
	}
	if strings.TrimSpace(requestID) == "" {
		return h.ErrJSON(ctx, http.StatusBadRequest, clusterop.ErrRequestIDRequired.Error())
	}
	op, ok := clusterOperations.GetByRequest(requestID)
	if !ok {
		return h.ErrJSON(ctx, http.StatusNotFound, "no such cluster operation")
	}
	return h.OkJSON(ctx, "success", op)
}

// GetClusterOperation serves one operation record.
func (h *Handlers) GetClusterOperation(ctx *fiber.Ctx) error {
	if clusterOperations == nil {
		return h.ErrJSON(ctx, http.StatusServiceUnavailable, orchestratorUnavailable)
	}

	op, ok := clusterOperations.Get(ctx.Params("id"))
	if !ok {
		return h.ErrJSON(ctx, http.StatusNotFound, "no such cluster operation")
	}
	return h.OkJSON(ctx, "success", op)
}

// ownerOf names the identity an operation belongs to. The signature is
// preferred because it is what authorized the operation; the access token's
// user is the fallback for the install-time path where there is no Olares ID
// on the machine yet.
func ownerOf(ctx *fiber.Ctx) string {
	if c, ok := ctx.Context().UserValue(client.ClIENT_CONTEXT).(ownerClient); ok && c != nil {
		if id := c.OlaresID(); id != "" {
			return id
		}
	}
	if u, ok := ctx.Context().UserValue(client.USER_CONTEXT).(*utils.ValidToken); ok && u != nil {
		return u.Username
	}
	return ""
}
