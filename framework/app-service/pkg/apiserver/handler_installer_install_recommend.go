package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/client/clientset"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	"github.com/beclab/Olares/framework/app-service/pkg/workflowinstaller"

	"github.com/emicklei/go-restful/v3"
	"github.com/go-resty/resty/v2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type KnowledgeInstallMsg struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func (h *Handler) notifyKnowledgeInstall(title, name, owner string) error {
	knowledgeAPI := "http://rss-svc.os-framework:3010/knowledge/algorithm/recommend/install"
	klog.Info("Start to notify knowledge to Install ", knowledgeAPI, title, name)

	msg := KnowledgeInstallMsg{
		ID:    name,
		Title: title,
	}
	body, jsonErr := json.Marshal(msg)
	if jsonErr != nil {
		return jsonErr
	}
	client := resty.New()
	resp, err := client.SetTimeout(10*time.Second).R().
		SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetHeader("X-Bfl-User", owner).
		SetBody(body).Post(knowledgeAPI)
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		klog.Errorf("Failed to notify knowledge to Install status=%s", resp.Status())
		return errors.New(resp.Status())
	}
	return nil
}

func (h *Handler) notifyKnowledgeUnInstall(name, owner string) error {
	knowledgeAPI := "http://rss-svc.os-framework:3010/knowledge/algorithm/recommend/uninstall"

	msg := KnowledgeInstallMsg{
		ID: name,
	}
	body, jsonErr := json.Marshal(msg)
	if jsonErr != nil {
		return jsonErr
	}
	klog.Info("Start to notify knowledge to unInstall ", knowledgeAPI)
	client := resty.New()
	resp, err := client.SetTimeout(10*time.Second).R().
		SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetHeader("X-Bfl-User", owner).
		SetBody(body).Post(knowledgeAPI)

	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		klog.Errorf("Failed to notify knowledge to Install status=%s", resp.Status())
		return errors.New(resp.Status())
	}
	return nil
}
func (h *Handler) cleanRecommendEntryData(name, owner string) error {
	knowledgeAPI := fmt.Sprintf("http://knowledge-base-api.user-system-%s:3010", owner)
	entryAPI := knowledgeAPI + "/knowledge/entry/algorithm/" + name
	klog.Info("Start to clean recommend entry data ", entryAPI)
	client := resty.New().SetTimeout(10*time.Second).
		SetHeader("X-Bfl-User", owner)
	entryResp, err := client.R().Get(entryAPI)
	if err != nil {
		return err
	}
	if entryResp.StatusCode() != http.StatusOK {
		klog.Errorf("Failed to get knowledge entry list status=%s", entryResp.Status())
		return errors.New(entryResp.Status())
	}
	var ret workflowinstaller.KnowledgeAPIResp
	err = json.Unmarshal(entryResp.Body(), &ret)
	if err != nil {
		return err
	}
	urlsCount := len(ret.Data)
	if urlsCount > 0 {
		limit := 100
		removeClient := resty.New()
		entryRemoveAPI := knowledgeAPI + "/knowledge/entry/" + name
		for i := 0; i*limit < urlsCount; i++ {
			start := i * limit
			end := start + limit
			if end > urlsCount {
				end = urlsCount
			}
			removeList := ret.Data[start:end]
			removeBody, _ := json.Marshal(removeList)
			res, _ := removeClient.SetTimeout(5*time.Second).R().SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
				SetBody(removeBody).Delete(entryRemoveAPI)

			if res.StatusCode() == http.StatusOK {
				klog.Info("Delete entry success page: ", i, len(removeList))
			} else {
				klog.Info("Clean recommend entry data error:", string(removeBody), string(res.Body()))
			}
		}

	}
	klog.Info("Delete entry success page: ", name, urlsCount)
	return nil
}

