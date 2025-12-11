package templates

import (
	"fmt"

	"bytetrade.io/web3os/bfl/pkg/constants"

	iamV1alpha2 "github.com/beclab/api/iam/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	applyCorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	applyMetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/utils/pointer"
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

func NewUserAndGlobalRoleBinding(u *UserCreateOption) (string, *iamV1alpha2.User, string, *iamV1alpha2.GlobalRoleBinding) {
	user := iamV1alpha2.User{
		TypeMeta: metav1.TypeMeta{
			APIVersion: iamV1alpha2.SchemeGroupVersion.String(),
			Kind:       iamV1alpha2.ResourceKindUser,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: u.Name,
			Annotations: map[string]string{
				"iam.kubesphere.io/uninitialized":       "true",
				constants.AnnotationUserCreator:         constants.Username,
				constants.UserAnnotationUninitialized:   "true",
				constants.UserAnnotationOwnerRole:       u.OwnerRole,
				constants.UserAnnotationIsEphemeral:     "true", // add is-ephemeral flag, 'true' it means need temporary domain
				constants.UserAnnotationTerminusNameKey: u.TerminusName,
				constants.UserLauncherAuthPolicy:        "two_factor",
				constants.UserLauncherAccessLevel:       "1",
				constants.UserAnnotationLimitsMemoryKey: u.MemoryLimit,
				constants.UserAnnotationLimitsCpuKey:    u.CpuLimit,
				"iam.kubesphere.io/sync-to-lldap":       "true",
				"iam.kubesphere.io/synced-to-lldap":     "false",
			},
		},
		Spec: iamV1alpha2.UserSpec{
			DisplayName: u.DisplayName,
			Email:       u.Email,
			//EncryptedPassword: u.Password,
			InitialPassword: u.Password,
			Description:     u.Description,
		},
		Status: iamV1alpha2.UserStatus{
			State: iamV1alpha2.UserActive,
		},
	}

	globalRoleBinding := iamV1alpha2.GlobalRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: iamV1alpha2.SchemeGroupVersion.String(),
			Kind:       iamV1alpha2.ResourceKindGlobalRoleBinding,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: u.Name,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: iamV1alpha2.SchemeGroupVersion.String(),
			Kind:     iamV1alpha2.ResourceKindGlobalRole,
			Name:     u.OwnerRole,
		},
		Subjects: []rbacv1.Subject{
			{
				APIGroup: iamV1alpha2.SchemeGroupVersion.String(),
				Kind:     iamV1alpha2.ResourceKindUser,
				Name:     u.Name,
			},
		},
	}

	return user.Name, &user, globalRoleBinding.Name, &globalRoleBinding

}

//func NewWorkspaceRoleBinding(u *UserCreateOption, workspace, workspaceRole string) (string, *iamV1alpha2.WorkspaceRoleBinding) {
//	name := u.Name
//
//	return name, &iamV1alpha2.WorkspaceRoleBinding{
//		TypeMeta: metav1.TypeMeta{
//			APIVersion: iamV1alpha2.SchemeGroupVersion.String(),
//			Kind:       iamV1alpha2.ResourceKindWorkspaceRoleBinding,
//		},
//		ObjectMeta: metav1.ObjectMeta{
//			Name: name,
//			Labels: map[string]string{
//				"iam.kubesphere.io/user-ref": name,
//				"kubesphere.io/workspace":    workspace,
//			},
//		},
//		RoleRef: rbacv1.RoleRef{
//			APIGroup: iamV1alpha2.SchemeGroupVersion.String(),
//			Kind:     iamV1alpha2.ResourceKindWorkspaceRole,
//			Name:     workspaceRole,
//		},
//		Subjects: []rbacv1.Subject{
//			{
//				APIGroup: iamV1alpha2.SchemeGroupVersion.String(),
//				Kind:     iamV1alpha2.ResourceKindUser,
//				Name:     name,
//			},
//		},
//	}
//}

type Userspace = corev1.Namespace

func NewUserspace(user string) (string, *Userspace) {
	name := fmt.Sprintf(constants.UserspaceNameFormat, user)

	return name, &Userspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,

			Annotations: map[string]string{
				"kubesphere.io/creator": user,
			},

			Labels: map[string]string{
				"kubernetes.io/metadata.name": name,
				"kubesphere.io/namespace":     name,
				"kubesphere.io/workspace":     "system-workspace",
			},

			Finalizers: []string{
				"finalizers.kubesphere.io/namespaces",
			},
		},
	}
}

type UserspaceRoleBinding = rbacv1.RoleBinding

func NewUserspaceRoleBinding(username, userspace, role string) *UserspaceRoleBinding {
	return &UserspaceRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", username, role),
			Namespace: userspace,
			Labels: map[string]string{
				"iam.kubesphere.io/user-ref": username,
			},
		},

		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     role,
		},

		Subjects: []rbacv1.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "User",
				Name:     username,
			},
		},
	}
}

func NewApplyConfigmap(namespace string, data map[string]string) *applyCorev1.ConfigMapApplyConfiguration {
	return &applyCorev1.ConfigMapApplyConfiguration{
		TypeMetaApplyConfiguration: applyMetav1.TypeMetaApplyConfiguration{
			Kind:       pointer.String("ConfigMap"),
			APIVersion: pointer.String(corev1.SchemeGroupVersion.String()),
		},
		ObjectMetaApplyConfiguration: &applyMetav1.ObjectMetaApplyConfiguration{
			Name:      pointer.String(constants.NameSSLConfigMapName),
			Namespace: pointer.String(namespace),
		},
		Data: data,
	}
}
