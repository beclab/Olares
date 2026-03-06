package provider

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"bytetrade.io/web3os/bfl/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	fsReconcileDelay                = 100 * time.Millisecond
	fsRetryDelay                    = 2 * time.Second
	fileServerNodeServiceAnnotation = "olare/files-node-server"
)

// FileserverReconciler manages Kubernetes resources (Services, ClusterRoles,
// ClusterRoleBindings) for fileserver pods. It is intentionally decoupled
// from the xDS Provider so that proxy-config generation and K8s resource
// management remain independent concerns.
type FileserverReconciler struct {
	cache        ctrlcache.Cache
	client       client.Client
	synced       int32
	debounceCh   chan struct{}
	OnReconciled func() // called after successful reconciliation to trigger downstream re-publish
}

func NewFileserverReconciler(c ctrlcache.Cache, cl client.Client) *FileserverReconciler {
	return &FileserverReconciler{
		cache:      c,
		client:     cl,
		debounceCh: make(chan struct{}, 1),
	}
}

func (r *FileserverReconciler) Name() string { return "fileserver-reconciler" }

func (r *FileserverReconciler) SetupWithManager(ctx context.Context) error {
	podInformer, err := r.cache.GetInformer(ctx, &corev1.Pod{})
	if err != nil {
		return fmt.Errorf("get pod informer: %w", err)
	}

	if _, err = podInformer.AddEventHandler(toolscache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				return false
			}
			return pod.Labels["app"] == "files"
		},
		Handler: toolscache.ResourceEventHandlerFuncs{
			AddFunc:    func(_ interface{}) { r.notifyChanged() },
			UpdateFunc: func(_, _ interface{}) { r.notifyChanged() },
			DeleteFunc: func(_ interface{}) { r.notifyChanged() },
		},
	}); err != nil {
		return fmt.Errorf("add fileserver pod event handler: %w", err)
	}

	klog.Info("fileserver-reconciler: event handler registered")
	return nil
}

func (r *FileserverReconciler) Start(ctx context.Context) error {
	atomic.StoreInt32(&r.synced, 1)
	klog.Info("fileserver-reconciler: cache synced, running initial reconcile")
	if err := r.reconcile(ctx); err != nil {
		klog.Errorf("fileserver-reconciler: initial reconcile failed, will retry: %v", err)
		r.scheduleRetry(ctx)
	}
	r.debounceLoop(ctx)
	klog.Info("fileserver-reconciler: stopped")
	return nil
}

func (r *FileserverReconciler) notifyChanged() {
	if atomic.LoadInt32(&r.synced) == 0 {
		return
	}
	select {
	case r.debounceCh <- struct{}{}:
	default:
	}
}

func (r *FileserverReconciler) debounceLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-r.debounceCh:
			timer := time.NewTimer(fsReconcileDelay)
		drain:
			for {
				select {
				case <-r.debounceCh:
					if !timer.Stop() {
						select {
						case <-timer.C:
						default:
						}
					}
					timer.Reset(fsReconcileDelay)
				case <-timer.C:
					break drain
				case <-ctx.Done():
					timer.Stop()
					return
				}
			}
			if err := r.reconcile(ctx); err != nil {
				klog.Errorf("fileserver-reconciler: reconcile failed, retrying in %v: %v", fsRetryDelay, err)
				r.scheduleRetry(ctx)
			}
		}
	}
}

func (r *FileserverReconciler) scheduleRetry(ctx context.Context) {
	go func() {
		select {
		case <-time.After(fsRetryDelay):
			r.notifyChanged()
		case <-ctx.Done():
		}
	}()
}

func (r *FileserverReconciler) reconcile(ctx context.Context) error {
	var podList corev1.PodList
	if err := r.cache.List(ctx, &podList, client.MatchingLabels{"app": "files"}); err != nil {
		return fmt.Errorf("list files pods: %w", err)
	}

	var nodeList corev1.NodeList
	if err := r.cache.List(ctx, &nodeList); err != nil {
		return fmt.Errorf("list nodes: %w", err)
	}

	podMap := make(map[string]*corev1.Pod)
	for i := range podList.Items {
		pod := &podList.Items[i]
		if pod.Labels["app"] == "files" && pod.Status.PodIP != "" {
			podMap[pod.Spec.NodeName] = pod
		}
	}

	serviceNamespace := fmt.Sprintf("user-system-%s", constants.Username)
	var serviceList []string

	for nodeName, pod := range podMap {
		svcCfg := proxyServiceConfig{NodeName: nodeName, Namespace: serviceNamespace}
		if err := svcCfg.upsert(ctx, r.client); err != nil {
			return fmt.Errorf("upsert proxy service for %s: %w", nodeName, err)
		}

		fsCfg := fileServerProviderConfig{
			NodeName:         nodeName,
			Pod:              pod,
			ServiceName:      svcCfg.serviceName(),
			ServiceNamespace: serviceNamespace,
		}
		if err := fsCfg.upsertRole(ctx, r.client); err != nil {
			return fmt.Errorf("upsert cluster role for %s: %w", nodeName, err)
		}
		if err := fsCfg.upsertRoleBinding(ctx, r.client); err != nil {
			return fmt.Errorf("upsert cluster role binding for %s: %w", nodeName, err)
		}

		serviceList = append(serviceList, svcCfg.serviceName())
	}

	r.cleanupOrphanedServices(ctx, serviceNamespace, serviceList)
	klog.Infof("fileserver-reconciler: reconciled %d fileserver nodes", len(podMap))

	if r.OnReconciled != nil {
		r.OnReconciled()
	}
	return nil
}

