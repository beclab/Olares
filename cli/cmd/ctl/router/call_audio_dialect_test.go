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

// The catalogue picks the order. Naming the model is the direct case; naming a
// category is the ordinary one, because `speak` is usually called with no
// --model at all and a category is deliberately absent from /v1/models.
func TestTheCatalogueSaysWhichSpellingToTryFirst(t *testing.T) {
	el := modelObject{ID: "Olares/Breeze", Mode: "tts", Supports: []string{"tts", "tts_clone", "tts_design"}}
	openai := modelObject{ID: "Olares/Qwen3-TTS", Mode: "tts", Supports: []string{"tts", "tts_clone"}}
	chat := modelObject{ID: "Olares/Qwen", Mode: "chat", Supports: []string{"reasoning"}}

	cases := []struct {
		name  string
		items []modelObject
		model string
		want  ttsDialect
	}{
		{"named model declares design", []modelObject{chat, el, openai}, "Olares/Breeze", dialectElevenLabs},
		{"named model does not", []modelObject{chat, el, openai}, "Olares/Qwen3-TTS", dialectOpenAI},
		{"a category, and the installed models agree", []modelObject{chat, el}, "default-tts", dialectElevenLabs},
		{"a category, and they disagree", []modelObject{el, openai}, "default-tts", dialectUnknown},
		{"a category with no synthesis installed", []modelObject{chat}, "default-tts", dialectUnknown},
		{"nothing in the catalogue at all", nil, "", dialectUnknown},
	}
	for _, c := range cases {
		if got := dialectFromCatalogue(c.items, c.model); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
	// A chat model is not consulted about synthesis even when it is the only
	// thing that matches nothing else.
	if got := dialectFromCatalogue([]modelObject{chat, openai}, "default-tts"); got != dialectOpenAI {
		t.Errorf("a non-tts row was allowed to disagree: got %v", got)
	}
}

// The hint decides the order and never the outcome. Both paths stay reachable
// whichever way it points, so an engine that breaks the correlation still works
// through the retry rather than becoming uncallable.
func TestAHintReordersTheAttemptsWithoutRemovingAny(t *testing.T) {
	for _, d := range []ttsDialect{dialectUnknown, dialectOpenAI, dialectElevenLabs} {
		got := speakRoutes(d, "en-f")
		if len(got) != 2 {
			t.Errorf("%v: %d routes, want both", d, len(got))
			continue
		}
		if got[0] == got[1] {
			t.Errorf("%v: the same route twice: %q", d, got[0])
		}
	}
	if got := speakRoutes(dialectElevenLabs, "en-f"); got[0] != epSpeakAs("en-f") {
		t.Errorf("the native spelling was not tried first: %q", got[0])
	}
	if got := speakRoutes(dialectUnknown, "en-f"); got[0] != epAudioSpeech {
		t.Errorf("with no hint the established order changed: %q", got[0])
	}
	// No voice means the path-addressing spelling has nowhere to put one, so
	// it is not an attempt worth making.
	if got := speakRoutes(dialectElevenLabs, "  "); len(got) != 1 || got[0] != epAudioSpeech {
		t.Errorf("a voiceless call tried to address a voice: %v", got)
	}
	if got := voicesRoutes(dialectElevenLabs); got[0] != epVoices {
		t.Errorf("the voice list did not follow the hint: %q", got[0])
	}
	if got := voicesRoutes(dialectUnknown); got[0] != epAudioVoices {
		t.Errorf("with no hint the established order changed: %q", got[0])
	}
}

// Where a model declares its routes there is nothing to correlate: the guess
// narrows to what the card names, and the spelling it does not name is not
// tried at all. Router refuses that one anyway, so the second attempt could
// only ever buy a second refusal and a second $0 row.
func TestADeclaredCatalogueLeavesOnlyTheRouteTheModelNames(t *testing.T) {
	yes := true
	el := modelObject{
		ID: "Olares/Breeze", Mode: "tts", Supports: []string{"tts", "tts_design"},
		Authoritative: &yes,
		Operations: []modelOperation{
			{ID: "speech.synthesize", Method: "POST", PathTemplate: epTextToSpeech + "/{voice_id}", Transport: "http"},
			{ID: "voice.list", Method: "GET", PathTemplate: epVoices, Transport: "http"},
		},
	}
	got := declaredFirst([]modelObject{el}, "Olares/Breeze", "POST", speakRoutes(dialectElevenLabs, "en-f"))
	if len(got) != 1 || got[0] != epSpeakAs("en-f") {
		t.Fatalf("expected only the declared spelling, got %v", got)
	}
	// The escaped voice has to survive the match: the template is matched a
	// segment at a time, and a voice id is one segment whatever is in it.
	odd := declaredFirst([]modelObject{el}, "Olares/Breeze", "POST", speakRoutes(dialectElevenLabs, "my voice/2"))
	if len(odd) != 1 {
		t.Fatalf("an escaped voice id stopped matching its own template: %v", odd)
	}
	if list := declaredFirst([]modelObject{el}, "Olares/Breeze", "GET", voicesRoutes(dialectElevenLabs)); len(list) != 1 || list[0] != epVoices {
		t.Fatalf("expected only the declared voice list, got %v", list)
	}
}

// A declared catalogue that names neither spelling is a model that does not do
// this job. One request gets Router's own `audio_operation_not_supported`,
// which says so; two get it twice.
func TestAModelThatDeclaresNeitherSpellingIsAskedOnce(t *testing.T) {
	yes := true
	m := modelObject{
		ID: "Olares/ASR", Mode: "tts", Authoritative: &yes,
		Operations: []modelOperation{{ID: "audio.transcribe", Method: "POST", PathTemplate: epAudioTranscriptions, Transport: "http"}},
	}
	candidates := speakRoutes(dialectUnknown, "en-f")
	got := declaredFirst([]modelObject{m}, "Olares/ASR", "POST", candidates)
	if len(got) != 1 || got[0] != candidates[0] {
		t.Fatalf("expected one attempt, got %v", got)
	}
}

// Nothing trustworthy narrows anything. An application whose chart predates the
// catalogue is exactly the case the correlation was written for, and a stale
// one describes an engine that may since have been relaunched onto other flags.
func TestOnlyATrustworthyCatalogueIsAllowedToNarrow(t *testing.T) {
	yes, no := true, false
	op := []modelOperation{{ID: "speech.synthesize", Method: "POST", PathTemplate: epTextToSpeech + "/{voice_id}", Transport: "http"}}
	both := speakRoutes(dialectUnknown, "en-f")

	cases := []struct {
		name string
		m    modelObject
	}{
		{"reconstructed from capabilities", modelObject{ID: "m", Mode: "tts", Authoritative: &no, Operations: op}},
		{"a Router that publishes none", modelObject{ID: "m", Mode: "tts", Operations: op}},
		{"declared but aged out", modelObject{ID: "m", Mode: "tts", Authoritative: &yes, CapabilityStale: true, Operations: op}},
		{"declared and empty", modelObject{ID: "m", Mode: "tts", Authoritative: &yes}},
	}
	for _, c := range cases {
		if got := declaredFirst([]modelObject{c.m}, "m", "POST", both); len(got) != 2 {
			t.Errorf("%s: narrowed to %v, and both spellings should stay reachable", c.name, got)
		}
	}
}

// `speak` is normally called with no --model, and a category matches no row.
// It is answered the way the dialect is — by what the installed models agree
// on — because Router picks which of them serves the category.
func TestACategoryIsAnsweredOnlyWhenTheModelsAgree(t *testing.T) {
	yes := true
	native := func(id string) modelObject {
		return modelObject{ID: id, Mode: "tts", Authoritative: &yes,
			Operations: []modelOperation{{ID: "speech.synthesize", Method: "POST", PathTemplate: epTextToSpeech + "/{voice_id}", Transport: "http"}}}
	}
	openai := modelObject{ID: "Olares/Qwen3-TTS", Mode: "tts", Authoritative: &yes,
		Operations: []modelOperation{{ID: "speech.synthesize", Method: "POST", PathTemplate: epAudioSpeech, Transport: "http"}}}
	chat := modelObject{ID: "Olares/Qwen", Mode: "chat"}

	agree := []modelObject{chat, native("Olares/Breeze"), native("Olares/Breeze2")}
	if got := declaredFirst(agree, "", "POST", speakRoutes(dialectUnknown, "en-f")); len(got) != 1 || got[0] != epSpeakAs("en-f") {
		t.Fatalf("agreeing models did not answer the category: %v", got)
	}
	disagree := []modelObject{native("Olares/Breeze"), openai}
	if got := declaredFirst(disagree, "", "POST", speakRoutes(dialectUnknown, "en-f")); len(got) != 2 {
		t.Fatalf("a category served by either shape must stay open to both: %v", got)
	}
	// One model without a trustworthy catalogue is enough to stop the whole
	// consensus: it may be the one Router routes the category to.
	mixed := []modelObject{native("Olares/Breeze"), {ID: "Olares/Old", Mode: "tts"}}
	if got := declaredFirst(mixed, "", "POST", speakRoutes(dialectUnknown, "en-f")); len(got) != 2 {
		t.Fatalf("an undeclared model was not allowed to withhold consensus: %v", got)
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
