package compute

import (
	"context"
	"errors"
	"fmt"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ErrAmbiguousComputeMode is returned by AutoSelectMode when more than one
// non-cpu mode declared by the app is also runnable on the cluster, so the
// caller must surface the ambiguity to the user (e.g. by returning the
// install compute plan and asking for an explicit selectedGpuType).
var ErrAmbiguousComputeMode = errors.New("multiple gpu types runnable on this cluster; please specify selectedGpuType")

// AutoSelectMode picks a compute mode for `appCfg` when the install caller
// did not supply one explicitly. It is meant to make the common cases
// "obvious choice, just pick it" frictionless without ever silently picking
// the wrong mode when more than one is plausible.
//
// Algorithm:
//
//  1. Build cluster.gpuTypes = the set of non-cpu GPU types that show up on
//     any node label gpu.bytetrade.io/type. The cluster always implicitly
//     supports cpu mode because every GPU node can also run a cpu-mode pod;
//     cpu therefore is never enumerated in cluster.gpuTypes.
//
//  2. Build app.modes = the set of modes the manifest declares support for.
//     See AppSupportedModes for the exact rules.
//
//  3. valid = (modes the app supports) ∩ (modes runnable on this cluster).
//     - cpu is "runnable" iff app.modes contains cpu.
//     - any non-cpu mode m is "runnable" iff m ∈ cluster.gpuTypes.
//
//  4. Decide:
//     - exactly one non-cpu entry in valid → return that. (Single GPU
//     candidate, no ambiguity even if cpu fallback is also valid.)
//     - zero non-cpu entries in valid, but cpu ∈ valid → return cpu.
//     (App is effectively cpu-only on this cluster.)
//     - otherwise (multiple non-cpu candidates, or the intersection is
//     empty) → return an error and let the caller require an explicit
//     SelectedGpuType.
//
// Worked examples (cluster has nvidia + cpu fallback):
//
//	app supports {nvidia, gb10, strix-halo}     -> nvidia
//	app supports {nvidia, gb10, cpu}            -> nvidia (one non-cpu match)
//	app supports {strix-halo, apple-m}          -> error (no overlap)
//	app supports {cpu}                          -> cpu
//
// Worked examples (cluster has nvidia + strix-halo + cpu fallback):
//
//	app supports {nvidia}                       -> nvidia
//	app supports {strix-halo}                   -> strix-halo
//	app supports {nvidia, strix-halo}           -> error (ambiguous)
//	app supports {cpu}                          -> cpu
//	app supports {cpu, apple-m}                 -> cpu (apple-m not on cluster)
func AutoSelectMode(ctx context.Context, c client.Client, appCfg *appcfg.ApplicationConfig) (string, error) {
	if appCfg == nil {
		return "", fmt.Errorf("auto select compute mode: nil app config")
	}
	// V2 shared-app installs reduce to a "client-only install" whenever the
	// app's shared subchart is already owned by some other user. The wider
	// shape of v2 shared-app installs is:
	//
	//   - The shared subchart can only be wired up once cluster-wide and
	//     only by an admin user. Non-admin users trying to install it
	//     first are rejected upstream by CheckDependencies (the manifest
	//     declares the cluster-scoped admin install of itself as a
	//     mandatory dependency for non-admin renders).
	//   - Subsequent installers — whether non-admin or a second admin —
	//     all render the manifest's "client" branch (no requiredGpu,
	//     clusterScoped=false, shared subchart skipped at helm time) and
	//     do not allocate compute on their own. The original admin's
	//     server install owns the GPU.
	//
	// resolveComputeTarget returning !manage is the precise predicate for
	// the second case regardless of the caller's admin status: it just
	// looks at whether the shared namespace is already owned by someone
	// other than appConfig.OwnerName.
	//
	// Force cpu here so SelectedGpuType stays consistent with the
	// {cpu, installable} placeholder BuildInstallComputePlan emits for
	// !manage and AppInstallable matches. Picking the manifest's
	// nvidia/etc. just because the cluster has those GPUs would be
	// misleading: this install never schedules onto a GPU node.
	//
	// Note: this short-circuit only affects the install path. Resume goes
	// through ApplyBindingSelection (handler_suspend.go::resume) which
	// uses ShouldIncludeSharedServerForResume to decide whether the
	// resuming admin should also re-spin the shared server, and routes
	// allocations onto the original server owner. AutoSelectMode is not
	// on the resume path.
	_, manage, err := resolveComputeTarget(ctx, c, appCfg, false)
	if err != nil {
		return "", fmt.Errorf("auto select compute mode: resolve compute target: %w", err)
	}
	if !manage {
		return utils.CPUType, nil
	}
	var nodes corev1.NodeList
	if err := c.List(ctx, &nodes); err != nil {
		return "", fmt.Errorf("auto select compute mode: list nodes: %w", err)
	}
	// Reuse utils.GetAllGpuTypesFromNodes — same source of truth used by
	// /api/v1/gpus and the install-time chart-render auto-detect path. It
	// already excludes cpu (cpu nodes don't carry the gpu-type label), so
	// the returned set is exactly the cluster's non-cpu GPU types.
	clusterGPUTypes, err := utils.GetAllGpuTypesFromNodes(&nodes)
	if err != nil {
		return "", fmt.Errorf("auto select compute mode: %w", err)
	}
	return autoSelectModeFromInputs(AppSupportedModes(appCfg), clusterGPUTypes)
}

// autoSelectModeFromInputs is the pure-data core of AutoSelectMode, factored
// out so unit tests can exercise the decision matrix without spinning up a
// fake controller-runtime client. See AutoSelectMode for the contract.
func autoSelectModeFromInputs(appModes []string, clusterGPUTypes map[string]struct{}) (string, error) {
	var validNonCPU []string
	var validCPU bool
	for _, mode := range appModes {
		if mode == utils.CPUType {
			validCPU = true
			continue
		}
		if _, ok := clusterGPUTypes[mode]; ok {
			validNonCPU = append(validNonCPU, mode)
		}
	}

	switch len(validNonCPU) {
	case 1:
		return validNonCPU[0], nil
	case 0:
		if validCPU {
			return utils.CPUType, nil
		}
		return "", fmt.Errorf("no compute mode runnable on this cluster; please install on a compatible cluster")
	default:
		return "", ErrAmbiguousComputeMode
	}
}

// AppSupportedModes returns the set of GPU modes the app declares support
// for, by deferring to appCfg.ComputeResourceModes() — the same accessor
// the rest of the compute package (BuildInstallComputePlan,
// SelectedResourceMode, …) uses — so the auto-selector and the rest of
// the install pipeline always see the same view of the manifest:
//
//   - New-format manifests (>= 0.12.0): ComputeResourceModes returns the
//     spec.resources matrix verbatim, one entry per declared mode.
//   - Legacy manifests (< 0.12.0) with non-zero requiredGpu: it synthesizes
//     a single {Mode: nvidia} entry. Legacy apps in practice only target
//     nvidia, so a single-element {nvidia} set is the right answer.
//   - Legacy manifests with no GPU requirement: it synthesizes a single
//     {Mode: cpu} entry.
//
// Modes are returned in declared order. They are assumed to already be
// canonical strings (lowercased, no whitespace) since the manifest pipeline
// and node-label readers are responsible for keeping that contract.
func AppSupportedModes(appCfg *appcfg.ApplicationConfig) []string {
	if appCfg == nil {
		return nil
	}
	modes := appCfg.ComputeResourceModes()
	out := make([]string, 0, len(modes))
	for _, m := range modes {
		out = append(out, m.Mode)
	}
	return out
}
