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
)

// ClusterScopedProbeNames returns the release names used by
// CheckClusterScopedFixedNames.
func ClusterScopedProbeNames() (a, b string) {
	return clusterScopedProbeA, clusterScopedProbeB
}

// CheckClusterScopedFixedNames compares two helm dry-run outputs rendered with
// different release names. Any cluster-scoped resource (empty namespace) whose
// kind+name is identical in both lists is treated as having a fixed name.
//
// All offenders are reported in a single multi-line error: one summary line
// carrying the remediation hint, followed by one bullet line per (kind, name)
// violation. Embedding explicit "\n" separators (instead of relying on
// errors.Join's invisible newlines) keeps the offenders visually separated
// even when a downstream logger or terminal collapses whitespace, and avoids
// repeating the long remediation hint once per violation.
func CheckClusterScopedFixedNames(listA, listB kube.ResourceList) error {
	fixed := intersectClusterScopedKeys(listA, listB)
	if len(fixed) == 0 {
		return nil
	}
	var b strings.Builder
	fmt.Fprintf(&b,
		"the following cluster-scoped resource(s) have a fixed metadata.name; use a release-unique name such as {{ .Release.Name }} when app is v1 or v3 with options.allowMultipleInstall is true\n",
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
