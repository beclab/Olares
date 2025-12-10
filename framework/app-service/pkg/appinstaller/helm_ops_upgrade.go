package appinstaller

import (
	"encoding/json"
	"fmt"
	"time"

	appv1alpha1 "bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/app-service/pkg/appcfg"
	"bytetrade.io/web3os/app-service/pkg/constants"
	"bytetrade.io/web3os/app-service/pkg/generated/clientset/versioned"
	"bytetrade.io/web3os/app-service/pkg/helm"
	"bytetrade.io/web3os/app-service/pkg/users/userspace"
	apputils "bytetrade.io/web3os/app-service/pkg/utils/app"
	helmrelease "helm.sh/helm/v3/pkg/release"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Upgrade do a upgrade operation for release.
func (h *HelmOps) Upgrade() error {
	status, err := h.status()
	if err != nil {
		klog.Errorf("get release status failed %v", err)
		return err
	}
	if status.Info.Status == helmrelease.StatusDeployed {
		return h.upgrade()
	}
	return fmt.Errorf("cannot upgrade release %s/%s, current state is %s", h.app.Namespace, h.app.AppName, status.Info.Status)
}

func (h *HelmOps) upgrade() error {
	values, err := h.SetValues()
	if err != nil {
		return err
	}

	err = h.TaprApply(values, "")
	if err != nil {
		return err
	}

	err = helm.UpgradeCharts(h.ctx, h.actionConfig, h.settings, h.app.AppName, h.app.ChartsName, h.app.RepoURL, h.app.Namespace, values, true)
	if err != nil {
		klog.Errorf("Failed to upgrade chart name=%s err=%v", h.app.AppName, err)
		return err
	}
	err = h.AddApplicationLabelsToDeployment()
	if err != nil {
		//h.rollBack()
		return err
	}

	isDepClusterScopedApp := false
	clientset, err := versioned.NewForConfig(h.kubeConfig)
	if err != nil {
		return err
	}
	apps, err := clientset.AppV1alpha1().Applications().List(h.ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, dep := range h.app.Dependencies {
		if dep.Type == constants.DependencyTypeSystem {
			continue
		}
		for _, app := range apps.Items {
			if app.Spec.Name == dep.Name && app.Spec.Settings["clusterScoped"] == "true" {
				isDepClusterScopedApp = true
				break
			}
		}
	}

	if isDepClusterScopedApp {
		err = h.AddLabelToNamespaceForDependClusterApp()
		if err != nil {
			//h.rollBack()
			return err
		}
	}

	if h.app.Type == appv1alpha1.Middleware.String() {
		return nil
	}

	err = h.UpdatePolicy()
	if err != nil {
		klog.Errorf("Failed to update policy err=%v", err)
		//h.rollBack()
		return err
	}

	if err = h.RegisterOrUnregisterAppProvider(Register); err != nil {
		klog.Errorf("Failed to register app provider err=%v", err)
		return err
	}

	ok, err := h.WaitForStartUp()
	if !ok {
		// canceled
		//h.rollBack()
		klog.Error("App upgrade start up failed, ", err)
		return err
	}

	klog.Infof("App %s upgraded successfully", h.app.AppName)
	return nil
}

func (h *HelmOps) UpdatePolicy() (err error) {
	appClient, err := versioned.NewForConfig(h.kubeConfig)
	if err != nil {
		return err
	}
	var workload client.Object
	if userspace.IsSysApp(h.app.AppName) {
		application, err := appClient.AppV1alpha1().Applications().Get(h.ctx,
			appv1alpha1.AppResourceName(h.app.AppName, h.app.Namespace), metav1.GetOptions{})
		if err != nil {
			return err
		}
		clientset, err := kubernetes.NewForConfig(h.kubeConfig)
		if err != nil {
			return err
		}
		workload, err = clientset.AppsV1().Deployments(h.app.Namespace).
			Get(h.ctx, application.Spec.DeploymentName, metav1.GetOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}

		if apierrors.IsNotFound(err) {
			// try to find statefulset
			workload, err = clientset.AppsV1().StatefulSets(h.app.Namespace).
				Get(h.ctx, application.Spec.DeploymentName, metav1.GetOptions{})
			if err != nil && !apierrors.IsNotFound(err) {
				return err
			}
		}

		if err == nil {
			// found workload
			entrancesLabel := workload.GetAnnotations()[constants.ApplicationEntrancesKey]
			entrances, err := ToEntrances(entrancesLabel)
			if err != nil {
				return err
			}
			h.app.Entrances = entrances
		}
	}
	for i, v := range h.app.Entrances {
		if v.AuthLevel == "" {
			h.app.Entrances[i].AuthLevel = constants.AuthorizationLevelOfPrivate
		}
	}

	var policyStr string
	if !userspace.IsSysApp(h.app.AppName) {
		if appCfg, err := appcfg.GetAppInstallationConfig(h.app.AppName, h.app.OwnerName); err != nil {
			klog.Infof("Failed to get app configuration appName=%s owner=%s err=%v", h.app.AppName, h.app.OwnerName, err)
		} else {
			policyStr, err = getApplicationPolicy(appCfg.Policies, h.app.Entrances)
			if err != nil {
				klog.Errorf("Failed to encode json err=%v", err)
			}
		}
	} else {
		// sys applications.
		type Policies struct {
			Policies []appcfg.Policy `json:"policies"`
		}
		var (
			applicationPoliciesFromAnnotation string
			ok                                bool
		)
		if workload != nil {
			applicationPoliciesFromAnnotation, ok = workload.GetAnnotations()[constants.ApplicationPolicies]
		}

		var policy Policies
		if ok {
			err := json.Unmarshal([]byte(applicationPoliciesFromAnnotation), &policy)
			if err != nil {
				klog.Errorf("Failed to unmarshal applicationPoliciesFromAnnotation err=%v", err)
			}
		}

		// transform from Policy to AppPolicy
		var appPolicies []appcfg.AppPolicy
		for _, p := range policy.Policies {
			d, _ := time.ParseDuration(p.Duration)
			appPolicies = append(appPolicies, appcfg.AppPolicy{
				EntranceName: p.EntranceName,
				URIRegex:     p.URIRegex,
				Level:        p.Level,
				OneTime:      p.OneTime,
				Duration:     d,
			})
		}
		policyStr, err = getApplicationPolicy(appPolicies, h.app.Entrances)
		if err != nil {
			klog.Errorf("Failed to encode json err=%v", err)
		}
	} // end if sys app

	patchData := map[string]interface{}{
		"spec": map[string]interface{}{
			"entrances": h.app.Entrances,
		},
	}
	if len(policyStr) > 0 {
		patchData = map[string]interface{}{
			"spec": map[string]interface{}{
				"entrances": h.app.Entrances,
				"settings": map[string]string{
					"policy": policyStr,
				},
			},
		}
	}
	patchByte, err := json.Marshal(patchData)
	if err != nil {
		klog.Errorf("Failed to marshal patch data err=%v", err)
		return err
	}
	name, _ := apputils.FmtAppMgrName(h.app.AppName, h.app.OwnerName, h.app.Namespace)
	_, err = appClient.AppV1alpha1().Applications().Patch(h.ctx, name, types.MergePatchType, patchByte, metav1.PatchOptions{})
	if err != nil {
		klog.Errorf("Failed to patch application %s err=%v", name, err)
		return err
	}

	return nil
}
