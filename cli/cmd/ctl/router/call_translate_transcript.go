package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// POST /v1/translate/transcript
//
// The fifth translate route, and the one that is not a text-at-a-time
// translation. A transcript is translated as a stretch: the model reads the
// turns around a line before rendering it, which is the difference between
// "对" arriving as "yes" and as "right" — a two-character turn carries no
// evidence of which, and the previous turn carries all of it.
//
// The answer is one result per turn, under the ids that were sent, never
// shorter than the request. That guarantee is the reason this route exists
// rather than a loop over /v1/translate: a caller matching by position cannot
// silently shift every translation onto the wrong turn, and a turn the model
// failed on comes back carrying its own reason instead of taking the whole
// page down.

// transcriptSegment is one turn to translate. The id is the caller's own name
// for it and comes back unchanged, so it has to be unique within a request.
type transcriptSegment struct {
	ID      string `json:"id"`
	Speaker string `json:"speaker,omitempty"`
	Text    string `json:"text"`
}

// transcriptLine is a turn the model reads and does not translate. It has no
// id on purpose: a line that cannot be named cannot be expected back, which is
// what keeps context out of the answer.
type transcriptLine struct {
	Speaker string `json:"speaker,omitempty"`
	Text    string `json:"text"`
}

type transcriptContext struct {
	Before []transcriptLine `json:"before,omitempty"`
}

type glossaryEntry struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type transcriptRequest struct {
	From       string              `json:"from,omitempty"`
	To         string              `json:"to"`
	Segments   []transcriptSegment `json:"segments"`
	Context    *transcriptContext  `json:"context,omitempty"`
	Background string              `json:"background,omitempty"`
	Glossary   []glossaryEntry     `json:"glossary,omitempty"`
}

// transcriptResult carries exactly one of Text and Error. A turn that failed
// is still a turn, in place, so the rest of the page is usable.
type transcriptResult struct {
	ID    string `json:"id"`
	Text  string `json:"text"`
	Error string `json:"error,omitempty"`
}

type transcriptResponse struct {
	Segments []transcriptResult `json:"segments"`
}

// transcriptSegmentLimit is the upstream's ceiling on one request. Checked
// here so a caller with an hour of dialogue is told to split it, rather than
// having the whole page refused with none of it translated.
const transcriptSegmentLimit = 64

type transcriptOptions struct {
	Path       string
	To         string
	From       string
	Background string
	Glossary   []string
	APIKey     string
	OutputIn   string
}

func runCallTranscript(ctx context.Context, f *cmdutil.Factory, opts transcriptOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(opts.OutputIn)
	if err != nil {
		return err
	}
	req, err := readTranscript(opts.Path)
	if err != nil {
		return err
	}
	glossary, err := parseGlossary(opts.Glossary)
	if err != nil {
		return err
	}
	req.To = opts.To
	if opts.From != "" {
		req.From = opts.From
	}
	if opts.Background != "" {
		req.Background = opts.Background
	}
	if len(glossary) > 0 {
		req.Glossary = glossary
	}
	if err := checkTranscript(req); err != nil {
		return err
	}
	pc, err := prepareLongRequest(ctx, f)
	if err != nil {
		return err
	}
	var resp transcriptResponse
	if err := dataPlane(pc, opts.APIKey).doJSON(ctx, "POST", epTranslateTranscript, req, &resp); err != nil {
		return callErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, resp)
	}
	return renderTranscript(os.Stdout, req.Segments, resp.Segments)
}

// readTranscript accepts the two shapes a transcript arrives in. A file
// written for this route is the request itself; a file that fell out of a
// diarizer or a meeting tool is a bare array of turns. Both are read, because
// requiring the wrapper would mean asking callers to edit a machine-produced
// file before it can be translated.
func readTranscript(path string) (*transcriptRequest, error) {
	var raw []byte
	var err error
	if strings.TrimSpace(path) == "-" {
		raw, err = io.ReadAll(os.Stdin)
	} else {
		raw, err = os.ReadFile(path)
	}
	if err != nil {
		return nil, fmt.Errorf("read the transcript: %w", err)
	}
	trimmed := strings.TrimLeft(string(raw), " \t\r\n")
	if strings.HasPrefix(trimmed, "[") {
		var segments []transcriptSegment
		if err := json.Unmarshal(raw, &segments); err != nil {
			return nil, fmt.Errorf("the transcript is an array but not an array of turns: %w", err)
		}
		return &transcriptRequest{Segments: segments}, nil
	}
	var req transcriptRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, fmt.Errorf("the transcript is not JSON this route can read: %w. "+
			"It is either an object with a \"segments\" array, or the array itself", err)
	}
	return &req, nil
}

// checkTranscript refuses locally what the upstream would refuse anyway, and
// says which turn is at fault. A duplicate id is the one worth catching here:
// upstream it is a 400 about the request, and here it is a line number.
func checkTranscript(req *transcriptRequest) error {
	if len(req.Segments) == 0 {
		return fmt.Errorf("the transcript has no turns to translate")
	}
	if len(req.Segments) > transcriptSegmentLimit {
		return fmt.Errorf("%d turns is more than the %d one request takes; split the transcript "+
			"and carry the tail of each part into the next one's \"context\"",
			len(req.Segments), transcriptSegmentLimit)
	}
	seen := make(map[string]int, len(req.Segments))
	for i, s := range req.Segments {
		if strings.TrimSpace(s.ID) == "" {
			return fmt.Errorf("turn %d has no id; the answer comes back under these ids, "+
				"so a turn without one cannot be matched", i+1)
		}
		if first, dup := seen[s.ID]; dup {
			return fmt.Errorf("turns %d and %d share the id %q; the answer is keyed by id, "+
				"so two turns cannot carry the same one", first+1, i+1, s.ID)
		}
		seen[s.ID] = i
	}
	return nil
}

func parseGlossary(pairs []string) ([]glossaryEntry, error) {
	out := make([]glossaryEntry, 0, len(pairs))
	for _, p := range pairs {
		source, target, ok := strings.Cut(p, "=")
		if !ok || strings.TrimSpace(source) == "" || strings.TrimSpace(target) == "" {
			return nil, fmt.Errorf("--term takes source=target, and %q is not that", p)
		}
		out = append(out, glossaryEntry{
			Source: strings.TrimSpace(source),
			Target: strings.TrimSpace(target),
		})
	}
	return out, nil
}

// renderTranscript prints the answer beside the turns it answers, because the
// point of this route is the alignment and a bare list of translations shows
// none of it. A turn the model failed on is printed as its reason rather than
// skipped: a gap in a transcript is worse than a marked hole.
func renderTranscript(w io.Writer, sent []transcriptSegment, got []transcriptResult) error {
	if len(got) == 0 {
		_, err := fmt.Fprintln(w, "the model answered with no turns, which this route does not do; "+
			"re-run with -o json to see what came back.")
		return err
	}
	speakers := make(map[string]string, len(sent))
	for _, s := range sent {
		speakers[s.ID] = s.Speaker
	}
	failed := 0
	t := newTable(w, "TURN", "SPEAKER", "TRANSLATION")
	for _, r := range got {
		text := r.Text
		if r.Error != "" {
			failed++
			text = "(failed: " + r.Error + ")"
		}
		t.row(r.ID, nonEmpty(speakers[r.ID]), text)
	}
	if err := t.flush(); err != nil {
		return err
	}
	if failed > 0 {
		_, err := fmt.Fprintf(w, "\n%d of %d turns failed and are marked above. The rest are "+
			"translated and in place; re-send only the failed ids.\n", failed, len(got))
		return err
	}
	return nil
}
