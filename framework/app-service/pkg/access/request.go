package access

import (
	"fmt"
	"strings"
)

const (
	RequestPending   = "Pending"
	RequestApproved  = "Approved"
	RequestDenied    = "Denied"
	RequestWithdrawn = "Withdrawn"
	RequestExpired   = "Expired"
)

// AccessRequestDecision is a terminal or pending decision on AppAccessRequest.
type AccessRequestDecision struct {
	Current string
	Next    string
}

// ValidateDecisionTransition enforces one-shot terminal transitions for confirm/review.
func ValidateDecisionTransition(d AccessRequestDecision) error {
	cur := d.Current
	if cur == "" {
		cur = RequestPending
	}
	next := d.Next
	switch cur {
	case RequestPending:
		switch next {
		case RequestApproved, RequestDenied, RequestWithdrawn, RequestExpired:
			return nil
		default:
			return fmt.Errorf("pending request cannot transition to %q", next)
		}
	case RequestApproved, RequestDenied, RequestWithdrawn, RequestExpired:
		if next != cur {
			return fmt.Errorf("terminal state %q cannot change to %q", cur, next)
		}
		return nil
	default:
		return fmt.Errorf("unknown request state %q", cur)
	}
}

// ValidateCapabilityExpansion requires re-review when match scope widens.
func ValidateCapabilityExpansion(oldRules, newRules []HTTPRule, reReviewed bool) error {
	if err := ValidateHTTPRules(newRules); err != nil {
		return err
	}
	if !scopeWidened(oldRules, newRules) {
		return nil
	}
	if !reReviewed {
		return fmt.Errorf("widening path match requires artifact re-review and user re-confirmation")
	}
	return nil
}

func scopeWidened(oldRules, newRules []HTTPRule) bool {
	if len(newRules) > len(oldRules) {
		return true
	}
	type key struct{ m, p, t string }
	old := map[key]struct{}{}
	for _, r := range oldRules {
		t := r.PathType
		if t == "" {
			t = PathTypeExact
		}
		old[key{strings.ToUpper(r.Method), r.Path, t}] = struct{}{}
	}
	for _, r := range newRules {
		t := r.PathType
		if t == "" {
			t = PathTypeExact
		}
		k := key{strings.ToUpper(r.Method), r.Path, t}
		if _, ok := old[k]; !ok {
			// New exact is widen; also prefix/regex vs exact is widen.
			return true
		}
		if t == PathTypePrefix || t == PathTypeRegularExpression {
			// Treat presence of non-exact as potentially wider than prior exact-only sets.
			for o := range old {
				if o.m == k.m && o.t == PathTypeExact && o.p == r.Path {
					return true
				}
			}
		}
	}
	return false
}
