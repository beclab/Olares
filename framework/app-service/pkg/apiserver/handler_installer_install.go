package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/appstate"
	"github.com/beclab/Olares/framework/app-service/pkg/compute"
	"github.com/beclab/Olares/framework/app-service/pkg/compute/validation"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	"github.com/beclab/Olares/framework/app-service/pkg/utils/config"
	"github.com/beclab/Olares/framework/oac"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	sysv1alpha1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	"github.com/beclab/api/pkg/generated/clientset/versioned"

	"github.com/emicklei/go-restful/v3"
	"helm.sh/helm/v3/pkg/time"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type depRequest struct {
	Data []appcfg.Dependency `json:"data"`
}

// getAppClient is the seam used by call sites (e.g. checkAppNameConflict)
// that need a cluster-wide app.bytetrade.io clientset constructed from the
// in-cluster / KUBECONFIG context. It defaults to utils.GetClient; unit
// tests in this package override it to inject an appfake.NewSimpleClientset.
// The return type is the versioned.Interface so both the real
// *versioned.Clientset and the generated *fake.Clientset satisfy it.
var getAppClient = func() (versioned.Interface, error) {
	return utils.GetClient()
}

type installHelperIntf interface {
	getAdminUsers() (admin []string, isAdmin bool, err error)
	getInstalledApps() (installed bool, app []*v1alpha1.Application, err error)
	getAppConfig(adminUsers []string, marketSource string, isAdmin, appInstalled bool, installedApps []*v1alpha1.Application, chartVersion, selectedGpuType, originOwner string) (appConfig *appcfg.ApplicationConfig, err error)
	setAppConfig(req *api.InstallRequest, appName string)
	validate(bool, []*v1alpha1.Application) error
	lintChart() error
	setAppEnv(overrides []sysv1alpha1.AppEnvVar) error
	applyAppEnv(ctx context.Context) error
	resolveAutoResources(ctx context.Context) error
	applyApplicationManager(marketSource string) (opID string, err error)
	// checkAppNameConflict enforces cluster-wide exclusivity between shared
	// and per-user installs of the same Spec.AppName.
	checkAppNameConflict(ctx context.Context, newShared bool) error
}

var _ installHelperIntf = (*installHandlerHelper)(nil)
var _ installHelperIntf = (*installHandlerHelperV2)(nil)
var _ installHelperIntf = (*installHandlerHelperV3)(nil)

type installHandlerHelper struct {
	h          *Handler
	req        *restful.Request
	resp       *restful.Response
	app        string
	rawAppName string
	owner      string
	chartOwner string
	token      string
	insReq     *api.InstallRequest
	appConfig  *appcfg.ApplicationConfig
	chartPath  string
	// client is the generated app.bytetrade.io clientset. Typed as the
	// interface (not the concrete *versioned.Clientset) so unit tests can
	// inject a fake via appfake.NewSimpleClientset. The runtime install
	// dispatcher still constructs a concrete *versioned.Clientset and
	// assigns it here; the interface type is purely a substitutability hook.
	client               versioned.Interface
	validateClusterScope func(isAdmin bool, installedApps []*v1alpha1.Application) (err error)
}

type installHandlerHelperV2 struct {
	installHandlerHelper
}

// installHandlerHelperV3 handles the v3 install pipeline. The same helper
// covers both shared cluster-wide apps (options.shared: true) and per-user
// v3 apps (options.shared: false / absent); the shared / per-user split
// is decided per-method by inspecting h.appConfig.IsShared() once the
// chart has been loaded, so no manifest-peek is needed at dispatch time.
// v1 apps stay on installHandlerHelper.
type installHandlerHelperV3 struct {
	installHandlerHelper
}

