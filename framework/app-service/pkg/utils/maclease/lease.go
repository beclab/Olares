// Package maclease serializes Pod admission for a fixed MAC across webhook replicas.
package maclease

import (
	"context"
	"fmt"
	"strings"
	"time"

	coordv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

const Namespace = "os-framework"
const ApplicationUIDLabel = "app.bytetrade.io/overlay-mac-application-uid"
const observedUID = "app.bytetrade.io/overlay-mac-pod-uid"

// Allow failed admissions to expire after API request/admission timeouts.
const PendingGrace = 2 * time.Minute

func Name(mac string) string        { return "overlay-mac-" + strings.ReplaceAll(mac, ":", "") }
func holder(pod *corev1.Pod) string { return pod.Namespace + "/" + pod.Name }
func expired(l *coordv1.Lease) bool {
	return l.Spec.AcquireTime != nil && time.Since(l.Spec.AcquireTime.Time) > PendingGrace
}

// Reserve uses Create/Update resourceVersion arbitration, not a process-local lock.
// The same Pod name may retry; Kubernetes itself makes that name unique.
func Reserve(ctx context.Context, kube kubernetes.Interface, pod *corev1.Pod, appName string, appUID types.UID, mac string) error {
	if pod.Name == "" || pod.Namespace == "" || appUID == "" {
		return fmt.Errorf("MAC reservation requires Pod name, namespace and application UID")
	}
	leases := kube.CoordinationV1().Leases(Namespace)
	return retry.OnError(retry.DefaultRetry, func(err error) bool { return apierrors.IsConflict(err) || apierrors.IsAlreadyExists(err) }, func() error {
		l, err := leases.Get(ctx, Name(mac), metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			h, now, duration := holder(pod), metav1.NowMicro(), int32(PendingGrace/time.Second)
			_, err = leases.Create(ctx, &coordv1.Lease{ObjectMeta: metav1.ObjectMeta{
				Name: Name(mac), Namespace: Namespace, Labels: map[string]string{ApplicationUIDLabel: string(appUID)},
				OwnerReferences: []metav1.OwnerReference{{APIVersion: "app.bytetrade.io/v1alpha1", Kind: "Application", Name: appName, UID: appUID}},
			}, Spec: coordv1.LeaseSpec{HolderIdentity: &h, AcquireTime: &now, LeaseDurationSeconds: &duration}}, metav1.CreateOptions{})
			return err
		}
		if err != nil {
			return err
		}
		if l.Labels[ApplicationUIDLabel] != string(appUID) {
			return fmt.Errorf("MAC lease belongs to another application")
		}
		if l.Spec.HolderIdentity == nil {
			return fmt.Errorf("MAC lease has no holder")
		}
		if *l.Spec.HolderIdentity == holder(pod) {
			// Refresh an admission retry before it can outlive a pending reservation.
			// Do not let a delayed retry race a takeover of an expired record.
			now := metav1.NowMicro()
			l.Spec.AcquireTime = &now
			_, getErr := kube.CoreV1().Pods(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
			if apierrors.IsNotFound(getErr) {
				delete(l.Annotations, observedUID)
			} else if getErr != nil {
				return getErr
			}
			_, err = leases.Update(ctx, l, metav1.UpdateOptions{})
			return err
		}
		ns, name, ok := strings.Cut(*l.Spec.HolderIdentity, "/")
		if !ok {
			return fmt.Errorf("invalid MAC lease holder")
		}
		_, err = kube.CoreV1().Pods(ns).Get(ctx, name, metav1.GetOptions{})
		if err == nil {
			return fmt.Errorf("MAC still occupied by Pod %s, including termination", *l.Spec.HolderIdentity)
		}
		if !apierrors.IsNotFound(err) {
			return err
		}
		if l.Annotations[observedUID] == "" && !expired(l) {
			return fmt.Errorf("MAC reserved by pending Pod %s; retry after admission completes", *l.Spec.HolderIdentity)
		}
		h, now := holder(pod), metav1.NowMicro()
		l.Spec.HolderIdentity, l.Spec.AcquireTime = &h, &now
		delete(l.Annotations, observedUID)
		_, err = leases.Update(ctx, l, metav1.UpdateOptions{})
		return err
	})
}

func Validate(ctx context.Context, kube kubernetes.Interface, pod *corev1.Pod, appUID types.UID, mac string) error {
	l, err := kube.CoordinationV1().Leases(Namespace).Get(ctx, Name(mac), metav1.GetOptions{})
	if err != nil {
		return err
	}
	if l.Labels[ApplicationUIDLabel] != string(appUID) || l.Spec.HolderIdentity == nil || *l.Spec.HolderIdentity != holder(pod) {
		return fmt.Errorf("Pod does not own the fixed MAC reservation")
	}
	return nil
}

// Reconcile observes committed Pods, permitting immediate reuse once they disappear.
// Uncommitted admissions have a bounded grace period; live/terminating Pods never expire.
func Reconcile(ctx context.Context, kube kubernetes.Interface, appUID types.UID) error {
	leases := kube.CoordinationV1().Leases(Namespace)
	list, err := leases.List(ctx, metav1.ListOptions{LabelSelector: ApplicationUIDLabel + "=" + string(appUID)})
	if err != nil {
		return err
	}
	for i := range list.Items {
		l := &list.Items[i]
		if l.Spec.HolderIdentity == nil {
			return fmt.Errorf("MAC lease %s has no holder", l.Name)
		}
		ns, name, ok := strings.Cut(*l.Spec.HolderIdentity, "/")
		if !ok {
			return fmt.Errorf("invalid MAC lease holder")
		}
		pod, err := kube.CoreV1().Pods(ns).Get(ctx, name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			if l.Annotations[observedUID] != "" || expired(l) {
				uid, rv := l.UID, l.ResourceVersion
				err = leases.Delete(ctx, l.Name, metav1.DeleteOptions{Preconditions: &metav1.Preconditions{UID: &uid, ResourceVersion: &rv}})
				if err != nil && !apierrors.IsNotFound(err) && !apierrors.IsConflict(err) {
					return err
				}
			}
			continue
		}
		if err != nil {
			return err
		}
		if l.Annotations[observedUID] != string(pod.UID) {
			if l.Annotations == nil {
				l.Annotations = map[string]string{}
			}
			l.Annotations[observedUID] = string(pod.UID)
			_, err = leases.Update(ctx, l, metav1.UpdateOptions{})
			if err != nil && !apierrors.IsConflict(err) {
				return err
			}
		}
	}
	return nil
}
