package apps

import (
	"context"
	"strings"

	"bytetrade.io/web3os/tapr/cmd/sys-event/watchers"
	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
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

			if strings.HasPrefix(app.Spec.Namespace, "user-space-") || strings.HasPrefix(app.Spec.Namespace, "user-system-") {
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
		if s.Notification != nil {
			return s.Notification.Send(ctx, app.Spec.Owner, "app "+app.Spec.Namespace+"/"+app.Spec.Name+" is installed",
				&watchers.EventPayload{
					Type: string(aprv1.AppInstall),
					Data: map[string]interface{}{
						"name": app.Spec.Name,
					},
				},
			)
		}
	case watchers.DELETE:
		klog.Info("app ", app.Spec.Namespace, "/", app.Spec.Name, " is uninstalled")
		if s.Notification != nil {
			return s.Notification.Send(ctx, app.Spec.Owner, "app "+app.Spec.Namespace+"/"+app.Spec.Name+" is uninstalled",
				&watchers.EventPayload{
					Type: string(aprv1.AppUninstall),
					Data: map[string]interface{}{
						"name": app.Spec.Name,
					},
				},
			)
		}
	case watchers.SUSPEND:
		klog.Info("app ", app.Spec.Namespace, "/", app.Spec.Name, " is suspended")
		if s.Notification != nil {
			admin, err := s.Notification.AdminUser(ctx)
			if err != nil {
				return err
			}

			return s.Notification.Send(ctx, admin,
				app.Spec.Owner+"'s app "+app.Spec.Namespace+"/"+app.Spec.Name+" was suspended, cause: "+app.Annotations[suspendCauseAnnotation],
				&watchers.EventPayload{
					Type: string(aprv1.AppSuspend),
					Data: map[string]interface{}{
						"name": app.Spec.Name,
					},
				},
			)
		}
	}
	return nil
}
