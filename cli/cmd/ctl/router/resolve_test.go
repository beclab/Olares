package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// requestLog is what the collection tests actually assert on: not that a
// function returned rows, which it did before the memo existed, but how many
// times Router was asked for them.
type requestLog struct {
	paths  []string
	users  []consoleUser
	models []adminModelRow
}

func (l *requestLog) serve(w http.ResponseWriter, r *http.Request) {
	l.paths = append(l.paths, r.Method+" "+r.URL.RequestURI())
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
		return
	}
	if r.URL.Path == epProviderModels {
		_ = json.NewEncoder(w).Encode(map[string]any{"items": l.models})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"items": l.users})
}

func (l *requestLog) gets(path string) int {
	n := 0
	for _, p := range l.paths {
		if p == "GET "+path {
			n++
		}
	}
	return n
}

func newTestClient(t *testing.T) (*preparedClient, *requestLog) {
	t.Helper()
	log := &requestLog{users: []consoleUser{{ID: "u-1", BflName: "alice"}}}
	srv := httptest.NewServer(http.HandlerFunc(log.serve))
	t.Cleanup(srv.Close)
	return &preparedClient{router: newRouterClient(srv.Client(), srv.URL, "alice@example.com")}, log
}

// `router usage list --user alice` reads the user list to build the filter and
// then reads it again to put a name on every row. That was two round trips over
// somebody's home network for one answer.
func TestACollectionIsReadOncePerInvocation(t *testing.T) {
	pc, log := newTestClient(t)
	ctx := context.Background()

	if _, err := listUsers(ctx, pc); err != nil {
		t.Fatalf("first read: %v", err)
	}
	if labels := userLabels(ctx, pc); labels["u-1"] != "alice" {
		t.Fatalf("labels did not come back: %v", labels)
	}
	if got := log.gets(epUsers); got != 1 {
		t.Errorf("asked Router for the user list %d times, want 1", got)
	}
}

// A filter is part of what was asked, so two filters are two questions.
func TestTwoFiltersAreTwoReads(t *testing.T) {
	pc, log := newTestClient(t)
	ctx := context.Background()

	if _, err := collection[consoleUser](ctx, pc, epUsers+"?status=active"); err != nil {
		t.Fatalf("first filter: %v", err)
	}
	if _, err := collection[consoleUser](ctx, pc, epUsers+"?status=disabled"); err != nil {
		t.Fatalf("second filter: %v", err)
	}
	if got := len(log.paths); got != 2 {
		t.Errorf("made %d requests, want one per filter", got)
	}
}

// The worst thing a cache here could do is show a verb the world as it was
// before its own write.
func TestAWriteInvalidatesWhatWasRead(t *testing.T) {
	pc, log := newTestClient(t)
	ctx := context.Background()

	if _, err := listUsers(ctx, pc); err != nil {
		t.Fatalf("first read: %v", err)
	}
	log.users = append(log.users, consoleUser{ID: "u-2", BflName: "bob"})
	if err := pc.router.doJSON(ctx, http.MethodPost, epUsers, map[string]any{"name": "bob"}, nil); err != nil {
		t.Fatalf("write: %v", err)
	}
	after, err := listUsers(ctx, pc)
	if err != nil {
		t.Fatalf("read after write: %v", err)
	}
	if len(after) != 2 {
		t.Errorf("read %d users after creating one, want 2", len(after))
	}
	if got := log.gets(epUsers); got != 2 {
		t.Errorf("read the list %d times, want a fresh one after the write", got)
	}
}

// A completion on the data plane changes nothing the console plane has read, so
// the clone that carries the `sk-*` bearer keeps its own count.
func TestADataPlaneCallDoesNotInvalidateTheConsole(t *testing.T) {
	pc, log := newTestClient(t)
	ctx := context.Background()

	if _, err := listUsers(ctx, pc); err != nil {
		t.Fatalf("first read: %v", err)
	}
	keyed := pc.router.withHeader("Authorization", "Bearer sk-test")
	if err := keyed.doJSON(ctx, http.MethodPost, epChatCompletions, map[string]any{"model": "m"}, nil); err != nil {
		t.Fatalf("completion: %v", err)
	}
	if _, err := listUsers(ctx, pc); err != nil {
		t.Fatalf("second read: %v", err)
	}
	if got := log.gets(epUsers); got != 1 {
		t.Errorf("read the user list %d times, want 1", got)
	}
}

// "There are none of these yet" and "yours is not among these" send a reader
// somewhere different, which is the whole reason this type exists.
func TestMissingTellsAnEmptyCollectionFromAMiss(t *testing.T) {
	empty := missing{
		noun: "provider", ref: "openai",
		none: "no provider is currently listed",
		note: "check for a stopped model app",
	}.err().Error()
	if !strings.Contains(empty, "no provider named \"openai\"") ||
		!strings.Contains(empty, "no provider is currently listed") ||
		!strings.Contains(empty, "check for a stopped model app") {
		t.Errorf("empty-collection miss reads badly: %s", empty)
	}
	if strings.Contains(empty, "the listed ones are") {
		t.Errorf("empty-collection miss introduced a list it does not have: %s", empty)
	}

	found := missing{
		noun: "provider", ref: "openai",
		known: []string{"anthropic", "local"},
		have:  "the listed ones are",
		none:  "no provider is currently listed",
	}.err().Error()
	if !strings.Contains(found, "the listed ones are anthropic, local") {
		t.Errorf("miss did not name what exists: %s", found)
	}
	if strings.Contains(found, "no provider is currently listed") {
		t.Errorf("miss claimed the collection is empty while naming its rows: %s", found)
	}
}

