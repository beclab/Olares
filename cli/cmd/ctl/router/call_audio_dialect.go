package router

import (
	"context"
	"strings"
)

// Which spelling of synthesis an engine serves.
//
// Router mounts both /v1/audio/speech and /v1/text-to-speech/<voice> and
// forwards each unchanged, translating between them only for the ElevenLabs
// vendor. A locally installed engine therefore answers one and 404s the other,
// and which one is a property of the image the model application was built
// from rather than anything Router or the caller decides.
//
// Trying one and retrying the other always works and costs a wasted request
// against the engine, which Router records as a failed call. Asking the
// catalogue first is cheaper: it is answered from Router's own rows, reaches no
// engine, and leaves nothing in the record. So the catalogue picks the order
// and the retry stays as the thing that makes a wrong guess harmless.
type ttsDialect int

const (
	dialectUnknown ttsDialect = iota
	dialectOpenAI
	dialectElevenLabs
)

// capTTSDesign is the flag that separates them. Designing a voice from a
// description is an ElevenLabs operation — /v1/text-to-voice/design — so an
// engine that declares it is an engine that implements that surface, and the
// images that speak the OpenAI shape declare speaking and cloning without it.
//
// This is a correlation across the engines that exist rather than a rule Router
// enforces, which is exactly why it only chooses the order.
const capTTSDesign = "tts_design"

// ttsDialectOf reads the catalogue the caller's own credential sees. Failure of
// any kind is dialectUnknown rather than an error: the worst a missing hint
// does is restore the behaviour of asking twice.
func ttsDialectOf(ctx context.Context, dp *routerClient, model string) ttsDialect {
	var resp modelsListResponse
	if err := dp.doJSON(ctx, "GET", epDataPlaneModels, nil, &resp); err != nil {
		return dialectUnknown
	}
	return dialectFromCatalogue(resp.Data, model)
}

// dialectFromCatalogue matches the reference against the list, and falls back
// to what the installed synthesis models agree on when it cannot.
//
// The fallback is what makes this useful at all: `speak` is normally called
// with no --model, and a default category is deliberately absent from /v1/models
// because it describes no single model. So a category is answered by whichever
// tts models are installed, and when they all speak one shape the category does
// too. When they disagree there is no honest answer and the retry decides.
func dialectFromCatalogue(items []modelObject, model string) ttsDialect {
	ref := strings.TrimSpace(model)
	for i := range items {
		m := &items[i]
		if ref != "" && (m.ID == ref || m.QualifiedID == ref) {
			return dialectOf(m.Supports)
		}
	}
	agreed := dialectUnknown
	for i := range items {
		m := &items[i]
		if m.Mode != "tts" {
			continue
		}
		switch d := dialectOf(m.Supports); {
		case agreed == dialectUnknown:
			agreed = d
		case agreed != d:
			return dialectUnknown
		}
	}
	return agreed
}

func dialectOf(supports []string) ttsDialect {
	for _, s := range supports {
		if s == capTTSDesign {
			return dialectElevenLabs
		}
	}
	return dialectOpenAI
}

// speakRoutes is every path worth trying, best first.
//
// Without a voice there is only one: the other spelling puts the voice in the
// path and has nowhere to put an absent one, so an engine of that shape cannot
// speak at all here and the refusal says so rather than a second 404 doing it.
func speakRoutes(d ttsDialect, voice string) []string {
	v := strings.TrimSpace(voice)
	if v == "" {
		return []string{epAudioSpeech}
	}
	if d == dialectElevenLabs {
		return []string{epSpeakAs(v), epAudioSpeech}
	}
	return []string{epAudioSpeech, epSpeakAs(v)}
}

// voicesRoutes is the same ordering for the catalogue read. The OpenAI spelling
// stays first when nothing is known, so an engine that serves both keeps
// answering the one it always answered.
func voicesRoutes(d ttsDialect) []string {
	if d == dialectElevenLabs {
		return []string{epVoices, epAudioVoices}
	}
	return []string{epAudioVoices, epVoices}
}
