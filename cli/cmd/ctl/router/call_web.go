package router

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"

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
		output string
		model  string
		limit  int
		apiKey string
	)
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "search the web",
		Long: `Search the web through whatever search provider Router is configured with.

Results come back in one shape whichever vendor answered — a title, a URL, a
snippet and sometimes a date — because Router normalises them. So a query
written against this stays written when the provider behind it changes.

This is an index lookup, not a fetch: the snippet is what the vendor holds.
"olares-cli router call scrape <url>" reads the page itself.

Searching needs a model whose mode is search; "olares-cli router model list --mode
search" shows the ones that qualify.

Examples:
  olares-cli router call search "olares self-hosted"
  olares-cli router call search "rust async runtime" --limit 5 -o json
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			query, err := readPromptArgs(args, "query")
			if err != nil {
				return err
			}
			req := searchRequest{Model: callModel(model, categorySearch), Query: query}
			if c.Flags().Changed("limit") {
				req.MaxResults = &limit
			}
			return runCallSearch(c.Context(), f, req, apiKey, output)
		},
	}
	cmd.Flags().StringVar(&model, "model", "", modelFlagHelp(categorySearch))
	cmd.Flags().IntVar(&limit, "limit", 0, "ask for at most this many results")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
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
	dp, _, err := dataPlane(ctx, pc, apiKey)
	if err != nil {
		return err
	}
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
	dp, _, err := dataPlane(ctx, pc, apiKey)
	if err != nil {
		return err
	}
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
