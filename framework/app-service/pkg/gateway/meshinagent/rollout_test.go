package meshinagent

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

func TestRolloutFingerprint(t *testing.T) {
	fp := RolloutFingerprint(true, "litellm", "3")
	if fp != "inject=true|edges=litellm|epoch=3" {
		t.Fatalf("fp=%s", fp)
	}
}

func TestRolloutQueuePromotedKeyAcquiresWithoutExtraSlot(t *testing.T) {
	q := NewRolloutQueue(2)
	if !q.TryAcquire("a") || !q.TryAcquire("b") {
		t.Fatal("expected first two acquires")
	}
	if q.TryAcquire("c") {
		t.Fatal("third acquire must wait")
	}
	if _, ok := q.Release(); !ok {
		t.Fatal("release must promote the waiting key")
	}
	if !q.TryAcquire("c") {
		t.Fatal("promoted key must be able to start")
	}
	if q.ActiveCount() != 2 {
		t.Fatalf("promoted key double counted: active=%d", q.ActiveCount())
	}
	if q.WaitingCount() != 0 {
		t.Fatalf("waiting=%d", q.WaitingCount())
	}
}

func TestBumpWorkloadMissingIsRetryable(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	err := BumpWorkload(context.Background(), c, WorkloadRef{Kind: "Deployment", Namespace: "ns", Name: "gone"})
	if !errors.Is(err, ErrWorkloadGone) {
		t.Fatalf("err=%v, want ErrWorkloadGone", err)
	}
}

func TestRolloutCompleteRequiresUpdatedReadyReplicas(t *testing.T) {
	one := int32(1)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Generation: 2},
		Spec:       appsv1.DeploymentSpec{Replicas: &one},
		Status: appsv1.DeploymentStatus{
			ObservedGeneration: 1,
			Replicas:           1,
			UpdatedReplicas:    1,
			ReadyReplicas:      1,
		},
	}
	if deploymentRolloutComplete(dep) {
		t.Fatal("stale observedGeneration must not count as complete")
	}
	dep.Status.ObservedGeneration = 2
	if !deploymentRolloutComplete(dep) {
		t.Fatal("updated and ready replicas must count as complete")
	}
	dep.Status.ReadyReplicas = 0
	if deploymentRolloutComplete(dep) {
		t.Fatal("unready replacement must not count as complete")
	}

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Generation: 1},
		Spec:       appsv1.StatefulSetSpec{Replicas: &one},
		Status: appsv1.StatefulSetStatus{
			ObservedGeneration: 1,
			UpdatedReplicas:    1,
			ReadyReplicas:      1,
			CurrentRevision:    "rev-1",
			UpdateRevision:     "rev-2",
		},
	}
	if statefulSetRolloutComplete(sts) {
		t.Fatal("pending revision must not count as complete")
	}
	sts.Status.CurrentRevision = "rev-2"
	if !statefulSetRolloutComplete(sts) {
		t.Fatal("settled revision with ready replicas must count as complete")
	}
}

func TestFilterWorkloadsNeedingInject(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	mkDeploy := func(name string) *appsv1.Deployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "ns"},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": name}},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": name}},
					Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c", Image: "x"}}},
				},
			},
		}
	}
	mkPod := func(name string, containers ...string) *corev1.Pod {
		ctrs := []corev1.Container{{Name: "c", Image: "x"}}
		for _, cn := range containers {
			ctrs = append(ctrs, corev1.Container{Name: cn, Image: "x"})
		}
		return &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: name + "-0", Namespace: "ns", Labels: map[string]string{"app": name}},
			Spec:       corev1.PodSpec{Containers: ctrs},
		}
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
		mkDeploy("bare"), mkPod("bare"),
		mkDeploy("meshed"), mkPod("meshed", ContainerName, linkerdProxyContainerName),
		mkDeploy("noPods"),
	).Build()

	refs := []WorkloadRef{
		{Kind: "Deployment", Namespace: "ns", Name: "bare"},
		{Kind: "Deployment", Namespace: "ns", Name: "meshed"},
		{Kind: "Deployment", Namespace: "ns", Name: "noPods"},
	}
	got := FilterWorkloadsNeedingInject(context.Background(), c, refs, SharedCallerInject(true))
	if len(got) != 1 || got[0].Name != "bare" {
		t.Fatalf("inject filter got=%v", got)
	}
	got = FilterWorkloadsNeedingInject(context.Background(), c, refs, SharedCallerInject(false))
	if len(got) != 1 || got[0].Name != "meshed" {
		t.Fatalf("opt-out filter got=%v", got)
	}
	got = FilterWorkloadsNeedingInject(context.Background(), c, refs, LinkerdOnlyInject())
	if len(got) != 1 || got[0].Name != "bare" {
		t.Fatalf("linkerd-only filter got=%v", got)
	}
}

