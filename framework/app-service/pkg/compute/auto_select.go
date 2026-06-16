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
// mode declared by the app is runnable on the cluster. cpu is treated as a
// first-class candidate, so an app that declares both cpu and a matching gpu
// mode is ambiguous too; the caller must surface the choice to the user (e.g.
// by returning the install compute plan and asking for an explicit
// selectedGpuType).
var ErrAmbiguousComputeMode = errors.New("multiple compute modes runnable on this cluster; please specify selectedGpuType")

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
//     - cpu is a first-class candidate: it is "runnable" iff app.modes
//     contains cpu (every cluster can run cpu pods).
//     - any non-cpu mode m is "runnable" iff m ∈ cluster.gpuTypes.
//
//  4. Decide purely on len(valid):
//     - exactly one entry → return it (the only choice; may be cpu).
//     - zero entries → error (the app declares only gpu modes, none of
//     which the cluster has).
//     - more than one entry → ErrAmbiguousComputeMode; the caller surfaces
//     the choice. cpu being one of the candidates does NOT auto-resolve it:
//     {cpu, nvidia} is ambiguous and asks the user to pick.
//
// Worked examples (cluster has nvidia, cpu always runnable):
//
//	app supports {nvidia, gb10, amd}            -> nvidia (only nvidia matches)
//	app supports {nvidia, gb10, cpu}            -> error (ambiguous: nvidia + cpu)
//	app supports {amd, apple-m}                 -> error (no overlap)
//	app supports {cpu}                          -> cpu
//
// Worked examples (cluster has nvidia + amd, cpu always runnable):
//
//	app supports {nvidia}                       -> nvidia
//	app supports {amd}                          -> amd
//	app supports {nvidia, amd}                  -> error (ambiguous)
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
	seen := make(map[string]struct{})
	var valid []string
	add := func(mode string) {
		if _, ok := seen[mode]; ok {
			return
		}
		seen[mode] = struct{}{}
		valid = append(valid, mode)
	}
	for _, mode := range appModes {
		if mode == utils.CPUType {
			// cpu is a first-class candidate: every cluster can run cpu
			// pods, so it always counts when the app declares it.
			add(utils.CPUType)
			continue
		}
		if _, ok := clusterGPUTypes[mode]; ok {
			add(mode)
		}
	}

	switch len(valid) {
	case 1:
		return valid[0], nil
	case 0:
		return "", fmt.Errorf("The Olares cluster cannot meet the required resource specifications. No matching GPU type")
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
