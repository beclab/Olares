package router

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// Every call verb has to say what it falls back to. Router refuses a request
// whose `model` is empty rather than picking something, and the CLI's answer to
// an omitted --model is the default category for that kind of work — so a verb
// that ships without one is a verb that fails on the ordinary invocation, and
// fails with Router's message rather than with anything pointing here.
//
// Ten verbs arrived at once. This is the check that the eleventh does not
// arrive without its category.
func TestEveryCallVerbFallsBackToACategory(t *testing.T) {
	// Translate is the deliberate exception: its routes resolve the translate
	// default themselves, per call, so there is no model for a caller to name
	// and no flag to give one.
	const noModelFlag = "translate"

	for _, verb := range callVerbs(t) {
		// `call models` sits here because it is answered by the data plane
		// over the same credential, but it sends no work and so has nothing
		// to fall back to.
		if verb.Name() == "models" {
			continue
		}
		flag := verb.Flags().Lookup("model")
		if verb.Name() == noModelFlag {
			if flag != nil {
				t.Errorf("call %s: has a --model flag; the translate routes choose their own model",
					verb.Name())
			}
			continue
		}
		if flag == nil {
			t.Errorf("call %s: no --model flag", verb.Name())
			continue
		}
		if !strings.Contains(flag.Usage, "default-") {
			t.Errorf("call %s: --model does not name the category it falls back to: %q",
				verb.Name(), flag.Usage)
		}
	}
}

// The categories are hand-copied from Router's own registry, which this tree
// cannot read. Nothing here can confirm a name exists upstream; what it can
// confirm is that every one of them is spelled like a default route and like a
// route at all. A name missing the prefix resolves as an ordinary alias, and an
// alias that happens not to exist is a 404 rather than the "nothing serves this
// kind of work" that was meant; a name Router's own rule would reject cannot be
// any route, which is the same 404 arriving by a different road.
//
// checkRouteName is not the test, because it refuses the prefix on purpose —
// that is the rule for a name a person may create. Here the prefix is the point,
// so what is borrowed is the rest of the rule.
func TestEveryCategoryIsSpelledLikeADefaultRoute(t *testing.T) {
	categories := map[string]string{
		"chat":        categoryChat,
		"embedding":   categoryEmbedding,
		"rerank":      categoryRerank,
		"search":      categorySearch,
		"scrape":      categoryScrape,
		"image":       categoryImage,
		"video":       categoryVideo,
		"ocr":         categoryOCR,
		"stt":         categorySTT,
		"tts":         categoryTTS,
		"vad":         categoryVAD,
		"diarization": categoryDiarization,
		"enhance":     categoryEnhance,
		"align":       categoryAlign,
	}
	for name, value := range categories {
		if !strings.HasPrefix(value, defaultNamePrefix) {
			t.Errorf("%s: %q does not begin %q, so Router would read it as an alias",
				name, value, defaultNamePrefix)
		}
		if err := checkRouteName(strings.TrimPrefix(value, defaultNamePrefix)); err != nil {
			t.Errorf("%s: %q is not a name Router would accept: %v", name, value, err)
		}
	}
}

// callModel prefers what was asked for and falls back to the category. An empty
// string reaching the wire is the one outcome that has no useful failure.
func TestCallModelNeverSendsNothing(t *testing.T) {
	if got := callModel("openai/gpt-5", categoryChat); got != "openai/gpt-5" {
		t.Errorf("explicit model: got %q", got)
	}
	if got := callModel("  ", categoryChat); got != categoryChat {
		t.Errorf("blank model: got %q want the category", got)
	}
	if got := callModel("", categoryOCR); got == "" {
		t.Error("omitted model resolved to nothing")
	}
}

func callVerbs(t *testing.T) []*cobra.Command {
	t.Helper()
	call := NewCallCommand(&cmdutil.Factory{})
	verbs := call.Commands()
	if len(verbs) < 15 {
		t.Fatalf("found %d call verbs; the subtree moved", len(verbs))
	}
	return verbs
}
