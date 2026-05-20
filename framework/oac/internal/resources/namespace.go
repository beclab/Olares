package resources

import (
	"errors"
	"fmt"
	"strings"

	"helm.sh/helm/v3/pkg/kube"
)

// AppNamespace is the namespace into which Olares installs every app's
// rendered chart. Charts that declare an explicit `metadata.namespace`
// other than this (or a user-system-* namespace for shared resources)
// land in the wrong place at install time, so CheckResourceNamespace
// surfaces it early.
//
// It mirrors helmrender.RenderNamespace so dry-run output (where
// .Release.Namespace == AppNamespace) is the canonical compliant case.
const AppNamespace = "app-namespace"

// userSystemNamespacePrefix is the only cross-namespace destination we
// allow for non-workload resources -- typically used by ProviderRegistry
// or similar bridge objects that need to be created in the owner's
// user-system-<owner> namespace.
const userSystemNamespacePrefix = "user-system-"

// CheckResourceNamespace enforces the install-time namespace contract on
// every helm-rendered resource:
//
//   - Workload kinds (Deployment / StatefulSet / DaemonSet) must declare
//     `metadata.namespace = AppNamespace` (i.e. "app-namespace"). Workloads
//     are the primary install-time artifacts and putting them anywhere else
//     would silently miss the app's quota / network policies.
//   - Any other namespaced resource (Service, ConfigMap, Secret, Role,
//     RoleBinding, ProviderRegistry, ...) must either land in AppNamespace
//     or in a user-system-* namespace.
//   - Cluster-scoped resources (ClusterRole, ClusterRoleBinding,
//     PersistentVolume, ...) are detected by an empty namespace and are
//     skipped: they have no notion of namespace by design.
//
// Violations are collected and returned as a single aggregated error so
// that one Lint run surfaces every offender, not just the first.
func CheckResourceNamespace(list kube.ResourceList) error {
	var errs []error
	for _, r := range list {
		ns := r.Namespace
		if ns == "" {
			continue
		}
		kind := r.Object.GetObjectKind().GroupVersionKind().Kind
		if isWorkloadKind(kind) {
			if ns != AppNamespace {
				errs = append(errs, fmt.Errorf(
					"illegal namespace: %s for %s, name %s", ns, kind, r.Name,
				))
			}
			continue
		}
		if ns != AppNamespace && !strings.HasPrefix(ns, userSystemNamespacePrefix) {
			errs = append(errs, fmt.Errorf(
				"illegal namespace: %s for %s, name %s", ns, kind, r.Name,
			))
		}
	}
	return errors.Join(errs...)
}

// isWorkloadKind reports whether a Kubernetes Kind represents a primary
// app workload that must live in AppNamespace. Currently restricted to
// the three controllers that always carry containers; pod templates
// that ship inside CRDs or Jobs are intentionally treated as "other
// namespaced resources" until we have a concrete need to lock them down.
func isWorkloadKind(kind string) bool {
	switch kind {
	case KindDeployment, KindStatefulSet, KindDaemonSet:
		return true
	default:
		return false
	}
}
