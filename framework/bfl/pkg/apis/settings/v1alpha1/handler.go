package v1alpha1

import (
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"

	"bytetrade.io/web3os/bfl/internal/log"
	"bytetrade.io/web3os/bfl/pkg/api/response"
	"bytetrade.io/web3os/bfl/pkg/apis"
	"bytetrade.io/web3os/bfl/pkg/apis/iam/v1alpha1/operator"
	"bytetrade.io/web3os/bfl/pkg/apiserver/runtime"
	"bytetrade.io/web3os/bfl/pkg/constants"
	"bytetrade.io/web3os/bfl/pkg/userenv"
	"bytetrade.io/web3os/bfl/pkg/utils"

	iamV1alpha2 "github.com/beclab/api/iam/v1alpha2"
	restful "github.com/emicklei/go-restful/v3"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

type Handler struct {
	apis.Base
}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) handleBindingUserZone(req *restful.Request, resp *restful.Response) {
	username, userErr := runtime.RequestUser(req)
	if userErr != nil {
		response.HandleBadRequest(resp, userErr)
		return
	}
	var post PostTerminusName
	err := req.ReadEntity(&post)
	if err != nil {
		response.HandleBadRequest(resp, errors.Errorf("binding zone: %v", err))
		return
	}

	op, err := operator.NewUserOperatorWithContext(req.Request.Context())
	if err != nil {
		response.HandleBadRequest(resp, errors.Errorf("binding zone: %v", err))
		return
	}

	user, err := op.GetUser(username)
	if err != nil {
		response.HandleError(resp, errors.Errorf("binding user zone: get user err, %v", err))
		return
	}

	if v, ok := user.Annotations[constants.UserTerminusWizardStatus]; ok {
		if v != string(constants.WaitActivateVault) {
			response.HandleError(resp, errors.Errorf("user '%s' wizard status err, %s", user.Name, v))
			return
		}
	}

	userPatches := []func(*iamV1alpha2.User){
		func(u *iamV1alpha2.User) {
			u.Annotations[constants.UserAnnotationTerminusNameKey] = u.Name
		},
	}

	if post.JWSSignature != "" {
		userPatches = append(userPatches, func(u *iamV1alpha2.User) {
			u.Annotations[constants.UserCertManagerJWSToken] = post.JWSSignature
		})
	}

	if post.DID != "" {
		userPatches = append(userPatches, func(u *iamV1alpha2.User) {
			u.Annotations[constants.UserCertManagerDID] = post.DID
		})
	}

	userPatches = append(userPatches, func(u *iamV1alpha2.User) {
		u.Annotations[constants.UserTerminusWizardStatus] = string(constants.Completed)
	})

	if err = op.UpdateUser(user, userPatches); err != nil {
		response.HandleError(resp, errors.Errorf("binding user zone err:  %v", err))
		return
	}

	response.SuccessNoData(resp)
}

func (h *Handler) handleGetLauncherAccessPolicy(req *restful.Request, resp *restful.Response) {
	username, userErr := runtime.RequestUser(req)
	if userErr != nil {
		response.HandleBadRequest(resp, userErr)
		return
	}
	userOp, err := operator.NewUserOperatorWithContext(req.Request.Context())
	if err != nil {
		response.HandleError(resp, errors.Errorf("get launcher access policy: new user operator err, %v", err))
		return
	}
	user, err := userOp.GetUser(username)
	if err != nil {
		response.HandleError(resp, errors.Errorf("get launcher access policy: get user err, %v", err))
		return
	}

	var accessLevel AccessLevel
	var allowCIDRs []string
	var authPolicy AuthPolicy

	err = func() error {
		level, err := userOp.GetLauncherAccessLevel(user)
		if err != nil {
			return errors.Errorf("no user access_level")
		}
		if level != nil {
			accessLevel = AccessLevel(*level)
		}
		allowCIDRs = userOp.GetLauncherAllowCIDR(user)

		authPolicy = AuthPolicy(userOp.GetLauncherAuthPolicy(user))
		if authPolicy == "" {
			authPolicy = DefaultAuthPolicy
		}
		return nil
	}()

	if err != nil {
		response.HandleError(resp, errors.Errorf("get launcher access policy: %v", err))
		return
	}

	response.Success(resp, LauncherAccessPolicy{AccessLevel: accessLevel, AllowCIDRs: allowCIDRs, AuthPolicy: authPolicy})
}

