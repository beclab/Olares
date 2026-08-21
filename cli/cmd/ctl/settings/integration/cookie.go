package integration

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/cookieparse"
)

// `olares-cli settings integration cookie ...`
//
// Backed by user-service's /api/cookie/* (cookie.controller.ts), served on
// the settings host. This is the same store the Settings SPA writes and
// download-server reads through integration-provider, so a cookie imported
// here is immediately usable by `knowledge download`.
//
//	GET  /api/cookie/all           → []DomainCookie          (list)
//	POST /api/cookie               → upsert one DomainCookie  (import)
//	POST /api/cookie/retrieve      → []DomainCookie by domain (validate, merge)
//	POST /api/cookie/delete {key}  → delete by store key      (rm)
//
// Cookie values never appear in command-line arguments, terminal output or
// error messages: arguments are world-readable through `ps` and land in
// shell history.
func NewCookieCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cookie",
		Short: "browser cookies used by downloads and Wise collection",
		Long: `Manage the per-domain cookie store that Olares uses to download
content behind a login (YouTube members-only videos, age-restricted
media, private Hugging Face repos, ...).

Subcommands:
  import   [--domain] <--file>
  list
  rm       <domain>
  validate <domain>

This is the same store as Settings -> Integration -> Cookies in the
browser, so cookies imported here show up there and vice versa.

Cookie values are secrets. Import reads them from a file or stdin and
never accepts them as a command-line argument, and no subcommand ever
prints a cookie value back.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newCookieImportCommand(f))
	cmd.AddCommand(newCookieListCommand(f))
	cmd.AddCommand(newCookieRemoveCommand(f))
	cmd.AddCommand(newCookieValidateCommand(f))
	return cmd
}

// domainCookie mirrors user-service's DomainCookie (src/utils.ts).
type domainCookie struct {
	Domain     string               `json:"domain"`
	Account    string               `json:"account"`
	Records    []cookieparse.Record `json:"records"`
	UpdateTime int64                `json:"updateTime"`
}

// storeKey mirrors DomainCookie.get_store_key(). The account segment is
// what download-server looks a cookie up by, so an empty account writes a
// row nothing can ever read.
func (d domainCookie) storeKey() string {
	if d.Account == "" {
		return "cookie:" + d.Domain
	}
	return "cookie:" + d.Domain + ":" + d.Account
}

// cookieAccount derives the account segment from the Olares ID, matching
// the SPA's `olaresId.split('@')[0]`.
func cookieAccount(olaresID string) (string, error) {
	account := strings.TrimSpace(olaresID)
	if idx := strings.Index(account, "@"); idx >= 0 {
		account = account[:idx]
	}
	if account == "" {
		return "", fmt.Errorf("cannot derive the cookie account: profile has no Olares ID")
	}
	return account, nil
}

func newCookieImportCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		domain     string
		file       string
		formatFlag string
		merge      bool
	)
	cmd := &cobra.Command{
		Use:   "import",
		Short: "import cookies from a file or stdin",
		Long: `Import cookies from a file, or from stdin with --file -.

Formats (--format, default auto-detect):

  netscape  cookies.txt, the format yt-dlp / curl / wget use and that
            browser extensions export by default.
  json      the array browser extensions such as EditThisCookie export.
  header    a raw header line. Two shapes are accepted: a request-style
            "a=b; c=d" (needs --domain, since it carries none) and a
            response-style "Set-Cookie: a=b; Domain=...; HttpOnly".

By default every host found in the file is written to its own store
key, matching Settings -> Integration -> Cookies. Pass --domain to
keep only buckets whose primary domain matches (e.g. youtube.com keeps
.youtube.com and music.youtube.com, not .google.com). --domain never
rewrites foreign hosts onto a different key.

Prefer the header form when a cookies.txt export is missing your login:
the browser's request Cookie header includes httpOnly cookies, which is
where the session actually lives.

Importing REPLACES every cookie stored for each written domain. Pass
--merge to keep existing records and only add or overwrite incoming names.

