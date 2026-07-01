package validation

import (
	"context"
	"fmt"
	"github.com/beclab/Olares/framework/app-service/pkg/compute"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
)

// clusterMetricsProvider is the indirection clusterCapacityValidator
// uses to fetch the kubesphere-reported cluster totals. It exists as a
// package-level variable so unit tests can swap in a deterministic
// stub without needing a live kubesphere instance or an HTTP-level
// mock; production callers get the real GetClusterResource by default.
//
// Mirroring apputils.GetClusterResource's signature exactly keeps the
// indirection lossless — same return shape, same error semantics.
var clusterMetricsProvider = apputils.GetClusterResource

// clusterCapacityValidator answers the most fundamental feasibility
// question: "is the cluster physically big enough to host this app at
// all?" It reads the kubesphere-aggregated cluster totals via
// GetClusterResource and compares Total (NOT Total-Usage) against the
// app's declared AddedResources. Current pod consumption is
// intentionally ignored — that is the job of
// k8sRequestValidator / clusterPressureValidator.
//
// Placed first in DefaultValidators because:
//
//   - It produces the most actionable "your cluster is just too small"
//     error, which is otherwise hidden behind a confusing pressure
//     message from a downstream validator.
//   - Failing here lets the chain short-circuit before we spend time
//     on heavier validators (user quota, compute-mode planning,
//     per-node pressure walks).
//
// Units (matching GetClusterResource / ClusterMetrics):
//
//   - CPU.Total    : whole cores (float64).      Compared against
//     AddedResources.CPU in milli-cores after a *1000
//     conversion.
//   - Memory.Total : bytes (float64).            Compared as int64.
//   - Disk.Total   : bytes (float64).            Compared as int64.
//
// Token: the kubesphere monitoring endpoint authenticates via a
// service account token in production; the validator forwards
// Input.Token. An empty token is intentionally allowed (the webhook
// caller and a few system paths run without one) — GetClusterResource
// will surface the resulting auth error and we propagate it as a
// validator error rather than silently passing.
type clusterCapacityValidator struct{}

func (clusterCapacityValidator) Name() string { return NameClusterCapacity }

// AppliesTo install and upgrade. Resume reuses the placement chosen
// at install — the cluster's total schedulable capacity hasn't shrunk
// between install and resume in any normal flow, and in pathological
// "cluster shrank while the app was stopped" cases the runtime gate
// (k8s-request / node-pressure) will catch the failure with a more
// actionable message.
//
// Upgrade is included so we reject an upgrade whose new chart declares
// resource requirements the cluster can never satisfy (e.g. the new
// version raised CPU/memory/disk past the cluster's total schedulable
// capacity) at HTTP submit time, before any helm work happens. None
// of the other validators in this file apply to UpgradeOp — runtime
// pressure / per-user quota are handled separately on the upgrade path.
func (clusterCapacityValidator) AppliesTo(op Op) bool {
	switch op {
	case v1alpha1.InstallOp, v1alpha1.UpgradeOp:
		return true
	}
	return false
}

func (clusterCapacityValidator) Validate(ctx context.Context, in Input) (Decision, error) {
	added := compute.AddedResourcesFromAppConfig(in.AppConfig)

	// Short-circuit: when the app declares no resource requirement
	// there is nothing to compare against, so we avoid an unnecessary
	// kubesphere round trip.
	if added.CPU <= 0 && added.Memory <= 0 && added.Disk <= 0 {
		return ok(), nil
	}

	metrics, _, err := clusterMetricsProvider(in.Token)
	if err != nil {
		return Decision{}, fmt.Errorf("fetch cluster metrics for cluster-capacity check: %w", err)
	}
	if metrics == nil {
		// Defensive: never assume the provider returns a metrics
		// struct on success. Surface the inconsistency loudly so it
		// can be debugged rather than producing a misleading "your
		// cluster has 0 cpu" failure further down.
		return Decision{}, fmt.Errorf("cluster metrics provider returned nil result with no error")
	}

	totalCPUMilli := int64(metrics.CPU.Total * 1000) // cores → milli
	totalMemBytes := int64(metrics.Memory.Total)
	totalDiskBytes := int64(metrics.Disk.Total)

	op := string(in.Op)
	if added.CPU > totalCPUMilli {
		return Decision{
			OK:       false,
			Resource: constants.CPU,
			Reason:   constants.ClusterCPUInsufficient,
			Message:  fmt.Sprintf(constants.ClusterCPUInsufficientMessage, op),
		}, nil
	}
	if added.Memory > totalMemBytes {
		return Decision{
			OK:       false,
			Resource: constants.Memory,
			Reason:   constants.ClusterMemoryInsufficient,
			Message:  fmt.Sprintf(constants.ClusterMemoryInsufficientMessage, op),
		}, nil
	}
	if added.Disk > totalDiskBytes {
		return Decision{
			OK:       false,
			Resource: constants.Disk,
			Reason:   constants.ClusterDiskInsufficient,
			Message:  fmt.Sprintf(constants.ClusterDiskInsufficientMessage, op),
		}, nil
	}
	return ok(), nil
}

