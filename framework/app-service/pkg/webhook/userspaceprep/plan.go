// Package userspaceprep computes what the platform-injected
// olares-prepare-userspace init container has to do for a given pod.
//
// Olares fixes the ownership and access identity of everything under a
// user's space at uid/gid 1000. Application processes get there through
// the pod-level securityContext, but the directories themselves do not:
// market apps mount their userspace through hostPath + DirectoryOrCreate,
// kubelet creates missing paths as root:root, and hostPath is explicitly
// excluded from fsGroup ownership management. So the process lands on
// 1000 while the directory stays root, and the app cannot write.
//
// This package closes that gap without ever walking the data tree.
// Correctness does not come from scanning exhaustively; it comes from
// ownership being self-sustaining. Once the writer is uid 1000, every
// file it creates is already 1000, so each start only has to fix the few
// directories that were *not* created by the app process: the volume
// roots, the conventional roots above them, and any static subPath
// kubernetes requires to exist before the mount succeeds. Everything is
// non-recursive, which is also what keeps upgrades off the `chown -R`
// EPERM path that fails on FUSE / JuiceFS inodes.
package userspaceprep

import (
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// mountRoot is the parent of the mount points the prepare container uses
// internally. The generated script addresses directories by their path
// inside that container, never by their host path.
const mountRoot = "/olares-prepare"

// extraVolumePrefix names the hostPath volumes this package synthesizes
// for conventional roots the chart itself did not declare.
const extraVolumePrefix = "olares-prepare-root-"

// Roots holds the absolute host paths of the four userspace conventional
// roots. A root is empty when the app was not granted that permission.
type Roots struct {
	// AppData is {userspace_hostpath}/Data/{appName}.
	AppData string
	// AppCache is {appcache_hostpath}/{appName}.
	AppCache string
	// UserData is {userspace_hostpath}/Home. Unlike the others this one
	// is shared across the user's apps, so subPath creation under it
	// becomes visible in the user's own Home directory. That is a chart
	// review concern, not a reason to treat it specially here.
	UserData string
	// AppCommon is the cluster-wide {rootPath}/rootfs/Common.
	AppCommon string
}

// IsEmpty reports whether the app was granted no userspace permission at
// all, in which case there is nothing to prepare.
func (r Roots) IsEmpty() bool {
	return r.AppData == "" && r.AppCache == "" && r.UserData == "" && r.AppCommon == ""
}

// list returns the non-empty roots ordered longest first, so prefix
// matching attributes a hostPath to the most specific root when two of
// them nest (Data/<app> lives under the same tree as Home).
func (r Roots) list() []string {
	var out []string
	for _, p := range []string{r.AppData, r.AppCache, r.UserData, r.AppCommon} {
		if p == "" {
			continue
		}
		out = append(out, filepath.Clean(p))
	}
	sort.Slice(out, func(i, j int) bool {
		if len(out[i]) != len(out[j]) {
			return len(out[i]) > len(out[j])
		}
		return out[i] < out[j]
	})
	return out
}

// Target is one directory prepare must bring to 1000:1000, plus the
// subPaths it must create underneath.
type Target struct {
	// HostPath is the absolute path on the node.
	HostPath string
	// MountPath is where prepare sees HostPath inside its own container.
	MountPath string
	// VolumeName is the pod volume prepare mounts to reach HostPath.
	VolumeName string
	// SubPaths are the static volumeMounts.subPath values declared
	// against this volume, relative to HostPath, deduped and sorted.
	SubPaths []string
}

// Plan is the full set of work for one pod.
type Plan struct {
	// Targets is list A: volume roots plus the conventional roots above
	// them, each to be chowned non-recursively.
	Targets []Target
	// ExtraVolumes are hostPath volumes for conventional roots the chart
	// did not declare. They exist only so prepare can see the parent
	// directory; no other container mounts them.
	ExtraVolumes []corev1.Volume
}

// BuildPlan collects the directories prepare has to own for pod.
//
// It returns ok=false when the pod mounts nothing under a granted
// conventional root, so callers can skip injecting an init container
// that would have no work to do.
//
// Both lists are deduplicated and sorted. That is not cosmetic: the
// generated command has to be byte-identical across admissions of the
// same pod spec, or every re-admission would produce a different patch.
func BuildPlan(pod *corev1.Pod, roots Roots) (Plan, bool) {
	if pod == nil || roots.IsEmpty() {
		return Plan{}, false
	}

	rootList := roots.list()

	// hostPath of every volume that resolves under a conventional root,
	// keyed by volume name.
	volumePaths := map[string]string{}
	// The set of directories to own, mapped to the volume that reaches
	// them. An empty volume name means "no chart volume covers this, we
	// have to synthesize one".
	owners := map[string]string{}
	// subPaths accumulated per host path.
	subPaths := map[string]map[string]struct{}{}

	for _, v := range pod.Spec.Volumes {
		if v.HostPath == nil {
			continue
		}
		p := filepath.Clean(v.HostPath.Path)
		root, ok := matchRoot(p, rootList)
		if !ok {
			continue
		}
		volumePaths[v.Name] = p
		owners[p] = v.Name
		// A path assembled below the conventional root leaves the root
		// itself as whatever kubelet created it as. Own it too, so we
		// never leave a root:root directory between the userspace and
		// the app's leaf.
		if p != root {
			if _, seen := owners[root]; !seen {
				owners[root] = ""
			}
		}
	}

	if len(owners) == 0 {
		return Plan{}, false
	}

	// A conventional root can be both mounted by the chart and inferred
	// from a deeper path. The chart's own volume wins so we do not mount
	// the same host path twice.
	for name, p := range volumePaths {
		owners[p] = name
	}

	for _, c := range append(append([]corev1.Container{}, pod.Spec.InitContainers...), pod.Spec.Containers...) {
		for _, vm := range c.VolumeMounts {
			// subPathExpr is not expanded at admission time (it is still
			// "$(VAR)" here, and downward-API references have no value
			// yet), so it cannot be prepared. oac rejects it at lint.
			if vm.SubPath == "" || vm.SubPathExpr != "" {
				continue
			}
			host, ok := volumePaths[vm.Name]
			if !ok {
				continue
			}
			sp := filepath.Clean(vm.SubPath)
			if sp == "." || sp == "/" || strings.HasPrefix(sp, "..") {
				continue
			}
			sp = strings.TrimPrefix(sp, "/")
			if subPaths[host] == nil {
				subPaths[host] = map[string]struct{}{}
			}
			subPaths[host][sp] = struct{}{}
		}
	}

	hostPaths := make([]string, 0, len(owners))
	for p := range owners {
		hostPaths = append(hostPaths, p)
	}
	sort.Strings(hostPaths)

	var plan Plan
	for i, p := range hostPaths {
		mountPath := filepath.Join(mountRoot, strconv.Itoa(i))
		volumeName := owners[p]
		if volumeName == "" {
			volumeName = extraVolumePrefix + strconv.Itoa(i)
			hostPathType := corev1.HostPathDirectoryOrCreate
			plan.ExtraVolumes = append(plan.ExtraVolumes, corev1.Volume{
				Name: volumeName,
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{Path: p, Type: &hostPathType},
				},
			})
		}
		plan.Targets = append(plan.Targets, Target{
			HostPath:   p,
			MountPath:  mountPath,
			VolumeName: volumeName,
			SubPaths:   sortedKeys(subPaths[p]),
		})
	}
	return plan, true
}

// matchRoot returns the longest conventional root that p sits under, or
// equals. Matching is path-boundary aware so ".../Data/app" does not
// match a root of ".../Data/ap".
func matchRoot(p string, roots []string) (string, bool) {
	for _, root := range roots {
		if p == root || strings.HasPrefix(p, root+"/") {
			return root, true
		}
	}
	return "", false
}

func sortedKeys(set map[string]struct{}) []string {
	if len(set) == 0 {
		return nil
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
