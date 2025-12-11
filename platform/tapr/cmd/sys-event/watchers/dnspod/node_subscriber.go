package dnspod

import (
	"context"

	"bytetrade.io/web3os/tapr/cmd/sys-event/watchers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

type NodeSubscriber struct {
	*watchers.Subscriber
	kubeClient    kubernetes.Interface
	dynamicClient *dynamic.DynamicClient
}

func (s *NodeSubscriber) WithKubeConfig(config *rest.Config) *NodeSubscriber {
	s.dynamicClient = dynamic.NewForConfigOrDie(config)
	s.kubeClient = kubernetes.NewForConfigOrDie(config)
	return s
}

func (s *NodeSubscriber) HandleEvent() cache.ResourceEventHandler {
	return cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			node, ok := obj.(*corev1.Node)
			if !ok {
				klog.Error("not node resource, invalid obj")
				return false
			}

			_, ok = node.Labels["node-role.kubernetes.io/control-plane"]

			return ok
		},

		Handler: cache.ResourceEventHandlerFuncs{

			DeleteFunc: func(obj interface{}) {
				eobj := watchers.EnqueueObj{
					Subscribe: s,
					Obj:       obj,
					Action:    watchers.DELETE,
				}
				s.Watchers.Enqueue(eobj)
			},

			UpdateFunc: func(oldObj, newObj interface{}) {
				eobj := watchers.EnqueueObj{
					Subscribe: s,
					Obj:       newObj,
					Action:    watchers.ADD,
				}
				s.Watchers.Enqueue(eobj)
			},

			AddFunc: func(obj interface{}) {
				eobj := watchers.EnqueueObj{
					Subscribe: s,
					Obj:       obj,
					Action:    watchers.ADD,
				}
				s.Watchers.Enqueue(eobj)
			},
		},
	}
}

func (s *NodeSubscriber) Do(ctx context.Context, obj interface{}, action watchers.Action) error {
	return watchers.RegenerateCorefile(ctx, s.kubeClient, s.dynamicClient)
}
