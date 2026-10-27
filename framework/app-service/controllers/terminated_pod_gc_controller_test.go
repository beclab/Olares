package controllers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestTerminatedPodGCControllerReconcile(t *testing.T) {
	now := time.Date(2026, time.July, 24, 10, 0, 0, 0, time.UTC)

	tests := []terminatedPodGCTestCase{
		{
			name:                "deletes failed pod replaced by ready pod from another ReplicaSet",
			phase:               corev1.PodFailed,
			replicas:            1,
			targetTerminalAge:   10 * time.Minute,
			replacementReady:    true,
			replacementReadyAge: 2 * time.Minute,
			wantDeleted:         true,
		},
		{
			name:                "deletes succeeded pod",
			phase:               corev1.PodSucceeded,
			replicas:            1,
			targetTerminalAge:   10 * time.Minute,
			replacementReady:    true,
			replacementReadyAge: 2 * time.Minute,
			wantDeleted:         true,
		},
		{
			name:                   "deletes terminal pod that carries no application labels",
			phase:                  corev1.PodFailed,
			replicas:               1,
			targetTerminalAge:      10 * time.Minute,
			replacementReady:       true,
			replacementReadyAge:    2 * time.Minute,
			targetWithoutAppLabels: true,
			wantDeleted:            true,
		},
		{
			name:                "keeps unknown pod",
			phase:               corev1.PodUnknown,
			replicas:            1,
			targetTerminalAge:   10 * time.Minute,
			replacementReady:    true,
			replacementReadyAge: 2 * time.Minute,
		},
		{
			name:                "waits for replacement to become ready",
			phase:               corev1.PodFailed,
			replicas:            1,
			targetTerminalAge:   10 * time.Minute,
			replacementReady:    false,
			replacementReadyAge: 2 * time.Minute,
		},
		{
			name:                "waits for replacement readiness delay",
			phase:               corev1.PodFailed,
			replicas:            1,
			targetTerminalAge:   10 * time.Minute,
			replacementReady:    true,
			replacementReadyAge: 30 * time.Second,
			wantRequeue:         true,
		},
		{
			name:                "waits for terminal pod minimum age",
			phase:               corev1.PodFailed,
			replicas:            1,
			targetTerminalAge:   time.Minute,
			replacementReady:    true,
			replacementReadyAge: 2 * time.Minute,
			wantRequeue:         true,
		},
		{
			name:                "waits while multi replica Deployment is short of ready replicas",
			phase:               corev1.PodFailed,
			replicas:            2,
			targetTerminalAge:   10 * time.Minute,
			replacementReady:    true,
			replacementReadyAge: 2 * time.Minute,
		},
		{
			name:                "deletes failed pod once every replica of a multi replica Deployment is ready",
			phase:               corev1.PodFailed,
			replicas:            2,
			targetTerminalAge:   10 * time.Minute,
			replacementReady:    true,
			replacementReadyAge: 2 * time.Minute,
			replacementCount:    2,
			wantDeleted:         true,
		},
		{
			name:              "deletes failed pod left behind by a stopped Deployment",
			phase:             corev1.PodFailed,
			replicas:          0,
			targetTerminalAge: 10 * time.Minute,
			noReplacement:     true,
			wantDeleted:       true,
		},
		{
			name:              "waits for terminal pod minimum age on a stopped Deployment",
			phase:             corev1.PodFailed,
			replicas:          0,
			targetTerminalAge: time.Minute,
			noReplacement:     true,
			wantRequeue:       true,
		},
		{
			name:                       "does not accept replacement owned by another Deployment",
			phase:                      corev1.PodFailed,
			replicas:                   1,
			targetTerminalAge:          10 * time.Minute,
			replacementReady:           true,
			replacementReadyAge:        2 * time.Minute,
			replacementOtherDeployment: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deploy, objects := terminatedPodGCFixture(now, tt)
			fakeClient := testutil.NewFakeClient(objects...)
			reconciler := &TerminatedPodGCController{
				Client:     fakeClient,
				minPodAge:  5 * time.Minute,
				readyDelay: time.Minute,
				now:        func() time.Time { return now },
			}

			result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
				NamespacedName: client.ObjectKeyFromObject(deploy),
			})
			require.NoError(t, err)

			var target corev1.Pod
			err = fakeClient.Get(
				context.Background(),
				types.NamespacedName{Namespace: deploy.Namespace, Name: "old-pod"},
				&target,
			)
			if tt.wantDeleted {
				assert.True(t, apierrors.IsNotFound(err), "terminal pod should be deleted")
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantRequeue, result.RequeueAfter > 0)
		})
	}
}

