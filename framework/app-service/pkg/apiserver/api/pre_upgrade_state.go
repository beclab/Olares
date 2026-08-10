package api

import (
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// SnapshotPreUpgradeState records the steady pre-op intent (Running or
// Stopped) on am.Annotations[AppPreUpgradeStateKey].
//
// Upgrade/ApplyEnv handlers and the AppEnv controller call this before
// flipping the AM into Upgrading/ApplyingEnv. upgrading_app and
// applying_env_app later read the annotation to decide whether to scale
// workloads back up or land in Stopped.
//
// Only Running/Stopped overwrite the annotation. Failed or transitional
// states (e.g. UpgradeFailed) keep an existing Running/Stopped value so a
// retry cannot forget that the app was Stopped and pull pods back up. If
// there is no usable prior value, default to Running (prefer starting over
// accidentally leaving a Running app stopped).
func SnapshotPreUpgradeState(annotations map[string]string, state appv1alpha1.ApplicationManagerState) {
	if annotations == nil {
		return
	}
	switch state {
	case appv1alpha1.Running, appv1alpha1.Stopped:
		annotations[AppPreUpgradeStateKey] = string(state)
		return
	}
	existing := annotations[AppPreUpgradeStateKey]
	if existing == string(appv1alpha1.Running) || existing == string(appv1alpha1.Stopped) {
		return
	}
	annotations[AppPreUpgradeStateKey] = string(appv1alpha1.Running)
}
