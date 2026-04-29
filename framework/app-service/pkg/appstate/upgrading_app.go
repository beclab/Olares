package appstate

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller/versioned"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/helm"
	"github.com/beclab/Olares/framework/app-service/pkg/images"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ OperationApp = &UpgradingApp{}

type UpgradingApp struct {
	*baseOperationApp
	imageClient          images.ImageManager
	finallyCh            chan func()
	isDownloading        bool
	isDownloaded         bool
	downloadTTL          time.Duration
	downloadedTime       *metav1.Time
	downloadingStartTime *metav1.Time
}

func (p *UpgradingApp) Finally() {
	if p.finallyCh == nil {
		return
	}
	if fn, ok := <-p.finallyCh; ok && fn != nil {
		fn()
	}
}

func (p *UpgradingApp) State() string {
	return p.GetManager().Status.State.String()
}

func NewUpgradingApp(c client.Client,
	manager *appsv1.ApplicationManager, downloadTTL, ttl time.Duration) (StatefulApp, StateError) {

	return appFactory.New(c, manager, ttl,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &UpgradingApp{
				baseOperationApp: &baseOperationApp{
					ttl: ttl,
					baseStatefulApp: &baseStatefulApp{
						manager: manager,
						client:  c,
					},
				},
				downloadTTL: downloadTTL,
				imageClient: images.NewImageManager(c),
			}
		})
}

func (p *UpgradingApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	p.finallyCh = make(chan func(), 1)

	opCtx, cancel := context.WithCancel(context.Background())
	return appFactory.execAndWatch(opCtx, p,
		func(c context.Context) (StatefulInProgressApp, error) {
			in := upgradingInProgressApp{
				UpgradingApp: p,
				baseStatefulInProgressApp: &baseStatefulInProgressApp{
					done:   c.Done,
					cancel: cancel,
				},
			}

			go func() {
				defer cancel()
				defer close(p.finallyCh)

				var execErr error
				defer func() {
					if r := recover(); r != nil {
						klog.Errorf("panic in upgrade exec goroutine: %v", r)
						execErr = fmt.Errorf("panic: %v", r)
					}
					if execErr != nil {
						p.finallyCh <- func() {
							klog.Info("upgrade app failed, update app status to upgradeFailed, ", p.manager.Name)
							opRecord := makeRecord(p.manager, appsv1.UpgradeFailed, fmt.Sprintf(constants.OperationFailedTpl, p.manager.Spec.OpType, execErr.Error()))

							updateErr := p.updateStatus(context.TODO(), p.manager, appsv1.UpgradeFailed, opRecord, execErr.Error(), appsv1.UpgradeFailed.String())
							if updateErr != nil {
								klog.Errorf("update appmgr state to upgradeFailed state failed %v", updateErr)
							}
						}
					} else {
						p.finallyCh <- func() {
							klog.Info("upgrade app success, update app status to initializing, ", p.manager.Name)
							updateErr := p.updateStatus(context.TODO(), p.manager, appsv1.Initializing, nil, appsv1.Initializing.String(), appsv1.Initializing.String())
							if updateErr != nil {
								klog.Errorf("update appmgr state to initializing state failed %v", updateErr)
							}
						}
					}
				}()

				execErr = p.exec(c)
			}()

			return &in, nil
		})
}

