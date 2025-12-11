package v2alpha1

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"bytetrade.io/web3os/system-server/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
)

func (h *handler) createRefForProvider(ctx context.Context, appName, appNamespace string, provider *Provider) error {
	providerRefs, err := h.getProviderRefs(appNamespace, provider)
	if err != nil {
		klog.Error(err)
		return err
	}

	// create cluster role  for provider refs
	for _, r := range providerRefs {
		if err := h.createCusterRoleForRef(ctx, r, provider); err != nil {
			klog.Error(err)
			return err
		}
	}

	// create service
	err = h.createServiceForProviderProxy(ctx, appName, provider.Service)
	if err != nil {
		klog.Error("create service for provider proxy err,", err)
	}
	return err
}

func (h *handler) createCusterRoleForRef(ctx context.Context, ref string, provider *Provider) error {
	// Create a ClusterRole for the provider reference
	klog.Info("Creating ClusterRole for provider reference, ", ref, ", provider: ", provider.Paths)

	roleName := h.getRoleNameForRef(ref)
	role := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: roleName,
			Annotations: map[string]string{
				ProviderRefAnnotation:     ref,
				ProviderServiceAnnotation: provider.Service,
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:           provider.Verbs,
				NonResourceURLs: provider.Paths,
			},
		},
	}

	_, err := h.kubeClient.RbacV1().ClusterRoles().Get(ctx, roleName, metav1.GetOptions{})
	if err == nil {
		klog.Infof("Cluster role %s already exists for provider reference %s", roleName, ref)
		err = h.kubeClient.RbacV1().ClusterRoles().Delete(ctx, roleName, metav1.DeleteOptions{})
		if err != nil {
			klog.Error("delete existing cluster role for provider err,", err)
			return err
		}
		klog.Infof("Deleted existing cluster role %s for provider reference %s", roleName, ref)
	}

	_, err = h.kubeClient.RbacV1().ClusterRoles().Create(ctx, role, metav1.CreateOptions{})
	if err != nil {
		klog.Error("create cluster role for privder err,", err)
		return err
	}

	return nil
}

func (h *handler) createServiceForProviderProxy(ctx context.Context, providerName, providerService string) error {
	klog.Info("Creating service for provider proxy, ", providerName, ", service: ", providerService)
	portStr := strings.Split(constants.ProxyServerListenAddress, ":")[1]
	port, err := strconv.Atoi(portStr)

	servicePortStrToken := strings.Split(providerService, ":")
	servicePort := 80
	if len(servicePortStrToken) > 1 {
		servicePort, err = strconv.Atoi(servicePortStrToken[1])
		if err != nil {
			klog.Error("invalid port for provider proxy, ", providerService)
			return err
		}
	}

	if err != nil {
		klog.Error("invalid port for provider proxy, ", constants.ProxyServerListenAddress)
		return err
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      providerName,
			Namespace: constants.MyNamespace,
		},
		Spec: corev1.ServiceSpec{
			Type:         corev1.ServiceTypeExternalName,
			ExternalName: "system-server.user-system-" + constants.Owner + ".svc.cluster.local",
			Ports: []corev1.ServicePort{
				{
					Port:       int32(servicePort),
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt32(int32(port)),
				},
			},
		},
	}

	// delete if exists
	if _, err := h.kubeClient.CoreV1().Services(constants.MyNamespace).Get(ctx, providerName, metav1.GetOptions{}); err == nil {
		klog.Info("Service for provider proxy already exists, delete it first.")
		if err = h.kubeClient.CoreV1().Services(constants.MyNamespace).Delete(ctx, providerName, metav1.DeleteOptions{}); err != nil {
			klog.Error("delete service for provider proxy err,", err)
			return err
		}
	}

	_, err = h.kubeClient.CoreV1().Services(constants.MyNamespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		klog.Error("create service for provider proxy err,", err)
		return err
	}

	return nil
}

func (h *handler) deleteRefForProvider(ctx context.Context, appName, appNamespace string, provider *Provider) error {
	providerRefs, err := h.getProviderRefs(appNamespace, provider)
	if err != nil {
		klog.Error(err)
		return err
	}

	errs := []error{}
	for _, p := range providerRefs {
		err := h.deleteCusterRoleForRef(ctx, p)
		if err != nil {
			errs = append(errs, err)
		}
	}

	err = h.deleteServiceForProviderProxy(ctx, provider.Service)
	if err != nil {
		errs = append(errs, err)
	}
	return errors.NewAggregate(errs)
}

func (h *handler) deleteServiceForProviderProxy(ctx context.Context, providerName string) error {
	klog.Info("Deleting service for provider proxy, ", providerName)
	if err := h.kubeClient.CoreV1().Services(constants.MyNamespace).Delete(ctx, providerName, metav1.DeleteOptions{}); err != nil {
		klog.Error("delete service for provider proxy err,", err)
		return err
	}
	return nil
}

func (h *handler) deleteCusterRoleForRef(ctx context.Context, ref string) error {
	roleName := h.getRoleNameForRef(ref)

	if err := h.kubeClient.RbacV1().ClusterRoles().Delete(ctx, roleName, metav1.DeleteOptions{}); err != nil {
		klog.Error("delete cluster role for provider err,", err)
		return err
	}

	return nil
}

func (h *handler) getRoleNameForRef(ref string) string {
	return GetRoleNameForRef(ref)
}

func (h *handler) getProviderRefs(appNamespace string, provider *Provider) ([]string, error) {
	providerRefs := []string{
		ProviderRefName(provider.Name, appNamespace),
	}
	if provider.Domain != "" {
		strToken := strings.Split(provider.Domain, ".")
		if len(strToken) < 3 { // must be <appid>.<username>.xxx.yy.zzz
			err := fmt.Errorf("invalid provider domain: %s", provider.Domain)
			klog.Error(err)
			return nil, err
		}

		providerRefs = append(providerRefs, ProviderRefFromHost(provider.Domain))
	}

	return providerRefs, nil
}
