package appstate

import (
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
)

// init wires the StateTransitions table into apputils.UpdateAppMgrStatus so
// every handler path that writes ApplicationManager.Status via the apiserver
// helpers (handler_installer_install, handler_suspend, handler_applyenv,
// handler_installer_upgrade, handler_installer_uninstall, handler_settings,
// handler_compute_resources, handler_middleware*, appenv_controller,
// pod_abnormal_suspend_app_controller, …) is held to the same transition
// invariant as pkg/appstate's own updateStatus.
//
// The indirection is deliberate: pkg/utils/app cannot import pkg/appstate
// because pkg/appstate already imports pkg/utils/app, so we register the
// guard at init time instead.
func init() {
	apputils.StateTransitionGuard = IsStateTransitionAllowed
}
