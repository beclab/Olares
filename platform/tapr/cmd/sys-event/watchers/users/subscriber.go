package users

import (
	"context"

	"bytetrade.io/web3os/tapr/cmd/sys-event/watchers"
	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	"bytetrade.io/web3os/tapr/pkg/kubesphere"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type task interface {
	doAdd(context.Context, *kubesphere.User) error
	doDelete(context.Context, *kubesphere.User) error
	doUpdate(context.Context, *kubesphere.User) error
}

var _ task = &Notify{}

// ////////////////////////////////
// notification task
type Notify struct {
	notification *watchers.Notification
}

// doAdd implements task.
func (n *Notify) doAdd(ctx context.Context, user *kubesphere.User) error {
	admin, err := n.notification.AdminUser(ctx)
	if err != nil {
		return err
	}

	if n.notification != nil {
		return n.notification.Send(ctx, admin, "user "+user.Name+" is created", &watchers.EventPayload{
			Type: string(aprv1.UserCreate),
			Data: map[string]interface{}{
				"user": user.Name,
			},
		})
	}

	return nil
}

// doDelete implements task.
func (n *Notify) doDelete(ctx context.Context, user *kubesphere.User) error {
	admin, err := n.notification.AdminUser(ctx)
	if err != nil {
		return err
	}

	if n.notification != nil {
		return n.notification.Send(ctx, admin, "user "+user.Name+" is deleted", &watchers.EventPayload{
			Type: string(aprv1.UserDelete),
			Data: map[string]interface{}{
				"user": user.Name,
			},
		})
	}

	return nil
}

// doUpdate implements task.
func (n *Notify) doUpdate(ctx context.Context, user *kubesphere.User) error { return nil }

// ////////////////////////////////
// update coredns task
var _ task = &UserDomain{}

type UserDomain struct {
	kubeClient    kubernetes.Interface
	dynamicClient dynamic.Interface
}

// doAdd implements task.
func (u *UserDomain) doAdd(context.Context, *kubesphere.User) error { return nil }

// doDelete implements task.
func (u *UserDomain) doDelete(ctx context.Context, user *kubesphere.User) error {
	return watchers.RegenerateCorefile(ctx, u.kubeClient, u.dynamicClient)
}

// doUpdate implements task.
func (u *UserDomain) doUpdate(ctx context.Context, user *kubesphere.User) error {
	return watchers.RegenerateCorefile(ctx, u.kubeClient, u.dynamicClient)
}

type Subscriber struct {
	tasks []task
}

func (s *Subscriber) Do(ctx context.Context, obj interface{}, action watchers.Action) error {
	user := obj.(*kubesphere.User)
	switch action {
	case watchers.ADD:
		klog.Info("user ", user.Name, " is created")
		for _, t := range s.tasks {
			if err := t.doAdd(ctx, user); err != nil {
				return err
			}
		}
	case watchers.DELETE:
		klog.Info("user ", user.Name, " is deleted")
		for _, t := range s.tasks {
			if err := t.doDelete(ctx, user); err != nil {
				return err
			}
		}
	case watchers.UPDATE:
		klog.Info("user ", user.Name, " is updated")
		for _, t := range s.tasks {
			if err := t.doUpdate(ctx, user); err != nil {
				return err
			}
		}
	}
	return nil
}
