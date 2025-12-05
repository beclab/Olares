package apiserver

import (
	"context"
	"encoding/json"

	"bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/app-service/pkg/apiserver/api"
	"bytetrade.io/web3os/app-service/pkg/appinstaller"
	"bytetrade.io/web3os/app-service/pkg/client/clientset"
	"bytetrade.io/web3os/app-service/pkg/constants"
	"bytetrade.io/web3os/app-service/pkg/middlewareinstaller"
	"bytetrade.io/web3os/app-service/pkg/utils"
	apputils "bytetrade.io/web3os/app-service/pkg/utils/app"
	installerv1 "bytetrade.io/web3os/app-service/pkg/workflowinstaller/v1"

	"sort"

	"github.com/emicklei/go-restful/v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

func (h *Handler) statusMiddleware(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamWorkflowName)
	owner := req.Attribute(constants.UserContextAttribute).(string)
	//client := req.Attribute(constants.KubeSphereClientAttribute).(*clientset.ClientSet)

	data, err := getMiddlewareStatus(req.Request.Context(), h.kubeConfig, app, owner)

	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(statusResp{
		Response: api.Response{Code: 200},
		Data:     data,
	})

}

func (h *Handler) statusMiddlewareList(req *restful.Request, resp *restful.Response) {
	//owner := req.Attribute(constants.UserContextAttribute).(string)
	//client := req.Attribute(constants.KubeSphereClientAttribute).(*clientset.ClientSet)

	var mgrs v1alpha1.ApplicationManagerList
	err := h.ctrlClient.List(req.Request.Context(), &mgrs)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	statusList := make([]*statusData, 0)
	for _, mgr := range mgrs.Items {
		if mgr.Spec.Type != v1alpha1.Middleware {
			continue
		}
		//if mgr.Spec.AppOwner != owner {
		//	continue
		//}

		data, err := getMiddlewareStatus(req.Request.Context(), h.kubeConfig, mgr.Spec.AppName, mgr.Spec.AppOwner)
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}

		statusList = append(statusList, data)
	}

	resp.WriteEntity(statusListResp{
		Response: api.Response{Code: 200},
		Data:     statusList,
	})
}

func getMiddlewareStatus(ctx context.Context, kubeConfig *rest.Config, app, owner string) (*statusData, error) {
	namespace, err := utils.AppNamespace(app, owner, "")
	if err != nil {
		return nil, err
	}

	helmClient, err := installerv1.NewHelmClient(ctx, kubeConfig, namespace)
	if err != nil {
		klog.Errorf("Failed to build helm client err=%v", err)
		return nil, err
	}

	installed, release, err := helmClient.Status(app)
	if err != nil {
		klog.Errorf("Failed to get install history app=%s err=%v", app, err)
		return nil, err
	}
	//status := v1alpha1.Uninstalled.String()
	//if installed {
	//	status = v1alpha1.Uninstalling.String()
	//}

	client, err := utils.GetClient()
	if err != nil {
		return nil, err
	}
	name, err := apputils.FmtAppMgrName(app, owner, namespace)
	if err != nil {
		return nil, err
	}

	mgr, err := client.AppV1alpha1().ApplicationManagers().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	//if mgr.Status.OpType == v1alpha1.UninstallOp && mgr.Status.State == v1alpha1.Uninstalling {
	//	res.ResourceStatus = v1alpha1.Uninstalling.String()
	//}
	//if mgr.Status.State == v1alpha1.Running {
	//	res.ResourceStatus = v1alpha1.AppRunning.String()
	//}
	res := statusData{
		UUID:           "",
		Namespace:      namespace,
		User:           owner,
		ResourceStatus: mgr.Status.State.String(),
		ResourceType:   v1alpha1.Middleware.String(),
		Metadata:       metadata{Name: app},
	}

	if release != nil {
		res.CreateTime = metav1.Time(release.Info.FirstDeployed)
		res.UpdataTime = metav1.Time(release.Info.LastDeployed)
	}
	var cfg middlewareinstaller.MiddlewareConfig
	err = json.Unmarshal([]byte(mgr.Spec.Config), &cfg)
	if err != nil {
		return nil, err
	}
	res.Title = cfg.Title

	if installed {
		version, err := helmClient.Version(app)
		if err != nil {
			klog.Errorf("Failed to get deployed chart version app=%s err=%v", app, err)
			return nil, err
		}

		res.Version = version
	}

	return &res, nil
}

func (h *Handler) operateMiddleware(req *restful.Request, resp *restful.Response) {
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

func (h *Handler) operateMiddlewareList(req *restful.Request, resp *restful.Response) {
	client := req.Attribute(constants.KubeSphereClientAttribute).(*clientset.ClientSet)
	owner := req.Attribute(constants.UserContextAttribute).(string)

	ams, err := client.AppClient.AppV1alpha1().ApplicationManagers().List(req.Request.Context(), metav1.ListOptions{})
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	filteredOperates := make([]appinstaller.Operate, 0)
	for _, am := range ams.Items {
		if am.Spec.AppOwner == owner && am.Spec.Type == v1alpha1.Middleware {
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
