package router

import (
	"context"
	"strings"
)

// Router's own view of the caller, read from GET /console/api/users/me.
//
// There is no verb of its own for it. Router derives the record from the
// X-BFL-USER header the Olares edge injects, and the role it carries is the
// Olares role rather than a second one to keep track of: `olares-cli profile
// whoami` already answers who the CLI is, and `router status` reports what
// Router made of that alongside the two other things a failing verb needs.

// consoleUser mirrors Router's users row as it appears on the wire.
type consoleUser struct {
	ID            string   `json:"id"`
	BflName       string   `json:"bfl_name"`
	Role          string   `json:"role"`
	Status        string   `json:"status"`
	Source        string   `json:"source"`
	OlaresID      string   `json:"olares_id,omitempty"`
	AllowedModels []string `json:"allowed_models,omitempty"`
}

func (u *consoleUser) isAdmin() bool {
	return strings.EqualFold(u.Role, "admin") || strings.EqualFold(u.Role, "owner")
}

func fetchConsoleUser(ctx context.Context, pc *preparedClient) (*consoleUser, error) {
	var me consoleUser
	if err := pc.router.doJSON(ctx, "GET", epMe, nil, &me); err != nil {
		return nil, err
	}
	return &me, nil
}

// userLabels maps Router's user ids to the names a person recognises, for the
// columns that would otherwise print a UUID. Failing to read the list is not
// worth failing the caller over — the id is still printed — so this returns an
// empty map rather than an error.
func userLabels(ctx context.Context, pc *preparedClient) map[string]string {
	users, err := listUsers(ctx, pc)
	if err != nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(users))
	for i := range users {
		u := &users[i]
		if u.ID != "" && u.BflName != "" {
			out[u.ID] = u.BflName
		}
	}
	return out
}
