package apiserver

import (
	"context"

	sysv1alpha1 "bytetrade.io/web3os/system-server/pkg/apis/sys/v1alpha1"
	"bytetrade.io/web3os/system-server/pkg/apiserver/v1alpha1/api"
	"bytetrade.io/web3os/system-server/pkg/apiserver/v1alpha1/api/response"
	permission "bytetrade.io/web3os/system-server/pkg/permission/v1alpha1"
	prodiverregistry "bytetrade.io/web3os/system-server/pkg/providerregistry/v1alpha1"
	serviceproxy "bytetrade.io/web3os/system-server/pkg/serviceproxy/v1alpha1"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/client-go/rest"
)

// Handler include several fields that used for managing interactions with associated services.
type Handler struct {
	serviceCtx     context.Context
	kubeConfig     *rest.Config // helm's kubeconfig. TODO: insecure
	proxy          *serviceproxy.Proxy
	dispatcher     *serviceproxy.Dispatcher
	permissionCtrl *permission.PermissionControlSet
}

func newAPIHandler(ctx context.Context, kubeconfig *rest.Config,
	registry *prodiverregistry.Registry,
	ctrlSet *permission.PermissionControlSet,
) (*Handler, error) {
	proxy := serviceproxy.NewProxy(registry)
	dispatcher := serviceproxy.NewDispatcher(ctx, registry)

	return &Handler{
		serviceCtx:     ctx,
		kubeConfig:     kubeconfig,
		proxy:          proxy,
		dispatcher:     dispatcher,
		permissionCtrl: ctrlSet,
	}, nil
}

func (h *Handler) get(req *restful.Request, resp *restful.Response) {
	h.handleProxy(sysv1alpha1.Get, req, resp)
}

func (h *Handler) list(req *restful.Request, resp *restful.Response) {
	h.handleProxy(sysv1alpha1.List, req, resp)
}

func (h *Handler) create(req *restful.Request, resp *restful.Response) {
	h.handleProxy(sysv1alpha1.Create, req, resp)
}

func (h *Handler) update(req *restful.Request, resp *restful.Response) {
	h.handleProxy(sysv1alpha1.Update, req, resp)
}

func (h *Handler) delete(req *restful.Request, resp *restful.Response) {
	h.handleProxy(sysv1alpha1.Delete, req, resp)
}

func (h *Handler) action(req *restful.Request, resp *restful.Response) {
	op := req.PathParameter(api.ParamAction)
	params := req.Request.URL.RawQuery
	if params != "" {
		op += "?" + params
	}

	h.handleProxy(op, req, resp)
}

func (h *Handler) handleProxy(op string, req *restful.Request, resp *restful.Response) {
	token := req.Request.Header.Get(api.AccessTokenHeader)
	appKey, err := permission.ValidateAccessTokenWithRequest(token, op, req, h.permissionCtrl)
	if err != nil {
		response.HandleForbidden(resp, err)
		return
	}

	proxyrequest, err := serviceproxy.NewProxyRequestFromOpRequest(appKey, op, req)
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	// invoke provider
	ret, _, err := h.proxy.DoRequest(req, op, proxyrequest)
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	// notify watcher
	switch op {
	case sysv1alpha1.Create, sysv1alpha1.Update, sysv1alpha1.Delete:
		dispatchRequest := serviceproxy.NewDispatchRequest(proxyrequest, ret)
		h.dispatcher.DoWatch(dispatchRequest)
	}

	response.Success(resp, ret)
}
