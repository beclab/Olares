package meshinagent

import (
	"sort"
	"strings"
)

const (
	AnnotDecide             = "gateway.olares.io/shared-caller-decide"
	AnnotDecideSource       = "gateway.olares.io/shared-caller-decide-source"
	AnnotDecideEdges        = "gateway.olares.io/shared-caller-edges"
	AnnotDecideRuleID       = "gateway.olares.io/shared-caller-rule-id"
	SettingOptOutMesh       = "mesh-inject"
	SettingAppRef           = "appRef"
	DecideSourceExplicit    = "explicit"
	DecideSourceRule        = "rule"
	DecideSourceEligibility = "eligibility"
	DecideSourceNone        = "none"

	// RuleAllowSharedLLMGateway auto-Decides Shared AI Router apps as callers
	// with empty edges (OPEN-01). Pure Shared callees (e.g. ollama*) do not match.
	RuleAllowSharedLLMGateway = "R-ALLOW-shared-llmgateway"
)

// Rule maps an application name (or prefix*) to named Shared callees.
// An allow rule with empty Callees still matches and injects (open domain).
type Rule struct {
	ID      string
	Match   string // exact app name or "prefix*"
	Deny    bool
	Callees []string
}

// RuleSet is the phase-1 decide rule table.
type RuleSet []Rule

// DefaultRules denies platform shells and allow-lists known Shared callers.
func DefaultRules() RuleSet {
	return RuleSet{
		{ID: "R-DENY-middleware", Match: "middleware*", Deny: true},
		{ID: "R-DENY-os-", Match: "os-", Deny: true},
		{ID: "R-DENY-olares-app", Match: "olares-app", Deny: true},
		{ID: "R-DENY-bfl", Match: "bfl", Deny: true},
		// Shared composite callers (opt-in by platform name). Empty Callees = OPEN-01.
		{ID: RuleAllowSharedLLMGateway, Match: "llmgateway*", Callees: nil},
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

// matchDenyExact matches deny rules by exact name or prefix* only (no bare HasPrefix),
// so "olares-app" does not deny "olares-application".
func matchDenyExact(appName string, r Rule) bool {
	m := strings.TrimSpace(r.Match)
	if m == "" {
		return false
	}
	if strings.HasSuffix(m, "*") {
		return strings.HasPrefix(appName, strings.TrimSuffix(m, "*"))
	}
	return appName == m
}

// ApplyRules returns generated callees or deny hit.
func ApplyRules(appName string, rules RuleSet) (callees []string, ruleID string, deny bool) {
	if id, hit := DenyRuleHit(appName, rules); hit {
		return nil, id, true
	}
	callees, ruleID = AllowRuleCallees(appName, rules)
	return callees, ruleID, false
}

// DenyRuleHit reports an exact/prefix* platform deny match.
func DenyRuleHit(appName string, rules RuleSet) (ruleID string, deny bool) {
	for _, r := range rules {
		if !r.Deny {
			continue
		}
		if matchDenyExact(appName, r) {
			return r.ID, true
		}
	}
	return "", false
}

// AllowRuleCallees returns callees from the first matching non-deny rule.
// A match with empty Callees is still a hit (ruleID != "") for open-domain callers.
func AllowRuleCallees(appName string, rules RuleSet) (callees []string, ruleID string) {
	for _, r := range rules {
		if r.Deny {
			continue
		}
		if !matchRule(appName, r) {
			continue
		}
		cp := append([]string(nil), r.Callees...)
		sort.Strings(cp)
		return cp, r.ID
	}
	return nil, ""
}

func isOptOut(settings map[string]string) bool {
	if settings == nil {
		return false
	}
	v := strings.ToLower(strings.TrimSpace(settings[SettingOptOutMesh]))
	return v == "disabled" || v == "false"
}

// Decide chooses mesh-in for a caller: deny and opt-out win, then named callees,
// then allow rules. Ordinary apps may fall through to eligibility (empty edges).
// Shared apps never use eligibility — only explicit callees, allow rules, or a
// preserved settings decide=true (ApplyDecide explicit path).
func Decide(appName string, settings map[string]string, rules RuleSet, sharedApp bool) DecideResult {
	if rules == nil {
		rules = DefaultRules()
	}
	res := DecideResult{Source: DecideSourceNone}
	if isOptOut(settings) {
		return res
	}
	if denyID, deny := DenyRuleHit(appName, rules); deny {
		res.RuleID = denyID
		return res
	}
	explicit := ParseCallees(settings)
	if len(explicit) > 0 {
		res.Inject = true
		res.Callees = explicit
		res.Source = DecideSourceExplicit
		return res
	}
	callees, ruleID := AllowRuleCallees(appName, rules)
	if ruleID != "" {
		res.Inject = true
		res.Callees = callees
		res.Source = DecideSourceRule
		res.RuleID = ruleID
		return res
	}
	if sharedApp {
		return res
	}
	res.Inject = true
	res.Source = DecideSourceEligibility
	return res
}

// DeclaresSharedCaller is the admission predicate: decide=true or named callees, not opt-out.
func DeclaresSharedCaller(settings map[string]string) bool {
	if settings == nil || isOptOut(settings) {
		return false
	}
	if d := strings.TrimSpace(settings[AnnotDecide]); strings.EqualFold(d, "false") {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(settings[AnnotDecide]), "true") {
		return true
	}
	return len(ParseCallees(settings)) > 0
}

// ApplyDecide mutates settings with decide facts and sets Changed vs previous annotate.
// When source is already explicit and decide=true (vendor settings patch), preserve
// those keys so rules cannot clear the opt-in. sharedApp disables eligibility.
func ApplyDecide(appName string, settings map[string]string, rules RuleSet, sharedApp bool) DecideResult {
	if settings == nil {
		return DecideResult{}
	}
	prevInject := strings.EqualFold(strings.TrimSpace(settings[AnnotDecide]), "true")
	prevEdges := strings.TrimSpace(settings[AnnotDecideEdges])
	prevSource := strings.TrimSpace(settings[AnnotDecideSource])
	if prevInject && strings.EqualFold(prevSource, DecideSourceExplicit) {
		res := DecideResult{
			Inject:  true,
			Source:  DecideSourceExplicit,
			Callees: splitCSV(prevEdges),
		}
		res.Changed = false
		return res
	}
	res := Decide(appName, settings, rules, sharedApp)
	if res.Inject {
		settings[AnnotDecide] = "true"
		settings[AnnotDecideSource] = res.Source
		settings[AnnotDecideEdges] = strings.Join(res.Callees, ",")
		if res.RuleID != "" {
			settings[AnnotDecideRuleID] = res.RuleID
		} else {
			delete(settings, AnnotDecideRuleID)
		}
		if res.Source == DecideSourceRule && len(res.Callees) > 0 && strings.TrimSpace(settings[SettingSharedAppDeps]) == "" {
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

func splitCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	seen := map[string]struct{}{}
	var out []string
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
	sort.Strings(out)
	return out
}

// ApplyExplicitSharedCaller writes vendor-equivalent opt-in settings
// (decide=true, source=explicit, edges empty). Used for manual AC before
// Manifest options.sharedCaller exists in beclab/api.
func ApplyExplicitSharedCaller(settings map[string]string) {
	if settings == nil {
		return
	}
	settings[AnnotDecide] = "true"
	settings[AnnotDecideSource] = DecideSourceExplicit
	settings[AnnotDecideEdges] = ""
	delete(settings, AnnotDecideRuleID)
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
