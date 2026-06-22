package apiserver

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	sysv1alpha1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"

	"github.com/emicklei/go-restful/v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type SystemEnvUpdateRequest struct {
	Value string `json:"value"`
}

// SystemEnvDetail extends SystemEnv with reference information
type SystemEnvDetail struct {
	sysv1alpha1.EnvVarSpec `json:",inline"`
	ReferencedBy           []AppEnvReferrer `json:"referencedBy"`
}

type AppEnvReferrer struct {
	AppName   string `json:"appName"`
	AppOwner  string `json:"appOwner"`
	Namespace string `json:"namespace,omitempty"`
}

// collectAppEnvReferrers builds a map from a referenced env name to the apps
// that reference it via AppEnv.ValueFrom. When ownerFilter is non-empty, only
// AppEnvs owned by that user are considered (used for user envs); an empty
// ownerFilter considers every app (used for system envs).
func (h *Handler) collectAppEnvReferrers(ctx context.Context, ownerFilter string) (map[string][]AppEnvReferrer, error) {
	var appEnvList sysv1alpha1.AppEnvList
	if err := h.ctrlClient.List(ctx, &appEnvList); err != nil {
		return nil, err
	}
	refMap := make(map[string][]AppEnvReferrer)
	for _, ae := range appEnvList.Items {
		if ownerFilter != "" && ae.AppOwner != ownerFilter {
			continue
		}
		seen := make(map[string]struct{})
		for _, envVar := range ae.Envs {
			if envVar.ValueFrom == nil || envVar.ValueFrom.EnvName == "" {
				continue
			}
			envName := envVar.ValueFrom.EnvName
			if _, ok := seen[envName]; ok {
				continue
			}
			seen[envName] = struct{}{}
			refMap[envName] = append(refMap[envName], AppEnvReferrer{
				AppName:   ae.AppName,
				AppOwner:  ae.AppOwner,
				Namespace: ae.Namespace,
			})
		}
	}
	return refMap, nil
}

func (h *Handler) ensureAdmin(req *restful.Request, resp *restful.Response) (string, bool) {
	owner := getCurrentUser(req)
	isAdmin, err := kubesphere.IsAdmin(req.Request.Context(), h.kubeConfig, owner)
	if err != nil {
		api.HandleError(resp, req, err)
		return "", false
	}
	if !isAdmin {
		api.HandleBadRequest(resp, req, fmt.Errorf("only admin user can operate systemenvs"))
		return "", false
	}
	return owner, true
}

// todo: is it allowed?
func (h *Handler) createSystemEnv(req *restful.Request, resp *restful.Response) {
	_, ok := h.ensureAdmin(req, resp)
	if !ok {
		return
	}
	var body sysv1alpha1.EnvVarSpec
	if err := req.ReadEntity(&body); err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	if body.EnvName == "" {
		api.HandleBadRequest(resp, req, fmt.Errorf("name is required"))
		return
	}

	// validate and normalize resource name
	resourceName, err := utils.EnvNameToResourceName(body.EnvName)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	obj := &sysv1alpha1.SystemEnv{
		ObjectMeta: metav1.ObjectMeta{
			Name: resourceName,
		},
		EnvVarSpec: body,
	}

	if err := obj.ValidateValue(body.Value); err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	if err := obj.ValidateValue(body.Default); err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	if err := h.ctrlClient.Create(req.Request.Context(), obj); err != nil {
		api.HandleError(resp, req, err)
		return
	}
	resp.WriteAsJson(obj.EnvVarSpec)
}

func (h *Handler) updateSystemEnv(req *restful.Request, resp *restful.Response) {
	_, ok := h.ensureAdmin(req, resp)
	if !ok {
		return
	}

	name := req.PathParameter(ParamEnvName)
	if name == "" {
		api.HandleBadRequest(resp, req, fmt.Errorf("systemenv name is required"))
		return
	}

	var body SystemEnvUpdateRequest
	if err := req.ReadEntity(&body); err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	ctx := req.Request.Context()
	// validate and normalize resource name from path
	resourceName, err := utils.EnvNameToResourceName(name)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	var current sysv1alpha1.SystemEnv
	if err := h.ctrlClient.Get(ctx, types.NamespacedName{Name: resourceName}, &current); err != nil {
		api.HandleError(resp, req, err)
		return
	}

	if !current.Editable {
		api.HandleBadRequest(resp, req, fmt.Errorf("systemenv '%s' is not editable", current.EnvName))
		return
	}

	if current.Required && current.Default == "" && body.Value == "" {
		api.HandleBadRequest(resp, req, fmt.Errorf("systemenv '%s' is required", current.EnvName))
		return
	}

	if current.Value != body.Value {
		err := current.ValidateValue(body.Value)
		if err != nil {
			api.HandleBadRequest(resp, req, err)
			return
		}
		klog.Infof("Updating SystemEnv %s value from '%s' to '%s'", resourceName, current.Value, body.Value)
		original := current.DeepCopy()
		current.Value = body.Value
		if err := h.ctrlClient.Patch(ctx, &current, client.MergeFrom(original)); err != nil {
			api.HandleError(resp, req, err)
			return
		}
	}

	resp.WriteAsJson(current.EnvVarSpec)
}

