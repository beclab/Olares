package cookieparse

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
)

type jsonParser struct{}

func (jsonParser) Format() Format { return FormatJSON }

func (jsonParser) Detect(text string) int {
	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") {
		return 0
	}
	if !json.Valid([]byte(trimmed)) {
		return 0
	}
	return 90
}

// jsonCookie accepts the field spellings the common browser extensions
// emit. expires may be a unix number or a date string, and EditThisCookie
// and friends call it expirationDate instead.
type jsonCookie struct {
	Domain         string          `json:"domain"`
	Name           string          `json:"name"`
	Value          string          `json:"value"`
	Path           string          `json:"path"`
	Expires        json.RawMessage `json:"expires"`
	ExpirationDate json.RawMessage `json:"expirationDate"`
	Secure         bool            `json:"secure"`
	HttpOnly       bool            `json:"httpOnly"`
	SameSite       string          `json:"sameSite"`
}

func (jsonParser) Parse(text, domain string) (*Result, error) {
	trimmed := strings.TrimSpace(text)

	var items []jsonCookie
	if err := decodeJSONCookies([]byte(trimmed), &items); err != nil {
		return nil, err
	}

	result := newResult()
	for i, item := range items {
		if item.Name == "" {
			result.reject("cookie object at index %d has no name", i)
			continue
		}
		recordDomain := strings.TrimSpace(item.Domain)
		if domain != "" {
			recordDomain = domain
		}
		if recordDomain == "" {
			result.reject("cookie has no domain; pass --domain")
			continue
		}

		path := item.Path
		if path == "" {
			path = "/"
		}

		rec := Record{
			Domain:   recordDomain,
			Name:     item.Name,
			Value:    item.Value,
			Path:     path,
			Secure:   item.Secure,
			HttpOnly: item.HttpOnly,
			SameSite: item.SameSite,
		}
		rec.setExpiry(coalesceExpiry(item.Expires, item.ExpirationDate))
		result.add(rec)
	}
	return result, nil
}

// decodeJSONCookies accepts a bare array, a {"cookies": [...]} wrapper, or
// a single cookie object.
func decodeJSONCookies(data []byte, out *[]jsonCookie) error {
	if err := json.Unmarshal(data, out); err == nil {
		return nil
	}

	var wrapper struct {
		Cookies []jsonCookie `json:"cookies"`
	}
	if err := json.Unmarshal(data, &wrapper); err == nil && wrapper.Cookies != nil {
		*out = wrapper.Cookies
		return nil
	}

	var single jsonCookie
	if err := json.Unmarshal(data, &single); err != nil {
		return fmt.Errorf("input is not valid cookie JSON: %w", err)
	}
	*out = []jsonCookie{single}
	return nil
}

func coalesceExpiry(candidates ...json.RawMessage) int64 {
	for _, raw := range candidates {
		if len(raw) == 0 {
			continue
		}
		if v, ok := parseExpiry(raw); ok {
			return v
		}
	}
	return 0
}

// parseExpiry reads either a unix-seconds number or a parseable date string.
func parseExpiry(raw json.RawMessage) (int64, bool) {
	var num float64
	if err := json.Unmarshal(raw, &num); err == nil {
		if math.IsNaN(num) || math.IsInf(num, 0) || num < 0 {
			return 0, false
		}
		return int64(num), true
	}

	var s string
	if err := json.Unmarshal(raw, &s); err != nil || strings.TrimSpace(s) == "" {
		return 0, false
	}
	for _, layout := range []string{time.RFC3339, time.RFC1123, http1123Layout} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Unix(), true
		}
	}
	return 0, false
}

// http1123Layout is the Set-Cookie / Expires wire format (GMT, not UTC).
const http1123Layout = "Mon, 02 Jan 2006 15:04:05 GMT"