func (h *Handler) installRecommend(req *restful.Request, resp *restful.Response) {
	insReq := &api.InstallRequest{}
	err := req.ReadEntity(insReq)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	app := req.PathParameter(ParamWorkflowName)
	owner := req.Attribute(constants.UserContextAttribute).(string)
	marketSource := req.HeaderParameter(constants.MarketSource)
	token, err := h.GetUserServiceAccountToken(req.Request.Context(), owner)
	if err != nil {
		klog.Error("Failed to get user service account token: ", err)
		api.HandleError(resp, req, err)
		return
	}

	klog.Infof("Download chart and get workflow config appName=%s repoURL=%s", app, insReq.RepoURL)
	//workflowCfg, err := getWorkflowConfigFromRepo(req.Request.Context(), owner, app, insReq.RepoURL, "", token, marketSource)

	workflowCfg, err := getWorkflowConfigFromRepo(req.Request.Context(), &apputils.ConfigOptions{
		App:          app,
		Owner:        owner,
		RepoURL:      insReq.RepoURL,
		Version:      "",
		Token:        token,
		MarketSource: marketSource,
	})
	if err != nil {
		klog.Errorf("Failed to get workflow config appName=%s repoURL=%s err=%v", app, insReq.RepoURL, err)
		api.HandleError(resp, req, err)
		return
	}

	satisfied, err := apputils.CheckMiddlewareRequirement(req.Request.Context(), h.ctrlClient, workflowCfg.Cfg.Middleware)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	if !satisfied {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, api.RequirementResp{
			Response: api.Response{Code: 400},
			Resource: "middleware",
			Message:  "middleware requirement can not be satisfied",
		})
		return
	}

	go h.notifyKnowledgeInstall(workflowCfg.Cfg.Metadata.Title, app, owner)

	client, _ := utils.GetClient()

	var a *v1alpha1.ApplicationManager
	//appNamespace, _ := utils.AppNamespace(app, owner, workflowCfg.Namespace)
	name, _ := apputils.FmtAppMgrName(app, owner, workflowCfg.Namespace)
	recommendMgr := &v1alpha1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", workflowCfg.Namespace, app),
		},
		Spec: v1alpha1.ApplicationManagerSpec{
			AppName:      app,
			AppNamespace: workflowCfg.Namespace,
			AppOwner:     owner,
			Source:       insReq.Source.String(),
			Type:         v1alpha1.Recommend,
		},
	}
	a, err = client.AppV1alpha1().ApplicationManagers().Get(req.Request.Context(), name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			api.HandleError(resp, req, err)
			return
		}
		a, err = client.AppV1alpha1().ApplicationManagers().Create(req.Request.Context(), recommendMgr, metav1.CreateOptions{})
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
	recommendStatus := v1alpha1.ApplicationManagerStatus{
		OpType:  v1alpha1.InstallOp,
		State:   v1alpha1.Installing,
		Message: "installing recommend",
		Payload: map[string]string{
			"version": workflowCfg.Cfg.Metadata.Version,
		},
		StatusTime: &now,
		UpdateTime: &now,
	}
	a, err = apputils.UpdateAppMgrStatus(a.Name, recommendStatus)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	opRecord := v1alpha1.OpRecord{
		OpType:    v1alpha1.InstallOp,
		Version:   workflowCfg.Cfg.Metadata.Version,
		Source:    a.Spec.Source,
		Status:    v1alpha1.Running,
		StateTime: &now,
	}

	klog.Info("Start to install workflow, ", workflowCfg)
	err = workflowinstaller.Install(req.Request.Context(), h.kubeConfig, workflowCfg)
	if err != nil {
		opRecord.Status = v1alpha1.Failed
		opRecord.Message = fmt.Sprintf(constants.OperationFailedTpl, a.Status.OpType, err.Error())
		e := apputils.UpdateStatus(a, opRecord.Status, &opRecord, opRecord.Message)
		if e != nil {
			klog.Errorf("Failed to update applicationmanager status name=%s err=%v", a.Name, e)
		}
		api.HandleError(resp, req, err)
		return
	}

	now = metav1.Now()
	opRecord.Message = fmt.Sprintf(constants.InstallOperationCompletedTpl, a.Spec.Type.String(), a.Spec.AppName)
	err = apputils.UpdateStatus(a, v1alpha1.Running, &opRecord, opRecord.Message)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(api.InstallationResponse{
		Response: api.Response{Code: 200},
		Data:     api.InstallationResponseData{UID: app},
	})
}

