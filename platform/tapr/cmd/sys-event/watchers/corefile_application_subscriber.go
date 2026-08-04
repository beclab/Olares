package watchers

import (
	"context"

	"bytetrade.io/web3os/tapr/pkg/app/application"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// CorefileApplicationSubscriber regenerates CoreDNS when Application shared
// entrances or the app-shared label change.
type CorefileApplicationSubscriber struct {
	*Subscriber
	kubeClient    kubernetes.Interface
	dynamicClient dynamic.Interface
}

func (s *CorefileApplicationSubscriber) HandleEvent() cache.ResourceEventHandler {
	enqueue := func(obj interface{}) {
		s.Watchers.Enqueue(EnqueueObj{
			Subscribe: s,
			Obj:       obj,
			Action:    UPDATE,
		})
	}
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    enqueue,
		UpdateFunc: func(_, newObj interface{}) { enqueue(newObj) },
		DeleteFunc: enqueue,
	}
}

func (s *CorefileApplicationSubscriber) Do(ctx context.Context, obj interface{}, action Action) error {
	_ = obj
	_ = action
	return RegenerateCorefile(ctx, s.kubeClient, s.dynamicClient)
}

// RegisterCorefileApplicationWatcher watches Application CRs and triggers RegenerateCorefile.
func RegisterCorefileApplicationWatcher(w *Watchers, kubeClient kubernetes.Interface, dynamicClient dynamic.Interface) error {
	sub := &CorefileApplicationSubscriber{
		Subscriber:    NewSubscriber(w),
		kubeClient:    kubeClient,
		dynamicClient: dynamicClient,
	}
	return AddToWatchers[application.Application](w, application.GVR, sub.HandleEvent())
}
