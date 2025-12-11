package controllers

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"bytetrade.io/web3os/bfl/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *NginxController) reconcileFileserverProvider(ctx context.Context) error {
	pods, err := r.getFilesPods(ctx)
	if err != nil {
		klog.Errorf("failed to list pods, %v", err)
		return err
	}

	var nodes corev1.NodeList
	err = r.List(ctx, &nodes)
	if err != nil {
		klog.Errorf("failed to list nodes, %v", err)
		return err
	}

	_, podMap := r.getFileserverPodMap(nodes, pods)

	var serviceList []string
	serviceNamespace := (&ProxyServiceConfig{}).ServiceNameSpace()
	for nodeName, pod := range podMap {
		// create or update proxy service
		config := &ProxyServiceConfig{nodeName}
		err = config.UpsertProxyService(ctx, r.Client)
		if err != nil {
			klog.Errorf("failed to upsert proxy service for %s, %v", nodeName, err)
			return err
		}

		// create or update fileserver provider
		providerConfig := &FileServerProviderConfig{
			NodeName:    nodeName,
			Fileserver:  pod,
			ProxyServer: config,
		}
		err = providerConfig.UpsertFileServerRole(ctx, r.Client)
		if err != nil {
			klog.Errorf("failed to upsert fileserver cluster role for %s, %v", nodeName, err)
			return err
		}

		err = providerConfig.UpsertFileServerRoleBinding(ctx, r.Client)
		if err != nil {
			klog.Errorf("failed to upsert fileserver cluster role binding for %s, %v", nodeName, err)
			return err
		}

		serviceList = append(serviceList, config.ServiceName())
	} // end of current nodes loop

	// try to find file proxy of not existing nodes to delete
	var currentServiceList corev1.ServiceList
	err = r.Client.List(ctx, &currentServiceList, client.InNamespace(serviceNamespace))
	if err != nil {
		klog.Errorf("failed to list services in namespace %s, %v", serviceNamespace)
		return nil
	}
	for _, service := range currentServiceList.Items {
		if strings.HasPrefix(service.Name, "files-") &&
			FilesNodeService(service).IsFilesNodeService() &&
			!slices.Contains(serviceList, service.Name) {
			klog.Info("Found orphaned file node service: %s", service.Name)
			err = r.Client.Delete(ctx, &service)
			if err != nil {
				klog.Errorf("failed to delete orphaned service %s, %v", service.Name, err)
			}
		}
	}
	return nil
}

type ProxyServiceConfig struct {
	NodeName string
}

func (p *ProxyServiceConfig) ServiceName() string {
	return fmt.Sprintf("files-%s", p.NodeName)
}

func (p *ProxyServiceConfig) ServiceNameSpace() string {
	return fmt.Sprintf("user-system-%s", constants.Username)
}

func (p *ProxyServiceConfig) ServiceHost() string {
	return fmt.Sprintf("http://%s.%s.svc.cluster.local:28080", p.ServiceName(), p.ServiceNameSpace())
}

func (p *ProxyServiceConfig) Selector() map[string]string {
	return map[string]string{"app": "systemserver"}
}

func (p *ProxyServiceConfig) Ports() []corev1.ServicePort {
	return []corev1.ServicePort{
		{
			Name:       "rbac-proxy",
			Port:       28080,
			TargetPort: intstr.FromInt(28080),
		},
	}
}

func (p *ProxyServiceConfig) UpsertProxyService(ctx context.Context, c client.Client) error {
	var proxyService corev1.Service

	key := types.NamespacedName{
		Namespace: p.ServiceNameSpace(),
		Name:      p.ServiceName(),
	}
	err := c.Get(ctx, key, &proxyService)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get proxy service %s: %w", p.ServiceName(), err)
		}

		// create the service if it does not exist
		proxyService = corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      p.ServiceName(),
				Namespace: p.ServiceNameSpace(),
			},
			Spec: corev1.ServiceSpec{
				Selector: p.Selector(),
				Type:     corev1.ServiceTypeClusterIP,
				Ports:    p.Ports(),
			},
		}

		FilesNodeService(proxyService).Wrap()
		return c.Create(ctx, &proxyService)
	}

	// do not need to update
	return nil
}

func (p *ProxyServiceConfig) DeleteProxyService(ctx context.Context, c client.Client) error {
	var proxyService corev1.Service

	key := types.NamespacedName{
		Namespace: p.ServiceNameSpace(),
		Name:      p.ServiceName(),
	}
	err := c.Get(ctx, key, &proxyService)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Service not found, nothing to delete
			return nil
		}
		return fmt.Errorf("failed to get proxy service %s: %w", p.ServiceName(), err)
	}

	return c.Delete(ctx, &proxyService)
}

type FilesNodeService corev1.Service