func (h *Handler) configDefaultAllowCIDR(req *restful.Request, level AccessLevel) []string {
	var allows []string

	switch level {
	case WorldWide, Public:
		allows = []string{"0.0.0.0/0"}
	case Private:
		allows = []string{"127.0.0.1/32", "192.168.0.0/16", "172.16.0.0/12", "10.0.0.0/8"}

		if external := utils.GetMyExternalIPAddr(); external != "" {
			allows = append(allows, external+"/32")
		}
	}
	return allows
}

func (h *Handler) handleUpdateLauncherAccessPolicy(req *restful.Request, resp *restful.Response) {
	username, userErr := runtime.RequestUser(req)
	if userErr != nil {
		response.HandleBadRequest(resp, userErr)
		return
	}
	var policy LauncherAccessPolicy
	req.ReadEntity(&policy)

	log.Infow("update launcher access policy", "policy", policy)

	if policy.AccessLevel == 0 {
		response.HandleError(resp, errors.New("update launcher access policy: no access level provieded"))
		return
	}

	if policy.AuthPolicy == "" {
		policy.AuthPolicy = DefaultAuthPolicy
	}

	err := func() error {
		userOp, err := operator.NewUserOperatorWithContext(req.Request.Context())
		if err != nil {
			return errors.Errorf("new user operator err, %v", err)
		}
		user, err := userOp.GetUser(username)
		if err != nil {
			return errors.Errorf("get user err, %v", err)
		}

		currentAuthPolicy := userOp.GetLauncherAuthPolicy(user)

		cidrs := userOp.GetLauncherAllowCIDR(user)
		if reflect.DeepEqual(cidrs, policy.AllowCIDRs) {
			if currentAuthPolicy != string(policy.AuthPolicy) {
				if err = userOp.UpdateUser(user, []func(*iamV1alpha2.User){
					func(u *iamV1alpha2.User) {
						u.Annotations[constants.UserLauncherAuthPolicy] = string(policy.AuthPolicy)
					},
				}); err != nil {
					return errors.Errorf("update user err, %v", err)
				}
			}
			return nil
		}

		var ipCIDRs []string

		if len(policy.AllowCIDRs) == 0 {
			ipCIDRs = h.configDefaultAllowCIDR(req, policy.AccessLevel)
		} else {
			for _, cidr := range policy.AllowCIDRs {
				if !strings.Contains(cidr, "/") {
					return errors.Errorf("%q is invalid ip cidr, missing subnet mask, eg: '/24'", cidr)
				}
				_, ipNet, err := net.ParseCIDR(cidr)
				if err != nil {
					return errors.Errorf("parse cidr err, %v", err)
				}
				ipCIDRs = append(ipCIDRs, ipNet.String())
			}
		}
		if policy.AccessLevel == Protected {
			if !utils.ListContains(ipCIDRs, DefaultPodsCIDR) {
				ipCIDRs = append(ipCIDRs, DefaultPodsCIDR)
			}
		}

		if err = userOp.UpdateUser(user, []func(*iamV1alpha2.User){
			func(u *iamV1alpha2.User) {
				u.Annotations[constants.UserLauncherAccessLevel] = fmt.Sprintf("%v", policy.AccessLevel)
				u.Annotations[constants.UserLauncherAllowCIDR] = strings.Join(ipCIDRs, ",")
				u.Annotations[constants.UserLauncherAuthPolicy] = string(policy.AuthPolicy)
			},
		}); err != nil {
			return errors.Errorf("update user err, %v", err)
		}
		return nil
	}()

	if err != nil {
		response.HandleError(resp, errors.Errorf("update launcher access policy: %v", err))
		return
	}

	response.SuccessNoData(resp)
}

