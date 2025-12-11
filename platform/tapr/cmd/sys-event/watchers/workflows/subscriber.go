package workflows

import (
	"context"
	"errors"
	"strings"

	"bytetrade.io/web3os/tapr/cmd/sys-event/watchers"
	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	aprclientset "bytetrade.io/web3os/tapr/pkg/generated/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

const (
	WorkflowNameLabel       = "workflows.app.bytetrade.io/name"
	WorkflowTitleAnnotation = "workflows.app.bytetrade.io/title"
)

type Subscriber struct {
	*watchers.Subscriber
	invoker       *watchers.CallbackInvoker
	aprClient     *aprclientset.Clientset
	dynamicClient *dynamic.DynamicClient
}

func (s *Subscriber) WithKubeConfig(config *rest.Config) *Subscriber {
	s.aprClient = aprclientset.NewForConfigOrDie(config)
	s.dynamicClient = dynamic.NewForConfigOrDie(config)
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

			if strings.HasPrefix(ns.Name, "user-space-") || strings.HasPrefix(ns.Name, "user-system-") {
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
	title := ns.Annotations[WorkflowTitleAnnotation]
	owner, ownerOK := ns.Labels["bytetrade.io/ns-owner"]
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

		if s.Notification != nil {
			if !ownerOK {
				owner, err = s.getWorkflowOwner(ctx, ns.Name)
				if err != nil {
					klog.Warning(err)
				}
			}

			if owner != "" {
				return s.Notification.Send(ctx, owner, "recommend "+title+" is installed",
					&watchers.EventPayload{
						Type: string(aprv1.RecommendInstall),
						Data: map[string]interface{}{
							"name": name,
						},
					},
				)
			}
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

		if s.Notification != nil {
			if !ownerOK {
				owner, err = s.getWorkflowOwner(ctx, ns.Name)
				if err != nil {
					klog.Warning(err)
				}
			}

			if owner != "" {
				return s.Notification.Send(ctx, owner, "recommend "+title+" is uninstalled",
					&watchers.EventPayload{
						Type: string(aprv1.RecommendUninstall),
						Data: map[string]interface{}{
							"name": name,
						},
					},
				)
			}
		}
	}

	return nil
}

func (s *Subscriber) getWorkflowOwner(ctx context.Context, namespace string) (string, error) {
	gvr := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "cronworkflows",
	}

	if workflows, err := s.dynamicClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{}); err == nil {
		for _, w := range workflows.Items {
			if w.GetLabels() == nil {
				continue
			}

			owner, ok := w.GetLabels()["workflows.app.bytetrade.io/owner"]
			if ok && owner != "" {
				return owner, nil
			}
		}
	}

	return "", errors.New("owner not found")
}
