// Package cookieparse turns the three cookie exchange formats a user can
// realistically get hold of (Netscape cookies.txt, JSON exports, raw HTTP
// headers) into the DomainCookie shape user-service stores in Infisical.
//
// Kept free of CLI and HTTP dependencies so the format handling can be
// tested on its own. The semantics mirror the Settings SPA's
// packages/app/src/stores/settings/cookieParser.ts, because both clients
// write to the same secret and users move between them.
package cookieparse

import (
	"fmt"
	"strings"
)

// Format identifies one of the supported input encodings.
type Format string

const (
	FormatNetscape Format = "netscape"
	FormatJSON     Format = "json"
	FormatHeader   Format = "header"
	FormatAuto     Format = "auto"
)

// SupportedFormats lists the values accepted by ParseFormat, in the order
// they should be shown to a user.
var SupportedFormats = []Format{FormatNetscape, FormatJSON, FormatHeader, FormatAuto}

// ParseFormat validates a user-supplied --format value.
func ParseFormat(s string) (Format, error) {
	switch Format(strings.ToLower(strings.TrimSpace(s))) {
	case "", FormatAuto:
		return FormatAuto, nil
	case FormatNetscape:
		return FormatNetscape, nil
	case FormatJSON:
		return FormatJSON, nil
	case FormatHeader:
		return FormatHeader, nil
	default:
		return "", fmt.Errorf("unsupported --format %q (allowed: netscape, json, header, auto)", s)
	}
}

// Record mirrors user-service's DomainCookieRecord (src/utils.ts).
//
// ExpirationDate duplicates Expires on purpose: the SPA and user-service
// write `expires`, while download-server and the yt-dlp daemon read
// `expirationDate`. Emitting both is what makes the expiry survive the
// whole round trip.
//
// Both are float64 because the SPA stores Date.now()/1000, so real stored
// records carry fractional seconds; an int64 here fails to decode them.
type Record struct {
	Domain         string  `json:"domain"`
	Name           string  `json:"name"`
	Value          string  `json:"value"`
	Expires        float64 `json:"expires"`
	ExpirationDate float64 `json:"expirationDate"`
	Path           string  `json:"path"`
	Secure         bool    `json:"secure"`
	HttpOnly       bool    `json:"httpOnly"`
	SameSite       string  `json:"sameSite,omitempty"`
}

func (r *Record) setExpiry(unixSeconds int64) {
	r.Expires = float64(unixSeconds)
	r.ExpirationDate = float64(unixSeconds)
}

// ExpiryUnix returns the record's expiry in whole unix seconds, preferring
// `expires` and falling back to `expirationDate`. 0 means a session cookie.
func (r Record) ExpiryUnix() int64 {
	expiry := r.Expires
	if expiry == 0 {
		expiry = r.ExpirationDate
	}
	if expiry <= 0 {
		return 0
	}
	return int64(expiry)
}

// Result is one parse pass: cookies grouped by the domain they belong to,
// plus the lines we refused, so the caller can report them without having
// to re-parse.
type Result struct {
	Cookies      map[string][]Record
	InvalidLines []string
}

func newResult() *Result {
	return &Result{Cookies: make(map[string][]Record)}
}

func (r *Result) add(rec Record) {
	r.Cookies[rec.Domain] = append(r.Cookies[rec.Domain], rec)
}

func (r *Result) reject(format string, args ...interface{}) {
	r.InvalidLines = append(r.InvalidLines, fmt.Sprintf(format, args...))
}

// Count returns the number of successfully parsed records.
func (r *Result) Count() int {
	n := 0
	for _, records := range r.Cookies {
		n += len(records)
	}
	return n
}

// Domains returns the domains present in the result.
func (r *Result) Domains() []string {
	out := make([]string, 0, len(r.Cookies))
	for domain := range r.Cookies {
		out = append(out, domain)
	}
	return out
}

// Parser is one format handler.
type Parser interface {
	Format() Format
	// Detect scores how much the input looks like this format, 0..100.
	Detect(text string) int
	// Parse converts the input. domain is the user-supplied --domain and
	// is only consulted by formats that cannot carry one themselves.
	Parse(text, domain string) (*Result, error)
}

func parsers() []Parser {
	return []Parser{netscapeParser{}, jsonParser{}, headerParser{}}
}

// Detect picks the best-scoring parser, or nil when nothing scores high
// enough to be worth guessing at.
func Detect(text string) Parser {
	var best Parser
	bestScore := 0
	for _, p := range parsers() {
		if score := p.Detect(text); score > bestScore {
			best, bestScore = p, score
		}
	}
	if bestScore < 5 {
		return nil
	}
	return best
}

// Parse converts text using the requested format. FormatAuto detects it.
//
// domain is required for a browser-style Cookie header, which carries no
// domain of its own; the other formats use it to override what they read.
func Parse(text string, format Format, domain string) (*Result, Format, error) {
	if strings.TrimSpace(text) == "" {
		return nil, "", fmt.Errorf("cookie input is empty")
	}

	if format == FormatAuto {
		p := Detect(text)
		if p == nil {
			msg := "could not tell which cookie format this is; pass --format netscape|json|header"
			if strings.Contains(text, ";") && strings.Contains(text, "=") {
				msg += "; a Cookie request header needs --format header --domain <host>"
			}
			return nil, "", fmt.Errorf("%s", msg)
		}
		res, err := p.Parse(text, domain)
		return res, p.Format(), err
	}

	for _, p := range parsers() {
		if p.Format() != format {
			continue
		}
		res, err := p.Parse(text, domain)
		return res, format, err
	}
	return nil, "", fmt.Errorf("unsupported format %q", format)
}
