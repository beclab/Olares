/*
Copyright 2020 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package am

import (
	"context"
	"fmt"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/globalrole"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/globalrolebinding"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog"

	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/informers"
	resourcev1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/clusterrole"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/clusterrolebinding"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/role"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/rolebinding"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
)

type AccessManagementInterface interface {
	GetClusterRoleOfUser(username string) (*rbacv1.ClusterRole, error)
	GetNamespaceRoleOfUser(username string, groups []string, namespace string) ([]*rbacv1.Role, error)
	ListRoles(namespace string, query *query.Query) (*api.ListResult, error)
	ListClusterRoles(query *query.Query) (*api.ListResult, error)
	ListClusterRoleBindings(username string) ([]*rbacv1.ClusterRoleBinding, error)
	ListRoleBindings(username string, groups []string, namespace string) ([]*rbacv1.RoleBinding, error)
	GetRoleReferenceRules(roleRef rbacv1.RoleRef, namespace string) (string, []rbacv1.PolicyRule, error)
	GetClusterRole(name string) (*rbacv1.ClusterRole, error)
	GetNamespaceRole(namespace string, name string) (*rbacv1.Role, error)
	RemoveUserFromNamespace(username string, namespace string) error
	RemoveUserFromCluster(username string) error

	GetGlobalRoleOfUser(username string) (*iamv1alpha2.GlobalRole, error)
	ListGlobalRoles(query *query.Query) (*api.ListResult, error)
	ListGlobalRoleBindings(username string) ([]*iamv1alpha2.GlobalRoleBinding, error)
	GetGlobalRole(globalRole string) (*iamv1alpha2.GlobalRole, error)
}

type amOperator struct {
	workspaceRoleBindingGetter resourcev1alpha3.Interface
	clusterRoleBindingGetter   resourcev1alpha3.Interface
	roleBindingGetter          resourcev1alpha3.Interface
	clusterRoleGetter          resourcev1alpha3.Interface
	roleGetter                 resourcev1alpha3.Interface
	namespaceLister            listersv1.NamespaceLister
	ksclient                   kubesphere.Interface
	k8sclient                  kubernetes.Interface

	globalRoleBindingGetter resourcev1alpha3.Interface
	globalRoleGetter        resourcev1alpha3.Interface
}

func NewReadOnlyOperator(factory informers.InformerFactory) AccessManagementInterface {
	operator := &amOperator{
		clusterRoleBindingGetter: clusterrolebinding.New(factory.KubernetesSharedInformerFactory()),
		roleBindingGetter:        rolebinding.New(factory.KubernetesSharedInformerFactory()),
		clusterRoleGetter:        clusterrole.New(factory.KubernetesSharedInformerFactory()),
		roleGetter:               role.New(factory.KubernetesSharedInformerFactory()),
		namespaceLister:          factory.KubernetesSharedInformerFactory().Core().V1().Namespaces().Lister(),
		globalRoleBindingGetter:  globalrolebinding.New(factory.KubeSphereSharedInformerFactory()),
		globalRoleGetter:         globalrole.New(factory.KubeSphereSharedInformerFactory()),
	}
	// no more CRDs of devopsprojects if the DevOps module was disabled
	return operator
}

func NewOperator(ksClient kubesphere.Interface, k8sClient kubernetes.Interface, factory informers.InformerFactory) AccessManagementInterface {
	amOperator := NewReadOnlyOperator(factory).(*amOperator)
	amOperator.ksclient = ksClient
	amOperator.k8sclient = k8sClient
	return amOperator
}

func (am *amOperator) GetGlobalRoleOfUser(username string) (*iamv1alpha2.GlobalRole, error) {
	globalRoleBindings, err := am.ListGlobalRoleBindings(username)
	if len(globalRoleBindings) > 0 {
		// Usually, only one globalRoleBinding will be found which is created from ks-console.
		if len(globalRoleBindings) > 1 {
			klog.Warningf("conflict global role binding, username: %s", username)
		}
		globalRole, err := am.GetGlobalRole(globalRoleBindings[0].RoleRef.Name)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		return globalRole, nil
	}

	err = errors.NewNotFound(iamv1alpha2.Resource(iamv1alpha2.ResourcesSingularGlobalRoleBinding), username)
	klog.V(4).Info(err)
	return nil, err
}

func (am *amOperator) ListGlobalRoleBindings(username string) ([]*iamv1alpha2.GlobalRoleBinding, error) {
	roleBindings, err := am.globalRoleBindingGetter.List("", query.New())
	if err != nil {
		return nil, err
	}

	result := make([]*iamv1alpha2.GlobalRoleBinding, 0)
	for _, obj := range roleBindings.Items {
		roleBinding := obj.(*iamv1alpha2.GlobalRoleBinding)
		if contains(roleBinding.Subjects, username, nil) {
			result = append(result, roleBinding)
		}
	}

	return result, nil
}

func (am *amOperator) ListGlobalRoles(query *query.Query) (*api.ListResult, error) {
	return am.globalRoleGetter.List("", query)
}

func (am *amOperator) GetGlobalRole(globalRole string) (*iamv1alpha2.GlobalRole, error) {
	obj, err := am.globalRoleGetter.Get("", globalRole)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return obj.(*iamv1alpha2.GlobalRole), nil
}

func (am *amOperator) GetNamespaceRoleOfUser(username string, groups []string, namespace string) ([]*rbacv1.Role, error) {

	userRoleBindings, err := am.ListRoleBindings(username, groups, namespace)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if len(userRoleBindings) > 0 {
		roles := make([]*rbacv1.Role, len(userRoleBindings))
		for i, roleBinding := range userRoleBindings {
			role, err := am.GetNamespaceRole(namespace, roleBinding.RoleRef.Name)
			if err != nil {
				klog.Error(err)
				return nil, err
			}

			out := role.DeepCopy()
			if out.Annotations == nil {
				out.Annotations = make(map[string]string, 0)
			}
			out.Annotations[iamv1alpha2.RoleAnnotation] = role.Name

			roles[i] = out
		}

		if len(userRoleBindings) > 1 {
			klog.Infof("conflict role binding, username: %s", username)
		}
		return roles, nil
	}

	err = errors.NewNotFound(iamv1alpha2.Resource(iamv1alpha2.ResourcesSingularRoleBinding), username)
	klog.V(4).Info(err)
	return nil, err
}

func (am *amOperator) GetClusterRoleOfUser(username string) (*rbacv1.ClusterRole, error) {
	userRoleBindings, err := am.ListClusterRoleBindings(username)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if len(userRoleBindings) > 0 {
		role, err := am.GetClusterRole(userRoleBindings[0].RoleRef.Name)
		if err != nil {
			klog.Error(err)
			return nil, err
		}

		if len(userRoleBindings) > 1 {
			klog.Warningf("conflict cluster role binding, username: %s", username)
		}

		out := role.DeepCopy()
		if out.Annotations == nil {
			out.Annotations = make(map[string]string, 0)
		}
		out.Annotations[iamv1alpha2.ClusterRoleAnnotation] = role.Name
		return out, nil
	}

	err = errors.NewNotFound(iamv1alpha2.Resource(iamv1alpha2.ResourcesSingularClusterRoleBinding), username)
	klog.V(4).Info(err)
	return nil, err
}

//func (am *amOperator) ListWorkspaceRoleBindings(username string, groups []string, workspace string) ([]*iamv1alpha2.WorkspaceRoleBinding, error) {
//	roleBindings, err := am.workspaceRoleBindingGetter.List("", query.New())
//
//	if err != nil {
//		return nil, err
//	}
//
//	result := make([]*iamv1alpha2.WorkspaceRoleBinding, 0)
//
//	for _, obj := range roleBindings.Items {
//		roleBinding := obj.(*iamv1alpha2.WorkspaceRoleBinding)
//		inSpecifiedWorkspace := workspace == "" || roleBinding.Labels[tenantv1alpha1.WorkspaceLabel] == workspace
//		if contains(roleBinding.Subjects, username, groups) && inSpecifiedWorkspace {
//			result = append(result, roleBinding)
//		}
//	}
//
//	return result, nil
//}

func (am *amOperator) ListClusterRoleBindings(username string) ([]*rbacv1.ClusterRoleBinding, error) {

	roleBindings, err := am.clusterRoleBindingGetter.List("", query.New())
	if err != nil {
		return nil, err
	}

	result := make([]*rbacv1.ClusterRoleBinding, 0)
	for _, obj := range roleBindings.Items {
		roleBinding := obj.(*rbacv1.ClusterRoleBinding)
		if contains(roleBinding.Subjects, username, nil) {
			result = append(result, roleBinding)
		}
	}

	return result, nil
}

func (am *amOperator) ListRoleBindings(username string, groups []string, namespace string) ([]*rbacv1.RoleBinding, error) {
	roleBindings, err := am.roleBindingGetter.List(namespace, query.New())
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	result := make([]*rbacv1.RoleBinding, 0)
	for _, obj := range roleBindings.Items {
		roleBinding := obj.(*rbacv1.RoleBinding)
		if contains(roleBinding.Subjects, username, groups) {
			result = append(result, roleBinding)
		}
	}
	return result, nil
}

func contains(subjects []rbacv1.Subject, username string, groups []string) bool {
	// if username is nil means list all role bindings
	if username == "" {
		return true
	}
	for _, subject := range subjects {
		if subject.Kind == rbacv1.UserKind && subject.Name == username {
			return true
		}
		if subject.Kind == rbacv1.GroupKind && sliceutil.HasString(groups, subject.Name) {
			return true
		}
	}
	return false
}

func (am *amOperator) ListRoles(namespace string, query *query.Query) (*api.ListResult, error) {
	return am.roleGetter.List(namespace, query)
}

func (am *amOperator) ListClusterRoles(query *query.Query) (*api.ListResult, error) {
	return am.clusterRoleGetter.List("", query)
}

func (am *amOperator) RemoveUserFromNamespace(username string, namespace string) error {

	roleBindings, err := am.ListRoleBindings(username, nil, namespace)
	if err != nil {
		klog.Error(err)
		return err
	}

	for _, roleBinding := range roleBindings {
		err := am.k8sclient.RbacV1().RoleBindings(namespace).Delete(context.Background(), roleBinding.Name, *metav1.NewDeleteOptions(0))
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			klog.Error(err)
			return err
		}
	}

	return nil
}

func (am *amOperator) RemoveUserFromCluster(username string) error {
	roleBindings, err := am.ListClusterRoleBindings(username)
	if err != nil {
		klog.Error(err)
		return err
	}

	for _, roleBinding := range roleBindings {
		err := am.k8sclient.RbacV1().ClusterRoleBindings().Delete(context.Background(), roleBinding.Name, *metav1.NewDeleteOptions(0))
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			klog.Error(err)
			return err
		}
	}

	return nil
}

// GetRoleReferenceRules attempts to resolve the RoleBinding or ClusterRoleBinding.
func (am *amOperator) GetRoleReferenceRules(roleRef rbacv1.RoleRef, namespace string) (regoPolicy string, rules []rbacv1.PolicyRule, err error) {

	empty := make([]rbacv1.PolicyRule, 0)

	switch roleRef.Kind {
	case iamv1alpha2.ResourceKindRole:
		role, err := am.GetNamespaceRole(namespace, roleRef.Name)
		if err != nil {
			if errors.IsNotFound(err) {
				return "", empty, nil
			}
			return "", nil, err
		}
		return role.Annotations[iamv1alpha2.RegoOverrideAnnotation], role.Rules, nil
	case iamv1alpha2.ResourceKindClusterRole:
		clusterRole, err := am.GetClusterRole(roleRef.Name)
		if err != nil {
			if errors.IsNotFound(err) {
				return "", empty, nil
			}
			return "", nil, err
		}
		return clusterRole.Annotations[iamv1alpha2.RegoOverrideAnnotation], clusterRole.Rules, nil
	case iamv1alpha2.ResourceKindGlobalRole:
		globalRole, err := am.GetGlobalRole(roleRef.Name)
		if err != nil {
			if errors.IsNotFound(err) {
				return "", empty, nil
			}
			return "", nil, err
		}
		return globalRole.Annotations[iamv1alpha2.RegoOverrideAnnotation], globalRole.Rules, nil
	default:
		return "", nil, fmt.Errorf("unsupported role reference kind: %q", roleRef.Kind)
	}
}

func (am *amOperator) GetNamespaceRole(namespace string, name string) (*rbacv1.Role, error) {
	obj, err := am.roleGetter.Get(namespace, name)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return obj.(*rbacv1.Role), nil
}

func (am *amOperator) GetClusterRole(name string) (*rbacv1.ClusterRole, error) {
	obj, err := am.clusterRoleGetter.Get("", name)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return obj.(*rbacv1.ClusterRole), nil
}
