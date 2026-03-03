package v2

import (
	"context"
	"errors"

	appv1alpha1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"

	v1 "github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/errcode"
	"github.com/beclab/Olares/framework/app-service/pkg/helm"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"helm.sh/helm/v3/pkg/action"
	helmrelease "helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

var _ v1.HelmOpsInterface = &HelmOpsV2{}

type HelmOpsV2 struct {
	*v1.HelmOps
}

func NewHelmOps(ctx context.Context, kubeConfig *rest.Config, app *appcfg.ApplicationConfig, token string, options v1.Opt) (v1.HelmOpsInterface, error) {
	v1Ops, err := v1.NewHelmOps(ctx, kubeConfig, app, token, options)
	if err != nil {
		klog.Errorf("Failed to create HelmOps: %v", err)
		return nil, err
	}

	return &HelmOpsV2{
		HelmOps: v1Ops.(*v1.HelmOps),
	}, nil
}

func (h *HelmOpsV2) Install() error {
	if !h.isMultiCharts() {
		// currently, only multi-charts support in v2 app
		return h.HelmOps.Install()
	}

	if h.isMultiCharts() && !h.hasClusterSharedCharts() {
		err := errors.New("multi-charts app must have at least one cluster shared chart")
		klog.Error(err)
		return err
	}

	var err error
	values, err := h.SetValues()
	if err != nil {
		klog.Errorf("set values err %v", err)
		return err
	}
	if values["isAdmin"].(bool) {
		// force set the admin is owner
		values["admin"] = h.App().OwnerName
	}

	// in v2, if app is multi-charts and has a cluster shared chart,
	//  middleware namespace is always os-platform
	middlewareNamespace := "os-platform"
	err = h.TaprApply(values, middlewareNamespace)
	if err != nil {
		klog.Errorf("tapr apply err %v", err)
		return err
	}

	// add application labels to shared namespace
	err = h.addApplicationLabelsToSharedNamespace()
	if err != nil {
		klog.Errorf("Failed to add application labels to shared namespace err=%v", err)
		return err
	}

	err, sharedInstalled := h.install(values)
	clear := func() {
		if sharedInstalled {
			h.UninstallAll()
		} else {
			h.Uninstall()
		}
	}
	if err != nil && !errors.Is(err, driver.ErrReleaseExists) {
		klog.Errorf("Failed to install chart err=%v", err)
		clear()
		return err
	}

	// just need to add labels to the main ( client ) chart's deployment
	err = h.AddApplicationLabelsToDeployment()
	if err != nil {
		klog.Errorf("Failed to add application labels to deployment err=%v", err)
		clear()
		return err
	}

	// v2 && multi-charts app, add labels to namespace for cluster scoped apps
	err = h.AddLabelToNamespaceForDependClusterApp()
	if err != nil {
		klog.Errorf("Failed to add labels to namespace for cluster scoped apps err=%v", err)
		clear()
		return err
	}

	if err = h.RegisterOrUnregisterAppProvider(v1.Register); err != nil {
		klog.Errorf("Failed to register app provider err=%v", err)
		clear()
		return err
	}
	if h.App().Type == appv1alpha1.Middleware.String() {
		return nil
	}
	ok, err := h.WaitForStartUp()
	if err != nil && (errors.Is(err, errcode.ErrPodPending) || errors.Is(err, errcode.ErrServerSidePodPending)) {
		klog.Errorf("App %s is pending, err=%v", h.App().AppName, err)
		return err
	}
	if !ok {
		klog.Errorf("App %s is not started, err=%v", h.App().AppName, err)
		clear()
		return err
	}

	return nil
}

func (h *HelmOpsV2) isMultiCharts() bool {
	return h.App().IsMultiCharts()
}

func (h *HelmOpsV2) hasClusterSharedCharts() bool {
	return h.App().HasClusterSharedCharts()
}

func (h *HelmOpsV2) install(values map[string]interface{}) (err error, sharedInstalled bool) {
	for _, chart := range h.App().SubCharts {
		if chart.Shared {
			isAdmin, err := kubesphere.IsAdmin(h.Context(), h.KubeConfig(), h.App().OwnerName)
			if err != nil {
				klog.Errorf("Failed to check if user is admin for chart %s: %v", chart.Name, err)
				return err, sharedInstalled
			}

			if !isAdmin {
				klog.Infof("Skipping installation of shared chart %s for non-admin user %s", chart.Name, h.App().OwnerName)
				continue
			}
		}

		_, err := h.status(chart.Name)
		if err == nil {
			if chart.Shared {
				klog.Infof("chart %s already installed, skipping", chart.Name)
				continue
			} else {
				klog.Errorf("chart %s already exists, cannot install again", chart.Name)
				return driver.ErrReleaseExists, sharedInstalled
			}
		}

		if !errors.Is(err, driver.ErrReleaseNotFound) {
			klog.Errorf("Failed to get status for chart %s: %v", chart.Name, err)
			return err, sharedInstalled
		}

		actionConfig := h.ActionConfig()
		settings := h.Settings()
		if chart.Shared {
			// re-create action config for shared chart
			actionConfig, settings, err = helm.InitConfig(h.KubeConfig(), chart.Namespace(h.App().OwnerName))
			if err != nil {
				klog.Errorf("Failed to create action config for shared chart %s: %v", chart.Name, err)
				return err, sharedInstalled
			}

			sharedInstalled = true
		}
		err = helm.InstallCharts(
			h.Context(),
			actionConfig,
			settings,
			chart.Name,
			chart.ChartPath(h.App().AppName),
			h.App().RepoURL,
			chart.Namespace(h.App().OwnerName),
			values,
		)

		if err != nil {
			klog.Errorf("Failed to install chart %s: %v", chart.Name, err)
			return err, sharedInstalled
		}

		klog.Infof("Successfully installed chart %s", chart.Name)
	} // end subcharts loop

	return nil, sharedInstalled
}

func (h *HelmOpsV2) status(releaseName string) (*helmrelease.Release, error) {
	actionConfig := h.ActionConfig()
	var err error
	for _, chart := range h.App().SubCharts {
		if chart.Shared && chart.Name == releaseName {
			// re-create action config for shared chart
			actionConfig, _, err = helm.InitConfig(h.KubeConfig(), chart.Namespace(h.App().OwnerName))
			if err != nil {
				klog.Errorf("Failed to create action config for shared chart %s: %v", chart.Name, err)
				return nil, err
			}
			break
		}
	}
	statusClient := action.NewStatus(actionConfig)
	status, err := statusClient.Run(releaseName)
	if err != nil {
		return nil, err
	}
	return status, nil
}

func (h *HelmOpsV2) addApplicationLabelsToSharedNamespace() error {
	k8s, err := kubernetes.NewForConfig(h.KubeConfig())
	if err != nil {
		return err
	}

	for _, chart := range h.App().SubCharts {
		if !chart.Shared {
			continue
		}

		// Use the shared namespace defined in the chart
		sharedNamespace := chart.Namespace(h.App().OwnerName)
		ns, err := k8s.CoreV1().Namespaces().Get(h.Context(), sharedNamespace, metav1.GetOptions{})
		create := false
		if err != nil {
			if !apierrors.IsNotFound(err) {
				klog.Errorf("Failed to get namespace %s: %v", sharedNamespace, err)
				return err
			}
			// try to create the namespace if not found
			create = true
			ns = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: sharedNamespace,
					Labels: map[string]string{
						"name":                   sharedNamespace,
						"bytetrade.io/ns-shared": "true",
					},
				},
			}
		}

		if ns.Labels == nil {
			ns.Labels = make(map[string]string)
		}
		ns.Labels[constants.ApplicationNameLabel] = h.App().AppName

		if ns.Labels[constants.ApplicationInstallUserLabel] == "" {
			ns.Labels[constants.ApplicationInstallUserLabel] = h.App().OwnerName
		}

		if create {
			if _, err := k8s.CoreV1().Namespaces().Create(h.Context(), ns, metav1.CreateOptions{}); err != nil {
				klog.Errorf("Failed to create namespace %s: %v", sharedNamespace, err)
				return err
			}
		} else {
			if _, err := k8s.CoreV1().Namespaces().Update(h.Context(), ns, metav1.UpdateOptions{}); err != nil {
				klog.Errorf("Failed to update namespace %s: %v", sharedNamespace, err)
				return err
			}
		}
	}

	return nil
}