const fileServerNodeServiceAnnotation = "olare/files-node-server"

func (f FilesNodeService) Wrap() {
	f.Annotations = map[string]string{
		fileServerNodeServiceAnnotation: "true",
	}
}

func (f FilesNodeService) IsFilesNodeService() bool {
	return f.Annotations[fileServerNodeServiceAnnotation] == "true"
}

type FileServerProviderConfig struct {
	NodeName    string
	Fileserver  *corev1.Pod
	ProxyServer *ProxyServiceConfig
}

func (f *FileServerProviderConfig) DomainClusterRoleName() string {
	return fmt.Sprintf("%s:files-frontend-domain-%s", constants.Username, f.NodeName)
}

func (f *FileServerProviderConfig) DomainClusterRoleBindingName() string {
	return fmt.Sprintf("user:%s:files-frontend-domain-%s", constants.Username, f.NodeName)
}

func (f *FileServerProviderConfig) ProviderRegistryRef() string {
	return fmt.Sprintf("%s/%s", f.ProxyServer.ServiceNameSpace(), f.ProxyServer.ServiceName())
}

func (f *FileServerProviderConfig) ProviderServiceRef() string {
	return fmt.Sprintf("%s:80", f.Fileserver.Status.PodIP)
}

func (f *FileServerProviderConfig) UpsertFileServerRole(ctx context.Context, c client.Client) error {
	var clusterRole rbacv1.ClusterRole

	clusterRoleKey := types.NamespacedName{
		Name: f.DomainClusterRoleName(),
	}

	err := c.Get(ctx, clusterRoleKey, &clusterRole)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get cluster role %s: %w", f.DomainClusterRoleName(), err)
		}

		// create the cluster role if it does not exist
		clusterRole = rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: f.DomainClusterRoleName(),
				Annotations: map[string]string{
					"provider-registry-ref": f.ProviderRegistryRef(),
					"provider-service-ref":  f.ProviderServiceRef(),
				},
			},
			Rules: []rbacv1.PolicyRule{
				{
					NonResourceURLs: []string{"*"},
					Verbs:           []string{"*"},
				},
			},
		}
		return c.Create(ctx, &clusterRole)
	}

	// update cluster role
	clusterRole.Annotations["provider-registry-ref"] = f.ProviderRegistryRef()
	clusterRole.Annotations["provider-service-ref"] = f.ProviderServiceRef()
	clusterRole.Rules = []rbacv1.PolicyRule{
		{
			NonResourceURLs: []string{"*"},
			Verbs:           []string{"*"},
		},
	}

	return c.Update(ctx, &clusterRole)
}

func (f *FileServerProviderConfig) DeleteFileServerRole(ctx context.Context, c client.Client) error {
	var clusterRole rbacv1.ClusterRole

	clusterRoleKey := types.NamespacedName{
		Name: f.DomainClusterRoleName(),
	}
	err := c.Get(ctx, clusterRoleKey, &clusterRole)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// ClusterRole not found, nothing to delete
			return nil
		}
		return fmt.Errorf("failed to get cluster role %s: %w", f.DomainClusterRoleName(), err)
	}

	return c.Delete(ctx, &clusterRole)
}

func (f *FileServerProviderConfig) UpsertFileServerRoleBinding(ctx context.Context, c client.Client) error {
	var clusterRoleBinding rbacv1.ClusterRoleBinding

	clusterRoleBindingKey := types.NamespacedName{
		Name: f.DomainClusterRoleBindingName(),
	}

	err := c.Get(ctx, clusterRoleBindingKey, &clusterRoleBinding)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get cluster role binding %s: %w", f.DomainClusterRoleBindingName(), err)
		}

		// create the cluster role binding if it does not exist
		clusterRoleBinding = rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: f.DomainClusterRoleBindingName(),
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.SchemeGroupVersion.Group,
				Kind:     "ClusterRole",
				Name:     f.DomainClusterRoleName(),
			},
			Subjects: []rbacv1.Subject{
				{
					Kind: "User",
					Name: constants.Username,
				},
			},
		}
		return c.Create(ctx, &clusterRoleBinding)
	}

	// do not need to update
	return nil
}

func (f *FileServerProviderConfig) DeleteFileServerRoleBinding(ctx context.Context, c client.Client) error {
	var clusterRoleBinding rbacv1.ClusterRoleBinding

	clusterRoleBindingKey := types.NamespacedName{
		Name: f.DomainClusterRoleBindingName(),
	}

	err := c.Get(ctx, clusterRoleBindingKey, &clusterRoleBinding)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// ClusterRoleBinding not found, nothing to delete
			return nil
		}
		return fmt.Errorf("failed to get cluster role binding %s: %w", f.DomainClusterRoleBindingName(), err)
	}

	return c.Delete(ctx, &clusterRoleBinding)
}
