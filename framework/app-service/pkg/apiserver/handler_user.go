package apiserver

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"

	"github.com/beclab/Olares/framework/app-service/pkg/users"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"

	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/emicklei/go-restful/v3"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
)

const userAnnotationLimitsCpuKey = "bytetrade.io/user-cpu-limit"
const userAnnotationLimitsMemoryKey = "bytetrade.io/user-memory-limit"

func (h *Handler) userResourceStatus(req *restful.Request, resp *restful.Response) {
	username := req.PathParameter(ParamUserName)
	metrics, err := prometheus.GetCurUserResource(req.Request.Context(), username)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	resp.WriteAsJson(metrics)
}

func (h *Handler) curUserResource(req *restful.Request, resp *restful.Response) {
	username := req.Attribute(constants.UserContextAttribute).(string)
	metrics, err := prometheus.GetCurUserResource(req.Request.Context(), username)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	resp.WriteAsJson(metrics)
}

func (h *Handler) userInfo(req *restful.Request, resp *restful.Response) {
	username := req.Attribute(constants.UserContextAttribute).(string)
	gvr := schema.GroupVersionResource{
		Group:    "iam.kubesphere.io",
		Version:  "v1alpha2",
		Resource: "users",
	}
	client, err := dynamic.NewForConfig(h.kubeConfig)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	user, err := client.Resource(gvr).Get(req.Request.Context(), username, metav1.GetOptions{})
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	annotations := user.GetAnnotations()
	role := annotations["bytetrade.io/owner-role"]
	userInfo := map[string]string{
		"username": user.GetName(),
		"role":     role,
	}

	resp.WriteAsJson(userInfo)
}

