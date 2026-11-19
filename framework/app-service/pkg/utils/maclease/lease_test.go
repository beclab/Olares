package maclease

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
)

const mac = "02:00:00:00:00:01"

func pod(name string) *corev1.Pod {
	return &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "apps", UID: types.UID(name)}}
}

func TestConcurrentDifferentPodsHaveOneWinner(t *testing.T) {
	kube := fake.NewSimpleClientset()
	var wg sync.WaitGroup
	var winners atomic.Int32
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if Reserve(t.Context(), kube, pod(fmt.Sprint(i)), "app", "uid", mac) == nil {
				winners.Add(1)
			}
		}(i)
	}
	wg.Wait()
	if winners.Load() != 1 {
		t.Fatalf("admitted %d different Pods", winners.Load())
	}
}
func TestRetryAndLivePodNeverExpire(t *testing.T) {
	kube := fake.NewSimpleClientset()
	a := pod("a")
	if err := Reserve(t.Context(), kube, a, "app", "uid", mac); err != nil {
		t.Fatal(err)
	}
	if err := Reserve(t.Context(), kube, a, "app", "uid", mac); err != nil {
		t.Fatal(err)
	}
	l, _ := kube.CoordinationV1().Leases(Namespace).Get(t.Context(), Name(mac), metav1.GetOptions{})
	old := metav1.NewMicroTime(time.Now().Add(-2 * PendingGrace))
	l.Spec.AcquireTime = &old
	kube.CoordinationV1().Leases(Namespace).Update(t.Context(), l, metav1.UpdateOptions{})
	now := metav1.Now()
	a.DeletionTimestamp = &now
	kube.CoreV1().Pods(a.Namespace).Create(t.Context(), a, metav1.CreateOptions{})
	if err := Reserve(t.Context(), kube, pod("b"), "app", "uid", mac); err == nil {
		t.Fatal("terminating Pod lost its reservation")
	}
	if err := Reconcile(t.Context(), kube, "uid"); err != nil {
		t.Fatal(err)
	}
	if err := Validate(t.Context(), kube, a, "uid", mac); err != nil {
		t.Fatal(err)
	}
	kube.CoreV1().Pods(a.Namespace).Delete(t.Context(), a.Name, metav1.DeleteOptions{})
	if err := Reserve(t.Context(), kube, pod("b"), "app", "uid", mac); err != nil {
		t.Fatal(err)
	}
	if err := Validate(t.Context(), kube, a, "uid", mac); err == nil {
		t.Fatal("old holder still accepted")
	}
}
func TestRejectedAdmissionExpiresButApplicationCannotSteal(t *testing.T) {
	kube := fake.NewSimpleClientset()
	if err := Reserve(t.Context(), kube, pod("a"), "app", "uid", mac); err != nil {
		t.Fatal(err)
	}
	if err := Reserve(t.Context(), kube, pod("b"), "app", "uid", mac); err == nil {
		t.Fatal("pending request lost reservation")
	}
	l, _ := kube.CoordinationV1().Leases(Namespace).Get(t.Context(), Name(mac), metav1.GetOptions{})
	old := metav1.NewMicroTime(time.Now().Add(-2 * PendingGrace))
	l.Spec.AcquireTime = &old
	kube.CoordinationV1().Leases(Namespace).Update(t.Context(), l, metav1.UpdateOptions{})
	if err := Reserve(t.Context(), kube, pod("b"), "other-app", "other-uid", mac); err == nil {
		t.Fatal("cross-application takeover")
	}
	if err := Reconcile(t.Context(), kube, "uid"); err != nil {
		t.Fatal(err)
	}
	if err := Reserve(t.Context(), kube, pod("b"), "app", "uid", mac); err != nil {
		t.Fatal(err)
	}
}

func TestSamePodRetryRefreshesExpiredPendingReservation(t *testing.T) {
	kube := fake.NewSimpleClientset()
	if err := Reserve(t.Context(), kube, pod("a"), "app", "uid", mac); err != nil {
		t.Fatal(err)
	}
	l, _ := kube.CoordinationV1().Leases(Namespace).Get(t.Context(), Name(mac), metav1.GetOptions{})
	old := metav1.NewMicroTime(time.Now().Add(-2 * PendingGrace))
	l.Spec.AcquireTime = &old
	kube.CoordinationV1().Leases(Namespace).Update(t.Context(), l, metav1.UpdateOptions{})
	if err := Reserve(t.Context(), kube, pod("a"), "app", "uid", mac); err != nil {
		t.Fatal(err)
	}
	if err := Reserve(t.Context(), kube, pod("b"), "app", "uid", mac); err == nil {
		t.Fatal("renewed admission lost its reservation")
	}
}
