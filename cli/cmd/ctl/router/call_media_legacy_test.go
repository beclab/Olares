package router

import (
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// The two released verbs offer their family's row and nothing else. An image has
// no operation because both routes run an image as generate; without an
// operation there is nothing for a reference image or a mask to belong to, so
// those go with it. A flag that could only ever be refused reads as a promise.
func TestImageAndVideoOfferTheirOwnFields(t *testing.T) {
	cases := map[string]struct{ has, lacks []string }{
		"image": {
			has: []string{flagSize, flagAspectRatio, flagQuality, flagFormat, flagSeed, flagNegative},
			lacks: []string{
				flagOperation, flagImage, flagMask, flagSource, flagN,
				flagResolution, flagDuration, flagFPS, flagLyrics, flagFormats,
			},
		},
		"video": {
			has: []string{
				flagOperation, flagResolution, flagDuration, flagFPS, flagAspectRatio,
				flagImage, flagMask, flagAudioIn, flagSource, flagN,
			},
			lacks: []string{flagLyrics, flagInstrumental, flagTexture, flagPBR, flagPolycount, flagFormats},
		},
	}
	for name, want := range cases {
		verb := mediaVerb(t, name)
		for _, flag := range want.has {
			if verb.Flags().Lookup(flag) == nil {
				t.Errorf("call %s has no --%s", name, flag)
			}
		}
		for _, flag := range want.lacks {
			if verb.Flags().Lookup(flag) != nil {
				t.Errorf("call %s offers --%s, which no route would honor for it", name, flag)
			}
		}
	}
}

// Image and video stay on the routes they were released on, and that is a
// decision rather than an omission: Router lifts every key those routes take
// onto the canonical field it means, and the image route additionally answers
// synchronously for a provider that keeps no generations. Submitting an image on
// /v1/generations would refuse the most widely installed provider there is.
func TestImageAndVideoStayOnTheirReleasedRoutes(t *testing.T) {
	if imageKind.submitPath != epImageGenerations || videoKind.submitPath != epVideos {
		t.Errorf("image submits to %s and video to %s", imageKind.submitPath, videoKind.submitPath)
	}
	for _, kind := range []mediaKind{imageKind, videoKind} {
		if kind.submitPath == epGenerations {
			t.Errorf("%s moved to the unified route, which has no synchronous answer", kind.verb)
		}
	}
}

// Every flag these verbs grew has to reach the body. The keys are the released
// spellings, which Router lifts; a key it does not lift is forwarded to the
// provider as a vendor option instead, so a misspelling here is silent.
func TestTheReleasedVerbsSendEveryFlagTheyOffer(t *testing.T) {
	cmd, flags := mediaCommand(t, imageFields,
		"--negative", "blurry", "--seed", "7", "--size", "1024x1024",
		"--quality", "high", "--format", "png",
	)
	body, err := flags.legacyBody(cmd, "OpenAI/gpt-image-1", "a bicycle")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	for key, want := range map[string]any{
		"model": "OpenAI/gpt-image-1", "prompt": "a bicycle", "negative_prompt": "blurry",
		"seed": int64(7), "size": "1024x1024", "quality": "high", "output_format": "png",
	} {
		if body[key] != want {
			t.Errorf("%s: got %v, want %v", key, body[key], want)
		}
	}
	if len(body) != 7 {
		t.Errorf("body carries %d keys: %v", len(body), body)
	}
}

// A video operation can be the whole request: extending a clip or syncing a
// mouth to a recording describes what to do without any words. An image cannot,
// so it still asks for a prompt — and a video with neither a prompt nor an input
// has no subject at all.
func TestAVideoOperationCanStandWithoutAPrompt(t *testing.T) {
	cmd, flags := mediaCommand(t, videoFields,
		"--operation", "extend", "--source-generation", "gen_01",
	)
	body, err := flags.legacyBody(cmd, "FlowStudio/wan2.2", "")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if _, carries := body["prompt"]; carries {
		t.Errorf("an empty prompt was sent: %v", body)
	}
	if !flags.namesAnInput(cmd) {
		t.Error("--source-generation was not read as something to work from")
	}

	factory := &cmdutil.Factory{}
	err = runVerb(t, newCallVideoCommand(factory), "--model", "FlowStudio/wan2.2")
	if err == nil || !strings.Contains(err.Error(), "nothing to work from") {
		t.Errorf("a video with neither a prompt nor an input: %v", err)
	}
}