type UserCreateRequest struct {
	Name        string `json:"name"`
	OwnerRole   string `json:"owner_role"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Description string `json:"description"`
	MemoryLimit string `json:"memory_limit"`
	CpuLimit    string `json:"cpu_limit"`
}

type UserCreateOption struct {
	UserCreateRequest
	TerminusName string
}

func (h *Handler) createUser(req *restful.Request, resp *restful.Response) {
	owner := req.Attribute(constants.UserContextAttribute).(string)

	var userReq UserCreateRequest
	err := req.ReadEntity(&userReq)
	if err != nil {
		api.HandleBadRequest(resp, req, errors.Errorf("user create: %v", err))
		return
	}

	klog.Infof("userReq: %#v", userReq)

	//// Basic field validation
	if userReq.Name == "" || userReq.Password == "" {
		api.HandleBadRequest(resp, req, errors.New("user create: no username or password provided"))
		return
	}

	if userReq.MemoryLimit == "" {
		api.HandleBadRequest(resp, req, errors.New("user create: memory_limit cannot be empty"))
		return
	}

	if userReq.CpuLimit == "" {
		api.HandleBadRequest(resp, req, errors.New("user create: cpu_limit cannot be empty"))
		return
	}

	if userReq.OwnerRole == "" || (userReq.OwnerRole != "owner" && userReq.OwnerRole != "admin" && userReq.OwnerRole != "normal") {
		api.HandleBadRequest(resp, req, errors.New("user create: invalid owner_role"))
		return
	}

	username := strings.ToLower(userReq.Name)
	// get terminusName
	var olaresName users.OlaresName
	if strings.Contains(username, "@") {
		olaresName = users.OlaresName(username)
		username = strings.Split(username, "@")[0]
	} else {
		olares, err := utils.GetTerminus(req.Request.Context(), h.ctrlClient)
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}
		domainName := olares.Spec.Settings["domainName"]
		if domainName == "" {
			api.HandleError(resp, req, errors.New("empty domainName"))
			return
		}
		olaresName = users.NewOlaresName(username, domainName)
	}

	user := iamv1alpha2.User{
		TypeMeta: metav1.TypeMeta{
			APIVersion: iamv1alpha2.SchemeGroupVersion.String(),
			Kind:       iamv1alpha2.ResourceKindUser,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: userReq.Name,
			Annotations: map[string]string{
				"iam.kubesphere.io/uninitialized":   "true",
				users.AnnotationUserCreator:         owner,
				users.UserAnnotationUninitialized:   "true",
				users.UserAnnotationOwnerRole:       userReq.OwnerRole,
				users.UserAnnotationIsEphemeral:     "true", // add is-ephemeral flag, 'true' it means need temporary domain
				users.UserAnnotationTerminusNameKey: string(olaresName),
				users.UserLauncherAuthPolicy:        "two_factor",
				users.UserLauncherAccessLevel:       "1",
				users.UserAnnotationLimitsMemoryKey: userReq.MemoryLimit,
				users.UserAnnotationLimitsCpuKey:    userReq.CpuLimit,
				"iam.kubesphere.io/sync-to-lldap":   "true",
				"iam.kubesphere.io/synced-to-lldap": "false",
			},
		},
		Spec: iamv1alpha2.UserSpec{
			DisplayName:     userReq.DisplayName,
			Email:           string(olaresName),
			InitialPassword: userReq.Password,
			Description:     userReq.Description,
		},
	}
	if userReq.OwnerRole == "owner" || userReq.OwnerRole == "admin" {
		user.Spec.Groups = append(user.Spec.Groups, "lldap_admin")
	}
	err = h.ctrlClient.Create(req.Request.Context(), &user)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteAsJson(map[string]interface{}{
		"code": 200,
		"data": map[string]string{
			"name": userReq.Name,
		},
	})
	return
}

func (h *Handler) deleteUser(req *restful.Request, resp *restful.Response) {
	owner := req.Attribute(constants.UserContextAttribute).(string)

	username := req.PathParameter("user")
	if username == "" {
		api.HandleBadRequest(resp, req, errors.New("user delete: no username provided"))
		return
	}

	var user iamv1alpha2.User
	err := h.ctrlClient.Get(req.Request.Context(), types.NamespacedName{Name: username}, &user)
	if err != nil && !apierrors.IsNotFound(err) {
		api.HandleError(resp, req, err)
		return
	}
	if err != nil {
		api.HandleBadRequest(resp, req, fmt.Errorf("user %s not found", username))
		return
	}
	user.Annotations[users.AnnotationUserDeleter] = owner
	err = h.ctrlClient.Update(req.Request.Context(), &user)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	if user.Status.State == "Creating" {
		api.HandleBadRequest(resp, req, fmt.Errorf("user %s is under creating", username))
		return
	}
	err = h.ctrlClient.Delete(req.Request.Context(), &user)
	if err != nil && !apierrors.IsNotFound(err) {
		api.HandleError(resp, req, err)
		return
	}
	resp.WriteAsJson(map[string]interface{}{
		"code": 200,
		"data": map[string]string{
			"name": username,
		},
	})
	return
}

func (h *Handler) userStatus(req *restful.Request, resp *restful.Response) {
	username := req.PathParameter("user")
	if username == "" {
		api.HandleBadRequest(resp, req, errors.New("user delete: no username provided"))
		return
	}

	var user iamv1alpha2.User
	err := h.ctrlClient.Get(req.Request.Context(), types.NamespacedName{Name: username}, &user)
	if err != nil && !apierrors.IsNotFound(err) {
		api.HandleError(resp, req, err)
		return
	}
	if err != nil {
		resp.WriteAsJson(map[string]interface{}{
			"code": 200,
			"data": map[string]interface{}{
				"name":   username,
				"status": "Deleted",
			},
			"message": fmt.Sprintf("user %s not exists", username),
		})
		return
	}

	isEphemeral, zone, err := h.getUserDomainType(&user)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	var address string
	if isEphemeral {
		address = fmt.Sprintf("wizard-%s.%s", username, zone)
	}

	resp.WriteAsJson(map[string]interface{}{
		"code": 200,
		"data": map[string]interface{}{
			"name":    username,
			"status":  user.Status.State,
			"message": user.Status.Reason,
			"address": map[string]string{
				"wizard": address,
			},
		},
		"message": user.Status.Reason,
	})
	return
}

func (h *Handler) getUserDomainType(user *iamv1alpha2.User) (bool, string, error) {
	zone := user.Annotations[users.UserAnnotationZoneKey]
	if zone != "" {
		return false, zone, nil
	}
	creatorUser, err := utils.FindOwnerUser(h.ctrlClient, user)
	if err != nil {
		klog.Error(err)
		return false, "", err
	}
	if v := creatorUser.Annotations[users.UserAnnotationZoneKey]; v != "" {
		return true, v, nil
	}

	return false, "", nil
}

type UserResourceLimit struct {
	MemoryLimit string `json:"memory_limit"`
	CpuLimit    string `json:"cpu_limit"`
}

func (h *Handler) handleUpdateUserLimits(req *restful.Request, resp *restful.Response) {
	var userResourceLimits UserResourceLimit
	if err := req.ReadEntity(&userResourceLimits); err != nil {
		api.HandleBadRequest(resp, req, errors.Errorf("update user's resource limit: %v", err))
		return
	}

	memory, err := resource.ParseQuantity(userResourceLimits.MemoryLimit)
	if err != nil {
		api.HandleBadRequest(resp, req, errors.New("user create: invalid format of memory limit"))
		return
	}

	cpu, err := resource.ParseQuantity(userResourceLimits.CpuLimit)
	if err != nil {
		api.HandleBadRequest(resp, req, errors.New("user create: invalid format of cpu limit"))
		return
	}

	defaultMemoryLimit, _ := resource.ParseQuantity(os.Getenv("USER_DEFAULT_MEMORY_LIMIT"))
	defaultCpuLimit, _ := resource.ParseQuantity(os.Getenv("USER_DEFAULT_CPU_LIMIT"))

	if defaultMemoryLimit.CmpInt64(int64(memory.AsApproximateFloat64())) > 0 {
		api.HandleBadRequest(resp, req, errors.Errorf("user create: memory limit can not less than %s",
			defaultMemoryLimit.String()))
		return
	}

	if defaultCpuLimit.CmpInt64(int64(cpu.AsApproximateFloat64())) > 0 {
		api.HandleBadRequest(resp, req, errors.Errorf("user create: cpu limit can not less than %s core",
			defaultCpuLimit.String()))
		return
	}

	username := req.PathParameter("user")
	var user iamv1alpha2.User
	err = h.ctrlClient.Get(req.Request.Context(), types.NamespacedName{Name: username}, &user)
	if err != nil {
		api.HandleError(resp, req, errors.Errorf("get user err: %v", err))
		return
	}

	user.Annotations[userAnnotationLimitsMemoryKey] = userResourceLimits.MemoryLimit
	user.Annotations[userAnnotationLimitsCpuKey] = userResourceLimits.CpuLimit
	err = h.ctrlClient.Update(req.Request.Context(), &user)
	if err != nil {
		api.HandleError(resp, req, errors.Errorf("update user err: %v", err))
		return
	}
	resp.WriteAsJson(map[string]interface{}{
		"code": 200,
	})

}

func (h *Handler) handleUsers(req *restful.Request, resp *restful.Response) {
	var userList iamv1alpha2.UserList
	err := h.ctrlClient.List(req.Request.Context(), &userList)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	userInfo := make([]UserInfo, 0)
	for _, user := range userList.Items {
		var roles []string
		roles = []string{user.Annotations[users.UserAnnotationOwnerRole]}
		u := UserInfo{
			UID:               string(user.UID),
			Name:              user.Name,
			DisplayName:       user.Spec.DisplayName,
			Description:       user.Spec.Description,
			Email:             user.Spec.Email,
			State:             string(user.Status.State),
			CreationTimestamp: user.CreationTimestamp.Unix(),
			Roles:             roles,
			Avatar:            "",
		}

		if terminusName, ok := user.Annotations[users.UserAnnotationTerminusNameKey]; ok {
			u.TerminusName = terminusName
		}

		if avatar, ok := user.Annotations[users.UserAvatar]; ok {
			u.Avatar = avatar
		}

		if memoryLimit, ok := user.Annotations[users.UserAnnotationLimitsMemoryKey]; ok {
			u.MemoryLimit = memoryLimit
		}
		if cpuLimit, ok := user.Annotations[users.UserAnnotationLimitsCpuKey]; ok {
			u.CpuLimit = cpuLimit
		}

		if user.Status.LastLoginTime != nil {
			u.LastLoginTime = pointer.Int64(user.Status.LastLoginTime.Unix())
		}
		if zone, ok := user.Annotations[users.UserAnnotationZoneKey]; ok {
			u.Zone = zone
		}

		u.WizardComplete = getWizardComplete(&user)
		userInfo = append(userInfo, u)
	}
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}
	resp.WriteAsJson(NewListResult(userInfo))
}

func (h *Handler) handleUser(req *restful.Request, resp *restful.Response) {
	username := req.PathParameter(ParamUserName)

	var user iamv1alpha2.User
	err := h.ctrlClient.Get(req.Request.Context(), types.NamespacedName{Name: username}, &user)
	if err != nil {
		if apierrors.IsNotFound(err) {
			api.HandleNotFound(resp, req, fmt.Errorf("user %s not found", username))
		}
		api.HandleError(resp, req, err)

		return
	}
	roles := []string{user.Annotations[users.UserAnnotationOwnerRole]}

	u := UserInfo{
		UID:               string(user.UID),
		Name:              user.Name,
		DisplayName:       user.Spec.DisplayName,
		Description:       user.Spec.Description,
		Email:             user.Spec.Email,
		State:             string(user.Status.State),
		CreationTimestamp: user.CreationTimestamp.Unix(),
	}

	if terminusName, ok := user.Annotations[users.UserAnnotationTerminusNameKey]; ok {
		u.TerminusName = terminusName
	}

	if avatar, ok := user.Annotations[users.UserAvatar]; ok {
		u.Avatar = avatar
	}

	if memoryLimit, ok := user.Annotations[users.UserAnnotationLimitsMemoryKey]; ok {
		u.MemoryLimit = memoryLimit
	}
	if cpuLimit, ok := user.Annotations[users.UserAnnotationLimitsCpuKey]; ok {
		u.CpuLimit = cpuLimit
	}

	if user.Status.LastLoginTime != nil {
		u.LastLoginTime = pointer.Int64(user.Status.LastLoginTime.Unix())
	}
	u.Roles = roles

	u.WizardComplete = getWizardComplete(&user)
	resp.WriteAsJson(map[string]interface{}{
		"code": 200,
		"data": u,
	})

}

func (h *Handler) getRolesByUserName(ctx context.Context, name string) ([]string, error) {
	var globalRoleBindingList iamv1alpha2.GlobalRoleBindingList
	err := h.ctrlClient.List(ctx, &globalRoleBindingList)
	if err != nil {
		return nil, err
	}

	var roles = sets.NewString()

	for _, binding := range globalRoleBindingList.Items {
		for _, subject := range binding.Subjects {
			if subject.Name == name {
				roles.Insert(binding.RoleRef.Name)
			}
		}
	}
	return roles.List(), nil
}

func getWizardComplete(user *iamv1alpha2.User) bool {
	if user == nil {
		return false
	}
	if wizardStatus, ok := user.Annotations["bytetrade.io/wizard-status"]; ok {
		if wizardStatus == "completed" || wizardStatus == "wait_reset_password" {
			return true
		}
	}
	return false
}

type UserInfo struct {
	UID               string `json:"uid"`
	Name              string `json:"name"`
	DisplayName       string `json:"display_name"`
	Description       string `json:"description"`
	Email             string `json:"email"`
	State             string `json:"state"`
	LastLoginTime     *int64 `json:"last_login_time"`
	CreationTimestamp int64  `json:"creation_timestamp"`
	Avatar            string `json:"avatar"`
	Zone              string `json:"zone"`

	TerminusName   string `json:"terminusName"`
	WizardComplete bool   `json:"wizard_complete"`

	Roles []string `json:"roles"`

	MemoryLimit string `json:"memory_limit"`
	CpuLimit    string `json:"cpu_limit"`
}
