package login

import (
	"context"

	"bytetrade.io/web3os/tapr/cmd/sys-event/watchers"
	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

type Subscriber struct {
	*watchers.Subscriber
}

func (s *Subscriber) HandleEvent() cache.ResourceEventHandler {
	return cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			return true
		},

		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				login := obj.(*LoginRecord)
				if watchers.ValidWatchDuration(&login.ObjectMeta) {
					eobj := watchers.EnqueueObj{
						Subscribe: s,
						Obj:       obj,
						Action:    watchers.ADD,
					}
					s.Watchers.Enqueue(eobj)
				}
			},
		},
	}
}

func (s *Subscriber) Do(ctx context.Context, obj interface{}, action watchers.Action) error {
	if action == watchers.ADD && s.Notification != nil {
		admin, err := s.Notification.AdminUser(ctx)
		if err != nil {
			return err
		}

		login := obj.(*LoginRecord)
		user := login.Labels["iam.kubesphere.io/user-ref"]

		msg := user + " login from " + login.Spec.SourceIP
		if err := s.Notification.Send(ctx, admin, msg, &watchers.EventPayload{
			Type: string(aprv1.UserLogin),
			Data: map[string]interface{}{
				"user": user,
			},
		}); err != nil {
			return err
		}

		if user != admin {
			return s.Notification.Send(ctx, user, msg, &watchers.EventPayload{
				Type: string(aprv1.UserLogin),
				Data: map[string]interface{}{
					"user": user,
				},
			})
		}
	}

	return nil
}