func (h *Handler) uninstallRecommend(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamWorkflowName)
	owner := req.Attribute(constants.UserContextAttribute).(string)
	client := req.Attribute(constants.KubeSphereClientAttribute).(*clientset.ClientSet)
	var err error

	//namespace := fmt.Sprintf("%s-%s", app, owner)
	namespace, err := utils.AppNamespace(app, owner, "")
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	workflowCfg := &workflowinstaller.WorkflowConfig{
		WorkflowName: app,
		Namespace:    namespace,
		OwnerName:    owner,
	}
	klog.Infof("Start to uninstall workflow name=%s", workflowCfg.WorkflowName)

	go h.cleanRecommendEntryData(app, owner)
	go h.notifyKnowledgeUnInstall(app, owner)

	now := metav1.Now()
	var recommendMgr *v1alpha1.ApplicationManager
	recommendStatus := v1alpha1.ApplicationManagerStatus{
		OpType:     v1alpha1.UninstallOp,
		State:      v1alpha1.Uninstalling,
		Message:    "try to uninstall a recommend",
		StatusTime: &now,
		UpdateTime: &now,
	}
	name, _ := apputils.FmtAppMgrName(app, owner, namespace)
	recommendMgr, err = apputils.UpdateAppMgrStatus(name, recommendStatus)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	defer func() {
		if err != nil {
			now := metav1.Now()
			message := fmt.Sprintf(constants.OperationFailedTpl, recommendMgr.Status.OpType, err.Error())
			opRecord := v1alpha1.OpRecord{
				OpType:    v1alpha1.UninstallOp,
				Message:   message,
				Source:    recommendMgr.Spec.Source,
				Version:   recommendMgr.Status.Payload["version"],
				Status:    v1alpha1.Failed,
				StateTime: &now,
			}
			e := apputils.UpdateStatus(recommendMgr, "failed", &opRecord, message)
			if e != nil {
				klog.Errorf("Failed to update applicationmanager status in uninstall Recommend name=%s err=%v", recommendMgr.Name, e)
			}
		}
	}()

	err = workflowinstaller.Uninstall(req.Request.Context(), h.kubeConfig, workflowCfg)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	klog.Infof("Start to delete namespace=%s", namespace)
	err = client.KubeClient.Kubernetes().CoreV1().Namespaces().Delete(req.Request.Context(), namespace, metav1.DeleteOptions{})
	if err != nil {
		klog.Errorf("Failed to delete workflow namespace=%s err=%v", namespace, err)
		api.HandleError(resp, req, err)
		return
	}
	go func() {
		timer := time.NewTicker(2 * time.Second)
		timeout := time.NewTimer(30 * time.Minute)
		defer timer.Stop()
		defer timeout.Stop()
		for {
			select {
			case <-timer.C:
				_, err := client.KubeClient.Kubernetes().CoreV1().Namespaces().
					Get(context.TODO(), namespace, metav1.GetOptions{})
				if err != nil {
					if apierrors.IsNotFound(err) {
						now := metav1.Now()
						message := fmt.Sprintf(constants.UninstallOperationCompletedTpl, recommendMgr.Spec.Type.String(), recommendMgr.Spec.AppName)
						opRecord := v1alpha1.OpRecord{
							OpType:    v1alpha1.UninstallOp,
							Message:   message,
							Source:    recommendMgr.Spec.Source,
							Version:   recommendMgr.Status.Payload["version"],
							Status:    v1alpha1.Running,
							StateTime: &now,
						}
						err = apputils.UpdateStatus(recommendMgr, opRecord.Status, &opRecord, message)
						if err != nil {
							klog.Errorf("Failed to update applicationmanager name=%s in uninstall Recommend err=%v", recommendMgr.Name, err)
						}
						return
					}

				}
			case <-timeout.C:
				klog.Errorf("Timeout to delete namespace=%s, please check it manually", namespace)
				return
			}
		}
	}()

	resp.WriteEntity(api.InstallationResponse{
		Response: api.Response{Code: 200},
		Data:     api.InstallationResponseData{UID: app},
	})
}

