package templates

import (
	"fmt"

	"bytetrade.io/web3os/app-service/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/*
apiVersion: v1
kind: Namespace
metadata:
  annotations:
    kubesphere.io/creator: {{ .Values.user.name }}
  finalizers:
  - finalizers.kubesphere.io/namespaces
  labels:
    kubernetes.io/metadata.name: {{ .Values.namespace.name }}
    kubesphere.io/namespace: {{ .Values.namespace.name }}
    kubesphere.io/workspace: system-workspace
  name: {{ .Values.namespace.name }}
*/

type Userspace corev1.Namespace

func NewUserspace(user string) *Userspace {
	name := utils.UserspaceName(user)

	return &Userspace{
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

func NewUserSystem(user string) *Userspace {
	name := "user-system-" + user

	return &Userspace{
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

/*
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    iam.kubesphere.io/user-ref: {{ .Values.user.name }}
  name: {{ .Values.user.name }}-{{ .Values.namespace.role }}
  namespace: {{ .Values.namespace.name }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: {{ .Values.namespace.role }}
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: {{ .Values.user.name }}
*/

// UserspaceRoleBinding is an alias for rbacv1.RoleBinding.
type UserspaceRoleBinding rbacv1.RoleBinding

// NewUserspaceRoleBinding creates a new userspace rolebinding.
func NewUserspaceRoleBinding(user, userspace, role string) *UserspaceRoleBinding {
	return &UserspaceRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", user, role),
			Namespace: userspace,
			Labels: map[string]string{
				"iam.kubesphere.io/user-ref": user,
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
				Name:     user,
			},
		},
	}
}
