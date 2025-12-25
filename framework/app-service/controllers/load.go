package controllers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	appv1alpha1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/appstate"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	appevent "github.com/beclab/Olares/framework/app-service/pkg/event"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

func LoadStatefulApp(ctx context.Context, appmgr *ApplicationManagerController, name string) (appstate.StatefulApp, appstate.StateError) {
	var am appv1alpha1.ApplicationManager
	err := appmgr.Get(ctx, types.NamespacedName{Name: name}, &am)
	if err != nil {
		if apierrors.IsNotFound(err) {

			var app appv1alpha1.Application
			if err = appmgr.Get(ctx, types.NamespacedName{Name: name}, &app); err == nil {
				klog.Infof("LoadStatefulApp: application manager %s not found, but application %s exists", name, app.Name)
				// If the application manager is not found, but the application exists,
				// we need force delete the application.
				return nil, appstate.NewErrorUnknownState(func() func(ctx context.Context) error {
					return func(ctx context.Context) error {
						go func() {
							delCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
							defer cancel()
							klog.Infof("LoadStatefulApp: force delete application %s", app.Name)
							if !apputils.IsProtectedNamespace(app.Spec.Namespace) {
								err := appmgr.Delete(delCtx,
									&corev1.Namespace{
										ObjectMeta: metav1.ObjectMeta{
											Name: app.Spec.Namespace,
										},
									})

								if err != nil {
									klog.Errorf("LoadStatefulApp: force delete application %s failed: %v", app.Name, err)
								} else {
									klog.Infof("LoadStatefulApp: force delete application %s successfully", app.Name)
								}
								opID := strconv.FormatInt(time.Now().Unix(), 10)
								appevent.PublishAppEventToQueue(utils.EventParams{
									Owner:      app.Spec.Owner,
									Name:       app.Spec.Name,
									OpType:     string(appv1alpha1.UninstallOp),
									OpID:       opID,
									State:      string(appv1alpha1.Uninstalling),
									RawAppName: app.Spec.RawAppName,
									Type:       "app",
									Title:      app.Spec.Settings["title"],
									Reason:     constants.AppForceUninstall,
									Message: func() string {
										if err != nil {
											return fmt.Sprintf("force delete application %s failed: %v", app.Name, err)
										}
										return fmt.Sprintf("force delete application %s successfully", app.Name)
									}(),
								})

								ticker := time.NewTicker(2 * time.Second)
								defer ticker.Stop()
								for {
									select {
									case <-delCtx.Done():
										return
									case <-ticker.C:
										var ns corev1.Namespace
										err = appmgr.Get(delCtx, types.NamespacedName{Name: app.Spec.Namespace}, &ns)
										klog.Infof("wait for namespace: %s to be deleted", app.Spec.Namespace)
										if apierrors.IsNotFound(err) {
											appevent.PublishAppEventToQueue(utils.EventParams{
												Owner:      app.Spec.Owner,
												Name:       app.Spec.Name,
												OpType:     string(appv1alpha1.UninstallOp),
												OpID:       opID,
												State:      string(appv1alpha1.Uninstalled),
												RawAppName: app.Spec.RawAppName,
												Type:       "app",
												Title:      app.Spec.Settings["title"],
												Reason:     constants.AppForceUninstalled,
												Message:    fmt.Sprintf("app %s was force uninstalled", app.Spec.Name),
											})
											return
										}
									}
								}

							}
						}()

						return nil
					}
				}, nil)
			}
			return nil, nil
		}
		return nil, appstate.NewStateError(err.Error())
	}

	klog.Infof("LoadStatefulApp name:%s, state: %v", am.Name, am.Status.State)

	retApp, serr := func() (appstate.StatefulApp, appstate.StateError) {
		switch am.Status.State {
		case appv1alpha1.Pending:
			return appstate.NewPendingApp(ctx, appmgr, &am, 24*time.Hour)
		case appv1alpha1.Downloading:
			return appstate.NewDownloadingApp(appmgr, &am, 24*time.Hour)
		case appv1alpha1.Installing:
			return appstate.NewInstallingApp(appmgr, &am, 30*time.Minute)
		case appv1alpha1.Initializing:
			return appstate.NewInitializingApp(appmgr, &am, 60*time.Minute)
		case appv1alpha1.Running:
			return appstate.NewRunningApp(ctx, appmgr, &am)
		case appv1alpha1.Stopping:
			return appstate.NewSuspendingApp(appmgr, &am, 30*time.Minute)
		case appv1alpha1.Upgrading:
			return appstate.NewUpgradingApp(appmgr, &am, 30*time.Minute)
		case appv1alpha1.ApplyingEnv:
			return appstate.NewApplyingEnvApp(appmgr, &am, 30*time.Minute)
		case appv1alpha1.Resuming:
			return appstate.NewResumingApp(appmgr, &am, 60*time.Minute)
		case appv1alpha1.PendingCanceling:
			return appstate.NewPendingCancelingApp(appmgr, &am)
		case appv1alpha1.DownloadingCanceling:
			return appstate.NewDownloadingCancelingApp(appmgr, &am)
		case appv1alpha1.InstallingCanceling:
			return appstate.NewInstallingCancelingApp(appmgr, &am, 30*time.Minute)
		case appv1alpha1.InitializingCanceling:
			return appstate.NewInitializingCancelingApp(appmgr, &am)
		case appv1alpha1.ResumingCanceling:
			return appstate.NewResumingCancelingApp(appmgr, &am)
		case appv1alpha1.UpgradingCanceling:
			return appstate.NewUpgradingCancelingApp(appmgr, &am)
		case appv1alpha1.ApplyingEnvCanceling:
			return appstate.NewApplyingEnvCancelingApp(appmgr, &am)
		case appv1alpha1.Uninstalling:
			return appstate.NewUninstallingApp(appmgr, &am, 15*time.Minute)
		case appv1alpha1.StopFailed:
			return appstate.NewSuspendFailedApp(appmgr, &am)
		case appv1alpha1.UninstallFailed:
			return appstate.NewUninstallFailedApp(appmgr, &am)
		case appv1alpha1.UpgradeFailed:
			return appstate.NewUpgradeFailedApp(appmgr, &am)
		case appv1alpha1.ApplyEnvFailed:
			return appstate.NewApplyEnvFailedApp(appmgr, &am)
		case appv1alpha1.ResumeFailed:
			return appstate.NewResumeFailedApp(appmgr, &am)

		case appv1alpha1.DownloadFailed,
			appv1alpha1.PendingCanceled, appv1alpha1.DownloadingCanceled,
			appv1alpha1.InstallingCanceled, appv1alpha1.InitializingCanceled,
			appv1alpha1.UpgradingCanceled, appv1alpha1.ApplyingEnvCanceled,
			appv1alpha1.ResumingCanceled, appv1alpha1.Stopped:
			return appstate.NewDoNothingApp(appmgr, &am)
		case appv1alpha1.InstallFailed:
			return appstate.NewInstallFailedApp(appmgr, &am)
		case appv1alpha1.PendingCancelFailed:
			return appstate.NewPendingCancelFailedApp(appmgr, &am)
		case appv1alpha1.DownloadingCancelFailed:
			return appstate.NewDownloadingCancelFailedApp(appmgr, &am)

		case appv1alpha1.InstallingCancelFailed:
			return appstate.NewInstallingCancelFailedApp(appmgr, &am)
		case appv1alpha1.UpgradingCancelFailed:
			return appstate.NewUpgradingCancelFailedApp(appmgr, &am)
		case appv1alpha1.ApplyingEnvCancelFailed:
			return appstate.NewApplyingEnvCancelFailedApp(appmgr, &am)
		case appv1alpha1.Uninstalled:
			return appstate.NewUninstalledApp(ctx, appmgr, &am)
		}

		return nil, appstate.NewErrorUnknownState(nil, nil)
	}()

	if serr != nil {
		klog.Infof("load stateful app name=%s, state=%s failed err %v", am.Name, am.Status.State, serr)
		return nil, serr
	}

	return retApp, nil
}
