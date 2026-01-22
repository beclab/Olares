package apiserver

import (
	"context"
	"encoding/json"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/appstate"
	"github.com/beclab/Olares/framework/app-service/pkg/client/clientset"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"

	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"

	"github.com/emicklei/go-restful/v3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (h *Handler) status(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)

	name, err := apputils.FmtAppMgrName(app, owner, "")
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	var am v1alpha1.ApplicationManager
	e := h.ctrlClient.Get(req.Request.Context(), types.NamespacedName{Name: name}, &am)
	if e != nil {
		if apierrors.IsNotFound(e) {
			api.HandleNotFound(resp, req, e)
			return
		}
		api.HandleError(resp, req, e)
		return
	}
	now := metav1.Now()
	sts := appinstaller.Status{
		Name:              am.Spec.AppName,
		AppID:             v1alpha1.AppName(am.Spec.AppName).GetAppID(),
		Namespace:         am.Spec.AppNamespace,
		CreationTimestamp: now,
		Source:            am.Spec.Source,
		AppStatus: v1alpha1.ApplicationStatus{
			State:      am.Status.State.String(),
			Progress:   am.Status.Progress,
			StatusTime: &now,
			UpdateTime: &now,
		},
	}

	resp.WriteAsJson(sts)
}

func (h *Handler) appsStatus(req *restful.Request, resp *restful.Response) {
	owner := req.Attribute(constants.UserContextAttribute).(string)
	isSysApp := req.QueryParameter("issysapp")
	state := req.QueryParameter("state")
	ss := make([]string, 0)
	if state != "" {
		ss = strings.Split(state, "|")
	}
	all := make([]string, 0)
	for _, a := range appstate.All {
		all = append(all, a.String())
	}
	stateSet := sets.NewString(all...)
	if len(ss) > 0 {
		stateSet = sets.String{}
	}
	for _, s := range ss {
		stateSet.Insert(s)
	}

	// filter by application's owner
	filteredApps := make([]appinstaller.Status, 0)

	appAms, err := h.appmgrLister.List(labels.Everything())
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	for _, am := range appAms {
		if am.Spec.AppOwner == owner {
			if !stateSet.Has(am.Status.State.String()) {
				continue
			}
			if len(isSysApp) > 0 && isSysApp == "true" && !userspace.IsSysApp(am.Spec.AppName) {
				continue
			}
			now := metav1.Now()
			status := appinstaller.Status{
				Name:              am.Spec.AppName,
				AppID:             v1alpha1.AppName(am.Spec.AppName).GetAppID(),
				Namespace:         am.Spec.AppNamespace,
				CreationTimestamp: now,
				Source:            am.Spec.Source,
				AppStatus: v1alpha1.ApplicationStatus{
					State:      am.Status.State.String(),
					Progress:   am.Status.Progress,
					StatusTime: &now,
					UpdateTime: &now,
				},
			}

			filteredApps = append(filteredApps, status)
		}
	}

	// sort by create time desc
	sort.Slice(filteredApps, func(i, j int) bool {
		return filteredApps[j].CreationTimestamp.Before(&filteredApps[i].CreationTimestamp)
	})

	resp.WriteAsJson(map[string]interface{}{"result": filteredApps})
}

func (h *Handler) operate(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)

	var am v1alpha1.ApplicationManager
	name, err := apputils.FmtAppMgrName(app, owner, "")
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	err = h.ctrlClient.Get(req.Request.Context(), types.NamespacedName{Name: name}, &am)
	if err != nil {
		if apierrors.IsNotFound(err) {
			api.HandleNotFound(resp, req, err)
			return
		}
		api.HandleError(resp, req, err)
		return
	}
	operate := appinstaller.Operate{
		AppName:           am.Spec.AppName,
		AppNamespace:      am.Spec.AppNamespace,
		AppOwner:          am.Spec.AppOwner,
		OpType:            am.Status.OpType,
		OpID:              am.Status.OpID,
		ResourceType:      am.Spec.Type.String(),
		State:             am.Status.State,
		Message:           am.Status.Message,
		CreationTimestamp: am.CreationTimestamp,
		Source:            am.Spec.Source,
		Progress:          am.Status.Progress,
	}

	resp.WriteAsJson(operate)
}

