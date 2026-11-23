// Package access validates access.olares.io authorization objects before admit.
package access

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	PathTypeExact              = "Exact"
	PathTypePrefix             = "PathPrefix"
	PathTypeRegularExpression  = "RegularExpression"
	MaxRulesPerCapability      = 64
	MaxManagedObjectsPerPolicy = 256
)

// HTTPRule is the admission view of one capability HTTP match rule.
type HTTPRule struct {
	Method   string
	Path     string
	PathType string
}

// ValidateHTTPRules enforces pathType and capacity limits from the CRD contract.
func ValidateHTTPRules(rules []HTTPRule) error {
	if len(rules) == 0 {
		return fmt.Errorf("at least one HTTP rule is required")
	}
	if len(rules) > MaxRulesPerCapability {
		return fmt.Errorf("capability exceeds %d HTTP rules", MaxRulesPerCapability)
	}
	for i, r := range rules {
		pt := r.PathType
		if pt == "" {
			pt = PathTypeExact
		}
		switch pt {
		case PathTypeExact, PathTypePrefix, PathTypeRegularExpression:
		default:
			return fmt.Errorf("rule[%d]: unsupported pathType %q", i, r.PathType)
		}
		if strings.TrimSpace(r.Method) == "" {
			return fmt.Errorf("rule[%d]: method is required", i)
		}
		if r.Path == "" {
			return fmt.Errorf("rule[%d]: path is required", i)
		}
		if pt == PathTypeRegularExpression {
			if !strings.HasPrefix(r.Path, "^/") {
				return fmt.Errorf("rule[%d]: RegularExpression path must start with ^/", i)
			}
			if !strings.HasSuffix(r.Path, "$") {
				return fmt.Errorf("rule[%d]: RegularExpression path must be anchored", i)
			}
			if _, err := regexp.Compile(r.Path); err != nil {
				return fmt.Errorf("rule[%d]: invalid RegularExpression", i)
			}
			continue
		}
		if !strings.HasPrefix(r.Path, "/") {
			return fmt.Errorf("rule[%d]: path must start with /", i)
		}
		if pt == PathTypePrefix && r.Path != "/" && strings.HasSuffix(r.Path, "/") {
			return fmt.Errorf("rule[%d]: PathPrefix cannot end with / except for root", i)
		}
	}
	return nil
}

// ValidateManagedObjectBudget rejects policy projections that exceed the per-policy cap.
func ValidateManagedObjectBudget(count int) error {
	if count < 0 {
		return fmt.Errorf("managed object count cannot be negative")
	}
	if count > MaxManagedObjectsPerPolicy {
		return fmt.Errorf("policy exceeds %d managed objects", MaxManagedObjectsPerPolicy)
	}
	return nil
}

// PolicyCompilationEnabled reports whether runtime ConfigMap enables compiler output.
func PolicyCompilationEnabled(raw string) bool {
	return strings.EqualFold(strings.TrimSpace(raw), "true") || strings.TrimSpace(raw) == "1"
}
