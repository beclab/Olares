package router

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// What a file is called and what is in it are separate facts, and the audio
// verbs used to let them disagree in silence. `router call speak … --out
// say.wav` wrote an mp3 and reported `wrote say.wav`: the extension takes no
// part in choosing the format, and the engines here answer mp3 unless asked
// otherwise. The lie then travelled — feeding that say.wav to `router call
// listen`, which takes headerless PCM, produced an empty transcript, a duration
// of half the recording, and an exit status of zero.
//
// Two checks close it. Before the call, an `--out` name that contradicts an
// explicit `--response-format` is refused, which costs nothing and is
// unambiguous. After it, the bytes are compared with the name and the
// difference is reported — but the file is kept, because the model has already
// done the work and been billed for it, and deleting the only copy to make a
// point about its name would be the worse failure.
//
// Deriving the format from the extension is deliberately not done. The
// vocabulary is the engine's, not a container list: the TTS applications here
// take `wav_16000` and reject a bare `wav`, so guessing would break commands
// that work today.

// audioHeaderBytes is enough to recognise every container below. `RIFF....WAVE`
// is the longest at twelve.
const audioHeaderBytes = 12

// audioRespFormatFlagUsage used to read "container format, e.g. mp3 or wav",
// which named two values every engine here refuses: they take a sample rate in
// the same string. The help now says the shape and leaves the list to the
// engine, which sends its own on a wrong value.
const audioRespFormatFlagUsage = "output format as the engine names it, e.g. wav_16000 or " +
	"mp3_44100_128; a bare wav is refused, and --out does not set this"

// namesFormatChoices reads a body Router passed through unchanged and reports
// whether it is an engine refusing the format it was asked for, listing the ones
// it takes. Router adds no envelope to an upstream refusal, so this is the body
// itself rather than a code to switch on — which is why the refusal arrived
// unexplained for as long as it did.
func namesFormatChoices(body []byte) bool {
	s := string(body)
	if !strings.Contains(s, "output_format") && !strings.Contains(s, "response_format") {
		return false
	}
	return strings.Contains(s, "must be one of")
}

// hintFormatRefusal explains the shape of the values rather than repeating them:
// the engine already sent its list, and it is printed directly above this line.
func hintFormatRefusal(err error) error {
	return fmt.Errorf("%w\nThese engines name a format by container and sample rate together, so a bare "+
		"`wav` or `mp3` is not one of them — pass one of the values listed above. `--out` does not "+
		"choose the format either; it only names the file the bytes go into", err)
}

// containerOf names the container a body is in, or "" when it is headerless or
// unrecognised. Raw PCM is the headerless case and is not detectable by design.
func containerOf(head []byte) string {
	switch {
	case len(head) >= 12 && string(head[:4]) == "RIFF" && string(head[8:12]) == "WAVE":
		return "wav"
	case len(head) >= 3 && string(head[:3]) == "ID3":
		return "mp3"
	case len(head) >= 2 && head[0] == 0xFF && head[1]&0xE0 == 0xE0:
		return "mp3"
	case len(head) >= 4 && string(head[:4]) == "OggS":
		return "ogg"
	case len(head) >= 4 && string(head[:4]) == "fLaC":
		return "flac"
	}
	return ""
}

// containerFromName reads the container an output path claims, or "" when the
// extension names nothing this can check.
func containerFromName(path string) string {
	switch strings.ToLower(strings.TrimPrefix(filepath.Ext(path), ".")) {
	case "wav", "wave":
		return "wav"
	case "mp3":
		return "mp3"
	case "ogg", "oga", "opus":
		return "ogg"
	case "flac":
		return "flac"
	}
	return ""
}

// containerFromResponseFormat reads the container out of a --response-format
// value. The engines spell a sample rate into it — `wav_16000`, `mp3_44100_128`
// — so the container is the part before the first underscore.
func containerFromResponseFormat(value string) string {
	v := strings.ToLower(strings.TrimSpace(value))
	if v == "" {
		return ""
	}
	if i := strings.IndexAny(v, "_-"); i > 0 {
		v = v[:i]
	}
	switch v {
	case "wav", "wave", "pcm":
		if v == "pcm" {
			return "pcm"
		}
		return "wav"
	case "mp3":
		return "mp3"
	case "ogg", "opus":
		return "ogg"
	case "flac":
		return "flac"
	}
	return ""
}

// refuseContradictedOutName stops a call whose two ways of naming a format
// disagree. Free to check and impossible to satisfy: whichever the engine
// honours, one of the two was wrong.
func refuseContradictedOutName(out, respFormat string) error {
	named := containerFromName(out)
	asked := containerFromResponseFormat(respFormat)
	if named == "" || asked == "" || named == asked {
		return nil
	}
	return fmt.Errorf("--out names a %s file and --response-format asks for %s; "+
		"one of the two has to give", named, asked)
}

// reportOutNameMismatch says so when the bytes are not what the name claims.
// Silence otherwise, including for a container it cannot recognise: raw PCM has
// no header, and reporting every unrecognised body would train the reader to
// ignore the line that matters.
func reportOutNameMismatch(w io.Writer, path string, head []byte) {
	named := containerFromName(path)
	actual := containerOf(head)
	if named == "" || actual == "" || named == actual {
		return
	}
	fmt.Fprintf(w, "warning: those bytes are %s, not the %s that name implies. The extension does not "+
		"choose the format — --response-format does, and these engines spell a sample rate into the "+
		"value, as in %s_16000.\n", actual, named, named)
}
