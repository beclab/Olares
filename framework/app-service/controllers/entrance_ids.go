package controllers

import (
	"context"
	"fmt"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	sysv1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	"github.com/beclab/api/pkg/entranceids"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// +kubebuilder:rbac:groups=sys.bytetrade.io,resources=terminuses,verbs=get;list;watch
func (r *ApplicationReconciler) originalDomainUser(ctx context.Context) (string, error) {
	var term sysv1.Terminus
	if r.AppClientset != nil {
		t, err := r.AppClientset.SysV1alpha1().Terminuses().Get(ctx, "terminus", metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		term = *t
	} else if err := r.Get(ctx, types.NamespacedName{Name: "terminus"}, &term); err != nil {
		return "", err
	}
	if term.Spec.LocalName == "" {
		return "", fmt.Errorf("Terminus.spec.localName is required to allocate entrance IDs")
	}
	return term.Spec.LocalName, nil
}

func (r *ApplicationReconciler) assignOrdinaryIDs(ctx context.Context, appid, owner string, old, desired []appv1.Entrance) ([]appv1.Entrance, error) {
	if result, err := appv1.AssignOrdinaryEntranceIDs(appid, owner, "", old, desired); err == nil {
		return result, nil
	}
	original, err := r.originalDomainUser(ctx)
	if err != nil {
		return nil, err
	}
	return appv1.AssignOrdinaryEntranceIDs(appid, owner, original, old, desired)
}

func (r *ApplicationReconciler) assignApplicationIDs(ctx context.Context, desired, old *appv1.Application) (*appv1.Application, error) {
	if result, err := appv1.AssignApplicationEntranceIDs(desired, old, ""); err == nil {
		return result, nil
	}
	original, err := r.originalDomainUser(ctx)
	if err != nil {
		return nil, err
	}
	return appv1.AssignApplicationEntranceIDs(desired, old, original)
}

func (r *ApplicationReconciler) entranceWriteContext(ctx context.Context) (context.Context, func(), error) {
	if r.Kubeconfig == nil {
		return nil, nil, fmt.Errorf("Kubernetes config is required for entrance allocation")
	}
	kube, err := kubernetes.NewForConfig(r.Kubeconfig)
	if err != nil {
		return nil, nil, err
	}
	return entranceids.AcquireWriteLock(ctx, kube)
}

func (r *ApplicationReconciler) validateEntranceClaims(ctx context.Context, app *appv1.Application) error {
	list, err := r.AppClientset.AppV1alpha1().Applications().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	return appv1.ValidateOrdinaryIDClaims(app, list.Items)
}
