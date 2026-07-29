package meshinagent

import (
	"testing"
	"time"
)

func TestDecideExplicitCallees(t *testing.T) {
	s := map[string]string{SettingSharedAppDeps: "ollama,litellm"}
	r := Decide("myapp", s, nil)
	if !r.Inject || r.Source != DecideSourceExplicit || len(r.Callees) != 2 {
		t.Fatalf("got %+v", r)
	}
}

func TestDecideEligibilityWithoutCallees(t *testing.T) {
	s := map[string]string{SettingNeedsSharedAccess: "true"}
	r := Decide("myapp", s, nil)
	if !r.Inject || r.Source != DecideSourceEligibility || len(r.Callees) != 0 {
		t.Fatalf("needsSharedAccess without callees must inject: %+v", r)
	}
	r2 := Decide("plain-app", map[string]string{}, nil)
	if !r2.Inject || r2.Source != DecideSourceEligibility {
		t.Fatalf("empty settings must inject with empty callees: %+v", r2)
	}
}

func TestDecideRuleAllow(t *testing.T) {
	rules := RuleSet{{ID: "R-ALLOW-demo", Match: "demo*", Callees: []string{"ollama"}}}
	r := Decide("demo-chat", map[string]string{}, rules)
	if !r.Inject || r.Source != DecideSourceRule || r.RuleID != "R-ALLOW-demo" {
		t.Fatalf("got %+v", r)
	}
}

func TestDecideRuleDeny(t *testing.T) {
	r := Decide("middleware-x", map[string]string{SettingSharedAppDeps: "x"}, DefaultRules())
	if r.Inject {
		t.Fatalf("deny must win over explicit callees: %+v", r)
	}
	r2 := Decide("middleware-x", map[string]string{}, DefaultRules())
	if r2.Inject {
		t.Fatalf("deny must block: %+v", r2)
	}
}

func TestDecidePlatformShellDeny(t *testing.T) {
	for _, name := range []string{"olares-app", "bfl"} {
		r := Decide(name, map[string]string{}, DefaultRules())
		if r.Inject {
			t.Fatalf("%s must be denied: %+v", name, r)
		}
		r2 := Decide(name, map[string]string{SettingAppRef: "ollama"}, DefaultRules())
		if r2.Inject {
			t.Fatalf("%s deny must win over appRef: %+v", name, r2)
		}
	}
	r3 := Decide("olares-application", map[string]string{}, DefaultRules())
	if !r3.Inject {
		t.Fatalf("olares-application must not match exact deny olares-app: %+v", r3)
	}
}

func TestDecideOptOut(t *testing.T) {
	s := map[string]string{SettingSharedAppDeps: "ollama", SettingOptOutMesh: "disabled"}
	r := Decide("myapp", s, nil)
	if r.Inject {
		t.Fatal("opt-out must win")
	}
}

func TestApplyDecideIdempotent(t *testing.T) {
	s := map[string]string{SettingClusterAppRef: "shared-a"}
	r1 := ApplyDecide("app", s, nil)
	if !r1.Inject || !r1.Changed {
		t.Fatalf("first: %+v", r1)
	}
	r2 := ApplyDecide("app", s, nil)
	if r2.Changed {
		t.Fatalf("second must not change: %+v settings=%v", r2, s)
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
