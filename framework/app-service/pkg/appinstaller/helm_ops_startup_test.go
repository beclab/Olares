package appinstaller

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/errcode"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestWaitForStartUp_DeadlineReturnsError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	time.Sleep(2 * time.Millisecond)

	ok, err := (&HelmOps{
		ctx: ctx,
		app: &appcfg.ApplicationConfig{AppName: "gcont"},
	}).WaitForStartUp()
	if ok || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ok=%v err=%v, want context deadline", ok, err)
	}
}

func TestStartupPendingErr(t *testing.T) {
	unschedulable := []corev1.PodCondition{{
		Type:    corev1.PodScheduled,
		Status:  corev1.ConditionFalse,
		Reason:  corev1.PodReasonUnschedulable,
		Message: "0/1 nodes are available: 1 node(s) didn't match Pod's node affinity/selector",
	}}
	scheduled := []corev1.PodCondition{{
		Type:   corev1.PodScheduled,
		Status: corev1.ConditionTrue,
	}}

	cases := []struct {
		name       string
		phase      corev1.PodPhase
		conditions []corev1.PodCondition
		pendingFor time.Duration
		serverSide bool
		want       error
	}{
		{
			name:       "unschedulable pending returns immediately",
			phase:      corev1.PodPending,
			conditions: unschedulable,
			pendingFor: time.Second,
			want:       errcode.ErrPodPending,
		},
		{
			name:       "unschedulable server pod returns the server-side sentinel",
			phase:      corev1.PodPending,
			conditions: unschedulable,
			pendingFor: time.Second,
			serverSide: true,
			want:       errcode.ErrServerSidePodPending,
		},
		{
			name:       "scheduled pending keeps waiting while the image pulls",
			phase:      corev1.PodPending,
			conditions: scheduled,
			pendingFor: time.Minute,
		},
		{
			name:       "pending without a schedule decision keeps waiting",
			phase:      corev1.PodPending,
			pendingFor: time.Minute,
		},
		{
			name:       "pending for more than 1 minute still fails",
			phase:      corev1.PodPending,
			conditions: scheduled,
			pendingFor: time.Minute + time.Second,
			want:       errcode.ErrPodPending,
		},
		{
			name:       "running pod is not a pending failure",
			phase:      corev1.PodRunning,
			conditions: scheduled,
			pendingFor: time.Hour,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "gcont"},
				Status: corev1.PodStatus{
					Phase:      tc.phase,
					Conditions: tc.conditions,
				},
			}
			err := startupPendingErr(pod, tc.pendingFor, tc.serverSide)
			if tc.want == nil {
				if err != nil {
					t.Fatalf("err = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}
