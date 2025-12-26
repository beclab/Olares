package operator

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"bytetrade.io/web3os/bfl/pkg/client/dynamic_client/users"
	"bytetrade.io/web3os/bfl/pkg/constants"

	iamV1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type UserOperator struct {
	client         *users.ResourceUserClient
	terminusClient *users.ResourceTerminusClient
	ctx            context.Context
}

func NewUserOperator() (*UserOperator, error) {
	userClient, err := users.NewResourceUserClient()
	if err != nil {
		return nil, err
	}

	terminusClient, err := users.NewResourceTerminusClient()
	if err != nil {
		return nil, err
	}

	return &UserOperator{
		client:         userClient,
		terminusClient: terminusClient,
		ctx:            context.TODO(),
	}, nil
}

func NewUserOperatorWithContext(ctx context.Context) (*UserOperator, error) {
	uop, err := NewUserOperator()
	if err != nil {
		return nil, err
	}
	uop.ctx = ctx
	return uop, nil
}

func NewUserOperatorOrDie() *UserOperator {
	return &UserOperator{
		client: users.NewResourceUserClientOrDie(),
		ctx:    context.TODO(),
	}
}

func (o *UserOperator) ListUsers() ([]*iamV1alpha2.User, error) {
	_users, err := o.client.List(o.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return _users, nil
}

func (o *UserOperator) GetOwnerUser() (*iamV1alpha2.User, error) {
	userList, err := o.ListUsers()
	if err != nil {
		klog.Errorf("get user list error %v", err)
		return nil, err
	}
	for _, u := range userList {
		if u.Annotations[constants.UserAnnotationOwnerRole] == constants.RoleOwner {
			return u, nil
		}
	}
	return nil, errors.New("owner user not found")
}

func (o *UserOperator) GetUser(name string) (*iamV1alpha2.User, error) {
	if name == "" {
		name = constants.Username
	}

	user, err := o.client.Get(o.ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (o *UserOperator) UpdateUser(user *iamV1alpha2.User, applyPatches []func(*iamV1alpha2.User)) error {
	// must reload user to retry
RETRY:
	userNew, err := o.client.Get(o.ctx, user.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	for _, p := range applyPatches {
		p(userNew)
	}

	_, err = o.client.Update(o.ctx, userNew, metav1.UpdateOptions{})
	if err != nil && apierrors.IsConflict(err) {
		goto RETRY
	} else if err != nil {
		return err
	}
	return nil
}

func (o *UserOperator) GetTerminusName(user *iamV1alpha2.User) string {
	return o.GetUserAnnotation(user, constants.UserAnnotationTerminusNameKey)
}

func (o *UserOperator) GetTerminusStatus(user *iamV1alpha2.User) string {
	return o.GetUserAnnotation(user, constants.UserTerminusWizardStatus)
}

func (o *UserOperator) GetReverseProxyType() (string, error) {
	user, err := o.GetUser("")
	if err != nil {
		return "", err
	}
	return o.GetUserAnnotation(user, constants.UserAnnotationReverseProxyType), nil
}

func (o *UserOperator) GetLoginBackground(user *iamV1alpha2.User) string {
	b := o.GetUserAnnotation(user, constants.UserLoginBackground)
	if b == "" {
		return "/bg/0.jpg"
	}

	return b
}

func (o *UserOperator) GetAvatar(user *iamV1alpha2.User) string {
	return o.GetUserAnnotation(user, constants.UserAvatar)
}

func (o *UserOperator) GetUserDID(user *iamV1alpha2.User) string {
	return o.GetUserAnnotation(user, constants.UserCertManagerDID)
}

func (o *UserOperator) BindingTerminusName(user *iamV1alpha2.User, domain string) error {
	if v, ok := user.Annotations[constants.UserAnnotationTerminusNameKey]; ok {
		return fmt.Errorf("user '%s' olaresName is already bind, olaresName: %s", user.Name, v)
	}

	// update terminus name to user annotation
	return o.UpdateUser(user, []func(*iamV1alpha2.User){
		func(u *iamV1alpha2.User) {
			u.Annotations[constants.UserAnnotationTerminusNameKey] = string(constants.NewTerminusName(user.Name, domain))
		},
	})
}

func (o *UserOperator) UnbindingTerminusName(user *iamV1alpha2.User) error {
	return o.UpdateUser(user, []func(*iamV1alpha2.User){
		func(u *iamV1alpha2.User) {
			delete(u.Annotations, constants.UserAnnotationTerminusNameKey)
		},
	})
}

func (o *UserOperator) UpdateAnnotation(user *iamV1alpha2.User, key, value string) error {
	return o.UpdateUser(user, []func(*iamV1alpha2.User){
		func(u *iamV1alpha2.User) {
			u.Annotations[key] = value
		},
	})
}

func (o *UserOperator) AnnotationExists(user *iamV1alpha2.User, key string) bool {
	_, ok := user.Annotations[key]
	return ok
}

func (o *UserOperator) GetUserAnnotation(user *iamV1alpha2.User, name string) string {
	if v, ok := user.Annotations[name]; ok {
		return v
	}
	return ""
}

func (o *UserOperator) GetUserZone(user *iamV1alpha2.User) string {
	zone := o.GetUserAnnotation(user, constants.UserAnnotationZoneKey)
	if zone != "" {
		return zone
	}

	creator := o.GetUserAnnotation(user, constants.AnnotationUserCreator)
	if creator != "" {
		if creator == "cli" {
			oUser, err := o.GetOwnerUser()
			if err != nil {
				klog.Errorf("failed to get user with owner role %v", err)
			}
			if err == nil {
				return oUser.Name
			}
		} else {
			creatorUser, err := o.GetUser(creator)
			if err != nil {
				klog.Errorf("failed to get creator user %v", err)
			}
			if err == nil && creatorUser != nil {
				return o.GetUserAnnotation(creatorUser, constants.UserAnnotationZoneKey)
			}
		}

	}
	return ""
}

func (o *UserOperator) UserIsEphemeralDomain(user *iamV1alpha2.User) (bool, error) {
	isEphemeralAnnotation := o.GetUserAnnotation(user, constants.UserAnnotationIsEphemeral)
	if isEphemeralAnnotation == "" {
		return false, nil
	}
	return strconv.ParseBool(isEphemeralAnnotation)
}

func (o *UserOperator) GetUserDomainType(user *iamV1alpha2.User) (bool, string, error) {
	zone := o.GetUserAnnotation(user, constants.UserAnnotationZoneKey)
	if zone != "" {
		return false, zone, nil
	}

	// Find the creator user's zone
	creatorUserName := o.GetUserAnnotation(user, constants.AnnotationUserCreator)
	if creatorUserName != "" {
		var creatorUser *iamV1alpha2.User
		var err error
		if creatorUserName == "cli" {
			creatorUser, err = o.GetOwnerUser()
			if err != nil {
				return false, "", err
			}
		} else {
			creatorUser, err = o.GetUser(creatorUserName)
			if err != nil {
				return false, "", err
			}
		}

		if v := o.GetUserAnnotation(creatorUser, constants.UserAnnotationZoneKey); v != "" {
			return true, v, nil
		}
	}

	return false, "", nil
}

func (o *UserOperator) GetLauncherAccessLevel(user *iamV1alpha2.User) (*uint64, error) {
	level := o.GetUserAnnotation(user, constants.UserLauncherAccessLevel)
	if level == "" {
		return nil, fmt.Errorf("do not configuration yet")
	}

	parsedLevel, err := strconv.ParseUint(level, 10, 64)
	if err != nil {
		return nil, err
	}
	return &parsedLevel, nil
}

func (o *UserOperator) GetLauncherAllowCIDR(user *iamV1alpha2.User) []string {
	cidr := o.GetUserAnnotation(user, constants.UserLauncherAllowCIDR)
	return strings.Split(cidr, ",")
}

func (o *UserOperator) GetLauncherAuthPolicy(user *iamV1alpha2.User) string {
	policy := o.GetUserAnnotation(user, constants.UserLauncherAuthPolicy)

	return policy
}

func (o *UserOperator) GetDenyAllPolicy(user *iamV1alpha2.User) string {
	policy := o.GetUserAnnotation(user, constants.UserDenyAllPolicy)
	return policy
}

func (o *UserOperator) GetDomain() (string, error) {
	terminus, err := o.terminusClient.Get(o.ctx, "terminus", metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	if name, ok := terminus.Spec.Settings[users.SettingsDomainNameKey]; !ok {
		return "", errors.New("olares domain name not found")
	} else {
		return name, nil
	}

}

func (o *UserOperator) SelfhostedAndOsVersion() (selfhosted, terminusd bool, version string, err error) {
	terminus, err := o.terminusClient.Get(o.ctx, "terminus", metav1.GetOptions{})
	if err != nil {
		return false, false, "", err
	}

	version = terminus.Spec.Version

	if selfhostedValue, ok := terminus.Spec.Settings[users.SettingsSelfhostedKey]; !ok {
		selfhosted = true
		return
	} else {
		selfhosted, err = strconv.ParseBool(selfhostedValue)
		if err != nil {
			klog.Error("parse selfhosted value error", err)
			return
		}
	}

	if terminusdValue, ok := terminus.Spec.Settings[users.SettingsTerminusdKey]; !ok {
		terminusd = false
		return
	} else {
		terminusd = terminusdValue == "1"
	}

	return
}

func (o *UserOperator) GetTerminusID() (string, error) {
	terminus, err := o.terminusClient.Get(o.ctx, "terminus", metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return string(terminus.ObjectMeta.UID), nil

}
