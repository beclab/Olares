package apiserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	"github.com/beclab/Olares/framework/app-service/pkg/utils/config"
	"github.com/beclab/Olares/framework/oac"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/klog/v2"
)

type upgradeHelperIntf interface {
	getAdminUsers() (admin []string, isAdmin bool, err error)
	getAppConfig(prevCfg *appcfg.ApplicationConfig, adminUsers []string, marketSource string, isAdmin bool) (appConfig *appcfg.ApplicationConfig, err error)
	validate() error
	lintChart() error
	applyAppEnv(ctx context.Context) error
	setAndEncodingAppCofnig(prevCfg *appcfg.ApplicationConfig) (string, error)
}

var _ upgradeHelperIntf = (upgradeHelperIntf)(nil)
var _ upgradeHelperIntf = (upgradeHelperIntf)(nil)

type upgradeHandlerHelper struct {
	h          *Handler
	req        *restful.Request
	resp       *restful.Response
	owner      string
	app        string
	rawAppName string
	request    *api.UpgradeRequest
	token      string
	appConfig  *appcfg.ApplicationConfig
	chartPath  string
}

type upgradeHandlerHelperV2 struct {
	*upgradeHandlerHelper
}

// upgradeHandlerHelperV3 covers both v3 shared apps and v3 per-user apps.
// For shared apps it treats the caller as admin (admin-only lifecycle;
// the gate above already verified that) and overrides getAppConfig to
// pass IsAdmin=true / Admin=h.owner so GetAppConfig resolves the cluster
// owner. For v3 per-user apps it falls through to the v1 upgrade flow
// unchanged.
type upgradeHandlerHelperV3 struct {
	*upgradeHandlerHelper
}

func (h *upgradeHandlerHelper) getAdminUsers() (admins []string, isAdmin bool, err error) {
	adminList, err := kubesphere.GetOwnerOrAdminList(h.req.Request.Context(), h.h.kubeConfig)
	if err != nil {
		api.HandleError(h.resp, h.req, err)
		return
	}
	for _, user := range adminList {
		admins = append(admins, user.Name)
	}
	isAdmin, err = kubesphere.IsAdmin(h.req.Request.Context(), h.h.kubeConfig, h.owner)
	if err != nil {
		api.HandleError(h.resp, h.req, err)
		return
	}

	return
}

func (h *upgradeHandlerHelper) getAppConfig(prevCfg *appcfg.ApplicationConfig, adminUsers []string, marketSource string, _ bool) (*appcfg.ApplicationConfig, error) {
	var admin string
	if !prevCfg.AppScope.ClusterScoped {
		// installed as non-admin
		admin = adminUsers[0]
		if len(adminUsers) > 1 {
			for _, user := range adminUsers {
				if user != h.owner {
					admin = user
					break
				}
			}
		}
	} else {
		admin = h.owner
	}

	appConfig, chartPath, err := apputils.GetAppConfig(h.req.Request.Context(), &apputils.ConfigOptions{
		App:          h.app,
		RawAppName:   h.rawAppName,
		Owner:        h.owner,
		RepoURL:      h.request.RepoURL,
		Version:      h.request.Version,
		Token:        h.token,
		Admin:        admin,
		MarketSource: marketSource,
		IsAdmin:      prevCfg.AppScope.ClusterScoped,
		SelectedGpu:  prevCfg.SelectedGpuType,
		ChartOwner:   prevCfg.ChartOwner,
	})
	if err != nil {
		api.HandleError(h.resp, h.req, err)
		return nil, err
	}

	h.appConfig = appConfig
	h.chartPath = chartPath
	return appConfig, nil
}

func (h *upgradeHandlerHelper) validate() error {
	if h.appConfig == nil {
		return fmt.Errorf("application config is nil")
	}

	err := apputils.CheckTailScaleACLs(h.appConfig.TailScale.ACLs)
	if err != nil {
		api.HandleError(h.resp, h.req, err)
		return err
	}

	if !utils.MatchVersion(h.appConfig.CfgFileVersion, config.MinCfgFileVersion) {
		api.HandleBadRequest(h.resp, h.req, fmt.Errorf("olaresManifest.version must %s", config.MinCfgFileVersion))
		return err
	}

	return nil
}

