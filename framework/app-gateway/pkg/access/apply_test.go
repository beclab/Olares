package access

import (
	"encoding/json"
	"testing"
)

func TestBuildLinkerdObjectsActive(t *testing.T) {
	p := PolicySpec{
		Name: "a-to-b", Namespace: "app-b", DesiredState: DesiredActive,
		RouteMode: RouteModeDirect, ServerPort: 8080,
		SourceSANamespace: "app-a", SourceSAName: "sa",
		Rules: []HTTPRule{{Method: "GET", Path: "/state", PathType: PathTypeExact}},
	}
	plan, err := CompilePolicy(p, true)
	if err != nil {
		t.Fatal(err)
	}
	objs, err := BuildLinkerdObjects(plan, p)
	if err != nil || len(objs) != 4 {
		t.Fatalf("objs=%d err=%v", len(objs), err)
	}
	for _, o := range objs {
		if o.Object.GetLabels()[ManagedByLabel] != ManagedByAccessPolicy {
			t.Fatal("managed-by label")
		}
	}
}

func TestBuildLinkerdObjectsSkipWhenDisabled(t *testing.T) {
	p := PolicySpec{Name: "x", Namespace: "ns"}
	plan, _ := CompilePolicy(p, false)
	objs, err := BuildLinkerdObjects(plan, p)
	if err != nil || len(objs) != 0 {
		t.Fatalf("expected no objects, got %d err=%v", len(objs), err)
	}
}

func TestBuildLinkerdObjectsRevokedDeletes(t *testing.T) {
	p := PolicySpec{
		Name: "a-to-b", Namespace: "app-b", DesiredState: DesiredRevoked,
		RouteMode: RouteModeDirect, ServerPort: 8080,
		SourceSANamespace: "app-a", SourceSAName: "sa",
		Rules: []HTTPRule{{Method: "GET", Path: "/state", PathType: PathTypeExact}},
	}
	// Names from a prior active compile
	active := p
	active.DesiredState = DesiredActive
	plan, _ := CompilePolicy(active, true)
	p.DesiredState = DesiredRevoked
	skip, _ := CompilePolicy(p, true)
	objs, err := BuildLinkerdObjects(skip, p)
	if err != nil {
		t.Fatal(err)
	}
	// Revoked with SkipCompile from DesiredRevoked
	if !skip.SkipCompile {
		t.Fatal("expected skip on revoked")
	}
	objs, err = BuildLinkerdObjects(CompiledObjects{
		SkipCompile: true, ServerName: plan.ServerName, HTTPRouteName: plan.HTTPRouteName,
		AuthorizationPolicyName: plan.AuthorizationPolicyName, MeshTLSAuthName: plan.MeshTLSAuthName,
	}, p)
	if err != nil || len(objs) == 0 {
		t.Fatalf("expected delete markers")
	}
	for _, o := range objs {
		if !o.Delete {
			t.Fatal("expected delete")
		}
	}
}

func TestAdmitAdmissionReview(t *testing.T) {
	raw := []byte(`{
		"apiVersion":"admission.k8s.io/v1","kind":"AdmissionReview",
		"request":{"uid":"u1","operation":"CREATE","object":{
			"spec":{"capabilities":[{"http":[{"method":"GET","path":"/x","pathType":"Exact"}]}]}
		}}
	}`)
	out, err := AdmitAdmissionReview(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !containsAllowed(out, true) {
		t.Fatalf("want allow: %s", out)
	}
	bad := []byte(`{
		"apiVersion":"admission.k8s.io/v1","kind":"AdmissionReview",
		"request":{"uid":"u2","object":{"spec":{"rules":[{"method":"GET","path":"x","pathType":"Exact"}]}}}
	}`)
	out, err = AdmitAdmissionReview(bad)
	if err != nil || containsAllowed(out, true) {
		t.Fatalf("want deny: %s err=%v", out, err)
	}
}

func containsAllowed(raw []byte, want bool) bool {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return false
	}
	resp, _ := m["response"].(map[string]any)
	allowed, _ := resp["allowed"].(bool)
	return allowed == want
}
