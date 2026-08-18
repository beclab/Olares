package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/cookieparse"
	"github.com/beclab/Olares/cli/pkg/credential"
)

func testClient(olaresID string) (*preparedClient, *fakeDoer) {
	doer := &fakeDoer{}
	return &preparedClient{
		profile: &credential.ResolvedProfile{OlaresID: olaresID},
		doer:    doer,
	}, doer
}

func TestCookieAccountMirrorsTheSPA(t *testing.T) {
	got, err := cookieAccount("alice@olares.com")
	if err != nil {
		t.Fatalf("cookieAccount: %v", err)
	}
	if got != "alice" {
		t.Fatalf("account = %q, want alice", got)
	}
	if _, err := cookieAccount("  "); err == nil {
		t.Fatal("expected an error for a profile without an Olares ID")
	}
}

// download-server looks a cookie up as cookie:{domain}:{user}, so an
// account-less key is a row nothing can read.
func TestStoreKeyCarriesTheAccount(t *testing.T) {
	withAccount := domainCookie{Domain: "youtube.com", Account: "alice"}.storeKey()
	if withAccount != "cookie:youtube.com:alice" {
		t.Fatalf("store key = %q", withAccount)
	}
	if bare := (domainCookie{Domain: "youtube.com"}).storeKey(); bare != "cookie:youtube.com" {
		t.Fatalf("store key = %q", bare)
	}
}

func TestStoreCookiesReplacesTheDomainByDefault(t *testing.T) {
	pc, doer := testClient("alice@olares.com")

	stored, err := storeCookies(context.Background(), pc, "youtube.com",
		[]cookieparse.Record{{Domain: "youtube.com", Name: "SID", Value: "x"}}, false)
	if err != nil {
		t.Fatalf("storeCookies: %v", err)
	}
	if len(doer.calls) != 1 {
		t.Fatalf("calls = %d, want 1 (no retrieve without --merge)", len(doer.calls))
	}
	call := doer.lastCall()
	if call.method != "POST" || call.path != "/api/cookie" {
		t.Fatalf("wrote to %s %s", call.method, call.path)
	}
	if stored.Account != "alice" || stored.storeKey() != "cookie:youtube.com:alice" {
		t.Fatalf("stored under %q", stored.storeKey())
	}
	if stored.UpdateTime == 0 {
		t.Fatal("updateTime not stamped")
	}
}

func TestStoreCookiesMergeKeepsUntouchedNames(t *testing.T) {
	pc, doer := testClient("alice@olares.com")
	doer.enqueueEnvelope([]domainCookie{
		{
			Domain:  "youtube.com",
			Account: "alice",
			Records: []cookieparse.Record{
				{Domain: "youtube.com", Name: "SID", Value: "old", Path: "/"},
				{Domain: "youtube.com", Name: "HSID", Value: "keep", Path: "/"},
			},
		},
		// Another user's row for the same domain must not leak in.
		{
			Domain:  "youtube.com",
			Account: "bob",
			Records: []cookieparse.Record{{Domain: "youtube.com", Name: "BOB", Value: "no", Path: "/"}},
		},
	})

	stored, err := storeCookies(context.Background(), pc, "youtube.com", []cookieparse.Record{
		{Domain: "youtube.com", Name: "SID", Value: "new", Path: "/"},
		{Domain: "youtube.com", Name: "PREF", Value: "added", Path: "/"},
	}, true)
	if err != nil {
		t.Fatalf("storeCookies: %v", err)
	}

	got := map[string]string{}
	for _, rec := range stored.Records {
		got[rec.Name] = rec.Value
	}
	want := map[string]string{"SID": "new", "HSID": "keep", "PREF": "added"}
	if len(got) != len(want) {
		t.Fatalf("merged records = %v, want %v", got, want)
	}
	for name, value := range want {
		if got[name] != value {
			t.Fatalf("%s = %q, want %q", name, got[name], value)
		}
	}
	if doer.calls[0].path != "/api/cookie/retrieve" {
		t.Fatalf("merge did not read first: %s", doer.calls[0].path)
	}
}

