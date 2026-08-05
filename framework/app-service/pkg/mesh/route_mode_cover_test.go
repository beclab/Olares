package mesh

import "testing"

func TestEvaluateEntranceRouteModeGatewayReadyVacuous(t *testing.T) {
	if !EvaluateEntranceRouteModeGatewayReady(nil) {
		t.Fatal("empty targets must be Ready")
	}
}

func TestEvaluateEntranceRouteModeGatewayReadyAllGateway(t *testing.T) {
	targets := []EntranceRouteModeTarget{
		{Namespace: "user-space-a", AppName: "app1", Mode: "gateway"},
		{Namespace: "user-space-a", AppName: "app2", Mode: "gateway"},
	}
	if !EvaluateEntranceRouteModeGatewayReady(targets) {
		t.Fatal("all gateway must be Ready")
	}
}

func TestEvaluateEntranceRouteModeGatewayReadyPartial(t *testing.T) {
	targets := []EntranceRouteModeTarget{
		{Namespace: "ns", AppName: "a", Mode: "gateway"},
		{Namespace: "ns", AppName: "b", Mode: ""},
	}
	if EvaluateEntranceRouteModeGatewayReady(targets) {
		t.Fatal("partial gateway must not be Ready")
	}
	if !HasPartialEntranceRouteModeGateway(targets) {
		t.Fatal("expected partial detection")
	}
}

func TestEvaluateEntranceRouteModeGatewayReadyDirect(t *testing.T) {
	targets := []EntranceRouteModeTarget{
		{Namespace: "ns", AppName: "a", Mode: "direct"},
	}
	if EvaluateEntranceRouteModeGatewayReady(targets) {
		t.Fatal("direct must not be Ready")
	}
}
