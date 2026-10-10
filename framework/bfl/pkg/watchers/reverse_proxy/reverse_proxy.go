package reverse_proxy

import (
	"bytetrade.io/web3os/bfl/pkg/apis/settings/v1alpha1"
	"bytetrade.io/web3os/bfl/pkg/constants"
	"bytetrade.io/web3os/bfl/pkg/watchers"
	"context"
	"errors"
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
}

func NewSubscriber(w *watchers.Watchers) (*Subscriber, error) {
	return &Subscriber{
		Watchers: w,
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

	// A local user has no domain during onboarding. Resolve the current user
	// for each configuration event so a later domain binding can take effect.
	configurator, err := v1alpha1.NewReverseProxyConfigurator()
	if errors.Is(err, v1alpha1.ErrDomainNotBound) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := configurator.Configure(ctx); err != nil {
		return fmt.Errorf("failed to get reverse proxy config configmap: %w", err)
	}

	return nil
}
