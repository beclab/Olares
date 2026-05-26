package authz

// Stable ext_authz denial codes for cluster-internal Shared access.
const (
	CodeMissingCallerIdentity  = "MISSING_CALLER_IDENTITY"
	CodeInvalidCallerIdentity  = "INVALID_CALLER_IDENTITY"
	CodeInvalidHostUser        = "INVALID_HOST_USER"
	CodeNotAuthorizedCaller    = "NOT_AUTHORIZED_CALLER"
	CodeDelegationNotAllowed   = "DELEGATION_NOT_ALLOWED"
	CodeMeshNotReady           = "MESH_NOT_READY"
	CodeExtAuthzUpstreamFail   = "EXT_AUTHZ_UPSTREAM_FAIL"
)
