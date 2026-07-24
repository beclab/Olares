package meshinagent

import (
	"sort"
	"strings"
)

const (
	AnnotDecide          = "gateway.olares.io/shared-caller-decide"
	AnnotDecideSource    = "gateway.olares.io/shared-caller-decide-source"
	AnnotDecideEdges     = "gateway.olares.io/shared-caller-edges"
	AnnotDecideRuleID    = "gateway.olares.io/shared-caller-rule-id"
	SettingOptOutMesh    = "mesh-inject"
	SettingAppRef        = "appRef"
	DecideSourceExplicit = "explicit"
	DecideSourceRule     = "rule"
	DecideSourceNone     = "none"
)

// Rule maps an application name (or prefix*) to named Shared callees.
type Rule struct {
	ID      string
	Match   string // exact app name or "prefix*"
	Deny    bool
	Callees []string
}

// RuleSet is the phase-1 decide rule table.
type RuleSet []Rule

// DefaultRules provides a minimal deny for platform middleware names.
func DefaultRules() RuleSet {
	return RuleSet{
		{ID: "R-DENY-middleware", Match: "middleware*", Deny: true},
		{ID: "R-DENY-os-", Match: "os-", Deny: true},
	}
}

// DecideResult is the install/upgrade SharedCallerDecide outcome.
type DecideResult struct {
	Inject  bool
	Callees []string
	Source  string
	RuleID  string
	Changed bool
}

// ParseCallees extracts named caller→callee edges from settings.
func ParseCallees(settings map[string]string) []string {
	if settings == nil {
		return nil
	}
	seen := map[string]struct{}{}
	var out []string
	add := func(raw string) {
		for _, p := range strings.Split(raw, ",") {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if _, ok := seen[p]; ok {
				continue
			}
			seen[p] = struct{}{}
			out = append(out, p)
		}
	}
	add(settings[SettingSharedAppDeps])
	add(settings[SettingClusterAppRef])
	add(settings[SettingAppRef])
	sort.Strings(out)
	return out
}

func matchRule(appName string, r Rule) bool {
	m := strings.TrimSpace(r.Match)
	if m == "" {
		return false
	}
	if strings.HasSuffix(m, "*") {
		return strings.HasPrefix(appName, strings.TrimSuffix(m, "*"))
	}
	return appName == m || strings.HasPrefix(appName, m)
}

// ApplyRules returns generated callees or deny hit.
func ApplyRules(appName string, rules RuleSet) (callees []string, ruleID string, deny bool) {
	for _, r := range rules {
		if !matchRule(appName, r) {
			continue
		}
		if r.Deny {
			return nil, r.ID, true
		}
		if len(r.Callees) > 0 {
			cp := append([]string(nil), r.Callees...)
			sort.Strings(cp)
			return cp, r.ID, false
		}
	}
	return nil, "", false
}

func isOptOut(settings map[string]string) bool {
	if settings == nil {
		return false
	}
	v := strings.ToLower(strings.TrimSpace(settings[SettingOptOutMesh]))
	return v == "disabled" || v == "false"
}

// Decide runs SharedCallerDecide: explicit named edges > rules; no callees => no inject.
// needsSharedAccess alone never injects (Q13).
func Decide(appName string, settings map[string]string, rules RuleSet) DecideResult {
	if rules == nil {
		rules = DefaultRules()
	}
	res := DecideResult{Source: DecideSourceNone}
	if isOptOut(settings) {
		return res
	}
	explicit := ParseCallees(settings)
	if len(explicit) > 0 {
		res.Inject = true
		res.Callees = explicit
		res.Source = DecideSourceExplicit
		return res
	}
	callees, ruleID, deny := ApplyRules(appName, rules)
	if deny {
		res.RuleID = ruleID
		return res
	}
	if len(callees) > 0 {
		res.Inject = true
		res.Callees = callees
		res.Source = DecideSourceRule
		res.RuleID = ruleID
	}
	return res
}

// DeclaresSharedCaller is the admission predicate: non-empty named callees (and not opt-out).
func DeclaresSharedCaller(settings map[string]string) bool {
	if settings == nil || isOptOut(settings) {
		return false
	}
	if d := strings.TrimSpace(settings[AnnotDecide]); strings.EqualFold(d, "false") {
		return false
	}
	return len(ParseCallees(settings)) > 0
}

// ApplyDecide mutates settings with decide facts and sets Changed vs previous annotate.
func ApplyDecide(appName string, settings map[string]string, rules RuleSet) DecideResult {
	if settings == nil {
		return DecideResult{}
	}
	prevInject := strings.EqualFold(strings.TrimSpace(settings[AnnotDecide]), "true")
	prevEdges := strings.TrimSpace(settings[AnnotDecideEdges])
	res := Decide(appName, settings, rules)
	if res.Inject {
		settings[AnnotDecide] = "true"
		settings[AnnotDecideSource] = res.Source
		settings[AnnotDecideEdges] = strings.Join(res.Callees, ",")
		if res.RuleID != "" {
			settings[AnnotDecideRuleID] = res.RuleID
		} else {
			delete(settings, AnnotDecideRuleID)
		}
		if res.Source == DecideSourceRule && strings.TrimSpace(settings[SettingSharedAppDeps]) == "" {
			settings[SettingSharedAppDeps] = strings.Join(res.Callees, ",")
		}
	} else {
		settings[AnnotDecide] = "false"
		settings[AnnotDecideSource] = res.Source
		settings[AnnotDecideEdges] = ""
		delete(settings, AnnotDecideRuleID)
	}
	res.Changed = prevInject != res.Inject || (res.Inject && prevEdges != strings.Join(res.Callees, ","))
	return res
}

// WriteDecideAnnotations copies decide settings onto Application annotations.
func WriteDecideAnnotations(ann map[string]string, settings map[string]string) map[string]string {
	if ann == nil {
		ann = map[string]string{}
	}
	if settings == nil {
		return ann
	}
	for _, k := range []string{AnnotDecide, AnnotDecideSource, AnnotDecideEdges, AnnotDecideRuleID} {
		if v := strings.TrimSpace(settings[k]); v != "" {
			ann[k] = v
		} else {
			delete(ann, k)
		}
	}
	return ann
}

// DecideSettingKeys are settings keys written by ApplyDecide that update paths must merge.
func DecideSettingKeys() []string {
	return []string{
		AnnotDecide, AnnotDecideSource, AnnotDecideEdges, AnnotDecideRuleID,
		SettingSharedAppDeps, SettingClusterAppRef, SettingAppRef, SettingNeedsSharedAccess,
	}
}