func TestSummarizeDomainCountsExpiryWithoutLeakingValues(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	summary := summarizeDomain(domainCookie{
		Domain:  "youtube.com",
		Account: "alice",
		Records: []cookieparse.Record{
			// Fractional, as the SPA writes it.
			{Name: "SID", Value: "secret", Expires: float64(now.Unix()-60) + 0.5, ExpirationDate: float64(now.Unix()-60) + 0.5},
			{Name: "HSID", Value: "secret", Expires: float64(now.Unix() + 3600), ExpirationDate: float64(now.Unix() + 3600)},
			{Name: "SESSION", Value: "secret"},
		},
		UpdateTime: now.UnixMilli(),
	}, now)

	if summary.Records != 3 {
		t.Fatalf("records = %d", summary.Records)
	}
	if summary.Expired != 1 {
		t.Fatalf("expired = %d, want 1", summary.Expired)
	}
	if summary.Session != 1 {
		t.Fatalf("session = %d, want 1", summary.Session)
	}
	wantNext := time.Unix(now.Unix()+3600, 0).UTC().Format(time.RFC3339)
	if summary.NextExpiry != wantNext {
		t.Fatalf("nextExpiry = %q, want %q", summary.NextExpiry, wantNext)
	}
	if summary.UpdatedAt == "" {
		t.Fatalf("summary = %+v", summary)
	}

	raw, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(raw), "secret") || strings.Contains(string(raw), "SID") {
		t.Fatalf("summary leaks cookie data: %s", raw)
	}

	var table bytes.Buffer
	if err := renderCookieTable(&table, []cookieSummary{summary}); err != nil {
		t.Fatalf("renderCookieTable: %v", err)
	}
	if strings.Contains(table.String(), "secret") || strings.Contains(table.String(), "SID") {
		t.Fatalf("table leaks cookie data: %s", table.String())
	}
}

// A domain whose every dated cookie has expired reports no next expiry.
func TestSummarizeDomainAllExpired(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	summary := summarizeDomain(domainCookie{
		Domain:  "youtube.com",
		Account: "alice",
		Records: []cookieparse.Record{
			{Name: "SID", Expires: float64(now.Unix() - 120)},
			{Name: "HSID", Expires: float64(now.Unix() - 60)},
			{Name: "SESSION"},
		},
	}, now)
	if summary.Expired != 2 {
		t.Fatalf("expired = %d, want 2", summary.Expired)
	}
	if summary.Session != 1 {
		t.Fatalf("session = %d, want 1", summary.Session)
	}
	if summary.NextExpiry != "" {
		t.Fatalf("nextExpiry = %q, want empty", summary.NextExpiry)
	}
	if !cookieValidateFailed(summary) {
		t.Fatal("all dated cookies expired (plus a session leftover) must fail validate")
	}
}

// A session-only domain has nothing to expire, so it must not be
// reported as expired and must not claim a next expiry.
func TestSummarizeDomainSessionOnly(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	summary := summarizeDomain(domainCookie{
		Domain:  "example.com",
		Records: []cookieparse.Record{{Name: "a"}, {Name: "b"}},
	}, now)
	if summary.Expired != 0 || summary.Session != 2 || summary.NextExpiry != "" {
		t.Fatalf("summary = %+v", summary)
	}
	if cookieValidateFailed(summary) {
		t.Fatal("session-only domains must pass validate")
	}
}

func TestCookieValidateFailedAllDatedExpired(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	allDated := summarizeDomain(domainCookie{
		Domain: "youtube.com",
		Records: []cookieparse.Record{
			{Name: "SID", Expires: float64(now.Unix() - 120)},
			{Name: "HSID", Expires: float64(now.Unix() - 60)},
		},
	}, now)
	if !cookieValidateFailed(allDated) {
		t.Fatal("every dated cookie expired must fail validate")
	}

	stillFresh := summarizeDomain(domainCookie{
		Domain: "youtube.com",
		Records: []cookieparse.Record{
			{Name: "SID", Expires: float64(now.Unix() - 120)},
			{Name: "HSID", Expires: float64(now.Unix() + 60)},
		},
	}, now)
	if cookieValidateFailed(stillFresh) {
		t.Fatal("a fresh dated cookie must pass validate")
	}
}

func TestCookieDomainSummaryPointsAtImportWhenEmpty(t *testing.T) {
	pc, doer := testClient("alice@olares.com")
	doer.enqueueEnvelope([]domainCookie{})

	_, err := cookieDomainSummary(context.Background(), pc, "youtube.com")
	if err == nil {
		t.Fatal("expected an error for a domain with no cookies")
	}
	if !strings.Contains(err.Error(), cookieImportHint("youtube.com")) {
		t.Fatalf("error does not give a runnable command: %v", err)
	}
}

