package resources

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"helm.sh/helm/v3/pkg/kube"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes/scheme"
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
// must name a real workload. The manifest validator already requires
// workloadReplicas on modern (olaresManifest.version >= 0.12.0) non-v2
// manifests; on legacy versions (and on v2) the field stays optional and
// callers should only invoke this when the manifest actually declares it.
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

// CheckWorkloadReplicasRendered asserts that every rendered Deployment and
// StatefulSet in list carries spec.replicas == want. list must be a helm render
// performed with `.Values.workloads.<name>.replicaCount` set to want for every
// declared workload.
//
// This is how the platform starts and stops an app: app-service does not patch
// workloads, it re-runs helm with that value overridden (HelmOps.Scale). A chart
// whose spec.replicas does not track the value therefore cannot be stopped at
// all -- the upgrade produces an identical manifest, nothing is written, the pods
// stay up, and the stop sits in Stopping until its 30-minute TTL turns into
// StopFailed. The same override drives the replicas=0 first phase of a two-phase
// install, so a chart that ignores it also runs its pods before the resource and
// GPU admission checks have passed.
//
// Static inspection of the template cannot decide this. `| default 1` neutralises
// a correct-looking reference because helm's default treats 0 as empty; a
// template that omits spec.replicas entirely takes the Kubernetes default of 1
// while naming no value at all; and a count reached through _helpers.tpl, an if
// block or a subchart is not visible on the replicas line. Comparing two real
// renders is immune to all of them.
//
// A workload with spec.replicas unset counts as 1, matching Kubernetes' own
// default rather than being skipped: "no value" is precisely the failure this
// check exists to catch.
func CheckWorkloadReplicasRendered(list kube.ResourceList, want int32) error {
	var errs []error
	for _, r := range list {
		kind := r.Object.GetObjectKind().GroupVersionKind().Kind
		var (
			name string
			got  *int32
		)
		switch kind {
		case KindDeployment:
			var dep appsv1.Deployment
			if err := scheme.Scheme.Convert(r.Object, &dep, nil); err != nil {
				return err
			}
			name, got = dep.Name, dep.Spec.Replicas
		case KindStatefulSet:
			var sts appsv1.StatefulSet
			if err := scheme.Scheme.Convert(r.Object, &sts, nil); err != nil {
				return err
			}
			name, got = sts.Name, sts.Spec.Replicas
		default:
			continue
		}
		effective := int32(1)
		if got != nil {
			effective = *got
		}
		if effective == want {
			continue
		}
		detail := fmt.Sprintf("%d", effective)
		if got == nil {
			detail = "unset (Kubernetes defaults it to 1)"
		}
		errs = append(errs, fmt.Errorf(
			"%s %q: spec.replicas rendered as %s while workloads.%s.replicaCount was %d; "+
				"the platform starts and stops an app by overriding that value, so spec.replicas must "+
				"reference it directly, without `default` or any other filter that discards a zero",
			strings.ToLower(kind), name, detail, name, want,
		))
	}
	return errors.Join(errs...)
}

// CheckReleaseNameWorkload reports whether at least one Deployment or
// StatefulSet in the rendered chart has its metadata.name templated as
// `{{ .Release.Name }}`. listA and listB must be two helm dry-runs of the
// same chart rendered with distinct release names probeA and probeB; a
// workload whose name is templated on the release name shows up as probeA
// in listA and probeB in listB simultaneously, while any workload with a
// fixed name (or with a name that mixes the release name with other tokens,
// e.g. `{{ .Release.Name }}-web`) won't satisfy both predicates at once.
//
// The check exists to back the options.allowMultipleInstall=true safety
// rule: multiple installs of the same chart need at least one
// release-scoped primary workload so that namespaced resources do not
// collide between installs. Callers should only invoke this function when
// that flag is true; the gate lives in the OAC package.
func CheckReleaseNameWorkload(listA, listB kube.ResourceList, probeA, probeB string) error {
	if hasReleaseNamedWorkload(listA, probeA) && hasReleaseNamedWorkload(listB, probeB) {
		return nil
	}
	return fmt.Errorf(
		"options.allowMultipleInstall=true requires at least one Deployment or StatefulSet whose metadata.name is set to {{ .Release.Name }}; without it multiple installs of this chart would collide on workload names",
	)
}

// hasReleaseNamedWorkload reports whether list contains a Deployment or
// StatefulSet whose metadata.name equals name. It is the per-render half of
// CheckReleaseNameWorkload's two-render comparison.
func hasReleaseNamedWorkload(list kube.ResourceList, name string) bool {
	for _, r := range list {
		kind := r.Object.GetObjectKind().GroupVersionKind().Kind
		if (kind == KindDeployment || kind == KindStatefulSet) && r.Name == name {
			return true
		}
	}
	return false
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