func (h *Handler) install(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)
	var err error
	token, err := h.GetUserServiceAccountToken(req.Request.Context(), owner)
	if err != nil {
		klog.Error("Failed to get user service account token: ", err)
		api.HandleError(resp, req, err)
		return
	}

	marketSource := req.HeaderParameter(constants.MarketSource)
	klog.Infof("install: user: %v, source: %v", owner, marketSource)

	insReq := &api.InstallRequest{}
	err = req.ReadEntity(insReq)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	klog.Infof("insReq: %#v", insReq)
	if insReq.Source != api.Market && insReq.Source != api.Custom && insReq.Source != api.DevBox {
		api.HandleBadRequest(resp, req, fmt.Errorf("unsupported chart source: %s", insReq.Source))
		return
	}
	rawAppName := app
	if insReq.RawAppName != "" {
		rawAppName = insReq.RawAppName
	}
	klog.Infof("rawAppName: %s", rawAppName)
	chartVersion := ""
	chartOwner := ""
	if insReq.RawAppName != "" && !insReq.TemplateClone {
		chartVersion, chartOwner, err = h.getOriginChartVersion(rawAppName, owner)
		if err != nil {
			api.HandleBadRequest(resp, req, err)
			return
		}
	}
	// For uploaded (non-market) apps record the uploading user as the chart
	// owner so the chart source path survives even when the app owner is
	// normalized to the cluster owner (shared apps). Market apps keep
	// chartOwner empty and fall back to the installing user downstream.
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
	klog.Infof("chartVersion: %s, originOwner: %s", chartVersion, chartOwner)

	if err != nil {
		klog.Errorf("Failed to get api version err=%v", err)
		api.HandleBadRequest(resp, req, err)
		return
	}
	klog.Infof("apiVersion: %s", apiVersion)

	client, err := utils.GetClient()
	if err != nil {
		klog.Errorf("Failed to get client err=%v", err)
		api.HandleError(resp, req, err)
		return
	}

	var helper installHelperIntf
	switch apiVersion {
	case appcfg.V1:
		klog.Info("Using install handler helper for V1")
		h := &installHandlerHelper{
			h:          h,
			req:        req,
			resp:       resp,
			app:        app,
			rawAppName: rawAppName,
			owner:      owner,
			chartOwner: chartOwner,
			token:      token,
			insReq:     insReq,
			client:     client,
		}

		h.validateClusterScope = h._validateClusterScope

		helper = h
	case appcfg.V2:
		klog.Info("Using install handler helper for V2")
		h := &installHandlerHelperV2{
			installHandlerHelper: installHandlerHelper{
				h:          h,
				req:        req,
				resp:       resp,
				app:        app,
				rawAppName: rawAppName,
				owner:      owner,
				chartOwner: chartOwner,
				token:      token,
				insReq:     insReq,
				client:     client,
			},
		}

		h.validateClusterScope = h._validateClusterScope
		helper = h
	case appcfg.V3:
		h := &installHandlerHelperV3{
			installHandlerHelper: installHandlerHelper{
				h:          h,
				req:        req,
				resp:       resp,
				app:        app,
				rawAppName: rawAppName,
				owner:      owner,
				chartOwner: chartOwner,
				token:      token,
				insReq:     insReq,
				client:     client,
			}}

		// v3 reuses the v1 cluster-scope validation; the V3 helper's own
		// overrides decide which methods diverge from the v1 base based
		// on the shared flag.
		h.validateClusterScope = h._validateClusterScope
		helper = h
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

	// V2: get current user role and check if the app is installed by admin
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
		err = errors.New("only admin users can install shared apps")
		api.HandleForbidden(resp, req, err)
		return
	}

	if err = helper.checkAppNameConflict(req.Request.Context(), appCfg.Shared); err != nil {
		klog.Errorf("Failed to check app name conflict err=%v", err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	err = helper.lintChart()
	if err != nil {
		klog.Errorf("Failed to lint chart err=%v", err)
		return
	}

	if !appCfg.AllowMultipleInstall && insReq.RawAppName != "" ||
		(appCfg.AllowMultipleInstall && apiVersion == appcfg.V2) {
		klog.Errorf("app %s can not be clone", app)
		api.HandleBadRequest(resp, req, fmt.Errorf("app %s can not be clone", app))
		return
	}

	// When the caller didn't explicitly pick a compute mode, try to derive
	// one from the cluster + manifest. An empty SelectedGpuType is treated
	// as "no selection" (NOT as "cpu") — callers who genuinely want cpu
	// must pass "cpu" explicitly. See compute.AutoSelectMode for the rules.
	//
	// If the caller did pass something explicitly we leave it alone here;
	// the subsequent compute.AppInstallable check is what surfaces the
	// "you picked a mode the cluster doesn't have" error.
	if insReq.SelectedGpuType == "" {
		var chosen string
		chosen, err = compute.AutoSelectMode(req.Request.Context(), h.ctrlClient, appCfg)
		// More than one declared mode is runnable on this cluster — let the
		// caller pick by surfacing the full per-mode install plan with a
		// dedicated code, instead of failing the install outright.
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
		// Reload appCfg with the chosen mode so its Requirement reflects
		// the GPU-mode resource needs we'll actually schedule against.
		// Just patching SelectedGpuType in place isn't enough: for legacy
		// apps the first GetAppConfig call (with empty selectedGpu) ran
		// ResolveRequirement under cpu mode and dropped the GPU values
		// from appCfg.Requirement, so we need to re-parse from the
		// manifest with the chosen mode to recover them.
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
		Op:        v1alpha1.InstallOp,
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

	err = helper.setAppEnv(insReq.Envs)
	if err != nil {
		klog.Errorf("Failed to set app env err=%v", err)
		return
	}

	err = helper.validate(isAdmin, installedApps)
	if err != nil {
		klog.Errorf("Failed to validate app install request err=%v", err)
		return
	}
	if insReq.RawAppName != "" && insReq.Title != "" {
		helper.setAppConfig(insReq, app)
	}

	err = helper.applyAppEnv(req.Request.Context())
	if err != nil {
		klog.Errorf("Failed to apply app env err=%v", err)
		return
	}

	// For template apps that declare auto-compute ("-1") resource fields, the
	// concrete demand is only known once the chosen mode + just-applied appenv
	// are rendered into the chart. Resolve it now — before the
	// ApplicationManager is created — so the persisted application config
	// already carries concrete resource requirements and the entire downstream
	// pipeline (state machine, compute allocation, HAMI binding, resume) runs
	// the normal flow against real values.
	err = helper.resolveAutoResources(req.Request.Context())
	if err != nil {
		klog.Errorf("Failed to resolve auto compute resources err=%v", err)
		api.HandleError(resp, req, err)
		return
	}

	// create ApplicationManager
	opID, err := helper.applyApplicationManager(marketSource)
	if err != nil {
		klog.Errorf("Failed to apply application manager err=%v", err)
		return
	}

	resp.WriteEntity(api.InstallationResponse{
		Response: api.Response{Code: 200},
		Data:     api.InstallationResponseData{UID: app, OpID: opID},
	})
}

func (h *Handler) getOriginChartVersion(rawAppName, owner string) (string, string, error) {
	var ams v1alpha1.ApplicationManagerList
	err := h.ctrlClient.List(context.TODO(), &ams)
	if err != nil {
		return "", "", err
	}
	for _, am := range ams.Items {
		isShared := appcfg.IsShared(&am)
		if (am.Spec.AppName == rawAppName && am.Spec.AppOwner == owner) || (am.Spec.AppName == rawAppName && isShared) {
			// For shared apps the caller (current admin) may differ from
			// the user that originally uploaded the chart, and AppOwner is
			// normalized to the cluster owner — so we cannot key the
			// chart-repo lookup off AppOwner. GetChartOwner reads the
			// app.bytetrade.io/chart-owner label (the original uploader)
			// and only falls back to AppOwner for legacy/market AMs. v3
			// per-user apps fall through to the owner-equality branch above
			// just like v1 apps.
			return am.Annotations[api.AppVersionKey], appcfg.GetChartOwner(&am), nil
		}
	}
	return "", "", fmt.Errorf("rawApp %s not found", rawAppName)
}

func (h *installHandlerHelper) getAdminUsers() (admin []string, isAdmin bool, err error) {
	adminList, err := kubesphere.GetOwnerOrAdminList(h.req.Request.Context(), h.h.kubeConfig)
	if err != nil {
		api.HandleError(h.resp, h.req, err)
		return
	}

	for _, user := range adminList {
		admin = append(admin, user.Name)
		if user.Name == h.owner {
			isAdmin = true
		}
	}

	return
}

func (h *installHandlerHelper) validate(isAdmin bool, installedApps []*v1alpha1.Application) (err error) {
	unSatisfiedDeps, err := CheckDependencies(h.req.Request.Context(), h.appConfig.Dependencies, h.h.ctrlClient, h.owner, true)

	responseBadRequest := func(e error) {
		err = e
		api.HandleBadRequest(h.resp, h.req, err)
	}
	result, err := apputils.CheckCloneEntrances(h.h.ctrlClient, h.appConfig, h.insReq)
	if err != nil {
		api.HandleError(h.resp, h.req, err)
		return err
	}
	if result != nil {
		api.HandleFailedCheck(h.resp, api.CheckTypeAppEntrance, result)
		return fmt.Errorf("invalid entrance config, check result: %#v", result)
	}

	err = apputils.CheckDependencies2(h.req.Request.Context(), h.h.ctrlClient, h.appConfig.Dependencies, h.owner, true)
	if err != nil {
		klog.Errorf("Failed to check dependencies err=%v", err)
		responseBadRequest(FormatDependencyError(unSatisfiedDeps))
		return
	}

	err = apputils.CheckConflicts(h.req.Request.Context(), h.appConfig.Conflicts, h.owner)
	if err != nil {
		klog.Errorf("Failed to check installed conflict app err=%v", err)
		api.HandleBadRequest(h.resp, h.req, err)
		return
	}

	err = apputils.CheckTailScaleACLs(h.appConfig.TailScale.ACLs)
	if err != nil {
		klog.Errorf("Failed to check TailScale ACLs err=%v", err)
		api.HandleBadRequest(h.resp, h.req, err)
		return
	}

	err = apputils.CheckCfgFileVersion(h.appConfig.CfgFileVersion, config.MinCfgFileVersion)
	if err != nil {
		responseBadRequest(err)
		return
	}

	err = apputils.CheckNamespace(h.appConfig.Namespace)
	if err != nil {
		responseBadRequest(err)
		return
	}

	if !isAdmin && h.appConfig.OnlyAdmin {
		responseBadRequest(errors.New("only admin user can install this app"))
		return
	}

	if !isAdmin && h.appConfig.AppScope.ClusterScoped {
		responseBadRequest(errors.New("only admin user can create cluster level app"))
		return
	}

	if err = h.validateClusterScope(isAdmin, installedApps); err != nil {
		klog.Errorf("Failed to validate cluster scope err=%v", err)
		api.HandleBadRequest(h.resp, h.req, err)
		return
	}

	satisfied, err := apputils.CheckMiddlewareRequirement(h.req.Request.Context(), h.h.ctrlClient, h.appConfig.Middleware)
	if err != nil {
		api.HandleError(h.resp, h.req, err)
		return
	}
	if !satisfied {
		err = fmt.Errorf("middleware requirement can not be satisfied")
		h.resp.WriteHeaderAndEntity(http.StatusBadRequest, api.RequirementResp{
			Response: api.Response{Code: 400},
			Resource: "middleware",
			Message:  "middleware requirement can not be satisfied",
		})
		return
	}

	ret, err := apputils.CheckAppEnvs(h.req.Request.Context(), h.h.ctrlClient, h.appConfig.Envs, h.owner)
	if err != nil {
		klog.Errorf("Failed to check app environment config err=%v", err)
		api.HandleInternalError(h.resp, h.req, err)
		return
	}
	if ret != nil {
		api.HandleFailedCheck(h.resp, api.CheckTypeAppEnv, ret)
		return fmt.Errorf("Invalid appenv config, check result: %#v", ret)
	}

	return
}

// lintChart runs oac.Lint against the freshly-downloaded chart so chart-level
// authoring issues (folder layout, manifest cross-fields, helm render +
// workload integrity, hostPath rolling-update, resource namespace, container
// resource limits, Chart.yaml <-> manifest version match) are caught before
// the ApplicationManager object is created.
//
// Owner / admin reflect the actual scenario this install will run under, so
// helm-render-dependent checks exercise the same template branches the real
// install will take.
func (h *installHandlerHelper) lintChart() (err error) {
	if h.chartPath == "" {
		return nil
	}
	// ignore v2
	err = oac.Lint(h.chartPath)
	if err != nil {
		klog.Errorf("Failed to lint chart at %s err=%v", h.chartPath, err)
		api.HandleBadRequest(h.resp, h.req, err)
	}
	return
}

func (h *installHandlerHelper) _validateClusterScope(isAdmin bool, installedApp []*v1alpha1.Application) (err error) {
	for _, installedApp := range installedApp {
		if h.appConfig.AppScope.ClusterScoped && appcfg.IsClusterScoped(installedApp) {
			return errors.New("only one cluster scoped app can install in on cluster")
		}
	}

	return
}

func (h *installHandlerHelper) getInstalledApps() (installed bool, app []*v1alpha1.Application, err error) {
	var apps *v1alpha1.ApplicationList
	apps, err = h.client.AppV1alpha1().Applications().List(h.req.Request.Context(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list applications err=%v", err)
		api.HandleError(h.resp, h.req, err)
		return
	}

	for _, a := range apps.Items {
		if a.Spec.Name == h.app {
			installed = true
			app = append(app, &a)
		}
	}

	return
}

func (h *installHandlerHelper) getAppConfig(adminUsers []string, marketSource string, isAdmin, appInstalled bool, installedApps []*v1alpha1.Application, chartVersion, selectedGpuType, chartOwner string) (appConfig *appcfg.ApplicationConfig, err error) {
	var (
		admin                   string
		installAsAdmin          bool
		cluserAppInstalled      bool
		installedCluserAppOwner string
	)

	if appInstalled && len(installedApps) > 0 {
		for _, installedApp := range installedApps {
			klog.Infof("app: %s is already installed by %s", installedApp.Spec.Name, installedApp.Spec.Owner)
			// if the app is already installed, and the app's owner is admin,
			appOwner := installedApp.Spec.Owner
			if slices.Contains(adminUsers, appOwner) {
				// check the app is installed as cluster scope
				if appcfg.IsClusterScoped(installedApp) {
					cluserAppInstalled = true
					installedCluserAppOwner = appOwner
				}
			}
		}
	}

	switch {
	case cluserAppInstalled:
		admin = installedCluserAppOwner
		installAsAdmin = false
	case !isAdmin:
		if len(adminUsers) == 0 {
			klog.Errorf("No admin user found")
			err = fmt.Errorf("no admin user found")
			api.HandleBadRequest(h.resp, h.req, err)
			return
		}
		admin = adminUsers[0]
		installAsAdmin = false
	default:
		admin = h.owner
		installAsAdmin = true
	}

	var chartPath string
	appConfig, chartPath, err = apputils.GetAppConfig(h.req.Request.Context(), &apputils.ConfigOptions{
		App:          h.app,
		RawAppName:   h.rawAppName,
		Owner:        h.owner,
		RepoURL:      h.insReq.RepoURL,
		Version:      chartVersion,
		Admin:        admin,
		IsAdmin:      installAsAdmin,
		MarketSource: marketSource,
		SelectedGpu:  selectedGpuType,
		ChartOwner:   chartOwner,
	})
	if err != nil {
		klog.Errorf("Failed to get appconfig err=%v", err)
		api.HandleBadRequest(h.resp, h.req, err)
		return
	}

	h.appConfig = appConfig
	h.chartPath = chartPath

	return
}

func (h *installHandlerHelper) setAppConfig(req *api.InstallRequest, appName string) {
	h.appConfig.AppName = appName
	h.appConfig.RawAppName = appName
	if req.RawAppName != "" {
		h.appConfig.RawAppName = req.RawAppName
	}
	h.appConfig.Title = req.Title
	var appid string
	if userspace.IsSysApp(req.RawAppName) {
		appid = appName
	} else {
		appid = utils.Md5String(appName)[:8]
	}
	h.appConfig.AppID = appid

	entranceMap := make(map[string]string)
	for _, e := range req.Entrances {
		entranceMap[e.Name] = e.Title
	}

	for i, e := range h.appConfig.Entrances {
		h.appConfig.Entrances[i].Title = entranceMap[e.Name]
	}
	return
}

func (h *installHandlerHelper) applyApplicationManager(marketSource string) (opID string, err error) {
	name, _ := apputils.FmtAppMgrName(h.app, h.owner, h.appConfig.Namespace)
	return h.applyAppMgr(name, nil, marketSource)
}

// checkAppNameConflict enforces cluster-wide exclusivity between shared and
// per-user installs of the same app name:
//   - if a shared AM with the same Spec.AppName exists (any owner) → no other
//     user (and no admin) may install another shared or per-user variant of
//     that app name;
//   - if any per-user AM with the same Spec.AppName exists → an admin may not
//     install the shared variant before that per-user AM is removed.
//
// Same-owner same-type collisions are NOT handled here: their AM names are
// deterministically identical, so applyAppMgr's Get(name) + IsOperationAllowed
// path already covers them (patch on reinstallable states, reject otherwise).
// Per-user same-name across different owners is allowed by design.
func (h *installHandlerHelper) checkAppNameConflict(ctx context.Context, newShared bool) error {
	clientset, err := getAppClient()
	if err != nil {
		klog.Errorf("checkAppNameConflict: failed to get clientset %v", err)
		return err
	}
	apps, err := clientset.AppV1alpha1().ApplicationManagers().List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Errorf("checkAppNameConflict: failed to get appmgr list %v", err)
		return err
	}

	for i := range apps.Items {
		am := &apps.Items[i]
		if am.Spec.AppName != h.app {
			continue
		}
		if appstate.IsTerminalReinstallable(am.Status.State) {
			continue
		}
		existingShared := am.Labels[constants.AppSharedLabel] == constants.AppSharedTrue
		// Same-type collisions fall through to applyAppMgr's name-based path.
		if existingShared == newShared {
			continue
		}

		existingKind := "per-user"
		if existingShared {
			existingKind = "shared"
		}
		newKind := "per-user"
		if newShared {
			newKind = "shared"
		}
		err = fmt.Errorf("app %q is already installed as %s (owner=%q, state=%s); "+
			"uninstall it before installing as %s",
			h.app, existingKind, am.Spec.AppOwner, am.Status.State, newKind)
		return err
	}
	return nil
}

// clonedFromValue derives the app.bytetrade.io/app-cloned-from label value
// from the install request. A regular (non-clone) install has an empty
// RawAppName and yields "" so the label is stamped with an empty value. A
// clone sets RawAppName; TemplateClone distinguishes a clone from a template
// ("template") from a clone of an existing app ("app").
func clonedFromValue(insReq *api.InstallRequest) string {
	if insReq == nil || insReq.RawAppName == "" {
		return ""
	}
	if insReq.TemplateClone {
		return constants.AppClonedFromTemplate
	}
	return constants.AppClonedFromApp
}

// applyAppMgr is the shared create-or-patch implementation used by V1 / V2
// (via installHandlerHelper.applyApplicationManager) and V3
// (via installHandlerHelperV3.applyApplicationManager).
//
// `name` is the deterministic AM object name; `extraLabels` is merged into
// metadata.labels on both Create and Patch — V1/V2 pass nil; V3 passes the
// shared markers to mark the AM as a shared app. On top of extraLabels the
// app.bytetrade.io/app-cloned-from label is always stamped (empty value when
// the install is not a clone), so the "labels" key is always present in the
// merge patch.
func (h *installHandlerHelper) applyAppMgr(name string, extraLabels map[string]string, marketSource string) (opID string, err error) {
	// Record the clone origin on the appConfig so it is persisted in the AM
	// Config and later propagated to the deployment / Application via the
	// app.bytetrade.io/app-cloned-from label. Empty for non-clone installs.
	clonedFrom := clonedFromValue(h.insReq)
	h.appConfig.ClonedFrom = clonedFrom

	// Record the chart owner on the appConfig so it is persisted in the AM
	// Config and later propagated to the deployment / Application via the
	// app.bytetrade.io/chart-owner label. Empty for market installs.
	h.appConfig.ChartOwner = h.chartOwner

	// The cloned-from label is always stamped on the AM (empty value when the
	// install is not a clone). Merge it on top of any version-specific extra
	// labels (e.g. the v3 shared markers).
	labels := make(map[string]string, len(extraLabels)+1)
	for k, v := range extraLabels {
		labels[k] = v
	}
	labels[constants.AppClonedFromKey] = clonedFrom
	labels[constants.AppChartOwnerKey] = h.chartOwner

	config, err := json.Marshal(h.appConfig)
	if err != nil {
		api.HandleError(h.resp, h.req, err)
		return
	}
	images := make([]api.Image, 0)
	if len(h.insReq.Images) != 0 {
		images = h.insReq.Images
	}
	imagesStr, _ := json.Marshal(images)
	// For v1/v2 and v3+per-user apps appConfig.OwnerName == h.owner (the
	// install caller). For shared apps it is the cluster owner that
	// GetAppConfig resolved at chart load time — independent of which admin
	// is currently operating, so the AM stays addressed to a stable user
	// across multi-admin install / upgrade cycles.
	appOwner := h.appConfig.OwnerName
	appMgr := &v1alpha1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				api.AppTokenKey:                 h.token,
				api.AppRepoURLKey:               h.insReq.RepoURL,
				api.AppVersionKey:               h.appConfig.Version,
				api.AppMarketSourceKey:          marketSource,
				api.AppInstallSourceKey:         "app-service",
				constants.ApplicationTitleLabel: h.appConfig.Title,
				constants.ApplicationImageLabel: string(imagesStr),
			},
		},
		Spec: v1alpha1.ApplicationManagerSpec{
			AppName:      h.app,
			RawAppName:   h.rawAppName,
			AppNamespace: h.appConfig.Namespace,
			AppOwner:     appOwner,
			Config:       string(config),
			Source:       h.insReq.Source.String(),
			Type:         v1alpha1.Type(h.appConfig.Type),
			OpType:       v1alpha1.InstallOp,
		},
	}
	appMgr.Labels = labels

	a, err := h.client.AppV1alpha1().ApplicationManagers().Get(h.req.Request.Context(), name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			api.HandleError(h.resp, h.req, err)
			return
		}
		_, err = h.client.AppV1alpha1().ApplicationManagers().Create(h.req.Request.Context(), appMgr, metav1.CreateOptions{})
		if err != nil {
			api.HandleError(h.resp, h.req, err)
			return
		}
	} else {
		if !appstate.IsOperationAllowed(a.Status.State, v1alpha1.InstallOp) {
			err = appstate.ExplainOperationNotAllowed(a.Status.State, v1alpha1.InstallOp)
			api.HandleBadRequest(h.resp, h.req, err)
			return
		}
		metadataPatch := map[string]interface{}{
			"annotations": map[string]interface{}{
				api.AppTokenKey:                 h.token,
				api.AppRepoURLKey:               h.insReq.RepoURL,
				api.AppVersionKey:               h.appConfig.Version,
				api.AppMarketSourceKey:          marketSource,
				api.AppInstallSourceKey:         "app-service",
				constants.ApplicationTitleLabel: h.appConfig.Title,
			},
		}
		labelsPatch := make(map[string]interface{}, len(labels))
		for k, v := range labels {
			labelsPatch[k] = v
		}
		metadataPatch["labels"] = labelsPatch
		patchData := map[string]interface{}{
			"metadata": metadataPatch,
			"spec": map[string]interface{}{
				"opType":     v1alpha1.InstallOp,
				"config":     string(config),
				"source":     h.insReq.Source.String(),
				"appOwner":   appOwner,
				"rawAppName": h.rawAppName,
			},
		}
		var patchByte []byte
		patchByte, err = json.Marshal(patchData)
		if err != nil {
			api.HandleError(h.resp, h.req, err)
			return
		}
		_, err = h.client.AppV1alpha1().ApplicationManagers().Patch(h.req.Request.Context(), a.Name, types.MergePatchType, patchByte, metav1.PatchOptions{})
		if err != nil {
			api.HandleError(h.resp, h.req, err)
			return
		}

	}

	opID = strconv.FormatInt(time.Now().Unix(), 10)

	now := metav1.Now()
	status := v1alpha1.ApplicationManagerStatus{
		OpType:     v1alpha1.InstallOp,
		State:      v1alpha1.Pending,
		OpID:       opID,
		Message:    "waiting for install",
		Progress:   "0.00",
		StatusTime: &now,
		UpdateTime: &now,
		OpTime:     &now,
	}

	_, err = apputils.UpdateAppMgrStatus(name, status)
	if err != nil {
		api.HandleError(h.resp, h.req, err)
		return
	}
	return
}

