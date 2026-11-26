package userspace

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestWizardReadyRequiresReadyPodInTargetNamespace(t *testing.T) {
	pod := func(name, namespace string, ready corev1.ConditionStatus) *corev1.Pod {
		return &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: map[string]string{"app": "wizard"}},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning, Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: ready}}},
		}
	}
	wizardA := pod("wizard-a", "user-space-alice", corev1.ConditionTrue)
	wizardB := pod("wizard-b", "user-space-bob", corev1.ConditionFalse)
	terminating := pod("wizard-old", wizardB.Namespace, corev1.ConditionTrue)
	now := metav1.Now()
	terminating.DeletionTimestamp = &now
	terminating.Finalizers = []string{"fixture.example/cleanup"}
	otherApp := pod("another-app", wizardB.Namespace, corev1.ConditionTrue)
	otherApp.Labels["app"] = "authelia"
	c := NewCreator(fake.NewClientBuilder().WithObjects(wizardA, wizardB, terminating, otherApp).Build(), nil, "bob")

	assertReady := func(namespace string, want bool) {
		t.Helper()
		got, err := c.wizardReady(t.Context(), namespace)
		if err != nil || got != want {
			t.Fatalf("wizard readiness in %s = %t, %v; want %t", namespace, got, err, want)
		}
	}
	assertReady(wizardA.Namespace, true)
	assertReady(wizardB.Namespace, false)
	assertReady("user-space-unenrolled", false)
	wizardB.Status.Conditions[0].Status = corev1.ConditionTrue
	if err := c.Status().Update(t.Context(), wizardB); err != nil {
		t.Fatal(err)
	}
	assertReady(wizardB.Namespace, true)
}
