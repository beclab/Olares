package systemcomponents

import (
	"context"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const testNode = "olares-sidecar"

// fakeSource serves a fixed cluster to the checker.
type fakeSource struct {
	statefulSets []*appsv1.StatefulSet
	pods         []*corev1.Pod
}

func (s *fakeSource) Namespaces(context.Context) ([]string, error) { return nil, nil }
func (s *fakeSource) Deployments(context.Context) ([]*appsv1.Deployment, error) {
	return nil, nil
}
func (s *fakeSource) StatefulSets(context.Context) ([]*appsv1.StatefulSet, error) {
	return s.statefulSets, nil
}
func (s *fakeSource) DaemonSets(context.Context) ([]*appsv1.DaemonSet, error) { return nil, nil }
func (s *fakeSource) Pods(context.Context) ([]*corev1.Pod, error)             { return s.pods, nil }

func natsStatefulSet(readyReplicas int32) *appsv1.StatefulSet {
	one := int32(1)
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Namespace: "os-platform", Name: "nats", Generation: 1},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &one,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app.kubernetes.io/name": "nats"},
			},
		},
		Status: appsv1.StatefulSetStatus{
			ObservedGeneration: 1,
			ReadyReplicas:      readyReplicas,
			UpdatedReplicas:    1,
			CurrentRevision:    "nats-1",
			UpdateRevision:     "nats-1",
		},
	}
}

func natsPodOnNode(node string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "os-platform",
			Name:      "nats-0",
			Labels:    map[string]string{"app.kubernetes.io/name": "nats"},
		},
		Spec: corev1.PodSpec{
			NodeName:   node,
			Containers: []corev1.Container{{Name: "nats"}},
		},
		Status: corev1.PodStatus{
			Phase:      corev1.PodRunning,
			Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}},
			ContainerStatuses: []corev1.ContainerStatus{{
				Name:  "nats",
				State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
				Ready: true,
			}},
		},
	}
}

var natsComponent = []Component{{
	Namespace: "os-platform",
	Kind:      StatefulSet,
	Name:      "nats",
	Presence:  Required,
}}

func checkNode(t *testing.T, src *fakeSource) error {
	t.Helper()
	return NewChecker(src).OnNode(testNode).Check(context.Background(), natsComponent)
}

// change-ip deletes every pod on the node and hands straight over to this check.
// While the controllers have not caught up, the StatefulSet still counts the pod
// it just lost as ready, and the node scoped pod scan sees nothing to contradict
// it. Trusting the status there reported a whole node as ready in 20ms, before
// any workload had come back.
func TestCheckOnNodeRejectsAStatusThatPredatesARestart(t *testing.T) {
	tests := []struct {
		name string
		pods []*corev1.Pod
	}{
		{
			name: "replacement not created yet",
			pods: nil,
		},
		{
			name: "replacement created but not scheduled",
			pods: []*corev1.Pod{natsPodOnNode("")},
		},
		{
			name: "replacement pending while the old pod drains elsewhere",
			pods: func() []*corev1.Pod {
				pod := natsPodOnNode("some-other-node")
				now := metav1.Now()
				pod.DeletionTimestamp = &now
				return []*corev1.Pod{pod}
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkNode(t, &fakeSource{
				statefulSets: []*appsv1.StatefulSet{natsStatefulSet(1)},
				pods:         tt.pods,
			})
			if err == nil {
				t.Fatal("node reported ready while its workload had not come back")
			}
			if want := "predates the restart"; !strings.Contains(err.Error(), want) {
				t.Errorf("got %q, want it to contain %q", err, want)
			}
		})
	}
}

// The fallback still has to do its job: a workload whose replicas legitimately
// live on another node has nothing to prove about this one.
func TestCheckOnNodeAcceptsReplicasServingElsewhere(t *testing.T) {
	if err := checkNode(t, &fakeSource{
		statefulSets: []*appsv1.StatefulSet{natsStatefulSet(1)},
		pods:         []*corev1.Pod{natsPodOnNode("some-other-node")},
	}); err != nil {
		t.Fatalf("workload serving from another node reported not ready: %v", err)
	}
}

// A pod that cannot be placed must not hold back a workload that is serving:
// only the unplaced pod is discounted, the running one still backs the claim.
func TestCheckOnNodeAcceptsAnUnschedulableSurgePod(t *testing.T) {
	surge := natsPodOnNode("")
	surge.Name = "nats-surge"

	if err := checkNode(t, &fakeSource{
		statefulSets: []*appsv1.StatefulSet{natsStatefulSet(1)},
		pods:         []*corev1.Pod{natsPodOnNode("some-other-node"), surge},
	}); err != nil {
		t.Fatalf("unschedulable surge pod blocked a serving workload: %v", err)
	}
}

// A pod that is still draining on the checked node is visible to the node scan,
// so it is rejected on its own state and the status fallback is never consulted.
func TestCheckOnNodeRejectsATerminatingPodItCanSee(t *testing.T) {
	pod := natsPodOnNode(testNode)
	now := metav1.Now()
	pod.DeletionTimestamp = &now

	err := checkNode(t, &fakeSource{
		statefulSets: []*appsv1.StatefulSet{natsStatefulSet(1)},
		pods:         []*corev1.Pod{pod},
	})
	if err == nil {
		t.Fatal("terminating pod reported ready")
	}
	if want := "is terminating"; !strings.Contains(err.Error(), want) {
		t.Errorf("got %q, want it to contain %q", err, want)
	}
}

// A live pod on the checked node is judged on its own state, and never reaches
// the status fallback at all.
func TestCheckOnNodeJudgesAPodItCanSee(t *testing.T) {
	if err := checkNode(t, &fakeSource{
		statefulSets: []*appsv1.StatefulSet{natsStatefulSet(1)},
		pods:         []*corev1.Pod{natsPodOnNode(testNode)},
	}); err != nil {
		t.Fatalf("ready pod on the node reported not ready: %v", err)
	}

	notReady := natsPodOnNode(testNode)
	notReady.Status.ContainerStatuses[0].Ready = false
	err := checkNode(t, &fakeSource{
		statefulSets: []*appsv1.StatefulSet{natsStatefulSet(1)},
		pods:         []*corev1.Pod{notReady},
	})
	if err == nil {
		t.Fatal("unready pod on the node reported ready")
	}
}