func (h *installHandlerHelper) setAppEnv(overrides []sysv1alpha1.AppEnvVar) (err error) {
	defer func() {
		if err != nil {
			api.HandleBadRequest(h.resp, h.req, err)
		}
	}()
	if len(overrides) == 0 {
		return nil
	}
	if h.appConfig == nil {
		return fmt.Errorf("refuse to set app env on nil appconfig")
	}
	if len(h.appConfig.Envs) == 0 {
		return fmt.Errorf("refuse to set app env on app: %s with no declared envs", h.appConfig.AppName)
	}
	for _, override := range overrides {
		var found bool
		for i := range h.appConfig.Envs {
			if h.appConfig.Envs[i].EnvName == override.EnvName {
				found = true
				h.appConfig.Envs[i].Value = override.Value
				if override.ValueFrom != nil {
					h.appConfig.Envs[i].ValueFrom = override.ValueFrom
				}
			}
		}
		if !found {
			return fmt.Errorf("app env '%s' not found in app config", override.EnvName)
		}
	}
	return nil
}

func (h *installHandlerHelper) applyAppEnv(ctx context.Context) (err error) {
	_, err = apputils.ApplyAppEnv(ctx, h.h.ctrlClient, h.appConfig)
	if err != nil {
		api.HandleError(h.resp, h.req, err)
	}
	return
}