// clusterPressureValidator wraps apputils.CheckAppRequirement which
// checks aggregate cluster headroom (disk/memory/CPU) against the app's
// declared requirements. Requires a kubesphere token because the
// underlying call hits the kubesphere monitoring API.
type clusterPressureValidator struct{}

func (clusterPressureValidator) Name() string { return NameClusterPressure }

// Applies to install and resume only. UpgradeOp is excluded (upgrade
// does not run validation).
func (clusterPressureValidator) AppliesTo(op Op) bool {
	switch op {
	case v1alpha1.InstallOp, v1alpha1.ResumeOp:
		return true
	}
	return false
}

func (clusterPressureValidator) Validate(ctx context.Context, in Input) (Decision, error) {
	resource, reason, err := apputils.CheckAppRequirement(in.Token, in.AppConfig, in.Op)
	if err != nil {
		// CheckAppRequirement returns an empty resource/reason only when
		// the check itself couldn't be evaluated (e.g. the kubesphere
		// monitoring call failed); a genuine "resource insufficient"
		// rejection always carries a populated resource + reason. Surface
		// the former as a fatal chain error so callers treat it as
		// "unknown" instead of telling the user their cluster is out of
		// resources.
		if resource == "" && reason == "" {
			return Decision{}, fmt.Errorf("cluster-pressure check: %w", err)
		}
		return Decision{
			OK:       false,
			Resource: resource,
			Reason:   reason,
			Message:  err.Error(),
		}, nil
	}
	return ok(), nil
}

// userQuotaValidator wraps apputils.CheckUserResRequirement which
// checks the owner's per-user memory / CPU quota via prometheus.
type userQuotaValidator struct{}

func (userQuotaValidator) Name() string { return NameUserQuota }

// Applies to install and resume only. UpgradeOp is excluded (upgrade
// does not run validation).
func (userQuotaValidator) AppliesTo(op Op) bool {
	switch op {
	case v1alpha1.InstallOp, v1alpha1.ResumeOp:
		return true
	}
	return false
}

func (userQuotaValidator) Validate(ctx context.Context, in Input) (Decision, error) {
	resource, reason, err := apputils.CheckUserResRequirement(ctx, in.AppConfig, in.Op)
	if err != nil {
		// Empty resource/reason means the prometheus user-metrics call
		// failed, not that the user is over quota. Treat it as a fatal
		// chain error rather than a spurious quota rejection.
		if resource == "" && reason == "" {
			return Decision{}, fmt.Errorf("user-quota check: %w", err)
		}
		return Decision{
			OK:       false,
			Resource: resource,
			Reason:   reason,
			Message:  err.Error(),
		}, nil
	}
	return ok(), nil
}

// k8sRequestValidator wraps apputils.CheckAppK8sRequestResource which
// sums (allocatable - scheduled pod requests) across nodes and ensures
// the cluster has room for the app's CPU/memory request. Runs as part
// of RuntimePressureValidators in installing_app after helm install and
// before Scale(-1).
type k8sRequestValidator struct{}

func (k8sRequestValidator) Name() string { return NameK8sRequest }

// Applies to install and resume only. UpgradeOp is excluded (upgrade
// does not run validation).
func (k8sRequestValidator) AppliesTo(op Op) bool {
	switch op {
	case v1alpha1.InstallOp, v1alpha1.ResumeOp:
		return true
	}
	return false
}

func (k8sRequestValidator) Validate(ctx context.Context, in Input) (Decision, error) {
	resource, reason, err := apputils.CheckAppK8sRequestResource(in.AppConfig, in.Op)
	if err != nil {
		// Empty resource/reason means the node/allocatable lookup failed
		// (or appConfig was nil), not that the cluster lacks room. Surface
		// it as a fatal chain error rather than a spurious capacity
		// rejection.
		if resource == "" && reason == "" {
			return Decision{}, fmt.Errorf("k8s-request check: %w", err)
		}
		return Decision{
			OK:       false,
			Resource: resource,
			Reason:   reason,
			Message:  err.Error(),
		}, nil
	}
	return ok(), nil
}

// computeModeValidator wraps compute.AppInstallable which validates
// that the selected GPU mode actually has a runnable placement on the
// cluster. Only meaningful at install time; resume reuses the
// already-bound allocation.
type computeModeValidator struct{}