func TestDeleteDomainCookiesSendsTheStoreKey(t *testing.T) {
	pc, doer := testClient("alice@olares.com")

	key, err := deleteDomainCookies(context.Background(), pc, "youtube.com")
	if err != nil {
		t.Fatalf("deleteDomainCookies: %v", err)
	}
	if key != "cookie:youtube.com:alice" {
		t.Fatalf("key = %q", key)
	}
	call := doer.lastCall()
	if call.method != "POST" || call.path != "/api/cookie/delete" {
		t.Fatalf("called %s %s", call.method, call.path)
	}
	body, ok := call.body.(map[string]string)
	if !ok || body["key"] != key {
		t.Fatalf("body = %#v", call.body)
	}
}

func TestCookieSummariesSortsByDomain(t *testing.T) {
	pc, doer := testClient("alice@olares.com")
	doer.enqueueEnvelope([]domainCookie{
		{Domain: "youtube.com", Account: "alice"},
		{Domain: "bilibili.com", Account: "alice"},
	})

	summaries, err := cookieSummaries(context.Background(), pc)
	if err != nil {
		t.Fatalf("cookieSummaries: %v", err)
	}
	if len(summaries) != 2 || summaries[0].Domain != "bilibili.com" {
		t.Fatalf("summaries = %+v", summaries)
	}
	if doer.lastCall().path != "/api/cookie/all" {
		t.Fatalf("read from %s", doer.lastCall().path)
	}
}

// /api/cookie/all is account-unscoped; list must keep only the current
// profile's rows, same filter fetchDomainCookies applies on retrieve.
func TestCookieSummariesFiltersByAccount(t *testing.T) {
	pc, doer := testClient("alice@olares.com")
	doer.enqueueEnvelope([]domainCookie{
		{Domain: "youtube.com", Account: "alice", Records: []cookieparse.Record{{Name: "SID"}}},
		{Domain: "youtube.com", Account: "bob", Records: []cookieparse.Record{{Name: "BOB"}}},
		{Domain: "bilibili.com", Account: "bob"},
		{Domain: "huggingface.co", Account: "alice"},
	})

	summaries, err := cookieSummaries(context.Background(), pc)
	if err != nil {
		t.Fatalf("cookieSummaries: %v", err)
	}
	if len(summaries) != 2 {
		t.Fatalf("summaries = %+v, want 2 (alice only)", summaries)
	}
	if summaries[0].Domain != "huggingface.co" || summaries[0].Account != "alice" {
		t.Fatalf("first = %+v", summaries[0])
	}
	if summaries[1].Domain != "youtube.com" || summaries[1].Account != "alice" {
		t.Fatalf("second = %+v", summaries[1])
	}
}

func TestRunCookieImportRejectsMissingFlags(t *testing.T) {
	cases := []struct {
		name string
		opts cookieImportOptions
		want string
	}{
		{"no file", cookieImportOptions{domain: "youtube.com"}, "--file is required"},
		{"bad format", cookieImportOptions{file: "-", format: "xml"}, "unsupported --format"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := runCookieImport(context.Background(), nil, strings.NewReader(""), tc.opts)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
		})
	}
}

// Nothing may reach the network before the input parses, so an input
// that yields no usable record has to fail on its own.
func TestRunCookieImportRejectsUnusableInput(t *testing.T) {
	stdin := strings.NewReader("# Netscape HTTP Cookie File\nnot a cookie line\n")
	err := runCookieImport(context.Background(), nil, stdin, cookieImportOptions{
		domain: "youtube.com",
		file:   "-",
		format: "netscape",
	})
	if err == nil || !strings.Contains(err.Error(), "no usable cookies") {
		t.Fatalf("err = %v", err)
	}
}

func TestCookieDomainsMatchFilterHost(t *testing.T) {
	cases := []struct {
		host, filter string
		want         bool
	}{
		{".youtube.com", "youtube.com", true},
		{"music.youtube.com", "youtube.com", true},
		{"WWW.YouTube.com", "youtube.com", true},
		{".google.com", "youtube.com", false},
		{"bbc.co.uk", "bbc.co.uk", true},
		{".bbc.co.uk", "bbc.co.uk", true},
		{"www.bbc.co.uk", "bbc.co.uk", true},
		{"amazon.co.uk", "bbc.co.uk", false},
		{"example.co.jp", "bbc.co.uk", false},
		{"notyoutube.com", "youtube.com", false},
	}
	for _, tc := range cases {
		if got := cookieDomainsMatch(tc.host, tc.filter); got != tc.want {
			t.Fatalf("cookieDomainsMatch(%q, %q) = %v, want %v", tc.host, tc.filter, got, tc.want)
		}
	}
}

