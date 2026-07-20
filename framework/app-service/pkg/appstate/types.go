package appstate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/compute"
	"github.com/beclab/Olares/framework/app-service/pkg/middlewareinstaller"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StatefulApp interface {
	GetManager() *appsv1.ApplicationManager
	State() string
	Finally()
}

type baseStatefulApp struct {
	finallyApp
	app     *appsv1.Application
	manager *appsv1.ApplicationManager
	client  client.Client
}

func (b *baseStatefulApp) GetManager() *appsv1.ApplicationManager {
	return b.manager
}

func (b *baseStatefulApp) State() string {
	return b.GetManager().Status.State.String()
}

// func (b *baseStatefulApp) GetApp() *appsv1.Application {
// 	return b.app
// }

func (b *baseStatefulApp) updateStatus(ctx context.Context, am *appsv1.ApplicationManager, state appsv1.ApplicationManagerState,
	opRecord *appsv1.OpRecord, message, reason string) error {
	// The read-modify-write below must be atomic: the same ApplicationManager
	// can be patched concurrently by the main reconcile, per-controller
	// reconciles, apiserver handlers and the appFactory watcher's Finally().
	// We use an optimistic-lock patch (resourceVersion precondition) and retry
	// on conflict so concurrent callers cannot clobber each other's
	// OpGeneration increment or OpRecords.
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := b.client.Get(ctx, types.NamespacedName{Name: am.Name}, am); err != nil {
			return err
		}

		// Reject writes that are not declared in StateTransitions. The check
		// runs INSIDE the retry loop so that if the persisted state has
		// changed underneath us (e.g. user cancelled while this goroutine
		// was racing toward InstallFailed) we refuse to clobber the new
		// terminal state with a stale transition. Same-state writes are
		// still allowed via IsStateTransitionAllowed so idempotent retries
		// and re-assertions (updated message/reason) keep working.
		if !IsStateTransitionAllowed(am.Status.State, state) {
			err := fmt.Errorf("invalid state transition for %s: %s -> %s (not declared in StateTransitions)",
				am.Name, am.Status.State, state)
			klog.Warningf("updateStatus rejected: %v", err)
			return err
		}

		now := metav1.Now()
		amCopy := am.DeepCopy()
		amCopy.Status.State = state
		amCopy.Status.Message = message
		if reason != "" {
			amCopy.Status.Reason = reason
		}
		amCopy.Status.StatusTime = &now
		amCopy.Status.UpdateTime = &now
		amCopy.Status.OpGeneration += 1
		if opRecord != nil {
			amCopy.Status.OpRecords = append([]appsv1.OpRecord{*opRecord}, amCopy.Status.OpRecords...)
		}
		if len(amCopy.Status.OpRecords) > 20 {
			amCopy.Status.OpRecords = amCopy.Status.OpRecords[:20:20]
		}
		err := b.client.Patch(ctx, amCopy, client.MergeFromWithOptions(am, client.MergeFromWithOptimisticLock{}))
		if err != nil {
			klog.Errorf("patch appmgr's  %s status failed %v", am.Name, err)
			return err
		}
		return nil
	})
}

func (p *baseStatefulApp) forceDeleteApp(ctx context.Context) error {
	token := p.manager.Annotations[api.AppTokenKey]
	if p.manager.Spec.Config == "" && p.manager.Spec.Source == "system" {
		klog.Infof("app %s config is empty, source is system", p.manager.Name)
		err := p.updateStatus(ctx, p.manager, appsv1.Uninstalled, nil, appsv1.Uninstalled.String(), appsv1.Uninstalled.String())
		if err != nil {
			klog.Errorf("update app manager %s to state %s failed", p.manager.Name, appsv1.Uninstalled)
			return err
		}

		return nil
	}

	var appCfg *appcfg.ApplicationConfig
	err := json.Unmarshal([]byte(p.manager.Spec.Config), &appCfg)
	if err != nil {
		klog.Errorf("unmarshal to appConfig failed %v", err)
		return err
	}

	kubeConfig, err := getKubeConfig()
	if err != nil {
		klog.Errorf("get kube config failed %v", err)
		return err
	}
	if appCfg.MiddlewareName == "mongodb" && appCfg.Namespace == "os-platform" {
		return p.oldMongodbUninstall(ctx, kubeConfig)
	}

	ops, err := newHelmOps(ctx, kubeConfig, appCfg, token, appinstaller.Opt{MarketSource: appcfg.GetMarketSource(p.manager)})
	if err != nil {
		klog.Errorf("make helm ops failed %v", err)
		return err
	}
	err = ops.Uninstall()
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			klog.Errorf("uninstall app %s failed err %v", appCfg.AppName, err)
			return err
		}
	}

	// forceDeleteApp is the shared exit toward Uninstalled for the force-delete
	// paths (UninstallFailed, RunningApp / UninstalledApp self-heal). The normal
	// Uninstalling -> Uninstalled flow releases the compute allocation, but
	// these paths bypass UninstallingApp, so release it here too or the app
	// would leak its GPU/compute reservation after the workload is gone.
	uninstallAll := p.manager.Annotations[api.AppUninstallAllKey] == "true"
	if _, err = compute.EnsureAllocationsDeletedForComputeTarget(ctx, p.client, appCfg, uninstallAll); err != nil {
		klog.Errorf("delete compute allocation for force-deleted app %s failed %v", appCfg.AppName, err)
		return err
	}

	// Wait for namespace to be fully deleted before updating status
	if err = p.waitForNamespaceDeleted(ctx); err != nil {
		klog.Errorf("wait for namespace %s deleted failed %v", p.manager.Spec.AppNamespace, err)
		return err
	}

	err = p.updateStatus(ctx, p.manager, appsv1.Uninstalled, nil, appsv1.Uninstalled.String(), appsv1.Uninstalled.String())
	if err != nil {
		klog.Errorf("update app manager %s to state %s failed", p.manager.Name, appsv1.Uninstalled)
		return err
	}
	return nil
}

