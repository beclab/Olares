package appstate

import (
	"testing"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	kappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// readyWorkload returns a Deployment plus a single Pod whose container
// reports Started=true, i.e. a workload isStartUp considers ready. The pod
// carries the deployment's selector labels so the label query matches.
func readyWorkload(name, ns string) (*kappsv1.Deployment, *corev1.Pod) {
	labels := map[string]string{"app": name}
	dep := &kappsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: kappsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: labels},
		},
	}
	started := true
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name + "-pod", Namespace: ns, Labels: labels},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "main"}}},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{{Name: "main", Started: &started}},
		},
	}
	return dep, pod
}

// pendingWorkload mirrors readyWorkload but the container is not yet started.
func pendingWorkload(name, ns string) (*kappsv1.Deployment, *corev1.Pod) {
	dep, pod := readyWorkload(name, ns)
	notStarted := false
	pod.Status.ContainerStatuses[0].Started = &notStarted
	return dep, pod
}

func TestIsStartUp_ReadyDeployment(t *testing.T) {
	am := buildAM("demo", appv1alpha1.App, appv1alpha1.Initializing, appv1alpha1.InstallOp, "")
	dep, pod := readyWorkload("demo", "demo-ns")
	c := newFakeClient(t, dep, pod)

	ok, err := isStartUp(am, c)
	if err != nil {
		t.Fatalf("isStartUp error: %v", err)
	}
	if !ok {
		t.Fatalf("expected ready workload to report started")
	}
}

func TestIsStartUp_PodNotStarted(t *testing.T) {
	am := buildAM("demo", appv1alpha1.App, appv1alpha1.Initializing, appv1alpha1.InstallOp, "")
	dep, pod := pendingWorkload("demo", "demo-ns")
	c := newFakeClient(t, dep, pod)

	ok, err := isStartUp(am, c)
	if err != nil {
		t.Fatalf("isStartUp error: %v", err)
	}
	if ok {
		t.Fatalf("expected not-started pod to report not ready")
	}
}

func TestIsStartUp_NoPods(t *testing.T) {
	am := buildAM("demo", appv1alpha1.App, appv1alpha1.Initializing, appv1alpha1.InstallOp, "")
	dep, _ := readyWorkload("demo", "demo-ns")
	c := newFakeClient(t, dep)

	ok, err := isStartUp(am, c)
	if err == nil {
		t.Fatalf("expected error when no pods are found")
	}
	if ok {
		t.Fatalf("expected not ready when no pods are found")
	}
}

// A completed (job-style) container is also treated as started.
func TestIsStartUp_CompletedContainerCountsAsStarted(t *testing.T) {
	am := buildAM("demo", appv1alpha1.App, appv1alpha1.Initializing, appv1alpha1.InstallOp, "")
	dep, pod := readyWorkload("demo", "demo-ns")
	notStarted := false
	pod.Status.ContainerStatuses[0].Started = &notStarted
	pod.Status.ContainerStatuses[0].State = corev1.ContainerState{
		Terminated: &corev1.ContainerStateTerminated{Reason: "Completed"},
	}
	c := newFakeClient(t, dep, pod)

	ok, err := isStartUp(am, c)
	if err != nil {
		t.Fatalf("isStartUp error: %v", err)
	}
	if !ok {
		t.Fatalf("expected completed container to count as started")
	}
}
