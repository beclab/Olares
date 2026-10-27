package apps

import (
	"context"
	"strings"

	"bytetrade.io/web3os/tapr/cmd/sys-event/watchers"
	"bytetrade.io/web3os/tapr/pkg/app/application"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

type Subscriber struct {
	*watchers.Subscriber
}

const suspendAnnotation = "bytetrade.io/suspend-by"
const suspendCauseAnnotation = "bytetrade.io/suspend-cause"

func (s *Subscriber) HandleEvent() cache.ResourceEventHandler {
	return cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			app, ok := obj.(*application.Application)
			if !ok {
				klog.Error("not application resource, invalid obj")
				return false
			}

			if strings.HasPrefix(app.Spec.Namespace, "user-space-") {
				return false
			}

			return true
		},

		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				app := obj.(*application.Application)
				if watchers.ValidWatchDuration(&app.ObjectMeta) {
					eobj := watchers.EnqueueObj{
						Subscribe: s,
						Obj:       obj,
						Action:    watchers.ADD,
					}
					s.Watchers.Enqueue(eobj)
				}
			},
			DeleteFunc: func(obj interface{}) {
				eobj := watchers.EnqueueObj{
					Subscribe: s,
					Obj:       obj,
					Action:    watchers.DELETE,
				}
				s.Watchers.Enqueue(eobj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				oldApp := oldObj.(*application.Application)
				newApp := newObj.(*application.Application)

				if a, ok := newApp.Annotations[suspendAnnotation]; ok && a != oldApp.Annotations[suspendAnnotation] {
					eobj := watchers.EnqueueObj{
						Subscribe: s,
						Obj:       newObj,
						Action:    watchers.SUSPEND,
					}
					s.Watchers.Enqueue(eobj)
				}
			},
		},
	}
}

func (s *Subscriber) Do(ctx context.Context, obj interface{}, action watchers.Action) error {
	app := obj.(*application.Application)
	switch action {
	case watchers.ADD:
		klog.Info("app ", app.Spec.Namespace, "/", app.Spec.Name, " is installed")
	case watchers.DELETE:
		klog.Info("app ", app.Spec.Namespace, "/", app.Spec.Name, " is uninstalled")
	case watchers.SUSPEND:
		klog.Info("app ", app.Spec.Namespace, "/", app.Spec.Name, " is suspended")
	}
	return nil
}
