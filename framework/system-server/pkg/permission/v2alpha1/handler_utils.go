package v2alpha1

import (
	"context"
	"fmt"

	providerv2alpha1 "bytetrade.io/web3os/system-server/pkg/providerregistry/v2alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func (h *handler) getProvider(providerName, providerDomain, providerAppNamespace string) []*rbacv1.ClusterRole {
	// we cannot assume the provider exists, so we need to mock  cluster roles for it
	var roles []*rbacv1.ClusterRole
	for _, ref := range []string{
		h.appProviderRef(providerName, providerAppNamespace),
		h.appDomainProviderRef(providerDomain),
	} {
		role := &rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: providerv2alpha1.GetRoleNameForRef(ref),
			},
			TypeMeta: metav1.TypeMeta{
				APIVersion: rbacv1.SchemeGroupVersion.String(),
				Kind:       "ClusterRole",
			},
		}
		roles = append(roles, role)
	}

	return roles
}

func (h *handler) bindingProvider(ctx context.Context, user, app, serviceAccount string, roles []*rbacv1.ClusterRole) error {
	appNamespace := fmt.Sprintf("%s-%s", app, user)
	for _, role := range roles {
		binding := &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: h.getProviderBindingName(appNamespace, serviceAccount, role.Name),
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      serviceAccount,
					Namespace: appNamespace,
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.SchemeGroupVersion.Group,
				Kind:     role.Kind,
				Name:     role.Name,
			},
		}

		if rb, err := h.kubeClient.RbacV1().ClusterRoleBindings().Get(ctx, binding.Name, metav1.GetOptions{}); err == nil {
			klog.Infof("Cluster role binding %s already exists for service account %s in app %s", rb.Name, serviceAccount, appNamespace)
			rb.Subjects = binding.Subjects
			rb.RoleRef = binding.RoleRef
			if _, err := h.kubeClient.RbacV1().ClusterRoleBindings().Update(ctx, rb, metav1.UpdateOptions{}); err != nil {
				klog.Errorf("Failed to update cluster role binding %s: %v", rb.Name, err)
				return err
			}
			continue
		}

		if _, err := h.kubeClient.RbacV1().ClusterRoles().Get(ctx, role.Name, metav1.GetOptions{}); err != nil {
			klog.Errorf("Cluster role %s does not exist for binding %s: %v", role.Name, binding.Name, err)
			// create a placeholder role to bind to
			placeholderRole := &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: role.Name,
				},
				Rules: []rbacv1.PolicyRule{
					{
						Verbs:           []string{"*"},
						NonResourceURLs: []string{"/placeholder-mock"},
					},
				},
			}
			if _, err := h.kubeClient.RbacV1().ClusterRoles().Create(ctx, placeholderRole, metav1.CreateOptions{}); err != nil {
				klog.Errorf("Failed to create placeholder cluster role %s: %v", role.Name, err)
				return err
			}
			klog.Infof("Created placeholder cluster role %s for binding %s", role.Name, binding.Name)
		}

		if _, err := h.kubeClient.RbacV1().ClusterRoleBindings().Create(ctx, binding, metav1.CreateOptions{}); err != nil {
			klog.Errorf("Failed to create cluster role binding %s: %v", binding.Name, err)
			return err
		}
		klog.Infof("Created cluster role binding %s for service account %s in app %s", binding.Name, serviceAccount, appNamespace)
	}
	return nil
}

func (h *handler) unbindingProvider(ctx context.Context, user, app, serviceAccount string, roles []*rbacv1.ClusterRole) error {
	appNamespace := fmt.Sprintf("%s-%s", app, user)
	for _, role := range roles {
		bindingName := h.getProviderBindingName(appNamespace, serviceAccount, role.Name)
		if err := h.kubeClient.RbacV1().ClusterRoleBindings().Delete(ctx, bindingName, metav1.DeleteOptions{}); err == nil {
			klog.Infof("Deleted cluster role binding %s for service account %s in app %s", bindingName, serviceAccount, appNamespace)
		} else {
			klog.Errorf("Failed to delete cluster role binding %s for service account %s in app %s: %v", bindingName, serviceAccount, appNamespace, err)
		}
	}

	return nil
}

func (h *handler) userProviderRef(user, providerName string) string {
	return fmt.Sprintf("user-system-%s/%s", user, providerName)
}

func (h *handler) appDomainProviderRef(domain string) string {
	return providerv2alpha1.ProviderRefFromHost(domain)
}

func (h *handler) appProviderRef(providerAppName, providerAppNamespace string) string {
	return fmt.Sprintf("%s/%s", providerAppNamespace, providerAppName)
}

func (h *handler) getProviderBindingName(appNamespace, serviceAccount, roleName string) string {
	return fmt.Sprintf("%s:%s:%s", appNamespace, serviceAccount, roleName)
}
