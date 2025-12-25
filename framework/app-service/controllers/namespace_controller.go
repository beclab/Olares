/*
Copyright 2019 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	"github.com/beclab/Olares/framework/app-service/pkg/utils/sliceutil"

	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

const (
	controllerNamespaceName = "namespace-controller"
)

// NamespaceReconciler reconciles a Namespace object
type NamespaceReconciler struct {
	client.Client
}

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerNamespaceName).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 1,
		}).
		For(&corev1.Namespace{}).
		Complete(r)
}

func (r *NamespaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	rootCtx := context.Background()
	namespace := &corev1.Namespace{}
	if err := r.Get(rootCtx, req.NamespacedName, namespace); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// name of your custom finalizer
	//finalizer := "finalizers.kubesphere.io/namespaces"

	if namespace.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !sliceutil.HasString(namespace.ObjectMeta.Finalizers, namespaceFinalizer) {
			// create only once, ignore already exists error
			if err := r.initCreatorRoleBinding(rootCtx, namespace); err != nil {
				return ctrl.Result{}, err
			}
			namespace.ObjectMeta.Finalizers = append(namespace.ObjectMeta.Finalizers, namespaceFinalizer)
			if namespace.Labels == nil {
				namespace.Labels = make(map[string]string)
			}
			// used for NetworkPolicyPeer.NamespaceSelector
			namespace.Labels["bytetrade.io/namespace"] = namespace.Name
			if err := r.Update(rootCtx, namespace); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if sliceutil.HasString(namespace.ObjectMeta.Finalizers, namespaceFinalizer) {
			// remove our finalizer from the list and update it.
			namespace.ObjectMeta.Finalizers = sliceutil.RemoveString(namespace.ObjectMeta.Finalizers, func(item string) bool {
				return item == namespaceFinalizer
			})
			if err := r.Update(rootCtx, namespace); err != nil {
				return ctrl.Result{}, err
			}
		}
		// Our finalizer has finished, so the reconciler can do nothing.
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

func (r *NamespaceReconciler) initCreatorRoleBinding(ctx context.Context, namespace *corev1.Namespace) error {
	creator := namespace.Annotations[creator]
	if creator == "" {
		return nil
	}
	var user iamv1alpha2.User
	if err := r.Get(ctx, types.NamespacedName{Name: creator}, &user); err != nil {
		return client.IgnoreNotFound(err)
	}
	creatorRoleBinding := newCreatorRoleBinding(creator, namespace.Name)
	if err := r.Client.Create(ctx, creatorRoleBinding); err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}
	return nil
}

func newCreatorRoleBinding(creator string, namespace string) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", creator, iamv1alpha2.NamespaceAdmin),
			Labels:    map[string]string{iamv1alpha2.UserReferenceLabel: creator},
			Namespace: namespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     iamv1alpha2.ResourceKindRole,
			Name:     iamv1alpha2.NamespaceAdmin,
		},
		Subjects: []rbacv1.Subject{
			{
				Name:     creator,
				Kind:     iamv1alpha2.ResourceKindUser,
				APIGroup: rbacv1.GroupName,
			},
		},
	}
}
