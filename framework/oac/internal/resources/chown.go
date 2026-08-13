package resources

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"helm.sh/helm/v3/pkg/kube"
	corev1 "k8s.io/api/core/v1"
)

// recursiveChownRe matches a chown invocation carrying a recursive flag,
// in any of the shapes charts actually write it: at the start of the
// script, after a separator (;, &&, ||, newline) or after a pipe. The
// flag group tolerates bundled short options (-Rf, -hR) and repeated
// flags (chown -R -h).
var recursiveChownRe = regexp.MustCompile(`(?:^|[;&|\n])\s*chown\s+(?:-[a-zA-Z]*\s+)*-[a-zA-Z]*R[a-zA-Z]*(?:\s+-[a-zA-Z]+)*\s+([^\n;&|]*)`)

// CheckBlindRecursiveChown flags containers that run a recursive chown
// over a hostPath mount.
//
// Olares brings userspace directories to 1000:1000 from a platform init
// container, non-recursively and only for the few paths that were not
// created by the application process itself. A chart-side `chown -R` on
// top of that is redundant, and it is the single biggest source of
// upgrade-time EPERM: the cost scales with the data tree, and one
// un-chownable inode deep inside a FUSE / JuiceFS mount fails the init
// container and leaves the pod stuck in Initializing.
//
// Only recursive invocations aimed at a hostPath mount are reported. A
// non-recursive `chown` is fine (that is what the platform itself does),
// and a recursive chown over an emptyDir or an image path is out of
// scope for this rule.
//
// The check is OFF by default so it never blocks the install/upgrade
// admission gate for existing apps; enable it with WithBlindChownCheck()
// on the PR gate, where new charts must be clean.
func CheckBlindRecursiveChown(list kube.ResourceList) error {
	var errs []error
	walkPodSpecs(list, func(kind, name string, spec corev1.PodSpec) {
		hostPathVolumes := hostPathVolumeNames(spec)
		if len(hostPathVolumes) == 0 {
			return
		}
		for _, c := range allContainers(spec) {
			mounts := hostPathMountPaths(c, hostPathVolumes)
			if len(mounts) == 0 {
				continue
			}
			for _, target := range recursiveChownTargets(c) {
				hit, ok := matchMountPath(target, mounts)
				if !ok {
					continue
				}
				errs = append(errs, fmt.Errorf(
					"blind recursive chown on hostPath mount %s: %s %s, container %s (target %q); "+
						"drop the `chown -R` -- the platform's olares-prepare-userspace init container "+
						"already owns these directories non-recursively",
					hit, kind, name, c.Name, target,
				))
			}
		}
	})
	return errors.Join(errs...)
}

// hostPathVolumeNames returns the names of the pod's hostPath volumes.
func hostPathVolumeNames(spec corev1.PodSpec) map[string]struct{} {
	out := map[string]struct{}{}
	for _, v := range spec.Volumes {
		if v.HostPath != nil {
			out[v.Name] = struct{}{}
		}
	}
	return out
}

// hostPathMountPaths returns the container's mountPaths that resolve to a
// hostPath volume, longest first so matchMountPath reports the most
// specific mount when several are nested.
func hostPathMountPaths(c corev1.Container, hostPathVolumes map[string]struct{}) []string {
	var out []string
	for _, vm := range c.VolumeMounts {
		if _, ok := hostPathVolumes[vm.Name]; !ok {
			continue
		}
		if p := strings.TrimSuffix(vm.MountPath, "/"); p != "" {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return len(out[i]) > len(out[j]) })
	return out
}

// recursiveChownTargets extracts the path arguments of every recursive
// chown found in the container's command and args. The owner token
// (1000:1000, root, ...) is dropped; everything after it is a target.
func recursiveChownTargets(c corev1.Container) []string {
	script := strings.Join(append(append([]string{}, c.Command...), c.Args...), "\n")
	var out []string
	for _, m := range recursiveChownRe.FindAllStringSubmatch(script, -1) {
		fields := strings.Fields(m[1])
		if len(fields) < 2 {
			// chown -R <owner> with no path, or a malformed invocation.
			continue
		}
		for _, f := range fields[1:] {
			if strings.HasPrefix(f, "-") {
				continue
			}
			out = append(out, strings.Trim(f, `"'`))
		}
	}
	return out
}

// matchMountPath reports whether target is one of mounts or lives under
// one of them, returning the mount it matched.
func matchMountPath(target string, mounts []string) (string, bool) {
	target = strings.TrimSuffix(target, "/")
	if target == "" {
		return "", false
	}
	for _, m := range mounts {
		if target == m || strings.HasPrefix(target, m+"/") {
			return m, true
		}
	}
	return "", false
}