func (h *Handler) appsOperate(req *restful.Request, resp *restful.Response) {
	owner := req.Attribute(constants.UserContextAttribute).(string)

	ams, err := h.appmgrLister.List(labels.Everything())
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	// filter by application's owner
	filteredOperates := make([]appinstaller.Operate, 0)
	for _, am := range ams {
		if am.Spec.Type != v1alpha1.App {
			continue
		}

		if am.Spec.AppOwner == owner {
			operate := appinstaller.Operate{
				AppName:           am.Spec.AppName,
				AppNamespace:      am.Spec.AppNamespace,
				AppOwner:          am.Spec.AppOwner,
				State:             am.Status.State,
				OpType:            am.Status.OpType,
				OpID:              am.Status.OpID,
				ResourceType:      am.Spec.Type.String(),
				Message:           am.Status.Message,
				CreationTimestamp: am.CreationTimestamp,
				Source:            am.Spec.Source,
				Progress:          am.Status.Progress,
			}
			filteredOperates = append(filteredOperates, operate)
		}
	}

	// sort by create time desc
	sort.Slice(filteredOperates, func(i, j int) bool {
		return filteredOperates[j].CreationTimestamp.Before(&filteredOperates[i].CreationTimestamp)
	})

	resp.WriteAsJson(map[string]interface{}{"result": filteredOperates})
}

func (h *Handler) operateHistory(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)

	var am v1alpha1.ApplicationManager
	name, err := apputils.FmtAppMgrName(app, owner, "")
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	key := types.NamespacedName{Name: name}
	err = h.ctrlClient.Get(req.Request.Context(), key, &am)

	if err != nil {
		if apierrors.IsNotFound(err) {
			api.HandleNotFound(resp, req, err)
			return
		}
		api.HandleError(resp, req, err)
		return
	}
	ops := make([]appinstaller.OperateHistory, 0, len(am.Status.OpRecords))
	for _, r := range am.Status.OpRecords {
		op := appinstaller.OperateHistory{
			AppName:      am.Spec.AppName,
			AppNamespace: am.Spec.AppNamespace,
			AppOwner:     am.Spec.AppOwner,
			ResourceType: am.Spec.Type.String(),
			OpRecord: v1alpha1.OpRecord{
				OpType:    r.OpType,
				OpID:      r.OpID,
				Message:   r.Message,
				Source:    r.Source,
				Version:   r.Version,
				Status:    r.Status,
				StateTime: r.StateTime,
			},
		}
		ops = append(ops, op)
	}

	resp.WriteAsJson(map[string]interface{}{"result": ops})
}

func (h *Handler) allAppManagers(req *restful.Request, resp *restful.Response) {
	ams, err := h.appmgrLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("failed to get appmgr list %v", err)
		api.HandleError(resp, req, err)
		return
	}
	ret := make([]v1alpha1.ApplicationManager, 0, len(ams))
	for _, am := range ams {
		if am.Spec.Type != v1alpha1.App && am.Spec.Type != v1alpha1.Middleware {
			continue
		}
		if userspace.IsSysApp(am.Spec.AppName) {
			continue
		}
		am.ManagedFields = make([]metav1.ManagedFieldsEntry, 0)
		am.Annotations[api.AppTokenKey] = ""
		am.Annotations[constants.ApplicationImageLabel] = ""
		am.Spec.Config = ""
		am.Status.OpRecords = make([]v1alpha1.OpRecord, 0)
		ret = append(ret, *am.DeepCopy())
	}
	resp.WriteAsJson(ret)
}