func (p *UpgradingApp) exec(ctx context.Context) error {
	var err error
	var version string
	var actionConfig *action.Configuration
	kubeConfig, err := ctrl.GetConfig()
	if err != nil {
		klog.Errorf("get kube config failed %v", err)
		return err
	}
	actionConfig, _, err = helm.InitConfig(kubeConfig, p.manager.Spec.AppNamespace)
	if err != nil {
		klog.Errorf("helm init config failed %v", err)
		return err
	}
	var appConfig *appcfg.ApplicationConfig
	deployedVersion, _, err := apputils.GetDeployedReleaseVersion(actionConfig, p.manager.Spec.AppName)
	if err != nil {
		klog.Errorf("Failed to get release revision err=%v", err)
		return err
	}

	if !utils.MatchVersion(version, ">= "+deployedVersion) {
		err = errors.New("upgrade version should great than deployed version")
		return err
	}

	annotations := p.manager.Annotations
	version = annotations[api.AppVersionKey]
	repoURL := annotations[api.AppRepoURLKey]
	token := annotations[api.AppTokenKey]
	marketSource := annotations[api.AppMarketSourceKey]
	//var chartPath string
	admin, err := kubesphere.GetAdminUsername(ctx, kubeConfig)
	if err != nil {
		klog.Errorf("get admin username failed %v", err)
		return err
	}
	isAdmin, err := kubesphere.IsAdmin(ctx, kubeConfig, p.manager.Spec.AppOwner)
	if err != nil {
		klog.Errorf("failed check is admin user %v", err)
		return err
	}
	getRawAppName := func(rawAppName string) string {
		if rawAppName == "" {
			return p.manager.Spec.AppName
		}
		return rawAppName
	}

	if isAdmin {
		admin = p.manager.Spec.AppOwner
	}

	if !userspace.IsSysApp(getRawAppName(p.manager.Spec.RawAppName)) {
		var cfg *appcfg.ApplicationConfig
		err = json.Unmarshal([]byte(p.manager.Spec.Config), &cfg)
		if err != nil {
			klog.Errorf("unmarshal to appConfig failed %v", err)
			return err
		}
		appConfig, _, err = apputils.GetAppConfig(ctx, &apputils.ConfigOptions{
			App:          p.manager.Spec.AppName,
			Owner:        p.manager.Spec.AppOwner,
			RawAppName:   getRawAppName(p.manager.Spec.RawAppName),
			RepoURL:      repoURL,
			Version:      version,
			Token:        token,
			Admin:        admin,
			MarketSource: marketSource,
			IsAdmin:      isAdmin,
			SelectedGpu:  cfg.SelectedGpuType,
		})

		if err != nil {
			klog.Errorf("get app config failed %v", err)
			return err
		}

		appConfig.Ports = cfg.Ports
		appConfig.TailScale = cfg.TailScale

	} else {
		_, err = apputils.GetIndexAndDownloadChart(ctx, &apputils.ConfigOptions{
			App:          p.manager.Spec.AppName,
			RawAppName:   getRawAppName(p.manager.Spec.RawAppName),
			RepoURL:      repoURL,
			Version:      version,
			Token:        token,
			Owner:        p.manager.Spec.AppOwner,
			MarketSource: marketSource,
		})

		if err != nil {
			klog.Errorf("download chart failed %v", err)
			return err
		}
		err = json.Unmarshal([]byte(p.manager.Spec.Config), &appConfig)
		if err != nil {
			klog.Errorf("unmarshal to appConfig failed %v", err)
			return err
		}
	}
	ops, err := versioned.NewHelmOps(ctx, kubeConfig, appConfig, token,
		appinstaller.Opt{Source: p.manager.Spec.Source, MarketSource: appcfg.GetMarketSource(p.manager)})
	if err != nil {
		klog.Errorf("make helmop failed %v", err)
		return err
	}

	values, err := appinstaller.BuildBaseHelmValues(ctx, kubeConfig, appConfig, p.manager.Spec.AppOwner, true)
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
	p.isDownloading = true
	p.downloadingStartTime = ptr.To(metav1.Now())
	err = p.imageClient.PollDownloadProgress(ctx, p.manager)
	if err != nil {
		klog.Errorf("poll image download progress failed %v", err)
		p.isDownloading = false
		p.downloadedTime = ptr.To(metav1.Now())
		return err
	}
	p.isDownloading = false
	p.downloadedTime = ptr.To(metav1.Now())
	p.isDownloaded = true

	err = ops.Upgrade()
	if err != nil {
		klog.Errorf("upgrade app %s failed %v", p.manager.Spec.AppName, err)
		return err
	}
	return nil
}

func (p *UpgradingApp) Cancel(ctx context.Context) error {
	var err error
	klog.Infof("execute upgrading cancel operation appName=%s", p.manager.Spec.AppName)
	err = p.imageClient.UpdateStatus(ctx, p.manager.Name, appsv1.DownloadingCanceled.String(), appsv1.DownloadingCanceled.String())
	if err != nil {
		klog.Errorf("update im name=%s to downloadingCanceled state failed %v", p.manager.Name, err)
		return err
	}

	if ok := appFactory.cancelOperation(p.manager.Name); !ok {
		klog.Errorf("app %s cancel operation is not allowed", p.manager.Name)
	}
	return nil
}

func (p *UpgradingApp) IsTimeout() bool {
	if p.isDownloading {
		if p.downloadTTL <= 0 {
			return false
		}
		return p.downloadingStartTime.Add(p.downloadTTL).Before(time.Now())
	}

	if !p.isDownloaded {
		return p.baseOperationApp.IsTimeout()
	}

	return p.downloadedTime.Add(p.ttl).Before(time.Now())
}

var _ StatefulInProgressApp = &upgradingInProgressApp{}

type upgradingInProgressApp struct {
	*UpgradingApp
	*baseStatefulInProgressApp
}

// override to avoid duplicate exec
func (p *upgradingInProgressApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	return nil, nil
}
