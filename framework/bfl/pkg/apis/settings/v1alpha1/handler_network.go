package v1alpha1

import (
	"bytetrade.io/web3os/bfl/pkg/api/response"
	"bytetrade.io/web3os/bfl/pkg/apiserver/runtime"
	"bytetrade.io/web3os/bfl/pkg/network"
	"github.com/emicklei/go-restful/v3"
)

func (h *Handler) networkStore() (*network.Store, error) {
	kc, err := runtime.NewKubeClientInCluster()
	if err != nil {
		return nil, err
	}
	return &network.Store{Client: kc.CtrlClient()}, nil
}

func (h *Handler) handleBindDefaultDomain(req *restful.Request, resp *restful.Response) {
	username, err := runtime.RequestUser(req)
	if err != nil {
		response.HandleBadRequest(resp, err)
		return
	}
	store, err := h.networkStore()
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	if err := store.RequireAdmin(req.Request.Context(), username); err != nil {
		response.HandleForbidden(resp, err)
		return
	}
	var body struct {
		Domain       string `json:"domain"`
		JWSSignature string `json:"jws_signature"`
	}
	if err := req.ReadEntity(&body); err != nil {
		response.HandleBadRequest(resp, err)
		return
	}
	if err := store.Bind(req.Request.Context(), username, body.Domain, body.JWSSignature); err != nil {
		response.HandleBadRequest(resp, err)
		return
	}
	h.handleGetDefaultDomain(req, resp)
}

func (h *Handler) handleGetDefaultDomain(req *restful.Request, resp *restful.Response) {
	store, err := h.networkStore()
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	binding, err := store.Binding(req.Request.Context())
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	_, _, status, err := store.Config(req.Request.Context())
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	if binding.Domain != "" && status.Domain.Phase == "unbound" {
		status.Domain.Phase = "pending"
	}
	response.Success(resp, network.DomainView{DomainBinding: binding, Status: status.Domain})
}

func (h *Handler) handleUpdateDefaultReverseProxy(req *restful.Request, resp *restful.Response) {
	username, err := runtime.RequestUser(req)
	if err != nil {
		response.HandleBadRequest(resp, err)
		return
	}
	store, err := h.networkStore()
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	if err := store.RequireAdmin(req.Request.Context(), username); err != nil {
		response.HandleForbidden(resp, err)
		return
	}
	var conf network.ProxyConfig
	if err := req.ReadEntity(&conf); err != nil {
		response.HandleBadRequest(resp, err)
		return
	}
	if err := store.SetProxy(req.Request.Context(), username, conf); err != nil {
		response.HandleBadRequest(resp, err)
		return
	}
	h.handleGetDefaultReverseProxy(req, resp)
}

func (h *Handler) handleGetDefaultReverseProxy(req *restful.Request, resp *restful.Response) {
	store, err := h.networkStore()
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	binding, err := store.Binding(req.Request.Context())
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	_, spec, status, err := store.Config(req.Request.Context())
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	token, err := store.AuthToken(req.Request.Context(), spec.Config.FRPAuthMethod)
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	if binding.Domain == "" && spec.Config.Mode != network.ModeNone {
		status.Proxy.Phase, status.Proxy.Message = "waiting_domain", "Bind the default domain to continue"
	}
	entrance, err := store.Entrance(req.Request.Context())
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	response.Success(resp, network.ProxyView{Config: spec.Config, Username: spec.Username, HasAuthToken: token != "", Status: status.Proxy, Entrance: entrance})
}
