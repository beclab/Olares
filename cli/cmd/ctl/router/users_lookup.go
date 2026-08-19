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
// `olares-cli settings users list`, and what is left here is the lookup the
// verbs that take a person need: `key issue --for-user`, `quota set --user`,
// `usage --user` and `audit --user` are given a name and have to send an id.

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
