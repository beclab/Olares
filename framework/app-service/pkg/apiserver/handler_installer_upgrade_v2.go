package apiserver

import (
	"context"
	"fmt"
	"slices"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// appUpgradeV2 is the rewrite of `appUpgrade`.
//
// Design notes vs. simpler v2 handlers:
//
//   - The outer shell is the largest of the six migrated handlers
//     because upgrade requires a full chart-load + lint pipeline
//     BEFORE the state machine is engaged. That pipeline (helper
//     interface + v1/v2/v3 dispatch, oac.Lint, applyAppEnv, expose-
//     port preservation, config JSON re-encoding) is preserved
//     verbatim from legacy; the state-machine cutover only touches
//     the tail of the flow.
//
//   - Two spec fields update: OpType=UpgradeOp AND Spec.Config
//     (the freshly-rendered chart config JSON the Upgrading state-
//     machine handler reads to run helm upgrade). Both live in one
//     MergeFrom patch alongside the upgrade annotations.
//
//   - Annotations stamped:
//
//   - AppRepoURLKey       — target chart repo
//
//   - AppTargetChartVersionKey — requested chart version. AppVersionKey
//     stays the deployed chart until the upgrade succeeds.
//
//   - AppPreUpgradeConfigKey — config of the deployed release, kept
//     across retries from UpgradeFailed
//
//   - AppTokenKey         — caller-scoped SA token
//
//   - AppMarketSourceKey  — market provenance
//
//   - AppPreUpgradeStateKey — the state we're upgrading FROM;
//     this is what the Upgrading state-machine handler keys on
//     (via handlers.AnnotPreUpgradeState) to route the tail of
//     the upgrade into either Initializing (default) or
//     Stopped (when we upgraded from Stopped so replicas stay
//     at 0). See handlers/upgrading.go for the routing table.
//
//   - Same nil-annotations defensive rejection as applyEnvV2 —
//     preserved verbatim so the v2 error surface stays byte-
//     identical to the legacy client contract.
//
//   - No `?type=` distinction. Upgrade is only user-triggered; no
//     timeout / system origin to disambiguate.
//
// Not wired into the router yet — same incremental-cutover story
// as cancelV2.
func (h *Handler) appUpgradeV2(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)
	marketSource := req.HeaderParameter(constants.MarketSource)

	if h.opDispatcher == nil {
		api.HandleError(resp, req, fmt.Errorf("appUpgradeV2: opDispatcher is not wired"))
		return
	}

	request := &api.UpgradeRequest{}
	if err := req.ReadEntity(request); err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	if !slices.Contains([]api.AppSource{
		api.Market, api.Custom, api.DevBox, api.System,
	}, request.Source) {
		api.HandleBadRequest(resp, req, fmt.Errorf("unsupported chart source: %s", request.Source))
		return
	}

	// Auth — the appUpgrade handler doesn't use
	// loadAuthorizedLifecycleAM because it needs isShared / prevCfg
	// side-channels. Inline the resolver + admin gate here so we
	// stay symmetric with legacy.
	var appMgr appv1alpha1.ApplicationManager
	appMgrName, isShared, err := apputils.ResolveAppMgrName(req.Request.Context(), app, owner)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	if err = h.ctrlClient.Get(req.Request.Context(), types.NamespacedName{Name: appMgrName}, &appMgr); err != nil {
		api.HandleError(resp, req, err)
		return
	}
	var prevCfg appcfg.ApplicationConfig
	if err = appcfg.GetAppConfig(&appMgr, &prevCfg); err != nil {
		klog.Errorf("Failed to get previous app config err=%v", err)
		api.HandleError(resp, req, err)
		return
	}

	if isShared || appcfg.IsShared(&appMgr) {
		isAdmin, ierr := kubesphere.IsAdmin(req.Request.Context(), h.kubeConfig, owner)
		if ierr != nil {
			api.HandleError(resp, req, ierr)
			return
		}
		if !isAdmin {
			api.HandleForbidden(resp, req, fmt.Errorf("only admin users can upgrade shared app %q", app))
			return
		}
	}

	if appMgr.Spec.Source != request.Source.String() {
		api.HandleBadRequest(resp, req, fmt.Errorf("unmatched chart source"))
		return
	}

	if blockIfRequestPending(resp, req, &appMgr, appv1alpha1.UpgradeOp) {
		return
	}
	if !domain.DefaultTable.AllowsOp(appMgr.Status.State, appv1alpha1.UpgradeOp) {
		api.HandleBadRequest(resp, req,
			fmt.Errorf("%s operation is not allowed for %s state", appv1alpha1.UpgradeOp, appMgr.Status.State))
		return
	}

	token, err := h.GetUserServiceAccountToken(req.Request.Context(), owner)
	if err != nil {
		klog.Error("Failed to get user service account token: ", err)
		api.HandleError(resp, req, err)
		return
	}
	rawAppName := appMgr.Spec.AppName
	if appMgr.Spec.RawAppName != "" {
		rawAppName = appMgr.Spec.RawAppName
	}
	apiVersion, err := apputils.GetAppConfigVersion(req.Request.Context(), &apputils.ConfigOptions{
		App:          app,
		RawAppName:   rawAppName,
		Owner:        prevCfg.OwnerName,
		RepoURL:      request.RepoURL,
		MarketSource: marketSource,
	})
	if err != nil {
		klog.Errorf("Failed to get api version err=%v", err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	var helper upgradeHelperIntf
	base := &upgradeHandlerHelper{
		h:          h,
		req:        req,
		resp:       resp,
		request:    request,
		app:        app,
		rawAppName: rawAppName,
		owner:      prevCfg.OwnerName,
		token:      token,
	}
	switch apiVersion {
	case appcfg.V1:
		klog.Info("Using upgrade handler helper for V1 (v2 flow)")
		helper = base
	case appcfg.V2:
		klog.Info("Using upgrade handler helper for V2 (v2 flow)")
		helper = &upgradeHandlerHelperV2{upgradeHandlerHelper: base}
	case appcfg.V3:
		klog.Info("Using upgrade handler helper for V3 (v2 flow)")
		helper = &upgradeHandlerHelperV3{upgradeHandlerHelper: base}
	default:
		klog.Errorf("Unsupported app config api version: %s", apiVersion)
		api.HandleBadRequest(resp, req, fmt.Errorf("unsupported app config api version: %s", apiVersion))
		return
	}

	adminUsers, isAdmin, err := helper.getAdminUsers()
	if err != nil {
		klog.Errorf("Failed to get admin users err=%v", err)
		return
	}

	if _, err = helper.getAppConfig(&prevCfg, adminUsers, marketSource, isAdmin); err != nil {
		klog.Errorf("Failed to get app config err=%v", err)
		return
	}
	if err = helper.validate(); err != nil {
		klog.Errorf("Failed to validate app config err=%v", err)
		return
	}
	if err = helper.lintChart(); err != nil {
		klog.Errorf("Failed to lint chart err=%v", err)
		return
	}

	// Hold env batch lease during upgrade kickoff to avoid the
	// AppEnv controller racing and flipping op/state to ApplyEnv
	// in this window. Preserved verbatim from legacy.
	userNamespace := utils.UserspaceName(prevCfg.OwnerName)
	releaseLease, err := h.acquireUserEnvBatchLease(req.Request.Context(), userNamespace)
	if err != nil {
		klog.Errorf("Failed to acquire user env batch lease err=%v", err)
		api.HandleError(resp, req, err)
		return
	}
	if releaseLease != nil {
		defer releaseLease()
	}

	if err = helper.applyAppEnv(req.Request.Context()); err != nil {
		klog.Errorf("Failed to apply app env err=%v", err)
		return
	}

	configJSON, err := helper.setAndEncodingAppCofnig(&prevCfg)
	if err != nil {
		klog.Errorf("Failed to encode app config err=%v", err)
		return
	}

	h.deliverUpgradeV2(req.Request.Context(), req, resp, appMgrName, &appMgr, app, upgradeV2Options{
		Token:           token,
		RepoURL:         request.RepoURL,
		Version:         request.Version,
		MarketSource:    marketSource,
		ConfigJSON:      configJSON,
		PreUpgradeState: string(appMgr.Status.State),
	})
}

// upgradeV2Options is the pre-decoded caller intent
// deliverUpgradeV2 acts on. Fields correspond 1:1 to the five
// annotations + Spec.Config write; the split keeps the translator
// unit-testable without dragging in the chart-load pipeline.
type upgradeV2Options struct {
	Token           string
	RepoURL         string
	Version         string
	MarketSource    string
	ConfigJSON      string
	PreUpgradeState string
}

// deliverUpgradeV2 is the pure translator step of appUpgradeV2.
//
// Contract:
//
//   - opDispatcher MUST be non-nil (checked by the outer shell).
//   - amPtr MUST NOT be nil.
//   - AM.Annotations == nil is rejected via api.HandleError (5xx-
//     shape). Legacy behaviour preserved so the two handlers keep
//     the same rejection surface for as long as they coexist.
//   - Writes Spec.OpType + Spec.Config + upgrade annotations through
//     one MergeFrom patch — atomic re: optimistic concurrency.
//     AppVersionKey is left as the deployed chart version.
//   - Delivers exactly one TriggerOp{UpgradeOp}.
//   - Does NOT touch AM.Status directly — the Runner owns Status.
func (h *Handler) deliverUpgradeV2(
	ctx context.Context,
	req *restful.Request,
	resp *restful.Response,
	name string,
	amPtr *appv1alpha1.ApplicationManager,
	app string,
	opts upgradeV2Options,
) {
	am := *amPtr

	// Defensive re-check: the outer shell already validated the
	// state, but the chart-load pipeline is expensive and can span
	// seconds during which the state may drift (e.g. an operator
	// hits cancel). Duplicating the guard here is cheap and keeps
	// deliverUpgradeV2 a stand-alone unit-test target.
	if blockIfRequestPending(resp, req, &am, appv1alpha1.UpgradeOp) {
		return
	}
	if !domain.DefaultTable.AllowsOp(am.Status.State, appv1alpha1.UpgradeOp) {
		api.HandleBadRequest(resp, req,
			fmt.Errorf("%s operation is not allowed for %s state",
				appv1alpha1.UpgradeOp, am.Status.State))
		return
	}

	// Preserved from legacy — an AM with no annotation map at all
	// is treated as an unsupported target for Upgrade. Mirrors
	// deliverApplyEnvV2 (same reasoning, same shape).
	if am.Annotations == nil {
		klog.Errorf("not support operation %s,name: %s", appv1alpha1.UpgradeOp, am.Spec.AppName)
		api.HandleError(resp, req,
			fmt.Errorf("not support operation %s", appv1alpha1.UpgradeOp))
		return
	}

	patched := am.DeepCopy()
	patched.Spec.OpType = appv1alpha1.UpgradeOp
	api.PrepareUpgradeMetadata(patched.Annotations, am.Spec.Config, opts.Version)
	patched.Spec.Config = opts.ConfigJSON
	patched.Annotations[api.AppRepoURLKey] = opts.RepoURL
	patched.Annotations[api.AppTokenKey] = opts.Token
	patched.Annotations[api.AppMarketSourceKey] = opts.MarketSource
	// Snapshot the pre-upgrade state so the Upgrading state-machine
	// handler can decide whether to scale workloads back up after
	// the helm upgrade. When the app was already Stopped, the
	// upgrade lands back in Stopped with the new chart version
	// installed at replicas=0. Key is byte-identical to
	// handlers.AnnotPreUpgradeState — the two names must stay
	// synced (pinned in payload_keys.go's comment).
	patched.Annotations[api.AppPreUpgradeStateKey] = opts.PreUpgradeState
	var opID string
	patched.Annotations, opID = stampDurableOperationRequest(patched.Annotations)

	if err := h.ctrlClient.Patch(ctx, patched, client.MergeFrom(&am)); err != nil {
		api.HandleError(resp, req, err)
		return
	}

	h.opDispatcher.Trigger(name, domain.TriggerOp{Op: appv1alpha1.UpgradeOp})

	resp.WriteEntity(api.InstallationResponse{
		Response: api.Response{Code: 200},
		Data:     api.InstallationResponseData{UID: app, OpID: opID},
	})
}
