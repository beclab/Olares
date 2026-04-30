package utils

import (
	"net/http"
	"net/http/httputil"
	"strings"
)

// sensitiveHeaders enumerates HTTP headers whose values must never be written
// to logs. Header names are matched case-insensitively.
var sensitiveHeaders = map[string]struct{}{
	"authorization":         {},
	"proxy-authorization":   {},
	"cookie":                {},
	"set-cookie":            {},
	"x-auth-signature":      {},
	"x-app-key":             {},
	"x-authorization-token": {},
	"x-backend-token":       {},
	"x-bfl-user":            {},
	"x-authelia-nonce":      {},
}

// RedactedHeader returns a shallow copy of h with the values of well-known
// sensitive headers replaced by "[REDACTED]". The input header is not
// modified.
func RedactedHeader(h http.Header) http.Header {
	if h == nil {
		return nil
	}
	out := make(http.Header, len(h))
	for k, v := range h {
		if _, ok := sensitiveHeaders[strings.ToLower(k)]; ok {
			out[k] = []string{"[REDACTED]"}
			continue
		}
		copied := make([]string, len(v))
		copy(copied, v)
		out[k] = copied
	}
	return out
}

// DumpRequestRedacted is a logging-safe replacement for httputil.DumpRequest.
// It strips sensitive headers before dumping and never includes the body to
// avoid leaking credentials carried in request payloads.
func DumpRequestRedacted(req *http.Request) ([]byte, error) {
	if req == nil {
		return nil, nil
	}
	clone := req.Clone(req.Context())
	clone.Header = RedactedHeader(req.Header)
	return httputil.DumpRequest(clone, false)
}

// DumpRequestOutRedacted is a logging-safe replacement for
// httputil.DumpRequestOut. It strips sensitive headers and excludes the body.
func DumpRequestOutRedacted(req *http.Request) ([]byte, error) {
	if req == nil {
		return nil, nil
	}
	clone := req.Clone(req.Context())
	clone.Header = RedactedHeader(req.Header)
	return httputil.DumpRequestOut(clone, false)
}

// DumpResponseRedacted dumps a response without its body and with sensitive
// headers (e.g. Set-Cookie) redacted.
func DumpResponseRedacted(resp *http.Response) ([]byte, error) {
	if resp == nil {
		return nil, nil
	}
	clone := *resp
	clone.Header = RedactedHeader(resp.Header)
	return httputil.DumpResponse(&clone, false)
}
