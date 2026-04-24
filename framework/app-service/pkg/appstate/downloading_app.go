package appstate

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	appsv1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/images"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ OperationApp = &DownloadingApp{}
var _ PollableStatefulInProgressApp = &downloadingInProgressApp{}

type downloadingInProgressApp struct {
	*DownloadingApp
	*basePollableStatefulInProgressApp
}

func (r *downloadingInProgressApp) poll(ctx context.Context) error {
	return r.imageClient.PollDownloadProgress(ctx, r.manager)
}

func (r *downloadingInProgressApp) WaitAsync(ctx context.Context) {
	appFactory.waitForPolling(ctx, r, func(err error) {
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				updateErr := r.updateStatus(context.TODO(), r.manager, appsv1.DownloadFailed, nil, err.Error(), "")
				if updateErr != nil {
					klog.Errorf("update app manager %s to %s state failed %v", r.manager.Name, appsv1.DownloadFailed.String(), updateErr)
					return
				}
			}
			// if the download is finished with error, we should not update the status to installing
			return
		}

		// Check Kubernetes request resources before transitioning to Installing
		var appConfig *appcfg.ApplicationConfig
		if err := json.Unmarshal([]byte(r.manager.Spec.Config), &appConfig); err != nil {
			klog.Errorf("failed to unmarshal app config for %s: %v", r.manager.Spec.AppName, err)
			updateErr := r.updateStatus(context.TODO(), r.manager, appsv1.InstallFailed, nil, fmt.Sprintf("invalid app config: %v", err), "")
			if updateErr != nil {
				klog.Errorf("update app manager %s to %s state failed %v", r.manager.Name, appsv1.InstallFailed.String(), updateErr)
			}
			return
		}

		_, conditionType, checkErr := apputils.CheckAppK8sRequestResource(appConfig, r.manager.Spec.OpType)
		if checkErr != nil {
			klog.Errorf("k8s request resource check failed for app %s: %v", r.manager.Spec.AppName, checkErr)
			opRecord := makeRecord(r.manager, appsv1.InstallFailed, checkErr.Error())
			updateErr := r.updateStatus(context.TODO(), r.manager, appsv1.InstallFailed, opRecord, checkErr.Error(), string(conditionType))
			if updateErr != nil {
				klog.Errorf("update app manager %s to %s state failed %v", r.manager.Name, appsv1.InstallFailed.String(), updateErr)
			}

			return
		}

		updateErr := r.updateStatus(context.TODO(), r.manager, appsv1.Installing, nil, appsv1.Installing.String(), "")
		if updateErr != nil {
			klog.Errorf("update app manager %s to %s state failed %v", r.manager.Name, appsv1.Installing.String(), updateErr)
			return
		}

	})
}

// override
func (r *downloadingInProgressApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	// do nothing here
	// do not exec duplicately
	return nil, nil
}

type DownloadingApp struct {
	*baseOperationApp
	imageClient images.ImageManager
}

func NewDownloadingApp(c client.Client,
	manager *appsv1.ApplicationManager, ttl time.Duration) (StatefulApp, StateError) {
	// TODO: check app state

	//

	return appFactory.New(c, manager, ttl,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &DownloadingApp{
				baseOperationApp: &baseOperationApp{
					baseStatefulApp: &baseStatefulApp{
						client:  c,
						manager: manager,
					},
					ttl: ttl,
				},
				imageClient: images.NewImageManager(c),
			}
		})
}

func (p *DownloadingApp) Cancel(ctx context.Context) error {
	// only cancel the downloading operation when the app is timeout
	klog.Infof("call timeout downloadingApp cancel....")
	err := p.updateStatus(ctx, p.manager, appsv1.DownloadingCanceling, nil, constants.OperationCanceledByTerminusTpl, appsv1.DownloadingCanceling.String())
	if err != nil {
		klog.Errorf("update app manager name=%s to downloadingCanceling state failed %v", p.manager.Name, err)
		return err
	}
	return nil
}

func (p *DownloadingApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	err := p.exec(ctx)
	if err != nil {
		klog.Errorf("app %s downloading failed %v", p.manager.Spec.AppName, err)
		opRecord := makeRecord(p.manager, appsv1.DownloadFailed, fmt.Sprintf(constants.OperationFailedTpl, p.manager.Spec.OpType, err.Error()))
		updateErr := p.updateStatus(ctx, p.manager, appsv1.DownloadFailed, opRecord, err.Error(), appsv1.DownloadFailed.String())
		if updateErr != nil {
			klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, appsv1.DownloadFailed.String(), updateErr)
			err = errors.Wrapf(err, "update status failed %v", updateErr)
		}
		return nil, err
	}

	return &downloadingInProgressApp{
		DownloadingApp:                    p,
		basePollableStatefulInProgressApp: &basePollableStatefulInProgressApp{},
	}, nil
}

func (p *DownloadingApp) exec(ctx context.Context) error {
	var err error
	var appConfig *appcfg.ApplicationConfig
	kubeConfig, err := ctrl.GetConfig()
	if err != nil {
		klog.Errorf("get kube config failed %v", err)
		return err
	}
	err = json.Unmarshal([]byte(p.manager.Spec.Config), &appConfig)
	if err != nil {
		klog.Errorf("unmarshal to appConfig failed %v", err)
		return err
	}

	values, err := appinstaller.BuildBaseHelmValues(ctx, kubeConfig, appConfig, p.manager.Spec.AppOwner, true, false)
	if err != nil {
		klog.Errorf("build base helm values failed %v", err)
		return err
	}

	refs, err := GetRefsForImageManager(appConfig, values)
	if err != nil {
		klog.Errorf("get image refs from resources failed %v", err)
		return err
	}

	err = p.imageClient.Create(ctx, p.manager, refs)

	if err != nil {
		klog.Errorf("create imagemanager failed %v", err)
		return err
	}

	return nil
}
