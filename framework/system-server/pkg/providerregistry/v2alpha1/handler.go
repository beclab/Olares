package v2alpha1

import (
	"bytetrade.io/web3os/system-server/pkg/apiserver/v1alpha1/api"
	"bytetrade.io/web3os/system-server/pkg/apiserver/v1alpha1/api/response"
	"bytetrade.io/web3os/system-server/pkg/utils/apitools"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type handler struct {
	*apitools.BaseHandler
	kubeClient kubernetes.Interface
}

func (h *handler) register(req *restful.Request, resp *restful.Response) {
	ok, username := h.Validate(req, resp)
	if !ok {
		return
	}

	var err error
	var providerReq ProviderRegisterRequest
	if err = req.ReadEntity(&providerReq); err != nil {
		api.HandleError(resp, req, err)
		return
	}

	for _, provider := range providerReq.Providers {
		if err := h.createRefForProvider(req.Request.Context(),
			providerReq.AppName, providerReq.AppNamespace, &provider); err != nil {
			klog.Error("failed to create reference for provider: ", err)
			api.HandleError(resp, req, err)
			return
		}
	}

	klog.Info("success to register provider, ", username, ", app=", providerReq.AppName)
	response.SuccessNoData(resp)
}

func (h *handler) unregister(req *restful.Request, resp *restful.Response) {
	ok, username := h.Validate(req, resp)
	if !ok {
		return
	}

	var err error
	var providerReq ProviderRegisterRequest
	if err = req.ReadEntity(&providerReq); err != nil {
		api.HandleError(resp, req, err)
		return
	}

	for _, provider := range providerReq.Providers {
		if err = h.deleteRefForProvider(req.Request.Context(),
			providerReq.AppName, providerReq.AppNamespace, &provider); err != nil {
			klog.Error("failed to delete reference for provider: ", err)
			api.HandleError(resp, req, err)
			return
		}
	}

	klog.Info("success to unregister provider, ", username, ", app=", providerReq.AppName)
	response.SuccessNoData(resp)
}
