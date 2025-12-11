package reverse_proxy

import (
	"bytetrade.io/web3os/bfl/pkg/apis/settings/v1alpha1"
	"bytetrade.io/web3os/bfl/pkg/constants"
	"bytetrade.io/web3os/bfl/pkg/watchers"
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

var GVR = schema.GroupVersionResource{
	Group: "", Version: "v1", Resource: "configmaps",
}

type Subscriber struct {
	*watchers.Watchers
	configurator *v1alpha1.ReverseProxyConfigurator
}

func NewSubscriber(w *watchers.Watchers) (*Subscriber, error) {
	configurator, err := v1alpha1.NewReverseProxyConfigurator()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize reverse proxy configurator: %w", err)
	}
	return &Subscriber{
		Watchers:     w,
		configurator: configurator,
	}, nil
}

func (s *Subscriber) Handler() cache.ResourceEventHandler {
	handleFunc := func(obj interface{}) {
		s.Watchers.Enqueue(
			watchers.EnqueueObj{
				Subscribe: s,
				Obj:       obj,
			},
		)
	}
	return cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			cm, ok := obj.(*corev1.ConfigMap)
			if !ok {
				klog.Error("not configmap resource, invalid obj")
				return false
			}

			if cm.Namespace != constants.Namespace || cm.Name != constants.ReverseProxyConfigMapName {
				return false
			}

			return true
		},

		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: handleFunc,
			UpdateFunc: func(_, new interface{}) {
				handleFunc(new)
			},
			DeleteFunc: func(_ interface{}) {
			},
		},
	}
}

func (s *Subscriber) Do(ctx context.Context, _ interface{}, _ watchers.Action) error {
	klog.Infof("handling reverse proxy config event")

	if err := s.configurator.Configure(ctx); err != nil {
		return fmt.Errorf("failed to get reverse proxy config configmap: %w", err)
	}

	return nil
}
