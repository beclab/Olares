package controllers

import (
	"context"
	"encoding/json"

	v1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// TailScaleACLController publishes every application's dynamic tailscale
// ports to the shared headscale in os-network. Subnet routes and the old
// per-user tailscale-acl ConfigMap are no longer written from here.
type TailScaleACLController struct {
	client.Client
	ports applicationPortsAPI
}

func (r *TailScaleACLController) SetUpWithManager(mgr ctrl.Manager) error {
	if r.ports == nil {
		r.ports = newHeadscalePolicyClient()
	}
	c, err := controller.New("headscale application ports controller", mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		return err
	}
	return c.Watch(source.Kind(
		mgr.GetCache(),
		&v1alpha1.Application{},
		handler.TypedEnqueueRequestsFromMapFunc(
			func(ctx context.Context, app *v1alpha1.Application) []reconcile.Request {
				return []reconcile.Request{{NamespacedName: types.NamespacedName{
					Name: applicationPortsReconcileName,
				}}}
			}),
		predicate.TypedFuncs[*v1alpha1.Application]{
			CreateFunc: func(e event.TypedCreateEvent[*v1alpha1.Application]) bool {
				return true
			},
			UpdateFunc: func(e event.TypedUpdateEvent[*v1alpha1.Application]) bool {
				return true
			},
			DeleteFunc: func(e event.TypedDeleteEvent[*v1alpha1.Application]) bool {
				return true
			},
		},
	))
}

func (r *TailScaleACLController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.Info("reconcile headscale application ports")
	var apps v1alpha1.ApplicationList
	if err := r.List(ctx, &apps); err != nil {
		klog.Errorf("list applications failed: %v", err)
		return ctrl.Result{}, err
	}

	klog.Infof("headscale application ports: listed %d applications", len(apps.Items))
	desired := applicationPortsFromApps(apps.Items)
	desiredRaw, err := marshalApplicationPorts(desired)
	if err != nil {
		return ctrl.Result{}, err
	}
	klog.Infof("headscale application ports desired (%d users): %s", len(desired), desiredRaw)

	current, err := r.ports.Get(ctx)
	if err != nil {
		klog.Errorf("get headscale application ports failed: %v", err)
		return ctrl.Result{}, err
	}
	currentRaw, err := marshalApplicationPorts(current)
	if err != nil {
		return ctrl.Result{}, err
	}
	klog.Infof("headscale application ports current (%d users): %s", len(current), currentRaw)
	if applicationPortsEqual(current, desired) {
		klog.Infof("headscale application ports unchanged, skip put")
		return ctrl.Result{}, nil
	}

	klog.Infof("replacing headscale application ports: %s", desiredRaw)
	if err := r.ports.Put(ctx, desired); err != nil {
		klog.Errorf("put headscale application ports failed: %v", err)
		return ctrl.Result{}, err
	}
	klog.Infof("replaced headscale application ports")
	return ctrl.Result{}, nil
}

func marshalApplicationPorts(ports []applicationPort) (string, error) {
	if ports == nil {
		ports = []applicationPort{}
	}
	raw, err := json.Marshal(struct {
		ApplicationPorts []applicationPort `json:"applicationPorts"`
	}{ApplicationPorts: ports})
	if err != nil {
		return "", err
	}
	return string(raw), nil
}