// resolveAutoResources resolves any auto-compute ("-1") resource field the app
// declares in its accelerator/resources matrix. It renders the chart once with
// the selected mode and the just-applied appenv (a side-effect-free dry-run),
// sums the workload container resource requests/limits, and rewrites the
// selected mode's sentinel fields with the concrete values. It then re-derives
// appConfig.Requirement so every downstream consumer sees a single resolved
// requirement.
//
// It is a no-op for apps that declare no sentinel, so it is safe to call
// unconditionally on the install / clone path.
func (h *installHandlerHelper) resolveAutoResources(ctx context.Context) (err error) {
	if h.appConfig == nil || !h.appConfig.HasAutoResource() {
		return nil
	}
	if h.chartPath == "" {
		return fmt.Errorf("cannot resolve auto compute resources: chart path is empty")
	}

	values, err := appinstaller.BuildBaseHelmValues(ctx, h.h.kubeConfig, h.appConfig, h.appConfig.OwnerName, true)
	if err != nil {
		return fmt.Errorf("build helm values for auto resource resolution: %w", err)
	}
	totals, err := utils.GetWorkloadResourcesFromChart(h.chartPath, values)
	if err != nil {
		return fmt.Errorf("sum workload resources for auto resource resolution: %w", err)
	}

	backfillAutoResourceMode(h.appConfig, totals)

	resolved, err := h.appConfig.ResolveRequirement(h.appConfig.SelectedGpuType)
	if err != nil {
		return fmt.Errorf("resolve requirement after auto resource backfill: %w", err)
	}
	h.appConfig.Requirement = *resolved
	return nil
}

