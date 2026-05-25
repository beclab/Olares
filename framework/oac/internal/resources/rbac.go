package resources

import (
	"fmt"
	"io"
	"strings"

	"github.com/thoas/go-funk"
	"helm.sh/helm/v3/pkg/kube"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
)

// DefaultForbiddenRules describes the rules that an application service
// account is not allowed to be granted. It is embedded here so that callers
// don't have to pass it in, but the CheckServiceAccountRules function still
// accepts an arbitrary rules argument to make the policy easy to override in
// tests and alternate hosts.
const DefaultForbiddenRules = `rules:
- apiGroups:
  - '*'
  resources:
  - nodes
  - networkpolicies
  verbs:
  - create
  - update
  - patch
  - delete
  - deletecollection
`

// PolicyRules is the YAML shape of DefaultForbiddenRules.
type PolicyRules struct {
	Rules []rbacv1.PolicyRule `json:"rules" yaml:"rules"`
}

// LoadForbiddenRules parses DefaultForbiddenRules (or any equivalent string)
// into a slice of rbacv1.PolicyRule.
func LoadForbiddenRules(source string) ([]rbacv1.PolicyRule, error) {
	if source == "" {
		source = DefaultForbiddenRules
	}
	var p PolicyRules
	if err := yaml.NewYAMLOrJSONDecoder(strings.NewReader(source), 4096).Decode(&p); err != nil && err != io.EOF {
		return nil, err
	}
	return p.Rules, nil
}

// CheckServiceAccountRules inspects every Role/ClusterRole bound (via
// RoleBinding/ClusterRoleBinding) to a ServiceAccount in list and returns an
// error if any of them grants something the forbidden rules prohibit.
func CheckServiceAccountRules(list kube.ResourceList, forbidden []rbacv1.PolicyRule) error {
	if !serviceAccountRulesOK(list, forbidden) {
		return fmt.Errorf("service account role grants forbidden permissions: %+v", forbidden)
	}
	return nil
}

func serviceAccountRulesOK(list kube.ResourceList, forbidden []rbacv1.PolicyRule) bool {
	roleNames := collectBoundRoleNames(list)
	requested := collectRequestedRules(list, roleNames)
	if len(requested) == 0 {
		return true
	}
	for _, rule := range requested {
		r := rule
		if !rulesAllow(&r, forbidden...) {
			return false
		}
	}
	return true
}

func collectBoundRoleNames(list kube.ResourceList) map[string]struct{} {
	names := make(map[string]struct{})
	for _, r := range list {
		kind := r.Object.GetObjectKind().GroupVersionKind().Kind
		switch kind {
		case KindClusterRoleBinding:
			var rb rbacv1.ClusterRoleBinding
			if err := scheme.Scheme.Convert(r.Object, &rb, nil); err != nil {
				continue
			}
			for _, sub := range rb.Subjects {
				if sub.Kind == KindServiceAccount {
					names[rb.RoleRef.Name] = struct{}{}
				}
			}
		case KindRoleBinding:
			var rb rbacv1.RoleBinding
			if err := scheme.Scheme.Convert(r.Object, &rb, nil); err != nil {
				continue
			}
			for _, sub := range rb.Subjects {
				if sub.Kind == KindServiceAccount {
					names[rb.RoleRef.Name] = struct{}{}
				}
			}
		}
	}
	return names
}

func collectRequestedRules(list kube.ResourceList, roleNames map[string]struct{}) []rbacv1.PolicyRule {
	var out []rbacv1.PolicyRule
	for _, r := range list {
		kind := r.Object.GetObjectKind().GroupVersionKind().Kind
		switch kind {
		case KindRole:
			var role rbacv1.Role
			if err := scheme.Scheme.Convert(r.Object, &role, nil); err != nil {
				continue
			}
			if _, ok := roleNames[role.Name]; ok {
				out = append(out, role.Rules...)
			}
		case KindClusterRole:
			var role rbacv1.ClusterRole
			if err := scheme.Scheme.Convert(r.Object, &role, nil); err != nil {
				continue
			}
			if _, ok := roleNames[role.Name]; ok {
				out = append(out, role.Rules...)
			}
		}
	}
	return out
}

