package controllers

import (
	"context"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type TailScaleACLConfigMapController struct {
	client.Client
}

func (r *TailScaleACLConfigMapController) SetUpWithManager(mgr ctrl.Manager) error {
	c, err := controller.New("app's tailscale acls configmap manager controller", mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		return err
	}

	err = c.Watch(source.Kind(
		mgr.GetCache(),
		&corev1.ConfigMap{},
		handler.TypedEnqueueRequestsFromMapFunc(
			func(ctx context.Context, cm *corev1.ConfigMap) []reconcile.Request {
				return []reconcile.Request{{NamespacedName: types.NamespacedName{
					Name:      tailScaleACLConfigMapName,
					Namespace: cm.Namespace,
				}}}
			}),
		predicate.TypedFuncs[*corev1.ConfigMap]{
			CreateFunc: func(e event.TypedCreateEvent[*corev1.ConfigMap]) bool {
				return isTailScalAclConfigmap(e.Object)
			},
			UpdateFunc: func(e event.TypedUpdateEvent[*corev1.ConfigMap]) bool {
				return isTailScalAclConfigmap(e.ObjectNew)
			},
			DeleteFunc: func(e event.TypedDeleteEvent[*corev1.ConfigMap]) bool {
				return false
			},
		},
	))
	if err != nil {
		return err
	}
	return nil
}

func (r *TailScaleACLConfigMapController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	klog.Infof("reconcile tailscale acls configmap request name=%v, owner=%v", req.Name, req.Namespace)

	headScaleNamespace := req.Namespace

	// acl.json has changed, restart headscale via pod template annotation bump.
	deploy := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Namespace: headScaleNamespace, Name: "headscale"}, deploy)
	if err != nil {
		klog.Errorf("get headscale deployment in ns %s failed (skip rolling update): %v", headScaleNamespace, err)
		return ctrl.Result{}, err
	}

	if deploy.Spec.Template.Annotations == nil {
		deploy.Spec.Template.Annotations = make(map[string]string)
	}
	deploy.Spec.Template.Annotations[headScaleUpdatedTimeKey] = time.Now().String()
	err = r.Update(ctx, deploy)
	if err != nil {
		klog.Errorf("update headscale deploy failed: %v", err)
		return ctrl.Result{}, err
	}
	klog.Infof("rolling update headscale...")

	return ctrl.Result{}, nil
}

func isTailScalAclConfigmap(obj client.Object) bool {
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return false
	}
	namespace := cm.Namespace
	if !strings.HasPrefix(namespace, tailScaleNamespacePrefix) {
		return false
	}
	if cm.Name != tailScaleACLConfigMapName {
		return false
	}
	owner := strings.TrimPrefix(namespace, tailScaleNamespacePrefix)
	if owner == "" {
		return false
	}
	return true
}
