package apiserver

import (
	"context"
	"errors"
	"fmt"

	"bytetrade.io/web3os/app-service/pkg/apiserver/api"
	"bytetrade.io/web3os/app-service/pkg/client/clientset"
	"bytetrade.io/web3os/app-service/pkg/constants"
	"bytetrade.io/web3os/app-service/pkg/helm"
	"bytetrade.io/web3os/app-service/pkg/utils"

	"github.com/emicklei/go-restful/v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/storage/driver"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func (h *Handler) enableServiceSync(req *restful.Request, resp *restful.Response) {
	/**
	values:
		bfl.username
		bfl.nodeName
		userspace.appCache
	*/
	owner := req.Attribute(constants.UserContextAttribute).(string)                     // get owner from request token
	client := req.Attribute(constants.KubeSphereClientAttribute).(*clientset.ClientSet) // session client
	userspace := utils.UserspaceName(owner)

	bfl, err := h.findBFL(req.Request.Context(), client, owner, userspace)
	if err != nil {
		klog.Errorf("Failed to find sts bfl owner=%s err=%v", owner, err)
		api.HandleError(resp, req, err)
		return
	}

	_, pvPath, err := h.findPVC(req.Request.Context(), client, userspace, bfl)
	if err != nil {
		klog.Errorf("Failed to find pvc owner=%s err=%v", owner, err)
		api.HandleError(resp, req, err)
		return
	}

	vals := make(map[string]interface{})
	vals["bfl"] = map[string]interface{}{
		"username": owner,
		"nodeName": bfl.Spec.NodeName,
	}

	// appCache hostpath to mount with apps
	vals["userspace"] = map[string]interface{}{
		"appCache": fmt.Sprintf("%s/Cache", pvPath),
		"userData": fmt.Sprintf("%s/Home", pvPath),
	}

	err = h.installService(req.Request.Context(), bfl, "sync", owner, userspace, vals)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(api.Response{Code: 200})
}

func (h *Handler) disableServiceSync(req *restful.Request, resp *restful.Response) {
	owner := req.Attribute(constants.UserContextAttribute).(string) // get owner from request token
	userspace := utils.UserspaceName(owner)

	err := h.uninstallService(req.Request.Context(), "sync", owner, userspace)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(api.Response{Code: 200})
}

func (h *Handler) enableServiceBackup(req *restful.Request, resp *restful.Response) {
	/**
	values:
		bfl.username
		bfl.nodeName
		pvc.userspace
	*/

	owner := req.Attribute(constants.UserContextAttribute).(string)                     // get owner from request token
	client := req.Attribute(constants.KubeSphereClientAttribute).(*clientset.ClientSet) // session client
	userspace := utils.UserspaceName(owner)

	bfl, err := h.findBFL(req.Request.Context(), client, owner, userspace)
	if err != nil {
		klog.Errorf("Failed to find sts bfl owner=%s err=%v", owner, err)
		api.HandleError(resp, req, err)
		return
	}

	_, pvPath, err := h.findPVC(req.Request.Context(), client, userspace, bfl)
	if err != nil {
		klog.Errorf("Failed to find pvc owner=%s err=%v", owner, err)
		api.HandleError(resp, req, err)
		return
	}

	vals := make(map[string]interface{})
	vals["bfl"] = map[string]interface{}{
		"username": owner,
		"nodeName": bfl.Spec.NodeName,
	}

	vals["userspace"] = map[string]interface{}{
		"appCache": fmt.Sprintf("%s/Cache", pvPath),
		"userData": fmt.Sprintf("%s/Home", pvPath),
	}

	err = h.installService(req.Request.Context(), bfl, "backup", owner, userspace, vals)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(api.Response{Code: 200})
}

func (h *Handler) disableServiceBackup(req *restful.Request, resp *restful.Response) {
	owner := req.Attribute(constants.UserContextAttribute).(string) // get owner from request token
	userspace := utils.UserspaceName(owner)

	err := h.uninstallService(req.Request.Context(), "backup", owner, userspace)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(api.Response{Code: 200})
}

func (h *Handler) findBFL(ctx context.Context, clientSet *clientset.ClientSet, owner, userspace string) (*corev1.Pod, error) {
	pods, err := clientSet.KubeClient.Kubernetes().CoreV1().Pods(userspace).
		List(ctx, metav1.ListOptions{LabelSelector: "tier=bfl"})
	if err != nil {
		klog.Errorf("Failed to get bfl pod owner=%s err=%v", owner, err)
		return nil, err
	}

	if len(pods.Items) == 0 {
		return nil, errors.New("launcher not found")
	}

	bfl := pods.Items[0]

	return &bfl, nil
}

/*
*
find launcher's pvc
returns pvc's Name, pv's local path, err
*/
func (h *Handler) findPVC(ctx context.Context, clientSet *clientset.ClientSet, userspace string, bfl *corev1.Pod) (string, string, error) {
	vols := bfl.Spec.Volumes
	if len(vols) < 1 {
		return "", "", errors.New("user space not found")
	}

	// find user space pvc
	for _, vol := range vols {
		if vol.Name == constants.UserSpaceDirPVC {
			if vol.PersistentVolumeClaim != nil {
				// find user space path
				pvc, err := clientSet.KubeClient.Kubernetes().CoreV1().PersistentVolumeClaims(userspace).Get(ctx,
					vol.PersistentVolumeClaim.ClaimName,
					metav1.GetOptions{})
				if err != nil {
					return "", "", err
				}

				pv, err := clientSet.KubeClient.Kubernetes().CoreV1().PersistentVolumes().Get(ctx, pvc.Spec.VolumeName, metav1.GetOptions{})
				if err != nil {
					return "", "", err
				}

				if pv.Spec.Local != nil {
					return pvc.Name, pv.Spec.Local.Path, nil
				}
				return pvc.Name, pv.Spec.HostPath.Path, nil

			}
		}
	}

	return "", "", errors.New("userspace PVC not found")
}

func (h *Handler) installService(ctx context.Context, bfl *corev1.Pod, servicename,
	owner, userspace string, vals map[string]interface{}) error {
	actionCfg, settings, err := helm.InitConfig(h.kubeConfig, userspace)
	if err != nil {
		klog.Errorf("Failed to init helm config owner=%s err=%v", owner, err)
		return err
	}

	name := helm.ReleaseName(servicename, owner)

	if installed, err := h.isInstalled(actionCfg, name); err != nil {
		klog.Errorf("Failed to get helm history owner=%s err=%v", owner, err)
		return err
	} else if installed {
		return errors.New("service is enabled. Do not do repeat")
	}

	err = helm.InstallCharts(ctx, actionCfg, settings,
		name, constants.UserChartsPath+"/"+servicename, "", bfl.Namespace, vals)

	if err != nil {
		klog.Errorf("Failed to install chart owner=%s err=%v", owner, err)
		return err
	}

	return nil
}

func (h *Handler) uninstallService(ctx context.Context, serviceName, owner, userspace string) error {
	actionCfg, _, err := helm.InitConfig(h.kubeConfig, userspace)
	if err != nil {
		klog.Errorf("Failed to build init config owner=%s service=%s err=%v", owner, serviceName, err)
		return err
	}

	releaseName := helm.ReleaseName(serviceName, owner)
	err = helm.UninstallCharts(actionCfg, releaseName)
	if err != nil {
		klog.Errorf("Failed to uninstall system service owner=%s service=%s err=%v", owner, serviceName, err)
		return err
	}

	return nil
}

func (h *Handler) isInstalled(actionconfig *action.Configuration, releasename string) (bool, error) {
	histClient := action.NewHistory(actionconfig)
	histClient.Max = 1
	_, err := histClient.Run(releasename)

	if err != nil {
		if err == driver.ErrReleaseNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