func (computeModeValidator) Name() string { return NameComputeMode }

// AppliesTo install only. UpgradeOp is excluded (upgrade does not run
// validation). Resume reuses the allocation chosen at install.
func (computeModeValidator) AppliesTo(op Op) bool {
	switch op {
	case v1alpha1.InstallOp:
		return true
	}
	return false
}

func (computeModeValidator) Validate(ctx context.Context, in Input) (Decision, error) {
	enough, err := compute.AppInstallable(ctx, in.Client, in.AppConfig)
	if err != nil {
		return Decision{}, err
	}
	if !enough {
		return Decision{
			OK:       false,
			Resource: constants.Compute,
			Reason:   constants.ComputeModeUnavailable,
			Message:  "compute resource is not enough for selected mode " + in.AppConfig.SelectedGpuType,
		}, nil
	}
	return ok(), nil
}

// nodePressureValidator wraps a "would any node accept adding the app's
// resources" check on top of compute.WouldPressure. Used in
// installing_app after helm install (workloads at replicas=0) and
// before Scale(-1): can the cluster actually take the app when we
// scale up?
//
// The check is conservative: we walk every node and if any node would
// stay below the pressure threshold once the app's CPU/memory request
// is added, we return OK. compute.PickAllocations already does the
// per-mode picking for GPU apps; this validator handles the CPU/memory
// pressure component.
type nodePressureValidator struct{}

func (nodePressureValidator) Name() string { return NameNodePressure }

// AppliesTo install and resume only. UpgradeOp is excluded (upgrade
// does not run validation).
func (nodePressureValidator) AppliesTo(op Op) bool {
	switch op {
	case v1alpha1.InstallOp, v1alpha1.ResumeOp:
		return true
	}
	return false
}

func (nodePressureValidator) Validate(ctx context.Context, in Input) (Decision, error) {
	snap, err := compute.FetchPressureSnapshot(ctx)
	if err != nil {
		return Decision{}, err
	}
	nodes, err := compute.FetchNodeComputeAllocations(ctx, in.Client)
	if err != nil {
		return Decision{}, err
	}
	added := compute.AddedResourcesFromAppConfig(in.AppConfig)
	for _, n := range nodes {
		if !snap.WouldPressure(n, added) {
			return ok(), nil
		}
	}
	return Decision{
		OK:       false,
		Resource: constants.Node,
		Reason:   constants.NodePressure,
		Message:  "no node has headroom under pressure threshold for the app's request",
	}, nil
}

// computeAllocationProvider is the indirection
// computeAllocationValidator uses to invoke
// compute.AllocateForInstall. The package-level var keeps unit tests
// hermetic — AllocateForInstall has real side effects (it writes
// Allocation records into the cluster), so production gets the real
// implementation by default while tests can swap a deterministic stub.
var computeAllocationProvider = compute.AllocateForInstall

// computeAllocationValidator wraps compute.AllocateForInstall, the
// scheduler step that picks a concrete node + GPU mode placement and
// records the allocation for the app. Unlike the other validators in
// this file it is NOT a pure feasibility check — a successful run
// writes Allocation records — but conceptually it shares the same
// shape: "can the cluster accept this app right now? if not, what was
// the failure?" so we model it as a validator and let the chain
// executor short-circuit on the first non-OK decision.
//
// Placement intent inside InstallRuntimePressureValidators (runs after
// helm install, before Scale(-1) in installing_app):
//
//   - cluster-pressure / k8s-request / node-pressure run first because
//     they are cheap, read-only signals.
//   - compute-allocation runs last because it does the heaviest work
//     and has side effects; running it only after the cheaper checks
//     pass avoids writing allocation records for an app the cluster
//     already cannot accept on simpler grounds.
//
// AppliesTo is install only — resume reuses the allocation chosen at
// install (so re-allocating would be wasteful and could spuriously
// fail on a transiently-degraded node). UpgradeOp is excluded from
// every validator in this file (upgrade does not run validation).
type computeAllocationValidator struct{}

func (computeAllocationValidator) Name() string { return NameComputeAllocation }

func (computeAllocationValidator) AppliesTo(op Op) bool {
	switch op {
	case v1alpha1.InstallOp:
		return true
	}
	return false
}

func (computeAllocationValidator) Validate(ctx context.Context, in Input) (Decision, error) {
	if _, err := computeAllocationProvider(ctx, in.Client, in.AppConfig); err != nil {
		// Surface the raw scheduler error through Decision.Message so
		// the install-time call site can synthesize an error from it
		// (it preserves the legacy
		// "Insufficient compute resource for selected mode %s: %v"
		// log line, which depends on the unwrapped err text).
		return Decision{
			OK:       false,
			Resource: constants.Compute,
			Reason:   constants.ComputeAllocationFailed,
			Message:  err.Error(),
		}, nil
	}
	return ok(), nil
}