// backfillAutoResourceMode replaces the auto-compute ("-1") fields of the
// selected resource mode (or every mode when no specific GPU type is selected)
// with the concrete totals summed from the rendered chart. Non-sentinel fields
// are left untouched so manifest-declared values still win.
func backfillAutoResourceMode(appConfig *appcfg.ApplicationConfig, totals utils.WorkloadResourceTotals) {
	// GPU memory is conventionally declared only as a limit; fall back across
	// requests/limits so a sentinel on either side resolves to a usable value.
	gpuReq := totals.RequestsGPUMem
	if gpuReq.IsZero() {
		gpuReq = totals.LimitsGPUMem
	}
	gpuLim := totals.LimitsGPUMem
	if gpuLim.IsZero() {
		gpuLim = totals.RequestsGPUMem
	}

	for i := range appConfig.Accelerator {
		mode := &appConfig.Accelerator[i]
		if appConfig.SelectedGpuType != "" && mode.Mode != appConfig.SelectedGpuType {
			continue
		}
		rr := &mode.ResourceRequirement
		if appcfg.IsAutoResource(rr.RequiredCPU) {
			rr.RequiredCPU = totals.RequestsCPU.String()
		}
		if appcfg.IsAutoResource(rr.LimitedCPU) {
			rr.LimitedCPU = totals.LimitsCPU.String()
		}
		if appcfg.IsAutoResource(rr.RequiredMemory) {
			rr.RequiredMemory = totals.RequestsMemory.String()
		}
		if appcfg.IsAutoResource(rr.LimitedMemory) {
			rr.LimitedMemory = totals.LimitsMemory.String()
		}
		if appcfg.IsAutoResource(rr.RequiredGPU) {
			rr.RequiredGPU = gpuMemMiBToByteString(gpuReq)
		}
		if appcfg.IsAutoResource(rr.LimitedGPU) {
			rr.LimitedGPU = gpuMemMiBToByteString(gpuLim)
		}
	}
}

