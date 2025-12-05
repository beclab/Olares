package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/app-service/pkg/apiserver/api"
	"bytetrade.io/web3os/app-service/pkg/constants"
	apputils "bytetrade.io/web3os/app-service/pkg/utils/app"
	"bytetrade.io/web3os/app-service/pkg/workflowinstaller"

	"github.com/emicklei/go-restful/v3"
	"github.com/go-resty/resty/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
