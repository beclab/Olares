package resources

import (
	"errors"
	"fmt"
	"strings"

	"helm.sh/helm/v3/pkg/kube"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

// CheckSecurityContextForNonBeclabImage rejects any non-beclab container
// (init or main) whose effective securityContext grants root-equivalent
// privileges. Beclab-published images are trusted to require those
// settings (they ship system-level workloads inside Olares); third-party
// images that need root are almost always a packaging mistake or a
// supply-chain risk, so they have to be reviewed manually.
//
// Forbidden settings (any of these flags the container):
//
//   - container.securityContext.privileged = true
//   - container.securityContext.runAsUser = 0
//   - container.securityContext.runAsNonRoot = false
//
// Pod-level securityContext is also examined: if the pod's
// runAsUser == 0 or runAsNonRoot == false, every non-beclab container in
// that pod inherits the unsafe default and is reported, because the
// pod-level value applies whenever the container does not override it.
//
// Images whose repository path contains "beclab/" (e.g. "beclab/foo",
// "docker.io/beclab/foo", "registry.example.com/beclab/foo") are
// exempt; everything else is in scope.
//
// The check walks the same three workload controllers other resource
// checks operate on -- Deployment, StatefulSet, DaemonSet -- and
// aggregates violations into a single returned error.
func CheckSecurityContextForNonBeclabImage(list kube.ResourceList) error {
	var errs []error
	walkPodSpecs(list, func(kind, name string, spec corev1.PodSpec) {
		appendPodSecurityContextErrors(kind, name, spec, &errs)
		for _, c := range spec.Containers {
			appendContainerSecurityContextError(kind, name, c, &errs)
		}
		for _, c := range spec.InitContainers {
			appendContainerSecurityContextError(kind, name, c, &errs)
		}
	})
	return errors.Join(errs...)
}

func appendContainerSecurityContextError(kind, name string, c corev1.Container, errs *[]error) {
	if isBeclabImage(c.Image) {
		return
	}
	sc := c.SecurityContext
	if sc == nil {
		return
	}
	privilegedBad := sc.Privileged != nil && *sc.Privileged
	runAsUserBad := sc.RunAsUser != nil && *sc.RunAsUser == 0
	runAsNonRootBad := sc.RunAsNonRoot != nil && !*sc.RunAsNonRoot
	if !privilegedBad && !runAsUserBad && !runAsNonRootBad {
		return
	}
	*errs = append(*errs, formatPrivilegedImageError(c.Image, kind, name, c.Name))
}

func appendPodSecurityContextErrors(kind, name string, spec corev1.PodSpec, errs *[]error) {
	podSC := spec.SecurityContext
	if podSC == nil {
		return
	}
	runAsUserBad := podSC.RunAsUser != nil && *podSC.RunAsUser == 0
	runAsNonRootBad := podSC.RunAsNonRoot != nil && !*podSC.RunAsNonRoot
	if !runAsUserBad && !runAsNonRootBad {
		return
	}
	report := func(c corev1.Container) {
		if isBeclabImage(c.Image) {
			return
		}
		*errs = append(*errs, formatPrivilegedImageError(c.Image, kind, name, c.Name))
	}
	for _, c := range spec.Containers {
		report(c)
	}
	for _, c := range spec.InitContainers {
		report(c)
	}
}

func formatPrivilegedImageError(image, kind, workloadName, containerName string) error {
	return fmt.Errorf(
		"non-beclab image %q runs with root-equivalent securityContext: %s %s, container %s",
		image, kind, workloadName, containerName,
	)
}

// isBeclabImage reports whether image references the beclab/ namespace,
// either directly ("beclab/foo[:tag]") or via a fully-qualified registry
// ("docker.io/beclab/foo", "registry.example.com/beclab/foo"). It does
// not require parsing the full reference grammar; matching on the path
// segment "/beclab/" (or the bare "beclab/" prefix) is enough to cover
// the registries Olares actually publishes from while keeping
// look-alike repository names (e.g. "beclab-bar/foo") out of the
// allow list.
func isBeclabImage(image string) bool {
	if image == "" {
		return false
	}
	return strings.HasPrefix(image, "beclab/") || strings.Contains(image, "/beclab/")
}

// walkPodSpecs iterates over Deployment / StatefulSet / DaemonSet entries
// in list and yields each workload's full PodSpec. Unlike
// walkPodContainers (which only surfaces containers) this helper hands
// the caller the entire PodSpec so checks that depend on pod-level
// fields (securityContext, volumes, nodeSelector, ...) can run without
// re-decoding the workload.
func walkPodSpecs(list kube.ResourceList, fn func(kind, name string, spec corev1.PodSpec)) {
	for _, r := range list {
		kind := r.Object.GetObjectKind().GroupVersionKind().Kind
		switch kind {
		case KindDeployment:
			var dep appsv1.Deployment
			if err := scheme.Scheme.Convert(r.Object, &dep, nil); err != nil {
				continue
			}
			fn(kind, dep.Name, dep.Spec.Template.Spec)
		case KindStatefulSet:
			var sts appsv1.StatefulSet
			if err := scheme.Scheme.Convert(r.Object, &sts, nil); err != nil {
				continue
			}
			fn(kind, sts.Name, sts.Spec.Template.Spec)
		case KindDaemonSet:
			var ds appsv1.DaemonSet
			if err := scheme.Scheme.Convert(r.Object, &ds, nil); err != nil {
				continue
			}
			fn(kind, ds.Name, ds.Spec.Template.Spec)
		}
	}
}
