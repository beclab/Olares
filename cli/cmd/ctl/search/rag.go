package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

const (
	ragMaxQueryChars = 2048
	ragMaxRoots      = 64
)

// ragSearchOptions contains only browser-facing fields accepted by
// user-service. In particular, it has no owner or canonical resource URI:
// user-service derives both from the authenticated profile and Files ACLs.
type ragSearchOptions struct {
	roots             []string
	mode              string
	candidateK        int
	topK              int
	rerank            bool
	rerankTopN        int
	deadlineMS        int
	includeHighlights bool
	output            string
}

type ragSearchRequest struct {
	Query             string            `json:"query"`
	DirectoryRoots    []string          `json:"directoryRoots,omitempty"`
	RetrievalMode     string            `json:"retrievalMode,omitempty"`
	CandidateK        *int              `json:"candidateK,omitempty"`
	TopK              *int              `json:"topK,omitempty"`
	Rerank            *ragRerankRequest `json:"rerank,omitempty"`
	DeadlineMS        *int              `json:"deadlineMs,omitempty"`
	IncludeHighlights *bool             `json:"includeHighlights,omitempty"`
}

type ragRerankRequest struct {
	Enabled bool `json:"enabled"`
	TopN    *int `json:"topN,omitempty"`
}

type ragRetrievalResponse struct {
	SchemaVersion string                    `json:"schema_version"`
	Status        string                    `json:"status"`
	Hits          []ragRetrievalHit         `json:"hits"`
	Truncated     bool                      `json:"truncated"`
	Degraded      []ragRetrievalDegradation `json:"degraded"`
	TraceID       string                    `json:"trace_id"`
	Error         *ragRetrievalError        `json:"error,omitempty"`
}

type ragRetrievalError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ragRetrievalDegradation struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ragRetrievalHit struct {
	Rank        int                    `json:"rank"`
	ChunkKey    string                 `json:"chunk_key"`
	Content     string                 `json:"content"`
	ResourceURI string                 `json:"resource_uri"`
	Modality    string                 `json:"modality"`
	Locator     json.RawMessage        `json:"locator"`
	Scores      map[string]*float64    `json:"scores"`
	Provenance  ragRetrievalProvenance `json:"provenance"`
}

type ragRetrievalProvenance struct {
	Channels         []string `json:"channels"`
	RetrievalProfile string   `json:"retrieval_profile"`
}

type ragJSONDoer interface {
	DoJSON(context.Context, string, string, interface{}, interface{}) error
}

