package ports

import "context"

// Identity centralises the "who is this user" lookups scattered
// across the legacy state machine. It backs the reads that let
// handlers branch on user privilege (admin vs regular) and on
// cluster-owner scope (shared apps installed by the cluster owner
// are visible to all users).
//
// Legacy call sites:
//   - upgrading_app.go:186-191       kubesphere.GetAdminUsername + deps.IsAdmin
//   - installing_canceling_app.go:108-112  deps.IsAdmin
//   - upgrading_app.go:214-225       apputils.GetClusterOwner (indirectly)
//
// All three are HANDLER-SAFE READS. The Runner materializes them
// into domain.World.Identity (field TBD when writing the first
// upgrade handler) so handlers read from World rather than calling
// this port directly.
type Identity interface {
	// IsAdmin returns true when the given user has admin privilege
	// (currently defined by kubesphere role bindings).
	IsAdmin(ctx context.Context, username string) (bool, error)

	// AdminUsername returns the tenant admin's username. Used by
	// upgrade to pick the identity that owns market chart fetches.
	AdminUsername(ctx context.Context) (string, error)

	// ClusterOwner returns the cluster-owner user (the account that
	// installed the cluster). Used by GetAppConfig to resolve shared
	// apps that live under the owner's namespace even when a
	// non-owner triggers the upgrade.
	ClusterOwner(ctx context.Context) (string, error)
}
