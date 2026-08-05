package upgrade

import (
	"context"
	"fmt"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/beclab/Olares/cli/pkg/core/logger"
)

// evaluateAcceptSuitePassed is the pure accept gate: inventory + inbound coverage
// conditions must all be true (atomic package). Partial coverage fails.
func evaluateAcceptSuitePassed(conds map[string]bool) bool {
	if conds == nil {
		return false
	}
	required := []string{
		"ZeroOesInventory",
		"EntranceExtAuthCovered",
		"EntranceCookieCovered",
		"EntranceProbeBypassReady",
		"EntranceAuxCovered",
		"RouteModeGateway",
	}
	for _, k := range required {
		if !conds[k] {
			return false
		}
	}
	return true
}

func runAcceptSuiteProbes(ctx context.Context, kube kubernetes.Interface, dc dynamic.Interface) (map[string]bool, string, error) {
	conds := map[string]bool{}
	invOK, msg := inventoryZeroUnauthorizedOES(ctx, kube)
	conds["ZeroOesInventory"] = invOK
	conds["KeyPodsReady"] = true

	if dc == nil {
		assignExtAuthDepConditions(conds, false, false)
		assignCookieDepConditions(conds, false)
		assignProbeBypassDepConditions(conds, false)
		assignAuxDepConditions(conds, false)
		assignRouteModeDepConditions(conds, false)
		conds["AcceptSuitePassed"] = false
		if !invOK {
			return conds, msg, nil
		}
		return conds, "dynamic client unavailable for accept probes", nil
	}

	egOK := true // EG readiness is tracked in WaitDeps; accept focuses on coverage.
	extOK, err := probeEntranceExtAuthCovered(ctx, dc)
	if err != nil {
		logger.Errorf("deenvy: accept ExtAuth probe: %v", err)
		extOK = false
	}
	assignExtAuthDepConditions(conds, egOK, extOK)
	_ = ensureCookieProbeOK(ctx, dc, conds)
	_ = ensureProbeBypassProbeOK(ctx, dc, conds)
	_ = ensureAuxProbeOK(ctx, dc, conds)
	_ = ensureRouteModeGatewayOK(ctx, dc, conds)

	passed := evaluateAcceptSuitePassed(conds)
	conds["AcceptSuitePassed"] = passed
	if !invOK {
		return conds, msg, nil
	}
	if !passed {
		return conds, fmt.Sprintf("accept suite incomplete: %#v", conds), nil
	}
	return conds, "accept suite passed", nil
}
