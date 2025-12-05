package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/app-service/pkg/apiserver/api"
	"bytetrade.io/web3os/app-service/pkg/constants"
	"bytetrade.io/web3os/app-service/pkg/kubesphere"
	"bytetrade.io/web3os/app-service/pkg/middlewareinstaller"
	"bytetrade.io/web3os/app-service/pkg/tapr"
	"bytetrade.io/web3os/app-service/pkg/utils"
	apputils "bytetrade.io/web3os/app-service/pkg/utils/app"

	"github.com/emicklei/go-restful/v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
)

var middlewareManager map[string]context.CancelFunc

func (h *Handler) installMiddleware(req *restful.Request, resp *restful.Response) {
	insReq := &api.InstallRequest{}
	err := req.ReadEntity(insReq)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)
	token, err := h.GetUserServiceAccountToken(req.Request.Context(), owner)
	if err != nil {
		klog.Error("Failed to get user service account token: ", err)
		api.HandleError(resp, req, err)
		return
	}
	marketSource := req.HeaderParameter(constants.MarketSource)

	middlewareConfig, err := getMiddlewareConfigFromRepo(req.Request.Context(), &apputils.ConfigOptions{
		App:          app,
		Owner:        owner,
		RepoURL:      insReq.RepoURL,
		Version:      "",
		Token:        token,
		MarketSource: marketSource,
	})

	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	client, err := utils.GetClient()
	if err != nil {
		klog.Errorf("Failed to get client err=%v", err)
		api.HandleError(resp, req, err)
		return
	}

	role, err := kubesphere.GetUserRole(req.Request.Context(), owner)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	if role != "owner" && role != "admin" {
		api.HandleBadRequest(resp, req, errors.New("only admin user can install this middleware"))
		return
	}

	cfg, err := json.Marshal(middlewareConfig)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	var a *v1alpha1.ApplicationManager
	name := fmt.Sprintf("%s-%s", middlewareConfig.Namespace, app)
	mgr := &v1alpha1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.ApplicationManagerSpec{
			AppName:      app,
			RawAppName:   app,
			AppNamespace: middlewareConfig.Namespace,
			AppOwner:     owner,
			Source:       insReq.Source.String(),
			Type:         v1alpha1.Middleware,
			Config:       string(cfg),
		},
	}
	a, err = client.AppV1alpha1().ApplicationManagers().Get(req.Request.Context(), name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			api.HandleError(resp, req, err)
			return
		}
		a, err = client.AppV1alpha1().ApplicationManagers().Create(req.Request.Context(), mgr, metav1.CreateOptions{})
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}
	} else {
		patchData := map[string]interface{}{
			"spec": map[string]interface{}{
				"source": insReq.Source.String(),
			},
		}
		patchByte, err := json.Marshal(patchData)
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}
		_, err = client.AppV1alpha1().ApplicationManagers().Patch(req.Request.Context(),
			a.Name, types.MergePatchType, patchByte, metav1.PatchOptions{})
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}
	}
	now := metav1.Now()
	opRecord := v1alpha1.OpRecord{
		OpType:    v1alpha1.InstallOp,
		Message:   fmt.Sprintf(constants.InstallOperationCompletedTpl, a.Spec.Type.String(), a.Spec.AppName),
		Version:   middlewareConfig.Cfg.Metadata.Version,
		Source:    a.Spec.Source,
		Status:    v1alpha1.Running,
		StateTime: &now,
	}
	opID := strconv.FormatInt(time.Now().Unix(), 10)

	defer func() {
		if err != nil {
			opRecord.Status = v1alpha1.InstallFailed
			opRecord.Message = fmt.Sprintf(constants.OperationFailedTpl, a.Status.OpType, err.Error())
			e := apputils.UpdateStatus(a, opRecord.Status, &opRecord, opRecord.Message)
			if e != nil {
				klog.Errorf("Failed to update applicationmanager status name=%s err=%v", a.Name, e)
			} else {
				err = middlewareinstaller.Uninstall(context.TODO(), h.kubeConfig, middlewareConfig)
				if err != nil {
					klog.Errorf("Failed to uninstall middleware err=%v", err)
				}
				utils.PublishMiddlewareEvent(owner, app, string(v1alpha1.InstallOp), opID, v1alpha1.InstallFailed.String(), "", nil, "")
			}
		}
	}()

	middlewareStatus := v1alpha1.ApplicationManagerStatus{
		OpType:  v1alpha1.InstallOp,
		State:   v1alpha1.Installing,
		OpID:    opID,
		Message: "installing middleware",
		Payload: map[string]string{
			"version":      middlewareConfig.Cfg.Metadata.Version,
			"marketSource": marketSource,
		},
		StatusTime: &now,
		UpdateTime: &now,
	}
	a, err = apputils.UpdateAppMgrStatus(a.Name, middlewareStatus)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	utils.PublishMiddlewareEvent(owner, app, string(v1alpha1.InstallOp), opID, v1alpha1.Installing.String(), "", nil, "")

	klog.Infof("Start to install middleware name=%v", middlewareConfig.MiddlewareName)
	err = middlewareinstaller.Install(req.Request.Context(), h.kubeConfig, middlewareConfig)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	dConfig, err := dynamic.NewForConfig(h.kubeConfig)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	dc, err := middlewareinstaller.NewMiddlewareMongodb(dConfig)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	if middlewareManager == nil {
		middlewareManager = make(map[string]context.CancelFunc)
	}
	middlewareManager[name] = cancel

	timer := time.NewTicker(1 * time.Second)
	done := false
	go func() {
		for {
			select {
			case <-timer.C:
				if done {
					err = middlewareinstaller.Uninstall(context.TODO(), h.kubeConfig, middlewareConfig)
					if err != nil {
						klog.Errorf("Failed to uninstall middleware err=%v", err)
					}
					mgr, err = client.AppV1alpha1().ApplicationManagers().Get(context.TODO(), name, metav1.GetOptions{})
					if err != nil {
						klog.Errorf("Failed to get applicationmanagers err=%v", err)
						return
					}
					if mgr.Status.OpType == v1alpha1.CancelOp {
						if mgr.Status.Message == "timeout" {
							opRecord.Message = constants.OperationCanceledByTerminusTpl
						} else {
							opRecord.Message = constants.OperationCanceledByUserTpl
						}
					}
					opRecord.OpType = v1alpha1.CancelOp
					opRecord.Status = v1alpha1.InstallingCanceled
					err = apputils.UpdateStatus(mgr, v1alpha1.InstallingCanceled, &opRecord, opRecord.Message)
					if err != nil {
						klog.Infof("Failed to update status err=%v", err)
						return
					}
					utils.PublishMiddlewareEvent(owner, app, string(v1alpha1.CancelOp), opID, v1alpha1.InstallingCanceled.String(), "", nil, "")
					return
				}
				klog.Infof("ticker get middleware status")
				m, err := dc.Get(ctx, middlewareConfig.Namespace, "mongo-cluster", metav1.GetOptions{})
				if err != nil {
					klog.Errorf("Failed to get crd PerconaServerMongoDB name=%v namespace=%v err=%v",
						middlewareConfig.MiddlewareName, middlewareConfig.Namespace, err)
				}
				klog.Infof("m is nil: %v", m == nil)
				state, _, err := unstructured.NestedString(m.Object, "status", "state")
				if err != nil {
					klog.Error(err)
				}
				klog.Infof("middleware state=%v", state)
				if state == "ready" {
					e := apputils.UpdateStatus(a, opRecord.Status, &opRecord, opRecord.Message)
					if e != nil {
						klog.Error(e)
					} else {
						utils.PublishMiddlewareEvent(owner, app, string(v1alpha1.InstallOp), opID, opRecord.Status.String(), "", nil, "")
					}
					delete(middlewareManager, name)
					return
				}
			case <-ctx.Done():
				done = true
				klog.Infof("ctx....Done")
				//return
			}
		}
	}()
	resp.WriteEntity(api.InstallationResponse{
		Response: api.Response{Code: 200},
		Data:     api.InstallationResponseData{UID: middlewareConfig.MiddlewareName, OpID: opID},
	})
}

