package runner

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
)

// orphanForceDeleteTimeout matches the legacy LoadStatefulApp
// force-delete goroutine, which gave namespace GC 30 minutes.
const orphanForceDeleteTimeout = 30 * time.Minute

// orphanCleanups records names whose force-delete goroutine is
// still running.
type orphanCleanups struct {
	mu       sync.Mutex
	inflight map[string]struct{}
}

func (o *orphanCleanups) begin(name string) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.inflight == nil {
		o.inflight = make(map[string]struct{})
	}
	if _, ok := o.inflight[name]; ok {
		return false
	}
	o.inflight[name] = struct{}{}
	return true
}

func (o *orphanCleanups) end(name string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	delete(o.inflight, name)
}

// cleanupOrphanApplication force-deletes the Application that
// outlived its ApplicationManager. The legacy LoadStatefulApp path
// did this on AM NotFound: delete the app namespace (unless it is
// protected) and publish ForceUninstall then ForceUninstalled.
//
// A middleware ApplicationManager never has an Application CR. The
// application controller skips workloads labeled middleware=true, so
// the name {app}-middleware-{app} is the only record of the
// namespace. A missing Application for any other name is a no-op. A
// lookup error is returned so the reconciler retries. The namespace
// delete itself runs in a goroutine bounded by Runner.Stop, because
// namespace GC can take the full timeout.
func (r *Runner) cleanupOrphanApplication(ctx context.Context, name string) error {
	app, err := r.amStore.GetApplication(ctx, name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			if ns, ok := middlewareNamespaceFromAMName(name); ok {
				r.forceDeleteOrphanNamespace(name, ns, nil)
			}
			return nil
		}
		return fmt.Errorf("runner: orphan application %s: %w", name, err)
	}
	r.forceDeleteOrphanNamespace(name, app.Spec.Namespace, app)
	return nil
}

// middlewareNamespaceFromAMName returns {app}-middleware for an
// ApplicationManager named {app}-middleware-{app}. That is the name
// FmtAppMgrName produces once GetAppConfig rewrites a middleware
// namespace to {app}-middleware.
func middlewareNamespaceFromAMName(name string) (string, bool) {
	const marker = "-middleware-"
	app, rest, ok := strings.Cut(name, marker)
	if !ok || app == "" || rest != app {
		return "", false
	}
	return app + "-middleware", true
}

func (r *Runner) forceDeleteOrphanNamespace(name, namespace string, app *appv1.Application) {
	if namespace == "" || apputils.IsProtectedNamespace(namespace) {
		return
	}
	if !r.orphans.begin(name) {
		return
	}
	if app != nil {
		app = app.DeepCopy()
		klog.Infof("runner: application manager %s not found, but application %s exists", name, app.Name)
	} else {
		klog.Infof("runner: application manager %s not found; deleting middleware namespace %s", name, namespace)
	}

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		defer r.orphans.end(name)

		delCtx, cancel := context.WithTimeout(r.ctx, orphanForceDeleteTimeout)
		defer cancel()

		opID := strconv.FormatInt(r.clock.Now().Unix(), 10)
		if app != nil {
			klog.Infof("runner: force delete application %s", app.Name)
			r.publishOrphanEvent(delCtx, app, opID, appv1.Uninstalling, constants.AppForceUninstall,
				fmt.Sprintf("force delete application %s successfully", app.Name))
		}

		res := r.workloadOps.WaitForNamespaceDeleted(delCtx, ports.NamespaceWaitRequest{
			Primary: namespace,
			Timeout: orphanForceDeleteTimeout,
		})
		if res.Err != nil {
			klog.Errorf("runner: force delete namespace %s for %s failed: %v", namespace, name, res.Err)
			if app != nil {
				r.publishOrphanEvent(delCtx, app, opID, appv1.Uninstalling, constants.AppForceUninstall,
					fmt.Sprintf("force delete application %s failed: %v", app.Name, res.Err))
			}
			return
		}
		klog.Infof("runner: force delete namespace %s for %s successfully", namespace, name)
		if app == nil {
			return
		}
		specName := app.Spec.Name
		if specName == "" {
			specName = app.Name
		}
		r.publishOrphanEvent(delCtx, app, opID, appv1.Uninstalled, constants.AppForceUninstalled,
			fmt.Sprintf("app %s was force uninstalled", specName))
	}()
}

func (r *Runner) publishOrphanEvent(ctx context.Context, app *appv1.Application, opID string, state appv1.ApplicationManagerState, reason, msg string) {
	if app.Spec.Owner == "" {
		klog.Errorf("runner: skip force-uninstall event for %s: owner is empty", app.Name)
		return
	}
	name := app.Spec.Name
	if name == "" {
		name = app.Name
	}
	raw := app.Spec.RawAppName
	if raw == "" {
		raw = name
	}
	title := ""
	if app.Spec.Settings != nil {
		title = app.Spec.Settings["title"]
	}
	market := ""
	if app.Annotations != nil {
		market = app.Annotations[constants.AppMarketSourceKey]
	}
	ev := ports.AppLifecycleEvent{
		EventID:      fmt.Sprintf("%s-%s-%s-%d", app.Spec.Owner, name, state, r.clock.Now().UnixNano()),
		Name:         name,
		RawAppName:   raw,
		Type:         string(appv1.App),
		OpType:       string(appv1.UninstallOp),
		OpID:         opID,
		State:        string(state),
		User:         app.Spec.Owner,
		Title:        title,
		Reason:       reason,
		Message:      msg,
		MarketSource: market,
		CreateTime:   r.clock.Now(),
	}
	if err := r.recorder.RecordAppEvent(ctx, ev); err != nil {
		klog.Errorf("runner: force-uninstall event for %s: %v", app.Name, err)
	}
}