Examples:
  olares-cli settings integration cookie import --file cookies.txt
  olares-cli settings integration cookie import --domain youtube.com --file cookies.txt
  pbpaste | olares-cli settings integration cookie import --domain youtube.com --file - --format header
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runCookieImport(c.Context(), f, c.InOrStdin(), cookieImportOptions{
				domain: domain,
				file:   file,
				format: formatFlag,
				merge:  merge,
			})
		},
	}
	cmd.Flags().StringVar(&domain, "domain", "", "optional: only import hosts matching this domain (required for a bare Cookie header)")
	cmd.Flags().StringVar(&file, "file", "", "file to read cookies from; \"-\" reads stdin")
	cmd.Flags().StringVar(&formatFlag, "format", "auto", "input format: netscape, json, header, auto")
	cmd.Flags().BoolVar(&merge, "merge", false, "merge into each domain's existing cookies instead of replacing them")
	return cmd
}

type cookieImportOptions struct {
	domain string
	file   string
	format string
	merge  bool
}

func runCookieImport(ctx context.Context, f *cmdutil.Factory, stdin io.Reader, opts cookieImportOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	domain := strings.TrimSpace(opts.domain)
	if strings.TrimSpace(opts.file) == "" {
		return fmt.Errorf("--file is required (use \"--file -\" to read stdin)")
	}
	format, err := cookieparse.ParseFormat(opts.format)
	if err != nil {
		return err
	}

	text, err := readCookieInput(stdin, opts.file)
	if err != nil {
		return err
	}

	buckets, detected, err := parseCookieImportBuckets(text, format, domain)
	if err != nil {
		return err
	}

	pc, err := prepareSettings(ctx, f)
	if err != nil {
		return err
	}

	verb := "Replaced"
	if opts.merge {
		verb = "Merged into"
	}
	total := 0
	keys := make([]string, 0, len(buckets))
	for _, host := range sortedCookieDomains(buckets) {
		stored, err := storeCookies(ctx, pc, host, buckets[host], opts.merge)
		if err != nil {
			return err
		}
		total += len(stored.Records)
		keys = append(keys, stored.storeKey())
		fmt.Printf("%s %s: %d record(s) as %s\n",
			verb, host, len(stored.Records), stored.storeKey())
	}
	fmt.Printf("Imported %d domain(s) (%d record(s), format: %s).\n",
		len(keys), total, detected)
	return nil
}

// parseCookieImportBuckets parses the input like the Settings SPA: keep
// each file host as its own bucket. --domain only filters; it does not
// rewrite foreign hosts onto another key. A bare Cookie header has no
// host of its own, so --domain is passed through to the parser then.
func parseCookieImportBuckets(
	text string,
	format cookieparse.Format,
	domainFilter string,
) (map[string][]cookieparse.Record, cookieparse.Format, error) {
	parsed, detected, err := cookieparse.Parse(text, format, "")
	if err != nil {
		return nil, "", err
	}
	// Request-style headers carry no Domain; retry with the filter so
	// the parser can assign the host (SPA addHeaderCookies).
	if parsed.Count() == 0 && domainFilter != "" {
		parsed, detected, err = cookieparse.Parse(text, format, domainFilter)
		if err != nil {
			return nil, "", err
		}
	}
	if parsed.Count() == 0 {
		if domainFilter == "" && parsed.NeedsDomain {
			return nil, detected, fmt.Errorf(
				"this input is a browser Cookie header, which carries no domain; re-run with --domain <host> (for example --domain youtube.com)")
		}
		if len(parsed.InvalidLines) > 0 {
			return nil, detected, fmt.Errorf("no usable cookies found; first problem: %s", parsed.InvalidLines[0])
		}
		return nil, detected, fmt.Errorf("no cookies found in the input")
	}

	buckets := selectCookieImportBuckets(parsed.Cookies, domainFilter)
	if len(buckets) == 0 {
		return nil, detected, fmt.Errorf("no cookies for domain %s in the input", domainFilter)
	}
	if len(parsed.InvalidLines) > 0 {
		fmt.Printf("Skipped %d unparsable line(s); first: %s\n",
			len(parsed.InvalidLines), parsed.InvalidLines[0])
	}
	return buckets, detected, nil
}

