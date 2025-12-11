package users

import (
	"context"
	"fmt"

	v1alpha1client "bytetrade.io/web3os/bfl/pkg/client/clientset/v1alpha1"
	"bytetrade.io/web3os/bfl/pkg/constants"
	"github.com/beclab/api/iam/v1alpha2"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type IamUser struct {
	ctx context.Context

	kc v1alpha1client.ClientInterface
}

func NewIamUser() (*IamUser, error) {
	kc, err := v1alpha1client.NewKubeClient(nil)
	if err != nil {
		return nil, err
	}
	return &IamUser{
		ctx: context.Background(),
		kc:  kc,
	}, nil
}

func (u *IamUser) GetUser() (*v1alpha2.User, error) {
	var user v1alpha2.User
	err := u.kc.CtrlClient().Get(u.ctx, types.NamespacedName{Name: constants.Username}, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *IamUser) ListUsers() ([]v1alpha2.User, error) {
	var users v1alpha2.UserList
	err := u.kc.CtrlClient().List(u.ctx, &users)
	if err != nil {
		return nil, err
	}
	return users.Items, nil
}

func (u *IamUser) UpdateUser(user *v1alpha2.User) error {
	err := u.kc.CtrlClient().Update(u.ctx, user)
	return err
}

func (u *IamUser) GetTerminusName() (string, error) {
	user, err := u.GetUser()
	if err != nil {
		return "", err
	}
	if name, ok := user.Annotations[constants.UserAnnotationTerminusNameKey]; ok && name != "" {
		return name, nil
	}
	return "", fmt.Errorf("user olares name not binding")
}

func (u *IamUser) BindingTerminusName(terminusName string) error {
	user, err := u.GetUser()
	if err != nil {
		return err
	}
	if v, ok := user.Annotations[constants.UserAnnotationTerminusNameKey]; ok {
		return fmt.Errorf("user '%s' olares name is already bind, olares name: %s", user.Name, v)
	}

	// update terminus to user annotation
	user.Annotations[constants.UserAnnotationTerminusNameKey] = terminusName
	return u.UpdateUser(user)
}

func (u *IamUser) UnbindingTerminusName() error {
	user, err := u.GetUser()
	if err != nil {
		return err
	}
	delete(user.Annotations, constants.UserAnnotationTerminusNameKey)

	return u.UpdateUser(user)
}

func (u *IamUser) UpdateUserAnnotation(key, value string) error {
RETRY:
	user, err := u.GetUser()
	if err != nil {
		return err
	}
	user.Annotations[key] = value
	if err = u.UpdateUser(user); err != nil && apierrors.IsConflict(err) {
		goto RETRY
	} else if err != nil {
		return err
	}
	return nil
}

func (u *IamUser) GetAnnotation(name string) string {
	user, err := u.GetUser()
	if err != nil {
		return ""
	}
	if v, ok := user.Annotations[name]; ok {
		return v
	}
	return ""
}
