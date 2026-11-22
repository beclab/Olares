package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// POST /v1/search, POST /v1/scrape
//
// The two web modes, and the only place in this tree where Router normalises an
// upstream instead of passing it through. Every search vendor has its own
// envelope, so Router flattens them to a title, a URL, a snippet and a date,
// and every scraper has its own document model, so Router flattens that to
// markdown. That is what makes the two routes worth calling through a gateway:
// swapping the vendor does not change the caller.
//
// Both are metered like any other call, and neither is a browser: a search
// returns what the vendor's index holds, and a scrape returns one page as text
// with no scripts run and nothing followed.

type searchRequest struct {
	Model      string `json:"model,omitempty"`
	Query      string `json:"query"`
	MaxResults *int   `json:"max_results,omitempty"`
}

type searchResponse struct {
	Object  string `json:"object"`
	Results []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Snippet string `json:"snippet"`
		Date    string `json:"date,omitempty"`
	} `json:"results"`
}

const (
	searchSourceWeb       = "web"
	searchSourceFiles     = "files"
	filesRAGSchemaVersion = "rag.retrieval.v1"
)

type filesRAGRerankRequest struct {
	Enabled bool `json:"enabled"`
	TopN    int  `json:"top_n"`
}

// filesRAGSearchRequest deliberately mirrors Search3's public request instead
// of its trusted internal state. In particular it has no owner or authorized
// roots: Router and Search3 derive those from the delegated final user.
type filesRAGSearchRequest struct {
	SchemaVersion     string                `json:"schema_version"`
	DirectoryRoots    []string              `json:"directory_roots,omitempty"`
	ResourceURIs      []string              `json:"resource_uris,omitempty"`
	OriginalQuery     string                `json:"original_query"`
	EffectiveQuery    string                `json:"effective_query"`
	RetrievalMode     string                `json:"retrieval_mode"`
	CandidateK        int                   `json:"candidate_k"`
	TopK              int                   `json:"top_k"`
	Rerank            filesRAGRerankRequest `json:"rerank"`
	DeadlineMS        int64                 `json:"deadline_ms"`
	IncludeHighlights bool                  `json:"include_highlights"`
}

type filesRAGSearchResponse struct {
	Hits []struct {
		Rank        int             `json:"rank"`
		ResourceURI string          `json:"resource_uri"`
		Content     string          `json:"content"`
		Scores      json.RawMessage `json:"scores"`
	} `json:"hits"`
	Abstention struct {
		Applied bool   `json:"applied"`
		Reason  string `json:"reason"`
	} `json:"abstention"`
	Degraded []struct {
		Code string `json:"code"`
	} `json:"degraded"`
}

type scrapeRequest struct {
	Model string `json:"model,omitempty"`
	URL   string `json:"url"`
}

type scrapeDocument struct {
	Object   string `json:"object"`
	URL      string `json:"url"`
	Title    string `json:"title,omitempty"`
	Markdown string `json:"markdown"`
	Status   string `json:"status"`
}

func newCallSearchCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output            string
		model             string
		limit             int
		apiKey            string
		source            string
		mode              string
		directoryRoots    []string
		resourceURIs      []string
		candidateK        int
		rerank            bool
		rerankTopN        int
		deadline          time.Duration
		includeHighlights bool
		effectiveQuery    string
	)
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "search the web or the current user's authorized files",
		Long: `Search the web, or retrieve evidence from the current user's authorized Files RAG index.

Web search is the default and preserves the existing command. Use --source
files from a host application such as Lares to call Router's delegated Files
RAG route. Router binds the final user supplied by the host; Search3 obtains
that user's searchable scope from Files. Directory and resource flags can only
narrow that scope.

For --source web, results come back in one shape whichever vendor answered — a
title, a URL, a snippet and sometimes a date — because Router normalises them.
For --source files, JSON is the stable Search3 evidence contract: ranked chunks,
source locators, channel scores, provenance and optional highlight ranges.

Web search is an index lookup, not a fetch: the snippet is what the vendor holds.
"olares-cli router call scrape <url>" reads the page itself.

Web search needs a model whose mode is search; "olares-cli router model list
--mode search" shows the ones that qualify. Files search uses the current RAG
index and takes no --model or --api-key.

Examples:
  olares-cli router call search "olares self-hosted"
  olares-cli router call search "rust async runtime" --limit 5 -o json
  olares-cli router call search --source files "港口安全标识" -o json
  olares-cli router call search --source files "错误码 E1042" --mode lexical --limit 5
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			query, err := readPromptArgs(args, "query")
			if err != nil {
				return err
			}
			switch strings.ToLower(strings.TrimSpace(source)) {
			case searchSourceWeb:
				if err := rejectFilesSearchFlags(c); err != nil {
					return err
				}
				return runCallSearch(c.Context(), f, searchRequest{
					Model: callModel(model, categorySearch), Query: query,
					MaxResults: changedInt(c, "limit", limit),
				}, apiKey, output)
			case searchSourceFiles:
				if c.Flags().Changed("model") || strings.TrimSpace(model) != "" {
					return fmt.Errorf("--model applies to --source web; Files RAG is an indexed service, not a model route")
				}
				if resolveDataPlaneAuth(apiKey).Mode != authHost {
					return fmt.Errorf("--source files requires a host-delegated final user via %s; an API key or a laptop profile cannot name the user whose Files permissions Search3 must enforce", dataPlaneURLEnv)
				}
				if !c.Flags().Changed("limit") {
					limit = 10
				}
				if strings.TrimSpace(effectiveQuery) == "" {
					effectiveQuery = query
				}
				req := filesRAGSearchRequest{
					SchemaVersion: filesRAGSchemaVersion, DirectoryRoots: directoryRoots,
					ResourceURIs: resourceURIs, OriginalQuery: query,
					EffectiveQuery: effectiveQuery, RetrievalMode: strings.ToLower(strings.TrimSpace(mode)),
					CandidateK: candidateK, TopK: limit,
					Rerank:     filesRAGRerankRequest{Enabled: rerank, TopN: rerankTopN},
					DeadlineMS: deadline.Milliseconds(), IncludeHighlights: includeHighlights,
				}
				if err := validateFilesRAGSearchRequest(req); err != nil {
					return err
				}
				return runCallFilesRAGSearch(c.Context(), f, req, output)
			default:
				return fmt.Errorf("unsupported --source %q (allowed: web, files)", source)
			}
		},
	}
	cmd.Flags().StringVar(&source, "source", searchSourceWeb, "search source: web, files")
	cmd.Flags().StringVar(&model, "model", "", modelFlagHelp(categorySearch))
	cmd.Flags().IntVar(&limit, "limit", 0, "ask for at most this many results")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	cmd.Flags().StringVar(&mode, "mode", "hybrid", "Files RAG retrieval mode: lexical, vector, hybrid")
	cmd.Flags().StringSliceVar(&directoryRoots, "directory-root", nil, "narrow Files RAG to this canonical directory root (repeatable)")
	cmd.Flags().StringSliceVar(&resourceURIs, "resource-uri", nil, "narrow Files RAG to this exact canonical file URI (repeatable)")
	cmd.Flags().IntVar(&candidateK, "candidate-k", 40, "Files RAG candidate count before final ranking (1-200)")
	cmd.Flags().BoolVar(&rerank, "rerank", false, "rerank Files RAG candidates with the configured rerank model")
	cmd.Flags().IntVar(&rerankTopN, "rerank-top-n", 5, "Files RAG candidates to rerank (1-10)")
	cmd.Flags().DurationVar(&deadline, "deadline", 5*time.Second, "Files RAG server-side deadline (maximum 30s)")
	cmd.Flags().BoolVar(&includeHighlights, "highlights", true, "include display and source highlight ranges in Files RAG JSON results")
	cmd.Flags().StringVar(&effectiveQuery, "effective-query", "", "optional rewritten Files RAG query; defaults to the original query")
	addOutputFlag(cmd, &output)
	return cmd
}

func changedInt(c *cobra.Command, name string, value int) *int {
	if c.Flags().Changed(name) {
		return &value
	}
	return nil
}

func rejectFilesSearchFlags(c *cobra.Command) error {
	for _, name := range []string{"mode", "directory-root", "resource-uri", "candidate-k", "rerank", "rerank-top-n", "deadline", "highlights", "effective-query"} {
		if c.Flags().Changed(name) {
			return fmt.Errorf("--%s applies only to --source files", name)
		}
	}
	return nil
}

func validateFilesRAGSearchRequest(req filesRAGSearchRequest) error {
	if req.RetrievalMode != "lexical" && req.RetrievalMode != "vector" && req.RetrievalMode != "hybrid" {
		return fmt.Errorf("unsupported --mode %q (allowed: lexical, vector, hybrid)", req.RetrievalMode)
	}
	if req.CandidateK < 1 || req.CandidateK > 200 {
		return fmt.Errorf("--candidate-k must be between 1 and 200")
	}
	if req.TopK < 1 || req.TopK > 50 || req.TopK > req.CandidateK {
		return fmt.Errorf("--limit must be between 1 and 50 and no greater than --candidate-k")
	}
	if req.Rerank.TopN < 1 || req.Rerank.TopN > 10 || req.Rerank.TopN > req.CandidateK {
		return fmt.Errorf("--rerank-top-n must be between 1 and 10 and no greater than --candidate-k")
	}
	if req.DeadlineMS < 1 || req.DeadlineMS > 30_000 {
		return fmt.Errorf("--deadline must be greater than zero and no more than 30s")
	}
	if len(req.DirectoryRoots) > 64 {
		return fmt.Errorf("--directory-root accepts at most 64 values")
	}
	if len(req.ResourceURIs) > 16 {
		return fmt.Errorf("--resource-uri accepts at most 16 values")
	}
	return nil
}

func runCallFilesRAGSearch(ctx context.Context, f *cmdutil.Factory, req filesRAGSearchRequest, outputRaw string) error {
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
	dp := dataPlane(pc, "")
	if resolveDataPlaneAuth("").Mode != authHost {
		return fmt.Errorf("Files RAG search lost its host-delegated identity before the request")
	}
	raw, err := callFilesRAGSearch(ctx, dp, req)
	if err != nil {
		return callErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, raw)
	}
	var resp filesRAGSearchResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return fmt.Errorf("decode Files RAG response for table output: %w", err)
	}
	return renderFilesRAGSearch(os.Stdout, &resp)
}

func callFilesRAGSearch(ctx context.Context, client *routerClient, req filesRAGSearchRequest) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := client.doJSON(ctx, "POST", epFilesRAGSearch, req, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func renderFilesRAGSearch(w io.Writer, resp *filesRAGSearchResponse) error {
	if len(resp.Hits) == 0 {
		if resp.Abstention.Applied && strings.TrimSpace(resp.Abstention.Reason) != "" {
			_, err := fmt.Fprintf(w, "no authorized evidence: %s\n", resp.Abstention.Reason)
			return err
		}
		_, err := fmt.Fprintln(w, "no authorized file evidence matched that query.")
		return err
	}
	t := newTable(w, "#", "SOURCE", "SCORE", "EVIDENCE")
	for i := range resp.Hits {
		h := &resp.Hits[i]
		t.row(strconv.Itoa(h.Rank), clip(h.ResourceURI, 54), clip(string(h.Scores), 24), clip(strings.Join(strings.Fields(h.Content), " "), 64))
	}
	if err := t.flush(); err != nil {
		return err
	}
	if len(resp.Degraded) > 0 {
		codes := make([]string, 0, len(resp.Degraded))
		for _, item := range resp.Degraded {
			codes = append(codes, item.Code)
		}
		_, _ = fmt.Fprintf(os.Stderr, "\ndegraded: %s\n", strings.Join(codes, ","))
	}
	_, err := fmt.Fprintln(os.Stderr, "\n-o json carries locators, channel scores, citations, and highlight ranges")
	return err
}

func runCallSearch(ctx context.Context, f *cmdutil.Factory, req searchRequest, apiKey, outputRaw string) error {
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
	dp := dataPlane(pc, apiKey)
	var resp searchResponse
	if err := dp.doJSON(ctx, "POST", epSearch, req, &resp); err != nil {
		return callErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, resp)
	}
	return renderSearch(os.Stdout, &resp)
}

func renderSearch(w io.Writer, resp *searchResponse) error {
	if len(resp.Results) == 0 {
		_, err := fmt.Fprintln(w, "the provider found nothing for that query.")
		return err
	}
	t := newTable(w, "#", "TITLE", "URL", "WHEN")
	for i := range resp.Results {
		r := &resp.Results[i]
		t.row(strconv.Itoa(i+1), clip(nonEmpty(r.Title), 48), clip(r.URL, 56), nonEmpty(r.Date))
	}
	if err := t.flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintln(os.Stderr, "\n-o json carries the snippets")
	return err
}

func newCallScrapeCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		model  string
		apiKey string
	)
	cmd := &cobra.Command{
		Use:   "scrape <url>",
		Short: "read one page as markdown",
		Long: `Fetch a page and return it as markdown.

