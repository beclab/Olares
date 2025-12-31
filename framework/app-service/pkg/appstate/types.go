package appstate

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	appsv1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller/versioned"
	appevent "github.com/beclab/Olares/framework/app-service/pkg/event"
	"github.com/beclab/Olares/framework/app-service/pkg/middlewareinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
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
	var err error

	err = b.client.Get(ctx, types.NamespacedName{Name: am.Name}, am)
	if err != nil {
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
	err = b.client.Patch(ctx, amCopy, client.MergeFrom(am))
	if err != nil {
		klog.Errorf("patch appmgr's  %s status failed %v", am.Name, err)
		return err
	}
	appevent.PublishAppEventToQueue(utils.EventParams{
		Owner:      b.manager.Spec.AppOwner,
		Name:       b.manager.Spec.AppName,
		OpType:     string(b.manager.Spec.OpType),
		OpID:       b.manager.Status.OpID,
		State:      state.String(),
		RawAppName: b.manager.Spec.RawAppName,
		Type:       b.manager.Spec.Type.String(),
		Title:      apputils.AppTitle(b.manager.Spec.Config),
		Reason:     reason,
		Message:    message,
	})

	return nil
}

func (p *baseStatefulApp) forceDeleteApp(ctx context.Context) error {
	token := p.manager.Annotations[api.AppTokenKey]
	if p.manager.Spec.Config == "" && p.manager.Spec.Source == "system" {
		klog.Infof("app %s config is empty, source is system", p.manager.Name)
		err := p.updateStatus(ctx, p.manager, appsv1.Uninstalled, nil, appsv1.Uninstalled.String(), "")
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

	kubeConfig, err := ctrl.GetConfig()
	if err != nil {
		klog.Errorf("get kube config failed %v", err)
		return err
	}
	if appCfg.MiddlewareName == "mongodb" && appCfg.Namespace == "os-platform" {
		return p.oldMongodbUninstall(ctx, kubeConfig)
	}

	ops, err := versioned.NewHelmOps(ctx, kubeConfig, appCfg, token, appinstaller.Opt{MarketSource: p.manager.GetMarketSource()})
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

	// Wait for namespace to be fully deleted before updating status
	if err = p.waitForNamespaceDeleted(ctx); err != nil {
		klog.Errorf("wait for namespace %s deleted failed %v", p.manager.Spec.AppNamespace, err)
		return err
	}

	err = p.updateStatus(ctx, p.manager, appsv1.Uninstalled, nil, appsv1.Uninstalled.String(), "")
	if err != nil {
		klog.Errorf("update app manager %s to state %s failed", p.manager.Name, appsv1.Uninstalled)
		return err
	}
	return nil
}

// waitForNamespaceDeleted waits for the namespace to be completely deleted
func (p *baseStatefulApp) waitForNamespaceDeleted(ctx context.Context) error {
	namespace := p.manager.Spec.AppNamespace
	if apputils.IsProtectedNamespace(namespace) {
		return nil
	}

	klog.Infof("waiting for namespace %s to be fully deleted", namespace)
	err := utilwait.PollImmediate(time.Second, 30*time.Minute, func() (done bool, err error) {
		var ns corev1.Namespace
		err = p.client.Get(ctx, types.NamespacedName{Name: namespace}, &ns)
		if err != nil && !apierrors.IsNotFound(err) {
			klog.Errorf("failed to get namespace %s: %v", namespace, err)
			return false, err
		}
		if apierrors.IsNotFound(err) {
			klog.Infof("namespace %s has been fully deleted", namespace)
			return true, nil
		}
		klog.Infof("namespace %s still exists, waiting...", namespace)
		return false, nil
	})

	return err
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
