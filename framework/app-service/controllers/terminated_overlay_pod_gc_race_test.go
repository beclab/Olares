package controllers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

type overlayRecoveryFixture struct {
	deploy  *appsv1.Deployment
	rs      *appsv1.ReplicaSet
	pod     *corev1.Pod
	app     *appv1.Application
	node    *corev1.Node
	objects []client.Object
}

func newOverlayRecoveryFixture(now time.Time) *overlayRecoveryFixture {
	deploy, objects := terminatedPodGCFixture(now, terminatedPodGCTestCase{
		phase: corev1.PodFailed, replicas: 1, targetTerminalAge: time.Hour, noReplacement: true,
	})
	f := &overlayRecoveryFixture{deploy: deploy}
	deploy.Spec.Strategy.Type = appsv1.RecreateDeploymentStrategyType
	deploy.Spec.Template.Labels = map[string]string{constants.ApplicationNameLabel: "test-app", constants.ApplicationOwnerLabel: "alice"}
	for _, obj := range objects {
		switch obj := obj.(type) {
		case *appsv1.ReplicaSet:
			if obj.Name == "test-app-old" {
				f.rs = obj
			}
		case *corev1.Pod:
			f.pod = obj
		}
	}
	f.pod.Labels[constants.ApplicationMacvlanInitLabel] = "true"
	f.pod.Spec.NodeName = "node"
	f.pod.Spec.Containers = []corev1.Container{{Name: "app"}}
	f.pod.Status.Reason = "UnexpectedAdmissionError"
	f.pod.Status.Conditions = append(f.pod.Status.Conditions, corev1.PodCondition{
		Type: corev1.PodReadyToStartContainers, Status: corev1.ConditionFalse, LastTransitionTime: metav1.NewTime(now.Add(-time.Hour)),
	})
	f.app = &appv1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: deploy.Namespace + "-test-app", UID: "app-uid"},
		Spec: appv1.ApplicationSpec{Namespace: deploy.Namespace, Name: "test-app", Owner: "alice", DeploymentName: deploy.Name,
			Settings: map[string]string{"enableOverlayGateway": "true"}},
	}
	f.node = &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node"}, Status: corev1.NodeStatus{
		Conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionTrue}},
	}}
	f.objects = append(objects, f.app, f.node)
	return f
}

func TestOverlayRecoveryRejectsStaleCache(t *testing.T) {
	for _, scenario := range []string{"pod updated", "pod recreated", "node NotReady", "deployment recreated", "deployment scaled", "ReplicaSet recreated", "application owner changed"} {
		t.Run(scenario, func(t *testing.T) {
			now := time.Now()
			f := newOverlayRecoveryFixture(now)
			d, pod, objects := f.deploy, f.pod, f.objects
			cached, live := testutil.NewFakeClient(objects...), testutil.NewFakeClient(objects...)
			switch scenario {
			case "pod updated", "pod recreated":
				var current corev1.Pod
				require.NoError(t, live.Get(t.Context(), client.ObjectKeyFromObject(pod), &current))
				if scenario == "pod recreated" {
					current.UID = "new-pod-uid"
				} else {
					current.Labels["changed"] = "true"
				}
				require.NoError(t, live.Update(t.Context(), &current))
			case "node NotReady":
				var node corev1.Node
				require.NoError(t, live.Get(t.Context(), client.ObjectKey{Name: "node"}, &node))
				node.Status.Conditions[0].Status = corev1.ConditionFalse
				require.NoError(t, live.Status().Update(t.Context(), &node))
			case "deployment recreated", "deployment scaled":
				var current appsv1.Deployment
				require.NoError(t, live.Get(t.Context(), client.ObjectKeyFromObject(d), &current))
				if scenario == "deployment recreated" {
					current.UID = "new-deployment-uid"
				} else {
					current.Spec.Replicas = ptr.To(int32(2))
				}
				require.NoError(t, live.Update(t.Context(), &current))
			case "ReplicaSet recreated":
				var rs appsv1.ReplicaSet
				require.NoError(t, live.Get(t.Context(), client.ObjectKey{Namespace: d.Namespace, Name: "test-app-old"}, &rs))
				rs.UID = "new-rs-uid"
				require.NoError(t, live.Update(t.Context(), &rs))
			case "application owner changed":
				var app appv1.Application
				require.NoError(t, live.Get(t.Context(), client.ObjectKey{Name: d.Namespace + "-test-app"}, &app))
				app.Spec.Owner = "bob"
				require.NoError(t, live.Update(t.Context(), &app))
			}
			deletes := 0
			c := interceptor.NewClient(cached.(client.WithWatch), interceptor.Funcs{Delete: func(context.Context, client.WithWatch, client.Object, ...client.DeleteOption) error {
				deletes++
				return nil
			}})
			r := &TerminatedPodGCController{Client: c, APIReader: live, minPodAge: time.Minute}
			result, err := r.Reconcile(t.Context(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(d)})
			require.NoError(t, err)
			require.Zero(t, deletes)
			if scenario == "application owner changed" {
				require.Zero(t, result.RequeueAfter)
			} else {
				require.Positive(t, result.RequeueAfter)
			}
		})
	}
}

