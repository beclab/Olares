package systemenv

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	"olares.com/backup-server/pkg/constant"
	"olares.com/backup-server/pkg/util/log"
	"olares.com/backup-server/pkg/watchers"
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
				log.Error("not systemenv resource, invalid obj")
				return false
			}
			m := *mPtr
			// SystemEnv inlines EnvVarSpec keys at top level
			if envName, ok := m["envName"].(string); ok {
				return envName == constant.EnvOlaresSystemRemoteService
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

func (s *Subscriber) Do(_ context.Context, obj interface{}, action watchers.Action) error {
	mPtr, ok := obj.(*map[string]interface{})
	if !ok || mPtr == nil {
		return fmt.Errorf("invalid object type")
	}
	m := *mPtr

	// effective value can be from value or default
	var newValue string
	v, ok := m["value"].(string)
	if ok && v != "" {
		newValue = v
	} else if d, ok := m["default"].(string); ok && d != "" {
		newValue = d
	}
	if newValue == "" {
		return nil
	}

	if constant.OlaresRemoteService == newValue {
		return nil
	}

	log.Infof("updating OlaresRemoteService from %s to %s", constant.OlaresRemoteService, newValue)
	constant.OlaresRemoteService = newValue

	if constant.OlaresRemoteService == "" {
		constant.OlaresRemoteService = constant.DefaultSyncServerURL
	}

	if err := constant.ReloadEnvDependantVars(); err != nil {
		return fmt.Errorf("reload env-dependent vars failed: %w", err)
	} else {
		log.Infof("remote space url: %s", constant.SyncServerURL)
	}
	return nil
}
