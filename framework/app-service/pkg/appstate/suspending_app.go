package appstate

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/compute"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	kbopv1alpha1 "github.com/apecloud/kubeblocks/apis/operations/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ OperationApp = &SuspendingApp{}

type SuspendingApp struct {
	*baseOperationApp
}

func NewSuspendingApp(deps Deps,
	manager *appsv1.ApplicationManager, ttl time.Duration) (StatefulApp, StateError) {
	// TODO: check app state

	return deps.Factory.New(deps, manager, ttl,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &SuspendingApp{
				&baseOperationApp{
					ttl: ttl,
					baseStatefulApp: &baseStatefulApp{
						manager: manager,
						client:  c,
					},
				},
			}
		})
}

func (p *SuspendingApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	opCtx, cancel := context.WithCancel(context.Background())

	return p.deps.Factory.execAndWatch(opCtx, p,
		func(c context.Context) (StatefulInProgressApp, error) {
			in := suspendingInProgressApp{
				SuspendingApp: p,
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
						klog.Infof("suspending app %s failed %v", p.manager.Spec.AppName, err)
						opRecord := makeRecord(p.manager, appsv1.StopFailed, fmt.Sprintf(constants.OperationFailedTpl, p.manager.Spec.OpType, err.Error()))
						updateErr := p.updateStatus(context.TODO(), p.manager, appsv1.StopFailed, opRecord, err.Error(), appsv1.StopFailed.String())
						if updateErr != nil {
							klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, appsv1.StopFailed.String(), err)
							err = errors.Wrapf(err, "update status failed %v", updateErr)
							return
						}
					}
					return
				}

				p.finally = func() {
					klog.Infof("suspended app %s success", p.manager.Spec.AppName)
					opRecord := makeRecord(p.manager, appsv1.Stopped, fmt.Sprintf(constants.StopOperationCompletedTpl, p.manager.Spec.AppName))
					// Read latest status directly from apiserver to avoid cache staleness
					reason := p.manager.Status.Reason
					if cli, getErr := utils.GetClient(); getErr == nil {
						if am, getErr := cli.AppV1alpha1().ApplicationManagers().Get(context.TODO(), p.manager.Name, metav1.GetOptions{}); getErr == nil && am != nil {
							if am.Status.Reason != "" {
								reason = am.Status.Reason
							}
						}
					}
					updateErr := p.updateStatus(context.TODO(), p.manager, appsv1.Stopped, opRecord, fmt.Sprintf(constants.StopOperationCompletedTpl, p.manager.Spec.AppName), reason)
					if updateErr != nil {
						klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, appsv1.Stopped.String(), updateErr)
						return
					}
				}
			}()

			return &in, nil
		})
}

func (p *SuspendingApp) waitForPodsGone(ctx context.Context) error {
	// Scale(0)/patch only requests termination; do not claim Stopped while
	// pods remain or upgrade resource checks that treat Stopped as "full"
	// would under-count live requests. Cancel interrupts via ctx.
	return utilwait.PollImmediate(time.Second, 30*time.Minute, func() (bool, error) {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		pods, err := listAppPods(p.manager, p.client)
		if err != nil {
			klog.Errorf("list pods while suspending app %s failed %v", p.manager.Spec.AppName, err)
			return false, err
		}
		if len(pods) == 0 {
			return true, nil
		}
		klog.Infof("suspend app %s waiting for %d pod(s) to terminate in namespace %s",
			p.manager.Spec.AppName, len(pods), p.manager.Spec.AppNamespace)
		return false, nil
	})
}

