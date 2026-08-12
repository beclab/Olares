package mesh

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestEvaluateCanRemoveOES(t *testing.T) {
	cases := []struct {
		name                                string
		steady, inbound, outbound, rollback bool
		want                                bool
	}{
		{"all true", true, true, true, true, true},
		{"no steady", false, true, true, true, false},
		{"no inbound", true, false, true, true, false},
		{"no outbound", true, true, false, true, false},
		{"no rollback", true, true, true, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := EvaluateCanRemoveOES(tc.steady, tc.inbound, tc.outbound, tc.rollback)
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestStoreLoadSteadyGate(t *testing.T) {
	kube := fake.NewSimpleClientset()
	st := &SteadyGateState{
		Phase:         SteadyGateReadyPhase,
		TargetVersion: "1.12.9",
		Checkpoint:    "commit",
		Conditions: map[string]bool{
			"ZeroOesInventory":    true,
			ConditionL4ProxyReady: true,
		},
	}
	if err := StoreSteadyGate(context.Background(), kube, st); err != nil {
		t.Fatal(err)
	}
	got, err := LoadSteadyGate(context.Background(), kube)
	if err != nil {
		t.Fatal(err)
	}
	if got.Phase != SteadyGateReadyPhase || got.Checkpoint != "commit" {
		t.Fatalf("got %#v", got)
	}
	if !IsSteadyGateReady(context.Background(), kube) {
		t.Fatal("expected ready")
	}
	if !IsL4EdgePEPReady(context.Background(), kube) {
		t.Fatal("expected L4ProxyReady condition to unlock edge PEP")
	}
}

func TestLoadSteadyGateMissing(t *testing.T) {
	got, err := LoadSteadyGate(context.Background(), fake.NewSimpleClientset())
	if err != nil {
		t.Fatal(err)
	}
	if IsSteadyGateReady(context.Background(), fake.NewSimpleClientset()) {
		t.Fatal("missing CM must not be ready")
	}
	if got.Phase != "Pending" {
		t.Fatalf("phase = %q", got.Phase)
	}
	_, _ = fake.NewSimpleClientset(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: SteadyGateConfigMap, Namespace: SteadyGateNamespace},
		Data:       map[string]string{"phase": "Pending"},
	}).CoreV1().ConfigMaps(SteadyGateNamespace).Get(context.Background(), SteadyGateConfigMap, metav1.GetOptions{})
}

func TestIsL4EdgePEPReadyRequiresConditionOrDeployment(t *testing.T) {
	kube := fake.NewSimpleClientset()
	if IsL4EdgePEPReady(context.Background(), kube) {
		t.Fatal("missing gate and l4 Deployment must not report L4 edge PEP ready")
	}
	st := &SteadyGateState{Phase: "Pending", Conditions: map[string]bool{ConditionL4ProxyReady: false}}
	if err := StoreSteadyGate(context.Background(), kube, st); err != nil {
		t.Fatal(err)
	}
	if IsL4EdgePEPReady(context.Background(), kube) {
		t.Fatal("false L4ProxyReady without l4 Deployment must not unlock")
	}
}