func (h *Handler) uninstallMiddleware(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)
	//client := req.Attribute(constants.KubeSphereClientAttribute).(*clientset.ClientSet)

	mrExists, err := h.isMiddlewareDependenciesExists(app)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	if mrExists {
		api.HandleBadRequest(resp, req,
			fmt.Errorf("can not delete middleware %s, there are still mr present", app))
		return
	}

	namespace, err := utils.AppNamespace(app, owner, "")
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	middlewareConfig := &middlewareinstaller.MiddlewareConfig{
		MiddlewareName: app,
		Namespace:      namespace,
		OwnerName:      owner,
	}

	now := metav1.Now()
	opID := strconv.FormatInt(time.Now().Unix(), 10)

	var mgr *v1alpha1.ApplicationManager
	middlewareStatus := v1alpha1.ApplicationManagerStatus{
		OpType:     v1alpha1.UninstallOp,
		State:      v1alpha1.Uninstalling,
		OpID:       opID,
		Message:    "try to uninstall a middleware",
		StatusTime: &now,
		UpdateTime: &now,
	}
	name, _ := apputils.FmtAppMgrName(app, owner, namespace)
	mgr, err = apputils.UpdateAppMgrStatus(name, middlewareStatus)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	utils.PublishMiddlewareEvent(owner, app, string(v1alpha1.UninstallOp), opID, v1alpha1.Uninstalling.String(), "", nil, "")

	now = metav1.Now()
	opRecord := v1alpha1.OpRecord{
		OpType:    v1alpha1.UninstallOp,
		Message:   "",
		Source:    mgr.Spec.Source,
		Version:   mgr.Status.Payload["version"],
		Status:    v1alpha1.UninstallFailed,
		StateTime: &now,
	}
	defer func() {
		if err != nil {
			opRecord.Message = fmt.Sprintf(constants.OperationFailedTpl, mgr.Status.OpType, err.Error())

			e := apputils.UpdateStatus(mgr, v1alpha1.UninstallFailed, &opRecord, opRecord.Message)
			if e != nil {
				klog.Errorf("Failed to update applicationmanager status in uninstall middleware name=%s err=%v", mgr.Name, e)
			} else {
				utils.PublishMiddlewareEvent(owner, app, string(v1alpha1.UninstallOp), opID, v1alpha1.UninstallFailed.String(), "", nil, "")
			}
		}
	}()
	err = middlewareinstaller.Uninstall(req.Request.Context(), h.kubeConfig, middlewareConfig)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	opRecord.Message = fmt.Sprintf(constants.UninstallOperationCompletedTpl, mgr.Spec.Type.String(), mgr.Spec.AppName)
	opRecord.Status = v1alpha1.Uninstalled
	err = apputils.UpdateStatus(mgr, v1alpha1.Uninstalled, &opRecord, opRecord.Message)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	utils.PublishMiddlewareEvent(owner, app, string(v1alpha1.Uninstalled), opID, v1alpha1.Uninstalled.String(), "", nil, "")

	resp.WriteEntity(api.InstallationResponse{
		Response: api.Response{Code: 200},
		Data:     api.InstallationResponseData{UID: app},
	})
}

