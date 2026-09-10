package cookieparse

import (
	"fmt"
	"strconv"
	"strings"
)

// httpOnlyPrefix is how the Netscape format marks an httpOnly cookie: the
// domain field is prefixed, which makes the line indistinguishable from a
// comment unless matched explicitly. Dropping these lines silently loses
// exactly the cookies that carry a login (YouTube's SID / __Secure-3PSID
// / HSID are all httpOnly).
const httpOnlyPrefix = "#HttpOnly_"

type netscapeParser struct{}

func (netscapeParser) Format() Format { return FormatNetscape }

// splitHTTPOnly strips the httpOnly marker and reports whether it was there.
func splitHTTPOnly(line string) (string, bool) {
	if len(line) >= len(httpOnlyPrefix) &&
		strings.EqualFold(line[:len(httpOnlyPrefix)], httpOnlyPrefix) {
		return line[len(httpOnlyPrefix):], true
	}
	return line, false
}

// isDataLine separates real records from comments. An httpOnly record
// starts with '#' yet is data, so it has to be checked before the comment
// rule.
func isDataLine(trimmed string) bool {
	if trimmed == "" {
		return false
	}
	if _, httpOnly := splitHTTPOnly(trimmed); httpOnly {
		return true
	}
	return !strings.HasPrefix(trimmed, "#")
}

func (netscapeParser) Detect(text string) int {
	score := 0
	sampled := 0
	for _, raw := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(raw)
		if !isDataLine(trimmed) {
			continue
		}
		if sampled >= 5 {
			break
		}
		sampled++

		line, _ := splitHTTPOnly(trimmed)
		fields := strings.Split(line, "\t")
		switch {
		case len(fields) >= 7:
			score += 10
		case len(fields) >= 4:
			score += 5
		}
		if len(fields) >= 7 && isNetscapeBool(fields[1]) && isNetscapeBool(fields[3]) {
			score += 8
		}
	}
	if score > 100 {
		return 100
	}
	return score
}

func isNetscapeBool(s string) bool {
	v := strings.ToUpper(strings.TrimSpace(s))
	return v == "TRUE" || v == "FALSE"
}

func (netscapeParser) Parse(text, domain string) (*Result, error) {
	if head := strings.TrimSpace(text); strings.HasPrefix(head, "{") || strings.HasPrefix(head, "[") {
		return nil, fmt.Errorf("input looks like JSON, not a Netscape cookies.txt file; use --format json")
	}

	result := newResult()
	for _, raw := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(raw)
		if !isDataLine(trimmed) {
			continue
		}

		line, httpOnly := splitHTTPOnly(trimmed)
		fields := strings.Split(line, "\t")
		if len(fields) < 7 {
			// Never echo the raw line: it carries the cookie value.
			result.reject("need at least 7 tab-separated fields")
			continue
		}

		recordDomain := strings.TrimSpace(fields[0])
		if recordDomain == "" {
			result.reject("empty domain")
			continue
		}
		name := strings.TrimSpace(fields[5])
		if name == "" {
			result.reject("empty cookie name")
			continue
		}

		// The value is the remainder: a value may legitimately contain tabs.
		value := strings.TrimSpace(strings.Join(fields[6:], "\t"))

		// expires=0 means a session cookie and must be kept, not rejected.
		var expires int64
		if raw := strings.TrimSpace(fields[4]); raw != "" {
			parsed, err := strconv.ParseInt(raw, 10, 64)
			if err != nil || parsed < 0 {
				result.reject("invalid expiration")
				continue
			}
			expires = parsed
		}

		path := strings.TrimSpace(fields[2])
		if path == "" {
			path = "/"
		}
		if !strings.HasPrefix(path, "/") {
			result.reject("invalid path")
			continue
		}

		if domain != "" {
			recordDomain = domain
		}

		rec := Record{
			Domain:   recordDomain,
			Name:     name,
			Value:    value,
			Path:     path,
			Secure:   strings.EqualFold(strings.TrimSpace(fields[3]), "TRUE"),
			HttpOnly: httpOnly,
		}
		rec.setExpiry(expires)
		result.add(rec)
	}
	return result, nil
}