func TestRunJobRecordsSuccessOnlyAfterRolloutCompletes(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	_ = appv1alpha1.AddToScheme(scheme)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "web", Namespace: "ns1"},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "web"}},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c", Image: "x"}}},
			},
		},
	}
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "web-ns1", Annotations: map[string]string{AnnotDecide: "true"}},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "web",
			Namespace: "ns1",
			Settings:  map[string]string{AnnotDecide: "true"},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(dep, app).Build()
	w := NewRolloutWorker(c, NewRolloutQueue(2))
	w.sleepFn = func(context.Context, time.Duration) {}

	job := RolloutJob{
		Key:           "app/web-ns1",
		Reason:        ReasonDecideFalseToTrue,
		AppObjectName: "web-ns1",
		Workloads:     []WorkloadRef{{Kind: "Deployment", Namespace: "ns1", Name: "web"}},
		Fingerprint:   RolloutFingerprint(true, "", "1"),
	}

	bumps := 0
	w.completeFn = func(context.Context, WorkloadRef) (bool, error) { return false, nil }
	w.runJob(context.Background(), job)
	var got appv1alpha1.Application
	if err := c.Get(context.Background(), types.NamespacedName{Name: "web-ns1"}, &got); err != nil {
		t.Fatal(err)
	}
	if got.Annotations[AnnotRolloutStatus] != ErrRolloutFailed {
		t.Fatalf("incomplete rollout status=%s", got.Annotations[AnnotRolloutStatus])
	}
	if got.Annotations[AnnotRolloutFingerprint] == job.Fingerprint {
		t.Fatal("incomplete rollout must not record the fingerprint")
	}

	w.completeFn = func(context.Context, WorkloadRef) (bool, error) {
		bumps++
		return true, nil
	}
	w.runJob(context.Background(), job)
	if err := c.Get(context.Background(), types.NamespacedName{Name: "web-ns1"}, &got); err != nil {
		t.Fatal(err)
	}
	if got.Annotations[AnnotRolloutStatus] != RolloutStatusOK ||
		got.Annotations[AnnotRolloutFingerprint] != job.Fingerprint {
		t.Fatalf("status=%s fp=%s", got.Annotations[AnnotRolloutStatus], got.Annotations[AnnotRolloutFingerprint])
	}
	if bumps != 1 {
		t.Fatalf("completion checked %d times, want 1", bumps)
	}
}

func TestBumpDeployment(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "web", Namespace: "ns1"},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "web"}},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c", Image: "x"}}},
			},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(dep).Build()
	if err := BumpWorkload(context.Background(), c, WorkloadRef{Kind: "Deployment", Namespace: "ns1", Name: "web"}); err != nil {
		t.Fatal(err)
	}
	var got appsv1.Deployment
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "ns1", Name: "web"}, &got); err != nil {
		t.Fatal(err)
	}
	if got.Spec.Template.Annotations[RestartedAtAnnotation] == "" {
		t.Fatal("missing restartedAt")
	}
	if got.Spec.Strategy.RollingUpdate == nil || got.Spec.Strategy.RollingUpdate.MaxUnavailable == nil {
		t.Fatal("missing maxUnavailable")
	}
	if got.Spec.Strategy.RollingUpdate.MaxUnavailable.Type != intstr.Int ||
		got.Spec.Strategy.RollingUpdate.MaxUnavailable.IntVal != 0 {
		t.Fatalf("maxUnavailable=%v", got.Spec.Strategy.RollingUpdate.MaxUnavailable)
	}
}