// rulesAllow reports whether request can be safely granted given the supplied
// forbidden rules. It returns false (i.e. forbidden) as soon as ANY rule in
// rules covers the request, and true only when none of them do.
//
// The previous implementation had the boolean flipped: it returned "allowed"
// as soon as any forbidden rule FAILED to cover the request, which happened
// to work when len(rules) == 1 (the default policy) but let matching requests
// slip through as soon as a second, non-overlapping forbidden rule was
// present.
func rulesAllow(request *rbacv1.PolicyRule, rules ...rbacv1.PolicyRule) bool {
	for i := range rules {
		if ruleAllows(request, &rules[i]) {
			return false
		}
	}
	return true
}

func ruleAllows(request *rbacv1.PolicyRule, rule *rbacv1.PolicyRule) bool {
	if len(request.NonResourceURLs) == 0 {
		return verbMatches(rule, request.Verbs) &&
			apiGroupMatches(rule, request.APIGroups) &&
			resourceMatches(rule, request.Resources) &&
			resourceNameMatches(rule, request.ResourceNames)
	}
	return verbMatches(rule, request.Verbs) &&
		nonResourceURLMatches(rule, request.NonResourceURLs)
}

func verbMatches(rule *rbacv1.PolicyRule, requested []string) bool {
	if len(rule.Verbs) == 0 && len(requested) != 0 {
		return true
	}
	if len(requested) == 0 {
		return true
	}
	for _, rv := range rule.Verbs {
		if rv == rbacv1.VerbAll {
			return true
		}
		for _, req := range requested {
			if funk.Contains(rule.Verbs, req) {
				return true
			}
		}
	}
	return false
}

func apiGroupMatches(rule *rbacv1.PolicyRule, requested []string) bool {
	if len(rule.APIGroups) == 0 && len(requested) != 0 {
		return true
	}
	if len(requested) == 0 {
		return true
	}
	for _, rg := range rule.APIGroups {
		if rg == rbacv1.APIGroupAll {
			return true
		}
		for _, req := range requested {
			if funk.Contains(rule.APIGroups, req) {
				return true
			}
		}
	}
	return false
}

func resourceMatches(rule *rbacv1.PolicyRule, requested []string) bool {
	if len(rule.Resources) == 0 && len(requested) != 0 {
		return true
	}
	if len(requested) == 0 {
		return true
	}
	for _, rr := range rule.Resources {
		if rr == rbacv1.ResourceAll {
			return true
		}
		for _, req := range requested {
			if funk.Contains(rule.Resources, req) {
				return true
			}
		}
	}
	return false
}

func resourceNameMatches(rule *rbacv1.PolicyRule, requested []string) bool {
	if len(rule.ResourceNames) == 0 {
		return true
	}
	if len(requested) == 0 {
		return true
	}
	for _, req := range requested {
		if funk.Contains(rule.ResourceNames, req) {
			return true
		}
	}
	return false
}

// nonResourceURLMatches reports whether rule covers every entry in requested.
// A requested URL is considered covered when at least one URL in
// rule.NonResourceURLs either (a) equals it, (b) is "*", or (c) is of the
// form "prefix*" and the request starts with "prefix".
//
// The previous implementation nested the loops in the opposite order and
// returned early whenever any ruleURL failed to cover a req, which meant a
// rule containing two or more NonResourceURLs would reject requests that
// were plainly covered by one of them.
func nonResourceURLMatches(rule *rbacv1.PolicyRule, requested []string) bool {
	if len(requested) == 0 {
		return true
	}
	if len(rule.NonResourceURLs) == 0 {
		return false
	}
	for _, req := range requested {
		if !nonResourceURLCoveredBy(req, rule.NonResourceURLs) {
			return false
		}
	}
	return true
}

func nonResourceURLCoveredBy(req string, ruleURLs []string) bool {
	for _, ruleURL := range ruleURLs {
		if ruleURL == rbacv1.NonResourceAll {
			return true
		}
		if ruleURL == req {
			return true
		}
		if strings.HasSuffix(ruleURL, "*") {
			prefix := strings.TrimSuffix(ruleURL, "*")
			if strings.HasPrefix(req, prefix) {
				return true
			}
		}
	}
	return false
}
