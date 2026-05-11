package dashboard

import (
	"context"
	"fmt"
)

// ResolveTargetUser returns the user to operate on for `--user`-aware
// commands (overview user, etc.).
//
//   - explicit (positional or --user) is honoured if the active profile is
//     admin; non-admins targeting a third party get an admin-required
//     error.
//   - empty falls back to the active profile's identity.
func ResolveTargetUser(ctx context.Context, c *Client, requested string) (*UserDetail, error) {
	u, err := c.EnsureUser(ctx)
	if err != nil {
		return nil, err
	}
	if requested == "" || requested == u.Name {
		return u, nil
	}
	if !u.IsAdmin() {
		return nil, fmt.Errorf("--user %q requires platform-admin; %s does not have that role", requested, u.Name)
	}
	// Admin path — return a synthetic UserDetail so callers can use
	// .Name without re-fetching IAM; downstream helpers only use Name.
	return &UserDetail{Name: requested, GlobalRole: "<admin-target>"}, nil
}
