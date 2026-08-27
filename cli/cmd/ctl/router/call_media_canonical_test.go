package router

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// mediaCommand is a verb's flag surface without the verb: the fields this family
// admits, registered on a command the test can parse a line with. Presence is
// read off the flag set, so nothing here can be tested by setting struct fields.
func mediaCommand(t *testing.T, names []string, line ...string) (*cobra.Command, *mediaFlags) {
	t.Helper()
	flags := &mediaFlags{}
	cmd := &cobra.Command{Use: "media"}
	flags.register(cmd, names...)
	if err := cmd.Flags().Parse(line); err != nil {
		t.Fatalf("parse %v: %v", line, err)
	}
	return cmd, flags
}

var (
	imageFields = []string{
		flagNegative, flagN, flagSeed, flagFormat, flagSize, flagAspectRatio,
		flagQuality, flagImage, flagMask, flagSource, flagOperation, flagProviderOption,
	}
	videoFields = []string{
		flagNegative, flagN, flagSeed, flagFormat, flagSize, flagAspectRatio,
		flagResolution, flagDuration, flagFPS, flagQuality, flagImage, flagMask,
		flagAudioIn, flagSource, flagOperation, flagProviderOption,
	}
	musicFields = []string{
		flagNegative, flagN, flagSeed, flagFormat, flagDuration, flagLyrics,
		flagInstrumental, flagProviderOption,
	}
	model3DFields = []string{
		flagNegative, flagN, flagSeed, flagFormats, flagTexture, flagPBR,
		flagPolycount, flagImage, flagProviderOption,
	}
)

// The whole point of building the body from the flag set rather than from the
// values: /v1/generations refuses a field the resolved provider has no
// parameter for, so a field nobody asked for must not appear. Sending
// `"n": 0` because a flag defaults to zero is asking for something.
func TestCanonicalSendsOnlyWhatWasAskedFor(t *testing.T) {
	cmd, flags := mediaCommand(t, musicFields)
	request, err := flags.canonical(cmd, "FlowStudio/track", "a slow waltz")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	encoded, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &keys); err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := map[string]bool{"model": true, "prompt": true}
	for key := range keys {
		if !want[key] {
			t.Errorf("%q was sent without being asked for: %s", key, encoded)
		}
	}
	if len(keys) != len(want) {
		t.Errorf("body carries %d fields, want %d: %s", len(keys), len(want), encoded)
	}
}

// A zero a caller wrote is a request, and it has to survive. seed 0 is a
// reproducible seed and instrumental=false is a track with vocals; both would
// disappear if presence were read from the value.
func TestCanonicalKeepsAZeroTheCallerAskedFor(t *testing.T) {
	cmd, flags := mediaCommand(t, musicFields, "--seed", "0", "--instrumental=false")
	request, err := flags.canonical(cmd, "FlowStudio/track", "a waltz")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if request.Seed == nil || *request.Seed != 0 {
		t.Errorf("seed: got %v, want 0", request.Seed)
	}
	if request.Music == nil || request.Music.Instrumental == nil || *request.Music.Instrumental {
		t.Errorf("music: got %+v, want instrumental=false", request.Music)
	}
}

func TestCanonicalNestsEachFamilysFieldsWhereRouterReadsThem(t *testing.T) {
	cmd, flags := mediaCommand(t, model3DFields,
		"--formats", "glb,obj", "--texture", "--pbr=false", "--polycount", "30000",
		"--image", "data:image/png;base64,AA==", "--n", "1",
		"--provider-option", "steps=8", "--provider-option", "sampler=euler",
	)
	request, err := flags.canonical(cmd, "FlowStudio/mesh", "a lantern")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	encoded, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	const want = `{"model":"FlowStudio/mesh","prompt":"a lantern","n":1,` +
		`"inputs":{"images":["data:image/png;base64,AA=="]},` +
		`"output":{"formats":["glb","obj"],"texture":true,"pbr":false,"target_polycount":30000},` +
		`"provider_options":{"sampler":"euler","steps":8}}`
	if string(encoded) != want {
		t.Errorf("body:\n got %s\nwant %s", encoded, want)
	}
}

// The released routes have their own names for the same fields, and Router lifts
// each onto the canonical field it means. Getting one of these spellings wrong
// does not fail the request: the key becomes a vendor option and is forwarded to
// a provider that has never heard of it.
func TestLegacyBodyUsesTheKeysTheReleasedRoutesLift(t *testing.T) {
	cmd, flags := mediaCommand(t, videoFields,
		"--negative", "blurry", "--seed", "7", "--resolution", "1080p",
		"--duration", "5", "--fps", "24", "--format", "mp4", "--quality", "high",
		"--operation", "lip_sync", "--audio", "https://example.test/a.wav",
		"--image", "https://example.test/frame.png", "--source-generation", "gen_01",
		"--provider-option", "guidance_scale=3.5",
	)
	body, err := flags.legacyBody(cmd, "OpenAI/sora-2", "waves")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	want := map[string]any{
		"model": "OpenAI/sora-2", "prompt": "waves", "operation": "lip_sync",
		"negative_prompt": "blurry", "seed": int64(7), "resolution": "1080p",
		"duration_seconds": 5.0, "fps": 24, "output_format": "mp4", "quality": "high",
		"audio": "https://example.test/a.wav", "source_generation_id": "gen_01",
		"reference_images": []string{"https://example.test/frame.png"},
		"guidance_scale":   json.RawMessage("3.5"),
	}
	if !reflect.DeepEqual(body, want) {
		t.Errorf("body:\n got %#v\nwant %#v", body, want)
	}
}

