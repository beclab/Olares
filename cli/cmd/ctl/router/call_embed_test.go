package router

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// A picture is one input object rather than one of the strings, because that is
// the extension Router publishes and the engine reads `type` to know which of a
// CLIP model's two towers is being asked for.
func TestAnImageIsSentAsTheInputObjectRouterExtendsTheBodyWith(t *testing.T) {
	path := filepath.Join(t.TempDir(), "shot.png")
	if err := os.WriteFile(path, []byte("\x89PNG\r\n\x1a\nnot really"), 0o600); err != nil {
		t.Fatal(err)
	}
	picture, err := embedImage(path)
	if err != nil {
		t.Fatalf("a readable picture was refused: %v", err)
	}
	if picture.Type != "image" || !strings.HasPrefix(picture.ImageURL, "data:image/png;base64,") {
		t.Fatalf("the picture did not become a data URL: %+v", picture)
	}
	body, err := json.Marshal(embeddingsRequest{Model: "m", Input: picture})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"input":{"type":"image","image_url":"data:image/png;base64,`) {
		t.Fatalf("the wire shape is not the one Router documents: %s", body)
	}

	// Text keeps the shape it always had: the field widened, the body did not.
	body, err = json.Marshal(embeddingsRequest{Model: "m", Input: []string{"one", "two"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"input":["one","two"]`) {
		t.Fatalf("text input changed shape: %s", body)
	}
}

// A data URL or a link is what the caller already chose, so it travels as
// written. Rewriting it would be this tree guessing at a form the engine may
// well prefer.
func TestAnImageAlreadyInASendableFormIsNotReEncoded(t *testing.T) {
	for _, value := range []string{"data:image/webp;base64,AAAA", "https://example.test/a.png"} {
		picture, err := embedImage(value)
		if err != nil {
			t.Fatalf("%s: %v", value, err)
		}
		if picture.ImageURL != value {
			t.Errorf("%s was rewritten to %s", value, picture.ImageURL)
		}
	}
	if _, err := embedImage("   "); err == nil {
		t.Error("--image was accepted with nothing in it")
	}
}

// The budget is the request's, not the file's: base64 adds a third, so the file
// that fits is smaller than the limit. Checking the file would admit one that
// the application then refuses, and the refusal would arrive after the upload.
func TestAPictureIsMeasuredAfterEncodingAndRefusedBeforeTheFactory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "large.png")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	// Three quarters of the allowance encodes to just over it.
	if err := os.Truncate(path, int64(embedImageMaxBytes)*3/4+1024); err != nil {
		t.Fatal(err)
	}
	err := runCallEmbed(context.Background(), nil, nil, path, categoryEmbedding, nil, "", "table")
	if err == nil {
		t.Fatal("an oversize picture reached the nil Factory")
	}
	if !strings.Contains(err.Error(), strconv.Itoa(embedImageMaxBytes)) ||
		!strings.Contains(err.Error(), "base64") {
		t.Fatalf("the refusal does not say what the limit is or why: %v", err)
	}
}

// One call carries one thing. The engine embeds a picture or the text, and a
// caller who passed both meant two calls.
func TestTextAndAPictureCannotShareOneCall(t *testing.T) {
	for _, args := range [][]string{
		{"--image", "shot.png", "some text"},
		{"--image", "shot.png", "--per-line"},
	} {
		cmd := newCallEmbedCommand(nil)
		cmd.SetArgs(args)
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SilenceUsage = true
		err := cmd.Execute()
		if err == nil || !strings.Contains(err.Error(), "embeds one picture") {
			t.Errorf("%v: got %v", args, err)
		}
	}
}
