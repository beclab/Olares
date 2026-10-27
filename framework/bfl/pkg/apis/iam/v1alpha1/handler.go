package v1alpha1

import (
	"context"
	"fmt"
	"net/http"

	"bytetrade.io/web3os/bfl/internal/log"
	"bytetrade.io/web3os/bfl/pkg/api"
	"bytetrade.io/web3os/bfl/pkg/api/response"
	app_service "bytetrade.io/web3os/bfl/pkg/app_service/v1"
	"bytetrade.io/web3os/bfl/pkg/constants"
	"bytetrade.io/web3os/bfl/pkg/lldap"
	iamV1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/emicklei/go-restful/v3"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var defaultGlobalRoles = []string{
	constants.RoleOwner,
	// "platform-regular",    useless role
	// "users-manager",
	constants.RoleAdmin,
	constants.RoleOwner,
}

type Handler struct {
	ctrlClient client.Client
}

func New(ctrlClient client.Client) *Handler {
	return &Handler{
		ctrlClient: ctrlClient,
	}
}

func (h *Handler) getRolesByUserName(ctx context.Context, name string) ([]string, error) {
	var globalRoleBindings iamV1alpha2.GlobalRoleBindingList
	err := h.ctrlClient.List(ctx, &globalRoleBindings)
	if err != nil {
		return nil, err
	}

	var roles = sets.NewString()

	for _, binding := range globalRoleBindings.Items {
		for _, subject := range binding.Subjects {
			if subject.Name == name {
				roles.Insert(binding.RoleRef.Name)
			}
		}
	}
	return roles.List(), nil
}

func (h *Handler) listUsers(ctx context.Context) ([]iamV1alpha2.User, error) {
	var users iamV1alpha2.UserList
	err := h.ctrlClient.List(ctx, &users)
	if err != nil {
		return nil, err
	}

	return users.Items, nil
}

func (h *Handler) handleListUsers(req *restful.Request, resp *restful.Response) {
	users, err := h.listUsers(req.Request.Context())
	if err != nil {
		if apierrors.IsNotFound(err) {
			response.Success(resp, []any{})
			return
		}
		response.HandleError(resp, errors.Errorf("list users: %v", err))
		return
	}

	usersInfo := make([]UserInfo, 0)

	for _, user := range users {
		var roles []string

		roles, err = h.getRolesByUserName(req.Request.Context(), user.Name)
		if err != nil {
			break
		}

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
		// terminus name
		if terminusName, ok := user.Annotations[constants.UserAnnotationTerminusNameKey]; ok {
			u.TerminusName = terminusName
		}

		if avatar, ok := user.Annotations[constants.UserAvatar]; ok {
			u.Avatar = avatar
		}

		if memoryLimit, ok := user.Annotations[constants.UserAnnotationLimitsMemoryKey]; ok {
			u.MemoryLimit = memoryLimit
		}
		if cpuLimit, ok := user.Annotations[constants.UserAnnotationLimitsCpuKey]; ok {
			u.CpuLimit = cpuLimit
		}

		if user.Status.LastLoginTime != nil {
			u.LastLoginTime = pointer.Int64(user.Status.LastLoginTime.Unix())
		}

		if status := user.Annotations[constants.UserTerminusWizardStatus]; status == string(constants.Completed) {
			u.WizardComplete = true
		}

		usersInfo = append(usersInfo, u)
	}

	if err != nil {
		response.HandleError(resp, errors.Errorf("list users: %v", err))
		return
	}

	response.Success(resp, api.NewListResult(usersInfo))
}

func (h *Handler) handleListUserLoginRecords(req *restful.Request, resp *restful.Response) {
	name := req.PathParameter("user")

	users, err := h.listUsers(req.Request.Context())
	if err != nil {
		response.HandleError(resp, errors.Errorf("list user login records: %v", err))
		return
	}

	userIsExists := func() bool {
		for _, user := range users {
			if user.Name == name {
				return true
			}
		}
		return false
	}
	if !userIsExists() {
		response.HandleError(resp, errors.Errorf("list user login records: user %q not exists", name))
		return
	}
	lldapClient, err := lldap.New()
	if err != nil {
		log.Errorf("make lldap client err=%v", err)
		return
	}
	loginRecords, err := lldapClient.Users().LoginRecords(req.Request.Context(), name)
	if err != nil {
		response.HandleError(resp, errors.Errorf("list user login records: %v", err))
		return
	}
	records := make([]LoginRecord, 0)
	for _, r := range loginRecords {
		records = append(records, LoginRecord{
			Type:      "Token",
			Success:   r.Success,
			UserAgent: r.UserAgent,
			Reason:    r.Reason,
			SourceIP:  r.SourceIp,
			LoginTime: func() *int64 {
				t := r.CreationDate.Unix()
				return &t
			}(),
		})
	}
	response.Success(resp, api.NewListResult(records))
}

func (h *Handler) handleListUserRoles(_ *restful.Request, resp *restful.Response) {
	response.Success(resp, api.NewListResult(defaultGlobalRoles))
}

func (h *Handler) handleResetUserPassword(req *restful.Request, resp *restful.Response) {
	var passwordReset PasswordReset
	if err := req.ReadEntity(&passwordReset); err != nil {
		response.HandleBadRequest(resp, errors.Errorf("reset password: %v", err))
		return
	}
	token := req.HeaderParameter(constants.UserAuthorizationTokenKey)

	if passwordReset.Password == "" {
		response.HandleError(resp, errors.New("reset password: new password is empty"))
		return
	}

	if passwordReset.Password == passwordReset.CurrentPassword {
		response.HandleBadRequest(resp, errors.New("reset password: the tow passwords must be different"))
		return
	}

	log.Info("start reset user password")

	userName := req.PathParameter("user")
	var user iamV1alpha2.User
	err := h.ctrlClient.Get(req.Request.Context(), types.NamespacedName{Name: userName}, &user)
	if err != nil {
		response.HandleError(resp, errors.Errorf("reset password: get user err, %v", err))
		return
	}

	url := fmt.Sprintf("http://authelia-backend.os-framework:9091/api/reset/%s/password", userName)
	client := resty.New()
	res, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader(constants.UserAuthorizationTokenKey, token).
		SetHeader(constants.HeaderBflUserKey, userName).
		SetBody(&passwordReset).
		Post(url)
	if err != nil {
		response.HandleError(resp, errors.Errorf("reset password: request authelia failed %v", err))
		return
	}
	if res.StatusCode() != http.StatusOK {
		response.HandleError(resp, errors.New(res.String()))
		return
	}

	response.SuccessNoData(resp)
}

func (h *Handler) handleGetUserMetrics(req *restful.Request, resp *restful.Response) {
	user := req.PathParameter("user")
	token := req.HeaderParameter(constants.UserAuthorizationTokenKey)
	appServiceClient := app_service.NewAppServiceClient().ForUser(user)

	r, err := appServiceClient.GetUserMetrics(user, token)
	if err != nil {
		response.HandleError(resp, err)
	}
	resp.WriteAsJson(r)
}
