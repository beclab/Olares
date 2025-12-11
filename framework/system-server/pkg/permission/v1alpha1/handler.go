package permission

import (
	"errors"
	"fmt"
	"time"

	"bytetrade.io/web3os/system-server/pkg/apiserver/v1alpha1/api"
	"bytetrade.io/web3os/system-server/pkg/apiserver/v1alpha1/api/response"
	"bytetrade.io/web3os/system-server/pkg/constants"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

type Handler struct {
	permissionCtrl *PermissionControl
	accessMgr      *AccessManager
	kubeconfig     *rest.Config
	kubeClientSet  *kubernetes.Clientset
}

func newHandler(ctrlSet *PermissionControlSet, kubeconfig *rest.Config) *Handler {

	client := kubernetes.NewForConfigOrDie(kubeconfig)

	return &Handler{
		permissionCtrl: ctrlSet.Ctrl,
		accessMgr:      ctrlSet.Mgr,
		kubeconfig:     kubeconfig,
		kubeClientSet:  client,
	}
}

func (h *Handler) auth(req *restful.Request, resp *restful.Response) {

	var accReq AccessTokenRequest
	err := req.ReadEntity(&accReq)
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	appPerm, err := h.permissionCtrl.getAppPermissionFromAppKey(req.Request.Context(), accReq.AppKey)
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	authToken, err := h.accessMgr.getAccessToken(&accReq, appPerm)
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	if !h.permissionCtrl.verifyPermission(appPerm, &accReq.Perm) {
		response.HandleForbidden(resp, errors.New("permission required is not allowed"))
		return
	}

	accReq.Perm.AppKey = accReq.AppKey
	h.accessMgr.cacheAccessToken(authToken, &accReq.Perm)

	token := AccessTokenResponse{
		AccessToken: authToken,
		ExpiredAt:   time.Now().Add(TokenCacheTTL),
	}

	response.Success(resp, token)
}

func (h *Handler) register(req *restful.Request, resp *restful.Response) {
	token := req.HeaderParameter(api.AuthorizationTokenHeader)
	user, err := validateToken(req.Request.Context(), h.kubeconfig, token)
	if err != nil {
		// internal api, return http code
		api.HandleUnauthorized(resp, req, err)
		return
	}

	if constants.MyNamespace != "user-system-"+user {
		api.HandleUnauthorized(resp, req, fmt.Errorf("invalid user, %s", user))
		return
	}

	var perm PermissionRegister

	if err = req.ReadEntity(&perm); err != nil {
		api.HandleError(resp, req, err)
		return
	}

	reg, err := h.permissionCtrl.applyPermission(req.Request.Context(), &perm)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	response.Success(resp, reg)
}

func (h *Handler) unregister(req *restful.Request, resp *restful.Response) {
	token := req.HeaderParameter(api.AuthorizationTokenHeader)
	user, err := validateToken(req.Request.Context(), h.kubeconfig, token)
	if err != nil {
		// internal api, return http code
		api.HandleUnauthorized(resp, req, err)
		return
	}

	if constants.MyNamespace != "user-system-"+user {
		api.HandleUnauthorized(resp, req, fmt.Errorf("invalid user, %s", user))
		return
	}

	var perm PermissionRegister

	if err = req.ReadEntity(&perm); err != nil {
		api.HandleError(resp, req, err)
		return
	}

	err = h.permissionCtrl.deletePermission(req.Request.Context(), perm.App)
	if err != nil {
		klog.Error("delete app ", perm.App, " permission error, ", err)
		api.HandleError(resp, req, err)
		return
	}

	klog.Info("app ", perm.App, " permission deleted")
	response.SuccessNoData(resp)
}