func (h *Handler) allOperateHistory(req *restful.Request, resp *restful.Response) {
	owner := req.Attribute(constants.UserContextAttribute).(string)
	source := req.QueryParameter("source")
	resourceType := req.QueryParameter("resourceType")

	filteredSources := constants.Sources
	filteredResourceTypes := constants.ResourceTypes
	if len(source) > 0 {
		filteredSources = sets.String{}
		for _, s := range strings.Split(source, "|") {
			filteredSources.Insert(s)
		}
	}
	if len(resourceType) > 0 {
		filteredResourceTypes = sets.String{}
		for _, s := range strings.Split(resourceType, "|") {
			filteredResourceTypes.Insert(s)
		}
	}

	ams, err := h.appmgrLister.List(labels.Everything())
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	ops := make([]appinstaller.OperateHistory, 0)

	for _, am := range ams {
		if !filteredResourceTypes.Has(am.Spec.Type.String()) {
			continue
		}
		if am.Spec.AppOwner != owner || userspace.IsSysApp(am.Spec.AppName) {
			continue
		}
		for _, r := range am.Status.OpRecords {
			if !filteredSources.Has(r.Source) {
				continue
			}
			op := appinstaller.OperateHistory{
				AppName:      am.Spec.AppName,
				AppNamespace: am.Spec.AppNamespace,
				AppOwner:     am.Spec.AppOwner,
				ResourceType: am.Spec.Type.String(),
				OpRecord: v1alpha1.OpRecord{
					OpType:    r.OpType,
					Message:   r.Message,
					Source:    r.Source,
					Version:   r.Version,
					Status:    r.Status,
					StateTime: r.StateTime,
				},
			}
			ops = append(ops, op)
		}
	}
	sort.Slice(ops, func(i, j int) bool {
		return ops[j].StateTime.Before(ops[i].StateTime)
	})

	resp.WriteAsJson(map[string]interface{}{"result": ops})
}

func (h *Handler) getApp(req *restful.Request, resp *restful.Response) {
	client := req.Attribute(constants.KubeSphereClientAttribute).(*clientset.ClientSet)
	owner := req.Attribute(constants.UserContextAttribute).(string)

	appName := req.PathParameter(ParamAppName)
	name, err := apputils.FmtAppMgrName(appName, owner, "")
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	var app *v1alpha1.Application

	app, err = client.AppClient.AppV1alpha1().Applications().Get(req.Request.Context(), name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		api.HandleError(resp, req, err)
		return
	}
	am, err := client.AppClient.AppV1alpha1().ApplicationManagers().Get(req.Request.Context(), name, metav1.GetOptions{})
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	app.Status.State = am.Status.State.String()

	resp.WriteAsJson(app)
}

func (h *Handler) apps(req *restful.Request, resp *restful.Response) {
	owner := req.Attribute(constants.UserContextAttribute).(string)
	isSysApp := req.QueryParameter("issysapp")
	state := req.QueryParameter("state")

	ss := make([]string, 0)
	if state != "" {
		ss = strings.Split(state, "|")
	}
	all := make([]string, 0)
	for _, a := range appstate.All {
		all = append(all, a.String())
	}
	stateSet := sets.NewString(all...)
	if len(ss) > 0 {
		stateSet = sets.String{}
	}
	for _, s := range ss {
		stateSet.Insert(s)
	}
	filteredApps := make([]v1alpha1.Application, 0)
	appsMap := make(map[string]*v1alpha1.Application)
	appsEntranceMap := make(map[string]*v1alpha1.Application)

	// get pending app's from app managers
	ams, err := h.appmgrLister.List(labels.Everything())
	if err != nil {
		klog.Infof("get app manager list failed %v", err)
		api.HandleError(resp, req, err)
		return
	}
	for _, am := range ams {
		if am.Spec.Type != v1alpha1.App {
			continue
		}
		if am.Spec.AppOwner != owner {
			continue
		}
		if len(isSysApp) > 0 && isSysApp == "true" {
			continue
		}
		if userspace.IsSysApp(am.Spec.AppName) {
			continue
		}
		if !stateSet.Has(am.Status.State.String()) {
			continue
		}

		var appconfig appcfg.ApplicationConfig
		err = json.Unmarshal([]byte(am.Spec.Config), &appconfig)
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}
		now := metav1.Now()
		name, _ := apputils.FmtAppMgrName(am.Spec.AppName, owner, appconfig.Namespace)
		app := &v1alpha1.Application{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:              name,
				CreationTimestamp: am.CreationTimestamp,
			},
			Spec: v1alpha1.ApplicationSpec{
				Name:            am.Spec.AppName,
				RawAppName:      am.Spec.RawAppName,
				Appid:           v1alpha1.AppName(am.Spec.AppName).GetAppID(),
				IsSysApp:        v1alpha1.AppName(am.Spec.AppName).IsSysApp(),
				Namespace:       am.Spec.AppNamespace,
				Owner:           owner,
				Entrances:       appconfig.Entrances,
				SharedEntrances: appconfig.SharedEntrances,
				Icon:            appconfig.Icon,
				Settings: map[string]string{
					"title": am.Annotations[constants.ApplicationTitleLabel],
				},
			},
			Status: v1alpha1.ApplicationStatus{
				State:      am.Status.State.String(),
				Progress:   am.Status.Progress,
				StatusTime: &now,
				UpdateTime: &now,
			},
		}
		appsMap[app.Name] = app
	}

	allApps, err := h.appLister.List(labels.Everything())
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	for _, a := range allApps {
		if a.Spec.Owner == owner {
			if len(isSysApp) > 0 && isSysApp == "true" && strconv.FormatBool(a.Spec.IsSysApp) != isSysApp {
				continue
			}
			appsEntranceMap[a.Name] = a

			if a.Spec.IsSysApp {
				appsMap[a.Name] = a
				continue
			}
			if v, ok := appsMap[a.Name]; ok {
				v.Spec.Settings = a.Spec.Settings
			}
		}
	}
	for _, app := range appsMap {
		if v, ok := appsEntranceMap[app.Name]; ok {
			app.Status.EntranceStatuses = v.Status.EntranceStatuses
		}
		filteredApps = append(filteredApps, *app)
	}

	// sort by create time desc
	sort.Slice(filteredApps, func(i, j int) bool {
		return filteredApps[i].CreationTimestamp.Before(&filteredApps[j].CreationTimestamp)
	})

	resp.WriteAsJson(filteredApps)
}