func (r *FileserverReconciler) cleanupOrphanedServices(ctx context.Context, namespace string, activeServices []string) {
	var svcList corev1.ServiceList
	if err := r.client.List(ctx, &svcList, client.InNamespace(namespace)); err != nil {
		klog.Errorf("fileserver-reconciler: list services in %s: %v", namespace, err)
		return
	}

	for i := range svcList.Items {
		svc := &svcList.Items[i]
		if !strings.HasPrefix(svc.Name, "files-") {
			continue
		}
		if svc.Annotations[fileServerNodeServiceAnnotation] != "true" {
			continue
		}
		if slices.Contains(activeServices, svc.Name) {
			continue
		}
		klog.Infof("fileserver-reconciler: deleting orphaned service %s/%s", namespace, svc.Name)
		if err := r.client.Delete(ctx, svc); err != nil {
			klog.Errorf("fileserver-reconciler: delete orphaned service %s: %v", svc.Name, err)
		}
	}
}

type proxyServiceConfig struct {
	NodeName  string
	Namespace string
}

func (p *proxyServiceConfig) serviceName() string {
	return fmt.Sprintf("files-%s", p.NodeName)
}

func (p *proxyServiceConfig) upsert(ctx context.Context, c client.Client) error {
	var svc corev1.Service
	key := types.NamespacedName{Namespace: p.Namespace, Name: p.serviceName()}
	err := c.Get(ctx, key, &svc)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("get proxy service %s: %w", p.serviceName(), err)
		}

		svc = corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      p.serviceName(),
				Namespace: p.Namespace,
				Annotations: map[string]string{
					fileServerNodeServiceAnnotation: "true",
				},
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{"app": "systemserver"},
				Type:     corev1.ServiceTypeClusterIP,
				Ports: []corev1.ServicePort{{
					Name:       "rbac-proxy",
					Port:       28080,
					TargetPort: intstr.FromInt(28080),
				}},
			},
		}
		return c.Create(ctx, &svc)
	}
	return nil
}

type fileServerProviderConfig struct {
	NodeName         string
	Pod              *corev1.Pod
	ServiceName      string
	ServiceNamespace string
}

func (f *fileServerProviderConfig) roleName() string {
	return fmt.Sprintf("%s:files-frontend-domain-%s", constants.Username, f.NodeName)
}

func (f *fileServerProviderConfig) roleBindingName() string {
	return fmt.Sprintf("user:%s:files-frontend-domain-%s", constants.Username, f.NodeName)
}

func (f *fileServerProviderConfig) providerRegistryRef() string {
	return fmt.Sprintf("%s/%s", f.ServiceNamespace, f.ServiceName)
}

func (f *fileServerProviderConfig) providerServiceRef() string {
	return fmt.Sprintf("%s:80", f.Pod.Status.PodIP)
}

func (f *fileServerProviderConfig) upsertRole(ctx context.Context, c client.Client) error {
	var role rbacv1.ClusterRole
	key := types.NamespacedName{Name: f.roleName()}

	err := c.Get(ctx, key, &role)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("get cluster role %s: %w", f.roleName(), err)
		}

		role = rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: f.roleName(),
				Annotations: map[string]string{
					"provider-registry-ref": f.providerRegistryRef(),
					"provider-service-ref":  f.providerServiceRef(),
				},
			},
			Rules: []rbacv1.PolicyRule{{
				NonResourceURLs: []string{"*"},
				Verbs:           []string{"*"},
			}},
		}
		return c.Create(ctx, &role)
	}

	if role.Annotations == nil {
		role.Annotations = make(map[string]string)
	}
	role.Annotations["provider-registry-ref"] = f.providerRegistryRef()
	role.Annotations["provider-service-ref"] = f.providerServiceRef()
	role.Rules = []rbacv1.PolicyRule{{
		NonResourceURLs: []string{"*"},
		Verbs:           []string{"*"},
	}}
	return c.Update(ctx, &role)
}

func (f *fileServerProviderConfig) upsertRoleBinding(ctx context.Context, c client.Client) error {
	var binding rbacv1.ClusterRoleBinding
	key := types.NamespacedName{Name: f.roleBindingName()}

	err := c.Get(ctx, key, &binding)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("get cluster role binding %s: %w", f.roleBindingName(), err)
		}

		binding = rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: f.roleBindingName(),
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.SchemeGroupVersion.Group,
				Kind:     "ClusterRole",
				Name:     f.roleName(),
			},
			Subjects: []rbacv1.Subject{{
				Kind: "User",
				Name: constants.Username,
			}},
		}
		return c.Create(ctx, &binding)
	}
	return nil
}