func (h *Handler) handleGetPublicDomainAccessPolicy(req *restful.Request, resp *restful.Response) {
	username, userErr := runtime.RequestUser(req)
	if userErr != nil {
		response.HandleBadRequest(resp, userErr)
		return
	}
	userOp, err := operator.NewUserOperatorWithContext(req.Request.Context())
	if err != nil {
		response.HandleError(resp, errors.Errorf("get public domain access policy: new user operator err, %v", err))
		return
	}
	user, err := userOp.GetUser(username)
	if err != nil {
		response.HandleError(resp, errors.Errorf("get public domain access policy: get user err, %v", err))
		return
	}

	var denyAllAnno string = userOp.GetDenyAllPolicy(user)

	var denyAll, _ = strconv.Atoi(denyAllAnno)

	response.Success(resp, PublicDomainAccessPolicy{DenyAll: denyAll}) //  AllowedDomains: strings.Split(allowedDomains, ",")
}

func (h *Handler) handleUpdatePublicDomainAccessPolicy(req *restful.Request, resp *restful.Response) {
	username, userErr := runtime.RequestUser(req)
	if userErr != nil {
		response.HandleBadRequest(resp, userErr)
		return
	}
	var policy PublicDomainAccessPolicy
	req.ReadEntity(&policy)

	log.Infow("update public domain access policy", "policy", policy)

	if policy.DenyAll < 0 || policy.DenyAll > 1 {
		response.HandleError(resp, errors.Errorf("update public domain access policy: deny all %d params invalid", policy.DenyAll))
		return
	}

	err := func() error {
		userOp, err := operator.NewUserOperatorWithContext(req.Request.Context())
		if err != nil {
			return errors.Errorf("new user operator err, %v", err)
		}
		user, err := userOp.GetUser(username)
		if err != nil {
			return errors.Errorf("get user err, %v", err)
		}

		if err = userOp.UpdateUser(user, []func(*iamV1alpha2.User){
			func(u *iamV1alpha2.User) {
				u.Annotations[constants.UserDenyAllPolicy] = strconv.Itoa(policy.DenyAll)
			},
		}); err != nil {
			return errors.Errorf("update user err, %v", err)
		}
		return nil
	}()

	if err != nil {
		response.HandleError(resp, errors.Errorf("update public domain access policy: %v", err))
		return
	}

	response.SuccessNoData(resp)
}

func (h *Handler) handleUpdateLocale(req *restful.Request, resp *restful.Response) {
	username, userErr := runtime.RequestUser(req)
	if userErr != nil {
		response.HandleBadRequest(resp, userErr)
		return
	}
	var locale apis.PostLocale
	userOp, err := operator.NewUserOperatorWithContext(req.Request.Context())
	if err != nil {
		response.HandleError(resp, errors.Errorf("update user locale err: new user operator err, %v", err))
		return
	}

	user, err := userOp.GetUser(username)
	if err != nil {
		response.HandleError(resp, errors.Errorf("update user locale err: get user err, %v", err))
		return
	}

	err = req.ReadEntity(&locale)
	if err != nil {
		klog.Error("read request body error, ", err)
		response.HandleError(resp, errors.Errorf("update user locale data error: %v", err))
		return
	}

	if locale.Location != "" {
		err = userOp.UpdateUser(user, []func(*iamV1alpha2.User){
			func(u *iamV1alpha2.User) {
				u.Annotations[constants.UserLocation] = locale.Location
			},
		})
		if err != nil {
			response.HandleError(resp, errors.Wrap(err, "update user location"))
			return
		}
	}

	// language/theme/timezone are owned by the per-user UserEnvs (see
	// build/user-env.yaml) and persisted directly to the UserEnv CRs as the
	// single source of truth (the env settings page reads/writes the same CRs).
	// These UserEnvs are provisioned when the user reaches the "Created" state,
	// before activation, so they exist by the time this runs.
	kc, err := runtime.NewKubeClientInCluster()
	if err != nil {
		response.HandleError(resp, errors.Errorf("update user locale err: new kube client err, %v", err))
		return
	}
	ctx := req.Request.Context()
	c := kc.CtrlClient()
	envValues := []struct {
		env string
		val string
	}{
		{userenv.EnvLanguage, locale.Language},
		{userenv.EnvTimezone, locale.Timezone},
		{userenv.EnvTheme, locale.Theme},
	}
	for _, ev := range envValues {
		if ev.val == "" {
			continue
		}
		ok, e := userenv.SetValue(ctx, c, username, ev.env, ev.val)
		if e != nil {
			err = e
			response.HandleError(resp, errors.Errorf("update user locale err: set userenv %s err, %v", ev.env, e))
			return
		}
		if !ok {
			err = errors.Errorf("update user locale err: userenv %s not provisioned", ev.env)
			response.HandleError(resp, err)
			return
		}
	}

	response.SuccessNoData(resp)
}