func TestParseCookieImportBucketsMultiDomainAndFilter(t *testing.T) {
	text := strings.Join([]string{
		"# Netscape HTTP Cookie File",
		".youtube.com\tTRUE\t/\tTRUE\t0\tSID\tyt-secret",
		".google.com\tTRUE\t/\tFALSE\t0\tNID\tg-secret",
		"music.youtube.com\tTRUE\t/\tTRUE\t0\tPREF\tm-secret",
	}, "\n")

	all, _, err := parseCookieImportBuckets(text, cookieparse.FormatNetscape, "")
	if err != nil {
		t.Fatalf("parse all: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("buckets = %v, want 3 hosts", sortedCookieDomains(all))
	}

	filtered, _, err := parseCookieImportBuckets(text, cookieparse.FormatNetscape, "youtube.com")
	if err != nil {
		t.Fatalf("parse filter: %v", err)
	}
	if len(filtered) != 2 {
		t.Fatalf("filtered = %v, want youtube hosts only", sortedCookieDomains(filtered))
	}
	if _, ok := filtered[".google.com"]; ok {
		t.Fatal("google must not survive a youtube.com filter")
	}
	if host := filtered[".youtube.com"][0].Domain; host != ".youtube.com" {
		t.Fatalf("file domain rewritten to %q", host)
	}

	_, _, err = parseCookieImportBuckets(text, cookieparse.FormatNetscape, "not-in-file.com")
	if err == nil || !strings.Contains(err.Error(), "no cookies for domain") {
		t.Fatalf("err = %v, want no cookies for domain", err)
	}
}

func TestParseCookieImportBucketsHeaderNeedsDomain(t *testing.T) {
	const want = "this input is a browser Cookie header, which carries no domain; re-run with --domain <host> (for example --domain youtube.com)"

	_, _, err := parseCookieImportBuckets("SID=sid-value; HSID=hsid-value", cookieparse.FormatHeader, "")
	if err == nil {
		t.Fatal("want bare header without --domain to fail")
	}
	if err.Error() != want {
		t.Fatalf("err = %q, want %q", err.Error(), want)
	}
	for _, leak := range []string{"SID", "HSID", "sid-value", "hsid-value", "no usable cookies", "first problem"} {
		if strings.Contains(err.Error(), leak) {
			t.Fatalf("error must not contain %q: %q", leak, err.Error())
		}
	}

	buckets, _, err := parseCookieImportBuckets("SID=sid-value; HSID=hsid-value", cookieparse.FormatHeader, "youtube.com")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(buckets) != 1 || len(buckets["youtube.com"]) != 2 {
		t.Fatalf("buckets = %+v", buckets)
	}
}

func TestParseCookieImportBucketsPreservesOtherFailures(t *testing.T) {
	_, _, err := parseCookieImportBuckets("not-a-cookie-line", cookieparse.FormatHeader, "youtube.com")
	if err == nil {
		t.Fatal("want malformed header with --domain to fail")
	}
	msg := err.Error()
	if !strings.Contains(msg, "no usable cookies found; first problem:") {
		t.Fatalf("err = %q, want first-problem message", msg)
	}
	if strings.Contains(msg, "browser Cookie header") {
		t.Fatalf("must not use the bare-header --domain hint: %q", msg)
	}

	_, _, err = parseCookieImportBuckets("# Netscape HTTP Cookie File\n# empty\n", cookieparse.FormatNetscape, "")
	if err == nil {
		t.Fatal("want empty netscape to fail")
	}
	if err.Error() != "no cookies found in the input" {
		t.Fatalf("err = %q, want no cookies found", err.Error())
	}

	_, _, err = parseCookieImportBuckets(
		".youtube.com\tTRUE\trelative\tFALSE\t0\tSID\tsecret\n",
		cookieparse.FormatNetscape,
		"",
	)
	if err == nil {
		t.Fatal("want invalid netscape rows to fail")
	}
	if !strings.Contains(err.Error(), "no usable cookies found; first problem: invalid path") {
		t.Fatalf("err = %q, want first problem: invalid path", err.Error())
	}
	if strings.Contains(err.Error(), "secret") {
		t.Fatalf("error must not embed cookie value: %q", err.Error())
	}
}

func TestMergeRecordsKeysOnNameAndPath(t *testing.T) {
	merged := mergeRecords(
		[]cookieparse.Record{
			{Name: "SID", Path: "/", Value: "old"},
			{Name: "SID", Path: "/watch", Value: "scoped"},
		},
		[]cookieparse.Record{{Name: "SID", Path: "/", Value: "new"}},
	)
	if len(merged) != 2 {
		t.Fatalf("merged = %+v", merged)
	}
	if merged[0].Value != "new" || merged[1].Value != "scoped" {
		t.Fatalf("merged = %+v", merged)
	}
}
