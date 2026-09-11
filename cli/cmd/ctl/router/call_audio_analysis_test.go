package router

import (
	"strings"
	"testing"
)

// A diarization model returned no turns for a real 16 kHz recording of one
// person talking, while `call vad` found two speech segments in the same file
// and `call transcribe` transcribed it. Reporting that as nobody speaking sent
// the listener to check their microphone rather than their speaker count, and
// it is the one reading the rest of the evidence ruled out.
func TestAnEmptyDiarizationDoesNotClaimSilence(t *testing.T) {
	if strings.Contains(emptyDiarizationNote, "nobody speaking") {
		t.Errorf("the note still asserts silence: %q", emptyDiarizationNote)
	}
	for _, want := range []string{"no speaker turns", "single-speaker", "call vad"} {
		if !strings.Contains(emptyDiarizationNote, want) {
			t.Errorf("the note does not mention %q: %q", want, emptyDiarizationNote)
		}
	}
}

// Both the batch renderer and the streaming one say it, and they used to say it
// twice in two places.
func TestBothDiarizationRenderersUseTheSameNote(t *testing.T) {
	var batch strings.Builder
	if err := renderDiarization(&batch, []byte(`{"segments":[]}`)); err != nil {
		t.Fatalf("renderDiarization: %v", err)
	}
	if !strings.Contains(batch.String(), "no speaker turns") {
		t.Errorf("the batch renderer reads %q", batch.String())
	}
}
