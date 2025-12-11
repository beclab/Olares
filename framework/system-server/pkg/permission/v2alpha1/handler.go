package v2alpha1

import (
	"context"
	"errors"

	"bytetrade.io/web3os/system-server/pkg/apiserver/v1alpha1/api"
	"bytetrade.io/web3os/system-server/pkg/apiserver/v1alpha1/api/response"
	"bytetrade.io/web3os/system-server/pkg/utils/apitools"
	"github.com/emicklei/go-restful/v3"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
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

	var app string
	if ok, app = h.execute(req, resp, username, h.bindingProvider); !ok {
		return
	}

	klog.Info("success to register provider, ", username, ", app=", app)
	reg := &RegisterResp{}
	response.Success(resp, reg)
}

func (h *handler) unregister(req *restful.Request, resp *restful.Response) {
	ok, username := h.Validate(req, resp)
	if !ok {
		return
	}

	_, app := h.execute(req, resp, username, h.unbindingProvider)

	klog.Info("success to unregister provider, ", username, ", app=", app)
	response.SuccessNoData(resp)
}

func (h *handler) execute(req *restful.Request, resp *restful.Response, username string,
	action func(ctx context.Context, user, app, serviceAccount string, roles []*rbacv1.ClusterRole) error) (success bool, appName string) {
	var err error
	var perm PermissionRegister

	if err = req.ReadEntity(&perm); err != nil {
		api.HandleError(resp, req, err)
		return
	}

	if perm.App == "" {
		err = errors.New("invalid app, app name is empty")
		klog.Error(err)
		api.HandleError(resp, req, err)
		return
	}

	for _, p := range perm.Perm {
		if p.ProviderName == "" {
			klog.Warning("provider name is empty")
			continue
		}

		if p.ProviderAppName == "" {
			klog.Warning("provider app name is empty")
			continue
		}

		if p.ProviderNamespace == "" {
			klog.Warning("provider namespace is empty")
			continue
		}

		if p.ServiceAccount == nil {
			p.ServiceAccount = ptr.To("default")
		}

		roles := h.getProvider(p.ProviderName, p.ProviderDomain, p.ProviderNamespace)

		if len(roles) == 0 {
			klog.Warning("no roles found for provider, ", appName)
			continue
		}

		if err = action(req.Request.Context(), username, perm.App, *p.ServiceAccount, roles); err != nil {
			klog.Error("fail to bind provider, ", err)
			api.HandleError(resp, req, err)
			return
		}

	} // end of for app perm loop

	success = true
	return
}