// Router refuses pixels alongside a shape described another way, and the answer
// does not depend on the provider — so the round trip is the error message.
func TestPixelsAndAShapeAreRefusedBeforeTheRequest(t *testing.T) {
	for _, line := range [][]string{
		{"--size", "1280x720", "--aspect-ratio", "16:9"},
		{"--size", "1280x720", "--resolution", "1080p"},
	} {
		cmd, flags := mediaCommand(t, videoFields, line...)
		if _, err := flags.canonical(cmd, "m", "p"); err == nil {
			t.Errorf("%v was accepted", line)
		} else if !strings.Contains(err.Error(), "same output shape") {
			t.Errorf("%v: %v", line, err)
		}
		if _, err := flags.legacyBody(cmd, "m", "p"); err == nil {
			t.Errorf("%v was accepted on the released route", line)
		}
	}
	// A ratio and a named resolution together are one shape, and it is how
	// several providers are asked. Only pixels are exclusive.
	cmd, flags := mediaCommand(t, videoFields, "--aspect-ratio", "16:9", "--resolution", "1080p")
	if _, err := flags.canonical(cmd, "m", "p"); err != nil {
		t.Errorf("a ratio at a resolution was refused: %v", err)
	}
}

func TestAMeasurementHasToBePositive(t *testing.T) {
	for _, line := range [][]string{
		{"--n", "0"}, {"--duration", "0"}, {"--fps", "-1"},
	} {
		cmd, flags := mediaCommand(t, videoFields, line...)
		if _, err := flags.canonical(cmd, "m", "p"); err == nil {
			t.Errorf("%v was accepted", line)
		}
	}
}

// A path is meaningful on this machine and nowhere else. It is also the only
// form a person can type: the alternative is a 10 MiB data URL on a command
// line.
func TestALocalFileBecomesADataURLAndAnythingElseIsLeftAlone(t *testing.T) {
	path := filepath.Join(t.TempDir(), "frame.png")
	if err := os.WriteFile(path, []byte("\x89PNG\r\n\x1a\n"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	cmd, flags := mediaCommand(t, imageFields, "--image", path, "--operation", "edit")
	request, err := flags.canonical(cmd, "FlowStudio/image", "brighten it")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if request.Inputs == nil || len(request.Inputs.Images) != 1 {
		t.Fatalf("inputs: %+v", request.Inputs)
	}
	if got := request.Inputs.Images[0]; !strings.HasPrefix(got, "data:image/png;base64,") {
		t.Errorf("a local file was not encoded: %q", got)
	}

	for _, passed := range []string{"data:image/png;base64,AA==", "https://example.test/a.png"} {
		cmd, flags := mediaCommand(t, imageFields, "--image", passed, "--operation", "edit")
		request, err := flags.canonical(cmd, "m", "p")
		if err != nil {
			t.Fatalf("%s: %v", passed, err)
		}
		if request.Inputs.Images[0] != passed {
			t.Errorf("%s was rewritten to %q", passed, request.Inputs.Images[0])
		}
	}

	cmd, flags = mediaCommand(t, imageFields, "--image", "/nonexistent/file.png", "--operation", "edit")
	if _, err := flags.canonical(cmd, "m", "p"); err == nil {
		t.Error("a path that does not exist was accepted")
	}
}

func TestProviderOptionsCarryJSONAsJSONAndTheRestAsStrings(t *testing.T) {
	cmd, flags := mediaCommand(t, imageFields,
		"--provider-option", "steps=8",
		"--provider-option", "style=vivid",
		"--provider-option", `crop={"x":1}`,
		"--provider-option", "sizes=2K",
	)
	request, err := flags.canonical(cmd, "m", "p")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	want := map[string]string{
		"steps": `8`, "style": `"vivid"`, "crop": `{"x":1}`, "sizes": `"2K"`,
	}
	for key, value := range want {
		if got := string(request.ProviderOptions[key]); got != value {
			t.Errorf("%s: got %s, want %s", key, got, value)
		}
	}
	cmd, flags = mediaCommand(t, imageFields, "--provider-option", "novalue")
	if _, err := flags.canonical(cmd, "m", "p"); err == nil {
		t.Error("a pair with no key=value shape was accepted")
	}
}

// A family's flag list is Router's table, not a superset of it. A flag that
// could only ever be refused reads as a promise, so the registration is the
// place where a mistake has to be loud.
func TestRegisteringAFlagThisTreeHasNoDefinitionForPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("an unknown flag name was registered silently")
		}
	}()
	flags := &mediaFlags{}
	flags.register(&cobra.Command{Use: "media"}, "no-such-flag")
}
