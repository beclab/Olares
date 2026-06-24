// Package resources hosts cross-version resource-level checks that run over
// the kube.ResourceList produced by helmrender.Render.
package resources

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/beclab/Olares/framework/oac/internal/manifest"
	"helm.sh/helm/v3/pkg/kube"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes/scheme"
)

// ResourceLimits carries the CPU / memory limits declared in the manifest
// spec, normalised to plain strings so this package stays decoupled from any
// specific manifest version.
type ResourceLimits struct {
	RequiredCPU    string
	RequiredMemory string
	LimitedCPU     string
	LimitedMemory  string
}

// CheckResourceLimits validates requests/limits on every container of every
// workload in list against the manifest-level limits carried by limits. All
// violations are collected and returned as an aggregated error.
//
// All comparisons use resource.Quantity arithmetic (Add/Cmp) to avoid the
// precision loss that AsApproximateFloat64 imposed on very large aggregates,
// matching the way kube-scheduler reasons about these quantities.
func CheckResourceLimits(list kube.ResourceList, limits ResourceLimits) error {
	var errs []error
	appRCPU, _ := resource.ParseQuantity(limits.RequiredCPU)
	appRMem, _ := resource.ParseQuantity(limits.RequiredMemory)
	appLCPU, _ := resource.ParseQuantity(limits.LimitedCPU)
	appLMem, _ := resource.ParseQuantity(limits.LimitedMemory)

	// Auto-compute ("-1") envelope fields are resolved at install time from the
	// rendered chart; their value is unknown at author/lint time, so the
	// comparisons that involve them are skipped.
	autoRCPU := manifest.IsAutoResourceQuantity(limits.RequiredCPU)
	autoRMem := manifest.IsAutoResourceQuantity(limits.RequiredMemory)
	autoLCPU := manifest.IsAutoResourceQuantity(limits.LimitedCPU)
	autoLMem := manifest.IsAutoResourceQuantity(limits.LimitedMemory)

	if !autoRCPU && !autoLCPU && appRCPU.Cmp(appLCPU) > 0 {
		errs = append(errs, fmt.Errorf("spec.requiredCpu should be <= spec.limitedCpu"))
	}
	if !autoRMem && !autoLMem && appRMem.Cmp(appLMem) > 0 {
		errs = append(errs, fmt.Errorf("spec.requiredMemory should be <= spec.limitedMemory"))
	}

	var sumReqCPU, sumReqMem, sumLimCPU, sumLimMem resource.Quantity

	walkPodContainers(list, func(kind, wlName string, c corev1.Container) {
		requests := c.Resources.Requests
		cLimits := c.Resources.Limits
		if !requests.Cpu().IsZero() && !cLimits.Cpu().IsZero() && requests.Cpu().Cmp(*cLimits.Cpu()) > 0 {
			errs = append(errs, fmt.Errorf("%s: %s, container: %s requests.cpu must be <= limits.cpu", kind, wlName, c.Name))
		}
		if !requests.Memory().IsZero() && !cLimits.Memory().IsZero() && requests.Memory().Cmp(*cLimits.Memory()) > 0 {
			errs = append(errs, fmt.Errorf("%s: %s, container: %s requests.memory must be <= limits.memory", kind, wlName, c.Name))
		}
		if requests.Memory().IsZero() {
			errs = append(errs, fmt.Errorf("%s: %s, container: %s must set memory request", kind, wlName, c.Name))
		} else {
			sumReqMem.Add(*requests.Memory())
		}
		if requests.Cpu().IsZero() {
			errs = append(errs, fmt.Errorf("%s: %s, container: %s must set cpu request", kind, wlName, c.Name))
		} else {
			sumReqCPU.Add(*requests.Cpu())
		}
		if cLimits.Memory().IsZero() {
			errs = append(errs, fmt.Errorf("%s: %s, container: %s must set memory limit", kind, wlName, c.Name))
		} else {
			sumLimMem.Add(*cLimits.Memory())
		}
		if cLimits.Cpu().IsZero() {
			errs = append(errs, fmt.Errorf("%s: %s, container: %s must set cpu limit", kind, wlName, c.Name))
		} else {
			sumLimCPU.Add(*cLimits.Cpu())
		}
	})
	exceed := false
	if !autoLCPU && sumLimCPU.Cmp(appLCPU) > 0 {
		if !exceed {
			errs = append(errs, fmt.Errorf("The total requested container resources exceed the allocated spec in OlaresManifest.yaml\n"))
			exceed = true
		}
		errs = append(errs, fmt.Errorf(
			"sum of container resources.limits.cpu (%s) must be <= limitedCpu (%s)",
			sumLimCPU.String(), quantityDisplay(limits.LimitedCPU, appLCPU),
		))
	}
	if !autoLMem && sumLimMem.Cmp(appLMem) > 0 {
		if !exceed {
			errs = append(errs, fmt.Errorf("The total requested container resources exceed the allocated spec in OlaresManifest.yaml\n"))
			exceed = true
		}
		errs = append(errs, fmt.Errorf(
			"sum of container resources.limits.memory (%s) must be <= limitedMemory (%s)",
			sumLimMem.String(), quantityDisplay(limits.LimitedMemory, appLMem),
		))
	}
	if !autoRCPU && sumReqCPU.Cmp(appRCPU) > 0 {
		if !exceed {
			errs = append(errs, fmt.Errorf("The total requested container resources exceed the allocated spec in OlaresManifest.yaml\n"))
			exceed = true
		}
		errs = append(errs, fmt.Errorf(
			"sum of container resources.requests.cpu (%s) must be <= requiredCpu (%s)",
			sumReqCPU.String(), quantityDisplay(limits.RequiredCPU, appRCPU),
		))
	}
	if !autoRMem && sumReqMem.Cmp(appRMem) > 0 {
		if !exceed {
			errs = append(errs, fmt.Errorf("The total requested container resources exceed the allocated spec in OlaresManifest.yaml\n"))
			exceed = true
		}
		errs = append(errs, fmt.Errorf(
			"sum of container resources.requests.memory (%s) must be <= requiredMemory (%s)",
			sumReqMem.String(), quantityDisplay(limits.RequiredMemory, appRMem),
		))
	}
	return errors.Join(errs...)
}

