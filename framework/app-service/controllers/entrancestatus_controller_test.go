package controllers

import (
	"testing"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func entranceNativeSidecarPod() *corev1.Pod {
	completed := corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 0}}
	running := corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}
	return &corev1.Pod{
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{
				{Name: "linkerd-init"},
				{Name: "linkerd-proxy", RestartPolicy: ptr.To(corev1.ContainerRestartPolicyAlways)},
				{Name: "copy-frontend"},
			},
			Containers: []corev1.Container{{Name: "frontend"}},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			InitContainerStatuses: []corev1.ContainerStatus{
				{Name: "linkerd-init", State: completed},
				{Name: "linkerd-proxy", State: running, Started: ptr.To(true), Ready: true},
				{Name: "copy-frontend", State: completed},
			},
			ContainerStatuses: []corev1.ContainerStatus{{Name: "frontend", State: running, Ready: true}},
		},
	}
}

func TestMakeEntranceStateWithNativeSidecar(t *testing.T) {
	for _, tc := range []struct {
		name    string
		edit    func(*corev1.Pod)
		want    appv1alpha1.EntranceState
		reason  string
		message string
	}{
		{name: "started and ready sidecar", want: appv1alpha1.EntranceRunning, reason: "Running"},
		{name: "sidecar startup still pending", edit: func(p *corev1.Pod) {
			p.Status.InitContainerStatuses[1].Started = ptr.To(false)
		}, want: appv1alpha1.EntranceNotReady, reason: "Init:1/3"},
		{name: "sidecar readiness probe failing", edit: func(p *corev1.Pod) {
			p.Status.InitContainerStatuses[1].Ready = false
		}, want: appv1alpha1.EntranceNotReady, reason: "Init:1/3"},
		{name: "sidecar restarting", edit: func(p *corev1.Pod) {
			p.Status.InitContainerStatuses[1].State = corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"}}
		}, want: appv1alpha1.EntranceNotReady, reason: "Init:CrashLoopBackOff"},
		{name: "sidecar termination preserves signal and message", edit: func(p *corev1.Pod) {
			p.Status.InitContainerStatuses[1].State = corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{Signal: 9, Message: "terminated"}}
		}, want: appv1alpha1.EntranceNotReady, reason: "Init:Signal:9", message: "terminated"},
		{name: "sidecar waiting preserves message", edit: func(p *corev1.Pod) {
			p.Status.InitContainerStatuses[1].State = corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "ImagePullBackOff", Message: "pull failed"}}
		}, want: appv1alpha1.EntranceNotReady, reason: "Init:ImagePullBackOff", message: "pull failed"},
		{name: "sidecar exited successfully but must restart", edit: func(p *corev1.Pod) {
			p.Status.InitContainerStatuses[1].State = corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{Reason: "Completed"}}
		}, want: appv1alpha1.EntranceNotReady, reason: "Init:Completed"},
		{name: "sidecar status missing", edit: func(p *corev1.Pod) {
			p.Status.InitContainerStatuses = append(p.Status.InitContainerStatuses[:1], p.Status.InitContainerStatuses[2:]...)
		}, want: appv1alpha1.EntranceNotReady, reason: "Init:1/3"},
		{name: "regular init still running after ready sidecar", edit: func(p *corev1.Pod) {
			p.Status.InitContainerStatuses[2].State = corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}
		}, want: appv1alpha1.EntranceNotReady, reason: "Init:2/3"},
		{name: "regular init failed", edit: func(p *corev1.Pod) {
			p.Status.InitContainerStatuses[2].State = corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 7}}
		}, want: appv1alpha1.EntranceNotReady, reason: "Init:ExitCode:7"},
		{name: "application not ready after successful initialization", edit: func(p *corev1.Pod) {
			p.Status.ContainerStatuses[0].Ready = false
		}, want: appv1alpha1.EntranceNotReady, reason: "Running"},
		{name: "frontend crash still blocks a ready proxy", edit: func(p *corev1.Pod) {
			p.Status.ContainerStatuses[0].State = corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff", Message: "frontend restarting"}}
		}, want: appv1alpha1.EntranceNotReady, reason: "CrashLoopBackOff", message: "frontend restarting"},
		{name: "traditional pod with no sidecar", edit: func(p *corev1.Pod) {
			p.Spec.InitContainers = p.Spec.InitContainers[2:]
			p.Status.InitContainerStatuses = p.Status.InitContainerStatuses[2:]
		}, want: appv1alpha1.EntranceRunning, reason: "Running"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pod := entranceNativeSidecarPod()
			if tc.edit != nil {
				tc.edit(pod)
			}
			got, reason, message := makeEntranceState(pod)
			if got != tc.want || reason != tc.reason || message != tc.message {
				t.Fatalf("entrance state = %s (%s: %s), want %s (%s: %s)", got, reason, message, tc.want, tc.reason, tc.message)
			}
		})
	}
}

func TestEntranceReconcileRecoversAfterNativeSidecarBecomesReady(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := appv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	pod := entranceNativeSidecarPod()
	pod.ObjectMeta = metav1.ObjectMeta{Name: "frontend", Namespace: "user-space-alice", Labels: map[string]string{"app": "frontend"}}
	pod.Status.InitContainerStatuses[1].Ready = false
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "frontend", Namespace: pod.Namespace},
		Spec:       corev1.ServiceSpec{Selector: map[string]string{"app": "frontend"}},
	}
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "frontend"},
		Spec: appv1alpha1.ApplicationSpec{
			Namespace: pod.Namespace,
			Entrances: []appv1alpha1.Entrance{{Name: "desktop", Host: svc.Name}, {Name: "files", Host: svc.Name}},
		},
		Status: appv1alpha1.ApplicationStatus{
			State: "notReady",
			EntranceStatuses: []appv1alpha1.EntranceStatus{
				{Name: "desktop", State: appv1alpha1.EntranceNotReady},
				{Name: "files", State: appv1alpha1.EntranceNotReady},
			},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod, svc, app).
		WithStatusSubresource(pod, app).Build()
	controller := &EntranceStatusManagerController{Client: c}
	ctx := t.Context()
	podKey := client.ObjectKeyFromObject(pod)
	appKey := client.ObjectKeyFromObject(app)
	for _, ready := range []bool{false, true, false} {
		if err := c.Get(ctx, podKey, pod); err != nil {
			t.Fatal(err)
		}
		pod.Status.InitContainerStatuses[1].Ready = ready
		if err := c.Status().Update(ctx, pod); err != nil {
			t.Fatal(err)
		}
		if _, err := controller.Reconcile(ctx, ctrl.Request{NamespacedName: podKey}); err != nil {
			t.Fatal(err)
		}
		if err := c.Get(ctx, appKey, app); err != nil {
			t.Fatal(err)
		}
		wantApp, wantEntrance := "notReady", appv1alpha1.EntranceNotReady
		if ready {
			wantApp, wantEntrance = "running", appv1alpha1.EntranceRunning
		}
		if app.Status.State != wantApp || len(app.Status.EntranceStatuses) != 2 {
			t.Fatalf("aggregate status = %+v, want %s with both entrances", app.Status, wantApp)
		}
		for _, status := range app.Status.EntranceStatuses {
			if status.State != wantEntrance {
				t.Fatalf("entrance %s has state %s, want %s", status.Name, status.State, wantEntrance)
			}
		}
	}
}
