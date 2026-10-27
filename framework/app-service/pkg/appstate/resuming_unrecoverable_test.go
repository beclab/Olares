package appstate

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/podhealth"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

// unreachableKubeConfig yields a config that kubernetes.NewForConfig accepts
// (non-empty Host) but whose every request fails immediately, so the
// event-based mount tier is effectively disabled without hanging the test.
func unreachableKubeConfig() *rest.Config {
	return &rest.Config{Host: "https://127.0.0.1:1"}
}

func TestResume_IsStartUp_FastFailsOnUnrecoverablePod(t *testing.T) {
	// Shorten the grace so the poll loop trips on the first detection tick.
	oldGrace := podhealth.HardErrorGrace
	podhealth.HardErrorGrace = 0
	defer func() { podhealth.HardErrorGrace = oldGrace }()

	am := buildAM("demo", appv1alpha1.App, appv1alpha1.Resuming, appv1alpha1.ResumeOp, "")
	dep, pod := pendingWorkload("demo", "demo-ns")
	// Make the pod's container unrecoverable (missing ConfigMap/Secret volume
	// surfaces as CreateContainerConfigError on the container).
	pod.Status.ContainerStatuses[0].State = corev1.ContainerState{
		Waiting: &corev1.ContainerStateWaiting{Reason: "CreateContainerConfigError", Message: "configmap \"x\" not found"},
	}
	c := newFakeClient(t, dep, pod, am)
	deps, tf := newTestDeps(c)
	// Unreachable apiserver: the clientset builds, but the mount tier's event
	// List fails fast so only the container-status tier is exercised.
	tf.kubeConfig = unreachableKubeConfig()

	rp := &ResumingApp{
		baseOperationApp: &baseOperationApp{
			ttl:             time.Hour,
			baseStatefulApp: &baseStatefulApp{manager: am, client: c, deps: deps},
		},
	}
	ip := &resumingInProgressApp{
		ResumingApp:                       rp,
		basePollableStatefulInProgressApp: &basePollableStatefulInProgressApp{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ok, err := ip.IsStartUp(ctx)
	if ok {
		t.Fatal("expected IsStartUp to report not started")
	}
	if err == nil || !errors.Is(err, errUnrecoverablePod) {
		t.Fatalf("expected errUnrecoverablePod, got %v", err)
	}
}

func TestResume_IsStartUp_ReturnsOnCancelWithoutError(t *testing.T) {
	am := buildAM("demo", appv1alpha1.App, appv1alpha1.Resuming, appv1alpha1.ResumeOp, "")
	dep, pod := pendingWorkload("demo", "demo-ns") // healthy-but-not-started
	c := newFakeClient(t, dep, pod, am)
	deps, tf := newTestDeps(c)
	tf.kubeConfig = unreachableKubeConfig()

	rp := &ResumingApp{
		baseOperationApp: &baseOperationApp{
			ttl:             time.Hour,
			baseStatefulApp: &baseStatefulApp{manager: am, client: c, deps: deps},
		},
	}
	ip := &resumingInProgressApp{
		ResumingApp:                       rp,
		basePollableStatefulInProgressApp: &basePollableStatefulInProgressApp{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	ok, err := ip.IsStartUp(ctx)
	if ok {
		t.Fatal("expected not started")
	}
	if err != nil {
		t.Fatalf("expected nil error on ctx cancel, got %v", err)
	}
}
