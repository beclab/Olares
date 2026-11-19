package handlers

import (
	"net/http"
	"strings"

	"github.com/beclab/Olares/daemon/pkg/cluster/clusterop"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

// upgradeStages is this node's stage runner, installed once the daemon has a
// state directory to record stages in. Until then the routes exist and refuse.
var upgradeStages clusterop.UpgradeStageRunner

// verifyUpgradeToken is indirected so the routes can be exercised without a
// cluster. The chain itself is never bypassed.
var verifyUpgradeToken = clusterop.VerifyUpgradeToken

const upgradeStagesUnavailable = "this node cannot run upgrade stages yet"

// InstallUpgradeStageRunner gives this node the ability to run upgrade stages.
func InstallUpgradeStageRunner(runner clusterop.UpgradeStageRunner) {
	upgradeStages = runner
}

// PostUpgradeStage runs one stage of a cluster upgrade on the node it reaches.
//
// Like the power endpoint, it names no target: the node it arrives at is the
// node it acts on, so reaching it grants nothing that running olares-cli on
// that machine would not already grant.
//
// What it does not do is check an owner signature. An upgrade runs for an hour
// and survives this daemon being restarted by its own stages, and a
// request-bound signature covers neither — see clusterop.UpgradeDeps.Auth. The
// authorization here is the per-operation token held in the cluster, which is
// exactly as long-lived as the operation and no longer.
func (h *Handlers) PostUpgradeStage(ctx *fiber.Ctx) error {
	if upgradeStages == nil {
		return h.errUpgrade(ctx, http.StatusServiceUnavailable,
			clusterop.CodeUpgradeUnsupported, upgradeStagesUnavailable)
	}

	var req clusterop.UpgradeStageRequest
	if err := ctx.BodyParser(&req); err != nil {
		return h.errUpgrade(ctx, http.StatusBadRequest, clusterop.CodeInvalidRequest, "unable to parse body")
	}
	req.OperationID = strings.TrimSpace(req.OperationID)
	req.Stage = strings.TrimSpace(req.Stage)
	if req.OperationID == "" || req.Stage == "" {
		return h.errUpgrade(ctx, http.StatusBadRequest, clusterop.CodeInvalidRequest,
			"the request names no upgrade stage to run")
	}

	if ok, err := h.authorizeUpgradeStage(ctx, req.OperationID, req.ClusterID); !ok {
		return err
	}

	state, err := upgradeStages.Start(ctx.Context(), req)
	if err != nil {
		klog.Error("start upgrade stage error, ", err)
		return h.errUpgrade(ctx, http.StatusInternalServerError,
			clusterop.CodeStageFailed, "this node could not start the upgrade stage")
	}
	return h.OkJSON(ctx, "success", state)
}

// GetUpgradeStageStatus reports how a stage this node was given is going. It is the
// read side of the same authorization: a caller that may not start a stage may
// not read one either, because the record names the version being rolled out.
func (h *Handlers) GetUpgradeStageStatus(ctx *fiber.Ctx) error {
	if upgradeStages == nil {
		return h.errUpgrade(ctx, http.StatusServiceUnavailable,
			clusterop.CodeUpgradeUnsupported, upgradeStagesUnavailable)
	}

	operationID := strings.TrimSpace(ctx.Query("operationId"))
	stageName := strings.TrimSpace(ctx.Query("stage"))
	if operationID == "" || stageName == "" {
		return h.errUpgrade(ctx, http.StatusBadRequest, clusterop.CodeInvalidRequest,
			"the request names no upgrade stage")
	}
	if ok, err := h.authorizeUpgradeStage(ctx, operationID, ""); !ok {
		return err
	}

	state, ok := upgradeStages.Status(operationID, stageName)
	if !ok {
		return h.errUpgrade(ctx, http.StatusNotFound, clusterop.CodeInvalidRequest,
			"this node has no record of that upgrade stage")
	}
	return h.OkJSON(ctx, "success", state)
}

// authorizeUpgradeStage checks the per-operation token, and the cluster it belongs to
// when the caller named one.
//
// The cluster check mirrors the power endpoints: a token is only meaningful
// inside the cluster that minted it, and this node reads its own kube-system
// UID rather than trusting the one in the request.
//
// It reports whether the caller may proceed, separately from the error the
// handler should return. Signalling a refusal through the error alone does not
// work here: ErrJSON writes the refusal and returns nil, so a caller checking
// only the error would write a 403 and then carry on and run the stage.
func (h *Handlers) authorizeUpgradeStage(ctx *fiber.Ctx, operationID, claimedClusterID string) (bool, error) {
	if claimedClusterID != "" {
		localClusterID, err := clusterIDOf(ctx.Context())
		if err != nil || localClusterID == "" || localClusterID != claimedClusterID {
			return false, h.errUpgrade(ctx, http.StatusForbidden,
				clusterop.CodeSignatureMismatch, "this upgrade is for a different cluster")
		}
	}

	token := ctx.Get(clusterop.UpgradeTokenHeader)
	if err := verifyUpgradeToken(ctx.Context(), operationID, token); err != nil {
		// The detail says which check failed, which is a hint to whoever is
		// guessing. It goes to this node's log.
		klog.Warning("reject upgrade stage: ", err)
		return false, h.errUpgrade(ctx, http.StatusForbidden,
			clusterop.CodeSignatureUnbound, "this request does not authorize an upgrade on this node")
	}
	return true, nil
}

// GetUpgradeReadiness answers whether this node can run a stage of an upgrade.
//
// It exists because the orchestrator cannot ask the general node-status route:
// that one requires a user credential, and an upgrade carries none — the
// owner's signature cannot cover an hour of work that restarts olaresd. The
// answer itself is the same capability that route would report; only the door
// is different.
func (h *Handlers) GetUpgradeReadiness(ctx *fiber.Ctx) error {
	operationID := strings.TrimSpace(ctx.Query("operationId"))
	if operationID == "" {
		return h.errUpgrade(ctx, http.StatusBadRequest, clusterop.CodeInvalidRequest,
			"the request names no upgrade")
	}
	if ok, err := h.authorizeUpgradeStage(ctx, operationID, ""); !ok {
		return err
	}

	// Whether this node has a runner is part of the answer, not a separate
	// question. The capability probe describes the machine — a host, not a
	// container, with an olares-cli on it — and this daemon may still have
	// failed to open the directory it records stages in. Answering "ready" and
	// then refusing the stage would move the failure to the middle of an
	// upgrade, which is the whole thing this probe exists to prevent.
	ready := clusterop.UpgradeReadiness{}
	if upgradeStages != nil {
		snap, _ := currentStateSnapshot()
		identity := identityFrom(ctx.Context(), snap)
		caps := nodestatus.Detect(ctx.Context(), nodestatus.ProbeInput{Identity: identity, State: snap})
		if c, ok := caps[nodestatus.CapUpgradeStages]; ok && c.Supported {
			ready.Supported = true
			ready.CLIVersion, _ = c.Config[nodestatus.CapConfigCLIVersion].(string)
		}
	}
	return h.OkJSON(ctx, "success", ready)
}

func (h *Handlers) errUpgrade(ctx *fiber.Ctx, status int, code, message string) error {
	return h.ErrJSON(ctx, status, message, fiber.Map{"reason": code})
}
