package appstate

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// newResumingInProgress wires a resumingInProgressApp against the fake client
// with an unreachable kube config (the event tier fails fast, matching the
// other resume tests).
func newResumingInProgress(t *testing.T, am *appv1alpha1.ApplicationManager) *resumingInProgressApp {
	t.Helper()
	dep, pod := pendingWorkload(am.Spec.AppName, am.Spec.AppNamespace)
	c := newFakeClient(t, dep, pod, am)
	deps, tf := newTestDeps(c)
	tf.kubeConfig = unreachableKubeConfig()

	rp := &ResumingApp{
		baseOperationApp: &baseOperationApp{
			ttl:             time.Hour,
			baseStatefulApp: &baseStatefulApp{manager: am, client: c, deps: deps},
		},
	}
	return &resumingInProgressApp{
		ResumingApp:                       rp,
		basePollableStatefulInProgressApp: &basePollableStatefulInProgressApp{},
	}
}

// TestResume_Poll_CanceledReturnsContextCanceled pins the fix for the
// ResumingCanceling -> ResumeFailed race: when the poll context is canceled
// (cancel-by-timeout moves the app to ResumingCanceling and stops polling),
// poll must report the context error so WaitAsync skips the ResumeFailed write
// that the transition guard would otherwise reject.
func TestResume_Poll_CanceledReturnsContextCanceled(t *testing.T) {
	am := buildAM("demo", appv1alpha1.App, appv1alpha1.Resuming, appv1alpha1.ResumeOp, "")
	ip := newResumingInProgress(t, am)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // simulate Cleanup -> stopPolling cancelling the poll context

	err := ip.poll(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

// TestResume_Poll_StopRequestedSurvivesCanceledCtx verifies the detached-context
// annotation read: even when the poll context is canceled, a pending-pod stop
// request is still detected (errStopRequestedDueToPendingPod) rather than being
// masked by the cancellation.
func TestResume_Poll_StopRequestedSurvivesCanceledCtx(t *testing.T) {
	am := buildAM("demo", appv1alpha1.App, appv1alpha1.Resuming, appv1alpha1.ResumeOp, "")
	am.Annotations = map[string]string{api.AppStopByControllerDuePendingPod: "true"}
	ip := newResumingInProgress(t, am)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := ip.poll(ctx)
	if !errors.Is(err, errStopRequestedDueToPendingPod) {
		t.Fatalf("expected errStopRequestedDueToPendingPod, got %v", err)
	}
}