// A deployment with two hundred keys should not turn one sentence into a
// screenful; past a dozen the reader runs the list verb regardless.
func TestMissingCapsTheNamesItLists(t *testing.T) {
	names := make([]string, 30)
	for i := range names {
		names[i] = "key-" + string(rune('a'+i%26))
	}
	msg := missing{noun: "key", ref: "nope", known: names, have: "yours are"}.err().Error()
	if !strings.Contains(msg, "and 18 more") {
		t.Errorf("uncapped list, or wrong remainder: %s", msg)
	}
	if strings.Count(msg, ", ") != namesInAMiss-1 {
		t.Errorf("listed something other than %d names: %s", namesInAMiss, msg)
	}
}

// olaresModelRow is one locally installed application's model, which is the
// only shape where a provider name is not enough to name a row.
func olaresModelRow(id, app, title, model string) adminModelRow {
	return adminModelRow{
		ProviderModelID: id,
		ProviderID:      "p-" + app,
		ProviderName:    "Olares",
		ProviderTitle:   &title,
		ProviderType:    "openai-compatible",
		ProviderSource:  "olares",
		ProviderStatus:  "active",
		OlaresAppName:   &app,
		Model:           providerModelRow{ID: id, Name: model, Mode: "chat", Enabled: true, Status: "active"},
	}
}

// The application name is what `router provider list` prints and what a person
// can retype; the title is the one the model list shows. Both name the row.
func TestAModelIsFoundByItsApplicationName(t *testing.T) {
	pc, log := newTestClient(t)
	running := "running"
	row := olaresModelRow("pm-1", "llamacppqwen3v3", "Qwen3 8B", "qwen3-8b")
	row.ProviderOlaresStatus = &running
	log.models = []adminModelRow{row}

	for _, ref := range []string{
		"llamacppqwen3v3/qwen3-8b",
		"Qwen3 8B/qwen3-8b",
		"Olares/qwen3-8b",
		"qwen3-8b",
		"pm-1",
	} {
		got, err := resolveModel(context.Background(), pc, ref)
		if err != nil {
			t.Errorf("%q did not resolve: %v", ref, err)
			continue
		}
		if got.ProviderModelID != "pm-1" {
			t.Errorf("%q resolved to %s", ref, got.ProviderModelID)
		}
	}
}

// Two applications serving one model name is the case the provider name cannot
// express, so the refusal has to hand back something that can be pasted.
func TestAnAmbiguousModelIsOfferedApplicationQualifiedNames(t *testing.T) {
	pc, log := newTestClient(t)
	log.models = []adminModelRow{
		olaresModelRow("pm-1", "llamacppqwen3v3", "Qwen3 8B (llama.cpp)", "qwen3-8b"),
		olaresModelRow("pm-2", "vllmqwen3v3", "Qwen3 8B (vLLM)", "qwen3-8b"),
	}

	_, err := resolveModel(context.Background(), pc, "qwen3-8b")
	if err == nil {
		t.Fatal("two rows with one name resolved to one of them")
	}
	msg := err.Error()
	for _, want := range []string{
		"llamacppqwen3v3/qwen3-8b [pm-1]",
		"vllmqwen3v3/qwen3-8b [pm-2]",
		"<app_name>/<model>",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("refusal is missing %q: %s", want, msg)
		}
	}

	if got, err := resolveModel(context.Background(), pc, "vllmqwen3v3/qwen3-8b"); err != nil {
		t.Errorf("the form the refusal offered did not work: %v", err)
	} else if got.ProviderModelID != "pm-2" {
		t.Errorf("resolved to %s, want pm-2", got.ProviderModelID)
	}
}

// A manual provider has no application name, so the candidate falls back to the
// label rather than printing a bare model name that names both rows.
func TestAnAmbiguousModelWithoutAnApplicationKeepsItsLabel(t *testing.T) {
	pc, log := newTestClient(t)
	rows := make([]adminModelRow, 2)
	for i, provider := range []string{"openai", "azure"} {
		rows[i] = adminModelRow{
			ProviderModelID: "pm-" + provider,
			ProviderName:    provider,
			ProviderSource:  "manual",
			ProviderStatus:  "active",
			Model:           providerModelRow{Name: "gpt-4o", Mode: "chat", Enabled: true, Status: "active"},
		}
	}
	log.models = rows

	_, err := resolveModel(context.Background(), pc, "gpt-4o")
	if err == nil {
		t.Fatal("two providers offering one model resolved to one of them")
	}
	if msg := err.Error(); !strings.Contains(msg, "openai/gpt-4o [pm-openai]") ||
		!strings.Contains(msg, "azure/gpt-4o [pm-azure]") {
		t.Errorf("refusal does not name both providers: %s", msg)
	}
}

func TestRequireRefNamesTheHandlesThatWouldWork(t *testing.T) {
	if _, err := requireRef("   ", "a provider name or id"); err == nil {
		t.Fatal("an empty argument was accepted")
	} else if !strings.Contains(err.Error(), "a provider name or id is required") {
		t.Errorf("refusal does not say what to pass: %v", err)
	}
	if got, err := requireRef("  openai\n", "a provider name or id"); err != nil || got != "openai" {
		t.Errorf("got %q, %v; want the trimmed word", got, err)
	}
}
