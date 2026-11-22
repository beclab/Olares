package access

import "testing"

func TestCompilePolicyDisabled(t *testing.T) {
	out, err := CompilePolicy(PolicySpec{Name: "a", Namespace: "ns"}, false)
	if err != nil || !out.SkipCompile {
		t.Fatalf("got %+v err=%v", out, err)
	}
}

func TestCompilePolicyDirectAndRevoke(t *testing.T) {
	spec := PolicySpec{
		Name: "market-catalog", Namespace: "market-alice", DesiredState: DesiredActive,
		RouteMode: RouteModeDirect, ServerPort: 8080,
		SourceSANamespace: "chart-a", SourceSAName: "chart-a",
		Rules: []HTTPRule{{Method: "GET", Path: "/catalog", PathType: PathTypeExact}},
	}
	out, err := CompilePolicy(spec, true)
	if err != nil || out.SkipCompile || out.IdentityPrincipal == "" {
		t.Fatalf("got %+v err=%v", out, err)
	}
	spec.DesiredState = DesiredRevoked
	out, err = CompilePolicy(spec, true)
	if err != nil || !out.SkipCompile {
		t.Fatalf("revoke: %+v err=%v", out, err)
	}
}

func TestCompilePolicyViaGatewayForbidsMissingGW(t *testing.T) {
	_, err := CompilePolicy(PolicySpec{
		Name: "x", Namespace: "b", RouteMode: RouteModeViaAppGateway, ServerPort: 8080,
		Rules: []HTTPRule{{Method: "GET", Path: "/state"}},
	}, true)
	if err == nil {
		t.Fatal("expected gateway SA error")
	}
}