func (h *Handler) pendingOrInstallingApps(req *restful.Request, resp *restful.Response) {
	ams, err := apputils.GetPendingOrRunningTask(req.Request.Context())
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	resp.WriteAsJson(ams)
}

func (h *Handler) terminusVersion(req *restful.Request, resp *restful.Response) {
	terminus, err := utils.GetTerminus(req.Request.Context(), h.ctrlClient)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	resp.WriteAsJson(map[string]interface{}{"version": terminus.Spec.Version})
}

func (h *Handler) nodes(req *restful.Request, resp *restful.Response) {
	var nodes corev1.NodeList
	err := h.ctrlClient.List(req.Request.Context(), &nodes, &client.ListOptions{})
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	resp.WriteAsJson(map[string]interface{}{"result": nodes.Items})
}

//func toProcessing(state v1alpha1.ApplicationManagerState) v1alpha1.ApplicationManagerState {
//	if state == v1alpha1.Installing || state == v1alpha1.Uninstalling ||
//		state == v1alpha1.Upgrading || state == v1alpha1.Resuming ||
//		state == v1alpha1.Canceling || state == v1alpha1.Pending {
//		return v1alpha1.Processing
//	}
//	return state
//}

func (h *Handler) operateRecommend(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamWorkflowName)
	owner := req.Attribute(constants.UserContextAttribute).(string)

	var am v1alpha1.ApplicationManager
	name, err := apputils.FmtAppMgrName(app, owner, "")
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	err = h.ctrlClient.Get(req.Request.Context(), types.NamespacedName{Name: name}, &am)

	if err != nil {
		if apierrors.IsNotFound(err) {
			api.HandleNotFound(resp, req, err)
			return
		}
		api.HandleError(resp, req, err)
		return
	}
	operate := appinstaller.Operate{
		AppName:           am.Spec.AppName,
		AppOwner:          am.Spec.AppOwner,
		OpType:            am.Status.OpType,
		ResourceType:      am.Spec.Type.String(),
		State:             am.Status.State,
		Message:           am.Status.Message,
		CreationTimestamp: am.CreationTimestamp,
		Source:            am.Spec.Source,
	}
	resp.WriteAsJson(operate)
}

