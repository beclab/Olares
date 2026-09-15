package router

import (
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// The voice surface sits at the root of /v1 rather than under /audio, because
// it is the ElevenLabs shape a synthesis engine serves alongside the OpenAI
// one. Spelling one of these under /audio reaches a path Router does not mount
// for it, which arrives as a 404 that reads like a missing capability.
func TestTheVoiceLibraryIsNotUnderTheAudioPrefix(t *testing.T) {
	for name, path := range map[string]string{
		"list":     epVoices,
		"add":      epVoicesAdd,
		"settings": epVoiceSettings,
		"design":   epTextToVoiceDesign,
		"keep":     epTextToVoice,
		"history":  epHistory,
		"one":      epVoice("v1"),
		"reading":  epHistoryItem("h1"),
		"audio":    epHistoryAudio("h1"),
	} {
		if strings.HasPrefix(path, dataPlaneAPI+"/audio/") {
			t.Errorf("%s is spelled %q, under the audio prefix Router does not mount it on", name, path)
		}
		if !strings.HasPrefix(path, dataPlaneAPI+"/") {
			t.Errorf("%s is spelled %q, which is not a data-plane route", name, path)
		}
	}
}

// An id with a slash or a space in it must not become part of the path. These
// ids come from the engine and nothing here constrains their alphabet.
func TestAnIdIsEscapedIntoItsPath(t *testing.T) {
	if got := epVoice("a/b"); got != epVoices+"/a%2Fb" {
		t.Errorf("epVoice did not escape the id: %q", got)
	}
	if got := epHistoryAudio("a b"); got != epHistory+"/a%20b/audio" {
		t.Errorf("epHistoryAudio did not escape the id: %q", got)
	}
}

// Each voice verb resolves the category its capability implies. They are not
// one category with three names: an engine that speaks may not clone and may
// not design, so sending a recording at `default-tts` would hand the upload to
// something with nothing to do with it.
func TestEachVoiceVerbResolvesTheCapabilityItNeeds(t *testing.T) {
	want := map[string]string{
		"list":     categoryTTS,
		"get":      categoryTTS,
		"delete":   categoryTTS,
		"settings": categoryTTS,
		"add":      categoryTTSClone,
		"design":   categoryTTSDesign,
	}
	noun := newCallVoiceCommand(cmdutil.NewFactory())
	seen := map[string]bool{}
	for _, sub := range noun.Commands() {
		flag := sub.Flags().Lookup("model")
		if flag == nil {
			t.Errorf("voice %s: no --model flag", sub.Name())
			continue
		}
		category, ok := want[sub.Name()]
		if !ok {
			t.Errorf("voice %s: a new verb with no category recorded here", sub.Name())
			continue
		}
		seen[sub.Name()] = true
		if !strings.Contains(flag.Usage, category) {
			t.Errorf("voice %s: --model does not resolve %s: %q", sub.Name(), category, flag.Usage)
		}
	}
	for name := range want {
		if !seen[name] {
			t.Errorf("voice %s is written down here and no longer exists", name)
		}
	}
}

// History reaches no category at all, so the refusal has to explain the
// absence rather than let the dispatcher answer about a model nobody named.
func TestHistoryRefusesToGuessAModel(t *testing.T) {
	err := requireHistoryModel("  ")
	if err == nil {
		t.Fatal("an empty --model was accepted, and Router has no default for history to resolve")
	}
	msg := err.Error()
	for _, want := range []string{"--model is required", "no default category", "model list"} {
		if !strings.Contains(msg, want) {
			t.Errorf("the refusal does not say %q: %s", want, msg)
		}
	}
	if err := requireHistoryModel("Olares/tts"); err != nil {
		t.Errorf("a named model was refused: %v", err)
	}
}

// Naming the category in the help text is not the same as sending it. Router's
// audio gate table gives most of these suffixes a default and does not give one
// to `/v1/voices/settings/default`, so a verb that leaves the model out because
// "Router knows" is refused for a missing model on exactly one of its routes.
// Sending the category always makes the two the same request.
func TestAVerbSendsTheCategoryItsHelpPromises(t *testing.T) {
	for _, tc := range []struct{ route, category string }{
		{epVoices, categoryTTS},
		{epVoiceSettings, categoryTTS},
		{epVoicesAdd, categoryTTSClone},
		{epTextToVoiceDesign, categoryTTSDesign},
	} {
		got := voicePath(tc.route, callModel("", tc.category))
		if !strings.Contains(got, "model="+tc.category) {
			t.Errorf("%s went out without %s: %q", tc.route, tc.category, got)
		}
	}
	// A named model still wins: the category is the fallback, not an override.
	if got := voicePath(epVoices, callModel("Olares/x", categoryTTS)); strings.Contains(got, categoryTTS) {
		t.Errorf("the category displaced the model the caller named: %q", got)
	}
}

// The model travels on the query. Several of these routes carry no body at
// all, so a body is not a place the reference could go.
func TestTheModelTravelsOnTheQueryForEveryVoiceRoute(t *testing.T) {
	if got := voicePath(epVoices, "Olares/x"); got != epVoices+"?model=Olares%2Fx" {
		t.Errorf("voicePath did not put the model on the query: %q", got)
	}
	if got := voicePath(epVoices, "   "); got != epVoices {
		t.Errorf("a blank model became a query parameter: %q", got)
	}
}

// A reading names its voice by name when it has one, because an id says
// nothing to the person reading the table.
func TestAReadingIsNamedByItsVoiceRatherThanItsId(t *testing.T) {
	named := historyItem{VoiceID: "vc_1", VoiceName: "Night Host"}
	if got := historyVoice(&named); got != "Night Host" {
		t.Errorf("the name was not preferred: %q", got)
	}
	bare := historyItem{VoiceID: "vc_1"}
	if got := historyVoice(&bare); got != "vc_1" {
		t.Errorf("the id was not the fallback: %q", got)
	}
}

// A reading with no timestamp is a reading whose time is unknown, not one
// performed in 1970.
func TestATimeThatIsNotThereIsNotPrintedAsAnEpoch(t *testing.T) {
	if got := historyWhen(0); got != "-" {
		t.Errorf("a missing timestamp rendered as %q", got)
	}
	if got := historyWhen(1756000000); strings.HasPrefix(got, "1970") {
		t.Errorf("a real timestamp rendered as %q", got)
	}
}
