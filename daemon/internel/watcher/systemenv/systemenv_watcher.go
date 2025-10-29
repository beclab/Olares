package systemenv

import (
	"context"
	"sync"
	"time"

	"github.com/beclab/Olares/daemon/internel/watcher"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/containerd"
	"github.com/beclab/Olares/daemon/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"
)

var _ watcher.Watcher = &systemEnvWatcher{}

type systemEnvWatcher struct {
	sync.Mutex
	watcher.Watcher
	running bool
	cancel  context.CancelFunc
}

func NewSystemEnvWatcher() watcher.Watcher {
	return &systemEnvWatcher{}
}

func (w *systemEnvWatcher) Watch(ctx context.Context) {
	w.Lock()
	defer w.Unlock()

	if state.CurrentState.TerminusState != state.TerminusRunning &&
		state.CurrentState.TerminusState != state.Upgrading &&
		state.CurrentState.TerminusState != state.Uninitialized &&
		state.CurrentState.TerminusState != state.Initializing &&
		state.CurrentState.TerminusState != state.InitializeFailed &&
		state.CurrentState.TerminusState != state.SystemError {
		if w.cancel != nil {
			w.cancel()
			w.cancel = nil
		}
		w.running = false
		klog.V(4).Info("systemenv watcher stopped: cluster not running")
		return
	}

	if w.running {
		return
	}

	execCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel
	w.running = true

	go func() {
		defer func() {
			w.running = false
			w.cancel = nil
			klog.V(4).Info("systemenv watcher exited")
		}()

		startSystemEnvWatch(execCtx, func(eventType watch.EventType, obj map[string]any) {
			klog.V(5).Infof("systemenv event: %s", eventType)

			if eventType != watch.Added && eventType != watch.Modified {
				return
			}

			envName, _ := obj["envName"].(string)
			if envName != "OLARES_SYSTEM_CDN_SERVICE" && envName != "OLARES_SYSTEM_REMOTE_SERVICE" && envName != "OLARES_SYSTEM_DOCKERHUB_SERVICE" {
				return
			}

			val, _ := obj["value"].(string)
			if val == "" {
				if def, ok := obj["default"].(string); ok && def != "" {
					val = def
				}
			}
			if val == "" {
				return
			}

			switch envName {
			case "OLARES_SYSTEM_CDN_SERVICE":
				if commands.OLARES_CDN_SERVICE != val {
					old := commands.OLARES_CDN_SERVICE
					commands.OLARES_CDN_SERVICE = val
					klog.Infof("updated OLARES_CDN_SERVICE: %s -> %s", old, val)
				}
			case "OLARES_SYSTEM_REMOTE_SERVICE":
				if commands.OLARES_REMOTE_SERVICE != val {
					old := commands.OLARES_REMOTE_SERVICE
					commands.OLARES_REMOTE_SERVICE = val
					klog.Infof("updated OLARES_REMOTE_SERVICE: %s -> %s", old, val)
				}
			case "OLARES_SYSTEM_DOCKERHUB_SERVICE":
				if val != "" {
					go func(endpoint string) {
						if updated, err := containerd.EnsureRegistryMirror(execCtx, containerd.DefaultRegistryName, endpoint); err != nil {
							klog.Errorf("failed to ensure docker.io mirror endpoint %s: %v", endpoint, err)
							return
						} else if updated {
							klog.Infof("ensured docker.io mirror endpoint: %s", endpoint)
						} else {
							klog.V(5).Infof("docker.io mirror endpoint already present: %s", endpoint)
						}
					}(val)
				}
			}
		})
	}()
}

var systemEnvGVR = schema.GroupVersionResource{
	Group:    "sys.bytetrade.io",
	Version:  "v1alpha1",
	Resource: "systemenvs",
}

func startSystemEnvWatch(ctx context.Context, handle func(watch.EventType, map[string]any)) {
	for {
		dc, err := utils.GetDynamicClient()
		if err != nil {
			klog.V(4).Infof("systemenv watcher: dynamic client not ready: %v", err)
			return
		}

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
			klog.V(4).Info("systemenv watch channel closed, restarting with fresh list")
			continue
		case <-ctx.Done():
			return
		}
	}
}
