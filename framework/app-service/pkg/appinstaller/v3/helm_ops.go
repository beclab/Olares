// Package v3 wraps the v1 HelmOps for v3 apps.
//
// v3 apps are single-chart by definition, so they reuse the v1 install /
// upgrade / uninstall / applyenv pipeline almost verbatim. The wrapper
// exists so that:
//
//  1. versioned.NewHelmOps can dispatch to a v3-typed object (callers can
//     type-assert when they need v3-specific behaviour).
//  2. v3-only divergences can be introduced as method overrides in this
//     package without touching v1 or v2.
//
// Currently the only behavioural difference is AddApplicationLabelsToDeployment;
// Install is also overridden purely to wire Go's method dispatch — its body
// mirrors v1.Install so that the in-flow call to AddApplicationLabelsToDeployment
// resolves to the v3 override (Go embedding does not do virtual dispatch).
package v3

import (
	"context"
	"errors"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	v1 "github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/errcode"
	"github.com/beclab/api/pkg/generated/clientset/versioned"
	"helm.sh/helm/v3/pkg/storage/driver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

var _ v1.HelmOpsInterface = &HelmOpsV3{}

// HelmOpsV3 implements v1.HelmOpsInterface for v3 apps by embedding the v1
// implementation. Callers can type-assert to *HelmOpsV3 in handlers that
// need v3-specific extension points.
type HelmOpsV3 struct {
	*v1.HelmOps
}

// NewHelmOps constructs a v3 HelmOps. It delegates to v1.NewHelmOps for the
// heavy lifting and wraps the result so the dispatcher in
// pkg/appinstaller/versioned can return a typed v3 value.
func NewHelmOps(ctx context.Context, kubeConfig *rest.Config, app *appcfg.ApplicationConfig, token string, options v1.Opt) (v1.HelmOpsInterface, error) {
	v1Ops, err := v1.NewHelmOps(ctx, kubeConfig, app, token, options)
	if err != nil {
		klog.Errorf("Failed to create v3 HelmOps: %v", err)
		return nil, err
	}
	return &HelmOpsV3{HelmOps: v1Ops.(*v1.HelmOps)}, nil
}

// Install mirrors v1.HelmOps.Install. It exists only so the in-flow call to
// AddApplicationLabelsToDeployment resolves to *HelmOpsV3 (Go embedding does
// not redispatch to the outer type). All other steps reuse v1 directly.
func (h *HelmOpsV3) Install() error {
	values, err := h.SetValues()
	if err != nil {
		klog.Errorf("set values err %v", err)
		return err
	}

	if err = h.TaprApply(values, ""); err != nil {
		return err
	}

	err = h.InstallChart(values)
	if err != nil && !errors.Is(err, driver.ErrReleaseExists) {
		klog.Errorf("Failed to install chart err=%v", err)
		h.Uninstall()
		return err
	}

	if err = h.AddApplicationLabelsToDeployment(); err != nil {
		h.Uninstall()
		return err
	}

	isDepClusterScopedApp := false
	client, err := versioned.NewForConfig(h.KubeConfig())
	if err != nil {
		return err
	}
	apps, err := client.AppV1alpha1().Applications().List(h.Context(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, dep := range h.App().Dependencies {
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
		if err = h.AddLabelToNamespaceForDependClusterApp(); err != nil {
			h.Uninstall()
			return err
		}
	}

	if err = h.RegisterOrUnregisterAppProvider(v1.Register); err != nil {
		klog.Errorf("Failed to register app provider err=%v", err)
		h.Uninstall()
		return err
	}

	ok, err := h.WaitForStartUp()
	if err != nil && (errors.Is(err, errcode.ErrPodPending) || errors.Is(err, errcode.ErrServerSidePodPending)) {
		return err
	}
	if !ok {
		h.Uninstall()
		return err
	}
	return nil
}

// AddApplicationLabelsToDeployment is the v3 variant. It reuses v1's patch
// data builder, layers the v3-only labels onto the in-memory payload, then
// hands it back to v1's applier so each resource is patched exactly once.
//
// v3-only labels:
//   - app.bytetrade.io/api-version=v3 on the namespace and on the main
//     deployment/statefulset, so reconcileNetworkPolicy can pick the
//     v3-specific branch instead of the v1/v2 app-np branch.
func (h *HelmOpsV3) AddApplicationLabelsToDeployment() error {
	nsLabels, workloadPatchData := h.BuildDeploymentLabelPatchData()

	nsLabels[constants.AppApiVersionLabel] = constants.AppVersionV3

	if meta, ok := workloadPatchData["metadata"].(map[string]interface{}); ok {
		if labels, ok := meta["labels"].(map[string]string); ok {
			labels[constants.AppApiVersionLabel] = constants.AppVersionV3
		}
	}

	return h.ApplyDeploymentLabelPatch(nsLabels, workloadPatchData)
}
