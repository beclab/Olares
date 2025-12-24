package userspace

import (
	"context"
	"errors"
	"fmt"
	"time"

	appv1alpha1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/helm"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace/templates"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"

	"helm.sh/helm/v3/pkg/storage/driver"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	checkLauncherRunning  = "Running"
	checkLauncherNoExists = "NotExists"
)

type Deleter struct {
	client.Client              // k8s session client
	k8sConfig     *rest.Config // k8s service account config
	user          string
	helmCfg       helm.Config
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

	userspaceName := utils.UserspaceName(d.user)

	actionCfg, settings, err := helm.InitConfig(d.k8sConfig, userspaceName)

	if err != nil {
		return err
	}
	d.helmCfg.ActionCfg = actionCfg
	d.helmCfg.Settings = settings

	err = d.uninstallSysApps(ctx)
	if err != nil {
		return err
	}
	err = d.clearLauncher(fmt.Sprintf("launcher-%s", d.user))
	if err != nil {
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

	userSystem := templates.NewUserSystem(d.user)

	err = d.Client.Get(ctx, types.NamespacedName{Name: userSystem.Name}, &ns)
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("failed to get user-system namespace: %s, %v", userSystem.Name, err)
		return err
	}
	if err == nil {
		err = d.Client.Delete(ctx, &ns)
		if err != nil && !apierrors.IsNotFound(err) {
			klog.Errorf("failed to delete user-system namespace: %s, %v", userSystem.Name, err)
			return err
		}
	}

	return nil
}

func (d *Deleter) clearLauncher(launchername string) error {

	err := helm.UninstallCharts(d.helmCfg.ActionCfg, launchername)
	if err != nil && !errors.Is(err, driver.ErrReleaseNotFound) {
		klog.Errorf("failed to uninstall %s", launchername)
		return err
	}

	return nil
}

func (d *Deleter) findPVC(ctx context.Context, userspace string) (
	res *struct{ userspacePvc, userspacePv, appCachePvc, appCacheHostPath, dbdataPvc, dbdataHostPath string },
	err error) {
	var sts appsv1.StatefulSet
	err = d.Get(ctx, types.NamespacedName{Name: "bfl", Namespace: userspace}, &sts)
	if err != nil {
		klog.Errorf("Failed to get sts bfl userspace=%s err=%v", userspace, err)
		return nil, err
	}

	var ok bool
	res = &struct{ userspacePvc, userspacePv, appCachePvc, appCacheHostPath, dbdataPvc, dbdataHostPath string }{}
	res.userspacePvc, ok = sts.Annotations["userspace_pvc"]
	if !ok {
		return nil, errors.New("userspace PVC not found")
	}

	res.userspacePv, ok = sts.Annotations["userspace_pv"]
	if !ok {
		return nil, errors.New("userspace PV not found")
	}

	res.appCachePvc, ok = sts.Annotations["appcache_pvc"]
	if !ok {
		return nil, errors.New("appcache PVC not found")
	}

	res.appCacheHostPath, ok = sts.Annotations["appcache_hostpath"]
	if !ok {
		return nil, errors.New("appcache PVC not found")
	}

	res.dbdataPvc, ok = sts.Annotations["dbdata_pvc"]
	if !ok {
		return nil, errors.New("dbdata PVC not found")
	}

	res.dbdataHostPath, ok = sts.Annotations["dbdata_hostpath"]
	if !ok {
		return nil, errors.New("dbdata PVC not found")
	}

	return res, nil
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

func (d *Deleter) uninstallSysApps(ctx context.Context) error {
	sysApps, err := userspace.GetAppsFromDirectory(constants.UserChartsPath + "/apps")
	if err != nil {
		return err
	}

	//var errsCount int

	for _, appname := range sysApps {
		appReleaseName := helm.ReleaseName(appname, d.user)
		err = helm.UninstallCharts(d.helmCfg.ActionCfg, appReleaseName)
		if err != nil && !errors.Is(err, driver.ErrReleaseNotFound) {
			klog.Errorf("Failed to uninstall chart user=%s appName=%s err=%v", d.user, appname, err)
			return fmt.Errorf("failed to uninstall chart user=%s app=%s err=%w", d.user, appname, err)
		}

	}
	var appmgrs appv1alpha1.ApplicationManagerList

	err = d.List(ctx, &appmgrs)
	if err != nil {
		return err
	}

	for _, am := range appmgrs.Items {
		if am.Spec.AppOwner == d.user {
			var appmgr appv1alpha1.ApplicationManager
			err = d.Get(ctx, types.NamespacedName{Name: am.Name}, &appmgr)
			if err != nil && !apierrors.IsNotFound(err) {
				klog.Errorf("failed to get appmgr %s %v", am.Name, err)
				return err
			}
			if err == nil {
				e := d.Delete(ctx, &appmgr)
				if e != nil && !apierrors.IsNotFound(err) {
					klog.Errorf("Failed to delete user's sys applicationmanager user=%s err=%v", d.user, err)
					return e
				}
			}
		}
	}
	return nil
}

func (d *Deleter) checkLauncher(ctx context.Context, userspace string, runningOrExists string) (*corev1.Pod, error) {
	var (
		bfl          *corev1.Pod
		observations int
	)

	err := wait.PollImmediate(time.Second, 5*time.Minute, func() (bool, error) {
		var pods corev1.PodList
		selector, _ := labels.Parse("tier=bfl")

		err := d.List(ctx, &pods, &client.ListOptions{LabelSelector: selector, Namespace: userspace})
		if err != nil {
			return true, err
		}

		// check not exists
		if len(pods.Items) == 0 && runningOrExists == checkLauncherNoExists {
			return true, nil
		}

		if err == nil && len(pods.Items) > 0 {
			observations++
		}

		if observations >= 2 {
			if len(pods.Items) > 0 {
				pod := pods.Items[0]
				// check bfl is running, and return the bfl
				if pod.Status.Phase == corev1.PodRunning && runningOrExists == checkLauncherRunning {
					bfl = &pod
					return true, nil
				}
			}
		}
		return false, nil
	})

	return bfl, err
}

func (d *Deleter) isSysApps(app *appv1alpha1.Application) bool {
	userspace := utils.UserspaceName(d.user)
	return app.Spec.Namespace == userspace
}
