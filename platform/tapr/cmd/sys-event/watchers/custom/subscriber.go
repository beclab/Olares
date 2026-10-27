package custom

import (
	"context"

	"bytetrade.io/web3os/tapr/cmd/sys-event/watchers"
	"k8s.io/klog/v2"
)

type Subscriber struct{}

func (s *Subscriber) Do(ctx context.Context, obj interface{}, action watchers.Action) error {
	event := obj.(*CustomEvent)
	switch action {
	case watchers.ADD:
		klog.Info("user ", event.User, " fire event ", event.Type, ", ", event.Message)
	}
	return nil
}
