package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	"github.com/beclab/Olares/framework/app-service/pkg/workflowinstaller"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/emicklei/go-restful/v3"
	"github.com/go-resty/resty/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// loadAuthorizedLifecycleAM is the shared prelude for mutating lifecycle
// handlers (uninstall, suspend, resume, applyenv, cancel). It:
//  1. Resolves the ApplicationManager name via apputils.ResolveAppMgrName so
//     admins can target a v3 AM regardless of who installed it.
//  2. Loads that ApplicationManager from the API.
//  3. If the AM is v3, requires cluster admin; otherwise non-v3
//     AMs remain scoped by ResolveAppMgrName/FmtAppMgrName to the caller.
//
// On any error or rejection it writes the response and returns ok=false; the
// caller should simply `return`.
func (h *Handler) loadAuthorizedLifecycleAM(ctx context.Context, req *restful.Request, resp *restful.Response, app, owner string) (name string, am *v1alpha1.ApplicationManager, ok bool) {
	resolved, isV3, err := apputils.ResolveAppMgrName(ctx, app, owner)
	if err != nil {
		api.HandleError(resp, req, err)
		return "", nil, false
	}
	var got v1alpha1.ApplicationManager
	if err := h.ctrlClient.Get(ctx, types.NamespacedName{Name: resolved}, &got); err != nil {
		api.HandleError(resp, req, err)
		return "", nil, false
	}
	// Treat the AM as shared if either the resolver picked the v3 name OR
	// the AM carries the scope label (defense in depth in case of weird
	// install state).
	if isV3 || appcfg.IsV3(&got) {
		isAdmin, ierr := kubesphere.IsAdmin(ctx, h.kubeConfig, owner)
		if ierr != nil {
			api.HandleError(resp, req, ierr)
			return "", nil, false
		}
		if !isAdmin {
			api.HandleForbidden(resp, req, fmt.Errorf("only admin users can manage v3 / shared app %q", app))
			return "", nil, false
		}
	}
	return resolved, &got, true
}

// listVisibilityCtx is the per-request context used by list/get endpoints to
// decide whether an Application / ApplicationManager is visible to `viewer`.
//
// Visibility model:
//   - v3 / shared apps: visible to ALL authenticated users (admin or not);
//     any user may open the app. Lifecycle handlers
//     (install/uninstall/upgrade/...) enforce admin-only management.
//   - v1 / v2 apps: legacy owner-by-namespace check is preserved unchanged
//     (only the installer sees them in their own list).
type listVisibilityCtx struct {
	Viewer string
}

// newListVisibilityCtx is kept for symmetry and future per-request caching
// (e.g. admin role memoisation if we ever need it elsewhere). It currently
// makes no API calls.
func (h *Handler) newListVisibilityCtx(_ context.Context, viewer string) (*listVisibilityCtx, error) {
	return &listVisibilityCtx{Viewer: viewer}, nil
}

// VisibleSharedApp always returns true — see the visibility model above.
// Kept as a method so all visibility decisions funnel through this file.
func (v *listVisibilityCtx) VisibleSharedApp(_ string) bool { return true }

// VisibleAM returns true if the AM is visible to the cached viewer.
func (v *listVisibilityCtx) VisibleAM(am *v1alpha1.ApplicationManager) bool {
	if am == nil {
		return false
	}
	if appcfg.IsV3(am) {
		return true
	}
	return am.Spec.AppOwner == v.Viewer
}

// VisibleApp mirrors VisibleAM for *v1alpha1.Application objects.
func (v *listVisibilityCtx) VisibleApp(a *v1alpha1.Application) bool {
	if a == nil {
		return false
	}
	if appcfg.IsV3(a) {
		return true
	}
	return a.Spec.Owner == v.Viewer
}

// requireAdmin returns true iff `caller` is owner/admin. On rejection it
// writes the 403/5xx response and returns false so the caller can simply
// `return`.
func (h *Handler) requireAdmin(req *restful.Request, resp *restful.Response, caller string) bool {
	isAdmin, err := kubesphere.IsAdmin(req.Request.Context(), h.kubeConfig, caller)
	if err != nil {
		api.HandleError(resp, req, err)
		return false
	}
	if !isAdmin {
		api.HandleForbidden(resp, req, fmt.Errorf("only admin users can perform this operation"))
		return false
	}
	return true
}