func (h *Handler) operateRecommendList(req *restful.Request, resp *restful.Response) {
	owner := req.Attribute(constants.UserContextAttribute).(string)

	ams, err := h.appmgrLister.List(labels.Everything())
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	filteredOperates := make([]appinstaller.Operate, 0)
	for _, am := range ams {
		if am.Spec.AppOwner == owner && am.Spec.Type == v1alpha1.Recommend {
			operate := appinstaller.Operate{
				AppName:           am.Spec.AppName,
				AppOwner:          am.Spec.AppOwner,
				State:             am.Status.State,
				OpType:            am.Status.OpType,
				ResourceType:      am.Spec.Type.String(),
				Message:           am.Status.Message,
				CreationTimestamp: am.CreationTimestamp,
				Source:            am.Spec.Source,
			}
			filteredOperates = append(filteredOperates, operate)
		}
	}
	// sort by create time desc
	sort.Slice(filteredOperates, func(i, j int) bool {
		return filteredOperates[j].CreationTimestamp.Before(&filteredOperates[i].CreationTimestamp)
	})

	resp.WriteAsJson(map[string]interface{}{"result": filteredOperates})
}

func (h *Handler) operateRecommendHistory(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamWorkflowName)
	owner := req.Attribute(constants.UserContextAttribute).(string)

	var am v1alpha1.ApplicationManager
	name, err := apputils.FmtAppMgrName(app, owner, "")
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	key := types.NamespacedName{Name: name}
	err = h.ctrlClient.Get(req.Request.Context(), key, &am)
	if err != nil {
		if apierrors.IsNotFound(err) {
			api.HandleNotFound(resp, req, err)
			return
		}
		api.HandleError(resp, req, err)
		return
	}

	ops := make([]appinstaller.OperateHistory, 0, len(am.Status.OpRecords))
	for _, r := range am.Status.OpRecords {
		op := appinstaller.OperateHistory{
			AppName:      am.Spec.AppName,
			AppNamespace: am.Spec.AppNamespace,
			AppOwner:     am.Spec.AppOwner,
			ResourceType: am.Spec.Type.String(),

			OpRecord: v1alpha1.OpRecord{
				OpType:    r.OpType,
				Message:   r.Message,
				Source:    r.Source,
				Version:   r.Version,
				Status:    r.Status,
				StateTime: r.StateTime,
			},
		}
		ops = append(ops, op)
	}
	resp.WriteAsJson(map[string]interface{}{"result": ops})
}

func (h *Handler) allOperateRecommendHistory(req *restful.Request, resp *restful.Response) {
	owner := req.Attribute(constants.UserContextAttribute).(string)

	ams, err := h.appmgrLister.List(labels.Everything())

	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	ops := make([]appinstaller.OperateHistory, 0)

	for _, am := range ams {
		if am.Spec.AppOwner != owner || userspace.IsSysApp(am.Spec.AppName) || am.Spec.Type != v1alpha1.Recommend {
			continue
		}
		for _, r := range am.Status.OpRecords {
			op := appinstaller.OperateHistory{
				AppName:      am.Spec.AppName,
				AppNamespace: am.Spec.AppNamespace,
				AppOwner:     am.Spec.AppOwner,
				ResourceType: am.Spec.Type.String(),

				OpRecord: v1alpha1.OpRecord{
					OpType:    r.OpType,
					Message:   r.Message,
					Source:    r.Source,
					Version:   r.Version,
					Status:    r.Status,
					StateTime: r.StateTime,
				},
			}
			ops = append(ops, op)
		}
	}
	sort.Slice(ops, func(i, j int) bool {
		return ops[j].StateTime.Before(ops[i].StateTime)
	})

	resp.WriteAsJson(map[string]interface{}{"result": ops})
}

