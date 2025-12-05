package users

import (
	"fmt"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type UserCreateOption struct {
	Name         string
	OwnerRole    string
	DisplayName  string
	Email        string
	Password     string
	Description  string
	TerminusName string
	MemoryLimit  string
	CpuLimit     string
}

var (
	AnnotationGroup       = "bytetrade.io"
	AnnotationUserCreator = "bytetrade.io/creator"
	AnnotationUserDeleter = "bytetrade.io/deleter"

	UserAnnotationTerminusNameKey = fmt.Sprintf("%s/terminus-name", AnnotationGroup)
	UserAvatar                    = fmt.Sprintf("%s/avatar", AnnotationGroup)

	UserAnnotationZoneKey = fmt.Sprintf("%s/zone", AnnotationGroup)

	UserAnnotationUninitialized = fmt.Sprintf("%s/uninitialized", AnnotationGroup)

	UserAnnotationOwnerRole = fmt.Sprintf("%s/owner-role", AnnotationGroup)

	UserAnnotationIsEphemeral = fmt.Sprintf("%s/is-ephemeral", AnnotationGroup)

	EnableSSLTaskResultAnnotationKey = fmt.Sprintf("%s/task-enable-ssl", AnnotationGroup)
	UserLauncherAuthPolicy           = fmt.Sprintf("%s/launcher-auth-policy", AnnotationGroup)
	UserLauncherAccessLevel          = fmt.Sprintf("%s/launcher-access-level", AnnotationGroup)
	UserAnnotationLimitsCpuKey       = "bytetrade.io/user-cpu-limit"
	UserAnnotationLimitsMemoryKey    = "bytetrade.io/user-memory-limit"
	UserAnnotationSyncToLldapKey     = "iam.kubesphere.io/sync-to-lldap"
	UserAnnotationSyncedToLldapKeyy  = "iam.kubesphere.io/synced-to-lldap"
)

func NewUserAndGlobalRoleBinding(u *UserCreateOption) (*iamv1alpha2.User, *iamv1alpha2.GlobalRoleBinding) {
	user := iamv1alpha2.User{
		TypeMeta: metav1.TypeMeta{
			APIVersion: iamv1alpha2.SchemeGroupVersion.String(),
			Kind:       iamv1alpha2.ResourceKindUser,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: u.Name,
			Annotations: map[string]string{
				"iam.kubesphere.io/uninitialized": "true",
				// TODO:hys
				AnnotationUserCreator:               "cli",
				UserAnnotationUninitialized:         "true",
				UserAnnotationOwnerRole:             u.OwnerRole,
				UserAnnotationIsEphemeral:           "true", // add is-ephemeral flag, 'true' it means need temporary domain
				UserAnnotationTerminusNameKey:       u.TerminusName,
				UserLauncherAuthPolicy:              "two_factor",
				UserLauncherAccessLevel:             "1",
				UserAnnotationLimitsMemoryKey:       u.MemoryLimit,
				UserAnnotationLimitsCpuKey:          u.CpuLimit,
				"iam.kubesphere.io/sync-to-lldap":   "true",
				"iam.kubesphere.io/synced-to-lldap": "false",
			},
		},
		Spec: iamv1alpha2.UserSpec{
			DisplayName: u.DisplayName,
			Email:       u.Email,
			//EncryptedPassword: u.Password,
			InitialPassword: u.Password,
			Description:     u.Description,
		},
		Status: iamv1alpha2.UserStatus{
			State: iamv1alpha2.UserActive,
		},
	}

	globalRoleBinding := iamv1alpha2.GlobalRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: iamv1alpha2.SchemeGroupVersion.String(),
			Kind:       iamv1alpha2.ResourceKindGlobalRoleBinding,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: u.Name,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: iamv1alpha2.SchemeGroupVersion.String(),
			Kind:     iamv1alpha2.ResourceKindGlobalRole,
			Name:     u.OwnerRole,
		},
		Subjects: []rbacv1.Subject{
			{
				APIGroup: iamv1alpha2.SchemeGroupVersion.String(),
				Kind:     iamv1alpha2.ResourceKindUser,
				Name:     u.Name,
			},
		},
	}

	return &user, &globalRoleBinding

}
