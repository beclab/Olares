package systemenv

import (
	"context"
	"fmt"

	"bytetrade.io/web3os/bfl/pkg/constants"
	"bytetrade.io/web3os/bfl/pkg/watchers"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

var GVR = schema.GroupVersionResource{
	Group:    "sys.bytetrade.io",
	Version:  "v1alpha1",
	Resource: "systemenvs",
}

type Subscriber struct {
	*watchers.Watchers
}

func NewSubscriber(w *watchers.Watchers) *Subscriber {
	return &Subscriber{Watchers: w}
}

func (s *Subscriber) Handler() cache.ResourceEventHandler {
	handle := func(obj interface{}, action watchers.Action) {
		s.Watchers.Enqueue(watchers.EnqueueObj{Subscribe: s, Obj: obj, Action: action})
	}
	return cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			mPtr, ok := obj.(*map[string]interface{})
			if !ok || mPtr == nil {
				klog.Error("not systemenv resource, invalid obj")
				return false
			}
			m := *mPtr
			// SystemEnv inlines EnvVarSpec keys at top level
			if envName, ok := m["envName"].(string); ok {
				return envName == constants.EnvOlaresSystemRemoteService
			}
			return false
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) { handle(obj, watchers.ADD) },
			UpdateFunc: func(_, newObj interface{}) { handle(newObj, watchers.UPDATE) },
			DeleteFunc: func(obj interface{}) { handle(obj, watchers.DELETE) },
		},
	}
}

func (s *Subscriber) Do(_ context.Context, obj interface{}, _ watchers.Action) error {
	mPtr, ok := obj.(*map[string]interface{})
	if !ok || mPtr == nil {
		return fmt.Errorf("invalid object type")
	}
	m := *mPtr

	// effective value can be from value or default
	var newValue string
	if v, ok := m["value"].(string); ok && v != "" {
		newValue = v
	} else if d, ok := m["default"].(string); ok && d != "" {
		newValue = d
	}
	if newValue == "" {
		return nil
	}

	if constants.OlaresRemoteService == newValue {
		return nil
	}

	klog.Infof("updating OlaresRemoteService from %s to %s", constants.OlaresRemoteService, newValue)
	constants.OlaresRemoteService = newValue
	if err := constants.ReloadEnvDependantVars(); err != nil {
		return fmt.Errorf("reload env-dependent vars failed: %w", err)
	}
	return nil
}