func TestOverlayRecoveryDeleteConflictPreservesPodAndRetries(t *testing.T) {
	now := time.Now()
	f := newOverlayRecoveryFixture(now)
	d, pod, objects := f.deploy, f.pod, f.objects
	live := testutil.NewFakeClient(objects...)
	c := interceptor.NewClient(live.(client.WithWatch), interceptor.Funcs{Delete: func(ctx context.Context, inner client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
		// Model a concurrent writer between the last read and DELETE.
		var current corev1.Pod
		require.NoError(t, inner.Get(ctx, client.ObjectKeyFromObject(pod), &current))
		current.Labels["concurrent-update"] = "true"
		require.NoError(t, inner.Update(ctx, &current))
		return inner.Delete(ctx, obj, opts...)
	}})
	r := &TerminatedPodGCController{Client: c, APIReader: live, minPodAge: time.Minute}
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(d)}
	_, err := r.Reconcile(t.Context(), req)
	require.True(t, apierrors.IsConflict(err), "stale DELETE must be rejected: %v", err)
	var kept corev1.Pod
	require.NoError(t, live.Get(t.Context(), client.ObjectKeyFromObject(pod), &kept))
	require.Nil(t, kept.DeletionTimestamp)
	// Retry from a fresh observation, without the racing writer.
	r.Client = live
	_, err = r.Reconcile(t.Context(), req)
	require.NoError(t, err)
	require.True(t, apierrors.IsNotFound(live.Get(t.Context(), client.ObjectKeyFromObject(pod), &kept)))
}

func TestOverlayRecoveryWaitsForApplicationAndStoppedWorkloadSandbox(t *testing.T) {
	now := time.Now()
	f := newOverlayRecoveryFixture(now)
	d, pod, objects := f.deploy, f.pod, f.objects
	c := testutil.NewFakeClient(objects...)
	var app appv1.Application
	appKey := client.ObjectKey{Name: d.Namespace + "-test-app"}
	require.NoError(t, c.Get(t.Context(), appKey, &app))
	require.NoError(t, c.Delete(t.Context(), &app))
	r := &TerminatedPodGCController{Client: c, minPodAge: time.Minute}
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(d)}
	result, err := r.Reconcile(t.Context(), req)
	require.NoError(t, err)
	require.Zero(t, result.RequeueAfter)
	var kept corev1.Pod
	require.NoError(t, c.Get(t.Context(), client.ObjectKeyFromObject(pod), &kept))
	app.ResourceVersion = ""
	require.NoError(t, c.Create(t.Context(), &app))
	requests := r.requestsForOverlayApplication(t.Context(), &app)
	require.Equal(t, []ctrl.Request{req}, requests)
	_, err = r.Reconcile(t.Context(), requests[0])
	require.NoError(t, err)
	require.True(t, apierrors.IsNotFound(c.Get(t.Context(), client.ObjectKeyFromObject(pod), &kept)))

	// Even at replicas=0 the ordinary GC readiness gate must not bypass
	// the overlay sandbox guard and release a MAC during teardown.
	f = newOverlayRecoveryFixture(now)
	d, pod, objects = f.deploy, f.pod, f.objects
	d.Spec.Replicas = ptr.To(int32(0))
	pod.Status.Conditions[1].Status = corev1.ConditionTrue
	c = testutil.NewFakeClient(objects...)
	r.Client = c
	_, err = r.Reconcile(t.Context(), req)
	require.NoError(t, err)
	require.NoError(t, c.Get(t.Context(), client.ObjectKeyFromObject(pod), &kept))
}

