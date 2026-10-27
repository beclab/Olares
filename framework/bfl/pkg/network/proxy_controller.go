package network

import (
	"context"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

// ProxyController keeps the agent current even while domain provisioning is
// waiting on a slow or unavailable cloud certificate service.
type ProxyController struct{ *Controller }

func (c *ProxyController) SetupWithManager(mgr ctrl.Manager) error {
	return setupNetworkController(mgr, "default-reverse-proxy", c, false)
}

func (c *ProxyController) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	binding, err := c.Store.Binding(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	_, spec, _, err := c.Store.Config(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	status := ProxyStatus{ObservedRevision: spec.Revision}
	var jws string
	if binding.Domain == "" && spec.Config.Mode != ModeNone {
		status.Phase, status.Message = "waiting_domain", "Bind the default domain to continue"
	} else {
		if spec.Config.Mode == ModeFRP && spec.Config.FRPAuthMethod == "jws" {
			jws, err = c.Store.AuthToken(ctx, "jws")
			if err == nil && jws == "" {
				jws, err = c.bindingJWS(ctx, binding)
			}
		}
		if err == nil {
			status, err = c.ensureProxy(ctx, binding, spec, jws)
		}
		if err != nil {
			status.Phase, status.Message = "failed", err.Error()
		}
	}
	if saveErr := c.Store.SaveProxyStatus(ctx, spec.Revision, status); saveErr != nil {
		return ctrl.Result{}, saveErr
	}
	if err != nil {
		return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
	}
	return ctrl.Result{RequeueAfter: time.Minute}, nil
}
