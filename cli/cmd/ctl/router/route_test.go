package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func member(name string, servable bool) routeMember {
	return routeMember{QualifiedName: name, ModelName: name, Servable: servable,
		ProviderModelID: "pm-" + name}
}

// A name that resolves and answers nothing is the failure this column exists
// for. It is not the same as a name that does not exist -- the caller gets a
// route, Router gets a request, and the refusal arrives from further in -- so
// enabled alone was never the question.
func TestARouteIsCallableOnlyWithSomethingBehindIt(t *testing.T) {
	cases := []struct {
		name  string
		route modelRoute
		want  bool
	}{
		{"enabled with a live member", modelRoute{Enabled: true,
			Members: []routeMember{member("a", true)}}, true},
		{"enabled with only dead members", modelRoute{Enabled: true,
			Members: []routeMember{member("a", false), member("b", false)}}, false},
		{"enabled with no members at all", modelRoute{Enabled: true}, false},
		{"disabled despite a live member", modelRoute{
			Members: []routeMember{member("a", true)}}, false},
		{"one live member among dead ones", modelRoute{Enabled: true,
			Members: []routeMember{member("a", false), member("b", true)}}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.route.callable(); got != tc.want {
				t.Errorf("callable() = %v, want %v", got, tc.want)
			}
		})
	}
}

// An empty category and an empty group mean different things, and the cell has
// to say which: nobody built the group, versus Router built the category and
// this machine has nothing that serves it. The second is fixed in Market.
func TestAnEmptyRouteSaysWhichKindOfEmptyItIs(t *testing.T) {
	group := modelRoute{Kind: routeKindGroup}
	if got := group.answersWith(); got != "nothing" {
		t.Errorf("empty group reads %q", got)
	}
	category := modelRoute{Kind: routeKindDefault}
	if got := category.answersWith(); !strings.Contains(got, "installed") {
		t.Errorf("empty category reads %q, which does not point at Market", got)
	}
}

func TestARouteWithSeveralBackendsNamesOneAndCounts(t *testing.T) {
	r := modelRoute{Members: []routeMember{member("a", true), member("b", true), member("c", true)}}
	if got := r.answersWith(); got != "a and 2 more" {
		t.Errorf("answersWith() = %q", got)
	}
	single := modelRoute{Members: []routeMember{member("a", true)}}
	if got := single.answersWith(); got != "a" {
		t.Errorf("a single backend reads %q, want no count", got)
	}
}

// Every local model application is a provider called `Olares`, so two members
// of one group render as the same string unless the application is named. An
// admin reordering them has to be able to tell which row is which.
func TestAMemberOfALocalApplicationIsNamedByIt(t *testing.T) {
	app, title := "qwen3-8b-a", "Qwen3 8B (A)"
	cases := []struct {
		name string
		m    routeMember
		want string
	}{
		{"application name wins", routeMember{QualifiedName: "Olares/qwen3-8b",
			OlaresAppName: &app, ProviderTitle: &title}, "Olares/qwen3-8b (qwen3-8b-a)"},
		{"title when there is no application name", routeMember{QualifiedName: "Olares/qwen3-8b",
			ProviderTitle: &title}, "Olares/qwen3-8b (Qwen3 8B (A))"},
		{"nothing to add", routeMember{QualifiedName: "openai/gpt-4o"}, "openai/gpt-4o"},
		{"qualified name assembled when absent", routeMember{ProviderName: "openai",
			ModelName: "gpt-4o"}, "openai/gpt-4o"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.m.label(); got != tc.want {
				t.Errorf("label() = %q, want %q", got, tc.want)
			}
		})
	}
}

// A title identical to the provider name adds nothing and would read as a
// stutter -- "openai/gpt-4o (openai)".
func TestALabelDoesNotRepeatTheProviderName(t *testing.T) {
	same := "openai"
	m := routeMember{QualifiedName: "openai/gpt-4o", ProviderName: "openai", ProviderTitle: &same}
	if got := m.label(); got != "openai/gpt-4o" {
		t.Errorf("label() = %q, want the name unadorned", got)
	}
}

func TestViaNamesTheHopOrNothing(t *testing.T) {
	cases := []struct {
		name  string
		route modelRoute
		want  string
	}{
		{"direct at a model", modelRoute{}, ""},
		{"through a group", modelRoute{Target: &routeTarget{
			RouteKind: "group", RouteName: "fast"}}, "group fast"},
		{"at a model, target carries no route", modelRoute{Target: &routeTarget{
			ProviderModelID: "pm-1"}}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.route.via(); got != tc.want {
				t.Errorf("via() = %q, want %q", got, tc.want)
			}
		})
	}
}

// A dangling target is worth saying out loud rather than rendering as blank:
// blank reads as "points straight at a model", which is a different and
// working configuration.
func TestADanglingTargetIsNotRenderedAsDirect(t *testing.T) {
	r := modelRoute{Target: &routeTarget{RouteID: "rt-gone"}}
	got := r.via()
	if !strings.Contains(got, "no longer exists") || !strings.Contains(got, "rt-gone") {
		t.Errorf("via() = %q, want it to name the missing route", got)
	}
}

