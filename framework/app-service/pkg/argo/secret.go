package argo

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
)

const (
	fakes3SecretName        = "argo-workflow-log-fakes3"
	workflowRoleName        = "workflow-role"
	workflowRoleBindingName = "workflow-rolebinding"
)

var (
	fakes3SecretTempl = `apiVersion: v1
kind: Secret
metadata:
  name: %s
  namespace: %s
type: Opaque
stringData:
  AWS_ACCESS_KEY_ID: S3RVER
  AWS_SECRET_ACCESS_KEY: S3RVER
`
	workflowRoleTempl = `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: %s
  namespace: %s
rules:
- apiGroups:
  - "*"
  resources:
  - pods
  verbs:
  - patch`

	workflowRolebindingTempl = `apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: %s
  namespace: %s
subjects:
  - kind: ServiceAccount
    namespace: %[2]s
    name: default
roleRef:
  kind: Role
  name: workflow-role
  apiGroup: rbac.authorization.k8s.io`
)

func createFakes3SecretAndRole(clientset *kubernetes.Clientset, namespace string) error {
	err := createFakes3Secret(clientset, namespace)
	if err != nil {
		klog.Errorf("Failed to createFakes3Secret namespace=%s err=%v", namespace, err)
		return err
	}

	err = createWorkflowRole(clientset, namespace)
	if err != nil {
		klog.Errorf("Failed to createWorkflowRole namespace=%s err=%v", namespace, err)
		return err
	}

	err = createWorkflowRoleBinding(clientset, namespace)
	if err != nil {
		klog.Errorf("Failed to createWorkflowRoleBinding namespace=%s err=%v", namespace, err)
		return err
	}

	return nil
}

func createFakes3Secret(clientset kubernetes.Interface, namespace string) error {
	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.Background(), fakes3SecretName, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	}

	secretManifest := fmt.Sprintf(fakes3SecretTempl, fakes3SecretName, namespace)
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(secretManifest), nil, nil)
	if err != nil {
		klog.Warningf("Failed to createFakes3Secret namespace=%s err=%v", namespace, err)
		return err
	}

	secret = obj.(*corev1.Secret)

	_, err = clientset.CoreV1().Secrets(namespace).Create(context.Background(), secret, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}

	return nil
}

func createWorkflowRole(clientset kubernetes.Interface, namespace string) error {
	rolesClient := clientset.RbacV1().Roles(namespace)
	role, err := rolesClient.Get(context.Background(), workflowRoleName, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	}

	roleManifest := fmt.Sprintf(workflowRoleTempl, workflowRoleName, namespace)
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(roleManifest), nil, nil)
	if err != nil {
		return err
	}

	role = obj.(*rbacv1.Role)

	_, err = rolesClient.Create(context.Background(), role, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}

	return nil
}

func createWorkflowRoleBinding(clientset kubernetes.Interface, namespace string) error {
	roleBindingsClient := clientset.RbacV1().RoleBindings(namespace)
	rolebinding, err := roleBindingsClient.Get(context.Background(), workflowRoleBindingName, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	}

	roleBindingManifest := fmt.Sprintf(workflowRolebindingTempl, workflowRoleBindingName, namespace)
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(roleBindingManifest), nil, nil)
	if err != nil {
		return err
	}

	rolebinding = obj.(*rbacv1.RoleBinding)

	_, err = roleBindingsClient.Create(context.Background(), rolebinding, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}

	return nil
}
