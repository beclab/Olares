/*
Copyright 2024 bytetrade.io

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package router

import "strings"

// modelPhase is one word of the vocabulary a Model Console uses for what the
// weights behind a model are doing, with the two things every reader of it
// needs: whether a call would be answered, and how to say so.
//
// This is one table because it used to be three — a servability set, a table
// cell in `model list`, and a sentence in `model status` — and a phase Router
// adds next has to arrive in one place rather than three. The two renderings
// stay separate fields instead of one string because they are read in
// different postures: a cell shares a line with five other columns and a
// sentence is the whole answer.
type modelPhase struct {
	// servable is whether a call would be dispatched during this phase.
	servable bool
	// cell is the CALLABLE column's reason in `model list`, printed after
	// "no · ". Empty for a servable phase, which never needs one.
	cell string
	// sentence is the STATE line in `model status`: what the phase means for
	// somebody waiting, rather than the word itself.
	sentence string
	// readiness is the word `router call models` uses for the same condition,
	// which is coarser than the phase: several phases are one kind of waiting.
	// It is a separate axis from servable — `failed` and `warming` both refuse
	// calls, and only one of them will stop doing so on its own.
	readiness string
}

// modelPhases is llm-init's lifecycle, and nothing here can check it. A phase
// this build has never heard of is treated as servable, which is the same
// choice made for a manual provider, an application running its own engine and
// an application before its first observation: the vocabulary belongs to the
// application, and not recognising a word is not evidence against the model
// behind it.
//
// `degraded` is servable on purpose. llm-init's health loop moves a model
// between ready and degraded while the engine usually still answers, and an
// engine that genuinely cannot answer refuses with its own 503 — which is a
// better answer than a list that hid the model.
var modelPhases = map[string]modelPhase{
	"init": {readiness: "warming", cell: "model service starting",
		sentence: "starting up; nothing has been fetched yet"},
	"download": {readiness: "warming", cell: "fetching weights",
		sentence: "fetching the weights; calls are refused until this finishes"},
	"loading": {readiness: "warming", cell: "engine loading",
		sentence: "the engine is starting on weights that are already on disk"},
	"failed": {readiness: "failed", cell: "model load failed",
		sentence: "stopped trying; `router model retry` re-enters the loop"},
	"ready": {servable: true, readiness: "ready",
		sentence: "serving"},
	"degraded": {servable: true, readiness: "ready",
		sentence: "serving in a reduced state; `router model progress` carries the reason"},
}

func lookupPhase(phase string) (modelPhase, bool) {
	p, ok := modelPhases[strings.ToLower(strings.TrimSpace(phase))]
	return p, ok
}

// phaseBlocksCalls reports a phase known to refuse calls. An unknown phase is
// not one of them; see modelPhases.
func phaseBlocksCalls(phase string) bool {
	p, ok := lookupPhase(phase)
	return ok && !p.servable
}

// phaseNote translates the lifecycle phase into what it means for a caller
// waiting on the model. The phase names are the Model Console's; the
// consequence is what a reader actually wants.
func phaseNote(phase string, h *localHealth) string {
	p, ok := lookupPhase(phase)
	if !ok {
		return ""
	}
	if p.servable && p.sentence == "serving" && h != nil && !h.ModelExists {
		return "serving, but the engine no longer reports the configured model"
	}
	return p.sentence
}
