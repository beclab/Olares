package permission

import (
	"context"
	"crypto"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"time"

	sysv1alpha1 "bytetrade.io/web3os/system-server/pkg/apis/sys/v1alpha1"
	"bytetrade.io/web3os/system-server/pkg/constants"
	clientset "bytetrade.io/web3os/system-server/pkg/generated/clientset/versioned"
	v1alpha1 "bytetrade.io/web3os/system-server/pkg/generated/listers/sys/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	MAX_RAND_INT = 1000000
)

type PermissionControl struct {
	permissionLister    v1alpha1.ApplicationPermissionLister
	permissionClientset clientset.Interface
	rand                *rand.Rand
}

func NewPermissionControl(clientset clientset.Interface, lister v1alpha1.ApplicationPermissionLister) *PermissionControl {

	return &PermissionControl{
		permissionLister:    lister,
		permissionClientset: clientset,
		rand:                rand.New(rand.NewSource(time.Now().Unix())),
	}
}

func (p *PermissionControl) getAppPermissionFromAppKey(_ context.Context, appkey string) (*sysv1alpha1.ApplicationPermission, error) {
	aps, err := p.permissionLister.ApplicationPermissions(constants.MyNamespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	if len(aps) == 0 {
		return nil, errors.New("none of app permissions")
	}

	for _, ap := range aps {
		if ap.Spec.Key == appkey {
			return ap, nil
		}
	}

	return nil, errors.New("app not found")
}

func (p *PermissionControl) verifyPermission(appPerm *sysv1alpha1.ApplicationPermission,
	reqPerm *sysv1alpha1.PermissionRequire) bool {
	for _, p := range appPerm.Spec.Permission {
		if p.Include(reqPerm, false) {
			return true
		}
	}

	return false
}

func (p *PermissionControl) applyPermission(ctx context.Context, permReg *PermissionRegister) (*RegisterResp, error) {
	oldAP, err := p.permissionLister.ApplicationPermissions(constants.MyNamespace).
		Get(permReg.App)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}

	perms := make([]sysv1alpha1.PermissionRequire, 0)
	for _, p := range permReg.Perm {
		perms = append(perms, *p.DeepCopy())
	}

	appPerm := sysv1alpha1.ApplicationPermission{
		ObjectMeta: metav1.ObjectMeta{
			Name: permReg.App,
		},

		Spec: sysv1alpha1.ApplicationPermissionSpec{
			App:        permReg.App,
			Appid:      permReg.AppID,
			Key:        "",
			Secret:     "",
			Permission: perms,
		},

		Status: sysv1alpha1.ApplicationPermissionStatus{
			State: sysv1alpha1.Active,
		},
	}

	if apierrors.IsNotFound(err) {
		k, s := p.genAppkeyAndSecret(permReg.App)
		appPerm.Spec.Key = k
		appPerm.Spec.Secret = s
		if _, err = p.permissionClientset.SysV1alpha1().
			ApplicationPermissions(constants.MyNamespace).
			Create(ctx, &appPerm, metav1.CreateOptions{}); err != nil {
			return nil, err
		}
	} else {
		appPerm.Spec.Key = oldAP.Spec.Key
		appPerm.Spec.Secret = oldAP.Spec.Secret
		oldAP.Spec.Permission = appPerm.Spec.Permission
		if _, err = p.permissionClientset.SysV1alpha1().
			ApplicationPermissions(constants.MyNamespace).
			Update(ctx, oldAP, metav1.UpdateOptions{}); err != nil {
			return nil, err
		}
	}

	return &RegisterResp{
		AppKey:    appPerm.Spec.Key,
		AppSecret: appPerm.Spec.Secret,
	}, nil
}

func (p *PermissionControl) deletePermission(ctx context.Context, name string) error {
	_, err := p.permissionLister.ApplicationPermissions(constants.MyNamespace).
		Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}

		return err
	}

	return p.permissionClientset.SysV1alpha1().ApplicationPermissions(constants.MyNamespace).
		Delete(ctx, name, metav1.DeleteOptions{})
}

func (p *PermissionControl) genAppkeyAndSecret(app string) (string, string) {
	key := fmt.Sprintf("bytetrade_%s_%d", app, p.rand.Intn(MAX_RAND_INT))
	secret := md5(fmt.Sprintf("%s|%d", key, time.Now().Unix()))

	return key, secret[:16]
}

func md5(str string) string {
	h := crypto.MD5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