// gpuMemMiBToByteString converts a summed pod nvidia.com/gpumem quantity (a
// plain integer count in MiB, the HAMi convention for the extended resource)
// into a byte-quantity string suitable for the resource mode's
// requiredGPUMemory/limitedGPUMemory fields, which the compute scheduler and
// the gpu-inject webhook interpret as bytes. e.g. 8000 (MiB) -> "8000Mi".
func gpuMemMiBToByteString(mib apiresource.Quantity) string {
	if mib.IsZero() {
		return mib.String()
	}
	return apiresource.NewQuantity(mib.Value()*1024*1024, apiresource.BinarySI).String()
}

func (h *installHandlerHelperV2) setAppConfig(req *api.InstallRequest, appName string) {
	return
}

func (h *installHandlerHelperV2) _validateClusterScope(isAdmin bool, installedApps []*v1alpha1.Application) (err error) {
	klog.Info("validate cluster scope for install handler v2")

	// check if subcharts has a client chart
	for _, subChart := range h.appConfig.SubCharts {
		if !subChart.Shared {
			if subChart.Name != h.app {
				err := fmt.Errorf("non-shared subchart must has the same name with the app, subchart name is %s but the main app is %s", subChart.Name, h.app)
				klog.Error(err)
				api.HandleBadRequest(h.resp, h.req, err)
				return err
			}
		}
	}

	// in V2, we do not check cluster scope here, the cluster scope app
	// will be checked if the cluster part is installed by another user in the installing phase

	return nil
}

