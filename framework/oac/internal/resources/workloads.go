package resources

import (
	"errors"
	"fmt"
	"sort"

	apimanifest "github.com/beclab/api/manifest"
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

// CheckWorkloadOptions validates the parts of workloadOptions that depend on
// the rendered workload: exact key coverage, target container names, and the
// single-replica/Recreate constraints required by instance device resources.
func CheckWorkloadOptions(list kube.ResourceList, options apimanifest.WorkloadOptions) error {
	workloads := CollectWorkloadNames(list)
	var errs []error

	for name := range workloads {
		if _, ok := options[name]; !ok {
			errs = append(errs, fmt.Errorf("workloadOptions is missing workload %q", name))
		}
	}
	for name := range options {
		if _, ok := workloads[name]; !ok {
			errs = append(errs, fmt.Errorf("workloadOptions entry %q does not match any rendered Deployment/StatefulSet", name))
		}
	}

	for _, r := range list {
		kind := r.Object.GetObjectKind().GroupVersionKind().Kind
		if kind != KindDeployment && kind != KindStatefulSet {
			continue
		}
		option, ok := options[r.Name]
		if !ok {
			continue
		}

		var containerNames map[string]struct{}
		var deploymentStrategy appsv1.DeploymentStrategyType
		switch kind {
		case KindDeployment:
			var deployment appsv1.Deployment
			if err := scheme.Scheme.Convert(r.Object, &deployment, nil); err != nil {
				return err
			}
			containerNames = make(map[string]struct{}, len(deployment.Spec.Template.Spec.Containers))
			for _, container := range deployment.Spec.Template.Spec.Containers {
				containerNames[container.Name] = struct{}{}
			}
			deploymentStrategy = deployment.Spec.Strategy.Type
		case KindStatefulSet:
			var statefulSet appsv1.StatefulSet
			if err := scheme.Scheme.Convert(r.Object, &statefulSet, nil); err != nil {
				return err
			}
			containerNames = make(map[string]struct{}, len(statefulSet.Spec.Template.Spec.Containers))
			for _, container := range statefulSet.Spec.Template.Spec.Containers {
				containerNames[container.Name] = struct{}{}
			}
		}

		hasDevice := false
		for i, capability := range option.Allow {
			for _, container := range capability.Containers {
				if _, exists := containerNames[container]; !exists {
					errs = append(errs, fmt.Errorf("workloadOptions.%s.allow[%d] targets unknown container %q", r.Name, i, container))
				}
			}
			if capability.Type == apimanifest.WorkloadAllowDeviceSerial ||
				capability.Type == apimanifest.WorkloadAllowDeviceVideo ||
				capability.Type == apimanifest.WorkloadAllowDeviceAudio ||
				capability.Type == apimanifest.WorkloadAllowDeviceHID {
				hasDevice = true
			}
		}
		if !hasDevice {
			continue
		}
		if option.Replicas == nil || *option.Replicas != 1 {
			errs = append(errs, fmt.Errorf("workloadOptions.%s.replicas must be 1 when a device capability is declared", r.Name))
		}
		if kind == KindDeployment && deploymentStrategy != appsv1.RecreateDeploymentStrategyType {
			errs = append(errs, fmt.Errorf("Deployment %q must use strategy.type=Recreate when a device capability is declared", r.Name))
		}
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