func (h *Handler) deleteSystemEnv(req *restful.Request, resp *restful.Response) {
	_, ok := h.ensureAdmin(req, resp)
	if !ok {
		return
	}

	name := req.PathParameter(ParamEnvName)
	if name == "" {
		api.HandleBadRequest(resp, req, fmt.Errorf("systemenv name is required"))
		return
	}

	ctx := req.Request.Context()
	resourceName, err := utils.EnvNameToResourceName(name)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	var current sysv1alpha1.SystemEnv
	if err := h.ctrlClient.Get(ctx, types.NamespacedName{Name: resourceName}, &current); err != nil {
		if apierrors.IsNotFound(err) {
			resp.WriteEntity(api.Response{Code: 200})
			return
		}
		api.HandleError(resp, req, err)
		return
	}
	if current.Required {
		api.HandleBadRequest(resp, req, fmt.Errorf("systemenv '%s' is required", current.EnvName))
		return
	}

	refMap, err := h.collectAppEnvReferrers(ctx, "")
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	if refs := refMap[current.EnvName]; len(refs) > 0 {
		api.HandleBadRequest(resp, req, fmt.Errorf("systemenv '%s' is referenced by %d app(s) and cannot be deleted", current.EnvName, len(refs)))
		return
	}

	if err := h.ctrlClient.Delete(ctx, &current); err != nil {
		if !apierrors.IsNotFound(err) {
			api.HandleError(resp, req, err)
			return
		}
	}

	resp.WriteEntity(api.Response{Code: 200})
}

// listSystemEnvs returns all system env specs. The referencedBy field is only
// populated for admin users.
func (h *Handler) listSystemEnvs(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	var list sysv1alpha1.SystemEnvList
	if err := h.ctrlClient.List(ctx, &list); err != nil {
		api.HandleError(resp, req, err)
		return
	}

	owner := getCurrentUser(req)
	isAdmin, err := kubesphere.IsAdmin(ctx, h.kubeConfig, owner)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	var refMap map[string][]AppEnvReferrer
	if isAdmin {
		refMap, err = h.collectAppEnvReferrers(ctx, "")
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}
	}

	result := make([]SystemEnvDetail, 0, len(list.Items))
	for _, item := range list.Items {
		detail := SystemEnvDetail{EnvVarSpec: item.EnvVarSpec}
		if isAdmin {
			detail.ReferencedBy = refMap[item.EnvName]
		}
		result = append(result, detail)
	}
	resp.WriteAsJson(result)
}

// getSystemEnvDetail returns a system env spec along with referencing app envs
func (h *Handler) getSystemEnvDetail(req *restful.Request, resp *restful.Response) {
	_, ok := h.ensureAdmin(req, resp)
	if !ok {
		return
	}

	name := req.PathParameter(ParamEnvName)
	if name == "" {
		api.HandleBadRequest(resp, req, fmt.Errorf("systemenv name is required"))
		return
	}

	ctx := req.Request.Context()
	resourceName, err := utils.EnvNameToResourceName(name)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	var current sysv1alpha1.SystemEnv
	if err := h.ctrlClient.Get(ctx, types.NamespacedName{Name: resourceName}, &current); err != nil {
		api.HandleError(resp, req, err)
		return
	}

	detail := SystemEnvDetail{EnvVarSpec: current.EnvVarSpec}

	refMap, err := h.collectAppEnvReferrers(ctx, "")
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	detail.ReferencedBy = refMap[current.EnvName]

	resp.WriteAsJson(detail)
}

func (h *Handler) batchUpdateSystemEnvs(req *restful.Request, resp *restful.Response) {
	_, ok := h.ensureAdmin(req, resp)
	if !ok {
		return
	}

	var items []sysv1alpha1.EnvVarSpec
	if err := req.ReadEntity(&items); err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	if len(items) == 0 {
		resp.WriteAsJson([]sysv1alpha1.EnvVarSpec{})
		return
	}

	ctx := req.Request.Context()
	results := make([]sysv1alpha1.EnvVarSpec, 0, len(items))

	for _, it := range items {
		if it.EnvName == "" {
			api.HandleBadRequest(resp, req, fmt.Errorf("systemenv name is required"))
			return
		}

		resourceName, err := utils.EnvNameToResourceName(it.EnvName)
		if err != nil {
			api.HandleBadRequest(resp, req, err)
			return
		}

		var current sysv1alpha1.SystemEnv
		if err := h.ctrlClient.Get(ctx, types.NamespacedName{Name: resourceName}, &current); err != nil {
			api.HandleError(resp, req, err)
			return
		}

		if !current.Editable {
			api.HandleBadRequest(resp, req, fmt.Errorf("systemenv '%s' is not editable", current.EnvName))
			return
		}

		if current.Required && current.Default == "" && it.Value == "" {
			api.HandleBadRequest(resp, req, fmt.Errorf("systemenv '%s' is required", current.EnvName))
			return
		}

		if current.Value != it.Value {
			if err := current.ValidateValue(it.Value); err != nil {
				api.HandleBadRequest(resp, req, err)
				return
			}
			klog.Infof("Updating SystemEnv %s value from '%s' to '%s'", resourceName, current.Value, it.Value)
			original := current.DeepCopy()
			current.Value = it.Value
			if err := h.ctrlClient.Patch(ctx, &current, client.MergeFrom(original)); err != nil {
				api.HandleError(resp, req, err)
				return
			}
		}

		results = append(results, current.EnvVarSpec)
	}

	resp.WriteAsJson(results)
}
