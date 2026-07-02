package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/compute"
	"github.com/beclab/Olares/framework/app-service/pkg/compute/validation"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/emicklei/go-restful/v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// installV2 is the rewrite of `install` — the largest
// and most consequential of the seven op handlers, because it is the
// only one that CREATES the ApplicationManager (the other six only
// mutate an existing AM).
//
// Design vs. the other v2 handlers:
//
//   - Outer shell mirrors legacy `install` verbatim: it walks the
//     full pre-flight pipeline (chart version resolution, apiVersion
//     dispatch to the installHelperIntf V1/V2/V3 helper, admin
//     resolution, installed-app inspection, chart load + lint, clone
//     rejection, compute auto-mode selection with ambiguity-plan
//     escape hatch, install-time validation via InstallabilityValidators,
//     setAppEnv override, per-version validate(), setAppConfig for
//     clone-with-title, applyAppEnv, resolveAutoResources). Every
//     line up to the "apply application manager" call is preserved.
//     None of these steps have state-machine ports yet, so they stay
//     in the apiserver shell for now.
//
//   - Only the tail collapses: instead of calling
//     helper.applyApplicationManager (which internally seeds
//     Status=Pending), v2 asks the helper for its (name,
//     extraLabels) target via installV2Target() and hands off to
//     deliverInstallV2. The translator runs Create-or-Patch on
//     Spec+labels+annotations ONLY (never Status) and then delivers
//     TriggerOp{InstallOp} to the Runner.
//
//   - Empty Status.State on a freshly-created AM is a first-class
//     input to the state machine: the transition table maps
//     InitialState → Pending on TriggerOp{InstallOp} (see
//     domain/table_default.go), and terminal-reinstallable states
//     (Uninstalled, *Canceled, *Failed) accept InstallOp too, so
//     reinstall follows the same Trigger path as fresh install.
//
//   - No Status write here. Legacy stamped a Unix-second opID onto the
//     initial Pending seed; the Runner now owns Status and copies the
//     durable request ID stamped on the annotations into Status.OpID
//     when it consumes the TriggerOp. The response reports that same
//     ID as opID, so a caller can poll on it immediately.
//
// Not wired into the router yet — same incremental-cutover story as
// cancelV2 / uninstallV2 / applyEnvV2 / suspendV2 / resumeV2 /
// upgradeV2.
func (h *Handler) installV2(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)

	token, err := h.GetUserServiceAccountToken(req.Request.Context(), owner)
	if err != nil {
		klog.Error("Failed to get user service account token: ", err)
		api.HandleError(resp, req, err)
		return
	}

	marketSource := req.HeaderParameter(constants.MarketSource)
	klog.Infof("installV2: user: %v, source: %v", owner, marketSource)

	insReq := &api.InstallRequest{}
	if err = req.ReadEntity(insReq); err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	if insReq.Source != api.Market && insReq.Source != api.Custom && insReq.Source != api.DevBox {
		api.HandleBadRequest(resp, req, fmt.Errorf("unsupported chart source: %s", insReq.Source))
		return
	}
	rawAppName := app
	if insReq.RawAppName != "" {
		rawAppName = insReq.RawAppName
	}
	chartVersion, chartOwner := "", ""
	if insReq.RawAppName != "" && !insReq.TemplateClone {
		chartVersion, chartOwner, err = h.getOriginChartVersion(rawAppName, owner)
		if err != nil {
			api.HandleBadRequest(resp, req, err)
			return
		}
	}
	if chartOwner == "" && insReq.Source != api.Market {
		chartOwner = owner
	}

	apiVersion, err := apputils.GetAppConfigVersion(req.Request.Context(), &apputils.ConfigOptions{
		App:          app,
		RawAppName:   rawAppName,
		Owner:        owner,
		RepoURL:      insReq.RepoURL,
		MarketSource: marketSource,
		Version:      chartVersion,
		SelectedGpu:  insReq.SelectedGpuType,
		ChartOwner:   chartOwner,
	})
	if err != nil {
		klog.Errorf("Failed to get api version err=%v", err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	typedClient, err := utils.GetClient()
	if err != nil {
		klog.Errorf("Failed to get client err=%v", err)
		api.HandleError(resp, req, err)
		return
	}

	var helper installHelperIntf
	base := installHandlerHelper{
		h:          h,
		req:        req,
		resp:       resp,
		app:        app,
		rawAppName: rawAppName,
		owner:      owner,
		chartOwner: chartOwner,
		token:      token,
		insReq:     insReq,
		client:     typedClient,
	}
	switch apiVersion {
	case appcfg.V1:
		v1 := &installHandlerHelper{}
		*v1 = base
		v1.validateClusterScope = v1._validateClusterScope
		helper = v1
	case appcfg.V2:
		v2 := &installHandlerHelperV2{installHandlerHelper: base}
		v2.validateClusterScope = v2._validateClusterScope
		helper = v2
	case appcfg.V3:
		v3 := &installHandlerHelperV3{installHandlerHelper: base}
		v3.validateClusterScope = v3._validateClusterScope
		helper = v3
	default:
		klog.Errorf("Unsupported app config api version: %s", apiVersion)
		api.HandleBadRequest(resp, req, fmt.Errorf("unsupported app config api version: %s", apiVersion))
		return
	}

	adminUsers, isAdmin, err := helper.getAdminUsers()
	if err != nil {
		klog.Errorf("Failed to get admin user err=%v", err)
		return
	}

	appInstalled, installedApps, err := helper.getInstalledApps()
	if err != nil {
		klog.Errorf("Failed to get installed app err=%v", err)
		return
	}

	appCfg, err := helper.getAppConfig(adminUsers, marketSource, isAdmin, appInstalled, installedApps, chartVersion, insReq.SelectedGpuType, chartOwner)
	if err != nil {
		klog.Errorf("Failed to get app config err=%v", err)
		return
	}

	if !isAdmin && appCfg.Shared {
		api.HandleForbidden(resp, req, errors.New("only admin users can install shared apps"))
		return
	}

	if err = helper.checkAppNameConflict(req.Request.Context(), appCfg.Shared); err != nil {
		klog.Errorf("Failed to check app name conflict err=%v", err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	if err = helper.lintChart(); err != nil {
		klog.Errorf("Failed to lint chart err=%v", err)
		return
	}

	if !appCfg.AllowMultipleInstall && insReq.RawAppName != "" ||
		(appCfg.AllowMultipleInstall && apiVersion == appcfg.V2) {
		api.HandleBadRequest(resp, req, fmt.Errorf("app %s can not be clone", app))
		return
	}

	// Compute auto-mode selection preserved verbatim: empty
	// SelectedGpuType triggers cluster+manifest derivation, with the
	// AmbiguousComputeMode branch surfacing the full install plan
	// via a dedicated failed-check code instead of failing outright.
	if insReq.SelectedGpuType == "" {
		var chosen string
		chosen, err = compute.AutoSelectMode(req.Request.Context(), h.ctrlClient, appCfg)
		if errors.Is(err, compute.ErrAmbiguousComputeMode) {
			plan, planErr := compute.BuildInstallComputePlan(req.Request.Context(), h.ctrlClient, appCfg)
			if planErr != nil {
				api.HandleError(resp, req, planErr)
				return
			}
			api.HandleFailedCheck(resp, api.CheckTypeComputeModeSelect, plan)
			return
		}
		if err != nil {
			klog.Errorf("Failed to auto-select compute mode for app %s: %v", app, err)
			api.HandleBadRequest(resp, req, err)
			return
		}
		if chosen != appCfg.SelectedGpuType {
			appCfg, err = helper.getAppConfig(adminUsers, marketSource, isAdmin, appInstalled, installedApps, chartVersion, chosen, chartOwner)
			if err != nil {
				klog.Errorf("Failed to reload appConfig with auto-selected gpu type %s: %v", chosen, err)
				api.HandleError(resp, req, err)
				return
			}
		}
	}

	decision, err := validation.Run(req.Request.Context(), validation.Input{
		Client:    h.ctrlClient,
		AppConfig: appCfg,
		Op:        appv1alpha1.InstallOp,
		Token:     token,
	}, validation.InstallabilityValidators()...)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	if !decision.OK {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, api.RequirementResp{
			Response: api.Response{Code: 400},
			Resource: decision.Resource.String(),
			Message:  decision.Message,
			Reason:   decision.Reason.String(),
		})
		return
	}

	if err = helper.setAppEnv(insReq.Envs); err != nil {
		klog.Errorf("Failed to set app env err=%v", err)
		return
	}

	if err = helper.validate(isAdmin, installedApps); err != nil {
		klog.Errorf("Failed to validate app install request err=%v", err)
		return
	}
	if insReq.RawAppName != "" && insReq.Title != "" {
		helper.setAppConfig(insReq, app)
	}

	if err = helper.applyAppEnv(req.Request.Context()); err != nil {
		klog.Errorf("Failed to apply app env err=%v", err)
		return
	}

	if err = helper.resolveAutoResources(req.Request.Context()); err != nil {
		klog.Errorf("Failed to resolve auto compute resources err=%v", err)
		api.HandleError(resp, req, err)
		return
	}

	name, extraLabels, err := helper.installV2Target()
	if err != nil {
		klog.Errorf("Failed to derive AM target err=%v", err)
		api.HandleError(resp, req, err)
		return
	}

	h.deliverInstallV2(req.Request.Context(), req, resp, name, appCfg, insReq, installV2Options{
		Token:        token,
		MarketSource: marketSource,
		ChartOwner:   chartOwner,
		ExtraLabels:  extraLabels,
		App:          app,
		RawAppName:   rawAppName,
	})
}

// installV2Options is the pre-decoded caller intent
// deliverInstallV2 acts on. Everything the translator needs to
// build the desired AM object shape lives here so the deliver step
// stays free of req/resp inspection.
type installV2Options struct {
	Token        string
	MarketSource string
	ChartOwner   string
	// ExtraLabels is the version-specific label set the helper's
	// installV2Target() returned. May be nil (V1/V2). Merged with
	// AppClonedFromKey + AppChartOwnerKey which are always stamped.
	ExtraLabels map[string]string
	// App is the AM Spec.AppName (path parameter, not RawAppName).
	App string
	// RawAppName is the original app name — differs from App only
	// on clones.
	RawAppName string
}

// deliverInstallV2 is the pure translator step of installV2 — the
// Create-or-Patch on Spec+labels+annotations. This is the ONLY v2
// handler that may Create an AM (all others assume the AM already
// exists). The ApplicationManager watch enqueues the Runner.
//
// Contract:
//
//   - appCfg MUST be a fully-rendered ApplicationConfig — the
//     translator writes it verbatim into Spec.Config; the pre-flight
//     pipeline in installV2 owns config resolution.
//   - insReq MUST NOT be nil.
//   - If NO AM exists at `name`: Create with the full desired
//     Spec+labels+annotations.
//   - If an AM already exists at `name`:
//   - Reject with 400 when AllowsOp(state, InstallOp)
//     is false. This is the reinstall gate: only terminal-
//     reinstallable states (Uninstalled, *Canceled, *Failed) may
//     accept a fresh install.
//   - Otherwise Patch spec + labels + annotations via MergeFrom
//     so Status is preserved (Runner-owned).
//   - Exactly one TriggerOp{InstallOp} dispatched on success.
//   - Response is 200 with UID + the durable request ID as OpID (the
//     Runner copies that same ID into Status when writing Pending).
func (h *Handler) deliverInstallV2(
	ctx context.Context,
	req *restful.Request,
	resp *restful.Response,
	name string,
	appCfg *appcfg.ApplicationConfig,
	insReq *api.InstallRequest,
	opts installV2Options,
) {
	// Persist the clone-origin + chart-owner on appCfg BEFORE
	// marshaling so the resulting AM.Spec.Config carries them for
	// the Runner and every downstream consumer (deployment /
	// Application labels, market analytics, resume flow). Matches
	// legacy applyAppMgr line-for-line.
	clonedFrom := clonedFromValue(insReq)
	appCfg.ClonedFrom = clonedFrom
	appCfg.ChartOwner = opts.ChartOwner

	labels := make(map[string]string, len(opts.ExtraLabels)+2)
	for k, v := range opts.ExtraLabels {
		labels[k] = v
	}
	// AppClonedFromKey is ALWAYS stamped (even when empty) so a
	// downstream label-selector on "has cloned-from key" works
	// uniformly. Same for AppChartOwnerKey.
	labels[constants.AppClonedFromKey] = clonedFrom
	labels[constants.AppChartOwnerKey] = opts.ChartOwner

	configJSON, err := json.Marshal(appCfg)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	images := insReq.Images
	if images == nil {
		images = make([]api.Image, 0)
	}
	imagesStr, _ := json.Marshal(images)

	annotations := map[string]string{
		api.AppTokenKey:                 opts.Token,
		api.AppRepoURLKey:               insReq.RepoURL,
		api.AppVersionKey:               appCfg.Version,
		api.AppMarketSourceKey:          opts.MarketSource,
		api.AppInstallSourceKey:         "app-service",
		constants.ApplicationTitleLabel: appCfg.Title,
		constants.ApplicationImageLabel: string(imagesStr),
	}
	annotations, opID := stampDurableOperationRequest(annotations)

	desiredSpec := appv1alpha1.ApplicationManagerSpec{
		AppName:      opts.App,
		RawAppName:   opts.RawAppName,
		AppNamespace: appCfg.Namespace,
		AppOwner:     appCfg.OwnerName,
		Config:       string(configJSON),
		Source:       insReq.Source.String(),
		Type:         appv1alpha1.Type(appCfg.Type),
		OpType:       appv1alpha1.InstallOp,
	}

	var existing appv1alpha1.ApplicationManager
	getErr := h.ctrlClient.Get(ctx, types.NamespacedName{Name: name}, &existing)
	switch {
	case apierrors.IsNotFound(getErr):
		// Fresh install: build the full AM and Create. Status is
		// left at its zero value; the Runner's first Reconcile
		// will observe InitialState + PendingActionQueue={TriggerOp}
		// and drive the bootstrap "" → Pending transition itself.
		fresh := &appv1alpha1.ApplicationManager{
			ObjectMeta: metav1.ObjectMeta{
				Name:        name,
				Labels:      labels,
				Annotations: annotations,
			},
			Spec: desiredSpec,
		}
		if err = h.ctrlClient.Create(ctx, fresh); err != nil {
			api.HandleError(resp, req, err)
			return
		}
	case getErr != nil:
		api.HandleError(resp, req, getErr)
		return
	default:
		// Reinstall gate. Terminal-reinstallable states
		// (Uninstalled, *Canceled, DownloadFailed) accept InstallOp per
		// the transition table; everything else must be rejected
		// so we don't clobber an in-flight lifecycle. An install
		// already written onto Spec but not yet copied into Status
		// still shows one of those states, so the pending request is
		// checked first.
		if blockIfRequestPending(resp, req, &existing, appv1alpha1.InstallOp) {
			return
		}
		if !domain.DefaultTable.AllowsOp(existing.Status.State, appv1alpha1.InstallOp) {
			api.HandleBadRequest(resp, req,
				explainOperationNotAllowed(existing.Status.State, appv1alpha1.InstallOp))
			return
		}
		// Patch spec + labels + annotations. Status is Runner-
		// owned so MergeFrom(&existing) preserves it — critical
		// because the reinstall path frequently runs against an
		// AM whose Status carries the final OpRecords from the
		// prior lifecycle (Uninstalled, InstallFailed, ...).
		patched := existing.DeepCopy()
		patched.Spec = desiredSpec
		if patched.Labels == nil {
			patched.Labels = make(map[string]string, len(labels))
		}
		for k, v := range labels {
			patched.Labels[k] = v
		}
		if patched.Annotations == nil {
			patched.Annotations = make(map[string]string, len(annotations))
		}
		for k, v := range annotations {
			patched.Annotations[k] = v
		}
		if err = h.ctrlClient.Patch(ctx, patched, client.MergeFrom(&existing)); err != nil {
			api.HandleError(resp, req, err)
			return
		}
	}

	resp.WriteEntity(api.InstallationResponse{
		Response: api.Response{Code: 200},
		Data:     api.InstallationResponseData{UID: opts.App, OpID: opID},
	})
}