// waitForNamespaceDeleted performs a single-shot check on the namespace and
// returns a RequeueError if the namespace is still present, asking the
// controller to re-enqueue the request after a short delay. This avoids the
// caller (typically the reconcile worker) blocking on a long PollImmediate
// loop and starving every other ApplicationManager — with MaxConcurrentReconciles=1
// a 30-minute synchronous wait here would freeze the whole controller.
//
// Callers must propagate the returned error verbatim so the reconciler can
// dispatch on appstate.RequeueError. The helm uninstall / compute cleanup
// steps that precede this check in forceDeleteApp are idempotent, so it is
// safe to re-run them on every requeue iteration.
func (p *baseStatefulApp) waitForNamespaceDeleted(ctx context.Context) error {
	namespace := p.manager.Spec.AppNamespace
	if apputils.IsProtectedNamespace(namespace) {
		return nil
	}

	var ns corev1.Namespace
	err := p.client.Get(ctx, types.NamespacedName{Name: namespace}, &ns)
	if err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("failed to get namespace %s: %v", namespace, err)
		return err
	}
	if apierrors.IsNotFound(err) {
		klog.Infof("namespace %s has been fully deleted", namespace)
		return nil
	}
	klog.Infof("namespace %s still exists, requeueing in 5s", namespace)
	return NewWaitingInLine(5)
}

type OperationApp interface {
	StatefulApp
	IsTimeout() bool
	Exec(ctx context.Context) (StatefulInProgressApp, error)

	// Cancel update the app to cancel state, into the next phase
	Cancel(ctx context.Context) error
}

type baseOperationApp struct {
	*baseStatefulApp
	ttl time.Duration
}

func (b *baseOperationApp) IsTimeout() bool {
	if b.ttl <= 0 {
		return false
	}
	return b.GetManager().Status.StatusTime.Add(b.ttl).Before(time.Now())
}

type CancelOperationApp interface {
	OperationApp
	IsAppCreated() bool
	// Failed() error
}

type StatefulInProgressApp interface {
	OperationApp

	// Cleanup Stop the current operation immediately and clean up the resource if necessary.
	Cleanup(ctx context.Context)
	Done() <-chan struct{}
}

type finallyApp struct {
	finally func()
}

func (f *finallyApp) Finally() {
	if f.finally != nil {
		f.finally()
	}
}

type baseStatefulInProgressApp struct {
	done   func() <-chan struct{}
	cancel context.CancelFunc
}

func (p *baseStatefulInProgressApp) Done() <-chan struct{} {
	if p.done != nil {
		return p.done()
	}

	return nil
}

func (p *baseStatefulInProgressApp) Cleanup(ctx context.Context) {
	if p.cancel != nil {
		p.cancel()
	}
}

// PollableStatefulInProgressApp is an interface for applications that can be polled for their state.
type PollableStatefulInProgressApp interface {
	StatefulInProgressApp
	poll(ctx context.Context) error
	stopPolling()
	WaitAsync(ctx context.Context)
	CreatePollContext() context.Context
}

type basePollableStatefulInProgressApp struct {
	cancelPoll context.CancelFunc
	ctxPoll    context.Context
}

// Cleanup implements PollableStatefulInProgressApp.
func (r *basePollableStatefulInProgressApp) Cleanup(ctx context.Context) {
	r.stopPolling()
}

func (r *basePollableStatefulInProgressApp) stopPolling() {
	if r != nil && r.cancelPoll != nil {
		r.cancelPoll()
	} else {
		klog.Errorf("call cancelPool failed with nil pointer r ")
	}
}

func (p *basePollableStatefulInProgressApp) Done() <-chan struct{} {
	if p.ctxPoll == nil {
		return nil
	}

	return p.ctxPoll.Done()
}

func (p *basePollableStatefulInProgressApp) CreatePollContext() context.Context {
	pollCtx, cancel := context.WithCancel(context.Background())
	p.cancelPoll = cancel
	p.ctxPoll = pollCtx

	return pollCtx
}

func (b *baseStatefulApp) oldMongodbUninstall(ctx context.Context, kubeConfig *rest.Config) error {
	mc := &middlewareinstaller.MiddlewareConfig{
		MiddlewareName: b.manager.Spec.AppName,
		Namespace:      b.manager.Spec.AppNamespace,
		OwnerName:      b.manager.Spec.AppOwner,
	}
	err := middlewareinstaller.Uninstall(ctx, kubeConfig, mc)
	if err != nil && err.Error() != "failed to delete release: mongodb" {
		klog.Errorf("failed to uninstall old mongodb %v", err)
		return err
	}
	var secret corev1.Secret

	err = b.client.Get(ctx, types.NamespacedName{Name: "sh.helm.release.v1.mongodb.v1", Namespace: mc.Namespace}, &secret)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if err = b.client.Delete(ctx, &secret); err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("failed to delete mongodb release secret: %s", secret.Name)
		return err
	}

	return nil
}