func TestOverlayRecoveryRetainsSafetyAfterLabelRemoval(t *testing.T) {
	for _, evidence := range []string{"network selection", "default network selection", "injected init container"} {
		t.Run(evidence, func(t *testing.T) {
			now := time.Now()
			f := newOverlayRecoveryFixture(now)
			d, pod, objects := f.deploy, f.pod, f.objects
			delete(pod.Labels, constants.ApplicationMacvlanInitLabel)
			switch evidence {
			case "network selection":
				pod.Annotations = map[string]string{"k8s.v1.cni.cncf.io/networks": `[{"name":"underlay-macvlan","namespace":"kube-system","interface":"net1"}]`}
			case "default network selection":
				pod.Annotations = map[string]string{"v1.multus-cni.io/default-network": "kube-system/underlay-macvlan"}
			case "injected init container":
				pod.Spec.InitContainers = []corev1.Container{{Name: constants.MacvlanReplyInitContainerName}}
				pod.Status.InitContainerStatuses = []corev1.ContainerStatus{{Name: constants.MacvlanReplyInitContainerName, State: pod.Status.ContainerStatuses[0].State}}
			}
			d.Spec.Replicas = ptr.To(int32(0))
			pod.Status.Conditions[1].Status = corev1.ConditionTrue
			c := testutil.NewFakeClient(objects...)
			r := &TerminatedPodGCController{Client: c, minPodAge: time.Minute}
			req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(d)}
			result, err := r.Reconcile(t.Context(), req)
			require.NoError(t, err)
			var kept corev1.Pod
			require.NoError(t, c.Get(t.Context(), client.ObjectKeyFromObject(pod), &kept), "removing the label must not bypass the sandbox guard")
			require.Positive(t, result.RequeueAfter)
			// Restore the desired replica count and let teardown complete. The
			// retained network evidence must also permit eventual recovery.
			var current appsv1.Deployment
			require.NoError(t, c.Get(t.Context(), client.ObjectKeyFromObject(d), &current))
			current.Spec.Replicas = ptr.To(int32(1))
			require.NoError(t, c.Update(t.Context(), &current))
			kept.Status.Conditions[1].Status = corev1.ConditionFalse
			require.NoError(t, c.Status().Update(t.Context(), &kept))
			_, err = r.Reconcile(t.Context(), req)
			require.NoError(t, err)
			require.True(t, apierrors.IsNotFound(c.Get(t.Context(), client.ObjectKeyFromObject(pod), &kept)))
		})
	}
}