func (h *Handler) allUsersApps(req *restful.Request, resp *restful.Response) {
	//owner := req.Attribute(constants.UserContextAttribute).(string)
	isSysApp := req.QueryParameter("issysapp")
	state := req.QueryParameter("state")

	ss := make([]string, 0)
	if state != "" {
		ss = strings.Split(state, "|")
	}
	all := make([]string, 0)
	for _, a := range appstate.All {
		all = append(all, a.String())
	}
	stateSet := sets.NewString(all...)
	if len(ss) > 0 {
		stateSet = sets.String{}
	}
	for _, s := range ss {
		stateSet.Insert(s)
	}

	filteredApps := make([]v1alpha1.Application, 0)
	appsMap := make(map[string]*v1alpha1.Application)
	appsEntranceMap := make(map[string]*v1alpha1.Application)
	// get pending app's from app managers
	ams, err := h.appmgrLister.List(labels.Everything())
	if err != nil {
		klog.Error(err)
		api.HandleError(resp, req, err)
		return
	}

	for _, am := range ams {
		if am.Spec.Type != v1alpha1.App {
			continue
		}

		if !stateSet.Has(am.Status.State.String()) {
			continue
		}
		if len(isSysApp) > 0 && isSysApp == "true" {
			continue
		}
		if userspace.IsSysApp(am.Spec.AppName) {
			continue
		}

		if am.Spec.Config == "" {
			continue
		}

		var appconfig appcfg.ApplicationConfig
		err = json.Unmarshal([]byte(am.Spec.Config), &appconfig)
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}

		now := metav1.Now()
		app := v1alpha1.Application{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:              am.Name,
				CreationTimestamp: am.CreationTimestamp,
			},
			Spec: v1alpha1.ApplicationSpec{
				Name:            am.Spec.AppName,
				RawAppName:      am.Spec.RawAppName,
				Appid:           v1alpha1.AppName(am.Spec.AppName).GetAppID(),
				IsSysApp:        v1alpha1.AppName(am.Spec.AppName).IsSysApp(),
				Namespace:       am.Spec.AppNamespace,
				Owner:           am.Spec.AppOwner,
				Entrances:       appconfig.Entrances,
				SharedEntrances: appconfig.SharedEntrances,
				Icon:            appconfig.Icon,
				Settings: map[string]string{
					"title": am.Annotations[constants.ApplicationTitleLabel],
				},
			},
			Status: v1alpha1.ApplicationStatus{
				State:      am.Status.State.String(),
				StatusTime: &now,
				UpdateTime: &now,
			},
		}
		appsMap[am.Name] = &app
	}

	allApps, err := h.appLister.List(labels.Everything())
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	// filter by application's owner
	for _, a := range allApps {
		if len(isSysApp) > 0 && strconv.FormatBool(a.Spec.IsSysApp) != isSysApp {
			continue
		}
		appsEntranceMap[a.Name] = a

		if a.Spec.IsSysApp {
			appsMap[a.Name] = a
			continue
		}
		if v, ok := appsMap[a.Name]; ok {
			v.Spec.Settings = a.Spec.Settings
		}
	}

	for _, app := range appsMap {
		entrances, err := app.GenEntranceURL(req.Request.Context())
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}
		app.Spec.Entrances = entrances

		sharedEntrances, err := app.GenSharedEntranceURL(req.Request.Context())
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}
		app.Spec.SharedEntrances = sharedEntrances

		if v, ok := appsEntranceMap[app.Name]; ok {
			app.Status.EntranceStatuses = v.Status.EntranceStatuses
		}
		filteredApps = append(filteredApps, *app)
	}

	// sort by create time desc
	sort.Slice(filteredApps, func(i, j int) bool {
		return filteredApps[j].CreationTimestamp.Before(&filteredApps[i].CreationTimestamp)
	})

	resp.WriteAsJson(filteredApps)
}

func (h *Handler) getAllUser() ([]string, error) {
	users := make([]string, 0)
	gvr := schema.GroupVersionResource{
		Group:    "iam.kubesphere.io",
		Version:  "v1alpha2",
		Resource: "users",
	}
	dClient, err := dynamic.NewForConfig(h.kubeConfig)
	if err != nil {
		return users, err
	}
	user, err := dClient.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return users, err
	}
	for _, u := range user.Items {
		if u.Object == nil {
			continue
		}
		users = append(users, u.GetName())
	}
	return users, nil
}