// The two rules with a reason behind them are the ones worth failing early:
// a slash makes a route name indistinguishable from a qualified reference, and
// the default- prefix impersonates a category Router maintains.
func TestARouteNameIsCheckedBeforeTheRoundTrip(t *testing.T) {
	cases := []struct {
		name, in, wantIn string
	}{
		{"empty", "", "required"},
		{"too long", strings.Repeat("a", 65), "at most 64"},
		{"impersonates a category", "default-chat", "impersonate"},
		{"contains a slash", "openai/gpt-4o", "'/'"},
		{"uppercase", "Fast", "lowercase"},
		{"punctuation", "fast!", "may contain"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkRouteName(tc.in)
			if err == nil {
				t.Fatalf("checkRouteName(%q) accepted it", tc.in)
			}
			if !strings.Contains(err.Error(), tc.wantIn) {
				t.Errorf("checkRouteName(%q) said %q, want it to mention %q", tc.in, err, tc.wantIn)
			}
		})
	}
	for _, ok := range []string{"fast", "fast-lane", "fast_lane", "gpt4", "a"} {
		if err := checkRouteName(ok); err != nil {
			t.Errorf("checkRouteName(%q) refused a valid name: %v", ok, err)
		}
	}
}

// An alias and a category refuse members for different reasons, and the reason
// is the whole value of the message: one is "this is not that kind of route",
// the other is "these backends are not yours".
func TestOnlyAGroupTakesMembers(t *testing.T) {
	if err := refuseNonGroupMembership(&modelRoute{Kind: routeKindGroup}, "add to"); err != nil {
		t.Errorf("a group refused a member: %v", err)
	}
	alias := &modelRoute{Kind: routeKindAlias, Name: "fast", Mode: "chat",
		Members: []routeMember{member("openai/gpt-4o", true)}}
	err := refuseNonGroupMembership(alias, "add to")
	if err == nil {
		t.Fatal("an alias accepted a member")
	}
	if !strings.Contains(err.Error(), "openai/gpt-4o") || !strings.Contains(err.Error(), "--kind group") {
		t.Errorf("the alias refusal does not say what it names or what to do: %v", err)
	}
	category := &modelRoute{Kind: routeKindDefault, Name: "default-chat"}
	err = refuseNonGroupMembership(category, "remove from")
	if err == nil {
		t.Fatal("a category accepted a membership change")
	}
	if !strings.Contains(err.Error(), "Router points itself") {
		t.Errorf("the category refusal does not say who owns it: %v", err)
	}
}

func TestAMemberIsFoundByAnyNameItHas(t *testing.T) {
	r := &modelRoute{Members: []routeMember{{
		ProviderModelID: "pm-1", QualifiedName: "openai/gpt-4o", ModelName: "gpt-4o",
	}}}
	for _, ref := range []string{"pm-1", "openai/gpt-4o", "gpt-4o", "OpenAI/GPT-4o"} {
		if got := memberIDInRoute(r, ref); got != "pm-1" {
			t.Errorf("memberIDInRoute(%q) = %q, want pm-1", ref, got)
		}
	}
	if got := memberIDInRoute(r, "claude"); got != "" {
		t.Errorf("memberIDInRoute found %q for a model that is not a member", got)
	}
}

func routeServer(t *testing.T, routes []modelRoute) *preparedClient {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"items": routes})
	}))
	t.Cleanup(ts.Close)
	return &preparedClient{router: newRouterClient(ts.Client(), ts.URL, "alice@example.com")}
}

// People think in "chat" and clients send "default-chat". Only Router mints
// these names, so expanding the short form cannot collide with an alias or a
// group somebody made.
func TestACategoryIsFoundByTheHalfThatVaries(t *testing.T) {
	pc := routeServer(t, []modelRoute{
		{ID: "rt-1", Name: "default-chat", Kind: routeKindDefault, Enabled: true},
		{ID: "rt-2", Name: "fast", Kind: routeKindGroup, Enabled: true},
	})
	ctx := context.Background()

	for _, ref := range []string{"chat", "default-chat", "CHAT"} {
		got, err := resolveRoute(ctx, pc, ref)
		if err != nil {
			t.Fatalf("resolveRoute(%q): %v", ref, err)
		}
		if got.Name != "default-chat" {
			t.Errorf("resolveRoute(%q) found %q", ref, got.Name)
		}
	}
}

// The expansion must not reach past an exact match. A group called `chat` is a
// name somebody chose, and silently answering with `default-chat` would send
// their traffic somewhere they did not configure.
func TestAnExactNameBeatsTheCategoryExpansion(t *testing.T) {
	pc := routeServer(t, []modelRoute{
		{ID: "rt-1", Name: "default-chat", Kind: routeKindDefault, Enabled: true},
		{ID: "rt-2", Name: "chat", Kind: routeKindGroup, Enabled: true},
	})
	got, err := resolveRoute(context.Background(), pc, "chat")
	if err != nil {
		t.Fatalf("resolveRoute: %v", err)
	}
	if got.Kind != routeKindGroup {
		t.Errorf("resolveRoute(\"chat\") answered with the %s, not the group of that name", got.Kind)
	}
}

// A name that does not exist is answered with the names that do, and with the
// fact that a model needs no route at all -- which is the likelier thing the
// reader was reaching for.
func TestAnUnknownRouteIsAnsweredWithTheOnesThatExist(t *testing.T) {
	pc := routeServer(t, []modelRoute{{ID: "rt-1", Name: "fast", Kind: routeKindGroup}})
	_, err := resolveRoute(context.Background(), pc, "slow")
	if err == nil {
		t.Fatal("an unknown route resolved")
	}
	if !strings.Contains(err.Error(), "fast") {
		t.Errorf("the refusal does not list the routes that exist: %v", err)
	}
	if !strings.Contains(err.Error(), "<provider>/<model>") {
		t.Errorf("the refusal does not mention calling without a route: %v", err)
	}
}
