package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"bytetrade.io/web3os/osnode-init/pkg/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

type SystemEnvWatcher struct {
	cfg *rest.Config
}

func NewSystemEnvWatcher(cfg *rest.Config) *SystemEnvWatcher {
	return &SystemEnvWatcher{cfg: cfg}
}

func (w *SystemEnvWatcher) Start(ctx context.Context) error {
	dc, err := dynamic.NewForConfig(w.cfg)
	if err != nil {
		klog.Infof("systemenv watcher: dynamic client not ready: %v", err)
		return fmt.Errorf("systemenv watcher: dynamic client not ready: %v", err)
	}

	startSystemEnvWatch(ctx, dc, func(eventType watch.EventType, obj map[string]any) {
		if eventType != watch.Added && eventType != watch.Modified {
			return
		}

		envName, _ := obj["envName"].(string)
		val, _ := obj["value"].(string)
		defVal, _ := obj["default"].(string)

		if envName != "OLARES_SYSTEM_REMOTE_SERVICE" {
			return
		}

		if val == "" && defVal != "" {
			val = defVal
		}

		val = strings.TrimSpace(val)

		if val == "" {
			return
		}

		if constants.OlaresRemoteService != val {
			old := constants.OlaresRemoteService
			constants.OlaresRemoteService = val
			klog.Infof("updated OlaresRemoteService: %s -> %s", old, val)
		}
	})

	<-ctx.Done()
	return nil
}

var systemEnvGVR = schema.GroupVersionResource{
	Group:    "sys.bytetrade.io",
	Version:  "v1alpha1",
	Resource: "systemenvs",
}

func startSystemEnvWatch(ctx context.Context, dc dynamic.Interface, handle func(watch.EventType, map[string]any)) {
	for {
		list, err := dc.Resource(systemEnvGVR).List(ctx, metav1.ListOptions{})
		if err != nil {
			select {
			case <-time.After(3 * time.Second):
				klog.V(3).Infof("systemenv list failed, retrying: %v", err)
				continue
			case <-ctx.Done():
				return
			}
		}

		for i := range list.Items {
			handle(watch.Added, list.Items[i].UnstructuredContent())
		}

		rv := list.GetResourceVersion()
		w, err := dc.Resource(systemEnvGVR).Watch(ctx, metav1.ListOptions{ResourceVersion: rv})
		if err != nil {
			select {
			case <-time.After(3 * time.Second):
				klog.V(3).Infof("systemenv watch create failed, retrying: %v", err)
				continue
			case <-ctx.Done():
				return
			}
		}

		ch := w.ResultChan()
		for {
			select {
			case e, ok := <-ch:
				if !ok {
					w.Stop()
					goto REWATCH
				}
				if u, ok := e.Object.(interface{ UnstructuredContent() map[string]any }); ok {
					handle(e.Type, u.UnstructuredContent())
				}
			case <-ctx.Done():
				w.Stop()
				return
			}
		}

	REWATCH:
		select {
		case <-time.After(1 * time.Second):
			klog.Info("systemenv watch channel closed, restarting with fresh list")
			continue
		case <-ctx.Done():
			return
		}
	}
}
