package resources

import (
	"errors"
	"fmt"

	"helm.sh/helm/v3/pkg/kube"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

// CheckHostPath rejects workloads that combine a hostPath volume with a
// rolling-update strategy. A rolling update brings up a new pod on a
// different node, which cannot see the old node's hostPath, so the new pod
// starves on a volume that doesn't exist for it.
//
// Recreate (Deployment) and OnDelete (StatefulSet) are accepted because
// they tear the old pod down before the new one starts, eliminating the
// cross-node visibility window. The default strategy is treated as
// RollingUpdate per Kubernetes' own defaults.
//
// Workloads with no hostPath volume are unaffected; the check runs over
// the helm-rendered kube.ResourceList so it only sees the workloads the
// install would actually create.
func CheckHostPath(list kube.ResourceList) error {
	var errs []error
	for _, r := range list {
		kind := r.Object.GetObjectKind().GroupVersionKind().Kind
		switch kind {
		case KindDeployment:
			var dep appsv1.Deployment
			if err := scheme.Scheme.Convert(r.Object, &dep, nil); err != nil {
				return err
			}
			if !isRollingDeployment(dep.Spec.Strategy.Type) {
				continue
			}
			for _, v := range dep.Spec.Template.Spec.Volumes {
				if v.HostPath == nil {
					continue
				}
				errs = append(errs, fmt.Errorf(
					"deployment %s can not enable rolling update with hostpath name:%s,path:%s",
					dep.Name, v.Name, v.HostPath.Path,
				))
			}
		case KindStatefulSet:
			var sts appsv1.StatefulSet
			if err := scheme.Scheme.Convert(r.Object, &sts, nil); err != nil {
				return err
			}
			if !isRollingStatefulSet(sts.Spec.UpdateStrategy.Type) {
				continue
			}
			for _, v := range sts.Spec.Template.Spec.Volumes {
				if v.HostPath == nil {
					continue
				}
				errs = append(errs, fmt.Errorf(
					"statefulset %s can not enable rolling update with hostpath name:%s,path:%s",
					sts.Name, v.Name, v.HostPath.Path,
				))
			}
		}
	}
	return errors.Join(errs...)
}

// isRollingDeployment reports whether a Deployment update strategy lets a
// new pod come up on a different node before the old pod terminates. An
// empty strategy collapses to the Kubernetes default of RollingUpdate.
func isRollingDeployment(t appsv1.DeploymentStrategyType) bool {
	return t == "" || t == appsv1.RollingUpdateDeploymentStrategyType
}

// isRollingStatefulSet reports whether a StatefulSet update strategy lets
// a new pod come up on a different node before the old pod terminates.
// Empty collapses to RollingUpdate; OnDelete is treated as non-rolling
// because it gates pod replacement on manual deletion.
func isRollingStatefulSet(t appsv1.StatefulSetUpdateStrategyType) bool {
	return t == "" || t == appsv1.RollingUpdateStatefulSetStrategyType
}
