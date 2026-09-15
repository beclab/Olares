package router

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Counting is Anthropic-shaped even though the CLI takes plain text, because
// that is the ingress Router mounts it on.
func TestAskingHowLargeATurnIsSpellsItTheWayTheRouteReadsIt(t *testing.T) {
	body, err := countTokensBody(countTokensOptions{
		Args: []string{"how", "big"}, Model: "default-chat", System: "be brief",
	})
	if err != nil {
		t.Fatal(err)
	}
	if body["model"] != "default-chat" {
		t.Errorf("the model did not travel: %v", body["model"])
	}
	if body["system"] != "be brief" {
		t.Errorf("the system prompt did not travel: %v", body["system"])
	}
	messages, ok := body["messages"].([]map[string]any)
	if !ok || len(messages) != 1 {
		t.Fatalf("the text did not become a turn: %#v", body["messages"])
	}
	if messages[0]["role"] != "user" || messages[0]["content"] != "how big" {
		t.Errorf("the turn is not the text that was given: %#v", messages[0])
	}
	if _, present := body["system"]; !present {
		t.Error("the system prompt vanished")
	}

	bare, err := countTokensBody(countTokensOptions{Args: []string{"hi"}, Model: "m"})
	if err != nil {
		t.Fatal(err)
	}
	if _, present := bare["system"]; present {
		t.Error("a system prompt nobody gave was sent")
	}
}

// A document written for one vendor names a model this Router may not have,
// and which provider is asked decides whether the count is official or
// estimated. So the resolved model replaces whatever the file said.
func TestAFileIsCountedAgainstTheModelThatWasNamedHere(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "turn.json")
	body := `{"model":"claude-3-from-somewhere-else","messages":[{"role":"user","content":"hi"}]}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := countTokensBody(countTokensOptions{File: path, Model: "anthropic/claude-sonnet-4"})
	if err != nil {
		t.Fatal(err)
	}
	if got["model"] != "anthropic/claude-sonnet-4" {
		t.Errorf("the file's own model was sent: %v", got["model"])
	}
	if got["messages"] == nil {
		t.Error("the turn was dropped")
	}
}

// A document with no messages is not a turn, and saying so here is better than
// a 400 about a field the caller cannot see from the command line.
func TestADocumentWithNoTurnIsRefusedHere(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(path, []byte(`{"model":"x"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := countTokensBody(countTokensOptions{File: path, Model: "m"})
	if err == nil || !strings.Contains(err.Error(), "messages") {
		t.Fatalf("a document with no turn was accepted: %v", err)
	}
}

// The counting route is mounted apart from the rest of the Anthropic ingress,
// above the quota line. Spelling it anywhere else reaches a metered route.
func TestCountingIsSpelledOnItsOwnRoute(t *testing.T) {
	if epMessagesCountTokens != dataPlaneAPI+"/messages/count_tokens" {
		t.Errorf("the counting route is spelled %q", epMessagesCountTokens)
	}
}
