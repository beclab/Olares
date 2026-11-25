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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestTerminatedOverlayPodRecovery(t *testing.T) {
	now := time.Date(2026, time.November, 25, 9, 10, 0, 0, time.UTC)
	tests := []struct {
		name        string
		change      func(*overlayRecoveryFixture)
		wantDelete  bool
		wantRequeue bool
	}{
		{name: "reclaims GPU admission failure without a replacement", wantDelete: true, wantRequeue: true},
		{name: "keeps succeeded singleton until replacement is ready", change: func(f *overlayRecoveryFixture) { f.pod.Status.Phase = corev1.PodSucceeded }, wantRequeue: true},
		{name: "keeps unrelated failure until replacement is ready", change: func(f *overlayRecoveryFixture) { f.pod.Status.Reason = "OtherFailure" }, wantRequeue: true},
		{name: "keeps first admission rejection without runtime observations", change: func(f *overlayRecoveryFixture) {
			f.pod.Status = corev1.PodStatus{Phase: corev1.PodFailed, Reason: "UnexpectedAdmissionError"}
		}, wantRequeue: true},
		{name: "waits for minimum terminal age", change: func(f *overlayRecoveryFixture) {
			f.pod.Status.ContainerStatuses[0].State.Terminated.FinishedAt = metav1.NewTime(now.Add(-30 * time.Second))
		}, wantRequeue: true},
		{name: "waits for sandbox teardown", change: func(f *overlayRecoveryFixture) { f.pod.Status.Conditions[1].Status = corev1.ConditionTrue }, wantRequeue: true},
		{name: "waits for missing sandbox evidence", change: func(f *overlayRecoveryFixture) { f.pod.Status.Conditions = nil }, wantRequeue: true},
		{name: "waits for unknown sandbox", change: func(f *overlayRecoveryFixture) { f.pod.Status.Conditions[1].Status = corev1.ConditionUnknown }, wantRequeue: true},
		{name: "waits for sandbox timestamp", change: func(f *overlayRecoveryFixture) { f.pod.Status.Conditions[1].LastTransitionTime = metav1.Time{} }, wantRequeue: true},
		{name: "waits after recent sandbox teardown", change: func(f *overlayRecoveryFixture) {
			f.pod.Status.Conditions[1].LastTransitionTime = metav1.NewTime(now.Add(-30 * time.Second))
		}, wantRequeue: true},
		{name: "keeps Unknown pod", change: func(f *overlayRecoveryFixture) { f.pod.Status.Phase = corev1.PodUnknown }},
		{name: "keeps Running pod", change: func(f *overlayRecoveryFixture) { f.pod.Status.Phase = corev1.PodRunning }},
		{name: "keeps Pending pod", change: func(f *overlayRecoveryFixture) { f.pod.Status.Phase = corev1.PodPending }},
		{name: "waits for container termination", change: func(f *overlayRecoveryFixture) {
			f.pod.Status.ContainerStatuses[0].State = corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}
		}, wantRequeue: true},
		{name: "keeps missing runtime status", change: func(f *overlayRecoveryFixture) { f.pod.Status.ContainerStatuses = nil }, wantRequeue: true},
		{name: "waits for init termination", change: func(f *overlayRecoveryFixture) {
			f.pod.Spec.InitContainers = []corev1.Container{{Name: "init"}}
			f.pod.Status.InitContainerStatuses = []corev1.ContainerStatus{{Name: "init", State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{}}}}
		}, wantRequeue: true},
		{name: "keeps pod with missing init status", change: func(f *overlayRecoveryFixture) {
			f.pod.Spec.InitContainers = []corev1.Container{{Name: "init"}}
		}, wantRequeue: true},
		{name: "waits for native sidecar termination", change: func(f *overlayRecoveryFixture) {
			f.pod.Spec.InitContainers = []corev1.Container{{Name: "sidecar", RestartPolicy: ptr.To(corev1.ContainerRestartPolicyAlways)}}
			f.pod.Status.InitContainerStatuses = []corev1.ContainerStatus{{Name: "sidecar", State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}}}
		}, wantRequeue: true},
		{name: "waits for ephemeral container termination", change: func(f *overlayRecoveryFixture) {
			f.pod.Spec.EphemeralContainers = []corev1.EphemeralContainer{{EphemeralContainerCommon: corev1.EphemeralContainerCommon{Name: "debug"}}}
			f.pod.Status.EphemeralContainerStatuses = []corev1.ContainerStatus{{Name: "debug", State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}}}
		}, wantRequeue: true},
		{name: "reclaims after init sidecar and ephemeral containers terminate", change: func(f *overlayRecoveryFixture) {
			f.pod.Spec.InitContainers = []corev1.Container{{Name: "sidecar", RestartPolicy: ptr.To(corev1.ContainerRestartPolicyAlways)}}
			f.pod.Status.InitContainerStatuses = []corev1.ContainerStatus{{Name: "sidecar", State: f.pod.Status.ContainerStatuses[0].State}}
			f.pod.Spec.EphemeralContainers = []corev1.EphemeralContainer{{EphemeralContainerCommon: corev1.EphemeralContainerCommon{Name: "debug"}}}
			f.pod.Status.EphemeralContainerStatuses = []corev1.ContainerStatus{{Name: "debug", State: f.pod.Status.ContainerStatuses[0].State}}
		}, wantDelete: true, wantRequeue: true},
		{name: "waits for minimum age after ephemeral termination", change: func(f *overlayRecoveryFixture) {
			f.pod.Spec.EphemeralContainers = []corev1.EphemeralContainer{{EphemeralContainerCommon: corev1.EphemeralContainerCommon{Name: "debug"}}}
			f.pod.Status.EphemeralContainerStatuses = []corev1.ContainerStatus{{Name: "debug", State: corev1.ContainerState{
				Terminated: &corev1.ContainerStateTerminated{FinishedAt: metav1.NewTime(now.Add(-30 * time.Second))},
			}}}
		}, wantRequeue: true},
		{name: "keeps terminating pod", change: func(f *overlayRecoveryFixture) {
			f.pod.DeletionTimestamp = ptr.To(metav1.NewTime(now))
			f.pod.Finalizers = []string{"test.example/teardown"}
		}},
		{name: "waits for NotReady node", change: func(f *overlayRecoveryFixture) { f.node.Status.Conditions[0].Status = corev1.ConditionFalse }, wantRequeue: true},
		{name: "waits for Unknown node", change: func(f *overlayRecoveryFixture) { f.node.Status.Conditions[0].Status = corev1.ConditionUnknown }, wantRequeue: true},
		{name: "waits for node conditions", change: func(f *overlayRecoveryFixture) { f.node.Status.Conditions = nil }, wantRequeue: true},
		{name: "waits for deleting node", change: func(f *overlayRecoveryFixture) {
			f.node.DeletionTimestamp = ptr.To(metav1.NewTime(now))
			f.node.Finalizers = []string{"test.example/teardown"}
		}, wantRequeue: true},
		{name: "waits for missing node", change: func(f *overlayRecoveryFixture) { f.pod.Spec.NodeName = "missing" }, wantRequeue: true},
		{name: "keeps pod without assigned node", change: func(f *overlayRecoveryFixture) { f.pod.Spec.NodeName = "" }, wantRequeue: true},
		{name: "keeps non-overlay pod", change: func(f *overlayRecoveryFixture) { delete(f.pod.Labels, constants.ApplicationMacvlanInitLabel) }},
		{name: "keeps multi replica workload", change: func(f *overlayRecoveryFixture) { f.deploy.Spec.Replicas = ptr.To(int32(2)) }, wantRequeue: true},
		{name: "keeps rolling workload", change: func(f *overlayRecoveryFixture) {
			f.deploy.Spec.Strategy.Type = appsv1.RollingUpdateDeploymentStrategyType
		}, wantRequeue: true},
		{name: "keeps mismatched ReplicaSet UID", change: func(f *overlayRecoveryFixture) { f.pod.OwnerReferences[0].UID = "stale" }},
		{name: "keeps mismatched Deployment UID", change: func(f *overlayRecoveryFixture) { f.rs.OwnerReferences[0].UID = "stale" }},
		{name: "keeps unowned pod", change: func(f *overlayRecoveryFixture) { f.pod.OwnerReferences = nil }},
		{name: "keeps missing Application", change: func(f *overlayRecoveryFixture) { f.app.Name = "other-app" }},
		{name: "keeps disabled overlay", change: func(f *overlayRecoveryFixture) { f.app.Spec.Settings["enableOverlayGateway"] = "false" }},
		{name: "keeps Application in another namespace", change: func(f *overlayRecoveryFixture) { f.app.Spec.Namespace = "other-namespace" }},
		{name: "keeps Application of another owner", change: func(f *overlayRecoveryFixture) { f.app.Spec.Owner = "bob" }},
		{name: "keeps another Application name", change: func(f *overlayRecoveryFixture) { f.app.Spec.Name = "other" }},
		{name: "reclaims secondary workload of same Application", change: func(f *overlayRecoveryFixture) { f.app.Spec.DeploymentName = "primary" }, wantDelete: true, wantRequeue: true},
		{name: "keeps mismatched template owner", change: func(f *overlayRecoveryFixture) {
			f.deploy.Spec.Template.Labels[constants.ApplicationOwnerLabel] = "bob"
		}},
		{name: "keeps missing application identity", change: func(f *overlayRecoveryFixture) { delete(f.pod.Labels, constants.ApplicationOwnerLabel) }, wantRequeue: true},
		{name: "keeps deleting Application", change: func(f *overlayRecoveryFixture) {
			f.app.DeletionTimestamp = ptr.To(metav1.NewTime(now))
			f.app.Finalizers = []string{"test.example/teardown"}
		}, wantRequeue: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newOverlayRecoveryFixture(now)
			if tt.change != nil {
				tt.change(f)
			}
			c := testutil.NewFakeClient(f.objects...)
			deleteCalls := 0
			c = interceptor.NewClient(c.(client.WithWatch), interceptor.Funcs{Delete: func(ctx context.Context, inner client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
				deleteCalls++
				options := &client.DeleteOptions{}
				options.ApplyOptions(opts)
				require.NotNil(t, options.Preconditions)
				require.Equal(t, f.pod.UID, *options.Preconditions.UID)
				require.NotNil(t, options.Preconditions.ResourceVersion)
				require.Equal(t, obj.GetResourceVersion(), *options.Preconditions.ResourceVersion)
				require.Nil(t, options.GracePeriodSeconds, "must use normal deletion, not force deletion")
				return inner.Delete(ctx, obj, opts...)
			}})
			r := &TerminatedPodGCController{Client: c, minPodAge: time.Minute, readyDelay: time.Minute, now: func() time.Time { return now }}
			req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(f.deploy)}
			result, err := r.Reconcile(t.Context(), req)
			require.NoError(t, err)
			if tt.wantDelete {
				require.Equal(t, 1, deleteCalls)
			} else {
				require.Zero(t, deleteCalls)
			}
			require.Equal(t, tt.wantRequeue, result.RequeueAfter > 0)
			var got corev1.Pod
			err = c.Get(t.Context(), client.ObjectKeyFromObject(f.pod), &got)
			if tt.wantDelete {
				require.True(t, apierrors.IsNotFound(err))
			} else {
				require.NoError(t, err)
			}
			// A fresh reconciler must also discover the terminal Pod after a
			// node recovers, without relying on an in-memory timer or Pod update.
			if tt.name == "waits for NotReady node" {
				var node corev1.Node
				require.NoError(t, c.Get(t.Context(), types.NamespacedName{Name: "node"}, &node))
				node.Status.Conditions[0].Status = corev1.ConditionTrue
				require.NoError(t, c.Status().Update(t.Context(), &node))
				restarted := &TerminatedPodGCController{Client: c, minPodAge: time.Minute, now: func() time.Time { return now }}
				_, err = restarted.Reconcile(t.Context(), req)
				require.NoError(t, err)
				require.True(t, apierrors.IsNotFound(c.Get(t.Context(), client.ObjectKeyFromObject(f.pod), &got)))
			}
		})
	}
}