// lintChart runs oac.Lint against the freshly-downloaded target chart so
// chart-level authoring issues (folder layout, manifest cross-fields, helm
// render + workload integrity, hostPath rolling-update, resource namespace,
// container resource limits, Chart.yaml <-> manifest version match) are
// caught before the upgrade is kicked off on the ApplicationManager.
//
// Owner / admin reflect the actual scenario this upgrade will run under, so
// helm-render-dependent checks exercise the same template branches the real
// upgrade will take.
func (h *upgradeHandlerHelper) lintChart() error {
	if h.chartPath == "" {
		return nil
	}
	// ignore v2
	err := oac.Lint(h.chartPath)
	if err != nil {
		klog.Errorf("Failed to lint chart at %s err=%v", h.chartPath, err)
		api.HandleBadRequest(h.resp, h.req, err)
		return err
	}
	return nil
}

func (h *upgradeHandlerHelper) setAndEncodingAppCofnig(prevCfg *appcfg.ApplicationConfig) (string, error) {
	// cloned app
	if h.app != h.rawAppName {
		h.appConfig.Title = prevCfg.Title
		entranceTitleMap := make(map[string]string)
		for _, e := range prevCfg.Entrances {
			if e.Invisible {
				continue
			}
			entranceTitleMap[e.Name] = e.Title
		}
		for i, e := range h.appConfig.Entrances {
			if e.Invisible {
				continue
			}
			if title, ok := entranceTitleMap[e.Name]; ok {
				h.appConfig.Entrances[i].Title = title
			}
		}
	}

	prevPortsMap := apputils.BuildPrevPortsMap(prevCfg)

	// Set expose ports for upgrade, preserving existing ports with same name
	err := apputils.SetExposePorts(context.TODO(), h.appConfig, prevPortsMap)
	if err != nil {
		klog.Errorf("set expose ports failed %v", err)
		return "", err
	}

	encoding, err := json.Marshal(h.appConfig)
	if err != nil {
		klog.Errorf("Failed to marshal app config err=%v", err)
		api.HandleError(h.resp, h.req, err)
		return "", err
	}
	return string(encoding), nil
}

func (h *upgradeHandlerHelper) applyAppEnv(ctx context.Context) (err error) {
	_, err = apputils.ApplyAppEnv(ctx, h.h.ctrlClient, h.appConfig)
	if err != nil {
		api.HandleError(h.resp, h.req, err)
	}
	return
}

func (h *upgradeHandlerHelperV3) getAppConfig(prevCfg *appcfg.ApplicationConfig, adminUsers []string, marketSource string, isAdmin bool) (*appcfg.ApplicationConfig, error) {
	klog.Info("Getting app config for V3 shared app")
	appConfig, chartPath, err := apputils.GetAppConfig(h.req.Request.Context(), &apputils.ConfigOptions{
		App:          h.app,
		Owner:        prevCfg.OwnerName,
		RepoURL:      h.request.RepoURL,
		Version:      h.request.Version,
		Token:        h.token,
		Admin:        h.owner,
		MarketSource: marketSource,
		IsAdmin:      true,
		RawAppName:   h.rawAppName,
		SelectedGpu:  prevCfg.SelectedGpuType,
		ChartOwner:   prevCfg.ChartOwner,
	})
	if err != nil {
		api.HandleError(h.resp, h.req, err)
		return nil, err
	}
	h.appConfig = appConfig
	h.chartPath = chartPath
	return appConfig, nil
}

func (h *upgradeHandlerHelperV2) getAppConfig(prevCfg *appcfg.ApplicationConfig, adminUsers []string, marketSource string, isAdmin bool) (*appcfg.ApplicationConfig, error) {
	klog.Info("Getting app config for V2")
	if len(adminUsers) == 0 {
		err := fmt.Errorf("no admin users found")
		klog.Error(err)
		api.HandleError(h.resp, h.req, err)
		return nil, err
	}

	var admin string
	if isAdmin {
		admin = h.owner
	} else {
		admin = adminUsers[0]
	}

	appConfig, chartPath, err := apputils.GetAppConfig(h.req.Request.Context(), &apputils.ConfigOptions{
		App:          h.app,
		Owner:        h.owner,
		RepoURL:      h.request.RepoURL,
		Version:      h.request.Version,
		Token:        h.token,
		Admin:        admin,
		MarketSource: marketSource,
		IsAdmin:      isAdmin,
		RawAppName:   h.rawAppName,
		SelectedGpu:  prevCfg.SelectedGpuType,
	})
	if err != nil {
		api.HandleError(h.resp, h.req, err)
		return nil, err
	}

	h.appConfig = appConfig
	h.chartPath = chartPath

	return appConfig, nil
}
