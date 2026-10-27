package meshinagent

import (
	"context"
	"errors"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ErrWorkloadGone marks a rollout target that disappeared; callers retry instead
// of recording success, so a workload recreated moments later is still rolled.
var ErrWorkloadGone = errors.New("rollout target not found")

// WorkloadRef identifies a Deployment or StatefulSet to roll.
type WorkloadRef struct {
	Kind      string // "Deployment" or "StatefulSet"
	Namespace string
	Name      string
}

// Key returns namespace/kind/name for queue dedupe.
func (w WorkloadRef) Key() string {
	return w.Namespace + "/" + w.Kind + "/" + w.Name
}

// BumpWorkload sets restartedAt on the pod template and, for Deployments,
// ensures RollingUpdate maxUnavailable=0 (surge-first / zero downtime).
func BumpWorkload(ctx context.Context, c client.Client, ref WorkloadRef) error {
	if c == nil || ref.Namespace == "" || ref.Name == "" {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339)
	switch ref.Kind {
	case "Deployment", "":
		return bumpDeployment(ctx, c, ref.Namespace, ref.Name, now)
	case "StatefulSet":
		return bumpStatefulSet(ctx, c, ref.Namespace, ref.Name, now)
	default:
		err := fmt.Errorf("mesh-in-rollout: unsupported kind %q for %s/%s", ref.Kind, ref.Namespace, ref.Name)
		klog.Errorf("%v", err)
		return err
	}
}

func bumpDeployment(ctx context.Context, c client.Client, ns, name, restartedAt string) error {
	var dep appsv1.Deployment
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, &dep); err != nil {
		if apierrors.IsNotFound(err) {
			klog.Warningf("mesh-in-rollout: deploy %s/%s not found, will retry", ns, name)
			return fmt.Errorf("%w: deployment %s/%s", ErrWorkloadGone, ns, name)
		}
		klog.Errorf("mesh-in-rollout: get deploy %s/%s failed: %v", ns, name, err)
		return err
	}
	base := dep.DeepCopy()
	if dep.Spec.Template.Annotations == nil {
		dep.Spec.Template.Annotations = map[string]string{}
	}
	dep.Spec.Template.Annotations[RestartedAtAnnotation] = restartedAt
	ensureSurgeFirstStrategy(&dep)
	if err := c.Patch(ctx, &dep, client.MergeFrom(base)); err != nil {
		klog.Errorf("mesh-in-rollout: patch deploy %s/%s failed: %v", ns, name, err)
		return err
	}
	klog.Infof("mesh-in-rollout: bumped Deployment %s/%s restartedAt=%s", ns, name, restartedAt)
	return nil
}

func bumpStatefulSet(ctx context.Context, c client.Client, ns, name, restartedAt string) error {
	var sts appsv1.StatefulSet
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, &sts); err != nil {
		if apierrors.IsNotFound(err) {
			klog.Warningf("mesh-in-rollout: sts %s/%s not found, will retry", ns, name)
			return fmt.Errorf("%w: statefulset %s/%s", ErrWorkloadGone, ns, name)
		}
		klog.Errorf("mesh-in-rollout: get sts %s/%s failed: %v", ns, name, err)
		return err
	}
	base := sts.DeepCopy()
	if sts.Spec.Template.Annotations == nil {
		sts.Spec.Template.Annotations = map[string]string{}
	}
	sts.Spec.Template.Annotations[RestartedAtAnnotation] = restartedAt
	if err := c.Patch(ctx, &sts, client.MergeFrom(base)); err != nil {
		klog.Errorf("mesh-in-rollout: patch sts %s/%s failed: %v", ns, name, err)
		return err
	}
	klog.Infof("mesh-in-rollout: bumped StatefulSet %s/%s restartedAt=%s", ns, name, restartedAt)
	return nil
}

func ensureSurgeFirstStrategy(dep *appsv1.Deployment) {
	if dep == nil {
		return
	}
	zero := intstr.FromInt(0)
	if dep.Spec.Strategy.Type != "" && dep.Spec.Strategy.Type != appsv1.RollingUpdateDeploymentStrategyType {
		return
	}
	dep.Spec.Strategy.Type = appsv1.RollingUpdateDeploymentStrategyType
	if dep.Spec.Strategy.RollingUpdate == nil {
		dep.Spec.Strategy.RollingUpdate = &appsv1.RollingUpdateDeployment{}
	}
	dep.Spec.Strategy.RollingUpdate.MaxUnavailable = &zero
}

// WorkloadRolloutComplete reports whether the workload finished replacing pods.
// The rollout is only recorded as done once the new generation is fully ready, so
// a stalled or crash-looping replacement is retried instead of marked successful.
func WorkloadRolloutComplete(ctx context.Context, c client.Client, ref WorkloadRef) (bool, error) {
	if c == nil {
		return false, nil
	}
	switch ref.Kind {
	case "Deployment", "":
		var dep appsv1.Deployment
		if err := c.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, &dep); err != nil {
			if apierrors.IsNotFound(err) {
				return false, fmt.Errorf("%w: deployment %s/%s", ErrWorkloadGone, ref.Namespace, ref.Name)
			}
			klog.Errorf("mesh-in-rollout: get deploy %s/%s status failed: %v", ref.Namespace, ref.Name, err)
			return false, err
		}
		return deploymentRolloutComplete(&dep), nil
	case "StatefulSet":
		var sts appsv1.StatefulSet
		if err := c.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, &sts); err != nil {
			if apierrors.IsNotFound(err) {
				return false, fmt.Errorf("%w: statefulset %s/%s", ErrWorkloadGone, ref.Namespace, ref.Name)
			}
			klog.Errorf("mesh-in-rollout: get sts %s/%s status failed: %v", ref.Namespace, ref.Name, err)
			return false, err
		}
		return statefulSetRolloutComplete(&sts), nil
	default:
		return false, fmt.Errorf("mesh-in-rollout: unsupported kind %q", ref.Kind)
	}
}

func deploymentRolloutComplete(dep *appsv1.Deployment) bool {
	desired := int32(1)
	if dep.Spec.Replicas != nil {
		desired = *dep.Spec.Replicas
	}
	if desired == 0 {
		return dep.Status.Replicas == 0
	}
	if dep.Status.ObservedGeneration < dep.Generation {
		return false
	}
	return dep.Status.UpdatedReplicas >= desired &&
		dep.Status.ReadyReplicas >= desired &&
		dep.Status.Replicas <= desired
}

func statefulSetRolloutComplete(sts *appsv1.StatefulSet) bool {
	desired := int32(1)
	if sts.Spec.Replicas != nil {
		desired = *sts.Spec.Replicas
	}
	if desired == 0 {
		return sts.Status.Replicas == 0
	}
	if sts.Status.ObservedGeneration < sts.Generation {
		return false
	}
	if sts.Status.UpdateRevision != "" && sts.Status.CurrentRevision != sts.Status.UpdateRevision {
		return false
	}
	return sts.Status.UpdatedReplicas >= desired && sts.Status.ReadyReplicas >= desired
}
