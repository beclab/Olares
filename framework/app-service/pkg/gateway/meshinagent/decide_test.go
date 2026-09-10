package meshinagent

import (
	"testing"
	"time"
)

func TestDecideExplicitCallees(t *testing.T) {
	s := map[string]string{SettingSharedAppDeps: "ollama,litellm"}
	r := Decide("myapp", s, nil, false)
	if !r.Inject || r.Source != DecideSourceExplicit || len(r.Callees) != 2 {
		t.Fatalf("got %+v", r)
	}
}

func TestDecideEligibilityWithoutCallees(t *testing.T) {
	s := map[string]string{SettingNeedsSharedAccess: "true"}
	r := Decide("myapp", s, nil, false)
	if !r.Inject || r.Source != DecideSourceEligibility || len(r.Callees) != 0 {
		t.Fatalf("needsSharedAccess without callees must inject: %+v", r)
	}
	r2 := Decide("plain-app", map[string]string{}, nil, false)
	if !r2.Inject || r2.Source != DecideSourceEligibility {
		t.Fatalf("empty settings must inject with empty callees: %+v", r2)
	}
}

func TestDecideSharedDefaultCaller(t *testing.T) {
	r := Decide("ollamallmbase", map[string]string{SettingNeedsSharedAccess: "true"}, DefaultRules(), true)
	if !r.Inject || r.Source != DecideSourceSharedDefault {
		t.Fatalf("Shared app must default to caller: %+v", r)
	}
	if len(r.Callees) != 0 {
		t.Fatalf("shared-default edges must be empty: %+v", r)
	}
}

func TestDecideSharedOptOutBlocksDefault(t *testing.T) {
	r := Decide("ollamallmbase", map[string]string{SettingOptOutMesh: "disabled"}, DefaultRules(), true)
	if r.Inject {
		t.Fatalf("opt-out must block shared-default: %+v", r)
	}
}

func TestDecideSharedLLMGatewayUsesSharedDefault(t *testing.T) {
	r := Decide("llmgatewayv3", map[string]string{}, DefaultRules(), true)
	if !r.Inject || r.Source != DecideSourceSharedDefault || r.RuleID != "" {
		t.Fatalf("Shared llmgateway* must use shared-default (no name allow rule): %+v", r)
	}
	if len(r.Callees) != 0 {
		t.Fatalf("shared-default edges must be empty (OPEN-01): %+v", r)
	}
}

func TestDecideRuleAllow(t *testing.T) {
	rules := RuleSet{{ID: "R-ALLOW-demo", Match: "demo*", Callees: []string{"ollama"}}}
	r := Decide("demo-chat", map[string]string{}, rules, false)
	if !r.Inject || r.Source != DecideSourceRule || r.RuleID != "R-ALLOW-demo" {
		t.Fatalf("got %+v", r)
	}
}

func TestDecideRuleAllowEmptyCallees(t *testing.T) {
	rules := RuleSet{{ID: "R-ALLOW-open", Match: "router*", Callees: nil}}
	r := Decide("router-a", map[string]string{}, rules, true)
	if !r.Inject || r.Source != DecideSourceRule || r.RuleID != "R-ALLOW-open" {
		t.Fatalf("empty-callee allow must inject: %+v", r)
	}
}

func TestDecideRuleDeny(t *testing.T) {
	rules := RuleSet{{ID: "R-DENY-demo", Match: "middleware*", Deny: true}}
	r := Decide("middleware-x", map[string]string{SettingSharedAppDeps: "x"}, rules, false)
	if r.Inject {
		t.Fatalf("deny must win over explicit callees: %+v", r)
	}
	r2 := Decide("middleware-x", map[string]string{}, rules, false)
	if r2.Inject {
		t.Fatalf("deny must block: %+v", r2)
	}
}

func TestDecidePlatformShellDeny(t *testing.T) {
	r := Decide("olares-app", map[string]string{}, DefaultRules(), false)
	if r.Inject {
		t.Fatalf("olares-app must be denied: %+v", r)
	}
	r2 := Decide("olares-app", map[string]string{SettingAppRef: "ollama"}, DefaultRules(), false)
	if r2.Inject {
		t.Fatalf("olares-app deny must win over appRef: %+v", r2)
	}
	r3 := Decide("olares-application", map[string]string{}, DefaultRules(), false)
	if !r3.Inject {
		t.Fatalf("olares-application must not match exact deny olares-app: %+v", r3)
	}
}

func TestDecideOptOut(t *testing.T) {
	s := map[string]string{SettingSharedAppDeps: "ollama", SettingOptOutMesh: "disabled"}
	r := Decide("myapp", s, nil, false)
	if r.Inject {
		t.Fatal("opt-out must win")
	}
}

func TestApplyDecideIdempotent(t *testing.T) {
	s := map[string]string{SettingClusterAppRef: "shared-a"}
	r1 := ApplyDecide("app", s, nil, false)
	if !r1.Inject || !r1.Changed {
		t.Fatalf("first: %+v", r1)
	}
	r2 := ApplyDecide("app", s, nil, false)
	if r2.Changed {
		t.Fatalf("second must not change: %+v settings=%v", r2, s)
	}
}