func (h *installHandlerHelperV2) getAppConfig(adminUsers []string, marketSource string, isAdmin, appInstalled bool, installedApps []*v1alpha1.Application, chartVersion, selectedGpuType, chartOwner string) (appConfig *appcfg.ApplicationConfig, err error) {
	klog.Info("get app config for install handler v2")

	var (
		admin string
	)

	if isAdmin {
		admin = h.owner
	} else {
		if len(adminUsers) == 0 {
			klog.Errorf("No admin user found")
			err = fmt.Errorf("no admin user found")
			api.HandleBadRequest(h.resp, h.req, err)
			return
		}
		admin = adminUsers[0]
	}
	var chartPath string
	appConfig, chartPath, err = apputils.GetAppConfig(h.req.Request.Context(), &apputils.ConfigOptions{
		App:          h.app,
		RawAppName:   h.rawAppName,
		Owner:        h.owner,
		RepoURL:      h.insReq.RepoURL,
		Version:      chartVersion,
		Token:        h.token,
		Admin:        admin,
		MarketSource: marketSource,
		IsAdmin:      isAdmin,
		SelectedGpu:  selectedGpuType,
		ChartOwner:   chartOwner,
	})
	if err != nil {
		klog.Errorf("Failed to get appconfig err=%v", err)
		api.HandleBadRequest(h.resp, h.req, err)
		return
	}

	h.appConfig = appConfig
	h.chartPath = chartPath

	return
}

