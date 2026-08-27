package router

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func mediaVerb(t *testing.T, name string) *cobra.Command {
	t.Helper()
	for _, verb := range NewCallCommand(&cmdutil.Factory{}).Commands() {
		if verb.Name() == name {
			return verb
		}
	}
	t.Fatalf("call %s is not registered", name)
	return nil
}

// runVerb executes a verb on its own, away from the tree. A command with a
// parent takes its arguments from the process rather than from SetArgs, which in
// a test binary means the test runner's own flags.
//
// Stdin is emptied for the duration, because a prompt can be piped in: left
// alone, a verb reading it would take whatever the runner was started with, and
// the answer would depend on whose terminal the suite is running under.
func runVerb(t *testing.T, cmd *cobra.Command, args ...string) error {
	t.Helper()
	empty, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	defer func() { _ = empty.Close() }()
	saved := os.Stdin
	os.Stdin = empty
	defer func() { os.Stdin = saved }()

	cmd.SetArgs(args)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	return cmd.Execute()
}

// Router has no default category for either family, so a call with no --model
// has to fail here with a reason. Sent as it stands it would resolve nothing and
// come back as a model that does not exist, which points the reader at their own
// spelling instead of at an absence.
func TestMusicAnd3DRefuseToRunWithoutAModel(t *testing.T) {
	factory := &cmdutil.Factory{}
	for name, verb := range map[string]*cobra.Command{
		"music": newCallMusicCommand(factory), "3d": newCall3DCommand(factory),
	} {
		err := runVerb(t, verb, "a lantern")
		if err == nil {
			t.Errorf("call %s ran with no model", name)
			continue
		}
		if !strings.Contains(err.Error(), "--model is required") ||
			!strings.Contains(err.Error(), "model list --mode") {
			t.Errorf("call %s: %v", name, err)
		}
	}
}

// Only the fields the family admits. A flag Router would refuse for this family
// reads as a promise, and the reader finds out it was not one after paying for
// the round trip.
func TestMusicAnd3DOfferOnlyTheirOwnFields(t *testing.T) {
	cases := map[string]struct{ has, lacks []string }{
		"music": {
			has:   []string{flagLyrics, flagInstrumental, flagDuration, flagFormat, flagSeed},
			lacks: []string{flagSize, flagAspectRatio, flagResolution, flagFPS, flagImage, flagOperation},
		},
		"3d": {
			has:   []string{flagFormats, flagTexture, flagPBR, flagPolycount, flagImage},
			lacks: []string{flagSize, flagDuration, flagFPS, flagFormat, flagLyrics, flagOperation},
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
				t.Errorf("call %s offers --%s, which this family cannot express", name, flag)
			}
		}
	}
}

// A mesh is usually asked for with a picture rather than with words, and 3D is
// the one family that admits an input on a plain generate. So a prompt is
// optional there and required everywhere else — but one of the two has to be
// present, or there is nothing to generate from.
func TestAMeshCanBeAskedForWithAPictureAlone(t *testing.T) {
	cmd, flags := mediaCommand(t, model3DFields, "--image", "data:image/png;base64,AA==")
	request, err := flags.canonical(cmd, "FlowStudio/mesh", "")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if request.Prompt != "" || request.Inputs == nil || len(request.Inputs.Images) != 1 {
		t.Errorf("request: %+v", request)
	}

	err = runVerb(t, newCall3DCommand(&cmdutil.Factory{}), "--model", "FlowStudio/mesh")
	if err == nil || !strings.Contains(err.Error(), "nothing to work from") {
		t.Errorf("a mesh with neither a prompt nor an image: %v", err)
	}
}

// Both verbs create on the unified route and read the record back from it. A
// generation created there is not addressable on the released routes, so a
// mismatch here is a submission that succeeds and an id that cannot be
// collected.
func TestMusicAnd3DUseTheUnifiedRoute(t *testing.T) {
	for _, kind := range []mediaKind{musicKind, model3DKind} {
		if kind.submitPath != epGenerations {
			t.Errorf("%s submits to %s", kind.verb, kind.submitPath)
		}
		if got := kind.get("gen_1"); got != epGeneration("gen_1") {
			t.Errorf("%s reads from %s", kind.verb, got)
		}
		if got := kind.content("gen_1"); got != epGenerationContent("gen_1") {
			t.Errorf("%s downloads from %s", kind.verb, got)
		}
	}
}

// The canonical body is what /v1/generations parses, and it parses strictly: a
// field under the wrong name is a 400 rather than a field that is ignored.
func TestATrackIsSpelledTheWayRouterReadsIt(t *testing.T) {
	cmd, flags := mediaCommand(t, musicFields,
		"--duration", "30", "--format", "mp3", "--lyrics", "la la", "--instrumental=false",
	)
	request, err := flags.canonical(cmd, "FlowStudio/ace-step", "a waltz")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	encoded, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	const want = `{"model":"FlowStudio/ace-step","prompt":"a waltz",` +
		`"output":{"format":"mp3","duration_seconds":30},` +
		`"music":{"lyrics":"la la","instrumental":false}}`
	if string(encoded) != want {
		t.Errorf("body:\n got %s\nwant %s", encoded, want)
	}
}

// A track and a mesh arrive as their own content types, and a file named
// otherwise is a file that opens wrong.
func TestAFileIsNamedAfterWhatArrived(t *testing.T) {
	cases := map[string]string{
		"audio/mpeg":         ".mp3",
		"audio/wav":          ".wav",
		"model/gltf-binary":  ".glb",
		"application/zip":    ".zip",
		"video/mp4":          ".mp4",
		"application/x-what": ".bin",
	}
	for contentType, want := range cases {
		if got := extForContentType(contentType, ".bin"); got != want {
			t.Errorf("%s: got %q want %q", contentType, got, want)
		}
	}
}
