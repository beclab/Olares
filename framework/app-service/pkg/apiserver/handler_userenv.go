package apiserver

import (
	"context"
	"fmt"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	sysv1alpha1 "github.com/beclab/Olares/framework/app-service/api/sys.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"

	"github.com/emicklei/go-restful/v3"
	coordinationv1 "k8s.io/api/coordination/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type UserEnvUpdateRequest struct {
	Value string `json:"value"`
}

// UserEnvDetail extends UserEnv with reference information
type UserEnvDetail struct {
	sysv1alpha1.EnvVarSpec `json:",inline"`
	ReferencedBy           []AppEnvReferrer `json:"referencedBy"`
}

func (h *Handler) createUserEnv(req *restful.Request, resp *restful.Response) {
	username := getCurrentUser(req)

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

	if err := body.ValidateValue(body.Value); err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}
	if err := body.ValidateValue(body.Default); err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	// Create UserEnv in the user's namespace
	userNamespace := utils.UserspaceName(username)
	obj := &sysv1alpha1.UserEnv{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: userNamespace,
		},
		EnvVarSpec: body,
	}

	if err := h.ctrlClient.Create(req.Request.Context(), obj); err != nil {
		api.HandleError(resp, req, err)
		return
	}
	resp.WriteAsJson(obj.EnvVarSpec)
}

func (h *Handler) updateUserEnv(req *restful.Request, resp *restful.Response) {
	username := getCurrentUser(req)

	name := req.PathParameter(ParamEnvName)
	if name == "" {
		api.HandleBadRequest(resp, req, fmt.Errorf("userenv name is required"))
		return
	}

	var body UserEnvUpdateRequest
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

	userNamespace := utils.UserspaceName(username)
	var current sysv1alpha1.UserEnv
	if err := h.ctrlClient.Get(ctx, types.NamespacedName{Name: resourceName, Namespace: userNamespace}, &current); err != nil {
		api.HandleError(resp, req, err)
		return
	}

	if !current.Editable {
		api.HandleBadRequest(resp, req, fmt.Errorf("userenv '%s' is not editable", current.EnvName))
		return
	}

	if current.Required && current.GetEffectiveValue() == "" && body.Value == "" {
		api.HandleBadRequest(resp, req, fmt.Errorf("userenv '%s' is required", current.EnvName))
		return
	}

	if current.Value != body.Value {
		err := current.ValidateValue(body.Value)
		if err != nil {
			api.HandleBadRequest(resp, req, err)
			return
		}
		klog.Infof("Updating UserEnv %s/%s value from '%s' to '%s'", userNamespace, resourceName, current.Value, body.Value)
		original := current.DeepCopy()
		current.Value = body.Value
		if err := h.ctrlClient.Patch(ctx, &current, client.MergeFrom(original)); err != nil {
			api.HandleError(resp, req, err)
			return
		}
	}

	resp.WriteAsJson(current.EnvVarSpec)
}