func newRAGCommand(f *cmdutil.Factory) *cobra.Command {
	o := &ragSearchOptions{}
	cmd := &cobra.Command{
		Use:   "rag <query>",
		Short: "Search indexed Files with RAG retrieval",
		Long: `Search the authenticated user's indexed Files through the trusted
user-service RAG endpoint. user-service derives the owner and authorized roots;
this command never accepts a trusted owner or Search3 canonical URI.

Only explicitly supplied tuning flags are sent. Otherwise the deployed
user-service/Search3 policy supplies its current defaults.

Examples:
  olares-cli search rag "refund error code"
  olares-cli search rag "北京大学" --mode hybrid --highlights
  olares-cli search rag "release plan" --root drive/Home/Documents/ -o json
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query, err := parseRAGQuery(args)
			if err != nil {
				return err
			}
			doer, err := newDoer(cmd.Context(), f)
			if err != nil {
				return err
			}
			request, err := buildRAGRequest(cmd, query, o)
			if err != nil {
				return err
			}
			return executeRAGSearch(cmd.Context(), doer, request, o.output)
		},
	}
	cmd.SilenceUsage = true
	cmd.Flags().StringArrayVar(&o.roots, "root", nil, "authorized frontend directory root (repeatable)")
	cmd.Flags().StringVar(&o.mode, "mode", "", "retrieval mode: lexical, vector, hybrid")
	cmd.Flags().IntVar(&o.candidateK, "candidate-k", 0, "candidate count override (1..200)")
	cmd.Flags().IntVar(&o.topK, "top-k", 0, "result count override (1..50)")
	cmd.Flags().BoolVar(&o.rerank, "rerank", false, "enable configured reranker")
	cmd.Flags().IntVar(&o.rerankTopN, "rerank-top-n", 0, "rerank candidate count override (1..10)")
	cmd.Flags().IntVar(&o.deadlineMS, "deadline-ms", 0, "request deadline override in milliseconds (1..30000)")
	cmd.Flags().BoolVar(&o.includeHighlights, "highlights", false, "include content and source highlights")
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", "output format: table, json")
	return cmd
}

func parseRAGQuery(args []string) (string, error) {
	query := strings.TrimSpace(strings.Join(args, " "))
	if query == "" {
		return "", fmt.Errorf("a non-empty RAG query is required")
	}
	if len([]rune(query)) > ragMaxQueryChars {
		return "", fmt.Errorf("RAG query must not exceed %d characters", ragMaxQueryChars)
	}
	return query, nil
}

func buildRAGRequest(cmd *cobra.Command, query string, o *ragSearchOptions) (ragSearchRequest, error) {
	format, err := parseFormat(o.output)
	if err != nil {
		return ragSearchRequest{}, err
	}
	o.output = string(format)
	if len(o.roots) > ragMaxRoots {
		return ragSearchRequest{}, fmt.Errorf("--root accepts at most %d entries", ragMaxRoots)
	}
	for _, root := range o.roots {
		if err := validateRAGRoot(root); err != nil {
			return ragSearchRequest{}, err
		}
	}
	if o.mode != "" && o.mode != "lexical" && o.mode != "vector" && o.mode != "hybrid" {
		return ragSearchRequest{}, fmt.Errorf("unsupported --mode %q (allowed: lexical, vector, hybrid)", o.mode)
	}
	if cmd.Flags().Changed("candidate-k") && (o.candidateK < 1 || o.candidateK > 200) {
		return ragSearchRequest{}, fmt.Errorf("--candidate-k must be 1..200")
	}
	if cmd.Flags().Changed("top-k") && (o.topK < 1 || o.topK > 50) {
		return ragSearchRequest{}, fmt.Errorf("--top-k must be 1..50")
	}
	if cmd.Flags().Changed("candidate-k") && cmd.Flags().Changed("top-k") && o.topK > o.candidateK {
		return ragSearchRequest{}, fmt.Errorf("--top-k must not exceed --candidate-k")
	}
	if cmd.Flags().Changed("rerank-top-n") && (o.rerankTopN < 1 || o.rerankTopN > 10) {
		return ragSearchRequest{}, fmt.Errorf("--rerank-top-n must be 1..10")
	}
	if cmd.Flags().Changed("deadline-ms") && (o.deadlineMS < 1 || o.deadlineMS > 30000) {
		return ragSearchRequest{}, fmt.Errorf("--deadline-ms must be 1..30000")
	}

	request := ragSearchRequest{Query: query, DirectoryRoots: o.roots, RetrievalMode: o.mode}
	if cmd.Flags().Changed("candidate-k") {
		request.CandidateK = intPointer(o.candidateK)
	}
	if cmd.Flags().Changed("top-k") {
		request.TopK = intPointer(o.topK)
	}
	if cmd.Flags().Changed("rerank") || cmd.Flags().Changed("rerank-top-n") {
		request.Rerank = &ragRerankRequest{Enabled: o.rerank}
		if cmd.Flags().Changed("rerank-top-n") {
			request.Rerank.TopN = intPointer(o.rerankTopN)
		}
	}
	if cmd.Flags().Changed("deadline-ms") {
		request.DeadlineMS = intPointer(o.deadlineMS)
	}
	if cmd.Flags().Changed("highlights") {
		request.IncludeHighlights = boolPointer(o.includeHighlights)
	}
	return request, nil
}

func validateRAGRoot(root string) error {
	trimmed := strings.TrimSpace(strings.TrimLeft(root, "/"))
	if trimmed == "" || strings.Contains(trimmed, "\\") || strings.Contains(trimmed, ":") || strings.Contains(trimmed, "//") {
		return fmt.Errorf("invalid --root %q: use a frontend path such as drive/Home/Documents/", root)
	}
	for _, part := range strings.Split(trimmed, "/") {
		if part == "." || part == ".." {
			return fmt.Errorf("invalid --root %q: dot path segments are not allowed", root)
		}
	}
	return nil
}

func executeRAGSearch(ctx context.Context, doer ragJSONDoer, request ragSearchRequest, output string) error {
	var raw json.RawMessage
	if err := doer.DoJSON(ctx, http.MethodPost, "/api/retrieval/query", request, &raw); err != nil {
		return err
	}
	var response ragRetrievalResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		return fmt.Errorf("decode RAG retrieval response: %w", err)
	}
	if response.Status != "ok" {
		if response.Error != nil {
			return fmt.Errorf("RAG retrieval %s: %s", response.Error.Code, response.Error.Message)
		}
		return fmt.Errorf("RAG retrieval returned status %q", response.Status)
	}
	if output == string(FormatJSON) {
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, raw, "", "  "); err != nil {
			return fmt.Errorf("format RAG retrieval response: %w", err)
		}
		_, err := fmt.Fprintln(os.Stdout, pretty.String())
		return err
	}
	return renderRAGResults(response)
}

func renderRAGResults(response ragRetrievalResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "RANK\tSCORE\tSOURCE\tMATCH"); err != nil {
		return err
	}
	for _, hit := range response.Hits {
		snippet := strings.Join(strings.Fields(hit.Content), " ")
		if len([]rune(snippet)) > 120 {
			snippet = string([]rune(snippet)[:117]) + "..."
		}
		if _, err := fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", hit.Rank, ragHitScore(hit), hit.ResourceURI, snippet); err != nil {
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if len(response.Hits) == 0 {
		fmt.Fprintln(os.Stderr, "No relevant RAG results.")
	}
	for _, degraded := range response.Degraded {
		fmt.Fprintf(os.Stderr, "RAG degraded (%s): %s\n", degraded.Code, degraded.Message)
	}
	if response.Truncated {
		fmt.Fprintln(os.Stderr, "RAG results were truncated; increase --top-k within the service limit if needed.")
	}
	return nil
}

func ragHitScore(hit ragRetrievalHit) string {
	for _, key := range []string{"rerank", "fused", "final", "vector", "lexical"} {
		if score := hit.Scores[key]; score != nil {
			return strconv.FormatFloat(*score, 'f', 4, 64)
		}
	}
	return "-"
}

func intPointer(value int) *int    { return &value }
func boolPointer(value bool) *bool { return &value }
