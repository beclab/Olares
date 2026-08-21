package router

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// POST /v1/translate, POST /v1/translate/batch, GET /v1/languages, POST /v1/detect
//
// The one modality with no model field. These four routes carry no `model` at
// all: Router resolves the translate default per call, and there is nothing for
// a caller to name. So this verb has no --model, which is a deliberate absence
// rather than an omission — a flag whose value the request cannot carry would be
// a flag that silently does nothing.
//
// Which pairs a model serves is a property of the model, not of translation, so
// --languages is the first thing to run against an unfamiliar deployment: a pair
// the model was not built for is refused rather than approximated through
// English.

type translateRequest struct {
	// From is optional: left out, the model detects it. Naming it is worth
	// doing on short text, where detection is least reliable.
	From string `json:"from,omitempty"`
	To   string `json:"to"`
	Text string `json:"text"`
	HTML bool   `json:"html,omitempty"`
}

type translateResponse struct {
	Result string `json:"result"`
}

type translateBatchRequest struct {
	From  string   `json:"from,omitempty"`
	To    string   `json:"to"`
	Texts []string `json:"texts"`
	HTML  bool     `json:"html,omitempty"`
}

type translateBatchResponse struct {
	Results []string `json:"results"`
}

type languagesResponse struct {
	Languages []string `json:"languages"`
	Pairs     []struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"pairs"`
}

type detectResponse struct {
	Language   string   `json:"language"`
	Confidence *float64 `json:"confidence,omitempty"`
}

// translateBatchLimit is the upstream's own ceiling. Checked here so a caller
// piping a long file is told to split it, rather than having the whole batch
// refused with none of it translated.
const translateBatchLimit = 64

func newCallTranslateCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output    string
		to        string
		from      string
		html      bool
		perLine   bool
		detect    bool
		languages bool
		apiKey    string
	)
	cmd := &cobra.Command{
		Use:   "translate [text…]",
		Short: "translate text, detect a language, or list the pairs",
		Long: `Translate text between languages.

The text is the arguments, or standard input when there are none. --per-line
translates each line separately in one request, which is how a file of strings
gets translated without losing which line is which.

--from is optional and detection fills it in. Naming it is worth doing on short
text: two words are not much for a detector to work with.

--languages lists what the model serves and does not translate anything. Run it
first against an unfamiliar deployment — a pair the model was not built for is
refused rather than routed through English. --detect reports what language the
text is in, and also translates nothing.

There is no --model here. These routes carry no model field: Router resolves the
translate default per call, so the model is a deployment's choice rather than a
caller's. "olares-cli router route list --kind default" says which one it is.

Examples:
  olares-cli router call translate --to zh "the quick brown fox"
  olares-cli router call translate --to en --from ja < notes.txt
  cat strings.txt | olares-cli router call translate --to de --per-line
  olares-cli router call translate --languages
  olares-cli router call translate --detect "Der Hund"
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			switch {
			case languages && detect:
				return fmt.Errorf("--languages lists the pairs and --detect reads one text; pass one or the other")
			case languages:
				if len(args) > 0 {
					return fmt.Errorf("--languages takes no text")
				}
				return runLanguages(c.Context(), f, apiKey, output)
			case detect:
				text, err := readPromptArgs(args, "text")
				if err != nil {
					return err
				}
				return runDetect(c.Context(), f, text, apiKey, output)
			}
			if strings.TrimSpace(to) == "" {
				return fmt.Errorf("name the language to translate into with --to; " +
					"`olares-cli router call translate --languages` lists what this model serves")
			}
			return runCallTranslate(c.Context(), f, translateOptions{
				To: strings.TrimSpace(to), From: strings.TrimSpace(from),
				HTML: html, PerLine: perLine, Args: args,
				APIKey: apiKey, OutputIn: output,
			})
		},
	}
	cmd.Flags().StringVar(&to, "to", "", "the language to translate into, e.g. en, zh, ja")
	cmd.Flags().StringVar(&from, "from", "", "the language the text is in; detected when omitted")
	cmd.Flags().BoolVar(&html, "html", false, "the text is HTML; keep the markup and translate around it")
	cmd.Flags().BoolVar(&perLine, "per-line", false, "translate each line of input separately")
	cmd.Flags().BoolVar(&detect, "detect", false, "report what language the text is in and translate nothing")
	cmd.Flags().BoolVar(&languages, "languages", false, "list the languages and pairs this model serves")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