func (h *Handler) deleteUserEnv(req *restful.Request, resp *restful.Response) {
	username := getCurrentUser(req)

	name := req.PathParameter(ParamEnvName)
	if name == "" {
		api.HandleBadRequest(resp, req, fmt.Errorf("userenv name is required"))
		return
	}

	ctx := req.Request.Context()
	resourceName, err := utils.EnvNameToResourceName(name)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	userNamespace := utils.UserspaceName(username)
	var current sysv1alpha1.UserEnv
	if err := h.ctrlClient.Get(ctx, types.NamespacedName{Name: resourceName, Namespace: userNamespace}, &current); err != nil {
		if apierrors.IsNotFound(err) {
			resp.WriteEntity(api.Response{Code: 200})
			return
		}
		api.HandleError(resp, req, err)
		return
	}
	if current.Required {
		api.HandleBadRequest(resp, req, fmt.Errorf("userenv '%s' is required", current.EnvName))
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

// listUserEnvs returns all user env specs for the current user
func (h *Handler) listUserEnvs(req *restful.Request, resp *restful.Response) {
	username := getCurrentUser(req)

	userNamespace := utils.UserspaceName(username)
	var list sysv1alpha1.UserEnvList
	if err := h.ctrlClient.List(req.Request.Context(), &list, client.InNamespace(userNamespace)); err != nil {
		api.HandleError(resp, req, err)
		return
	}

	result := make([]sysv1alpha1.EnvVarSpec, 0, len(list.Items))
	for _, item := range list.Items {
		result = append(result, item.EnvVarSpec)
	}
	resp.WriteAsJson(result)
}

// getUserEnvDetail returns a user env spec along with referencing app envs
func (h *Handler) getUserEnvDetail(req *restful.Request, resp *restful.Response) {
	username := getCurrentUser(req)

	name := req.PathParameter(ParamEnvName)
	if name == "" {
		api.HandleBadRequest(resp, req, fmt.Errorf("userenv name is required"))
		return
	}

	ctx := req.Request.Context()
	resourceName, err := utils.EnvNameToResourceName(name)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	userNamespace := utils.UserspaceName(username)
	var current sysv1alpha1.UserEnv
	if err := h.ctrlClient.Get(ctx, types.NamespacedName{Name: resourceName, Namespace: userNamespace}, &current); err != nil {
		api.HandleError(resp, req, err)
		return
	}

	detail := UserEnvDetail{EnvVarSpec: current.EnvVarSpec}

	var appEnvList sysv1alpha1.AppEnvList
	if err := h.ctrlClient.List(ctx, &appEnvList); err != nil {
		api.HandleError(resp, req, err)
		return
	}
	for _, ae := range appEnvList.Items {
		// Only check AppEnvs that belong to the same user
		if ae.AppOwner != username {
			continue
		}
		for _, envVar := range ae.Envs {
			if envVar.ValueFrom != nil && envVar.ValueFrom.EnvName == current.EnvName {
				detail.ReferencedBy = append(detail.ReferencedBy, AppEnvReferrer{
					AppName:   ae.AppName,
					AppOwner:  ae.AppOwner,
					Namespace: ae.Namespace,
				})
				break
			}
		}
	}

	resp.WriteAsJson(detail)
}

func (h *Handler) acquireUserEnvBatchLease(ctx context.Context, userNamespace string) (func(), error) {
	const leaseName = "env-batch-lock"
	lease := &coordinationv1.Lease{}
	holder := fmt.Sprintf("app-service-%d", time.Now().UnixNano())
	duration := int32(5)
	now := metav1.MicroTime{Time: time.Now()}

	if err := h.ctrlClient.Get(ctx, types.NamespacedName{Name: leaseName, Namespace: userNamespace}, lease); err != nil {
		if apierrors.IsNotFound(err) {
			lease = &coordinationv1.Lease{
				ObjectMeta: metav1.ObjectMeta{Name: leaseName, Namespace: userNamespace},
				Spec: coordinationv1.LeaseSpec{
					HolderIdentity:       &holder,
					LeaseDurationSeconds: &duration,
					AcquireTime:          &now,
					RenewTime:            &now,
				},
			}
			if err := h.ctrlClient.Create(ctx, lease); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		if lease.Spec.RenewTime == nil || lease.Spec.LeaseDurationSeconds == nil ||
			time.Now().After(lease.Spec.RenewTime.Add(time.Duration(*lease.Spec.LeaseDurationSeconds)*time.Second)) {
			original := lease.DeepCopy()
			lease.Spec.HolderIdentity = &holder
			lease.Spec.LeaseDurationSeconds = &duration
			lease.Spec.RenewTime = &now
			if lease.Spec.AcquireTime == nil {
				lease.Spec.AcquireTime = &now
			}
			if err := h.ctrlClient.Patch(ctx, lease, client.MergeFrom(original)); err != nil {
				return nil, err
			}
		}
	}

	release := func() {
		// Best-effort delete to release immediately; controller also relies on expiry
		l := &coordinationv1.Lease{}
		if err := h.ctrlClient.Get(ctx, types.NamespacedName{Name: leaseName, Namespace: userNamespace}, l); err == nil {
			_ = h.ctrlClient.Delete(ctx, l)
		}
	}
	return release, nil
}

func (h *Handler) batchUpdateUserEnvs(req *restful.Request, resp *restful.Response) {
	username := getCurrentUser(req)

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
	userNamespace := utils.UserspaceName(username)
	results := make([]sysv1alpha1.EnvVarSpec, 0, len(items))

	dedup := make(map[string]string)
	for _, it := range items {
		if it.EnvName == "" {
			api.HandleBadRequest(resp, req, fmt.Errorf("userenv name is required"))
			return
		}
		resourceName, err := utils.EnvNameToResourceName(it.EnvName)
		if err != nil {
			api.HandleBadRequest(resp, req, err)
			return
		}
		dedup[resourceName] = it.Value
	}

	if len(dedup) > 1 {
		release, err := h.acquireUserEnvBatchLease(ctx, userNamespace)
		if err != nil {
			klog.Errorf("Failed to acquire user env batch lease: %v", err)
		}
		if release != nil {
			defer release()
		}
	}

	for resourceName, value := range dedup {
		var current sysv1alpha1.UserEnv
		if err := h.ctrlClient.Get(ctx, types.NamespacedName{Name: resourceName, Namespace: userNamespace}, &current); err != nil {
			api.HandleError(resp, req, err)
			return
		}

		if !current.Editable {
			api.HandleBadRequest(resp, req, fmt.Errorf("userenv '%s' is not editable", current.EnvName))
			return
		}

		if current.Required && current.GetEffectiveValue() == "" && value == "" {
			api.HandleBadRequest(resp, req, fmt.Errorf("userenv '%s' is required", current.EnvName))
			return
		}

		if current.Value != value {
			if err := current.ValidateValue(value); err != nil {
				api.HandleBadRequest(resp, req, err)
				return
			}
			klog.Infof("Updating UserEnv %s/%s value from '%s' to '%s'", userNamespace, resourceName, current.Value, value)
			original := current.DeepCopy()
			current.Value = value
			if err := h.ctrlClient.Patch(ctx, &current, client.MergeFrom(original)); err != nil {
				api.HandleError(resp, req, err)
				return
			}
		}

		results = append(results, current.EnvVarSpec)
	}

	resp.WriteAsJson(results)
}