func TestTerminatedOverlayPodGCPropagatesReadAndDeleteFailures(t *testing.T) {
	// Fail closed on API errors; a failed delete must remain retryable.
	for _, kind := range []string{"application", "node", "pod", "deployment", "replicaset", "delete"} {
		t.Run(kind, func(t *testing.T) {
			f := newOverlayRecoveryFixture(time.Now())
			base := testutil.NewFakeClient(f.objects...)
			failure := errors.New("API temporarily unavailable")
			injected := false
			live := interceptor.NewClient(base.(client.WithWatch), interceptor.Funcs{
				Get: func(ctx context.Context, inner client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					var resource string
					switch obj.(type) {
					case *appv1.Application:
						resource = "application"
					case *corev1.Node:
						resource = "node"
					case *corev1.Pod:
						resource = "pod"
					case *appsv1.Deployment:
						resource = "deployment"
					case *appsv1.ReplicaSet:
						resource = "replicaset"
					}
					if kind == resource {
						injected = true
						return failure
					}
					return inner.Get(ctx, key, obj, opts...)
				},
			})
			c := interceptor.NewClient(base.(client.WithWatch), interceptor.Funcs{
				Delete: func(context.Context, client.WithWatch, client.Object, ...client.DeleteOption) error {
					require.Equal(t, "delete", kind, "a failed read must prevent deletion")
					injected = true
					return failure
				},
			})
			r := &TerminatedPodGCController{Client: c, APIReader: live, minPodAge: time.Minute}
			_, err := r.Reconcile(t.Context(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(f.deploy)})
			require.ErrorIs(t, err, failure)
			require.True(t, injected)
			var pod corev1.Pod
			require.NoError(t, base.Get(t.Context(), client.ObjectKeyFromObject(f.pod), &pod))
			require.Nil(t, pod.DeletionTimestamp)
		})
	}
}