func (h *Handler) handlerUpdateUserLoginBackground(req *restful.Request, resp *restful.Response) {
	username, userErr := runtime.RequestUser(req)
	if userErr != nil {
		response.HandleBadRequest(resp, userErr)
		return
	}
	var background struct {
		Background string `json:"background"`
		Style      string `json:"style"`
	}

	err := req.ReadEntity(&background)
	if err != nil {
		klog.Error("read request body error, ", err)
		response.HandleError(resp, errors.Errorf("update user login background error: %v", err))
		return
	}

	userOp, err := operator.NewUserOperatorWithContext(req.Request.Context())
	if err != nil {
		response.HandleError(resp, errors.Errorf("update user login background err: new user operator err, %v", err))
		return
	}

	user, err := userOp.GetUser(username)
	if err != nil {
		response.HandleError(resp, errors.Errorf("update user login background err: get user err, %v", err))
		return
	}

	err = userOp.UpdateUser(user, []func(*iamV1alpha2.User){
		func(u *iamV1alpha2.User) {
			u.Annotations[constants.UserLoginBackground] = background.Background
			u.Annotations[constants.UserLoginBackgroundStyle] = background.Style
		},
	})

	if err != nil {
		klog.Errorf("update user err, %v", err)
		response.HandleError(resp, errors.Errorf("update user login background error: %v", err))
		return
	}

	response.SuccessNoData(resp)
}

func (h *Handler) handlerUpdateUserAvatar(req *restful.Request, resp *restful.Response) {
	username, userErr := runtime.RequestUser(req)
	if userErr != nil {
		response.HandleBadRequest(resp, userErr)
		return
	}
	var avatar struct {
		Avatar string `json:"avatar"`
	}

	err := req.ReadEntity(&avatar)
	if err != nil {
		klog.Error("read request body error, ", err)
		response.HandleError(resp, errors.Errorf("update user avatar error: %v", err))
		return
	}

	userOp, err := operator.NewUserOperatorWithContext(req.Request.Context())
	if err != nil {
		response.HandleError(resp, errors.Errorf("update user avatar err: new user operator err, %v", err))
		return
	}

	user, err := userOp.GetUser(username)
	if err != nil {
		response.HandleError(resp, errors.Errorf("update user avatar err: get user err, %v", err))
		return
	}

	err = userOp.UpdateUser(user, []func(*iamV1alpha2.User){
		func(u *iamV1alpha2.User) {
			u.Annotations[constants.UserAvatar] = avatar.Avatar
		},
	})

	if err != nil {
		klog.Errorf("update user err, %v", err)
		response.HandleError(resp, errors.Errorf("update user avatar error: %v", err))
		return
	}

	response.SuccessNoData(resp)
}
