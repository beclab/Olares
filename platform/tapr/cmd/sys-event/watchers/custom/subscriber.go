package custom

import (
	"context"

	"bytetrade.io/web3os/tapr/cmd/sys-event/watchers"
	"k8s.io/klog/v2"
)

type Subscriber struct {
	notification *watchers.Notification
}

func (s *Subscriber) WithNotification(n *watchers.Notification) *Subscriber {
	s.notification = n
	return s
}

func (s *Subscriber) Do(ctx context.Context, obj interface{}, action watchers.Action) error {
	event := obj.(*CustomEvent)
	switch action {
	case watchers.ADD:
		klog.Info("user ", event.User, " fire event ", event.Type, ", ", event.Message)
		if s.notification != nil {
			return s.notification.Send(ctx, event.User, event.Message, &watchers.EventPayload{
				Type: event.Type,
				Data: event.Data,
			})
		}
	}
	return nil
}
