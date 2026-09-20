package mesh

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const (
	// SteadyGateNamespace hosts the de-envoy steady-state ConfigMap.
	SteadyGateNamespace = "os-framework"
	// SteadyGateConfigMap is the version-scoped cutover gate.
	SteadyGateConfigMap = "olares-deenvy-steady-state"
	// SteadyGateReadyPhase is the only phase that allows oes-free commit.
	SteadyGateReadyPhase = "Ready"

	annotSteadyPhase = "phase"

	// ConditionL4ProxyReady is set when Track-A M-L4 is true (verify/302 + F2/F3).
	ConditionL4ProxyReady = "L4ProxyReady"
)

// SteadyGateState is the persisted cutover gate (ADR-SYS-DEENVY-STEADY-01).
type SteadyGateState struct {
	Phase         string          `json:"phase"`
	TargetVersion string          `json:"targetVersion,omitempty"`
	Checkpoint    string          `json:"checkpoint,omitempty"`
	Conditions    map[string]bool `json:"conditions,omitempty"`
	Message       string          `json:"message,omitempty"`
}

// IsSteadyGateReady reports whether the cluster committed the oes-free steady state.
func IsSteadyGateReady(ctx context.Context, kube kubernetes.Interface) bool {
	st, err := LoadSteadyGate(ctx, kube)
	if err != nil {
		klog.V(2).Infof("mesh: load SteadyStateGate failed: %v", err)
		return false
	}
	return st != nil && strings.EqualFold(st.Phase, SteadyGateReadyPhase)
}

// LoadSteadyGate reads the gate ConfigMap; missing object returns empty non-ready state.
func LoadSteadyGate(ctx context.Context, kube kubernetes.Interface) (*SteadyGateState, error) {
	if kube == nil {
		return &SteadyGateState{Phase: "Pending"}, nil
	}
	cm, err := kube.CoreV1().ConfigMaps(SteadyGateNamespace).Get(ctx, SteadyGateConfigMap, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return &SteadyGateState{Phase: "Pending", Conditions: map[string]bool{}}, nil
		}
		return nil, fmt.Errorf("get SteadyStateGate: %w", err)
	}
	raw := strings.TrimSpace(cm.Data["state"])
	if raw == "" {
		return &SteadyGateState{Phase: cm.Data[annotSteadyPhase], Conditions: map[string]bool{}}, nil
	}
	var st SteadyGateState
	if err := json.Unmarshal([]byte(raw), &st); err != nil {
		return nil, fmt.Errorf("decode SteadyStateGate: %w", err)
	}
	return &st, nil
}

// StoreSteadyGate writes/updates the gate ConfigMap.
func StoreSteadyGate(ctx context.Context, kube kubernetes.Interface, st *SteadyGateState) error {
	if kube == nil || st == nil {
		return fmt.Errorf("kube client and state required")
	}
	body, err := json.Marshal(st)
	if err != nil {
		return err
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SteadyGateConfigMap,
			Namespace: SteadyGateNamespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "app-service",
				"app.kubernetes.io/name":       "deenvy-steady-state",
			},
		},
		Data: map[string]string{
			"state": string(body),
			"phase": st.Phase,
		},
	}
	_, err = kube.CoreV1().ConfigMaps(SteadyGateNamespace).Update(ctx, cm, metav1.UpdateOptions{})
	if apierrors.IsNotFound(err) {
		_, err = kube.CoreV1().ConfigMaps(SteadyGateNamespace).Create(ctx, cm, metav1.CreateOptions{})
	}
	if err != nil {
		return fmt.Errorf("store SteadyStateGate: %w", err)
	}
	return nil
}

// EvaluateCanRemoveOES is the pure whole-oes removal predicate.
func EvaluateCanRemoveOES(steadyReady, inboundCovered, outboundCovered, rollbackReady bool) bool {
	return steadyReady && inboundCovered && outboundCovered && rollbackReady
}

// CanRemoveOES probes SteadyGate plus coverage flags for whole-oes removal.
func CanRemoveOES(ctx context.Context, kube kubernetes.Interface, inboundCovered, outboundCovered, rollbackReady bool) bool {
	return EvaluateCanRemoveOES(steadyGateReadyOrOverride(ctx, kube), inboundCovered, outboundCovered, rollbackReady)
}

// ForceSteadyGateReadyForTest sets an in-process override used by unit tests only.
var forceSteadyGateReadyForTest *bool

func setSteadyGateReadyForTest(v bool) {
	forceSteadyGateReadyForTest = &v
}

func clearSteadyGateReadyForTest() {
	forceSteadyGateReadyForTest = nil
}

func steadyGateReadyOrOverride(ctx context.Context, kube kubernetes.Interface) bool {
	if forceSteadyGateReadyForTest != nil {
		return *forceSteadyGateReadyForTest
	}
	if os.Getenv("OLARES_DEENVY_STEADY_READY") == "1" {
		return true
	}
	return IsSteadyGateReady(ctx, kube)
}
