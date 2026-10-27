package userspace

import (
	"context"
	"errors"
	"fmt"
	"github.com/beclab/Olares/framework/app-service/pkg/helm"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"

	"helm.sh/helm/v3/pkg/storage/driver"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Deleter struct {
	client.Client              // k8s session client
	k8sConfig     *rest.Config // k8s service account config
	user          string
}

func NewDeleter(client client.Client, config *rest.Config, user string) *Deleter {
	return &Deleter{
		Client:    client,
		k8sConfig: config,
		user:      user,
	}
}

func (d *Deleter) DeleteUserResource(ctx context.Context) error {
	err := d.uninstallUserApps(ctx)
	if err != nil {
		klog.Infof("failed to uninstall user apps %v", err)
		return err
	}

	err = d.deleteNamespace(ctx)
	if err != nil {
		klog.Errorf("failed to delete namespace %v", err)
		return err
	}
	var globalRBD iamv1alpha2.GlobalRoleBinding
	err = d.Get(ctx, types.NamespacedName{Name: d.user}, &globalRBD)
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("failed to get globalrolebinding name: %s %v", d.user, err)
		return err
	}
	if err == nil {
		e := d.Delete(ctx, &globalRBD)
		if e != nil && !apierrors.IsNotFound(e) {
			klog.Errorf("failed to delete globalrolebinding name: %s %v", d.user, err)
			return e
		}
	}
	return nil
}

func (d *Deleter) deleteNamespace(ctx context.Context) error {
	userspaceNs := fmt.Sprintf("user-space-%s", d.user)
	var ns corev1.Namespace
	err := d.Client.Get(ctx, types.NamespacedName{Name: userspaceNs}, &ns)
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("failed to get user-space namespace: %s, %v", userspaceNs, err)
		return err
	}
	if err == nil {
		err = d.Client.Delete(ctx, &ns)
		if err != nil && !apierrors.IsNotFound(err) {
			klog.Errorf("failed to delete user-space namespace: %s, %v", userspaceNs, err)
			return err
		}
	}

	return nil
}

func (d *Deleter) uninstallUserApps(ctx context.Context) error {
	var appList appv1alpha1.ApplicationList
	err := d.List(ctx, &appList)
	if err != nil {
		return err
	}

	// filter by application's owner
	for _, a := range appList.Items {
		if a.Spec.Owner == d.user && !d.isSysApps(&a) {
			actionCfg, _, err := helm.InitConfig(d.k8sConfig, a.Spec.Namespace)
			if err != nil {
				klog.Errorf("Failed to delete user's application on config init, owner=%s err=%v", a.Spec.Owner, err)
				continue
			}

			err = helm.UninstallCharts(actionCfg, a.Spec.Name)
			if err != nil && !errors.Is(err, driver.ErrReleaseNotFound) {
				klog.Errorf("Failed to delete user's application owner=%s, err=%v", a.Spec.Owner, err)
			}
			name := fmt.Sprintf("%s-%s-%s", a.Spec.Name, a.Spec.Owner, a.Spec.Name)
			var am appv1alpha1.ApplicationManager
			err = d.Get(ctx, types.NamespacedName{Name: name}, &am)

			if err == nil {
				err = d.Delete(ctx, &am)
				if err != nil && !apierrors.IsNotFound(err) {
					klog.Errorf("Failed to delete user's applicationmanager name=%s owner=%s err=%v", name, a.Spec.Owner, err)
				}
			}

			var ns corev1.Namespace
			err = d.Get(ctx, types.NamespacedName{Name: a.Spec.Namespace}, &ns)
			if err == nil {
				err = d.Delete(ctx, &ns)
				if err != nil && !apierrors.IsNotFound(err) {
					klog.Errorf("Failed to delete user's namespace=%s err=%v", a.Spec.Namespace, err)
				}
			}
		}
	}

	return nil
}

func (d *Deleter) isSysApps(app *appv1alpha1.Application) bool {
	userspace := utils.UserspaceName(d.user)
	return app.Spec.Namespace == userspace
}
