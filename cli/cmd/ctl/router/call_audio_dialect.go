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

// The operation catalogue answers the same question outright.
//
// Everything above is a correlation: an engine that designs voices is probably
// an engine that addresses them in the path. Since ADR-68 a model application
// can declare the routes it serves, and Router publishes that declaration on
// /v1/models. Where it exists there is nothing to infer — the model names the
// path — and Router enforces it, refusing an operation the catalogue does not
// list with `audio_operation_not_supported` rather than forwarding it. Trying
// the other spelling after that refusal is not a recovery, it is a second
// refusal.
//
// So the catalogue decides where it is trustworthy, and the correlation decides
// where it is not. Both remain: an application whose chart predates the
// catalogue publishes nothing, Router reconstructs an unauthoritative list from
// the capability flags, and that is exactly the case the guess was written for.

// synthesisRoutes is every path `speak` should try, best first.
func synthesisRoutes(ctx context.Context, dp *routerClient, model, voice string) []string {
	catalogue := audioCatalogue(ctx, dp)
	guess := speakRoutes(dialectFromCatalogue(catalogue, model), voice)
	return declaredFirst(catalogue, model, "POST", guess)
}

// voiceListRoutes is the same decision for reading a model's voices.
func voiceListRoutes(ctx context.Context, dp *routerClient, model string) []string {
	catalogue := audioCatalogue(ctx, dp)
	guess := voicesRoutes(dialectFromCatalogue(catalogue, model))
	return declaredFirst(catalogue, model, "GET", guess)
}

// audioCatalogue reads the list the caller's own credential sees, asking for
// the per-model operations along with it. Failure of any kind is an empty list
// rather than an error: the worst a missing catalogue does is restore the
// behaviour of guessing and asking twice.
func audioCatalogue(ctx context.Context, dp *routerClient) []modelObject {
	var resp modelsListResponse
	if err := dp.doJSON(ctx, "GET", modelsPath(false, true), nil, &resp); err != nil {
		return nil
	}
	return resp.Data
}

// declaredFirst narrows the guessed candidates to the ones the model declares.
//
// Only a trustworthy catalogue may narrow anything, and narrowing to nothing is
// not the same as having no opinion: a model with a declared catalogue that
// lists neither spelling does not serve this operation, and the useful outcome
// is Router saying so once rather than the engine 404ing twice. So the first
// candidate is kept and the request goes, which is what makes the refusal
// arrive with `audio_operation_not_supported` on it.
func declaredFirst(catalogue []modelObject, model, method string, candidates []string) []string {
	m := trustedEntry(catalogue, model)
	if m == nil || len(candidates) == 0 {
		return candidates
	}
	kept := make([]string, 0, len(candidates))
	for _, route := range candidates {
		if _, ok := m.declaresOperation(method, route, "http"); ok {
			kept = append(kept, route)
		}
	}
	if len(kept) == 0 {
		return candidates[:1]
	}
	return kept
}

// trustedEntry finds the row whose catalogue may be believed, or nil.
//
// A category — `default-tts`, or no --model at all — matches no row, because a
// category describes no single model and is deliberately absent from
// /v1/models. It is answered the way the dialect is: by what the installed
// synthesis models agree on, and only when they all agree, since Router picks
// which of them serves the category and disagreement means the answer depends
// on a choice not made yet.
func trustedEntry(catalogue []modelObject, model string) *modelObject {
	ref := strings.TrimSpace(model)
	for i := range catalogue {
		m := &catalogue[i]
		if ref != "" && (m.ID == ref || m.QualifiedID == ref) {
			if m.trustworthyCatalogue() {
				return m
			}
			return nil
		}
	}
	var agreed *modelObject
	for i := range catalogue {
		m := &catalogue[i]
		if m.Mode != "tts" {
			continue
		}
		if !m.trustworthyCatalogue() {
			return nil
		}
		if agreed == nil {
			agreed = m
			continue
		}
		if !sameDeclaredPaths(agreed, m) {
			return nil
		}
	}
	return agreed
}

func sameDeclaredPaths(a, b *modelObject) bool {
	if len(a.Operations) != len(b.Operations) {
		return false
	}
	seen := make(map[string]int, len(a.Operations))
	for i := range a.Operations {
		seen[a.Operations[i].Method+" "+a.Operations[i].PathTemplate]++
	}
	for i := range b.Operations {
		seen[b.Operations[i].Method+" "+b.Operations[i].PathTemplate]--
	}
	for _, n := range seen {
		if n != 0 {
			return false
		}
	}
	return true
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