type terminatedPodGCTestCase struct {
	name                       string
	phase                      corev1.PodPhase
	replicas                   int32
	targetTerminalAge          time.Duration
	replacementReady           bool
	replacementReadyAge        time.Duration
	replacementOtherDeployment bool
	replacementCount           int
	noReplacement              bool
	targetWithoutAppLabels     bool
	wantDeleted                bool
	wantRequeue                bool
}

func terminatedPodGCFixture(
	now time.Time,
	values terminatedPodGCTestCase,
) (*appsv1.Deployment, []client.Object) {
	appLabels := map[string]string{
		"workload":                      "test-app",
		constants.ApplicationNameLabel:  "test-app",
		constants.ApplicationOwnerLabel: "alice",
	}
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-app-alice",
			Name:      "test-app",
			UID:       types.UID("deployment-uid"),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(values.replicas),
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"workload": "test-app"}},
		},
	}
	oldRS := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       deploy.Namespace,
			Name:            "test-app-old",
			UID:             types.UID("old-rs-uid"),
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(deploy, appsv1.SchemeGroupVersion.WithKind(deployment))},
		},
	}

	replacementOwner := deploy
	objects := []client.Object{deploy, oldRS}
	if values.replacementOtherDeployment {
		replacementOwner = deploy.DeepCopy()
		replacementOwner.Name = "other-app"
		replacementOwner.UID = types.UID("other-deployment-uid")
		objects = append(objects, replacementOwner)
	}
	newRS := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       deploy.Namespace,
			Name:            "test-app-new",
			UID:             types.UID("new-rs-uid"),
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(replacementOwner, appsv1.SchemeGroupVersion.WithKind(deployment))},
		},
	}

	// Only the Deployment selector has to match; the controller is not scoped to
	// application workloads.
	targetLabels := appLabels
	if values.targetWithoutAppLabels {
		targetLabels = map[string]string{"workload": "test-app"}
	}

	targetFinishedAt := metav1.NewTime(now.Add(-values.targetTerminalAge))
	target := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         deploy.Namespace,
			Name:              "old-pod",
			UID:               types.UID("old-pod-uid"),
			Labels:            targetLabels,
			CreationTimestamp: metav1.NewTime(now.Add(-time.Hour)),
			OwnerReferences:   []metav1.OwnerReference{*metav1.NewControllerRef(oldRS, appsv1.SchemeGroupVersion.WithKind(replicaSet))},
		},
		Status: corev1.PodStatus{
			Phase: values.phase,
			ContainerStatuses: []corev1.ContainerStatus{{
				Name: "app",
				State: corev1.ContainerState{
					Terminated: &corev1.ContainerStateTerminated{FinishedAt: targetFinishedAt},
				},
			}},
			Conditions: []corev1.PodCondition{{
				Type:               corev1.PodReady,
				Status:             corev1.ConditionFalse,
				LastTransitionTime: targetFinishedAt,
			}},
		},
	}

	readyStatus := corev1.ConditionFalse
	if values.replacementReady {
		readyStatus = corev1.ConditionTrue
	}
	objects = append(objects, newRS, target)

	replacementCount := values.replacementCount
	if replacementCount == 0 {
		replacementCount = 1
	}
	if values.noReplacement {
		replacementCount = 0
	}
	for i := 0; i < replacementCount; i++ {
		name := fmt.Sprintf("new-pod-%d", i)
		objects = append(objects, &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:         deploy.Namespace,
				Name:              name,
				UID:               types.UID(name + "-uid"),
				Labels:            appLabels,
				CreationTimestamp: metav1.NewTime(now.Add(-10 * time.Minute)),
				OwnerReferences:   []metav1.OwnerReference{*metav1.NewControllerRef(newRS, appsv1.SchemeGroupVersion.WithKind(replicaSet))},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
				Conditions: []corev1.PodCondition{{
					Type:               corev1.PodReady,
					Status:             readyStatus,
					LastTransitionTime: metav1.NewTime(now.Add(-values.replacementReadyAge)),
				}},
			},
		})
	}

	return deploy, objects
}