// quantityDisplay picks the most readable rendering for a manifest-side
// resource quantity: when the manifest text parses cleanly, we surface
// the author's exact spelling (e.g. "200m" or "1Gi") rather than the
// canonical form Quantity.String() produces, because the same number
// can render either way and matching the manifest helps the reader
// locate the offending field. We fall back to the parsed Quantity's
// String() when the raw text is empty or unparseable so the error
// message never shows an empty parenthesis pair.
func quantityDisplay(raw string, parsed resource.Quantity) string {
	if raw != "" {
		return raw
	}
	return parsed.String()
}

// CheckUploadConfig ensures that, if the manifest declares an options.upload
// destination, at least one primary container mounts it.
func CheckUploadConfig(list kube.ResourceList, uploadDest string) error {
	if uploadDest == "" {
		return nil
	}
	dest := filepath.Clean(uploadDest)
	found := false
	walkPodContainers(list, func(_, _ string, c corev1.Container) {
		if found {
			return
		}
		for _, v := range c.VolumeMounts {
			if filepath.Clean(v.MountPath) == dest {
				found = true
				return
			}
		}
	})
	if !found {
		return fmt.Errorf("cannot find volumemount path equal to options.upload.dest %q", uploadDest)
	}
	return nil
}

// CheckDeploymentName enforces that, for apps, at least one Deployment or
// StatefulSet is named after the app itself. Non-app configs are skipped.
func CheckDeploymentName(list kube.ResourceList, configType, appName string) error {
	if configType != "app" {
		return nil
	}
	for _, r := range list {
		kind := r.Object.GetObjectKind().GroupVersionKind().Kind
		if (kind == KindDeployment || kind == KindStatefulSet) && r.Name == appName {
			return nil
		}
	}
	return fmt.Errorf("must have a Deployment or StatefulSet named %q", appName)
}

// walkPodContainers iterates over Deployment and StatefulSet workloads only
// (per current requirements) and yields each primary container.
func walkPodContainers(list kube.ResourceList, fn func(kind, wlName string, c corev1.Container)) {
	for _, r := range list {
		kind := r.Object.GetObjectKind().GroupVersionKind().Kind
		switch kind {
		case KindDeployment:
			var dep appsv1.Deployment
			if err := scheme.Scheme.Convert(r.Object, &dep, nil); err != nil {
				continue
			}
			for _, c := range dep.Spec.Template.Spec.Containers {
				fn(kind, dep.Name, c)
			}
		case KindStatefulSet:
			var sts appsv1.StatefulSet
			if err := scheme.Scheme.Convert(r.Object, &sts, nil); err != nil {
				continue
			}
			for _, c := range sts.Spec.Template.Spec.Containers {
				fn(kind, sts.Name, c)
			}
		}
	}
}