// InstallabilityValidators returns the structural feasibility chain.
// These answer the question "can this cluster ever host this app?" —
// they look at static / slowly-changing properties (total schedulable
// capacity, GPU mode availability, per-user quota) and ignore the
// momentary level of pod scheduling pressure.
//
// Used at HTTP submit time (install handler) to reject requests the
// cluster fundamentally cannot accommodate before any helm work starts.
// They are intentionally NOT re-run in installing_app — runtime pressure
// and allocation run once after helm install and before Scale(-1).
//
// Upgrade uses its own chain (UpgradabilityValidators) instead of this
// one — see that function for the rationale.
//
// Ordering matches the user-facing failure mode they reveal:
//
//  1. cluster-capacity     — "your cluster is too small, period."
//  2. compute-mode         — "no node has the requested GPU mode."
//  3. user-quota           — "your account is over its limit."
func InstallabilityValidators() []Validator {
	return []Validator{
		clusterCapacityValidator{},
		computeModeValidator{},
		userQuotaValidator{},
	}
}

// UpgradabilityValidators returns the structural feasibility chain
// applied at HTTP submit time by the upgrade handler. The upgrade flow
// only needs to ask "is the cluster big enough to host the NEW chart's
// declared requirements at all?", because:
//
//   - compute-mode: the existing app already has an allocation, the
//     upgrade reuses prevCfg.SelectedGpuType. Re-running the planner
//     would either no-op or spuriously fail on a transiently-degraded
//     node.
//   - user-quota: the running deployment is already counted against
//     the owner's quota, so re-checking on upgrade would double-count
//     for templates whose new version raises requests only marginally.
//   - runtime-pressure / k8s-request / node-pressure: the upgrade goes
//     through helm upgrade which schedules its own replacement pods;
//     kubelet/kube-scheduler are the authoritative gate there.
//
// So the chain is intentionally just clusterCapacityValidator. We keep
// it as a separate exported chain (rather than letting callers pass a
// single validator inline) so that:
//
//   - The "what validates on upgrade?" answer lives in one place that
//     matches the InstallabilityValidators / RuntimePressureValidators
//     pattern.
//   - The clusterCapacityValidator type stays unexported.
//   - Adding another upgrade-time gate later only touches this file
//     and TestChainShapes, not every upgrade call site.
func UpgradabilityValidators() []Validator {
	return []Validator{
		clusterCapacityValidator{},
	}
}

// RuntimePressureValidators returns the "is the cluster currently
// under pressure that would block this app's pods from starting?"
// chain. These look at right-now-state: kubesphere monitoring
// (usage vs total), allocatable minus already-scheduled pod requests,
// and per-node pressure walks.
//
// Used at HTTP submit time (resume handler) and inside installing_app
// after helm install — once workloads are rendered at replicas=0, we
// re-check pressure before Scale(-1). Upgrade does not use this package.
//
// Ordering: cheapest aggregate signal first (kubesphere monitoring),
// then the more expensive per-node walks.
func RuntimePressureValidators() []Validator {
	return []Validator{
		clusterPressureValidator{},
		k8sRequestValidator{},
		nodePressureValidator{},
	}
}

// InstallRuntimePressureValidators is the install-flow extension of
// RuntimePressureValidators: it appends the compute-allocation
// scheduler step after the read-only pressure checks.
//
// Order is intentional: the cheap, read-only pressure validators run
// first, then compute-allocation picks a node and writes the
// Allocation record. If any earlier validator short-circuits the chain,
// no Allocation is written.
//
// Used by pkg/appstate/installing_app.go: before helm install for legacy
// apps (no workloadReplicas), after helm install and before Scale(-1)
// for two-phase apps.
func InstallRuntimePressureValidators() []Validator {
	return append(RuntimePressureValidators(), computeAllocationValidator{})
}

func ResumePressureValidators() []Validator {
	return []Validator{
		clusterPressureValidator{},
		k8sRequestValidator{},
		userQuotaValidator{},
	}
}

// DefaultValidators returns InstallabilityValidators ++
// RuntimePressureValidators in that order. Used by Run when callers
// pass no explicit chain — structural checks short-circuit first so we
// don't pay for kubesphere round trips on apps the cluster can never
// host.
func DefaultValidators() []Validator {
	out := InstallabilityValidators()
	out = append(out, RuntimePressureValidators()...)
	return out
}
