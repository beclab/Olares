package meshinagent

import (
	"context"
	"reflect"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNativeMeshInPrecedesApplicationInit(t *testing.T) {
	always := corev1.ContainerRestartPolicyAlways
	for _, linkerdFirst := range []bool{false, true} {
		pod := &corev1.Pod{Spec: corev1.PodSpec{InitContainers: []corev1.Container{{Name: "app-setup"}}}}
		linkerd := []corev1.Container{{Name: "linkerd-init"}, {Name: mesh.LinkerdProxyContainer, RestartPolicy: &always}}
		if linkerdFirst {
			pod.Spec.InitContainers = append(linkerd, pod.Spec.InitContainers...)
		}
		InsertNativeMeshIn(pod, "10.42.0.1")
		if !linkerdFirst {
			// The upstream injector adds the native pair at positions 0 and 1.
			pod.Spec.InitContainers = append(linkerd, pod.Spec.InitContainers...)
		}
		var names []string
		for _, container := range pod.Spec.InitContainers {
			names = append(names, container.Name)
		}
		want := []string{"linkerd-init", mesh.LinkerdProxyContainer, InitContainerName, ContainerName, "app-setup"}
		if !reflect.DeepEqual(names, want) {
			t.Fatalf("init sequence %v, want %v", names, want)
		}
		if !HasNativeMeshIn(pod) || pod.Spec.InitContainers[3].StartupProbe == nil || pod.Spec.InitContainers[3].StartupProbe.Exec == nil {
			t.Fatal("mesh-in must run as a restartable init with a local startup probe")
		}
	}
}

func TestNativeMigrationRebuildsInstalledAppOnce(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	for _, add := range []func(*runtime.Scheme) error{appsv1.AddToScheme, corev1.AddToScheme, appv1alpha1.AddToScheme} {
		if err := add(scheme); err != nil {
			t.Fatal(err)
		}
	}
	const ns = "photos-alice"
	dep := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "photos", UID: "dep"}}
	dep.Spec.Template.Annotations = map[string]string{"existing": "keep"}
	rs := &appsv1.ReplicaSet{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "photos-old", UID: "rs",
		OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(dep, appsv1.SchemeGroupVersion.WithKind("Deployment"))}}}
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "photos-old-1",
		OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(rs, appsv1.SchemeGroupVersion.WithKind("ReplicaSet"))}}}
	app := &appv1alpha1.ApplicationManager{ObjectMeta: metav1.ObjectMeta{Name: "photos-app"},
		Spec: appv1alpha1.ApplicationManagerSpec{Type: appv1alpha1.App, AppName: "photos", AppOwner: "alice", AppNamespace: ns}}
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(namespace, app, dep, rs, pod).Build()
	w := NewRolloutWorker(c, NewRolloutQueue(1))
	w.completeFn = func(context.Context, WorkloadRef) (bool, error) { return true, nil }
	if err := w.EnqueueNativeSidecarMigration(ctx); err != nil {
		t.Fatal(err)
	}
	job, ok := w.dequeue()
	if !ok || job.Reason != ReasonNativeSidecars || len(job.Workloads) != 1 || job.Workloads[0].Name != dep.Name {
		t.Fatalf("expected one migration for the legacy app, got %+v", job)
	}
	w.runJob(ctx, job)
	var got appsv1.Deployment
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: dep.Name}, &got); err != nil {
		t.Fatal(err)
	}
	if got.Spec.Template.Annotations["existing"] != "keep" || got.Spec.Template.Annotations[nativeSidecarMigrationAnnotation] == "" {
		t.Fatalf("migration marker missing or existing annotations lost: %v", got.Spec.Template.Annotations)
	}
	if err := w.EnqueueNativeSidecarMigration(ctx); err != nil {
		t.Fatal(err)
	}
	if _, ok := w.dequeue(); ok {
		t.Fatal("still-visible old replicas must not trigger another migration")
	}
}

func TestInstalledAppOwnershipAndMigrationOptOut(t *testing.T) {
	app := &appv1alpha1.ApplicationManager{Spec: appv1alpha1.ApplicationManagerSpec{
		Type: appv1alpha1.App, AppNamespace: "photos-alice", AppName: "photos", AppOwner: "alice",
	}}
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "photos-worker",
		Labels: map[string]string{constants.ApplicationNameLabel: "photos", constants.ApplicationInstallUserLabel: "bob"}}}
	if InstalledAppOwnsNamespace(ns, app) {
		t.Fatal("a subchart installed by a different owner must not qualify")
	}
	ns.Labels[constants.ApplicationInstallUserLabel] = "alice"
	if !InstalledAppOwnsNamespace(ns, app) {
		t.Fatal("the matching app subchart must qualify")
	}
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{nativeSidecarOptOut: "false"}},
		Spec: corev1.PodSpec{
			Containers:     []corev1.Container{{Name: mesh.LinkerdProxyContainer}},
			InitContainers: []corev1.Container{{Name: ContainerName, RestartPolicy: ptr.To(corev1.ContainerRestartPolicyAlways)}},
		}}
	if needsNativeSidecars(pod, true) {
		t.Fatal("do not restart a migrated app solely for an explicit Linkerd opt-out")
	}
}
