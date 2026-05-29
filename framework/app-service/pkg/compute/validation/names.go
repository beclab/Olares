package validation

// Validator name constants. Call sites compare Decision.Validator against
// these rather than spelling the strings inline.
const (
	NameClusterCapacity   = "cluster-capacity"
	NameClusterPressure   = "cluster-pressure"
	NameUserQuota         = "user-quota"
	NameK8sRequest        = "k8s-request"
	NameComputeMode       = "compute-mode"
	NameNodePressure      = "node-pressure"
	NameComputeAllocation = "compute-allocation"
)
