package operator

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"bytetrade.io/web3os/bfl/pkg/apiserver/runtime"
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

func (o *UserOperator) ListUsers() ([]*iamV1alpha2.User, error) {
	_users, err := o.client.List(o.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return _users, nil
}

func (o *UserOperator) GetUser(name string) (*iamV1alpha2.User, error) {
	if name == "" {
		return nil, errors.New("username is required")
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

	if userNew.Annotations == nil {
		userNew.Annotations = make(map[string]string)
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
	kc, err := runtime.NewKubeClientInCluster()
	if err != nil {
		return "", err
	}
	cm, err := kc.Kubernetes().CoreV1().ConfigMaps("os-network").Get(o.ctx, "default-network", metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return constants.ReverseProxyTypeNone, nil
	}
	if err != nil {
		return "", err
	}
	var spec struct {
		Config struct {
			Mode string `json:"mode"`
		} `json:"config"`
	}
	if raw := cm.Data["proxy"]; raw != "" {
		if err := json.Unmarshal([]byte(raw), &spec); err != nil {
			return "", err
		}
	}
	switch spec.Config.Mode {
	case "frp":
		return constants.ReverseProxyTypeFRP, nil
	case "public-ip":
		return constants.ReverseProxyTypePublic, nil
	default:
		return constants.ReverseProxyTypeNone, nil
	}
}

func (o *UserOperator) GetLoginBackground(user *iamV1alpha2.User) (string, string) {
	b := o.GetUserAnnotation(user, constants.UserLoginBackground)
	s := o.GetUserAnnotation(user, constants.UserLoginBackgroundStyle)
	if b == "" {
		b = "/bg/0.jpg"
	}

	if s == "" {
		s = "fill"
	}

	return b, s
}

func (o *UserOperator) GetAvatar(user *iamV1alpha2.User) string {
	return o.GetUserAnnotation(user, constants.UserAvatar)
}

func (o *UserOperator) GetUserDID(user *iamV1alpha2.User) string {
	return o.GetUserAnnotation(user, constants.UserCertManagerDID)
}

func (o *UserOperator) GetUserAnnotation(user *iamV1alpha2.User, name string) string {
	if v, ok := user.Annotations[name]; ok {
		return v
	}
	return ""
}

func (o *UserOperator) GetUserZone(_ *iamV1alpha2.User) string {
	domain, err := o.GetDefaultDomain()
	if err != nil {
		klog.Errorf("get default domain: %v", err)
		return ""
	}
	return domain
}

func (o *UserOperator) GetUserDomainType(_ *iamV1alpha2.User) (bool, string, error) {
	// App IDs already identify their owner; all users share the machine domain.
	domain, err := o.GetDefaultDomain()
	return false, domain, err
}

func (o *UserOperator) GetDefaultDomain() (string, error) {
	domain, err := o.GetDomain()
	if err != nil {
		return "", err
	}
	domain = strings.TrimSuffix(domain, ".")
	if domain == "olares.local" {
		return "", nil
	}
	return domain, nil
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

	return terminus.Spec.Settings[users.SettingsDomainNameKey], nil

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