func TestApplyDecidePreservesExplicit(t *testing.T) {
	s := map[string]string{
		AnnotDecide:       "true",
		AnnotDecideSource: DecideSourceExplicit,
		AnnotDecideEdges:  "",
	}
	r := ApplyDecide("middleware-x", s, RuleSet{{ID: "R-DENY-demo", Match: "middleware*", Deny: true}}, true)
	if !r.Inject || r.Source != DecideSourceExplicit {
		t.Fatalf("explicit must be preserved even under deny rule: %+v", r)
	}
	if s[AnnotDecide] != "true" || s[AnnotDecideSource] != DecideSourceExplicit {
		t.Fatalf("settings mutated: %v", s)
	}
	if r.Changed {
		t.Fatalf("preserve path must report Changed=false: %+v", r)
	}
}

func TestApplyDecideSharedLLMGatewaySharedDefault(t *testing.T) {
	s := map[string]string{}
	r := ApplyDecide("llmgatewayv3", s, DefaultRules(), true)
	if !r.Inject || r.Source != DecideSourceSharedDefault || s[AnnotDecide] != "true" {
		t.Fatalf("got %+v settings=%v", r, s)
	}
	if s[AnnotDecideSource] != DecideSourceSharedDefault {
		t.Fatalf("source: %v", s)
	}
	if s[AnnotDecideRuleID] != "" {
		t.Fatalf("rule id must be empty: %v", s)
	}
	if s[AnnotDecideEdges] != "" {
		t.Fatalf("edges must stay empty: %v", s)
	}
}

func TestApplyDecideSharedDefaultCaller(t *testing.T) {
	s := map[string]string{SettingNeedsSharedAccess: "true"}
	r := ApplyDecide("ollamallmbase", s, DefaultRules(), true)
	if !r.Inject || r.Source != DecideSourceSharedDefault || s[AnnotDecide] != "true" {
		t.Fatalf("Shared must default decide: %+v settings=%v", r, s)
	}
	if s[AnnotDecideSource] != DecideSourceSharedDefault || s[AnnotDecideEdges] != "" {
		t.Fatalf("shared-default settings: %v", s)
	}
}

func TestApplyDecideSharedRenameStillCaller(t *testing.T) {
	s := map[string]string{}
	r := ApplyDecide("router", s, DefaultRules(), true)
	if !r.Inject || r.Source != DecideSourceSharedDefault {
		t.Fatalf("renamed Shared must still decide: %+v", r)
	}
}

func TestApplyExplicitSharedCaller(t *testing.T) {
	s := map[string]string{}
	ApplyExplicitSharedCaller(s)
	if s[AnnotDecide] != "true" || s[AnnotDecideSource] != DecideSourceExplicit || s[AnnotDecideEdges] != "" {
		t.Fatalf("got %v", s)
	}
	if !DeclaresSharedCaller(s) {
		t.Fatal("explicit settings must declare")
	}
	// Survive ApplyDecide under Shared mode.
	r := ApplyDecide("custom-shared-app", s, DefaultRules(), true)
	if !r.Inject || r.Source != DecideSourceExplicit {
		t.Fatalf("manual explicit must survive: %+v", r)
	}
}

func TestDeclaresSharedCaller(t *testing.T) {
	if DeclaresSharedCaller(map[string]string{SettingNeedsSharedAccess: "true"}) {
		t.Fatal("needsSharedAccess alone without decide annotate must not declare")
	}
	if !DeclaresSharedCaller(map[string]string{SettingAppRef: "x"}) {
		t.Fatal("appRef")
	}
	if !DeclaresSharedCaller(map[string]string{AnnotDecide: "true", AnnotDecideSource: DecideSourceEligibility}) {
		t.Fatal("decide=true must declare")
	}
	if DeclaresSharedCaller(map[string]string{AnnotDecide: "false"}) {
		t.Fatal("decide=false must not declare")
	}
}

func TestRolloutQueueK2(t *testing.T) {
	q := NewRolloutQueue(2)
	if !q.TryAcquire("a") || !q.TryAcquire("b") {
		t.Fatal("first two must acquire")
	}
	if q.TryAcquire("c") {
		t.Fatal("third must wait")
	}
	if q.ActiveCount() != 2 || q.WaitingCount() != 1 {
		t.Fatalf("active=%d waiting=%d", q.ActiveCount(), q.WaitingCount())
	}
	next, ok := q.Release()
	if !ok || next != "c" {
		t.Fatalf("next=%q ok=%v", next, ok)
	}
	if q.ActiveCount() != 2 {
		t.Fatalf("after promote active=%d", q.ActiveCount())
	}
}

func TestRetryBackoff(t *testing.T) {
	if RetryBackoff(0) != 10*time.Second {
		t.Fatal(RetryBackoff(0))
	}
	if RetryBackoff(10) != RolloutBackoffCap {
		t.Fatal(RetryBackoff(10))
	}
}
