package validation

import (
	"context"
	"errors"
	"fmt"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/compute"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"

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
var checkAppRequirement = apputils.CheckAppRequirement
var checkUserResRequirement = apputils.CheckUserResRequirement
var checkAppK8sRequestResource = apputils.CheckAppK8sRequestResource

// kubeClientset builds the client used to inspect a deployed workload
// directly. Kept as a variable for the same reason as the providers
// above: the checks that need it must be exercisable without a live
// cluster.
var kubeClientset = func() (kubernetes.Interface, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

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
// capacity) at HTTP submit time, before any helm work happens. This
// validator uses the new chart's ABSOLUTE requirements on upgrade;
// cluster-pressure / k8s-request handle the headroom side against the
// delta (new − old) — see UpgradabilityValidators.
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

	op := string(in.Op)
	pressure, err := apputils.EvaluatePhysicalCapacity(apputils.ResourceState{
		CPU: added.CPU, Memory: added.Memory, Disk: added.Disk,
	}, metrics, apputils.ResourceDimensions{
		CPU: added.CPU > 0, Memory: added.Memory > 0, Disk: added.Disk > 0,
	})
	if err != nil {
		if resourceType, ok := apputils.MetricsFailureResource(err); ok {
			return Decision{
				OK:       false,
				Resource: resourceType,
				Reason:   constants.MetricsUnavailable,
				Message:  fmt.Sprintf(constants.MetricsUnavailableMessage, op),
			}, nil
		}
		return Decision{}, fmt.Errorf("evaluate cluster capacity: %w", err)
	}
	if len(pressure) == 0 {
		return ok(), nil
	}
	switch pressure[0].Resource {
	case string(constants.CPU):
		return Decision{
			OK:       false,
			Resource: constants.CPU,
			Reason:   constants.ClusterCPUInsufficient,
			Message:  fmt.Sprintf(constants.ClusterCPUInsufficientMessage, op),
		}, nil
	case string(constants.Memory):
		return Decision{
			OK:       false,
			Resource: constants.Memory,
			Reason:   constants.ClusterMemoryInsufficient,
			Message:  fmt.Sprintf(constants.ClusterMemoryInsufficientMessage, op),
		}, nil
	default:
		return Decision{
			OK:       false,
			Resource: constants.Disk,
			Reason:   constants.ClusterDiskInsufficient,
			Message:  fmt.Sprintf(constants.ClusterDiskInsufficientMessage, op),
		}, nil
	}
}

// clusterPressureValidator wraps apputils.CheckAppRequirement which
// checks aggregate cluster headroom (disk/memory/CPU) against the app's
// declared requirements. Requires a kubesphere token because the
// underlying call hits the kubesphere monitoring API.
//
// On UpgradeOp the check uses the non-negative resource delta
// (new − old) so the running deployment already counted in cluster
// usage is not double-counted — but only while the deployed version
// still has pods on the cluster. Upgrading a workload that is down falls
// back to the new chart's absolute requirements.
type clusterPressureValidator struct{}

func (clusterPressureValidator) Name() string { return NameClusterPressure }

func (clusterPressureValidator) AppliesTo(op Op) bool {
	switch op {
	case v1alpha1.InstallOp, v1alpha1.ResumeOp, v1alpha1.UpgradeOp:
		return true
	}
	return false
}

func (clusterPressureValidator) Validate(ctx context.Context, in Input) (Decision, error) {
	cfg, skip, err := requirementConfigForOp(ctx, in)
	if err != nil {
		return Decision{}, fmt.Errorf("cluster-pressure check: %w", err)
	}
	if skip {
		return ok(), nil
	}
	resource, reason, err := checkAppRequirement(in.Token, cfg, in.Op)
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

func (userQuotaValidator) AppliesTo(op Op) bool {
	switch op {
	case v1alpha1.InstallOp, v1alpha1.ResumeOp:
		return true
	}
	return false
}

func (userQuotaValidator) Validate(ctx context.Context, in Input) (Decision, error) {
	cfg, skip, err := requirementConfigForOp(ctx, in)
	if err != nil {
		return Decision{}, fmt.Errorf("user-quota check: %w", err)
	}
	if skip {
		return ok(), nil
	}
	resource, reason, err := checkUserResRequirement(ctx, cfg, in.Op)
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
//
// On UpgradeOp the check uses the non-negative resource delta
// (new − old) so already-scheduled requests of the running deployment
// are not double-counted, falling back to the new chart's absolute
// requirements when the deployed version has no pods left holding them.
// Rolling-update surge is left to kube-scheduler.
type k8sRequestValidator struct{}

func (k8sRequestValidator) Name() string { return NameK8sRequest }

func (k8sRequestValidator) AppliesTo(op Op) bool {
	switch op {
	case v1alpha1.InstallOp, v1alpha1.ResumeOp, v1alpha1.UpgradeOp:
		return true
	}
	return false
}

func (k8sRequestValidator) Validate(ctx context.Context, in Input) (Decision, error) {
	cfg, skip, err := requirementConfigForOp(ctx, in)
	if err != nil {
		return Decision{}, fmt.Errorf("k8s-request check: %w", err)
	}
	if skip {
		return ok(), nil
	}
	resource, reason, err := checkAppK8sRequestResource(cfg, in.Op)
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

// requirementConfigForOp returns the ApplicationConfig whose declared
// Requirement should be checked. For install/resume that is AppConfig
// as-is. For upgrade it is a shallow copy with Requirement set to
// max(0, new−old), but only when the previous version still holds its
// requests (see upgradePrevHoldsRequests); otherwise the new chart's
// absolute requirements are checked, because nothing is being freed.
// skip is true when the delta is zero on every dimension (caller should
// treat as OK without hitting metrics).
func requirementConfigForOp(ctx context.Context, in Input) (cfg *appcfg.ApplicationConfig, skip bool, err error) {
	if in.Op != v1alpha1.UpgradeOp {
		return in.AppConfig, false, nil
	}
	holds, err := upgradePrevHoldsRequests(ctx, in.PrevAppConfig, in.PrevState, in.PreUpgradeSteadyState)
	if err != nil {
		return nil, false, err
	}
	if !holds {
		return in.AppConfig, false, nil
	}
	delta := compute.UpgradeResourceDelta(in.PrevAppConfig, in.AppConfig)
	if delta.CPU <= 0 && delta.Memory <= 0 && delta.Disk <= 0 {
		return nil, true, nil
	}
	return compute.AppConfigWithRequirement(in.AppConfig, delta), false, nil
}

// upgradePrevHoldsRequests reports whether the deployed version an
// upgrade starts from still holds its requests, so only the (new − old)
// increase needs to fit in live headroom.
//
// Decision order:
//
//   - Running → delta (pods are expected to be up)
//   - Stopped → full (nothing to discount)
//   - UpgradeFailed + preUpgradeSteadyState running/stopped → same as above
//     (AppPreUpgradeStateKey snapshot from the attempt that failed)
//   - UpgradeFailed with missing/illegal preUpgradeSteadyState, and every
//     other state → live-pod probe in the previous config's namespace.
//     Terminated and unscheduled pods hold nothing, so Failed,
//     Succeeded and pods without a NodeName are excluded.
//
// Every uncertainty on the probe path is an error rather than a
// fallback: an unreadable namespace, an unknown previous config or one
// without a namespace all leave us unable to tell whether the discount
// is real, and assuming headroom that may not exist is the failure mode
// this guards against.
func upgradePrevHoldsRequests(ctx context.Context, cfg *appcfg.ApplicationConfig, state v1alpha1.ApplicationManagerState, preUpgradeSteadyState string) (bool, error) {
	switch state {
	case v1alpha1.Running:
		return true, nil
	case v1alpha1.Stopped:
		return false, nil
	case v1alpha1.UpgradeFailed:
		switch preUpgradeSteadyState {
		case string(v1alpha1.Running):
			return true, nil
		case string(v1alpha1.Stopped):
			return false, nil
		}
	}

	if cfg == nil {
		return false, errors.New("no previous app config to locate the deployed workload")
	}
	if cfg.Namespace == "" {
		return false, fmt.Errorf("app config %q has no namespace", cfg.AppName)
	}

	kube, err := kubeClientset()
	if err != nil {
		return false, fmt.Errorf("build kube client: %w", err)
	}
	if _, err = kube.CoreV1().Namespaces().Get(ctx, cfg.Namespace, metav1.GetOptions{}); err != nil {
		return false, fmt.Errorf("get namespace %s: %w", cfg.Namespace, err)
	}

	pods, err := kube.CoreV1().Pods(cfg.Namespace).List(ctx, metav1.ListOptions{
		FieldSelector: "status.phase!=Failed,status.phase!=Succeeded",
	})
	if err != nil {
		return false, fmt.Errorf("list pods in namespace %s: %w", cfg.Namespace, err)
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase != corev1.PodFailed &&
			pod.Status.Phase != corev1.PodSucceeded &&
			pod.Spec.NodeName != "" {
			return true, nil
		}
	}
	return false, nil
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

// UpgradabilityValidators returns the feasibility chain applied at HTTP
// submit time by the upgrade handler. It answers two questions:
//
//   - cluster-capacity: "is the cluster physically big enough to host
//     the NEW chart's declared requirements at all?" — evaluated
//     against the new chart's ABSOLUTE requirements (Total, not
//     Total−Usage), since a running old version does not shrink the
//     cluster's total schedulable size.
//   - cluster-pressure / k8s-request: "can live headroom absorb the
//     resource INCREASE this upgrade introduces?" — each runs against
//     the non-negative delta (new − old) computed from
//     Input.PrevAppConfig, so the running deployment already reflected
//     in cluster usage / scheduled requests is not double-counted.
//     When the delta is zero on every dimension (the common case where
//     a new version keeps or lowers its requests) these validators
//     short-circuit to OK without any metrics round trip. Upgrades of a
//     workload whose pods are gone have nothing to discount, so they are
//     checked against the new chart's absolute requirements instead —
//     see upgradePrevHoldsRequests.
//
// Intentionally excluded:
//
//   - user-quota: install/resume remain the per-user quota gates;
//     upgrade does not re-check owner quota.
//   - compute-mode: the existing app already has an allocation, the
//     upgrade reuses prevCfg.SelectedGpuType. Re-running the planner
//     would either no-op or spuriously fail on a transiently-degraded
//     node.
//   - node-pressure / compute-allocation: rolling-update surge and pod
//     placement go through helm upgrade + kube-scheduler, which are the
//     authoritative gate there.
//
// The delta-aware validators REQUIRE Input.PrevAppConfig to be set for
// UpgradeOp; the upgrade handler is the only caller of this chain and
// always supplies it.
func UpgradabilityValidators() []Validator {
	return []Validator{
		clusterCapacityValidator{},
		clusterPressureValidator{},
		k8sRequestValidator{},
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
