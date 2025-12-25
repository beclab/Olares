package appstate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	appsv1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
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

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ OperationApp = &UpgradingApp{}

type UpgradingApp struct {
	*baseOperationApp
	imageClient images.ImageManager
}

func (p *UpgradingApp) State() string {
	return p.GetManager().Status.State.String()
}

func NewUpgradingApp(c client.Client,
	manager *appsv1.ApplicationManager, ttl time.Duration) (StatefulApp, StateError) {

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
				imageClient: images.NewImageManager(c),
			}
		})
}

func (p *UpgradingApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {

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

				err := p.exec(c)
				if err != nil {
					p.finally = func() {
						klog.Info("upgrade app failed, update app status to upgradeFailed, ", p.manager.Name)
						opRecord := makeRecord(p.manager, appsv1.UpgradeFailed, fmt.Sprintf(constants.OperationFailedTpl, p.manager.Spec.OpType, err.Error()))

						updateErr := p.updateStatus(context.TODO(), p.manager, appsv1.UpgradeFailed, opRecord, err.Error(), "")
						if updateErr != nil {
							klog.Errorf("update appmgr state to upgradeFailed state failed %v", updateErr)
						}
					}
					return
				}

				p.finally = func() {
					klog.Info("upgrade app success, update app status to initializing, ", p.manager.Name)
					updateErr := p.updateStatus(context.TODO(), p.manager, appsv1.Initializing, nil, appsv1.Initializing.String(), "")
					if updateErr != nil {
						klog.Errorf("update appmgr state to initializing state failed %v", updateErr)
						return
					}

				}

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

	if !userspace.IsSysApp(getRawAppName(p.manager.Spec.RawAppName)) {
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
		})

		if err != nil {
			klog.Errorf("get app config failed %v", err)
			return err
		}
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
		appinstaller.Opt{Source: p.manager.Spec.Source, MarketSource: p.manager.GetMarketSource()})
	if err != nil {
		klog.Errorf("make helmop failed %v", err)
		return err
	}
	if isAdmin {
		admin = p.manager.Spec.AppOwner
	}
	values := map[string]interface{}{
		"admin": admin,
		"bfl": map[string]string{
			"username": p.manager.Spec.AppOwner,
		},
	}
	var nodes corev1.NodeList
	err = p.client.List(ctx, &nodes, &client.ListOptions{})
	if err != nil {
		klog.Errorf("list node failed %v", err)
		return err
	}
	gpuType, err := utils.FindGpuTypeFromNodes(&nodes)
	if err != nil {
		klog.Errorf("get gpu type failed %v", gpuType)
		return err
	}
	values["GPU"] = map[string]interface{}{
		"Type": gpuType,
		"Cuda": os.Getenv("OLARES_SYSTEM_CUDA_VERSION"),
	}

	terminus, err := utils.GetTerminusVersion(ctx, kubeConfig)
	if err != nil {
		klog.Infof("get terminus error %v", err)
		return err
	}
	values["sysVersion"] = terminus.Spec.Version

	nodeInfo, err := utils.GetNodeInfo(ctx)
	if err != nil {
		klog.Errorf("failed to get node info %v", err)
		return err
	}
	values["nodes"] = nodeInfo

	deviceName, err := utils.GetDeviceName()
	if err != nil {
		klog.Errorf("failed to get deviceName %v", err)
		return err
	}

	values["deviceName"] = deviceName

	refs, err := p.getRefsForImageManager(appConfig, values)
	if err != nil {
		klog.Errorf("get image refs from resources failed %v", err)
		return err
	}
	err = p.imageClient.Create(ctx, p.manager, refs)
	if err != nil {
		klog.Errorf("create imagemanager failed %v", err)
		return err
	}
	err = p.imageClient.PollDownloadProgress(ctx, p.manager)
	if err != nil {
		klog.Errorf("poll image download progress failed %v", err)
		return err
	}
	err = ops.Upgrade()
	if err != nil {
		klog.Errorf("upgrade app %s failed %v", p.manager.Spec.AppName, err)
		return err
	}
	return nil
}

func (p *UpgradingApp) Cancel(ctx context.Context) error {
	err := p.updateStatus(ctx, p.manager, appsv1.UpgradingCanceling, nil, constants.OperationCanceledByTerminusTpl, "")
	if err != nil {
		klog.Errorf("update appmgr state to upgradingCanceling state failed %v", err)
		return err
	}
	return nil
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

func (p *UpgradingApp) getRefsForImageManager(appConfig *appcfg.ApplicationConfig, values map[string]interface{}) (refs []appsv1.Ref, err error) {
	switch {
	case appConfig.APIVersion == appcfg.V2 && appConfig.IsMultiCharts():
		// For V2 multi-charts, we need to get refs from each chart
		var chartRefs []appsv1.Ref
		for _, chart := range appConfig.SubCharts {
			chartRefs, err = utils.GetRefFromResourceList(chart.ChartPath(appConfig.AppName), values, appConfig.Images)
			if err != nil {
				klog.Errorf("get refs from chart %s failed %v", chart.Name, err)
				return
			}

			refs = append(refs, chartRefs...)
		}
	default:
		refs, err = utils.GetRefFromResourceList(appConfig.ChartsName, values, appConfig.Images)
	}
	return
}
