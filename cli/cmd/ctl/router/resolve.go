package router

// Reading a collection, and turning a typed word into one of its rows.
//
// Five verbs take a name where a route wants an id — a provider, a key, a user,
// a calling application, a model on a provider — and each did the same three
// things: read the whole collection, compare the word against every handle a
// row answers to, and explain itself when nothing matched. The lookups differ
// enough to stay separate (a key is also findable by the prefix that appears in
// a log, a provider by an id the list may not even contain), but the two ends
// of that shape are the same everywhere and now live here.
//
// The reading matters more than it looks. Six routes answer with `{"items":
// [...]}` and nothing else, and three of them were read twice inside one verb:
// `router usage list --user alice` looked users up to build the filter and then
// looked the same list up again to put a name on each row. A verb runs once per
// process, so a collection read twice is a round trip on somebody's home
// network that answered the same question twice.
//
// The explaining matters because of which sentence it picks. "There are none of
// these yet" and "yours is not among these" send a reader somewhere different,
// and a bare "not found" sends them nowhere.

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// entityID matches the UUID Router gives every row it owns, which is how a
// lookup decides whether the word it was handed can skip the list.
var entityID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// page is the envelope every list route with paging answers with. Total is the
// count before the window, which is what the footer under a table needs; Limit
// and Offset come back so a caller can report the window Router actually
// applied rather than the one it asked for.
type page[T any] struct {
	Items  []T `json:"items"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// collection reads a route that answers with items and nothing else, once per
// invocation.
//
// The memo is keyed by the full path, query included, so two different filters
// are two different reads. It is dropped as soon as anything is written through
// this client: a verb that creates a provider and then lists them has to see
// the one it just made, and holding a pre-write snapshot would show it the
// world as it was a moment ago — the single most confusing thing a cache can
// do. Comparing the write counter rather than clearing the map keeps that true
// without every write site having to remember.
//
// A loop that waits for something to change must not read through here. Nothing
// writes while it waits, so the memo would answer every turn of the loop with
// the snapshot the first one took, and the wait would never end. A poller reads
// its route directly.
func collection[T any](ctx context.Context, pc *preparedClient, path string) ([]T, error) {
	if memo, ok := pc.collections[path]; ok && memo.writes == pc.router.writes {
		if items, ok := memo.items.([]T); ok {
			return items, nil
		}
	}
	var env struct {
		Items []T `json:"items"`
	}
	if err := pc.router.doJSON(ctx, http.MethodGet, path, nil, &env); err != nil {
		return nil, err
	}
	if pc.collections == nil {
		pc.collections = map[string]memoizedCollection{}
	}
	pc.collections[path] = memoizedCollection{items: env.Items, writes: pc.router.writes}
	return env.Items, nil
}

type memoizedCollection struct {
	items  any
	writes int
}

// requireRef trims what was typed and refuses an empty argument, naming the
// handles that would have worked. what is spelled as it appears in the
// sentence: "a provider name or id".
func requireRef(ref, what string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", fmt.Errorf("%s is required", what)
	}
	return ref, nil
}

// missing is the refusal a lookup owes a word that named nothing.
//
// noun is spelled as it should read after "no", and may carry the scope it was
// looked for in: "model on openai" produces "no model on openai named ...".
type missing struct {
	noun  string   // singular, as it reads after "no"
	ref   string   // what the person typed
	known []string // what does exist, spelled as they would type it back
	have  string   // the clause introducing known, when there is any
	none  string   // said instead when nothing of this kind exists at all
	note  string   // what to do about it, said either way
}

// namesInAMiss caps the list. A deployment with two hundred keys turns one
// helpful sentence into a screenful, and past a dozen the reader is going to
// run the list verb regardless.
const namesInAMiss = 12

func (m missing) err() error {
	var b strings.Builder
	fmt.Fprintf(&b, "no %s named %q; ", m.noun, m.ref)
	if len(m.known) == 0 {
		b.WriteString(m.none)
	} else {
		shown := m.known
		extra := 0
		if len(shown) > namesInAMiss {
			extra = len(shown) - namesInAMiss
			shown = shown[:namesInAMiss]
		}
		b.WriteString(m.have + " " + strings.Join(shown, ", "))
		if extra > 0 {
			fmt.Fprintf(&b, " and %d more", extra)
		}
	}
	if note := strings.TrimSpace(m.note); note != "" {
		b.WriteString(". " + note)
	}
	return fmt.Errorf("%s", b.String())
}