func (p *SuspendingApp) exec(ctx context.Context) error {
	// Check if stop-all is requested for V2 apps to also stop server-side shared charts
	stopServer := p.manager.Annotations[api.AppStopAllKey] == "true"

	// v1/v3 apps that declare workloadReplicas scale to 0 via a helm
	// upgrade (HelmOps.Scale(0)) instead of patching workloads directly.
	// v2 apps and legacy v1/v3 manifests fall back to the original
	// suspendV2AppAll / suspendV1AppOrV2Client patch path.
	if err := p.scaleOrPatchSuspend(ctx, stopServer); err != nil {
		// Cold-start webhook/Service lag: keep retrying Scale/patch until
		// success or the Stopping TTL Cancel interrupts ctx, instead of
		// permanently landing in StopFailed on a transient failure.
		if IsRetryableWebhookError(err) {
			klog.Infof("suspend app %s hit retryable webhook error, retrying: %v", p.manager.Spec.AppName, err)
			err = utilwait.PollImmediate(5*time.Second, 30*time.Minute, func() (bool, error) {
				if err := ctx.Err(); err != nil {
					return false, err
				}
				if scaleErr := p.scaleOrPatchSuspend(ctx, stopServer); scaleErr != nil {
					if IsRetryableWebhookError(scaleErr) {
						klog.Infof("suspend app %s still hit retryable webhook error: %v", p.manager.Spec.AppName, scaleErr)
						return false, nil
					}
					return false, scaleErr
				}
				return true, nil
			})
		}
		if err != nil {
			return err
		}
	}
	var appCfg appcfg.ApplicationConfig
	if err := json.Unmarshal([]byte(p.manager.Spec.Config), &appCfg); err != nil {
		klog.Errorf("unmarshal app config for compute cleanup failed %v", err)
		return err
	}

	if stopServer && appCfg.APIVersion == appcfg.V2 {
		// For V2 cluster-scoped apps, when server is down, stop all other users' clients
		// because they share the same server and cannot function without it
		klog.Infof("stopping other users' clients for v2 app %s", p.manager.Spec.AppName)

		var appManagerList appsv1.ApplicationManagerList
		if err := p.client.List(ctx, &appManagerList); err != nil {
			klog.Errorf("failed to list application managers: %v", err)
		} else {
			// find all ApplicationManagers with same AppName but different AppOwner
			for _, am := range appManagerList.Items {
				// Skip if same owner (already handled) or different app
				if am.Spec.AppName != p.manager.Spec.AppName || am.Spec.AppOwner == p.manager.Spec.AppOwner {
					continue
				}

				if am.Spec.Type != appsv1.App && am.Spec.Type != appsv1.Middleware {
					continue
				}

				if am.Status.State == appsv1.Stopped || am.Status.State == appsv1.Stopping {
					klog.Infof("app %s owner %s already in stopped/stopping state, skip", am.Spec.AppName, am.Spec.AppOwner)
					continue
				}

				if !IsOperationAllowed(am.Status.State, appsv1.StopOp) {
					klog.Infof("app %s owner %s not allowed do stop operation, skip", am.Spec.AppName, am.Spec.AppOwner)
					continue
				}
				opID := strconv.FormatInt(time.Now().Unix(), 10)
				now := metav1.Now()
				status := appsv1.ApplicationManagerStatus{
					OpType:     appsv1.StopOp,
					OpID:       opID,
					State:      appsv1.Stopping,
					StatusTime: &now,
					UpdateTime: &now,
					Reason:     p.manager.Status.Reason,
					Message:    p.manager.Status.Message,
				}
				if _, err := apputils.UpdateAppMgrStatus(am.Name, status); err != nil {
					return err
				}

				klog.Infof("stopping client for user %s, app %s", am.Spec.AppOwner, am.Spec.AppName)

			}
		}
	}

	if p.manager.Spec.Type == appsv1.Middleware && userspace.IsKbMiddlewares(p.manager.Spec.AppName) {
		err := p.execMiddleware(ctx)
		if err != nil {
			klog.Errorf("suspend middleware %s failed %v", p.manager.Spec.AppName, err)
			return err
		}
	}

	if err := p.waitForPodsGone(ctx); err != nil {
		klog.Errorf("waiting for app %s pods to terminate failed %v", p.manager.Spec.AppName, err)
		return err
	}

	if err := compute.DeleteAllocationsForComputeTarget(ctx, p.client, &appCfg, stopServer); err != nil {
		klog.Errorf("delete compute allocation for suspended app %s failed %v", p.manager.Spec.AppName, err)
		return err
	}
	return nil
}