func TestListTargets(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = appv1alpha1.AddToScheme(scheme)

	sharedNS := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "shared-litellm",
			Labels: map[string]string{
				security.NamespaceSharedLabel: "true",
			},
			Annotations: map[string]string{
				mesh.LinkerdInjectAnnotation: mesh.LinkerdInjectEnabled,
			},
		},
	}
	sharedDep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "litellm", Namespace: "shared-litellm"},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "litellm"}},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c", Image: "x"}}},
			},
		},
	}
	egDep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "envoy-dp", Namespace: EGDataplaneNamespace},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{mesh.LinkerdInjectAnnotation: mesh.LinkerdInjectEnabled},
				},
				Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "c", Image: "x"}}},
			},
		},
	}
	appDep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "jellyfin",
			Namespace: "jellyfin-u",
			Labels:    map[string]string{constants.ApplicationNameLabel: "jellyfin"},
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "jellyfin"}},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c", Image: "x"}}},
			},
		},
	}
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "jellyfin-jellyfin-u"},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "jellyfin",
			Namespace: "jellyfin-u",
			Settings:  map[string]string{AnnotDecide: "true"},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(sharedNS, sharedDep, egDep, appDep, app).Build()

	refs, err := ListAppWorkloads(context.Background(), c, "jellyfin-u", "jellyfin")
	if err != nil || len(refs) != 1 || refs[0].Name != "jellyfin" {
		t.Fatalf("app refs=%v err=%v", refs, err)
	}
	shared, err := ListSharedInjectWorkloads(context.Background(), c)
	if err != nil || len(shared) != 1 {
		t.Fatalf("shared=%v err=%v", shared, err)
	}
	eg, err := ListEGDataplaneWorkloads(context.Background(), c)
	if err != nil || len(eg) != 1 {
		t.Fatalf("eg=%v err=%v", eg, err)
	}
	apps, err := ListDecideTrueApplications(context.Background(), c)
	if err != nil || len(apps) != 1 {
		t.Fatalf("apps=%d err=%v", len(apps), err)
	}
}

func TestEnqueueIdempotentAndFailedNoRollback(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = appv1alpha1.AddToScheme(scheme)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web",
			Namespace: "ns1",
			Labels:    map[string]string{constants.ApplicationNameLabel: "web"},
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "web"}},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c", Image: "x"}}},
			},
		},
	}
	fp := RolloutFingerprint(true, "", "1")
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name: "web-ns1",
			Annotations: map[string]string{
				AnnotDecide:             "true",
				AnnotRolloutFingerprint: fp,
				AnnotRolloutStatus:      RolloutStatusOK,
			},
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "web",
			Namespace: "ns1",
			Settings:  map[string]string{AnnotDecide: "true"},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(dep, app).Build()
	w := NewRolloutWorker(c, NewRolloutQueue(2))
	w.sleepFn = func(context.Context, time.Duration) {}

	w.EnqueueAppRollout(context.Background(), app, ReasonDecideFalseToTrue, "1")
	w.mu.Lock()
	pending := len(w.pending)
	w.mu.Unlock()
	if pending != 0 {
		t.Fatalf("expected idempotent skip, pending=%d", pending)
	}

	// Failure path: mark failed must not clear decide annotation.
	app.Annotations[AnnotRolloutFingerprint] = ""
	app.Annotations[AnnotRolloutStatus] = ""
	_ = c.Update(context.Background(), app)
	job := RolloutJob{
		Key:           "app/web-ns1",
		Reason:        ReasonDecideFalseToTrue,
		AppObjectName: "web-ns1",
		Workloads:     []WorkloadRef{{Kind: "Deployment", Namespace: "missing", Name: "nope"}},
		Fingerprint:   RolloutFingerprint(true, "", "2"),
	}
	// Force bump failure by using a client that cannot find the deploy — Bump returns nil on NotFound.
	// Use a closed context after first attempt simulation via markFailed directly.
	w.markFailed(context.Background(), job)
	var got appv1alpha1.Application
	if err := c.Get(context.Background(), types.NamespacedName{Name: "web-ns1"}, &got); err != nil {
		t.Fatal(err)
	}
	if got.Annotations[AnnotDecide] != "true" {
		t.Fatal("decide fact must not be rolled back")
	}
	if got.Annotations[AnnotRolloutStatus] != ErrRolloutFailed {
		t.Fatalf("status=%s", got.Annotations[AnnotRolloutStatus])
	}
}

