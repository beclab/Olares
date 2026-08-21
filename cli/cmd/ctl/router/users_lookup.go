package router

import (
	"context"
	"fmt"
	"strings"
)

// Router's user records — GET /console/api/users.
//
// There is no verb for this. Router's users are not Olares' users: a row appears
// the first time somebody's session reaches Router, created from the X-BFL-USER
// header the edge injects, and the role on it is the Olares role read back
// rather than a second one to keep track of. So the people on this Olares are
// `olares-cli settings users list`, their role is `olares-cli profile whoami`,
// and what is left here is the lookup the verbs that take a person need: `key
// issue --for-user`, `quota set --user`, `usage --user` and `audit --user` are
// given a name and have to send an id.

// consoleUser is the part of a user row those lookups read. The row carries
// more — a role, a source, a status that the directory's delete cascade
// tombstones — and none of it is read here, because nothing left in this tree
// prints a person.
type consoleUser struct {
	ID       string `json:"id"`
	BflName  string `json:"bfl_name"`
	OlaresID string `json:"olares_id,omitempty"`
}

func listUsers(ctx context.Context, pc *preparedClient) ([]consoleUser, error) {
	return collection[consoleUser](ctx, pc, epUsers)
}

// resolveUserID turns a name into the id the routes take. Names are what people
// know; the id appears only in Router's own rows.
func resolveUserID(ctx context.Context, pc *preparedClient, ref string) (string, error) {
	ref, err := requireRef(ref, "a user name or id")
	if err != nil {
		return "", err
	}
	if entityID.MatchString(ref) {
		return ref, nil
	}
	users, err := listUsers(ctx, pc)
	if err != nil {
		return "", fmt.Errorf("look up user %q: %w", ref, err)
	}
	known := make([]string, 0, len(users))
	for i := range users {
		if strings.EqualFold(users[i].BflName, ref) || strings.EqualFold(users[i].OlaresID, ref) {
			return users[i].ID, nil
		}
		known = append(known, users[i].BflName)
	}
	return "", missing{
		noun:  "user",
		ref:   ref,
		known: known,
		have:  "Router knows",
		none:  "Router has no user records yet, which happens before anybody has opened it",
		note:  "A user appears only after their first visit to Router",
	}.err()
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