func (p *SuspendingApp) Cancel(ctx context.Context) error {
	klog.Infof("cancel suspending operation appName=%s", p.manager.Spec.AppName)
	if ok := p.deps.Factory.cancelOperation(p.manager.Name); !ok {
		klog.Errorf("app %s operation is not in progress", p.manager.Name)
	}

	opRecord := makeRecord(p.manager, appsv1.StopFailed,
		fmt.Sprintf(constants.OperationFailedTpl, p.manager.Spec.OpType, "stopping ttl exceeded"))
	return p.updateStatus(ctx, p.manager, appsv1.StopFailed, opRecord,
		"stopping ttl exceeded", appsv1.StopFailed.String())
}

// scaleOrPatchSuspend chooses between the helm-upgrade-based Scale(0)
// flow (apps with workloadReplicas) and the legacy direct-patch
// suspendOrResumeApp flow (apps without workloadReplicas in their
// manifest). Routing mirrors resuming_app.scaleOrPatchResume.
func (p *SuspendingApp) scaleOrPatchSuspend(ctx context.Context, stopServer bool) error {
	var appCfg appcfg.ApplicationConfig
	if err := json.Unmarshal([]byte(p.manager.Spec.Config), &appCfg); err != nil {
		klog.Warningf("unmarshal app config for suspend routing failed %v", err)
		// fall back to legacy patch path on unmarshal failure
		return p.suspendViaPatch(ctx, stopServer)
	}
	if !appCfg.HasWorkloadReplicas() {
		return p.suspendViaPatch(ctx, stopServer)
	}

	kubeConfig, err := p.deps.KubeConfig()
	if err != nil {
		klog.Errorf("get kube config failed %v", err)
		return err
	}
	token := p.manager.Annotations[api.AppTokenKey]
	ops, err := p.deps.NewHelmOps(ctx, kubeConfig, &appCfg, token,
		appinstaller.Opt{
			Source:       p.manager.Spec.Source,
			MarketSource: appcfg.GetMarketSource(p.manager),
		})
	if err != nil {
		klog.Errorf("make helm ops for suspend failed %v", err)
		return err
	}
	if err := ops.Scale(0); err != nil {
		klog.Errorf("scale-to-zero for app %s failed %v", p.manager.Spec.AppName, err)
		return fmt.Errorf("suspend app %s failed %w", p.manager.Spec.AppName, err)
	}
	return nil
}

// suspendViaPatch is the legacy direct-patch implementation used by v2
// apps and v1/v3 apps without workloadReplicas. It preserves the
// stop-all branching for V2 shared-server charts.
func (p *SuspendingApp) suspendViaPatch(ctx context.Context, stopServer bool) error {
	if stopServer {
		if err := suspendV2AppAll(ctx, p.client, p.manager); err != nil {
			klog.Errorf("suspend v2 app %s %s failed %v", p.manager.Spec.Type, p.manager.Spec.AppName, err)
			return fmt.Errorf("suspend v2 app %s failed %w", p.manager.Spec.AppName, err)
		}
		return nil
	}
	if err := suspendV1AppOrV2Client(ctx, p.client, p.manager); err != nil {
		klog.Errorf("suspend app %s %s failed %v", p.manager.Spec.Type, p.manager.Spec.AppName, err)
		return fmt.Errorf("suspend app %s failed %w", p.manager.Spec.AppName, err)
	}
	return nil
}

func (p *SuspendingApp) execMiddleware(ctx context.Context) error {
	op := p.deps.NewMiddlewareOp(ctx, kbopv1alpha1.StopType, p.manager, p.client)
	err := op.Stop()
	if err != nil {
		klog.Errorf("failed to stop middleware %s,err=%v", p.manager.Spec.AppName, err)
		return err
	}
	return nil
}

var _ StatefulInProgressApp = &suspendingInProgressApp{}

type suspendingInProgressApp struct {
	*SuspendingApp
	*baseStatefulInProgressApp
}

// override to avoid duplicate exec
func (p *suspendingInProgressApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	return nil, nil
}