func (h *Handler) upgradeRecommend(req *restful.Request, resp *restful.Response) {
	app := req.PathParameter(ParamWorkflowName)
	owner := req.Attribute(constants.UserContextAttribute).(string)
	var err error
	token, err := h.GetUserServiceAccountToken(req.Request.Context(), owner)
	if err != nil {
		klog.Error("Failed to get user service account token: ", err)
		api.HandleError(resp, req, err)
		return
	}

	marketSource := req.HeaderParameter(constants.MarketSource)
	upReq := &api.UpgradeRequest{}
	err = req.ReadEntity(upReq)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	var recommendMgr *v1alpha1.ApplicationManager
	var workflowCfg *workflowinstaller.WorkflowConfig

	defer func() {
		now := metav1.Now()
		opRecord := v1alpha1.OpRecord{
			OpType:    v1alpha1.UpgradeOp,
			Message:   fmt.Sprintf(constants.UpgradeOperationCompletedTpl, recommendMgr.Spec.Type.String(), recommendMgr.Spec.AppName),
			Source:    recommendMgr.Spec.Source,
			Version:   workflowCfg.Cfg.Metadata.Version,
			Status:    v1alpha1.Running,
			StateTime: &now,
		}
		if err != nil {
			opRecord.Status = v1alpha1.Failed
			opRecord.Message = fmt.Sprintf(constants.OperationFailedTpl, recommendMgr.Status.OpType, err.Error())
		}
		e := apputils.UpdateStatus(recommendMgr, opRecord.Status, &opRecord, opRecord.Message)
		if e != nil {
			klog.Errorf("Failed to update applicationmanager status in upgrade recommend name=%s err=%v", recommendMgr.Name, e)
		}

	}()

	now := metav1.Now()
	recommendStatus := v1alpha1.ApplicationManagerStatus{
		OpType:     v1alpha1.UpgradeOp,
		State:      v1alpha1.Upgrading,
		Message:    "try to upgrade a recommend",
		StatusTime: &now,
		UpdateTime: &now,
	}

	klog.Infof("Download latest version chart and get workflow config name=%s repoURL=%s", app, upReq.RepoURL)

	workflowCfg, err = getWorkflowConfigFromRepo(req.Request.Context(), &apputils.ConfigOptions{
		App:          app,
		Owner:        owner,
		RepoURL:      upReq.RepoURL,
		Version:      "",
		Token:        token,
		MarketSource: marketSource,
	})
	if err != nil {
		klog.Errorf("Failed to get workflow config name=%s repoURL=%s err=%v, ", app, upReq.RepoURL, err)
		api.HandleError(resp, req, err)
		return
	}
	name, _ := apputils.FmtAppMgrName(app, owner, workflowCfg.Namespace)
	recommendMgr, err = apputils.UpdateAppMgrStatus(name, recommendStatus)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	klog.Infof("Start to upgrade workflow name=%s", workflowCfg.WorkflowName)
	err = workflowinstaller.Upgrade(req.Request.Context(), h.kubeConfig, workflowCfg)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(api.InstallationResponse{
		Response: api.Response{Code: 200},
		Data:     api.InstallationResponseData{UID: app},
	})

}
