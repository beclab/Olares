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
	// The verbs Router resolves no default for. Each is listed with its reason
	// at the check below, because "it needs a model" is a claim about Router's
	// registry rather than a choice made here.
	requiresAModel := map[string]bool{"responses": true, "music": true, "3d": true}

	for _, verb := range callVerbs(t) {
		// `call models` sits here because it is answered by the data plane
		// over the same credential, but it sends no work and so has nothing
		// to fall back to.
		if verb.Name() == "models" {
			continue
		}
		// `call task` submits nothing either: it follows work another verb
		// already submitted, and a task exists only on the backend that
		// accepted it. Its subcommands need the model the submission
		// resolved, so falling back to a category would send them at
		// whichever backend the category happens to name today — a different
		// one is the ordinary case, and it answers 404 for an id it never saw.
		if verb.Name() == "task" {
			assertTaskSubcommandsRequireAModel(t, verb)
			continue
		}
		flag := verb.Flags().Lookup("model")
		// The other exceptions do take a --model, and Router resolves no
		// default for any of them. So the flag has to be there, has to say it
		// is required, and must not promise a category — naming one would send
		// callers at a route that does not exist, which arrives as "no such
		// model" rather than as the absence it is.
		//
		// Responses is provider-model-only, and routing/category_test.go
		// asserts that absence deliberately so it cannot be quietly reversed.
		// Music and 3D are the newer pair: Router's registry says a category
		// waits for a second implementation, since with one apiece a default
		// would name that one thing while reading like a choice.
		if requiresAModel[verb.Name()] {
			switch {
			case flag == nil:
				t.Errorf("call %s: no --model flag, but it cannot fall back to anything",
					verb.Name())
			case strings.Contains(flag.Usage, "default-"):
				t.Errorf("call %s: --model names a category, but Router resolves no "+
					"default for this mode: %q", verb.Name(), flag.Usage)
			case !strings.Contains(flag.Usage, "required"):
				t.Errorf("call %s: --model does not say it is required: %q",
					verb.Name(), flag.Usage)
			}
			continue
		}
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

func assertTaskSubcommandsRequireAModel(t *testing.T, task *cobra.Command) {
	t.Helper()
	subs := task.Commands()
	if len(subs) == 0 {
		t.Fatal("call task: no subcommands")
	}
	for _, sub := range subs {
		flag := sub.Flags().Lookup("model")
		if flag == nil {
			t.Errorf("call task %s: no --model flag, but a task id alone does not "+
				"say which backend holds it", sub.Name())
			continue
		}
		if strings.Contains(flag.Usage, "default-") {
			t.Errorf("call task %s: --model names a category, which would resolve a "+
				"backend that never saw this task: %q", sub.Name(), flag.Usage)
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
//
// Every category constant is listed. A constant left out is the one nobody
// checks, and sound effects were exactly that: reachable only through
// --sound-fx, so a typo in it would have surfaced as a 404 from a flag rather
// than as a missing route.
func TestEveryCategoryIsSpelledLikeADefaultRoute(t *testing.T) {
	categories := map[string]string{
		"chat":          categoryChat,
		"embedding":     categoryEmbedding,
		"rerank":        categoryRerank,
		"search":        categorySearch,
		"scrape":        categoryScrape,
		"image":         categoryImage,
		"video":         categoryVideo,
		"ocr":           categoryOCR,
		"stt":           categorySTT,
		"stt_stream":    categorySTTStream,
		"align":         categoryAlign,
		"tts":           categoryTTS,
		"tts_clone":     categoryTTSClone,
		"dialogue":      categoryTTSDialogue,
		"vad":           categoryVAD,
		"diarization":   categoryDiarization,
		"diar_stream":   categoryDiarStream,
		"speaker_embed": categorySpeakerEmbed,
		"enhance":       categoryEnhance,
		"sound_fx":      categorySoundFX,
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
