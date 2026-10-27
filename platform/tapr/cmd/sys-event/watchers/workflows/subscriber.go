package workflows

import (
	"context"
	"strings"

	"bytetrade.io/web3os/tapr/cmd/sys-event/watchers"
	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	aprclientset "bytetrade.io/web3os/tapr/pkg/generated/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

const WorkflowNameLabel = "workflows.app.bytetrade.io/name"

type Subscriber struct {
	*watchers.Subscriber
	invoker   *watchers.CallbackInvoker
	aprClient *aprclientset.Clientset
}

func (s *Subscriber) WithKubeConfig(config *rest.Config) *Subscriber {
	s.aprClient = aprclientset.NewForConfigOrDie(config)
	s.invoker = &watchers.CallbackInvoker{
		AprClient: s.aprClient,
		Retriable: func(err error) bool { return true },
	}
	return s
}

func (s *Subscriber) HandleEvent() cache.ResourceEventHandler {
	return cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			ns, ok := obj.(*corev1.Namespace)
			if !ok {
				klog.Error("not namespace resource, invalid obj")
				return false
			}

			if strings.HasPrefix(ns.Name, "user-space-") {
				return false
			}

			return true
		},

		Handler: cache.ResourceEventHandlerFuncs{

			DeleteFunc: func(obj interface{}) {
				if _, ok := obj.(*corev1.Namespace).Labels[WorkflowNameLabel]; ok {
					eobj := watchers.EnqueueObj{
						Subscribe: s,
						Obj:       obj,
						Action:    watchers.DELETE,
					}
					s.Watchers.Enqueue(eobj)
				}
			},

			UpdateFunc: func(oldObj, newObj interface{}) {
				oldNs := oldObj.(*corev1.Namespace)
				newNs := newObj.(*corev1.Namespace)

				if _, ok := newNs.Labels[WorkflowNameLabel]; ok {
					if _, ok := oldNs.Labels[WorkflowNameLabel]; !ok {
						eobj := watchers.EnqueueObj{
							Subscribe: s,
							Obj:       newObj,
							Action:    watchers.ADD,
						}
						s.Watchers.Enqueue(eobj)
					}
				}
			},
		},
	}
}

func (s *Subscriber) Do(ctx context.Context, obj interface{}, action watchers.Action) error {
	ns := obj.(*corev1.Namespace)
	name := ns.Labels[WorkflowNameLabel]
	switch action {
	case watchers.ADD:
		klog.Info("recommend ", ns, "/", name, " is installed")
		err := s.invoker.Invoke(ctx,
			func(cb *aprv1.SysEventRegistry) bool {
				return cb.Spec.Type == aprv1.Subscriber && cb.Spec.Event == aprv1.RecommendInstall
			},
			map[string]interface{}{
				"name": name,
			},
		)

		if err != nil {
			klog.Warning(err)
		}

	case watchers.DELETE:
		klog.Info("recommend ", ns, "/", name, " is uninstalled")
		err := s.invoker.Invoke(ctx,
			func(cb *aprv1.SysEventRegistry) bool {
				return cb.Spec.Type == aprv1.Subscriber && cb.Spec.Event == aprv1.RecommendUninstall
			},
			map[string]interface{}{
				"name": name,
			},
		)

		if err != nil {
			klog.Warning(err)
		}

	}

	return nil
}
