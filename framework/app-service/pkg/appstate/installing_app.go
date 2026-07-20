package appstate

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/compute"
	"github.com/beclab/Olares/framework/app-service/pkg/compute/validation"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/errcode"
	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ OperationApp = &InstallingApp{}

type InstallingApp struct {
	*baseOperationApp
}

func NewInstallingApp(deps Deps,
	manager *appsv1.ApplicationManager, ttl time.Duration) (StatefulApp, StateError) {
	// TODO: check app state

	return deps.Factory.New(deps, manager, ttl,
		func(c client.Client, manager *appsv1.ApplicationManager, ttl time.Duration) StatefulApp {
			return &InstallingApp{
				baseOperationApp: &baseOperationApp{
					baseStatefulApp: &baseStatefulApp{
						manager: manager,
						client:  c,
					},

					ttl: ttl,
				},
			}
		})
}

func (p *InstallingApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	var err error
	token := p.manager.Annotations[api.AppTokenKey]
	var appCfg *appcfg.ApplicationConfig
	err = json.Unmarshal([]byte(p.manager.Spec.Config), &appCfg)
	if err != nil {
		klog.Errorf("unmarshal to appConfig failed %v", err)
		return nil, err
	}
	kubeConfig, err := p.deps.KubeConfig()
	if err != nil {
		klog.Errorf("get kube config failed %v", err)
		return nil, err
	}
	err = p.deps.SetExposePorts(ctx, appCfg, nil)
	if err != nil {
		klog.Errorf("set expose ports failed %v", err)
		return nil, err
	}

	updatedConfig, err := json.Marshal(appCfg)
	if err != nil {
		klog.Errorf("marshal appConfig failed %v", err)
		return nil, err
	}
	managerCopy := p.manager.DeepCopy()
	managerCopy.Spec.Config = string(updatedConfig)
	err = p.client.Patch(ctx, managerCopy, client.MergeFrom(p.manager))
	if err != nil {
		klog.Errorf("update ApplicationManager config failed %v", err)
		return nil, err
	}

	opCtx, cancel := context.WithCancel(context.Background())

	return p.deps.Factory.execAndWatch(opCtx, p,
		func(c context.Context) (StatefulInProgressApp, error) {
			in := installingInProgressApp{
				InstallingApp: p,
				baseStatefulInProgressApp: &baseStatefulInProgressApp{
					done:   c.Done,
					cancel: cancel,
				},
			}

			go func() {
				defer cancel()
				var allocationErr error
				allocationMessage := ""
				rejectionReason := constants.AppUnschedulable
				twoPhase := appCfg.HasWorkloadReplicas()
				valIn := validation.Input{
					Client:    p.client,
					AppConfig: appCfg,
					Op:        p.manager.Spec.OpType,
					Token:     token,
				}

				// toInstallFailed is the single entry point for every "this install
				// failed, mark the AM InstallFailed" path. Before flipping the
				// status it runs cleanupAfterInstallFailure synchronously: that
				// helper is idempotent and handles the resource gaps left by
				// post-helm-validation (path D) and Scale (path E) failures, and
				// — critically — blocks until manager.Spec.AppNamespace is
				// confirmed IsNotFound (up to installFailureNSDeletionTimeout)
				// so callers observing InstallFailed can trust the strong
				// invariant "NS already gone, AM safely deletable".
				//
				// If cleanup times out the helper returns context.DeadlineExceeded;
				// we still transition to InstallFailed (we must not block the
				// state machine indefinitely) but tag Status.Message with the
				// timeout so operators can investigate, and InstallFailedApp.Exec
				// will keep retrying the same helper on subsequent reconciles.
				toInstallFailed := func(msg string) {
					cleanupErr := cleanupAfterInstallFailure(context.TODO(), p.client, p.manager)
					finalMsg := msg
					if cleanupErr != nil {
						klog.Warningf("install-failure cleanup for %s timed out waiting for NS: %v; will retry in InstallFailedApp.Exec",
							p.manager.Name, cleanupErr)
						finalMsg = fmt.Sprintf("%s; cleanup timeout: %v", msg, cleanupErr)
					}
					p.finally = func() {
						klog.Errorf("app %s install failed, update app state to installFailed", p.manager.Spec.AppName)
						opRecord := makeRecord(p.manager, appsv1.InstallFailed, fmt.Sprintf(constants.OperationFailedTpl, p.manager.Spec.OpType, msg))
						updateErr := p.updateStatus(context.TODO(), p.manager, appsv1.InstallFailed, opRecord, finalMsg, appsv1.InstallFailed.String())
						if updateErr != nil {
							klog.Errorf("update status failed %v", updateErr)
						}
					}
				}

				recordRejection := func(decision validation.Decision) {
					if decision.Reason != "" {
						rejectionReason = decision.Reason.String()
					}
					if decision.Validator == validation.NameComputeAllocation {
						allocationErr = errors.New(decision.Message)
						klog.Errorf("allocate compute resource for app %s failed %v", p.manager.Spec.AppName, allocationErr)
						allocationMessage = fmt.Sprintf("Insufficient compute resource for selected mode %s: %v", appCfg.SelectedGpuType, allocationErr)
						compute.PublishComputeInsufficientNotification(appCfg, allocationErr)
						return
					}
					klog.Errorf("app %s install validation rejected by %s: %s", p.manager.Spec.AppName, decision.Validator, decision.Message)
					allocationErr = errors.New(decision.Message)
					allocationMessage = decision.Message
				}

				runValidation := func() (validation.Decision, error) {
					return p.deps.RunInstallValidation(c, valIn)
				}

				if !twoPhase {
					decision, runErr := runValidation()
					if runErr != nil {
						toInstallFailed(runErr.Error())
						return
					}
					if !decision.OK {
						recordRejection(decision)
						toInstallFailed(allocationMessage)
						return
					}
				}

				var ops appinstaller.HelmOpsInterface
				ops, err = p.deps.NewHelmOps(c, kubeConfig, appCfg, token,
					appinstaller.Opt{
						Source:       p.manager.Spec.Source,
						MarketSource: appcfg.GetMarketSource(p.manager),
					})
				if err != nil {
					klog.Errorf("make helm ops failed %v", err)
					toInstallFailed(err.Error())
					return
				}

				err = ops.Install()
				if err != nil {
					klog.Errorf("install app %s failed %v", p.manager.Spec.AppName, err)
					// Release compute allocation up-front so the pending-pod paths
					// (ErrServerSidePodPending / ErrPodPending → Stopping) don't leak
					// the GPU/compute reservation. The generic InstallFailed branch
					// below re-runs the same cleanup via cleanupAfterInstallFailure
					// but that call is idempotent.
					if cleanupErr := compute.DeleteAllocationsForApp(context.TODO(), p.client, appCfg.AppName, appCfg.OwnerName); cleanupErr != nil {
						klog.Warningf("cleanup compute allocation for failed install %s failed: %v", appCfg.AppName, cleanupErr)
					}
					if errors.Is(err, errcode.ErrServerSidePodPending) {
						p.finally = func() {
							klog.Infof("app %s server side pods is pending, set stop-all annotation and update app state to stopping", p.manager.Spec.AppName)

							var am appsv1.ApplicationManager
							if err := p.client.Get(context.TODO(), types.NamespacedName{Name: p.manager.Name}, &am); err != nil {
								klog.Errorf("failed to get application manager: %v", err)
								return
							}

							if am.Annotations == nil {
								am.Annotations = make(map[string]string)
							}
							am.Annotations[api.AppStopAllKey] = "true"

							if err := p.client.Update(ctx, &am); err != nil {
								klog.Errorf("failed to set stop-all annotation: %v", err)
								return
							}
							reason := constants.AppUnschedulable
							if errors.Is(err, errcode.ErrHamiUnschedulable) {
								reason = constants.AppHamiSchedulable
							}
							updateErr := p.updateStatus(ctx, &am, appsv1.Stopping, nil, err.Error(), reason)
							if updateErr != nil {
								klog.Errorf("update status failed %v", updateErr)
								return
							}
						}

						return
					}

					if errors.Is(err, errcode.ErrPodPending) {
						p.finally = func() {
							klog.Infof("app %s pods is still pending, update app state to stopping", p.manager.Spec.AppName)
							reason := constants.AppUnschedulable
							if errors.Is(err, errcode.ErrHamiUnschedulable) {
								reason = constants.AppHamiSchedulable
							}
							updateErr := p.updateStatus(context.TODO(), p.manager, appsv1.Stopping, nil, err.Error(), reason)
							if updateErr != nil {
								klog.Errorf("update status failed %v", updateErr)
								return
							}
						}

						return
					}

					toInstallFailed(err.Error())
					return
				} // end of err != nil

				if twoPhase {
					decision, runErr := runValidation()
					if runErr != nil {
						// Path D: ops.Install() already succeeded, so helm release + main
						// NS + permissions + Provider are still in the cluster. cleanupAfter
						// InstallFailure (invoked by toInstallFailed) is the only thing that
						// actually tears these down and confirms NS is gone before we mark
						// InstallFailed; the old standalone compute.DeleteAllocationsForApp
						// call has been folded into the helper.
						toInstallFailed(runErr.Error())
						return
					}
					if !decision.OK {
						recordRejection(decision)
					}
				}

				if allocationErr != nil {
					// Post-helm resource rejection (workloadReplicas apps only).
					// Switch to Stop op and Stopping; the helm release stays at
					// replicas=0 until the user retries.
					p.finally = func() {
						manager := p.manager
						var am appsv1.ApplicationManager
						if err := p.client.Get(context.TODO(), types.NamespacedName{Name: p.manager.Name}, &am); err != nil {
							klog.Errorf("failed to get application manager: %v", err)
							return
						}
						managesSharedServer, targetErr := compute.ManagesSharedServer(context.TODO(), p.client, appCfg)
						if targetErr != nil {
							klog.Warningf("failed to resolve compute target for app %s: %v", appCfg.AppName, targetErr)
						}
						if managesSharedServer {
							if am.Annotations == nil {
								am.Annotations = make(map[string]string)
							}
							am.Annotations[api.AppStopAllKey] = "true"
						}
						am.Spec.OpType = appsv1.StopOp
						am.Status.OpType = appsv1.StopOp
						if err := p.client.Update(context.TODO(), &am); err != nil {
							klog.Errorf("failed to update app manager stop metadata: %v", err)
							return
						}
						manager = &am
						updateErr := p.updateStatus(context.TODO(), manager, appsv1.Stopping, nil, allocationMessage, rejectionReason)
						if updateErr != nil {
							klog.Errorf("update status failed %v", updateErr)
						}
					}
					return
				}

				// Scale up to the manifest-declared replica counts via a
				// second helm upgrade. Any failure here is treated as a
				// generic install failure to mirror the v1 behavior.
				//
				// Path E: ops.Install() already succeeded, so the helm release is in
				// the cluster. cleanupAfterInstallFailure (invoked by toInstallFailed)
				// owns the teardown + NS-gone confirmation before InstallFailed; the
				// old standalone compute.DeleteAllocationsForApp call has been folded
				// into the helper.
				if scaleErr := ops.Scale(-1); scaleErr != nil {
					klog.Errorf("scale-up after install failed for app %s: %v", p.manager.Spec.AppName, scaleErr)
					toInstallFailed(scaleErr.Error())
					return
				}

				// Pods exist now; wait for readiness. Middleware apps use
				// WaitForLaunch instead — HelmOps.Install already skips
				// startup polling for them.
				if p.manager.Spec.Type != appsv1.Middleware {
					if ok, waitErr := ops.WaitForStartUp(); !ok {
						// A cancel (DELETE /apps/{name}/install) cancels opCtx,
						// the only path on which WaitForStartUp returns (false, nil).
						// That is NOT a startup failure: the InstallingCanceling
						// state machine has already taken over this ApplicationManager
						// and owns the terminal transition. Writing Stopping/InitFailed
						// here would both mislabel a user cancel as an init failure and
						// race the cancel path into a spurious installingCancelFailed.
						// Bail out quietly and let the cancel handler drive the state.
						if waitErr == nil || c.Err() != nil {
							klog.Infof("install of app %s canceled while waiting for startup; leaving terminal state to the cancel handler", p.manager.Spec.AppName)
							return
						}

						klog.Errorf("wait for app %s startup after scale-up failed %v", p.manager.Spec.AppName, waitErr)
						reason := constants.AppStopDueToStartUpFailed
						wrappedWaitErr := errors.Wrapf(waitErr, "wait for app %s startup after scale-up failed", p.manager.Spec.AppName)
						if errors.Is(waitErr, errcode.ErrPodPending) || errors.Is(waitErr, errcode.ErrServerSidePodPending) {
							reason = constants.AppUnschedulable
							if errors.Is(waitErr, errcode.ErrHamiUnschedulable) {
								reason = constants.AppHamiSchedulable
							}
						}
						msg := wrappedWaitErr.Error()
						p.finally = func() {
							updateErr := p.updateStatus(context.TODO(), p.manager, appsv1.Stopping, nil, msg, reason)
							if updateErr != nil {
								klog.Errorf("update status failed %v", updateErr)
							}
						}
						return
					}
				}

				if p.manager.Spec.Type == appsv1.Middleware {
					ok, err := ops.WaitForLaunch()
					if !ok {
						// A cancel (DELETE /apps/{name}/install) or a force
						// uninstall cancels opCtx, on which WaitForLaunch returns
						// (false, ctx.Err()). That is NOT a launch failure: the
						// initiator has already written the terminal target state
						// (InstallingCanceling for cancel, Uninstalling for force
						// uninstall) and owns the transition. Writing
						// InstallingCanceling here would mislabel the initiator's
						// intent and race the cancel/uninstall path. Bail out
						// quietly and let the initiator drive the state.
						if err == nil || c.Err() != nil {
							klog.Infof("install of middleware %s canceled while waiting for launch; leaving terminal state to the initiator", p.manager.Spec.AppName)
							return
						}

						klog.Errorf("wait for middleware %s launch failed %v", p.manager.Spec.AppName, err)
						p.finally = func() {
							klog.Info("update app manager status to installing canceling, ", p.manager.Name)
							updateErr := p.updateStatus(context.TODO(), p.manager, appsv1.InstallingCanceling, nil, appsv1.InstallingCanceling.String(), constants.AppStopDueToStartUpFailed)
							if updateErr != nil {
								klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, appsv1.InstallingCanceling, updateErr)
								return
							}

						}
						return
					}
					p.finally = func() {
						message := fmt.Sprintf(constants.InstallOperationCompletedTpl, p.manager.Spec.Type.String(), p.manager.Spec.AppName)
						opRecord := makeRecord(p.manager, appsv1.Running, message)
						updateErr := p.updateStatus(context.TODO(), p.manager, appsv1.Running, opRecord, appsv1.Running.String(), appsv1.Running.String())
						if updateErr != nil {
							klog.Errorf("update app manager %s to %s state failed %v", p.manager.Name, appsv1.Running, updateErr)
							return
						}
					}
				} else {
					p.finally = func() {
						klog.Infof("app %s install successfully, update app state to initializing", p.manager.Spec.AppName)
						updateErr := p.updateStatus(context.TODO(), p.manager, appsv1.Initializing, nil, appsv1.Initializing.String(), appsv1.Initializing.String())
						if updateErr != nil {
							klog.Errorf("update status failed %v", updateErr)
							return
						}

					}
				}
			}()

			return &in, nil
		},
	)
}

func (p *InstallingApp) Cancel(ctx context.Context) error {
	err := p.updateStatus(ctx, p.manager, appsv1.InstallingCanceling, nil, constants.InstallCanceledByTimeout, constants.InstallCancelBySystem)
	if err != nil {
		klog.Errorf("update appmgr state to installingCanceling state failed %v", err)
		return err
	}

	return nil
}

var _ StatefulInProgressApp = &installingInProgressApp{}

type installingInProgressApp struct {
	*InstallingApp
	*baseStatefulInProgressApp
}

// override to avoid duplicate exec
func (p *installingInProgressApp) Exec(ctx context.Context) (StatefulInProgressApp, error) {
	return nil, nil
}
