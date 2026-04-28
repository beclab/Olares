package v2

import (
	"context"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/helm"
	"github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// In v2, when we do uninstall operation, we just uninstall the client by default.
// Only the UninstallAll method is used to uninstall all related resources including shared charts.
func (h *HelmOpsV2) UninstallAll() error {
	if !h.isMultiCharts() {
		return h.HelmOps.Uninstall()
	}

	client, err := kubernetes.NewForConfig(h.KubeConfig())
	if err != nil {
		return err
	}

	// uninstall shared charts in priority
	for _, chart := range h.App().SubCharts {
		if chart.Shared {
			namespace := chart.Namespace(h.App().OwnerName)
			appCacheDirs, err := h.tryToGetSharedAppCache(h.Context(), client, namespace)
			if err != nil {
				klog.Warningf("get app %s cache dir failed %v", namespace, err)
			}

			actionConfig, _, err := helm.InitConfig(h.KubeConfig(), namespace)
			if err != nil {
				klog.Errorf("Failed to init helm config namespace=%s err=%v", namespace, err)
				return err
			}
			err = h.Uninstall_(client, actionConfig, namespace, chart.Name)
			if err != nil {
				klog.Errorf("Failed to uninstall app %s err=%v", namespace, err)
				return err
			}

			h.ClearMiddlewareRequests("os-platform")

			err = h.ClearCache(client, appCacheDirs)
			if err != nil {
				klog.Errorf("Failed to clear app cache dirs %v err=%v", appCacheDirs, err)
				return err
			}

			err = h.DeleteNamespace(client, namespace)
			if err != nil {
				klog.Errorf("Failed to delete namespace %s err=%v", namespace, err)
			}
		}
	}

	err = h.HelmOps.Uninstall()
	if err != nil {
		klog.Errorf("Failed to uninstall client err=%v", err)
		return err
	}

	return nil
}

func (h *HelmOpsV2) tryToGetSharedAppCache(ctx context.Context, client kubernetes.Interface, namespace string) (appCacheDirs []string, err error) {

	deployments, err := client.AppsV1().Deployments(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list deployments in namespace %s: %v", namespace, err)
		return nil, err
	}

	appCachePath := "/olares/userdata/Cache/"
	appDirSet := sets.NewString()

	find := func(v *corev1.Volume) {
		if v.HostPath != nil && strings.HasPrefix(v.HostPath.Path, appCachePath) && len(v.HostPath.Path) > len(appCachePath) {
			appDir := app.GetFirstSubDir(v.HostPath.Path, appCachePath)
			if appDir != "" {
				if appDirSet.Has(appDir) {
					return
				}
				appCacheDirs = append(appCacheDirs, appDir)
				appDirSet.Insert(appDir)
			}
		}
	}

	for _, deployment := range deployments.Items {
		for _, v := range deployment.Spec.Template.Spec.Volumes {
			find(&v)
		}
	}

	statefulSets, err := client.AppsV1().StatefulSets(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list statefulsets in namespace %s: %v", namespace, err)
		return nil, err
	}

	for _, statefulSet := range statefulSets.Items {
		for _, v := range statefulSet.Spec.Template.Spec.Volumes {
			find(&v)
		}
	}

	return

}
