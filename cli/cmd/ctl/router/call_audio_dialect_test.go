package router

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// Synthesis is spelled two ways on this platform and Router translates between
// them only for the ElevenLabs vendor, so a locally installed engine answers
// one spelling and 404s the other. These tests pin the reading of that 404,
// because getting it wrong in either direction is silent: too narrow and
// `speak` stays broken against half the engines, too wide and a rejected key
// or an unresolved model is retried as if it were a missing route.
func TestOnlyAMissingRouteIsWorthAskingTheOtherWay(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"the engine has no such path", &RouterError{Status: 404}, true},
		{"still readable through a hint", fmt.Errorf("%w\nadvice", &RouterError{Status: 404}), true},
		{"Router's own 404 about a model", &RouterError{Status: 404, Code: "model_not_found"}, false},
		{"the key was refused", &RouterError{Status: 401, Code: "invalid_api_key"}, false},
		{"the model is the wrong mode", &RouterError{Status: 422, Code: "audio_mode_mismatch"}, false},
		{"nothing went wrong", nil, false},
		{"not Router at all", errors.New("dial tcp: connection refused"), false},
	}
	for _, c := range cases {
		if got := routeAbsent(c.err); got != c.want {
			t.Errorf("%s: routeAbsent = %v, want %v", c.name, got, c.want)
		}
	}
}

// The second spelling addresses the voice in the path. An id that needs
// escaping is the ordinary case for a cloned voice, whose name comes from
// whatever the caller typed.
func TestTheOtherSpellingCarriesTheVoiceInThePath(t *testing.T) {
	if got := epSpeakAs("en-f"); got != epTextToSpeech+"/en-f" {
		t.Errorf("got %q", got)
	}
	if got := epSpeakAs("my voice/2"); got != epTextToSpeech+"/my%20voice%2F2" {
		t.Errorf("an id was not escaped into its path: %q", got)
	}
	// Both spellings still carry the model, which is what Router resolves on.
	if got := audioRequestPath(epSpeakAs("en-f"), "default-tts", false); !strings.Contains(got, "model=default-tts") {
		t.Errorf("the retry dropped the model: %q", got)
	}
}

// The body does not change between the spellings: the fields this verb sends
// are the ones both shapes read. If that stops being true the retry silently
// speaks with the wrong voice or the wrong format rather than failing.
func TestOneBodyIsReadByBothSpellings(t *testing.T) {
	speed := 1.25
	req := buildSpeakRequest("hello", speakOptions{
		Model: "Olares/tts", Voice: "en-f", RespFormat: "wav", Speed: &speed,
	})
	for _, key := range []string{"input", "voice", "response_format"} {
		if _, ok := req[key]; !ok {
			t.Errorf("%q is missing, and it is the key the native shape reads", key)
		}
	}
}

// A voice is optional in the body and mandatory in the path, so the one case
// the retry cannot serve is a caller who named none. Reporting that as the
// generic "wrong engine" 404 sends them looking for a model they already have.
func TestSpeakingWithNoVoiceExplainsTheSpellingRatherThanTheEngine(t *testing.T) {
	hinted := callErr(&RouterError{Status: 404})
	if !strings.Contains(hinted.Error(), "wrong engine") {
		t.Fatal("the generic audio hint has changed; this test guards replacing it")
	}
	got := noVoiceToAddressErr(hinted).Error()
	if strings.Contains(got, "wrong engine") {
		t.Error("the generic hint survived alongside the real reason")
	}
	for _, want := range []string{"--voices", "--voice <id>"} {
		if !strings.Contains(got, want) {
			t.Errorf("the refusal does not say %q: %s", want, got)
		}
	}
	// Something that is not Router's is passed through rather than relabelled.
	plain := errors.New("dial tcp: connection refused")
	if got := noVoiceToAddressErr(plain); got != plain {
		t.Errorf("a transport failure was explained as a spelling: %v", got)
	}
}
