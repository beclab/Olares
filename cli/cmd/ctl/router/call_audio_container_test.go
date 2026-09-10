package router

import (
	"strings"
	"testing"
)

// The report that started this: `call speak … --out say.wav` wrote an mp3 and
// said `wrote say.wav`. Feeding that file to `call listen`, which takes
// headerless PCM, then produced an empty transcript at exit 0 — a silent lie
// that read as a model hearing nothing.
func TestAMisnamedContainerIsReported(t *testing.T) {
	var out strings.Builder
	reportOutNameMismatch(&out, "say.wav", []byte("ID3\x04\x00\x00\x00"))
	if !strings.Contains(out.String(), "mp3") || !strings.Contains(out.String(), "wav") {
		t.Errorf("an mp3 written as .wav is not reported: %q", out.String())
	}
	if !strings.Contains(out.String(), "--response-format") {
		t.Errorf("the report does not say what to change: %q", out.String())
	}
}

func TestAContainerThatMatchesItsNameIsSilent(t *testing.T) {
	cases := []struct {
		name string
		path string
		head []byte
	}{
		{"a wav called .wav", "a.wav", []byte("RIFF$\x94\x02\x00WAVE")},
		{"an mp3 called .mp3", "a.mp3", []byte("ID3\x04\x00\x00\x00")},
		{"a bare mpeg frame called .mp3", "a.mp3", []byte{0xFF, 0xFB, 0x90, 0x00}},
		{"an ogg called .ogg", "a.ogg", []byte("OggS\x00\x02\x00\x00")},
		{"a flac called .flac", "a.flac", []byte("fLaC\x00\x00\x00\x22")},
		// Headerless by design: raw PCM cannot be recognised, and reporting
		// every body this cannot read would train the reader to skip the line.
		{"raw pcm", "a.wav", []byte{0x00, 0x01, 0x00, 0xFF, 0x02, 0x00}},
		{"an extension this does not check", "a.bin", []byte("ID3\x04\x00\x00\x00")},
		{"no --out at all", "", []byte("ID3\x04\x00\x00\x00")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out strings.Builder
			reportOutNameMismatch(&out, tc.path, tc.head)
			if out.Len() != 0 {
				t.Errorf("reported %q", out.String())
			}
		})
	}
}

// The two ways of naming a format cannot both be honoured, so a call that names
// both differently is refused before anything is spent on it.
func TestContradictingOutAndResponseFormatIsRefusedUpFront(t *testing.T) {
	if err := refuseContradictedOutName("say.wav", "mp3_44100_128"); err == nil {
		t.Error("a .wav asked for as mp3 was accepted")
	}
	for _, tc := range []struct{ out, format string }{
		// The engines here spell a sample rate into the value, which is why the
		// container is read off the front rather than compared whole.
		{"say.wav", "wav_16000"},
		{"say.mp3", "mp3_44100_192"},
		{"say.wav", ""},
		{"", "mp3"},
		{"say.bin", "wav_16000"},
	} {
		if err := refuseContradictedOutName(tc.out, tc.format); err != nil {
			t.Errorf("--out %q with --response-format %q was refused: %v", tc.out, tc.format, err)
		}
	}
}

// `pcm` is its own answer rather than a kind of wav: the engines offer
// pcm_24000 beside wav_24000, and the two differ by exactly the header this
// checks for.
func TestPcmIsNotReadAsWav(t *testing.T) {
	if got := containerFromResponseFormat("pcm_24000"); got != "pcm" {
		t.Errorf("containerFromResponseFormat(pcm_24000) = %q", got)
	}
	if err := refuseContradictedOutName("say.wav", "pcm_24000"); err == nil {
		t.Error("headerless pcm asked for under a .wav name was accepted")
	}
}
