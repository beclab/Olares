package download

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewCookiesCommand assembles `olares-cli knowledge download cookies`.
func NewCookiesCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cookies",
		Short: "manage provider cookies (yt-dlp Netscape cookies)",
		Long: `Store and inspect per-domain provider cookies used by download
providers (e.g. yt-dlp) to fetch gated content.

Cookies are supplied as a Netscape cookies.txt file. The listing and
retrieve responses never expose the stored cookie text unless you pass
-o json (retrieve only).`,
	}
	cmd.AddCommand(newCookiesListCommand(f))
	cmd.AddCommand(newCookiesSetCommand(f))
	cmd.AddCommand(newCookiesDeleteCommand(f))
	cmd.AddCommand(newCookiesRetrieveCommand(f))
	cmd.AddCommand(newCookiesHealthCommand(f))
	return cmd
}

func newCookiesListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list stored cookie domains",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runCookiesList(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runCookiesList(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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
	var result CookieListResult
	if err := doGet(ctx, pc.doer, "/api/integration/cookies", &result); err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, result)
	default:
		return renderCookieList(os.Stdout, result)
	}
}

func renderCookieList(w io.Writer, result CookieListResult) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "DOMAIN\tPROVIDER\tHAS_COOKIE\tUPDATED")
	for _, ck := range result.List {
		fmt.Fprintf(tw, "%s\t%s\t%v\t%s\n",
			orDash(ck.Domain),
			orDash(ck.Provider),
			ck.HasCookie,
			formatUnix(ck.UpdatedAt),
		)
	}
	return tw.Flush()
}

func newCookiesSetCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		domain     string
		provider   string
		cookieFile string
		output     string
	)
	cmd := &cobra.Command{
		Use:   "set",
		Short: "store or replace the cookie for a domain",
		Long: `Store or replace a domain's cookie (PUT /api/integration/cookies).

--cookie-file points to a local Netscape cookies.txt file whose full text
is uploaded. --provider defaults to the server default (yt-dlp) when unset.`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runCookiesSet(c.Context(), f, domain, provider, cookieFile, output)
		},
	}
	addOutputFlag(cmd, &output)
	cmd.Flags().StringVar(&domain, "domain", "", "cookie domain, e.g. youtube.com (required)")
	cmd.Flags().StringVar(&provider, "provider", "", "provider name (default: server default, yt-dlp)")
	cmd.Flags().StringVar(&cookieFile, "cookie-file", "", "path to a local Netscape cookies.txt file (required)")
	_ = cmd.MarkFlagRequired("domain")
	_ = cmd.MarkFlagRequired("cookie-file")
	return cmd
}

func runCookiesSet(ctx context.Context, f *cmdutil.Factory, domain, provider, cookieFile, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return fmt.Errorf("--domain is required")
	}
	cookieFile = strings.TrimSpace(cookieFile)
	if cookieFile == "" {
		return fmt.Errorf("--cookie-file is required")
	}
	raw, err := os.ReadFile(cookieFile)
	if err != nil {
		return fmt.Errorf("read cookie file: %w", err)
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	req := UpsertCookieReq{
		Domain:   domain,
		Provider: strings.TrimSpace(provider),
		Cookie:   string(raw),
	}
	var summary CookieSummary
	if err := doMutate(ctx, pc.doer, "PUT", "/api/integration/cookies", req, &summary); err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, summary)
	default:
		fmt.Printf("set cookie for %s (provider=%s)\n", orDash(summary.Domain), orDash(summary.Provider))
		return nil
	}
}

func newCookiesDeleteCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		domain string
		output string
	)
	cmd := &cobra.Command{
		Use:     "delete",
		Aliases: []string{"rm"},
		Short:   "delete the stored cookie for a domain",
		Args:    cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runCookiesDelete(c.Context(), f, domain, output)
		},
	}
	addOutputFlag(cmd, &output)
	cmd.Flags().StringVar(&domain, "domain", "", "cookie domain to delete (required)")
	_ = cmd.MarkFlagRequired("domain")
	return cmd
}

func runCookiesDelete(ctx context.Context, f *cmdutil.Factory, domain, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, err := parseFormat(outputRaw); err != nil {
		return err
	}
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return fmt.Errorf("--domain is required")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	// The :user path segment is not trusted by the server (real identity is
	// the gateway-injected X-Bfl-User); the profile OlaresID is just a
	// placeholder to satisfy the route.
	path := fmt.Sprintf("/api/integration/cookies/%s/%s",
		url.PathEscape(pc.profile.OlaresID), url.PathEscape(domain))
	if err := doMutate(ctx, pc.doer, "DELETE", path, nil, nil); err != nil {
		return err
	}
	fmt.Printf("removed cookie for %s\n", domain)
	return nil
}

func newCookiesRetrieveCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		domain   string
		provider string
		output   string
	)
	cmd := &cobra.Command{
		Use:   "retrieve",
		Short: "test-retrieve the stored cookie for a domain",
		Long: `Retrieve the stored cookie for a domain
(POST /api/integration/cookies/retrieve).

The plaintext cookie is only printed with -o json; the table output shows
whether a cookie was found and when it was updated.`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runCookiesRetrieve(c.Context(), f, domain, provider, output)
		},
	}
	addOutputFlag(cmd, &output)
	cmd.Flags().StringVar(&domain, "domain", "", "cookie domain to retrieve (required)")
	cmd.Flags().StringVar(&provider, "provider", "", "provider name (default: server default, yt-dlp)")
	_ = cmd.MarkFlagRequired("domain")
	return cmd
}

func runCookiesRetrieve(ctx context.Context, f *cmdutil.Factory, domain, provider, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return fmt.Errorf("--domain is required")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	req := RetrieveCookieReq{Domain: domain, Provider: strings.TrimSpace(provider)}
	var res RetrieveCookieResult
	if err := doMutate(ctx, pc.doer, "POST", "/api/integration/cookies/retrieve", req, &res); err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, res)
	default:
		fmt.Printf("Domain:  %s\n", orDash(res.Domain))
		fmt.Printf("Found:   %v\n", res.Found)
		if res.Found && res.UpdatedAt > 0 {
			fmt.Printf("Updated: %s\n", formatUnix(res.UpdatedAt))
		}
		return nil
	}
}

func newCookiesHealthCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "health",
		Short: "report provider availability for cookie integrations",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runCookiesHealth(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runCookiesHealth(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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
	var res IntegrationHealth
	if err := doGet(ctx, pc.doer, "/api/integration/healthz", &res); err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, res)
	default:
		fmt.Printf("Healthy: %v\n", res.Healthy)
		tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		for name, status := range res.Providers {
			fmt.Fprintf(tw, "%s\t%s\n", name, orDash(status))
		}
		return tw.Flush()
	}
}
