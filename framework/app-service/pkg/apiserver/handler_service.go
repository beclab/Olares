package apiserver

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	appv1alpha1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appstate"
	"github.com/beclab/Olares/framework/app-service/pkg/client/clientset"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"

	"github.com/emicklei/go-restful/v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
)

// TODO: get owner's app with appname, instead to get by namespace and name
func (h *Handler) get(req *restful.Request, resp *restful.Response) {
	client := req.Attribute(constants.KubeSphereClientAttribute).(*clientset.ClientSet)
	owner := req.Attribute(constants.UserContextAttribute).(string)

	appName := req.PathParameter(ParamAppName)
	appNamespace := req.PathParameter(ParamAppNamespace)

	name := utils.FmtAppName(appName, appNamespace)

	// run with request context for incoming client
	app, err := client.AppClient.AppV1alpha1().Applications().Get(req.Request.Context(), name, metav1.GetOptions{})
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	if apierrors.IsNotFound(err) {
		am, err := client.AppClient.AppV1alpha1().ApplicationManagers().Get(req.Request.Context(), name, metav1.GetOptions{})
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}
		var appconfig appcfg.ApplicationConfig
		err = json.Unmarshal([]byte(am.Spec.Config), &appconfig)
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}
		app = &appv1alpha1.Application{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:              name,
				CreationTimestamp: metav1.Now(),
			},
			Spec: appv1alpha1.ApplicationSpec{
				Name:            am.Spec.AppName,
				RawAppName:      am.Spec.RawAppName,
				Appid:           appv1alpha1.AppName(am.Spec.AppName).GetAppID(),
				IsSysApp:        appv1alpha1.AppName(am.Spec.AppName).IsSysApp(),
				Namespace:       am.Spec.AppNamespace,
				Owner:           owner,
				Entrances:       appconfig.Entrances,
				SharedEntrances: appconfig.SharedEntrances,
				Icon:            appconfig.Icon,
			},
		}
	}

	resp.WriteAsJson(app)
}

func (h *Handler) list(req *restful.Request, resp *restful.Response) {
	client := req.Attribute(constants.KubeSphereClientAttribute).(*clientset.ClientSet)
	owner := req.Attribute(constants.UserContextAttribute).(string)

	// run with request context for incoming client
	allApps, err := client.AppClient.AppV1alpha1().Applications().List(req.Request.Context(), metav1.ListOptions{})
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	// filter by application's owner
	var filteredApps []appv1alpha1.Application
	for _, a := range allApps.Items {
		if a.Spec.Owner == owner {
			filteredApps = append(filteredApps, a)
		}
	}

	// get pending app's from app managers
	//var pendingApplications []appv1alpha1.Application
	appMgrs, err := client.AppClient.AppV1alpha1().ApplicationManagers().List(req.Request.Context(), metav1.ListOptions{})
	for _, am := range appMgrs.Items {
		var appconfig appcfg.ApplicationConfig
		err = json.Unmarshal([]byte(am.Spec.Config), &appconfig)
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}
		if am.Spec.AppOwner == owner && am.Status.State == appv1alpha1.Pending {
			app := appv1alpha1.Application{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              am.Name,
					CreationTimestamp: metav1.Now(),
				},
				Spec: appv1alpha1.ApplicationSpec{
					Name:            am.Spec.AppName,
					RawAppName:      am.Spec.RawAppName,
					Appid:           appv1alpha1.AppName(am.Spec.AppName).GetAppID(),
					IsSysApp:        appv1alpha1.AppName(am.Spec.AppName).IsSysApp(),
					Namespace:       am.Spec.AppNamespace,
					Owner:           owner,
					Entrances:       appconfig.Entrances,
					SharedEntrances: appconfig.SharedEntrances,
					Icon:            appconfig.Icon,
				},
			}
			filteredApps = append(filteredApps, app)
		}
	}

	// sort by create time desc
	sort.Slice(filteredApps, func(i, j int) bool {
		return filteredApps[i].CreationTimestamp.Before(&filteredApps[j].CreationTimestamp)
	})

	resp.WriteAsJson(filteredApps)
}

func (h *Handler) listBackend(req *restful.Request, resp *restful.Response) {
	owner := req.PathParameter(ParamUserName)
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
	filteredApps := make([]appv1alpha1.Application, 0)
	appsMap := make(map[string]*appv1alpha1.Application)
	appsEntranceMap := make(map[string]*appv1alpha1.Application)

	ams, err := h.appmgrLister.List(labels.Everything())
	if err != nil {
		klog.Infof("get app manager list failed %v", err)
		api.HandleError(resp, req, err)
		return
	}

	sharedEntranceApps := make(map[string]*appv1alpha1.Application)
	for _, am := range ams {
		if am.Spec.Type != appv1alpha1.App {
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
		for i := range appconfig.Entrances {
			if appconfig.Entrances[i].AuthLevel == "" {
				appconfig.Entrances[i].AuthLevel = "private"
			}
		}

		appconfig.SharedEntrances, err = appconfig.GenSharedEntranceURL(req.Request.Context())
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}

		if am.Spec.AppOwner != owner && len(appconfig.SharedEntrances) == 0 {
			continue
		}

		now := metav1.Now()
		name, _ := apputils.FmtAppMgrName(am.Spec.AppName, owner, appconfig.Namespace)
		app := &appv1alpha1.Application{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:              name,
				CreationTimestamp: am.CreationTimestamp,
			},
			Spec: appv1alpha1.ApplicationSpec{
				Name:            am.Spec.AppName,
				RawAppName:      am.Spec.RawAppName,
				Appid:           appv1alpha1.AppName(am.Spec.AppName).GetAppID(),
				IsSysApp:        appv1alpha1.AppName(am.Spec.AppName).IsSysApp(),
				Namespace:       am.Spec.AppNamespace,
				Owner:           am.Spec.AppOwner,
				Entrances:       appconfig.Entrances,
				Ports:           appconfig.Ports,
				SharedEntrances: appconfig.SharedEntrances,
				Icon:            appconfig.Icon,
				Settings: map[string]string{
					"title": am.Annotations[constants.ApplicationTitleLabel],
				},
			},
			Status: appv1alpha1.ApplicationStatus{
				State:      am.Status.State.String(),
				Progress:   am.Status.Progress,
				StatusTime: &now,
				UpdateTime: &now,
			},
		}

		if am.Spec.AppOwner != owner {
			// only show shared entrances
			// mask other infos
			app.Spec.Entrances = []appv1alpha1.Entrance{}

			if _, ok := sharedEntranceApps[app.Spec.Appid]; !ok {
				sharedEntranceApps[app.Spec.Appid] = app
			}
		} else {

			if len(appconfig.SharedEntrances) > 0 {
				// force to set the shared entrance app with my app
				sharedEntranceApps[app.Spec.Appid] = app
			}

			appsMap[app.Name] = app
		}
	}

	// add shared entrance apps
	for _, app := range sharedEntranceApps {
		if _, ok := appsMap[app.Name]; ok {
			continue
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
				v.Spec.Entrances = a.Spec.Entrances
<<<<<<< HEAD
				v.Spec.Ports = a.Spec.Ports
=======
>>>>>>> b27854b8 (fix: v2 app stop)
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
		if !filteredApps[i].CreationTimestamp.Equal(&filteredApps[j].CreationTimestamp) {
			return filteredApps[i].CreationTimestamp.Before(&filteredApps[j].CreationTimestamp)
		}
		return filteredApps[i].Spec.Name < filteredApps[j].Spec.Name
	})

	resp.WriteAsJson(filteredApps)
}