// gateSharedAppWrite enforces admin-only mutation on v3 Applications.
// v1/v2 apps pass through unchanged. On rejection it writes
// the 403 response and returns ok=false so the caller can simply `return`.
func (h *Handler) gateSharedAppWrite(req *restful.Request, resp *restful.Response, app *v1alpha1.Application) (ok bool) {
	if !appcfg.IsV3(app) {
		return true
	}
	caller, _ := req.Attribute(constants.UserContextAttribute).(string)
	return h.requireAdmin(req, resp, caller)
}

// UpdateAppState update applicationmanager state, message
func (h *Handler) UpdateAppState(ctx context.Context, name string, state v1alpha1.ApplicationManagerState, message string) error {
	var appMgr v1alpha1.ApplicationManager
	key := types.NamespacedName{Name: name}
	err := h.ctrlClient.Get(ctx, key, &appMgr)
	if err != nil {
		return err
	}
	appMgrCopy := appMgr.DeepCopy()
	now := metav1.Now()
	appMgr.Status.State = state
	appMgr.Status.Message = message
	appMgr.Status.StatusTime = &now
	appMgr.Status.UpdateTime = &now
	err = h.ctrlClient.Status().Patch(ctx, &appMgr, client.MergeFrom(appMgrCopy))
	return err
}

func (h *Handler) checkDependencies(req *restful.Request, resp *restful.Response) {
	owner := req.Attribute(constants.UserContextAttribute) // get owner from request token
	var err error
	depReq := depRequest{}
	err = req.ReadEntity(&depReq)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	unSatisfiedDeps, _ := apputils.CheckDependencies(req.Request.Context(), h.ctrlClient, depReq.Data, owner.(string), true)
	klog.Infof("Check application dependencies unSatisfiedDeps=%v", unSatisfiedDeps)

	data := make([]api.DependenciesRespData, 0)
	for _, dep := range unSatisfiedDeps {
		data = append(data, api.DependenciesRespData{
			Name:    dep.Name,
			Version: dep.Version,
			Type:    dep.Type,
		})
	}
	resp.WriteEntity(api.DependenciesResp{
		Response: api.Response{Code: 200},
		Data:     data,
	})
}

func (h *Handler) cleanRecommendFeedData(name, owner string) error {
	knowledgeAPI := fmt.Sprintf("http://knowledge-base-api.user-system-%s:3010", owner)
	feedAPI := knowledgeAPI + "/knowledge/feed/algorithm/" + name

	client := resty.New()
	response, err := client.R().Get(feedAPI)
	if err != nil {
		return err
	}
	if response.StatusCode() != http.StatusOK {
		klog.Errorf("Failed to get knowledge feed list status=%s body=%s", response.Status(), response.String())
		return errors.New(response.Status())
	}
	var ret workflowinstaller.KnowledgeAPIResp
	err = json.Unmarshal(response.Body(), &ret)
	if err != nil {
		return err
	}
	feedUrls := ret.Data
	klog.Info("Start to clean recommend feed data ", feedAPI, len(feedUrls))
	if len(feedUrls) > 0 {
		limit := 10
		removeClient := resty.New()
		for i := 0; i*limit < len(feedUrls); i++ {
			start := i * limit
			end := start + limit
			if end > len(feedUrls) {
				end = len(feedUrls)
			}
			removeList := feedUrls[start:end]
			reqData := workflowinstaller.KnowledgeFeedDelReq{FeedUrls: removeList}
			removeBody, _ := json.Marshal(reqData)
			res, _ := removeClient.SetTimeout(5*time.Second).R().SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
				SetBody(removeBody).Delete(feedAPI)

			if res.StatusCode() == http.StatusOK {
				klog.Info("Delete feed success: ", i, len(removeList))
			} else {
				klog.Errorf("Failed to clean recommend feed data err=%s", string(res.Body()))
			}
		}
	}
	klog.Info("Delete entry success page: ", name, len(feedUrls))
	return nil
}
