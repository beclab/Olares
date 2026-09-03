package appinstaller

import (
	"github.com/beclab/Olares/framework/app-service/pkg/helm"
	"k8s.io/klog/v2"
)

// Scale issues a helm upgrade whose only purpose is to retarget
// .Values.workloads.<name>.replicaCount. It is the second leg of the
// two-phase install / upgrade flow: SetValues(isInstallOp=true) forces
// replicas to 0 during the initial helm install so the pressure gate
// can run against a quiescent release, and Scale then drives the
// workloads to their final size.
//
// The implementation only sends the .workloads sub-tree and relies on
// helm UpgradeCharts(ReuseValues) to merge it over the previous release's
// values — every other input (admin / userspace / domain / service) is
// preserved. ReuseValues (not ResetThenReuseValues) is intentional: Scale
// must not pick up new chart defaults (e.g. image) just because the chart
// on disk changed. Behavior by argument:
//
//   - replicas >= 0: every workload is forced to this exact value. Used
//     by suspending_app to scale to 0, and by upgrading_app to land an
//     upgrade-from-Stopped release back at 0 after the helm upgrade
//     applied the manifest-declared (non-zero) defaults.
//   - replicas <  0: every workload uses its manifest-declared value
//     (falls back to 1 for any name that is not listed). Used by
//     installing_app / resuming_app to bring the app up after the
//     initial replicas=0 install or after a stop.
//
// Apps without a WorkloadReplicas declaration are a no-op: there is
// nothing for Scale to override and callers (suspending_app /
// resuming_app) already branch on appcfg.HasWorkloadReplicas before
// reaching here. We treat it as a successful no-op rather than an
// error so the upgrade-from-Stopped tail in upgrading_app stays
// version-agnostic.
//
// Scale is unsupported for v2 apps; HelmOpsV2 overrides this to return
// an explicit error and callers should branch on appcfg.IsV2() first.
func (h *HelmOps) Scale(replicas int32) error {
	if !h.app.HasWorkloadReplicas() {
		klog.V(4).Infof("Scale no-op for app=%s (no workloadReplicas declared)", h.app.AppName)
		return nil
	}
	values := make(map[string]interface{})
	values["workloads"] = buildWorkloadsValues(h.app.WorkloadReplicas, replicas)

	if err := helm.UpgradeCharts(h.ctx, h.actionConfig, h.settings,
		h.app.AppName, h.app.ChartsName, h.app.RepoURL, h.app.Namespace,
		values, helm.ReuseValues); err != nil {
		klog.Errorf("Failed to scale chart appName=%s replicas=%d err=%v", h.app.AppName, replicas, err)
		return err
	}
	klog.Infof("Scaled app=%s to replicas=%d", h.app.AppName, replicas)
	return nil
}
