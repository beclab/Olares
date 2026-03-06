package provider

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"bytetrade.io/web3os/bfl/internal/ingress/api/app.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/bfl/pkg/constants"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	svcReconcileDelay = 100 * time.Millisecond
	svcRetryDelay     = 2 * time.Second
)

type BflSvcReconciler struct {
	cache        ctrlcache.Cache
	client       client.Client
	synced       int32
	debounceCh   chan struct{}
	OnReconciled func() // called after successful reconciliation to trigger downstream re-publish
}

func NewBflSvcReconciler(c ctrlcache.Cache, cl client.Client) *BflSvcReconciler {
	return &BflSvcReconciler{
		cache:      c,
		client:     cl,
		debounceCh: make(chan struct{}, 1),
	}
}

func (r *BflSvcReconciler) Name() string { return "bflsvc-reconciler" }

func (r *BflSvcReconciler) SetupWithManager(ctx context.Context) error {
	appInformer, err := r.cache.GetInformer(ctx, &v1alpha1.Application{})
	if err != nil {
		return fmt.Errorf("get application informer: %w", err)
	}

	if _, err = appInformer.AddEventHandler(toolscache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			app, ok := obj.(*v1alpha1.Application)
			if !ok {
				return false
			}
			return app.Spec.Owner == constants.Username
		},
		Handler: toolscache.ResourceEventHandlerFuncs{
			AddFunc:    func(_ interface{}) { r.notifyChanged() },
			UpdateFunc: func(_, _ interface{}) { r.notifyChanged() },
			DeleteFunc: func(_ interface{}) { r.notifyChanged() },
		},
	}); err != nil {
		return fmt.Errorf("add bflsvc app event handler: %w", err)
	}

	klog.Info("bflsvc-reconciler: event handler registered")
	return nil
}

func (r *BflSvcReconciler) Start(ctx context.Context) error {
	atomic.StoreInt32(&r.synced, 1)
	klog.Info("bflsvc-reconciler: cache synced, running initial reconcile")
	if err := r.reconcile(ctx); err != nil {
		klog.Errorf("bflsvc-reconciler: initial reconcile failed, will retry: %v", err)
		r.scheduleRetry(ctx)
	}
	r.debounceLoop(ctx)
	klog.Info("bflsvc-reconciler: stopped")
	return nil
}

func (r *BflSvcReconciler) notifyChanged() {
	if atomic.LoadInt32(&r.synced) == 0 {
		return
	}
	select {
	case r.debounceCh <- struct{}{}:
	default:
	}
}

func (r *BflSvcReconciler) debounceLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-r.debounceCh:
			timer := time.NewTimer(svcReconcileDelay)
		drain:
			for {
				select {
				case <-r.debounceCh:
					if !timer.Stop() {
						select {
						case <-timer.C:
						default:
						}
					}
					timer.Reset(svcReconcileDelay)
				case <-timer.C:
					break drain
				case <-ctx.Done():
					timer.Stop()
					return
				}
			}
			if err := r.reconcile(ctx); err != nil {
				klog.Errorf("bflsvc-reconciler: reconcile failed, retrying in %v: %v", svcRetryDelay, err)
				r.scheduleRetry(ctx)
			}
		}
	}
}

func (r *BflSvcReconciler) scheduleRetry(ctx context.Context) {
	go func() {
		select {
		case <-time.After(svcRetryDelay):
			r.notifyChanged()
		case <-ctx.Done():
		}
	}()
}

func (r *BflSvcReconciler) reconcile(ctx context.Context) error {
	var appList v1alpha1.ApplicationList
	if err := r.cache.List(ctx, &appList, client.InNamespace("")); err != nil {
		return fmt.Errorf("list applications: %w", err)
	}

	var svc corev1.Service
	err := r.cache.Get(ctx, types.NamespacedName{Namespace: constants.Namespace, Name: constants.BFLServiceName}, &svc)
	if err != nil {
		return fmt.Errorf("get bfl svc failed: %w", err)
	}
	n := len(svc.Spec.Ports)

	for i := 0; i < n; {
		if strings.HasPrefix(svc.Spec.Ports[i].Name, "stream-tcp-") ||
			strings.HasPrefix(svc.Spec.Ports[i].Name, "stream-udp-") {
			svc.Spec.Ports = append(svc.Spec.Ports[:i], svc.Spec.Ports[i+1:]...)
			n--
		} else {
			i++
		}
	}
	for _, app := range appList.Items {
		if len(app.Spec.Ports) == 0 {
			continue
		}
		for _, p := range app.Spec.Ports {
			if p.Host == "" || p.ExposePort < 1 || p.ExposePort > 65535 {
				klog.Warningf("invalid port app:%s, host: %s,exportPort: %d,skipping", app.Spec.Name, p.Host, p.ExposePort)
				continue
			}
			svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
				Name:       fmt.Sprintf("stream-%s-%d", strings.ToLower(p.Protocol), p.ExposePort),
				Protocol:   corev1.Protocol(strings.ToUpper(p.Protocol)),
				Port:       p.ExposePort,
				TargetPort: intstr.FromInt(int(p.ExposePort)),
			})
		}
	}
	err = r.client.Update(ctx, &svc)
	if err != nil {
		return fmt.Errorf("update bfl svc failed: %w", err)
	}

	if r.OnReconciled != nil {
		r.OnReconciled()
	}
	return nil
}