// selectCookieImportBuckets returns the domain buckets to write. An empty
// filter keeps every host (SPA paste). A filter keeps the host itself and
// its subdomains (youtube.com ↔ .youtube.com / music.youtube.com), without
// collapsing multi-part public suffixes the way a naive two-label climb
// would (bbc.co.uk must not match amazon.co.uk).
func selectCookieImportBuckets(
	cookies map[string][]cookieparse.Record,
	filter string,
) map[string][]cookieparse.Record {
	out := make(map[string][]cookieparse.Record, len(cookies))
	filter = strings.TrimSpace(filter)
	for host, records := range cookies {
		if len(records) == 0 {
			continue
		}
		if filter != "" && !cookieDomainsMatch(host, filter) {
			continue
		}
		out[host] = records
	}
	return out
}

func sortedCookieDomains(buckets map[string][]cookieparse.Record) []string {
	out := make([]string, 0, len(buckets))
	for host := range buckets {
		out = append(out, host)
	}
	sort.Strings(out)
	return out
}

func cookieDomainHost(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	host = strings.TrimPrefix(host, ".")
	host = strings.TrimPrefix(host, "www.")
	return host
}

// cookieDomainsMatch reports whether cookieHost is the filter or a
// subdomain of it. Symmetric primary-domain equality is intentionally
// avoided: it over-matches compound TLDs (*.co.uk, *.co.jp, …).
func cookieDomainsMatch(cookieHost, filterHost string) bool {
	h, f := cookieDomainHost(cookieHost), cookieDomainHost(filterHost)
	if h == "" || f == "" {
		return false
	}
	if h == f {
		return true
	}
	return strings.HasSuffix(h, "."+f)
}

// storeCookies writes one domain's records and returns what was stored.
// merge overlays the incoming records onto the stored ones; without it
// the POST replaces the domain wholesale, which is what user-service
// does to the underlying secret either way.
func storeCookies(
	ctx context.Context,
	pc *preparedClient,
	domain string,
	records []cookieparse.Record,
	merge bool,
) (domainCookie, error) {
	account, err := cookieAccount(pc.profile.OlaresID)
	if err != nil {
		return domainCookie{}, err
	}
	if merge {
		existing, err := fetchDomainCookies(ctx, pc, domain, account)
		if err != nil {
			return domainCookie{}, err
		}
		records = mergeRecords(existing, records)
	}
	payload := domainCookie{
		Domain:     domain,
		Account:    account,
		Records:    records,
		UpdateTime: time.Now().UnixMilli(),
	}
	if err := doMutateEnvelope(ctx, pc.doer, "POST", "/api/cookie", payload, nil); err != nil {
		return domainCookie{}, err
	}
	return payload, nil
}

func readCookieInput(stdin io.Reader, file string) (string, error) {
	if file == "-" {
		if stdin == nil {
			stdin = os.Stdin
		}
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("read cookies from stdin: %w", err)
		}
		return string(data), nil
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("read cookies from %s: %w", file, err)
	}
	return string(data), nil
}

// mergeRecords overlays incoming records onto the stored ones, keyed the
// way a browser identifies a cookie: (name, path).
func mergeRecords(existing, incoming []cookieparse.Record) []cookieparse.Record {
	index := make(map[string]int, len(existing))
	merged := make([]cookieparse.Record, 0, len(existing)+len(incoming))
	for _, rec := range existing {
		index[rec.Name+"|"+rec.Path] = len(merged)
		merged = append(merged, rec)
	}
	for _, rec := range incoming {
		key := rec.Name + "|" + rec.Path
		if pos, ok := index[key]; ok {
			merged[pos] = rec
			continue
		}
		index[key] = len(merged)
		merged = append(merged, rec)
	}
	return merged
}

func newCookieListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list the domains that have stored cookies",
		Long: `List every domain with stored cookies, how many records each holds and
when the earliest of them expires.

Cookie names and values are never printed, in any output format.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runCookieList(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

// cookieSummary is the redacted view of a domain's cookies. It carries no
// names or values so it is safe to print and to pipe into a log.
type cookieSummary struct {
	Domain     string `json:"domain"`
	Account    string `json:"account"`
	Records    int    `json:"records"`
	Session    int    `json:"session"`
	Expired    int    `json:"expired"`
	NextExpiry string `json:"nextExpiry,omitempty"`
	UpdatedAt  string `json:"updatedAt,omitempty"`
}

func summarizeDomain(dc domainCookie, now time.Time) cookieSummary {
	summary := cookieSummary{
		Domain:  dc.Domain,
		Account: dc.Account,
		Records: len(dc.Records),
	}
	nowUnix := now.Unix()
	var earliest int64
	for _, rec := range dc.Records {
		// 0 means a session cookie: no expiry to report, never expired.
		expiry := rec.ExpiryUnix()
		if expiry == 0 {
			summary.Session++
			continue
		}
		if expiry <= nowUnix {
			summary.Expired++
			continue
		}
		if earliest == 0 || expiry < earliest {
			earliest = expiry
		}
	}
	if earliest > 0 {
		summary.NextExpiry = time.Unix(earliest, 0).UTC().Format(time.RFC3339)
	}
	if dc.UpdateTime > 0 {
		summary.UpdatedAt = time.UnixMilli(dc.UpdateTime).UTC().Format(time.RFC3339)
	}
	return summary
}

func runCookieList(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}

	pc, err := prepareSettings(ctx, f)
	if err != nil {
		return err
	}
	summaries, err := cookieSummaries(ctx, pc)
	if err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, summaries)
	default:
		return renderCookieTable(os.Stdout, summaries)
	}
}

func cookieSummaries(ctx context.Context, pc *preparedClient) ([]cookieSummary, error) {
	account, err := cookieAccount(pc.profile.OlaresID)
	if err != nil {
		return nil, err
	}
	var rows []domainCookie
	if err := doGetEnvelope(ctx, pc.doer, "/api/cookie/all", &rows); err != nil {
		return nil, err
	}
	now := time.Now()
	summaries := make([]cookieSummary, 0, len(rows))
	for _, row := range rows {
		if row.Account != account {
			continue
		}
		summaries = append(summaries, summarizeDomain(row, now))
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Domain < summaries[j].Domain
	})
	return summaries, nil
}

func renderCookieTable(w io.Writer, rows []cookieSummary) error {
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, "no stored cookies")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "DOMAIN\tACCOUNT\tRECORDS\tEXPIRED\tNEXT EXPIRY\tUPDATED"); err != nil {
		return err
	}
	for _, r := range rows {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%s\t%s\n",
			nonEmpty(r.Domain),
			nonEmpty(r.Account),
			r.Records,
			r.Expired,
			nonEmpty(r.NextExpiry),
			nonEmpty(r.UpdatedAt),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func newCookieRemoveCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm <domain>",
		Short: "delete every stored cookie for a domain",
		Long: `Delete every cookie stored for a domain.

Use "cookie list" first to see the exact domain strings; a leading dot
(".youtube.com") is a different key from the bare host.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runCookieRemove(c.Context(), f, args[0])
		},
	}
	return cmd
}

func runCookieRemove(ctx context.Context, f *cmdutil.Factory, domain string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return fmt.Errorf("domain is required")
	}

	pc, err := prepareSettings(ctx, f)
	if err != nil {
		return err
	}
	key, err := deleteDomainCookies(ctx, pc, domain)
	if err != nil {
		return err
	}
	fmt.Printf("Deleted cookies for %s (%s).\n", domain, key)
	return nil
}

func deleteDomainCookies(ctx context.Context, pc *preparedClient, domain string) (string, error) {
	account, err := cookieAccount(pc.profile.OlaresID)
	if err != nil {
		return "", err
	}
	key := domainCookie{Domain: domain, Account: account}.storeKey()
	if err := doMutateEnvelope(ctx, pc.doer, "POST", "/api/cookie/delete",
		map[string]string{"key": key}, nil); err != nil {
		return "", err
	}
	return key, nil
}

func newCookieValidateCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "validate <domain>",
		Short: "report whether a domain's stored cookies are still usable",
		Long: `Check the cookies stored for a domain and report how many have expired.

Exits non-zero when the domain has no cookies at all, or when every
non-session cookie has expired — so it can gate a script before a
download that needs a login. A leftover session cookie does not keep
the domain usable once all dated cookies have expired.

Session cookies (no expiry) are reported but never counted as expired.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runCookieValidate(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runCookieValidate(ctx context.Context, f *cmdutil.Factory, domain, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return fmt.Errorf("domain is required")
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}

	pc, err := prepareSettings(ctx, f)
	if err != nil {
		return err
	}
	summary, err := cookieDomainSummary(ctx, pc, domain)
	if err != nil {
		return err
	}

	if format == FormatJSON {
		if err := printJSON(os.Stdout, summary); err != nil {
			return err
		}
	} else {
		fmt.Printf("Domain:      %s\n", summary.Domain)
		fmt.Printf("Account:     %s\n", summary.Account)
		fmt.Printf("Records:     %d\n", summary.Records)
		fmt.Printf("Session:     %d\n", summary.Session)
		fmt.Printf("Expired:     %d\n", summary.Expired)
		fmt.Printf("Next expiry: %s\n", nonEmpty(summary.NextExpiry))
	}

	// Fail when every dated (non-session) cookie has expired. Session
	// cookies alone do not keep a login-gated download usable.
	if cookieValidateFailed(summary) {
		return fmt.Errorf("every non-session cookie for %s has expired; re-import them with:\n  %s",
			domain, cookieImportHint(domain))
	}
	return nil
}

// cookieValidateFailed is true when every dated cookie has expired.
// Session cookies are ignored for the gate (they never inflate Expired).
func cookieValidateFailed(summary cookieSummary) bool {
	nonSession := summary.Records - summary.Session
	return summary.Expired > 0 && summary.Expired == nonSession
}

func cookieDomainSummary(ctx context.Context, pc *preparedClient, domain string) (cookieSummary, error) {
	account, err := cookieAccount(pc.profile.OlaresID)
	if err != nil {
		return cookieSummary{}, err
	}
	records, err := fetchDomainCookies(ctx, pc, domain, account)
	if err != nil {
		return cookieSummary{}, err
	}
	if len(records) == 0 {
		return cookieSummary{}, fmt.Errorf("no cookies stored for %s; import them with:\n  %s",
			domain, cookieImportHint(domain))
	}
	return summarizeDomain(domainCookie{
		Domain:  domain,
		Account: account,
		Records: records,
	}, time.Now()), nil
}

// fetchDomainCookies returns the stored records for one (domain, account).
// user-service filters /api/cookie/retrieve by domain only; list (/api/cookie/all)
// is likewise unscoped. Both list and retrieve therefore filter by account
// on the client — same as the SPA's getAllCookies / fetchDomainCookies.
func fetchDomainCookies(ctx context.Context, pc *preparedClient, domain, account string) ([]cookieparse.Record, error) {
	var rows []domainCookie
	if err := doMutateEnvelope(ctx, pc.doer, "POST", "/api/cookie/retrieve",
		map[string]string{"domain": domain}, &rows); err != nil {
		return nil, err
	}
	var out []cookieparse.Record
	for _, row := range rows {
		if row.Account != account {
			continue
		}
		out = append(out, row.Records...)
	}
	return out, nil
}

// cookieImportHint is the copy-pasteable command we point users at from
// every "you have no usable cookies" path, including the download-side
// 501 handler in another command family.
func cookieImportHint(domain string) string {
	if domain == "" {
		domain = "<domain>"
	}
	return fmt.Sprintf(
		"olares-cli settings integration cookie import --domain %s --file cookies.txt", domain)
}