func TestTerminatedPodGCNonOverlayKeepsOriginalPolicy(t *testing.T) {
	for _, tc := range []struct {
		name        string
		replicas    int32
		ready       bool
		multus      bool
		platformMAC bool
		application bool
		wantDelete  bool
	}{
		{name: "plain Deployment scaled to zero", wantDelete: true},
		{name: "Multus Deployment with Overlay disabled scaled to zero", multus: true, application: true, wantDelete: true},
		{name: "Multus replacement is stably ready", replicas: 1, ready: true, multus: true, application: true, wantDelete: true},
		{name: "Multus still waits for a replacement", replicas: 1, multus: true, application: true},
		{name: "stale platform attachment with Overlay disabled", platformMAC: true, application: true, wantDelete: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Now()
			deploy, objects := terminatedPodGCFixture(now, terminatedPodGCTestCase{
				phase: corev1.PodFailed, replicas: tc.replicas, targetTerminalAge: time.Hour,
				replacementReady: tc.ready, replacementReadyAge: time.Hour, noReplacement: !tc.ready,
			})
			deploy.Spec.Template.Labels = map[string]string{constants.ApplicationNameLabel: "test-app", constants.ApplicationOwnerLabel: "alice"}
			if tc.application {
				objects = append(objects, &appv1.Application{
					ObjectMeta: metav1.ObjectMeta{Name: deploy.Namespace + "-test-app", UID: "app-uid"},
					Spec: appv1.ApplicationSpec{Namespace: deploy.Namespace, Name: "test-app", Owner: "alice",
						Settings: map[string]string{"enableOverlayGateway": "false"}},
				})
			}
			for _, obj := range objects {
				pod, ok := obj.(*corev1.Pod)
				if !ok || pod.Name != "old-pod" {
					continue
				}
				if tc.multus {
					pod.Annotations = map[string]string{"k8s.v1.cni.cncf.io/networks": "secondary", "v1.multus-cni.io/default-network": "underlay"}
				}
				if tc.platformMAC {
					pod.Annotations = map[string]string{"k8s.v1.cni.cncf.io/networks": `[{"name":"underlay-macvlan","namespace":"kube-system","interface":"net1"}]`}
				}
				pod.Spec.NodeName = "missing-node"
				pod.Status.Reason = "UnexpectedAdmissionError"
				pod.Status.ContainerStatuses = nil
				pod.Status.Conditions = append(pod.Status.Conditions, corev1.PodCondition{
					Type: corev1.PodReadyToStartContainers, Status: corev1.ConditionFalse, LastTransitionTime: metav1.NewTime(now),
				})
				pod.Status.EphemeralContainerStatuses = []corev1.ContainerStatus{{
					Name: "debug", State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{FinishedAt: metav1.NewTime(now)}},
				}}
			}

			backing := testutil.NewFakeClient(objects...)
			deletes := 0
			c := interceptor.NewClient(backing.(client.WithWatch), interceptor.Funcs{
				Get: func(ctx context.Context, inner client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					switch obj.(type) {
					case *corev1.Pod, *corev1.Node:
						t.Fatalf("non-Overlay cleanup requested a live runtime read: %T", obj)
					case *appv1.Application:
						if !tc.platformMAC {
							return errors.New("Application API is unavailable")
						}
					}
					return inner.Get(ctx, key, obj, opts...)
				},
				Delete: func(ctx context.Context, inner client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
					deletes++
					options := &client.DeleteOptions{}
					options.ApplyOptions(opts)
					require.NotNil(t, options.Preconditions)
					require.Equal(t, obj.GetUID(), *options.Preconditions.UID)
					require.Nil(t, options.Preconditions.ResourceVersion)
					return inner.Delete(ctx, obj, opts...)
				},
			})
			r := &TerminatedPodGCController{Client: c, APIReader: c, now: func() time.Time { return now }, minPodAge: time.Minute, readyDelay: time.Minute}
			result, err := r.Reconcile(t.Context(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(deploy)})
			require.NoError(t, err)
			require.Zero(t, result.RequeueAfter)

			var old corev1.Pod
			err = backing.Get(t.Context(), client.ObjectKey{Namespace: deploy.Namespace, Name: "old-pod"}, &old)
			if tc.wantDelete {
				require.Equal(t, 1, deletes)
				require.True(t, apierrors.IsNotFound(err), "old Pod should be deleted: %v", err)
			} else {
				require.Zero(t, deletes)
				require.NoError(t, err)
			}
		})
	}
}

func TestOverlayRecoveryRechecksEnablementBeforeEarlyDelete(t *testing.T) {
	now := time.Now()
	f := newOverlayRecoveryFixture(now)
	deploy, pod := f.deploy, f.pod
	backing := testutil.NewFakeClient(f.objects...)
	changed := false
	c := interceptor.NewClient(backing.(client.WithWatch), interceptor.Funcs{
		Get: func(ctx context.Context, inner client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			if _, checkingNode := obj.(*corev1.Node); checkingNode && !changed {
				var app appv1.Application
				require.NoError(t, inner.Get(ctx, client.ObjectKey{Name: deploy.Namespace + "-test-app"}, &app))
				app.Spec.Settings["enableOverlayGateway"] = "false"
				require.NoError(t, inner.Update(ctx, &app))
				changed = true
			}
			return inner.Get(ctx, key, obj, opts...)
		},
	})
	r := &TerminatedPodGCController{Client: c, APIReader: c, now: func() time.Time { return now }, minPodAge: time.Minute}
	_, err := r.Reconcile(t.Context(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(deploy)})
	require.NoError(t, err)
	require.True(t, changed)
	var old corev1.Pod
	require.NoError(t, backing.Get(t.Context(), client.ObjectKeyFromObject(pod), &old))
	require.Nil(t, old.DeletionTimestamp)
}
