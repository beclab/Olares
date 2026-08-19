package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// POST /v1/rerank
//
// Router routes on the `model` field and forwards the rest untouched, so the
// answer is whatever the upstream sent. The vendors agree on enough of it to
// render — an index, a score, and sometimes the document echoed back — and
// anything past that is what -o json is for.
//
// Reranking is the second half of a search: an embedding search returns fifty
// candidates cheaply and roughly, and a reranker reads the query against each
// one and orders them properly. That is why the input is a query plus documents
// and the output is positions rather than text.

type rerankRequest struct {
	Model     string   `json:"model"`
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
	TopN      *int     `json:"top_n,omitempty"`
	// ReturnDocuments asks the upstream to echo each document back. Cohere and
	// Jina both honour it and both default to leaving it out, which would make
	// the table a column of numbers with nothing to read them against.
	ReturnDocuments bool `json:"return_documents,omitempty"`
}

// rerankResponse is the shape the vendors share. `document` is a string for
// some and an object with a `text` field for others, so it stays raw and is
// unpacked when rendering.
type rerankResponse struct {
	Model   string `json:"model"`
	Results []struct {
		Index          int             `json:"index"`
		RelevanceScore float64         `json:"relevance_score"`
		Document       json.RawMessage `json:"document,omitempty"`
	} `json:"results"`
	Usage *struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}

func newCallRerankCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output    string
		model     string
		documents []string
		topN      int
		apiKey    string
	)
	cmd := &cobra.Command{
		Use:   "rerank <query>",
		Short: "order documents by how well they answer a query",
		Long: `Score documents against a query and order them.

Each --document is one candidate; with none, the candidates are the lines of
standard input, which is how the output of a search gets reranked in a pipe.

The answer is positions and scores, not text: the row numbers refer back to the
order the documents went in. Scores are the model's own scale — comparable
within one call and not across models — so what matters is the order.

Reranking needs a model whose mode is rerank. A chat or embedding model refuses
the route; "olares-cli router list --mode rerank" shows the ones that qualify.

Examples:
  olares-cli router call rerank "how do I reset my password" \
    --document "Password reset guide" --document "Billing FAQ"
  grep -h . candidates.txt | olares-cli router call rerank "cheap flights" --top-n 5
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			query, err := readPromptArgs(args, "query")
			if err != nil {
				return err
			}
			docs, err := rerankDocuments(documents)
			if err != nil {
				return err
			}
			var top *int
			if c.Flags().Changed("top-n") {
				top = &topN
			}
			return runCallRerank(c.Context(), f, rerankRequest{
				Model:           callModel(model, categoryRerank),
				Query:           query,
				Documents:       docs,
				TopN:            top,
				ReturnDocuments: true,
			}, apiKey, output)
		},
	}
	cmd.Flags().StringVar(&model, "model", "", modelFlagHelp(categoryRerank))
	cmd.Flags().StringArrayVar(&documents, "document", nil, "a candidate document; repeatable")
	cmd.Flags().IntVar(&topN, "top-n", 0, "return only this many of the best matches")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

// rerankDocuments takes the documents from the flag, or from stdin one per
// line. A query on the command line and the candidates on stdin is the shape
// this verb is used in, so the two cannot both come from stdin.
func rerankDocuments(fromFlags []string) ([]string, error) {
	if len(fromFlags) > 0 {
		return fromFlags, nil
	}
	if isTerminal(os.Stdin) {
		return nil, fmt.Errorf("no documents given; pass --document for each, or pipe them in one per line")
	}
	lines, err := readLines(os.Stdin)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("no non-empty lines on stdin to rank")
	}
	return lines, nil
}

func runCallRerank(ctx context.Context, f *cmdutil.Factory, req rerankRequest, apiKey, outputRaw string) error {
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
	dp, _, err := dataPlane(ctx, pc, apiKey)
	if err != nil {
		return err
	}
	var resp rerankResponse
	if err := dp.doJSON(ctx, "POST", epRerank, req, &resp); err != nil {
		return callErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, resp)
	}
	return renderRerank(os.Stdout, &resp, req.Documents)
}

func renderRerank(w io.Writer, resp *rerankResponse, docs []string) error {
	if len(resp.Results) == 0 {
		_, err := fmt.Fprintln(w, "the model ranked nothing.")
		return err
	}
	t := newTable(w, "RANK", "INPUT #", "SCORE", "DOCUMENT")
	for i := range resp.Results {
		r := &resp.Results[i]
		text := rerankDocumentText(r.Document)
		if text == "" && r.Index >= 0 && r.Index < len(docs) {
			text = docs[r.Index]
		}
		t.row(strconv.Itoa(i+1), strconv.Itoa(r.Index),
			fmt.Sprintf("%.4f", r.RelevanceScore), clip(text, 60))
	}
	if err := t.flush(); err != nil {
		return err
	}
	line := "\n" + nonEmpty(resp.Model)
	if resp.Usage != nil {
		line += fmt.Sprintf("  %d tokens", resp.Usage.TotalTokens)
	}
	_, err := fmt.Fprintln(os.Stderr, line)
	return err
}

// rerankDocumentText unpacks the echoed document. Cohere sends an object with a
// `text` field and Jina sends the string; an unrecognised shape yields nothing
// so the caller falls back to the input it sent.
func rerankDocumentText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var plain string
	if json.Unmarshal(raw, &plain) == nil {
		return strings.TrimSpace(plain)
	}
	var obj struct {
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &obj) == nil {
		return strings.TrimSpace(obj.Text)
	}
	return ""
}