The provider does the fetching, so this reaches pages a script on this machine
may not — and equally, a page behind a login is not readable by either. Nothing
is executed and no link is followed: one URL in, one document out.

Only http and https are accepted, and an address on a private network or a cloud
metadata endpoint is refused. Router does the refusing, because a scraper that
can be pointed inward is a way to read things the caller has no access to.

Scraping needs a model whose mode is scrape; "olares-cli router model list --mode
scrape" shows the ones that qualify.

Examples:
  olares-cli router call scrape https://olares.com
  olares-cli router call scrape https://example.com/post > post.md
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			target, err := checkScrapeURL(args[0])
			if err != nil {
				return err
			}
			return runCallScrape(c.Context(), f, scrapeRequest{
				Model: callModel(model, categoryScrape),
				URL:   target,
			}, apiKey, output)
		},
	}
	cmd.Flags().StringVar(&model, "model", "", modelFlagHelp(categoryScrape))
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

// checkScrapeURL catches the mistake worth catching locally: a bare hostname,
// which is what people type. Router enforces the rest — the scheme, and that the
// address is not on a private network — and it has to, since it is the one
// making the request.
func checkScrapeURL(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", fmt.Errorf("name a page to read")
	}
	parsed, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("%q is not a URL: %w", s, err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("%q needs a scheme and a host, as in https://%s", s, strings.TrimPrefix(s, "//"))
	}
	return s, nil
}

func runCallScrape(ctx context.Context, f *cmdutil.Factory, req scrapeRequest, apiKey, outputRaw string) error {
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
	dp := dataPlane(pc, apiKey)
	var doc scrapeDocument
	if err := dp.doJSON(ctx, "POST", epScrape, req, &doc); err != nil {
		return callErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, doc)
	}
	// The markdown goes to stdout unadorned so it can be redirected into a
	// file; what it is and where it came from goes to stderr.
	if _, err := fmt.Fprintln(os.Stdout, strings.TrimRight(doc.Markdown, "\n")); err != nil {
		return err
	}
	_, err = fmt.Fprintf(os.Stderr, "\n%s  %s\n", nonEmpty(doc.Title), nonEmpty(doc.URL))
	return err
}
