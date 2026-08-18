package cookieparse

import (
	"strconv"
	"strings"
	"time"
)

type headerParser struct{}

func (headerParser) Format() Format { return FormatHeader }

func (headerParser) Detect(text string) int {
	score := 0
	sampled := 0
	for _, raw := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if sampled >= 5 {
			break
		}
		sampled++

		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "set-cookie:") {
			score += 10
			continue
		}

		// Strip optional "Cookie:" so a DevTools paste matches a bare line.
		line := trimmed
		if strings.HasPrefix(lower, "cookie:") {
			line = strings.TrimSpace(trimmed[len("cookie:"):])
		}

		// Request Cookie lines (`a=b; c=d`) must clear Detect's threshold of 5.
		parts := splitAndTrim(line, ";")
		if isBrowserCookieLine(parts) {
			score += 10
			continue
		}
		if hasCookieAttribute(line) {
			score += 5
		}
		if strings.Contains(line, ";") && strings.Contains(line, "=") {
			score += 2
		}
	}
	if score > 100 {
		return 100
	}
	return score
}

func hasCookieAttribute(line string) bool {
	for _, part := range strings.Split(line, ";")[1:] {
		if attributeName(part) != "" {
			return true
		}
	}
	return false
}

// attributeName returns the canonical name of a Set-Cookie attribute, or
// "" when the segment is a plain name=value pair.
func attributeName(part string) string {
	lower := strings.ToLower(strings.TrimSpace(part))
	switch lower {
	case "secure", "httponly":
		return lower
	}
	for _, prefix := range []string{"domain=", "path=", "expires=", "max-age=", "samesite="} {
		if strings.HasPrefix(lower, prefix) {
			return strings.TrimSuffix(prefix, "=")
		}
	}
	return ""
}

func (headerParser) Parse(text, domain string) (*Result, error) {
	result := newResult()

	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if lower := strings.ToLower(line); strings.HasPrefix(lower, "set-cookie:") {
			line = strings.TrimSpace(line[len("set-cookie:"):])
		} else if strings.HasPrefix(lower, "cookie:") {
			line = strings.TrimSpace(line[len("cookie:"):])
		}

		parts := splitAndTrim(line, ";")
		if len(parts) == 0 {
			continue
		}

		if isBrowserCookieLine(parts) {
			parseBrowserCookieLine(result, parts, domain)
			continue
		}
		parseSetCookieLine(result, parts, domain)
	}
	return result, nil
}

// isBrowserCookieLine distinguishes a request header (`a=b; c=d`, every
// segment a pair, no attributes) from a Set-Cookie response header.
func isBrowserCookieLine(parts []string) bool {
	if len(parts) < 2 {
		return false
	}
	for _, part := range parts {
		if attributeName(part) != "" {
			return false
		}
		if !strings.Contains(part, "=") {
			return false
		}
	}
	return true
}

// parseBrowserCookieLine expands `a=b; c=d` into one record per pair. Such
// a header carries no metadata, so the domain must come from the caller;
// path defaults to "/" and the flags stay false, matching the SPA.
func parseBrowserCookieLine(result *Result, parts []string, domain string) {
	if domain == "" {
		// One reject for the whole line — never per pair, never embed names.
		result.NeedsDomain = true
		result.reject("a browser Cookie header carries no domain; pass --domain")
		return
	}
	for _, part := range parts {
		name, value, ok := splitPair(part)
		if !ok || name == "" {
			continue
		}
		rec := Record{
			Domain: domain,
			Name:   name,
			Value:  value,
			Path:   "/",
		}
		rec.setExpiry(0)
		result.add(rec)
	}
}

func parseSetCookieLine(result *Result, parts []string, domain string) {
	name, value, ok := splitPair(parts[0])
	if !ok {
		// Never echo parts[0]: it may be name=value.
		result.reject("not a cookie assignment")
		return
	}
	if name == "" {
		result.reject("empty cookie name")
		return
	}

	rec := Record{Name: name, Value: value, Path: "/"}
	var expires int64

	for _, part := range parts[1:] {
		attrValue := ""
		if _, v, ok := splitPair(part); ok {
			attrValue = v
		}
		switch attributeName(part) {
		case "secure":
			rec.Secure = true
		case "httponly":
			rec.HttpOnly = true
		case "domain":
			rec.Domain = attrValue
		case "path":
			rec.Path = attrValue
		case "samesite":
			rec.SameSite = attrValue
		case "expires":
			if t, ok := parseHeaderDate(attrValue); ok {
				expires = t
			} else {
				result.reject("invalid expires date")
			}
		case "max-age":
			if seconds, err := strconv.ParseInt(strings.TrimSpace(attrValue), 10, 64); err == nil {
				expires = time.Now().Unix() + seconds
			}
		}
	}

	if domain != "" {
		rec.Domain = domain
	}
	if rec.Domain == "" {
		result.reject("cookie has no Domain attribute; pass --domain")
		return
	}
	if rec.Path == "" {
		rec.Path = "/"
	}

	rec.setExpiry(expires)
	result.add(rec)
}

func parseHeaderDate(s string) (int64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	for _, layout := range []string{
		http1123Layout,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC850,
		time.ANSIC,
		time.RFC3339,
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Unix(), true
		}
	}
	return 0, false
}

func splitAndTrim(s, sep string) []string {
	var out []string
	for _, part := range strings.Split(s, sep) {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// splitPair splits on the first '=' so values containing '=' survive.
func splitPair(s string) (name, value string, ok bool) {
	idx := strings.Index(s, "=")
	if idx < 0 {
		return "", "", false
	}
	return strings.TrimSpace(s[:idx]), strings.TrimSpace(s[idx+1:]), true
}