func TestWorkerProcessesUnderK(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)

	mk := func(name string) *appsv1.Deployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "ns"},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": name}},
					Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c", Image: "x"}}},
				},
			},
		}
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(mk("da"), mk("db"), mk("dc")).Build()
	q := NewRolloutQueue(2)
	w := NewRolloutWorker(c, q)
	w.sleepFn = func(context.Context, time.Duration) {}
	w.completeFn = func(context.Context, WorkloadRef) (bool, error) { return true, nil }
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w.Start(ctx)

	var mu sync.Mutex
	maxActive := 0
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 40; i++ {
			mu.Lock()
			if a := q.ActiveCount(); a > maxActive {
				maxActive = a
			}
			mu.Unlock()
			time.Sleep(5 * time.Millisecond)
		}
	}()

	for _, name := range []string{"da", "db", "dc"} {
		w.Enqueue(RolloutJob{
			Key:       "job-" + name,
			Reason:    ReasonMeshReady,
			Workloads: []WorkloadRef{{Kind: "Deployment", Namespace: "ns", Name: name}},
		})
	}
	time.Sleep(200 * time.Millisecond)
	<-done
	mu.Lock()
	defer mu.Unlock()
	if maxActive > 2 {
		t.Fatalf("maxActive=%d exceeds K=2", maxActive)
	}
}

func TestNextMeshReadyEpoch(t *testing.T) {
	if NextMeshReadyEpoch("0") != "1" || NextMeshReadyEpoch("9") != "10" {
		t.Fatal(NextMeshReadyEpoch("0"), NextMeshReadyEpoch("9"))
	}
	if NextMeshReadyEpoch("not-a-number") != "1" {
		t.Fatalf("garbage epoch=%s", NextMeshReadyEpoch("not-a-number"))
	}
}

