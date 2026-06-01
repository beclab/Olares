package resources

import (
	"errors"
	"fmt"
	"sort"

	"helm.sh/helm/v3/pkg/kube"
)

// CollectWorkloadNames returns the set of Deployment and StatefulSet names
// present in list. Because the chart is helm-rendered with the release name
// set to the app name, templates that name a workload `{{ .Release.Name }}`
// already appear here under the substituted app name -- callers can therefore
// compare these names directly against manifest fields that reference the app
// name.
func CollectWorkloadNames(list kube.ResourceList) map[string]struct{} {
	names := make(map[string]struct{})
	for _, r := range list {
		kind := r.Object.GetObjectKind().GroupVersionKind().Kind
		if kind == KindDeployment || kind == KindStatefulSet {
			names[r.Name] = struct{}{}
		}
	}
	return names
}

// CheckWorkloadReplicas enforces an exact correspondence between the manifest's
// workloadReplicas keys and the rendered Deployment/StatefulSet names: every
// workload must have a workloadReplicas entry, and every workloadReplicas entry
// must name a real workload. Callers should only invoke this when the manifest
// declares workloadReplicas at all (the field is optional).
func CheckWorkloadReplicas(list kube.ResourceList, replicas map[string]int32) error {
	workloads := CollectWorkloadNames(list)
	var errs []error

	missing := make([]string, 0)
	for name := range workloads {
		if _, ok := replicas[name]; !ok {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	for _, name := range missing {
		errs = append(errs, fmt.Errorf(
			"workloadReplicas is missing an entry for workload %q; every Deployment/StatefulSet must be listed",
			name,
		))
	}

	unknown := make([]string, 0)
	for name := range replicas {
		if _, ok := workloads[name]; !ok {
			unknown = append(unknown, name)
		}
	}
	sort.Strings(unknown)
	for _, name := range unknown {
		errs = append(errs, fmt.Errorf(
			"workloadReplicas entry %q does not match any rendered Deployment/StatefulSet",
			name,
		))
	}

	return errors.Join(errs...)
}

// CheckOverlayGatewayWorkloads verifies that every overlayGateway entrance
// references a workload that exists as a rendered Deployment or StatefulSet.
// Like CheckWorkloadReplicas it relies on the release-name == app-name render
// so `{{ .Release.Name }}` workloads resolve to the app name before the
// comparison. workloads carries the entrance workload references in declaration
// order so error messages can point at the offending entrance.
func CheckOverlayGatewayWorkloads(list kube.ResourceList, workloads []string) error {
	names := CollectWorkloadNames(list)
	var errs []error
	for i, w := range workloads {
		if w == "" {
			continue
		}
		if _, ok := names[w]; !ok {
			errs = append(errs, fmt.Errorf(
				"overlayGateway.entrances[%d]: workload %q does not match any rendered Deployment/StatefulSet",
				i, w,
			))
		}
	}
	return errors.Join(errs...)
}
