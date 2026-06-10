package resources

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"helm.sh/helm/v3/pkg/kube"
)

// clusterScopedProbeA and clusterScopedProbeB are synthetic helm release names
// used to detect cluster-scoped resources whose metadata.name does not vary
// with the release (i.e. a fixed name that would collide when
// options.allowMultipleInstall allows several instances).
const (
	clusterScopedProbeA = "oac-lint-probe-a"
	clusterScopedProbeB = "oac-lint-probe-b"
	clusterScopedUserProbeA = "oac-lint-user-a"
	clusterScopedUserProbeB = "oac-lint-user-b"
	clusterScopedUserRelease = "oac-lint-user-release"
)

// ClusterScopedProbeNames returns the release names used by
// CheckClusterScopedFixedNames.
func ClusterScopedProbeNames() (a, b string) {
	return clusterScopedProbeA, clusterScopedProbeB
}

// ClusterScopedUsernameProbeNames returns the .Values.bfl.username values and
// shared release name used to detect cluster-scoped resources whose
// metadata.name is templated on {{ .Values.bfl.username }}.
func ClusterScopedUsernameProbeNames() (userA, userB, release string) {
	return clusterScopedUserProbeA, clusterScopedUserProbeB, clusterScopedUserRelease
}

// ClusterScopedFixedNameOpts configures CheckClusterScopedFixedNames.
type ClusterScopedFixedNameOpts struct {
	// AllowMultipleInstall mirrors options.allowMultipleInstall. When false,
	// cluster-scoped resources whose names vary with .Values.bfl.username
	// (detected via UsernameProbeA/B) are exempted.
	AllowMultipleInstall bool
	// UsernameProbeA and UsernameProbeB are two dry-runs of the same chart
	// with different .Values.bfl.username but an identical release name.
	// Required when AllowMultipleInstall is false.
	UsernameProbeA kube.ResourceList
	UsernameProbeB kube.ResourceList
}

// CheckClusterScopedFixedNames compares two helm dry-run outputs rendered with
// different release names. Any cluster-scoped resource (empty namespace) whose
// kind+name is identical in both lists is treated as having a fixed name.
//
// When opts.AllowMultipleInstall is false and username probe lists are
// provided, a fixed offender is exempted if the corresponding cluster-scoped
// resource (matched by stable kind+ordinal order) renders with different
// metadata.name across the two username probes — the pattern produced by
// names templated on {{ .Values.bfl.username }}.
//
// All offenders are reported in a single multi-line error: one summary line
// carrying the remediation hint, followed by one bullet line per (kind, name)
// violation. Embedding explicit "\n" separators (instead of relying on
// errors.Join's invisible newlines) keeps the offenders visually separated
// even when a downstream logger or terminal collapses whitespace, and avoids
// repeating the long remediation hint once per violation.
func CheckClusterScopedFixedNames(listA, listB kube.ResourceList, opts ClusterScopedFixedNameOpts) error {
	fixed := intersectClusterScopedKeys(listA, listB)
	if !opts.AllowMultipleInstall && len(opts.UsernameProbeA) > 0 && len(opts.UsernameProbeB) > 0 {
		fixed = filterUsernameScopedClusterKeys(fixed, listA, opts.UsernameProbeA, opts.UsernameProbeB)
	}
	if len(fixed) == 0 {
		return nil
	}
	var b strings.Builder
	fmt.Fprintf(&b,
		"the following cluster-scoped resource(s) have a fixed metadata.name; use a unique name such as {{ .Release.Namespace }} when app is v1 or v3 with options.allowMultipleInstall is true\n",
	)
	for _, key := range fixed {
		kind, name := splitClusterScopedKey(key)
		fmt.Fprintf(&b, "\n  - %s %q", kind, name)
	}
	return errors.New(b.String())
}

func clusterScopedKeys(list kube.ResourceList) map[string]struct{} {
	out := make(map[string]struct{})
	for _, r := range list {
		if r.Namespace != "" {
			continue
		}
		kind := r.Object.GetObjectKind().GroupVersionKind().Kind
		if kind == "" {
			kind = "Resource"
		}
		out[clusterScopedKey(kind, r.Name)] = struct{}{}
	}
	return out
}

func intersectClusterScopedKeys(listA, listB kube.ResourceList) []string {
	a := clusterScopedKeys(listA)
	var fixed []string
	for key := range clusterScopedKeys(listB) {
		if _, ok := a[key]; ok {
			fixed = append(fixed, key)
		}
	}
	sort.Strings(fixed)
	return fixed
}

func clusterScopedKey(kind, name string) string {
	return kind + "/" + name
}

func splitClusterScopedKey(key string) (kind, name string) {
	for i := 0; i < len(key); i++ {
		if key[i] == '/' {
			return key[:i], key[i+1:]
		}
	}
	return "Resource", key
}

type clusterScopedEntry struct {
	kind string
	name string
}

func sortedClusterScopedEntries(list kube.ResourceList) []clusterScopedEntry {
	var entries []clusterScopedEntry
	for _, r := range list {
		if r.Namespace != "" {
			continue
		}
		kind := r.Object.GetObjectKind().GroupVersionKind().Kind
		if kind == "" {
			kind = "Resource"
		}
		entries = append(entries, clusterScopedEntry{kind: kind, name: r.Name})
	}
	sort.Slice(entries, func(i, j int) bool {
		ki, kj := entries[i].kind+"/"+entries[i].name, entries[j].kind+"/"+entries[j].name
		return ki < kj
	})
	return entries
}

// filterUsernameScopedClusterKeys drops fixed offenders whose metadata.name
// changes when .Values.bfl.username changes. Resources are paired across
// probe lists by index in the kind+name-sorted cluster-scoped sequence, which
// is stable as long as the chart emits the same set of cluster-scoped kinds
// for every probe.
func filterUsernameScopedClusterKeys(fixed []string, releaseProbe, userA, userB kube.ResourceList) []string {
	releaseEntries := sortedClusterScopedEntries(releaseProbe)
	userAEntries := sortedClusterScopedEntries(userA)
	userBEntries := sortedClusterScopedEntries(userB)
	if len(userAEntries) != len(releaseEntries) || len(userBEntries) != len(releaseEntries) {
		return fixed
	}
	var remaining []string
	for _, key := range fixed {
		idx := -1
		for i, e := range releaseEntries {
			if clusterScopedKey(e.kind, e.name) == key {
				idx = i
				break
			}
		}
		if idx < 0 {
			remaining = append(remaining, key)
			continue
		}
		a, b := userAEntries[idx], userBEntries[idx]
		if a.kind != b.kind || a.kind != releaseEntries[idx].kind {
			remaining = append(remaining, key)
			continue
		}
		if a.name != b.name {
			continue
		}
		remaining = append(remaining, key)
	}
	return remaining
}
