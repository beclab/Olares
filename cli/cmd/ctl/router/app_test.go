package router

import (
	"bytes"
	"strings"
	"testing"
)

// The catalog's TAKES column is the only place a user learns that a row cannot
// be installed at all, so each of its three answers is worth pinning: a
// template says clone, an ordinary application says install, and a row the
// Market did not answer for says neither.
func TestCatalogSaysWhichVerbEachRowTakes(t *testing.T) {
	items := []marketApp{
		{AppName: "llamacppllmbasev3", Title: "llama.cpp base"},
		{AppName: "embeddinggemmav3", Title: "EmbeddingGemma"},
		{AppName: "fromanothersource", Title: "Elsewhere"},
	}
	templates := map[string]bool{
		"llamacppllmbasev3": true,
		"embeddinggemmav3":  false,
	}

	var out bytes.Buffer
	if err := renderCatalog(&out, items, templates); err != nil {
		t.Fatalf("renderCatalog: %v", err)
	}
	got := out.String()

	for _, want := range []string{
		"llamacppllmbasev3",
		"clone",
		"install",
		"market clone",
		"did not answer",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("catalog output is missing %q:\n%s", want, got)
		}
	}
}

// A catalog of ordinary applications must not explain clone or an unanswered
// row: both notes describe something the reader is not looking at.
func TestCatalogExplainsOnlyWhatItShows(t *testing.T) {
	var out bytes.Buffer
	if err := renderCatalog(&out, []marketApp{{AppName: "embeddinggemmav3"}}, map[string]bool{"embeddinggemmav3": false}); err != nil {
		t.Fatalf("renderCatalog: %v", err)
	}
	got := out.String()
	for _, unwanted := range []string{"market clone", "did not answer"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("catalog output should not mention %q:\n%s", unwanted, got)
		}
	}
}