// ----- v3 helper overrides -----

// getAdminUsers returns the admin user list. Shared apps reject non-admin
// callers with a 403 (admin-only lifecycle); v3 per-user apps fall through
// to the v1 helper, which doesn't gate on admin status here.
func (h *installHandlerHelperV3) getAdminUsers() (admin []string, isAdmin bool, err error) {
	admin, isAdmin, err = h.installHandlerHelper.getAdminUsers()
	if err != nil {
		return
	}
	return
}

// getInstalledApps returns no installed apps for shared apps so the v1
// "install on behalf of admin / install as the cluster owner" branching
// in installHandlerHelper.getAppConfig is bypassed — the V3 path always
// resolves to the cluster owner. v3 per-user apps fall through to the v1
// helper so the regular cluster-scope / clone duplication checks fire.
func (h *installHandlerHelperV3) getInstalledApps() (installed bool, apps []*v1alpha1.Application, err error) {
	return
}

// getAppConfig loads the manifest. For shared apps the chart is loaded
// with admin context (the caller is admin — gated above) and the
// namespace / owner are forced to the deterministic shared values. For
// v3 per-user apps the v1 install branching applies as-is so the app is
// installed in the regular per-user namespace under the installing user.
func (h *installHandlerHelperV3) getAppConfig(adminUsers []string, marketSource string, isAdmin, appInstalled bool, installedApps []*v1alpha1.Application, chartVersion, selectedGpuType, chartOwner string) (appConfig *appcfg.ApplicationConfig, err error) {
	var chartPath string
	appConfig, chartPath, err = apputils.GetAppConfig(h.req.Request.Context(), &apputils.ConfigOptions{
		App:          h.app,
		RawAppName:   h.rawAppName,
		Owner:        h.owner,
		RepoURL:      h.insReq.RepoURL,
		Version:      chartVersion,
		Admin:        h.owner,
		IsAdmin:      true,
		MarketSource: marketSource,
		SelectedGpu:  selectedGpuType,
		ChartOwner:   chartOwner,
	})
	if err != nil {
		klog.Errorf("Failed to get appconfig err=%v", err)
		api.HandleBadRequest(h.resp, h.req, err)
		return
	}
	h.appConfig = appConfig
	h.chartPath = chartPath
	return
}

// applyApplicationManager creates / updates the AM. Shared apps land at
// the cluster-wide V3AppMgrName(app) with both the api-version=v3 schema
// label and the shared=true marker so listers / proxy / NetworkPolicy /
// NATS fan-out can identify them. v3 per-user apps go to the regular
// per-user FmtAppMgrName(app, owner, "") with only the api-version=v3
// schema label — they are NOT shared and must not trigger any
// shared-app behavior downstream.
func (h *installHandlerHelperV3) applyApplicationManager(marketSource string) (opID string, err error) {
	name, _ := apputils.FmtAppMgrName(h.app, h.owner, h.appConfig.Namespace)
	labels := map[string]string{
		constants.AppApiVersionLabel: constants.AppVersionV3,
		constants.AppSharedLabel: func() string {
			if h.appConfig.Shared {
				return "true"
			}
			return "false"
		}(),
	}
	return h.applyAppMgr(name, labels, marketSource)
}

func (h *Handler) isDeployAllowed(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)

	// Use ResolveAppMgrName so the "can I deploy?" check observes the
	// shared cluster-wide AM (if any) and not a phantom per-user name that
	// nobody ever created. Without this an admin could think a shared
	// install slot is free just because no per-user AM exists.
	name, _, err := apputils.ResolveAppMgrName(req.Request.Context(), app, owner)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	var am v1alpha1.ApplicationManager
	err = h.ctrlClient.Get(req.Request.Context(), types.NamespacedName{Name: name}, &am)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			api.HandleError(resp, req, err)
			return
		}
		resp.WriteEntity(api.CanDeployResponse{
			Response: api.Response{Code: 200},
			Data: api.CanDeployResponseData{
				CanOp: true,
			},
		})
		return
	}
	if am.Status.State == v1alpha1.Uninstalled {
		resp.WriteEntity(api.CanDeployResponse{
			Response: api.Response{Code: 200},
			Data: api.CanDeployResponseData{
				CanOp: true,
			},
		})
		return
	}

	canOp := false
	if appstate.IsOperationAllowed(am.Status.State, v1alpha1.UninstallOp) {
		canOp = true
	}
	resp.WriteEntity(api.CanDeployResponse{
		Response: api.Response{Code: 200},
		Data: api.CanDeployResponseData{
			CanOp: canOp,
		},
	})
}
