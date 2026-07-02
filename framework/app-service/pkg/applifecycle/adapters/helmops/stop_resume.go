package helmops

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
)

// scaleOrPatchStop chooses between HelmOps.Scale(0) (apps that
// declare workloadReplicas) and the original
// SuspendV2AppAll / SuspendV1AppOrV2Client patch path.
// Mirrors pkg/appstate/suspending_app.go:scaleOrPatchSuspend.
func (o *Ops) scaleOrPatchStop(ctx context.Context, req ports.WorkloadRequest, stopServer bool) error {
	var appCfg appcfg.ApplicationConfig
	if req.AM == nil {
		return fmt.Errorf("helmops: scaleOrPatchStop: AM is nil")
	}
	if err := json.Unmarshal([]byte(req.AM.Spec.Config), &appCfg); err != nil {
		klog.Warningf("unmarshal app config for suspend routing failed %v", err)
		return o.suspendViaPatch(ctx, req.AM, stopServer)
	}
	if !appCfg.HasWorkloadReplicas() {
		return o.suspendViaPatch(ctx, req.AM, stopServer)
	}

	ops, err := o.buildHelmOps(ctx, req)
	if err != nil {
		return err
	}
	if err := ops.Scale(0); err != nil {
		klog.Errorf("scale-to-zero for app %s failed %v", req.AM.Spec.AppName, err)
		return fmt.Errorf("suspend app %s failed %w", req.AM.Spec.AppName, err)
	}
	return nil
}

func (o *Ops) suspendViaPatch(ctx context.Context, am *appv1.ApplicationManager, stopServer bool) error {
	if stopServer {
		if err := SuspendV2AppAll(ctx, o.c, am); err != nil {
			klog.Errorf("suspend v2 app %s %s failed %v", am.Spec.Type, am.Spec.AppName, err)
			return fmt.Errorf("suspend v2 app %s failed %w", am.Spec.AppName, err)
		}
		return nil
	}
	if err := SuspendV1AppOrV2Client(ctx, o.c, am); err != nil {
		klog.Errorf("suspend app %s %s failed %v", am.Spec.Type, am.Spec.AppName, err)
		return fmt.Errorf("suspend app %s failed %w", am.Spec.AppName, err)
	}
	return nil
}

// scaleOrPatchResume chooses between HelmOps.Scale(-1) and the
// original ResumeV2AppAll / ResumeV1AppOrV2AppClient patch path.
// Mirrors pkg/appstate/resuming_app.go:scaleOrPatchResume.
func (o *Ops) scaleOrPatchResume(ctx context.Context, req ports.WorkloadRequest, resumeServer bool) error {
	var appCfg appcfg.ApplicationConfig
	if req.AM == nil {
		return fmt.Errorf("helmops: scaleOrPatchResume: AM is nil")
	}
	if err := json.Unmarshal([]byte(req.AM.Spec.Config), &appCfg); err != nil {
		klog.Warningf("unmarshal app config for resume routing failed %v", err)
		return o.resumeViaPatch(ctx, req.AM, resumeServer)
	}
	if !appCfg.HasWorkloadReplicas() {
		return o.resumeViaPatch(ctx, req.AM, resumeServer)
	}

	ops, err := o.buildHelmOps(ctx, req)
	if err != nil {
		return err
	}
	if err := ops.Scale(-1); err != nil {
		klog.Errorf("scale-up for app %s failed %v", req.AM.Spec.AppName, err)
		return fmt.Errorf("resume app %s failed %w", req.AM.Spec.AppName, err)
	}
	return nil
}

func (o *Ops) resumeViaPatch(ctx context.Context, am *appv1.ApplicationManager, resumeServer bool) error {
	if resumeServer {
		if err := ResumeV2AppAll(ctx, o.c, am); err != nil {
			klog.Errorf("resume v2 app %s %s failed %v", am.Spec.Type, am.Spec.AppName, err)
			return fmt.Errorf("resume v2 app %s failed %w", am.Spec.AppName, err)
		}
		return nil
	}
	if err := ResumeV1AppOrV2AppClient(ctx, o.c, am); err != nil {
		klog.Errorf("resume v2 app %s %s failed %v", am.Spec.Type, am.Spec.AppName, err)
		return fmt.Errorf("resume v2 app %s failed %w", am.Spec.AppName, err)
	}
	return nil
}

func retryScaleOrPatch(ctx context.Context, appName string, run func() error) error {
	err := run()
	if err == nil {
		return nil
	}
	if !IsRetryableWebhookError(err) {
		return err
	}
	klog.Infof("suspend app %s hit retryable webhook error, retrying: %v", appName, err)
	return utilwait.PollImmediate(5*time.Second, 30*time.Minute, func() (bool, error) {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		if scaleErr := run(); scaleErr != nil {
			if IsRetryableWebhookError(scaleErr) {
				klog.Infof("suspend app %s still hit retryable webhook error: %v", appName, scaleErr)
				return false, nil
			}
			return false, scaleErr
		}
		return true, nil
	})
}

func annotationTrue(am *appv1.ApplicationManager, key string) bool {
	if am == nil || am.Annotations == nil {
		return false
	}
	return am.Annotations[key] == "true"
}

func stopAllRequested(am *appv1.ApplicationManager) bool {
	return annotationTrue(am, api.AppStopAllKey)
}

func resumeAllRequested(am *appv1.ApplicationManager) bool {
	return annotationTrue(am, api.AppResumeAllKey)
}

// WaitForPodsGone implements the optional runner podStopWaiter
// extension. Delegates to WaitForPodsGone.
func (o *Ops) WaitForPodsGone(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	if req.AM == nil {
		return ports.WorkloadResult{Err: fmt.Errorf("helmops: WaitForPodsGone: AM is nil")}
	}
	if err := WaitForPodsGone(ctx, req.AM, o.c); err != nil {
		return ports.WorkloadResult{Err: err}
	}
	return ports.WorkloadResult{}
}