type translateOptions struct {
	To       string
	From     string
	HTML     bool
	PerLine  bool
	Args     []string
	APIKey   string
	OutputIn string
}

func runCallTranslate(ctx context.Context, f *cmdutil.Factory, opts translateOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(opts.OutputIn)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	dp, err := dataPlane(ctx, pc, opts.APIKey)
	if err != nil {
		return err
	}

	if !opts.PerLine {
		text, terr := readPromptArgs(opts.Args, "text")
		if terr != nil {
			return terr
		}
		var resp translateResponse
		req := translateRequest{From: opts.From, To: opts.To, Text: text, HTML: opts.HTML}
		if err := dp.doJSON(ctx, "POST", epTranslate, req, &resp); err != nil {
			return callErr(err)
		}
		if format == FormatJSON {
			return printJSON(os.Stdout, resp)
		}
		_, werr := fmt.Fprintln(os.Stdout, resp.Result)
		return werr
	}

	texts, err := translateLines(opts.Args)
	if err != nil {
		return err
	}
	var resp translateBatchResponse
	req := translateBatchRequest{From: opts.From, To: opts.To, Texts: texts, HTML: opts.HTML}
	if err := dp.doJSON(ctx, "POST", epTranslateBatch, req, &resp); err != nil {
		return callErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, resp)
	}
	// One translation per line, in the order they went in, so the output can be
	// pasted back where the input came from.
	for _, line := range resp.Results {
		if _, err := fmt.Fprintln(os.Stdout, line); err != nil {
			return err
		}
	}
	return nil
}

func translateLines(args []string) ([]string, error) {
	var lines []string
	if len(args) > 0 {
		lines = args
	} else {
		if isTerminal(os.Stdin) {
			return nil, fmt.Errorf("no text given; pass it as arguments or pipe it in")
		}
		read, err := readLines(os.Stdin)
		if err != nil {
			return nil, err
		}
		lines = read
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("no non-empty lines to translate")
	}
	if len(lines) > translateBatchLimit {
		return nil, fmt.Errorf("%d lines is more than the %d one request takes; "+
			"split the input, for instance with `split -l %d`",
			len(lines), translateBatchLimit, translateBatchLimit)
	}
	return lines, nil
}

func runLanguages(ctx context.Context, f *cmdutil.Factory, apiKey, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	dp, err := dataPlane(ctx, pc, apiKey)
	if err != nil {
		return err
	}
	var resp languagesResponse
	if err := dp.doJSON(ctx, "GET", epLanguages, nil, &resp); err != nil {
		return callErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, resp)
	}
	return renderLanguages(os.Stdout, &resp)
}

func renderLanguages(w io.Writer, resp *languagesResponse) error {
	if len(resp.Languages) == 0 && len(resp.Pairs) == 0 {
		_, err := fmt.Fprintln(w, "the model declares no languages.")
		return err
	}
	// Pairs are grouped by source, because "what can I translate this into" is
	// the question being asked, and a flat list of pairs answers it slowly.
	into := make(map[string][]string, len(resp.Pairs))
	for _, p := range resp.Pairs {
		into[p.From] = append(into[p.From], p.To)
	}
	if len(into) == 0 {
		_, err := fmt.Fprintf(w, "any pair among: %s\n", strings.Join(sortedCopy(resp.Languages), " "))
		return err
	}
	froms := make([]string, 0, len(into))
	for k := range into {
		froms = append(froms, k)
	}
	sort.Strings(froms)
	t := newTable(w, "FROM", "INTO")
	for _, from := range froms {
		t.row(from, strings.Join(sortedCopy(into[from]), " "))
	}
	return t.flush()
}

func sortedCopy(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}

func runDetect(ctx context.Context, f *cmdutil.Factory, text, apiKey, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	dp, err := dataPlane(ctx, pc, apiKey)
	if err != nil {
		return err
	}
	var resp detectResponse
	if err := dp.doJSON(ctx, "POST", epDetect, map[string]any{"text": text}, &resp); err != nil {
		return callErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, resp)
	}
	line := nonEmpty(resp.Language)
	if resp.Confidence != nil {
		line += fmt.Sprintf("  %.0f%% sure", *resp.Confidence*100)
	}
	_, err = fmt.Fprintln(os.Stdout, line)
	return err
}