func TestEnqueueSweepSplitsDataplanePerWorkload(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = appv1alpha1.AddToScheme(scheme)

	mkDeploy := func(ns, name string, labels, tplAnnotations map[string]string) *appsv1.Deployment {
		podLabels := map[string]string{"app": name}
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns, Labels: labels},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{MatchLabels: podLabels},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: podLabels, Annotations: tplAnnotations},
					Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c", Image: "x"}}},
				},
			},
		}
	}
	mkPod := func(ns, name string) *corev1.Pod {
		return &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: name + "-0", Namespace: ns, Labels: map[string]string{"app": name}},
			Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c", Image: "x"}}},
		}
	}
	injectAnnotation := map[string]string{mesh.LinkerdInjectAnnotation: mesh.LinkerdInjectEnabled}
	sharedNS := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "shared-litellm",
			Labels:      map[string]string{security.NamespaceSharedLabel: "true"},
			Annotations: injectAnnotation,
		},
	}
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "jellyfin-u", Annotations: map[string]string{AnnotDecide: "true"}},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "jellyfin",
			Namespace: "jellyfin-u",
			Settings:  map[string]string{AnnotDecide: "true"},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
		sharedNS,
		mkDeploy("shared-litellm", "litellm", nil, nil), mkPod("shared-litellm", "litellm"),
		mkDeploy("shared-litellm", "litellm-worker", nil, nil), mkPod("shared-litellm", "litellm-worker"),
		mkDeploy(EGDataplaneNamespace, "envoy-dp", nil, injectAnnotation), mkPod(EGDataplaneNamespace, "envoy-dp"),
		mkDeploy("jellyfin-u", "jellyfin", map[string]string{constants.ApplicationNameLabel: "jellyfin"}, nil),
		mkPod("jellyfin-u", "jellyfin"),
		app,
	).Build()

	w := NewRolloutWorker(c, NewRolloutQueue(2))
	w.EnqueueMeshReadySweep(context.Background(), "1")

	w.mu.Lock()
	keys := make([]string, 0, len(w.pending))
	for k, job := range w.pending {
		keys = append(keys, k)
		if len(job.Workloads) != 1 {
			t.Fatalf("job %s bundles %d workloads, want 1", k, len(job.Workloads))
		}
	}
	w.mu.Unlock()
	if len(keys) != 4 {
		t.Fatalf("pending keys=%v, want one app job plus three dataplane jobs", keys)
	}
}

func TestMaybeEnqueueAfterDecideOnlyOnIntentChange(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = appv1alpha1.AddToScheme(scheme)

	podLabels := map[string]string{"app": "web"}
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web",
			Namespace: "ns1",
			Labels:    map[string]string{constants.ApplicationNameLabel: "web"},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: podLabels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: podLabels},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c", Image: "x"}}},
			},
		},
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "web-0", Namespace: "ns1", Labels: podLabels},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c", Image: "x"}}},
	}
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "web-ns1"},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "web",
			Namespace: "ns1",
			Settings:  map[string]string{AnnotDecide: "true"},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(dep, pod, app).Build()
	w := NewRolloutWorker(c, NewRolloutQueue(2))

	prev := DefaultWorker()
	SetDefaultWorker(w)
	defer SetDefaultWorker(prev)

	MaybeEnqueueAfterDecide(context.Background(), app, true, true, "litellm", "litellm", "")
	w.mu.Lock()
	unchanged := len(w.pending)
	w.mu.Unlock()
	if unchanged != 0 {
		t.Fatalf("unchanged intent enqueued %d jobs", unchanged)
	}

	EnqueueCreateIfInject(context.Background(), app)
	w.mu.Lock()
	created := len(w.pending)
	w.mu.Unlock()
	if created != 1 {
		t.Fatalf("newly declared caller enqueued %d jobs, want 1", created)
	}
}

func TestWorkloadRolloutCompleteReportsMissingWorkload(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	if _, err := WorkloadRolloutComplete(context.Background(), c,
		WorkloadRef{Kind: "StatefulSet", Namespace: "ns", Name: "gone"}); !errors.Is(err, ErrWorkloadGone) {
		t.Fatalf("err=%v, want ErrWorkloadGone", err)
	}
}

func TestEnsureSurgeFirstKeepsRecreateStrategy(t *testing.T) {
	dep := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Strategy: appsv1.DeploymentStrategy{Type: appsv1.RecreateDeploymentStrategyType},
		},
	}
	ensureSurgeFirstStrategy(dep)
	if dep.Spec.Strategy.Type != appsv1.RecreateDeploymentStrategyType || dep.Spec.Strategy.RollingUpdate != nil {
		t.Fatal("Recreate strategy must be left alone")
	}
}

func TestMeshInjectStateRoundTrip(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	if err := StoreMeshInjectRolloutState(context.Background(), c, true, "2"); err != nil {
		t.Fatal(err)
	}
	ready, epoch, err := LoadMeshInjectRolloutState(context.Background(), c)
	if err != nil || !ready || epoch != "2" {
		t.Fatalf("ready=%v epoch=%s err=%v", ready, epoch, err)
	}
}
