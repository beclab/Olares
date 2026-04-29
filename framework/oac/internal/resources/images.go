package resources

import (
	"sort"

	"helm.sh/helm/v3/pkg/kube"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

// ExtractWorkloadImages walks Deployment and StatefulSet resources in list
// and returns the distinct set of container images declared on the primary
// Pod template (initContainers are not included, per the "basic" decision
// in the refactor plan).
//
// DaemonSet is intentionally skipped: the resource-limits and upload-mount
// checks in walkPodContainers only inspect Deployment / StatefulSet, and the
// image listing must stay aligned with the set of workloads that actually
// go through lint — otherwise a DaemonSet could contribute images to the
// pull list while silently bypassing every other resource-level check.
//
// The returned slice is deterministic: sorted alphabetically with duplicates
// collapsed.
func ExtractWorkloadImages(list kube.ResourceList) []string {
	set := make(map[string]struct{})
	for _, r := range list {
		kind := r.Object.GetObjectKind().GroupVersionKind().Kind
		switch kind {
		case KindDeployment:
			var dep appsv1.Deployment
			if err := scheme.Scheme.Convert(r.Object, &dep, nil); err != nil {
				continue
			}
			for _, c := range dep.Spec.Template.Spec.Containers {
				if c.Image != "" {
					set[c.Image] = struct{}{}
				}
			}
		case KindStatefulSet:
			var sts appsv1.StatefulSet
			if err := scheme.Scheme.Convert(r.Object, &sts, nil); err != nil {
				continue
			}
			for _, c := range sts.Spec.Template.Spec.Containers {
				if c.Image != "" {
					set[c.Image] = struct{}{}
				}
			}
		}
	}
	out := make([]string, 0, len(set))
	for img := range set {
		out = append(out, img)
	}
	sort.Strings(out)
	return out
}

// MergeImages merges additional image entries into an existing, sorted list
// and returns the sorted, deduped result. Existing order of base is preserved
// for stability when no new entries change the set.
func MergeImages(base, extra []string) []string {
	set := make(map[string]struct{}, len(base)+len(extra))
	for _, img := range base {
		if img != "" {
			set[img] = struct{}{}
		}
	}
	for _, img := range extra {
		if img != "" {
			set[img] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for img := range set {
		out = append(out, img)
	}
	sort.Strings(out)
	return out
}