func (h *Handler) renderManifest(req *restful.Request, resp *restful.Response) {
	owner := req.Attribute(constants.UserContextAttribute).(string)

	request := api.ManifestRenderRequest{}
	err := req.ReadEntity(&request)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	admin, err := kubesphere.GetAdminUsername(req.Request.Context(), h.kubeConfig)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	isAdmin, err := kubesphere.IsAdmin(req.Request.Context(), h.kubeConfig, owner)
	if err != nil {
		klog.Error(err)
		api.HandleError(resp, req, err)
		return
	}
	renderedYAML, err := utils.RenderManifestFromContent([]byte(request.Content), owner, admin, isAdmin)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	resp.WriteEntity(api.ManifestRenderResponse{
		Response: api.Response{Code: 200},
		Data:     api.ManifestRenderRespData{Content: renderedYAML},
	})
}

func (h *Handler) adminUsername(req *restful.Request, resp *restful.Response) {
	config, err := ctrl.GetConfig()
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	username, err := kubesphere.GetAdminUsername(req.Request.Context(), config)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	resp.WriteEntity(api.AdminUsernameResponse{
		Response: api.Response{Code: 200},
		Data:     api.AdminUsernameRespData{Username: username},
	})
}

func (h *Handler) adminUserList(req *restful.Request, resp *restful.Response) {
	config, err := ctrl.GetConfig()
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	adminList, err := kubesphere.GetAdminUserList(req.Request.Context(), config)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	admins := make([]string, 0, len(adminList))
	for _, a := range adminList {
		admins = append(admins, a.Name)
	}
	resp.WriteEntity(api.AdminListResponse{
		Response: api.Response{Code: 200},
		Data:     admins,
	})
}

func (h *Handler) oamValues(req *restful.Request, resp *restful.Response) {
	values := map[string]interface{}{
		"admin": "admin",
		"bfl": map[string]string{
			"username": "admin",
		},
	}

	var nodes corev1.NodeList
	err := h.ctrlClient.List(req.Request.Context(), &nodes, &client.ListOptions{})
	if err != nil {
		klog.Errorf("list node failed %v", err)
		api.HandleError(resp, req, err)
		return
	}
	gpuType, err := utils.FindGpuTypeFromNodes(&nodes)
	if err != nil {
		klog.Errorf("get gpu type failed %v", gpuType)
		api.HandleError(resp, req, err)
		return
	}
	values["GPU"] = map[string]interface{}{
		"Type": gpuType,
		"Cuda": os.Getenv("OLARES_SYSTEM_CUDA_VERSION"),
	}
	values["user"] = map[string]interface{}{
		"zone": "user-zone",
	}
	values["schedule"] = map[string]interface{}{
		"nodeName": "node",
	}
	values["oidc"] = map[string]interface{}{
		"client": map[string]interface{}{},
		"issuer": "issuer",
	}
	values["userspace"] = map[string]interface{}{
		"appCache": "appcache",
		"userData": "userspace/Home",
	}
	values["os"] = map[string]interface{}{
		"appKey":    "appKey",
		"appSecret": "appSecret",
	}

	values["domain"] = map[string]string{}
	values["cluster"] = map[string]string{}
	values["dep"] = map[string]interface{}{}
	values["postgres"] = map[string]interface{}{
		"databases": map[string]interface{}{},
	}
	values["mariadb"] = map[string]interface{}{
		"databases": map[string]interface{}{},
	}
	values["mysql"] = map[string]interface{}{
		"databases": map[string]interface{}{},
	}
	values["minio"] = map[string]interface{}{
		"buckets": map[string]interface{}{},
	}
	values["rabbitmq"] = map[string]interface{}{
		"vhosts": map[string]interface{}{},
	}
	values["elasticsearch"] = map[string]interface{}{
		"indexes": map[string]interface{}{},
	}
	values["redis"] = map[string]interface{}{}
	values["mongodb"] = map[string]interface{}{
		"databases": map[string]interface{}{},
	}
	values["clickhouse"] = map[string]interface{}{
		"databases": map[string]interface{}{},
	}
	values["svcs"] = map[string]interface{}{}
	values["nats"] = map[string]interface{}{
		"subjects": map[string]interface{}{},
		"refs":     map[string]interface{}{},
	}
	resp.WriteAsJson(values)
}
