package appinstaller

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/helm"
	"github.com/beclab/Olares/framework/app-service/pkg/tapr"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"

	"github.com/emicklei/go-restful/v3"
	"github.com/go-resty/resty/v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/storage/driver"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// UninstallAll do a uninstall operation for release.
func (h *HelmOps) UninstallAll() error {
	client, err := kubernetes.NewForConfig(h.kubeConfig)
	if err != nil {
		return err
	}

	appCacheDirs, err := apputils.TryToGetAppdataDirFromDeployment(h.ctx, h.app.Namespace, h.app.AppName, h.app.OwnerName)
	if err != nil {
		klog.Warningf("get app %s cache dir failed %v", h.app.AppName, err)
	}

	err = h.Uninstall_(client, h.actionConfig, h.app.Namespace, h.app.AppName)
	if err != nil {
		klog.Errorf("Failed to uninstall app %s err=%v", h.app.AppName, err)
		return err
	}

	h.ClearMiddlewareRequests(fmt.Sprintf("%s-%s", "user-system", h.app.OwnerName))

	err = h.ClearCache(client, appCacheDirs)
	if err != nil {
		klog.Errorf("Failed to clear app cache dirs %v err=%v", appCacheDirs, err)
		return err
	}

	err = h.DeleteNamespace(client, h.app.Namespace)
	if err != nil {
		klog.Errorf("Failed to delete namespace %s err=%v", h.app.Namespace, err)
	}

	return err
}

func (h *HelmOps) Uninstall_(client kubernetes.Interface, actionConfig *action.Configuration,
	namespace, releaseName string) error {
	if !apputils.IsProtectedNamespace(namespace) {
		pvcs, err := client.CoreV1().PersistentVolumeClaims(namespace).List(h.ctx, metav1.ListOptions{})
		if err != nil {
			return err
		}
		for _, pvc := range pvcs.Items {
			err = client.CoreV1().PersistentVolumeClaims(pvc.Namespace).Delete(h.ctx, pvc.Name, metav1.DeleteOptions{})
			if err != nil && !apierrors.IsNotFound(err) {
				return err
			}
		}
	}

	err := helm.UninstallCharts(actionConfig, releaseName)
	if err != nil && !errors.Is(err, driver.ErrReleaseNotFound) {
		klog.Errorf("failed to uninstall app %s, err=%v", releaseName, err)
		return err
	}

	h.app.Permission = parseAppPermission(h.app.Permission)
	var perm []appcfg.ProviderPermission
	for _, p := range h.app.Permission {
		if t, ok := p.([]appcfg.ProviderPermission); ok {
			perm = append(perm, t...)
		}
	}

	permCfg, err := apputils.ProviderPermissionsConvertor(perm).ToPermissionCfg(h.ctx, h.app.OwnerName, h.options.MarketSource)
	if err != nil {
		klog.Errorf("Failed to convert app permissions for %s: %v", h.app.AppName, err)
		return err
	}

	err = h.unregisterAppPerm(h.app.ServiceAccountName, h.app.OwnerName, permCfg)
	if err != nil {
		klog.Warningf("Failed to unregister app err=%v", err)
	}

	err = h.RegisterOrUnregisterAppProvider(Unregister)
	if err != nil {
		klog.Warningf("Failed to unregister app provider err=%v", err)
	}

	return nil
}

func (h *HelmOps) ClearCache(client kubernetes.Interface, appCacheDirs []string) error {
	if len(appCacheDirs) > 0 {
		klog.Infof("clear app cache dirs: %v", appCacheDirs)

		c := resty.New().SetTimeout(2 * time.Second).
			SetAuthToken(h.token)
		nodes, e := client.CoreV1().Nodes().List(h.ctx, metav1.ListOptions{})

		if e == nil {
			formattedAppCacheDirs := apputils.FormatCacheDirs(appCacheDirs)

			for _, n := range nodes.Items {
				URL := fmt.Sprintf(constants.AppDataDirURL, n.Name)
				c.SetHeader("X-Terminus-Node", n.Name)
				c.SetHeader("X-Bfl-User", h.app.OwnerName)
				res, e := c.R().SetBody(map[string]interface{}{
					"dirents": formattedAppCacheDirs,
				}).Delete(URL)
				if e != nil {
					klog.Errorf("Failed to delete dir err=%v", e)
				}
				if res.StatusCode() != http.StatusOK {
					klog.Infof("delete app cache failed with: %v", res.String())
				}
			}
		} else {
			klog.Errorf("Failed to get nodes err=%v", e)
		}
	}
	return nil
}

func (h *HelmOps) ClearMiddlewareRequests(middlewareNamespace string) {
	// delete middleware requests crd
	for _, mt := range middlewareTypes {
		name := fmt.Sprintf("%s-%s", h.app.AppName, mt)
		err := tapr.DeleteMiddlewareRequest(h.ctx, h.kubeConfig, middlewareNamespace, name)
		if err != nil && !apierrors.IsNotFound(err) {
			klog.Errorf("Failed to delete middleware request namespace=%s name=%s err=%v", middlewareNamespace, name, err)
		}
	}
}

func (h *HelmOps) DeleteNamespace(client kubernetes.Interface, namespace string) error {
	if !apputils.IsProtectedNamespace(namespace) {
		klog.Infof("deleting namespace %s", namespace)
		err := client.CoreV1().Namespaces().Delete(h.ctx, namespace, metav1.DeleteOptions{})
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}

func (h *HelmOps) unregisterAppPerm(sa *string, ownerName string, perm []appcfg.PermissionCfg) error {
	requires := make([]appcfg.PermissionRequire, 0, len(perm))
	for _, p := range perm {
		requires = append(requires, appcfg.PermissionRequire{
			ProviderName:      p.ProviderName,
			ProviderNamespace: p.GetNamespace(ownerName),
			ServiceAccount:    sa,
			ProviderAppName:   p.AppName,
			ProviderDomain:    p.Domain,
		})
	}

	register := appcfg.PermissionRegister{
		App:   h.app.AppName,
		AppID: h.app.AppID,
		Perm:  requires,
	}

	url := fmt.Sprintf("http://%s/permission/v2alpha1/unregister", h.systemServerHost())
	client := resty.New()

	resp, err := client.SetTimeout(2*time.Second).R().
		SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetAuthToken(h.token).
		SetBody(register).Post(url)
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		dump, e := httputil.DumpRequest(resp.Request.RawRequest, true)
		if e == nil {
			klog.Errorf("Failed to get response body=%s url=%s", string(dump), url)
		}

		return errors.New(string(resp.Body()))
	}

	return nil
}

func (h *HelmOps) Uninstall() error {
	return h.UninstallAll()
}