func (h *Handler) cancelMiddleware(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamAppName)
	owner := req.Attribute(constants.UserContextAttribute).(string)
	// type = timeout | operate
	cancelType := req.QueryParameter("type")
	if cancelType == "" {
		cancelType = "operate"
	}
	name, err := apputils.FmtAppMgrName(app, owner, "")
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	cancel, ok := middlewareManager[name]
	if !ok {
		api.HandleError(resp, req, errors.New("can not execute cancel"))
		return
	}
	cancel()
	now := metav1.Now()
	opID := strconv.FormatInt(time.Now().Unix(), 10)

	status := v1alpha1.ApplicationManagerStatus{
		OpType:     v1alpha1.CancelOp,
		State:      v1alpha1.InstallingCanceling,
		OpID:       opID,
		Progress:   "0.00",
		Message:    cancelType,
		StatusTime: &now,
		UpdateTime: &now,
	}

	_, err = apputils.UpdateAppMgrStatus(name, status)

	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	utils.PublishMiddlewareEvent(owner, app, string(v1alpha1.CancelOp), opID, v1alpha1.InstallingCanceling.String(), "", nil, "")

	resp.WriteAsJson(api.InstallationResponse{
		Response: api.Response{Code: 200},
		Data:     api.InstallationResponseData{UID: app},
	})
}

func (h *Handler) isMiddlewareDependenciesExists(middleware string) (bool, error) {
	dConfig, err := dynamic.NewForConfig(h.kubeConfig)
	if err != nil {
		return false, err
	}
	dc, err := tapr.NewMiddlewareRequest(dConfig)
	if err != nil {
		return false, err
	}
	mrs, err := dc.List(context.TODO(), metav1.NamespaceAll, metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	if len(mrs.Items) > 0 {
		for _, mr := range mrs.Items {
			m, _, err := unstructured.NestedString(mr.Object, "spec", "middleware")
			if err != nil {
				return false, err
			}
			if m == middleware {
				return true, nil
			}
		}
	}
	return false, nil
}
